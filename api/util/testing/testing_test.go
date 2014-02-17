package testing

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDirectoryTree(t *testing.T) {
	files := map[string]string{
		"/foo.dat":         "one",
		"/foo/bar.dat":     "two",
		"/bar/foo/bar.dat": "three"}
	root, cleanup, err := CreateDirectoryTree(files, "TestCreateDirectoryTree")
	for path, content := range files {
		ret, err := ioutil.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Errorf("Could not read %q: %v", path, err)
			continue
		}
		if string(ret) != content {
			t.Errorf("Content mismatch for %q: %q vs expected %q", path,
				string(ret), content)
		}
	}
	cleanup()
	_, err = os.Open(root)
	if !os.IsNotExist(err) {
		t.Errorf("Cleanup did not remove the directory tree.")
	}
}
