package main

import (
	utesting "github.com/monsti/util/testing"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetNav(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{
		"/navigation.yaml": `
- name: foo Page
  target: foo
- name: bar Page
  target: bar`,
		"/foo/__empty__":      "",
		"/bar/cruz/__empty__": "",
		"/bar/navigation.yaml": `
- name: Absolute
  target: /absolute
- name: Cruz
  target: cruz`}, "TestGetNav")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	tests := []struct {
		Path, Active string
		Recursive    bool
		Expected     navigation
		Root         string
	}{
		{"/", "", false, navigation{
			{Name: "foo Page", Target: "foo"},
			{Name: "bar Page", Target: "bar"}}, ""},
		{"/", "foo", false, navigation{
			{Name: "foo Page", Target: "foo", Active: true},
			{Name: "bar Page", Target: "bar"}}, ""},
		{"/foo", "foo", false, nil, ""},
		{"/bar", "bar", false, navigation{
			{Name: "Absolute", Target: "/absolute", Active: false},
			{Name: "Cruz", Target: "cruz", Active: false}}, ""},
		{"/bar", "bar", true, navigation{
			{Name: "Absolute", Target: "/absolute", Active: false},
			{Name: "Cruz", Target: "cruz", Active: false}}, "/bar"},
		{"/foo", "foo", true, nil, ""},
		{"/bar/cruz", "cruz", true, navigation{
			{Name: "Absolute", Target: "/absolute", Active: false},
			{Name: "Cruz", Target: "cruz", Active: true}}, "/bar"},
		{"/", "", false, navigation{
			{Name: "foo Page", Target: "foo"},
			{Name: "bar Page", Target: "bar"}}, ""}}
	for _, test := range tests {
		ret, retRoot := getNav(test.Path, test.Active, test.Recursive, root)
		if !reflect.DeepEqual(ret, test.Expected) || retRoot != test.Root {
			t.Errorf(`getNav(%q, %q, %v, _) = %v, %v, should be %v, %v`,
				test.Path, test.Active, test.Recursive, ret, retRoot,
				test.Expected, test.Root)
		}
	}
}

func TestNavigationDump(t *testing.T) {
	nav := navigation{
		{Name: "foo Page", Target: "foo", Active: true},
		{Name: "bar Page", Target: "bar", Active: false}}
	root, err := ioutil.TempDir("", "_monsti_TestNavigationDump")
	if err != nil {
		t.Fatalf("Could not create temp dir: %s", err)
	}
	defer os.RemoveAll(root)
	err = os.Mkdir(filepath.Join(root, "foonode"), 0700)
	if err != nil {
		t.Fatalf("Could not create node directory: %s", err)
	}
	nav.Dump("/foonode", root)
	contents, err := ioutil.ReadFile(filepath.Join(root, "foonode",
		"navigation.yaml"))
	expected := `- name: foo Page
  target: foo
- name: bar Page
  target: bar
`
	if string(contents) != expected {
		t.Errorf(`dumpNav("%s", _, _) = %q, should be %q`,
			nav, contents, expected)
	}
}

func TestNavigationAdd(t *testing.T) {
	nav := navigation{
		{Name: "foo Page", Target: "foo"},
		{Name: "bar Page", Target: "bar"}}
	nav.Add("cruz Page", "cruz")
	link := navLink{Name: "cruz Page", Target: "cruz"}
	if nav[2] != link {
		t.Errorf(`navigation.Add("cruz Page", "cruz") = %q, this is wrong.`,
			nav)
	}
}

func TestNavigationRemove(t *testing.T) {
	nav := navigation{
		{Name: "foo Page", Target: "foo"},
		{Name: "bar Page", Target: "bar"},
		{Name: "again foo Page", Target: "foo"}}
	nav.Remove("foo")
	expected := navigation{{Name: "bar Page", Target: "bar"}}
	if !reflect.DeepEqual(nav, expected) {
		t.Errorf(`navigation.Remove("foo") = %q, should be %q`,
			nav, expected)
	}
}

func TestNavigationMakeAbsolute(t *testing.T) {
	nav := navigation{
		{Target: "foo"},
		{Target: "/bar"}}
	nav.MakeAbsolute("/root")
	expected := navigation{
		{Target: "/root/foo"},
		{Target: "/bar"}}
	if !reflect.DeepEqual(nav, expected) {
		t.Errorf(`navigation.MakeAbsolute("/root") = %q, should be %q`,
			nav, expected)
	}
}

func TestRemoveNode(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{
		"/navigation.yaml": `
- name: foo Page
  target: foo
- name: bar Page
  target: bar`,
		"/foo/__empty__":       "",
		"/bar/navigation.yaml": ""}, "TestRemoveNode")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	removeNode("/foo", root)
	if f, err := os.Open(filepath.Join(root, "foo")); !os.IsNotExist(err) {
		f.Close()
		t.Errorf(`/foo does still exist, should be removed`)
	}
	nav, _ := getNav("/", "", false, root)
	expectedNav := navigation{
		{Name: "bar Page", Target: "bar"}}
	if !reflect.DeepEqual(nav, expectedNav) {
		t.Errorf(`Navigation should be %v after removal, but is %v.`,
			expectedNav, nav)
	}
}
