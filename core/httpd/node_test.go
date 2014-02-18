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
	"path"
	"reflect"
	"testing"

	"pkg.monsti.org/monsti/api/service"
)

func TestGetNav(t *testing.T) {
	nodes := map[string]struct {
		Node     service.NodeFields
		Children []string
	}{
		"/": {service.NodeFields{Title: "Root"},
			[]string{"foo", "bar", "hideme", "cruz"}},
		"/foo": {service.NodeFields{Title: "Node Foo", ShortTitle: "Foo"},
			[]string{"child1", "child2"}},
		"/foo/child1": {
			service.NodeFields{Title: "Node Foo Child 1",
				ShortTitle: "Foo Child 1"}, []string{}},
		"/foo/child2": {
			service.NodeFields{Title: "Node Foo Child 2",
				ShortTitle: "Foo Child 2"}, []string{"child1"}},
		"/foo/child2/child1": {
			service.NodeFields{Title: "Node Foo Child 2 Child 1",
				ShortTitle: "Foo Child 2 Child 1"}, []string{}},
		"/bar": {
			service.NodeFields{Title: "Node Bar", Order: 2,
				ShortTitle: "Bar"}, []string{}},
		"/hideme": {
			service.NodeFields{Title: "Node Hide me!", Hide: true,
				ShortTitle: "Hide me!"}, []string{}},
		"/cruz": {
			service.NodeFields{Title: "Node Cruz", Order: -2,
				ShortTitle: "Cruz"}, []string{"child1"}},
		"/cruz/child1": {
			service.NodeFields{Title: "Node Cruz Child 1",
				ShortTitle: "Cruz Child 1"}, []string{}}}
	getNodeFn := func(nodePath string) (*service.NodeFields, error) {
		if val, ok := nodes[nodePath]; ok {
			val.Node.Path = nodePath
			return &val.Node, nil
		} else {
			t.Fatalf("Could not find node %q", nodePath)
		}
		return nil, nil
	}
	getChildrenFn := func(nodePath string) ([]*service.NodeFields, error) {
		children := make([]*service.NodeFields, 0)
		for _, child := range nodes[nodePath].Children {
			node, _ := getNodeFn(path.Join(nodePath, child))
			children = append(children, node)
		}
		return children, nil
	}
	tests := []struct {
		Path, Active string
		Expected     navigation
	}{
		{"/", "/", navigation{
			{Name: "Cruz", Target: "/cruz", Child: true, Order: -2},
			{Name: "Foo", Target: "/foo", Child: true},
			{Name: "Bar", Target: "/bar", Child: true, Order: 2}}},
		{"/", "/foo/child1/child2", navigation{
			{Name: "Cruz", Target: "/cruz", Child: true, Order: -2},
			{Name: "Foo", Target: "/foo", Child: true, Active: true},
			{Name: "Bar", Target: "/bar", Child: true, Order: 2}}},
		{"/foo", "/foo", navigation{
			{Name: "Foo", Target: "/foo", Active: true},
			{Name: "Foo Child 1", Target: "/foo/child1", Child: true},
			{Name: "Foo Child 2", Target: "/foo/child2", Child: true}}},
		{"/foo/child1", "/foo/child1", navigation{
			{Name: "Foo", Target: "/foo"},
			{Name: "Foo Child 1", Target: "/foo/child1", Child: true, Active: true},
			{Name: "Foo Child 2", Target: "/foo/child2", Child: true}}},
		{"/foo/child2", "/foo/child2", navigation{
			{Name: "Foo Child 1", Target: "/foo/child1"},
			{Name: "Foo Child 2", Target: "/foo/child2", Active: true},
			{Name: "Foo Child 2 Child 1", Target: "/foo/child2/child1", Child: true}}},
		{"/foo/child2/child1", "/foo/child2/child1", navigation{
			{Name: "Foo Child 1", Target: "/foo/child1"},
			{Name: "Foo Child 2", Target: "/foo/child2"},
			{Name: "Foo Child 2 Child 1", Target: "/foo/child2/child1", Active: true, Child: true}}},
		{"/bar", "/bar", navigation{}},
		{"/cruz", "/cruz", navigation{
			{Name: "Cruz", Target: "/cruz", Active: true, Order: -2},
			{Name: "Cruz Child 1", Target: "/cruz/child1", Child: true}}}}
	for _, test := range tests {
		ret, err := getNav(test.Path, test.Active, getNodeFn, getChildrenFn)
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
