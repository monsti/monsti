package main

import (
	"github.com/monsti/rpc/client"
	"github.com/monsti/util/template"
	"io/ioutil"
	"path/filepath"
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
	root, err := ioutil.TempDir("", "_monsti_TestRenderInMaster")
	if err != nil {
		t.Fatalf("Could not create temp dir: %s", err)
	}
	tests := []struct {
		Title, Description string
		Flags              masterTmplFlags
		Master             string
		Content            string
		Rendered           string
	}{
		{"Foo", "Bar!", 0, "{{.Page.Title}}|{{.Page.Description}}|{{.Page.Content}}",
			"The content.", "Foo|Bar!|The content."},
		{}}
	for i, v := range tests {
		session := client.Session{
			User: &client.User{Login: "admin", Name: "Administrator"}}
		node := client.Node{Title: v.Title, Description: v.Description}
		env := masterTmplEnv{node, &session, v.Title, v.Description, 0}
		renderer := template.Renderer{Root: root}
		if err = ioutil.WriteFile(filepath.Join(root, "master.html"),
			[]byte(v.Master), 0600); err != nil {
			t.Fatalf("Could not write master template: %s", err)
		}
		ret := renderInMaster(renderer, []byte(v.Content), env, settings{},
			site{}, "")
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
