package main

import (
	"github.com/monsti/rpc/client"
	"github.com/monsti/util/template"
	utesting "github.com/monsti/util/testing"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitFirstDir(t *testing.T) {
	tests := []struct {
		Path, First string
	}{
		{"", ""},
		{"foo", "foo"},
		{"foo/", "foo"},
		{"foo/bar", "foo"},
		{"/", ""},
		{"/foo", "foo"},
		{"/foo/", "foo"},
		{"/foo/bar", "foo"}}
	for _, test := range tests {
		ret := splitFirstDir(test.Path)
		if ret != test.First {
			t.Errorf("splitFirstDir(%q) = %q, should be %q", test.Path, ret,
				test.First)
		}
	}
}

func TestRenderInMaster(t *testing.T) {
	masterTmpl := `{{.Page.Title}}
{{.Page.Description}}
{{range .Page.PrimaryNav}}#{{if .Active}}a{{end}}|{{.Target}}|{{.Name}}{{end}}
{{if .Page.ShowSecondaryNav}}
{{range .Page.SecondaryNav}}#{{if .Active}}a{{end}}{{if .Child}}c{{end}}|{{.Target}}|{{.Name}}{{end}}
{{end}}
{{with .Page.Sidebar}}
{{.}}
{{end}}
{{.Page.Content}}`
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{
		"/data/foo/node.yaml":               "title: Foo",
		"/data/foo/child1/node.yaml":        "title: Foo Child 1",
		"/data/foo/child2/node.yaml":        "title: Foo Child 2",
		"/data/foo/child2/child1/node.yaml": "title: Foo Child 2 Child 1",
		"/data/bar/node.yaml":               "title: Bar",
		"/data/cruz/node.yaml":              "title: Cruz",
		"/templates/master.html":            masterTmpl}, "_monsti_TestRenderInMaster")
	if err != nil {
		t.Fatalf("Could not create temporary files: ", err)
	}
	defer cleanup()
	renderer := template.Renderer{Root: filepath.Join(root, "templates")}
	site := site{}
	site.Directories.Data = filepath.Join(root, "data")
	tests := []struct {
		Node              client.Node
		Flags             masterTmplFlags
		Content, Rendered string
	}{
		{client.Node{Title: "Foo Child 2", Description: "Bar!", Path: "/foo/child2"}, 0,
			"The content.", `Foo Child 2
Bar!
#|/bar/|Bar#|/cruz/|Cruz#a|/foo/|Foo
#|/foo/child1/|Foo Child 1#a|/foo/child2/|Foo Child 2#c|/foo/child2/child1/|Foo Child 2 Child 1
The content.`}}
	for i, v := range tests {
		session := client.Session{
			User: &client.User{Login: "admin", Name: "Administrator"}}
		env := masterTmplEnv{v.Node, &session, "", "", 0}
		ret := renderInMaster(renderer, []byte(v.Content), env, settings{},
			site, "")
		for strings.Contains(ret, "\n\n") {
			ret = strings.Replace(ret, "\n\n", "\n", -1)
		}
		if ret != v.Rendered {
			t.Errorf(`Test %v: renderInMaster(...) returned:
================================================================
%v
================================================================

Should be:
================================================================
%v
================================================================`,
				i, ret, v.Rendered)
		}
	}
}
