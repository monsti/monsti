// This file is part of Monsti, a web content management system.
// Copyright 2012-2014 Christian Neumann
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
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"path/filepath"
	"pkg.monsti.org/monsti/api/service"
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
	ret, err := getConfig(filepath.Join(root, "nonexisting.json"), "foo")
	if err != nil || ret != nil {
		t.Errorf("getConfig for non existing config file should"+
			"return nil,nil, got %v,%v", ret, err)
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

func TestFindAddableNodeTypes(t *testing.T) {
	tests := []struct {
		NodeTypes map[string]*service.NodeType
		NodeType  string
		Expected  []string
	}{
		{
			NodeTypes: map[string]*service.NodeType{
				"A": &service.NodeType{Id: "A", AddableTo: []string{"B"}},
				"B": &service.NodeType{Id: "B", AddableTo: []string{}},
				"C": &service.NodeType{Id: "C", AddableTo: []string{"B"}},
			},
			NodeType: "B",
			Expected: []string{"A", "C"},
		},
		{
			NodeTypes: map[string]*service.NodeType{
				"Foo.A": &service.NodeType{Id: "Foo.A", AddableTo: nil},
				"Foo.B": &service.NodeType{Id: "Foo.B", AddableTo: []string{"Foo.B"}},
				"Foo.C": &service.NodeType{Id: "Foo.C", AddableTo: []string{}},
				"Foo.D": &service.NodeType{Id: "Foo.D", AddableTo: []string{"."}},
				"Foo.E": &service.NodeType{Id: "Foo.E", AddableTo: []string{"Foo."}},
			},
			NodeType: "Foo.B",
			Expected: []string{"Foo.B", "Foo.D", "Foo.E"},
		},
	}
	for i, test := range tests {
		ret := findAddableNodeTypes(test.NodeType, test.NodeTypes)
		for _, retType := range test.Expected {
			found := false
			for _, expectedType := range ret {
				if retType == expectedType {
					found = true
				}
			}
			if !found || len(ret) != len(test.Expected) {
				t.Fatalf("findAddableNodeTypes#%v returned %v, expected %v",
					i, ret, test.Expected)
			}
		}
	}
}

func TestCache(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{}, "TestCache")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	err = toCache(root, "/foo/bar/cruz", "foo.some_cache", []byte("test"), nil)
	if err != nil {
		t.Fatalf("Could not cache data: %v", err)
	}
	err = toCache(root, "/foo/bar", "foo.another_cache", []byte("test2"),
		&service.CacheMods{Deps: []service.CacheDep{{Node: "/foo/bar/cruz"}}})
	if err != nil {
		t.Fatalf("Could not cache data: %v", err)
	}
	err = toCache(root, "/foo", "foo.another_cache", []byte("test3"),
		&service.CacheMods{Deps: []service.CacheDep{{Node: "/foo/bar/cruz"}}})
	if err != nil {
		t.Fatalf("Could not cache data: %v", err)
	}
	ret, err := fromCache(root, "/foo", "foo.another_cache")
	if err != nil {
		t.Fatalf("Could not get cached data: %v", err)
	}
	if !reflect.DeepEqual(ret, []byte("test3")) {
		t.Fatalf("test3 should be in cache, got %v", string(ret))
	}
	err = markDep(root, service.CacheDep{Node: "/foo/bar/cruz"}, 0)
	if err != nil {
		t.Fatalf("Could not mark dep: %v", err)
	}
	ret, err = fromCache(root, "/foo", "foo.another_cache")
	if err != nil {
		t.Fatalf("Could not get cached data: %v", err)
	}
	if ret != nil {
		t.Fatalf("Cache should be nil, got %v", string(ret))
	}
}

func TestCacheMarkDescend(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{}, "TestCache")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	err = toCache(root, "/foo/bar/cruz", "foo.some_cache", []byte("test"), nil)
	if err != nil {
		t.Fatalf("Could not cache data: %v", err)
	}
	err = toCache(root, "/foo/bar", "foo.some_cache", []byte("test2"), nil)
	if err != nil {
		t.Fatalf("Could not cache data: %v", err)
	}

	// Descend one level
	var ret []byte
	err = toCache(root, "/foo", "foo.another_cache", []byte("test3"),
		&service.CacheMods{Deps: []service.CacheDep{{Node: "/foo", Descend: 1}}})
	if err != nil {
		t.Fatalf("Could not cache data: %v", err)
	}
	err = markDep(root, service.CacheDep{Node: "/foo/bar/cruz"}, 0)
	if err != nil {
		t.Fatalf("Could not mark dep: %v", err)
	}
	ret, err = fromCache(root, "/foo", "foo.another_cache")
	if err != nil {
		t.Fatalf("Could not get cached data: %v", err)
	}
	if ret == nil {
		t.Errorf("Cache should not be nil")
	}
	err = markDep(root, service.CacheDep{Node: "/foo/bar"}, 0)
	if err != nil {
		t.Fatalf("Could not mark dep: %v", err)
	}
	ret, err = fromCache(root, "/foo", "foo.another_cache")
	if err != nil {
		t.Fatalf("Could not get cached data: %v", err)
	}
	if ret != nil {
		t.Errorf("Cache should be nil, got %v", string(ret))
	}

	// Descend all levels
	err = toCache(root, "/foo", "foo.another_cache", []byte("test3"),
		&service.CacheMods{Deps: []service.CacheDep{{Node: "/foo", Descend: -1}}})
	if err != nil {
		t.Fatalf("Could not cache data: %v", err)
	}
	err = markDep(root, service.CacheDep{Node: "/foo/bar/cruz"}, 0)
	if err != nil {
		t.Fatalf("Could not mark dep: %v", err)
	}
	ret, err = fromCache(root, "/foo", "foo.another_cache")
	if err != nil {
		t.Fatalf("Could not get cached data: %v", err)
	}
	if ret != nil {
		t.Errorf("Cache should be nil, got %v", string(ret))
	}
}

func TestCacheExpire(t *testing.T) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{}, "TestCache")
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	defer cleanup()
	err = toCache(root, "/foo", "foo.foo", []byte("test"),
		&service.CacheMods{Expire: time.Now().AddDate(-1, 0, 0)})
	if err != nil {
		t.Fatalf("Could not cache data: %v", err)
	}
	ret, err := fromCache(root, "/foo", "foo.foo")
	if err != nil {
		t.Fatalf("Could not get cached data: %v", err)
	}
	if ret != nil {
		t.Errorf("Cache should have been expired.")
	}
}
