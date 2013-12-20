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
	"testing"

	"pkg.monsti.org/service"
	utesting "pkg.monsti.org/util/testing"
)

func TestGetNode(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{
		"/foo/node.yaml": "title: Node Foo\nshorttitle: Foo"},
		"TestGetNode")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	ret, err := getNode(root, "/foo")
	expected := service.NodeInfo{
		Path:       "/foo",
		Title:      "Node Foo",
		ShortTitle: "Foo"}
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	if *ret != expected {
		t.Fatalf(`getNode(%q, "/foo") = %v, nil, should be %v, nil`,
			root, ret, expected)
	}

}

func TestGetChildren(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{
		"/foo/node.yaml": "title: Node Foo\nshorttitle: Foo",
		"/foo/child1/node.yaml": "title: Node a Foo Child 1\n" +
			"shorttitle: a Foo Child 1",
		"/foo/child2/node.yaml": "title: Node Foo Child 2\n" +
			"shorttitle: Foo Child 2",
		"/foo/child2/child1/node.yaml": "title: Node a Foo Child 2 Child 1\n" +
			"shorttitle: a Foo Child 2 Child 1",
		"/bar/node.yaml": "title: Node Bar\norder: 2\n" +
			"shorttitle: Bar"}, "TestGetChildren")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	tests := []struct {
		Path     string
		Children []string
	}{
		{"/foo", []string{"/foo/child1", "/foo/child2"}},
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
			if ret[i].Path != test.Children[i] {
				t.Errorf(`Item %v of getChildren(%q) is %v, should be %v`,
					i, test.Path, ret[i].Path, test.Children[i])
			}
		}
	}
}
