package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetNav(t *testing.T) {
	root, err := ioutil.TempDir("", "_monsti_TestGetNav")
	if err != nil {
		t.Fatalf("Could not create temp dir: %s", err)
	}
	defer os.RemoveAll(root)
	nav := []byte(`- name: foo Page
  target: foo
- name: bar Page
  target: bar
`)
	if err = ioutil.WriteFile(filepath.Join(root, "navigation.yaml"),
		nav, 0600); err != nil {
		t.Fatalf("Could not write navigation: %s", err)
	}

	if err = os.Mkdir(filepath.Join(root, "foo"), 0700); err != nil {
		t.Fatalf("Could not create node directory: %s", err)
	}
	if err = os.Mkdir(filepath.Join(root, "bar"), 0700); err != nil {
		t.Fatalf("Could not create node directory: %s", err)
	}
	if err = ioutil.WriteFile(filepath.Join(root, "bar", "navigation.yaml"),
		[]byte(""), 0600); err != nil {
		t.Fatalf("Could not write navigation: %s", err)
	}
	tests := []struct {
		Path, Active string
		Recursive    bool
		Expected     navigation
	}{
		{"/", "", false, navigation{
			{Name: "foo Page", Target: "foo"},
			{Name: "bar Page", Target: "bar"}}},
		{"/", "foo", false, navigation{
			{Name: "foo Page", Target: "foo", Active: true},
			{Name: "bar Page", Target: "bar"}}},
		{"/foo", "foo", false, nil},
		{"/bar", "foo", false, navigation{}},
		{"/foo", "foo", true, navigation{
			{Name: "foo Page", Target: "foo", Active: true},
			{Name: "bar Page", Target: "bar"}}},
		{"/", "", false, navigation{
			{Name: "foo Page", Target: "foo"},
			{Name: "bar Page", Target: "bar"}}}}
	for _, test := range tests {
		ret := getNav(test.Path, test.Active, test.Recursive, root)
		if !reflect.DeepEqual(ret, test.Expected) {
			t.Errorf(`getNav(%q, %q, %v, _) = %v, should be %v`,
				test.Path, test.Active, test.Recursive, ret, test.Expected)
		}
	}
}

func TestDumpNav(t *testing.T) {
	nav := navigation{
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

func TestAddToNav(t *testing.T) {
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
