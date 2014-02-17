// Package testing contains utility/convenience functions usable for testing.
package testing

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// CreateDirectoryTree creates a temporary directory tree containing the given
// files and having the given prefix.
//
// Returns the path to the directory and a cleanup function which removes all
// files. If the tree could not be generated, the resulting error is set and
// the tree (if any) will be removed.
//
// The cleanup will panic if any error occurs.
//
//     files := map[string]string{
//       "/foo/bar.html": "<b>Hello World</b>",
//       "/bar/foo/foo.txt": "Hey World."
//     }
//     root, cleanup, err := CreateDirectoryTree(files, "TestDoSomethingFunc")
//     if err != nil {
//       panic("Could not create directory tree.")
//     }
//     defer cleanup()
func CreateDirectoryTree(files map[string]string, prefix string) (string,
	func(), error) {

	root, err := ioutil.TempDir("", "_monsti_"+prefix)
	if err != nil {
		return "", nil, fmt.Errorf("Could not create temp dir: %v", err)
	}
	cleanup := func() {
		if err := os.RemoveAll(root); err != nil {
			panic(fmt.Sprint("Could not clean up: ", err))
		}
	}
	for path, content := range files {
		if err = os.MkdirAll(filepath.Join(root, filepath.Dir(path)), 0700); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("Could not create directory: %v", err)
		}
		if err = ioutil.WriteFile(filepath.Join(root, path), []byte(content),
			0600); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("Could not write file: %v", err)
		}
	}
	return root, cleanup, nil
}
