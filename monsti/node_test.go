package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDumpNav(t *testing.T) {
	nav := []navLink{
		{Name: "foo Page", Target: "foo", Active: true},
		{Name: "bar Page", Target: "bar", Active: false}}
	root, err := ioutil.TempDir("", "_monsti_TestDumpNav")
	if err != nil {
		t.Fatalf("Could not create temp dir: %s", err)
	}
	defer os.RemoveAll(root)
	err = os.Mkdir(filepath.Join(root, "foonode"), 0700)
	if err != nil {
		t.Fatalf("Could not create node directory: %s", err)
	}
	dumpNav(nav, "/foonode", root)
	contents, err := ioutil.ReadFile(filepath.Join(root, "foonode",
		"navigation.yaml"))
	expected := `- name: foo Page
  target: foo
- name: bar Page
  target: bar
`
	if string(contents) != expected {
		t.Errorf(`dumpNav("%s", _, _) = "%q", should be "%q"`,
			nav, contents, expected)
	}
}
