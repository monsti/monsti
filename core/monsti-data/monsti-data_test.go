// Copyright 2012-2013 Christian Neumann
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"path/filepath"
	utesting "pkg.monsti.org/monsti/api/util/testing"
)

func TestGetNode(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{
		"/foo/node.json": `{"Type":"core.Foo"}`},
		"TestGetNode")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	ret, err := getNode(root, "/foo")
	expected := `{"Path":"/foo","Type":"core.Foo"}`
	if err != nil {
		t.Errorf("Got error: %v", err)
	} else if string(ret) != expected {
		t.Fatalf(`getNode(%q, "/foo") = %v, nil, should be %v, nil`,
			root, string(ret), expected)
	}
	ret, err = getNode(root, "/unavailable")
	if err != nil {
		t.Errorf("Got error: %v", err)
	} else if ret != nil {
		t.Errorf(`getNode(%q, "/unavailable") = %v, nil, should be nil, nil`,
			root, string(ret))
	}
}

func TestGetChildren(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{
		"/foo/node.json": `{"title":"Node Foo","shorttitle":"Foo"}`,
		"/foo/child1/node.json": `{"title":"Node a Foo Child 1",` +
			`"shorttitle":"a Foo Child 1"}`,
		"/foo/child2/node.json": `{"title":"Node Foo Child 2",` +
			`"shorttitle":"Foo Child 2"}`,
		"/foo/child2/child1/node.json": `{"title":"Node a Foo Child 2 Child 1",` +
			`"shorttitle":"a Foo Child 2 Child 1"}`,
		"/bar/node.json": `{"title":"Node Bar","order":"2",` +
			`"shorttitle":"Bar"}`}, "TestGetChildren")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	err = os.Symlink("child2", filepath.Join(root, "/foo/child3"))
	if err != nil {
		t.Fatalf("Could not create symlink: %v", err)
	}
	tests := []struct {
		Path     string
		Children []string
	}{
		{"/foo", []string{"/foo/child1", "/foo/child2", "/foo/child3"}},
		{"/bar", []string{}}}
	for _, test := range tests {
		ret, err := getChildren(root, test.Path)
		if err != nil {
			t.Errorf(`getChildren(%q, %q) = %v, %v, should be _, nil`,
				root, test.Path, ret, err)
		}
		if len(ret) != len(test.Children) {
			t.Errorf(`getChildren(%q) returns %v items (%v), should be %v`,
				test.Path, len(ret), ret, len(test.Children))
		}
		for i := range ret {
			if strings.Contains(test.Children[i], string(ret[i])) {
				t.Errorf(`Item %v of getChildren(%q) is %v, should be %v`,
					i, test.Path, ret[i], test.Children[i])
			}
		}
	}
}

func TestGetConfig(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{
		"/foo.json": `{"foo":{"foobar":"foobarvalue"},"bar":"barvalue"}`,
	}, "TestGetSection")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	tests := []struct{ Name, Value string }{
		{"", `{"Value":{"foo":{"foobar":"foobarvalue"},"bar":"barvalue"}}`},
		{"foo", `{"Value":{"foobar":"foobarvalue"}}`},
		{"foo.foobar", `{"Value":"foobarvalue"}`},
		{"bar", `{"Value":"barvalue"}`},
		{"unknown", `{"Value": null}`},
	}
	for _, test := range tests {
		unmarshal := func(in []byte) (out interface{}) {
			if err = json.Unmarshal(in, &out); err != nil {
				t.Errorf("Could not unmarshal for %v: %v", test.Name, err)
			}
			return
		}
		ret, err := getConfig(filepath.Join(root, "foo.json"), test.Name)
		switch {
		case err != nil:
			t.Errorf("getConfig(_, %q) returned error: %v", test.Name, err)
		case !reflect.DeepEqual(unmarshal(ret), unmarshal([]byte(test.Value))):
			t.Errorf("getConfig(_, %q) = `%s`, _ should be `%s`", test.Name, ret,
				test.Value)
		}
	}
}
