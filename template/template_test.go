package template

import (
	mtesting "pkg.monsti.org/util/testing"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestGetIncludes(t *testing.T) {
	root, cleanup, err := mtesting.CreateDirectoryTree(map[string]string{
		"/first/include":                        "one\ntwo\nfour\n\n",
		"/first/master.include":                 "three",
		"/first/foo/bar/include":                "four",
		"/first/foo/bar/cruz/include":           "five",
		"/first/foo/bar/cruz/template.include":  "six",
		"/second/include":                       "seven",
		"/second/foo/bar/cruz/template.include": "eight"}, "TestGetIncludes")
	if err != nil {
		t.Fatalf("Could not create test directory tree: %v", err)
	}
	defer cleanup()
	includes := getIncludes([]string{filepath.Join(root, "first"),
		filepath.Join(root, "second")},
		"foo/bar/cruz/template")
	sort.Strings(includes)
	expected := []string{
		"eight", "five", "four", "one", "seven", "six", "two"}
	if !reflect.DeepEqual(includes, expected) {
		t.Errorf("getIncludes returned: %v, should be %v", includes, expected)
	}
}
