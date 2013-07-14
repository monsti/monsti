// This file is part of Monsti, a web content management system.
// Copyright 2012-2013 Christian Neumann
// 
// Monsti is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// Monsti is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Monsti.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	utesting "pkg.monsti.org/util/testing"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetNav(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{
		"/foo/node.yaml": "title: Node Foo\nshorttitle: Foo",
		"/foo/child1/node.yaml": "title: Node a Foo Child 1\n" +
			"shorttitle: a Foo Child 1",
		"/foo/child2/node.yaml":        "title: Node Foo Child 2\n" +
			"shorttitle: Foo Child 2",
		"/foo/child2/child1/node.yaml": "title: Node a Foo Child 2 Child 1\n" +
			"shorttitle: a Foo Child 2 Child 1",
		"/bar/node.yaml":               "title: Node Bar\norder: 2\n" +
			"shorttitle: Bar",
		"/hideme/node.yaml":            "title: Node Hide me!\nhide: true\n" +
			"shorttitle: Hide me!",
		"/cruz/node.yaml":              "title: Node Cruz\norder: -2\n" +
			"shorttitle: Cruz",
		"/cruz/child1/node.yaml":       "title: Node Cruz Child 1\n" +
			"shorttitle: Cruz Child 1"}, "TestGetNav")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	tests := []struct {
		Path, Active string
		Expected     navigation
	}{
		{"/", "/", navigation{
			{Name: "Cruz", Target: "cruz", Child: true, Order: -2},
			{Name: "Foo", Target: "foo", Child: true},
			{Name: "Bar", Target: "bar", Child: true, Order: 2}}},
		{"/", "/foo/child1/child2", navigation{
			{Name: "Cruz", Target: "../../../cruz", Child: true, Order: -2},
			{Name: "Foo", Target: "../..", Child: true, Active: true},
			{Name: "Bar", Target: "../../../bar", Child: true, Order: 2}}},
		{"/foo", "/foo", navigation{
			{Name: "Foo", Target: ".", Active: true},
			{Name: "Foo Child 2", Target: "child2", Child: true},
			{Name: "a Foo Child 1", Target: "child1", Child: true}}},
		{"/foo/child1", "/foo/child1", navigation{
			{Name: "Foo", Target: ".."},
			{Name: "Foo Child 2", Target: "../child2", Child: true},
			{Name: "a Foo Child 1", Target: ".", Child: true, Active: true}}},
		{"/foo/child2", "/foo/child2", navigation{
			{Name: "Foo Child 2", Target: ".", Active: true},
			{Name: "a Foo Child 2 Child 1", Target: "child1", Child: true},
			{Name: "a Foo Child 1", Target: "../child1"}}},
		{"/foo/child2/child1", "/foo/child2/child1", navigation{
			{Name: "Foo Child 2", Target: ".."},
			{Name: "a Foo Child 2 Child 1", Target: ".", Active: true, Child: true},
			{Name: "a Foo Child 1", Target: "../../child1"}}},
		{"/bar", "/bar", navigation{}},
		{"/cruz", "/cruz", navigation{
			{Name: "Cruz", Target: ".", Active: true, Order: -2},
			{Name: "Cruz Child 1", Target: "child1", Child: true}}}}
	for _, test := range tests {
		ret, err := getNav(test.Path, test.Active, root)
		if err != nil || !(len(ret) == 0 && len(test.Expected) == 0 || reflect.DeepEqual(ret, test.Expected)) {
			t.Errorf(`getNav(%q, %q, _) = %v, %v, should be %v, nil`,
				test.Path, test.Active, ret, err, test.Expected)
		}
	}
}

func TestNavigationMakeAbsolute(t *testing.T) {
	nav := navigation{
		{Target: "foo"},
		{Target: "."},
		{Target: "../bar"}}
	nav.MakeAbsolute("/root")
	expected := navigation{
		{Target: "/root/foo/"},
		{Target: "/root/"},
		{Target: "/bar/"}}
	if !(reflect.DeepEqual(nav, expected)) {
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
}
