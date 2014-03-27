package template

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	mtesting "pkg.monsti.org/monsti/api/util/testing"
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
	includes, err := getIncludes([]string{filepath.Join(root, "first"),
		filepath.Join(root, "second")},
		"foo/bar/cruz/template")
	sort.Strings(includes)
	expected := []string{
		"eight", "five", "four", "one", "seven", "six", "two"}
	if !reflect.DeepEqual(includes, expected) || err != nil {
		t.Errorf("getIncludes returned: %v, %v should be %v, nil",
			includes, err, expected)
	}
}
