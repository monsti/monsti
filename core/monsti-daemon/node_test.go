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
		Node     service.Node
		Children []string
	}{
		"/": {
			Children: []string{"foo", "bar", "hideme", "cruz"}},
		"/foo": {
			Children: []string{"child1", "child2"}},
		"/foo/child1": {
			Children: []string{}},
		"/foo/child2": {
			Children: []string{"child1"}},
		"/foo/child2/child1": {
			Children: []string{}},
		"/bar": {
			Node:     service.Node{Order: 2},
			Children: []string{}},
		"/hideme": {
			Node:     service.Node{Hide: true},
			Children: []string{}},
		"/cruz": {
			Node:     service.Node{Order: -2},
			Children: []string{"child1"}},
		"/cruz/child1": {
			Children: []string{}}}
	getNodeFn := func(nodePath string) (*service.Node, error) {
		if val, ok := nodes[nodePath]; ok {
			val.Node.Path = nodePath
			val.Node.Type = new(service.NodeType)
			val.Node.Public = true
			return &val.Node, nil
		} else {
			t.Fatalf("Could not find node %q", nodePath)
		}
		return nil, nil
	}
	getChildrenFn := func(nodePath string) ([]*service.Node, error) {
		children := make([]*service.Node, 0)
		for _, child := range nodes[nodePath].Children {
			node, _ := getNodeFn(path.Join(nodePath, child))
			children = append(children, node)
		}
		return children, nil
	}
	tests := []struct {
		Path, Active string
		Public       bool
		Expected     navigation
	}{
		{"/", "/", true, navigation{
			{Target: "/", Child: false, Active: true},
			{Target: "/cruz", Child: true, Order: -2},
			{Target: "/foo", Child: true},
			{Target: "/bar", Child: true, Order: 2}}},
		{"/", "/foo/child2/child1", true, navigation{
			{Target: "/", Child: false, Active: false, ActiveBelow: true},
			{Target: "/cruz", Child: true, Order: -2},
			{Target: "/foo", Child: true, ActiveBelow: true},
			{Target: "/bar", Child: true, Order: 2}}},
		{"/foo", "/foo", true, navigation{
			{Target: "/foo", Active: true},
			{Target: "/foo/child1", Child: true},
			{Target: "/foo/child2", Child: true}}},
		{"/foo/child1", "/foo/child1", true, navigation{
			{Target: "/foo", Active: false, ActiveBelow: true},
			{Target: "/foo/child1", Child: true, Active: true},
			{Target: "/foo/child2", Child: true}}},
		{"/foo/child2", "/foo/child2", true, navigation{
			{Target: "/foo/child1"},
			{Target: "/foo/child2", Active: true},
			{Target: "/foo/child2/child1", Child: true}}},
		{"/foo/child2/child1", "/foo/child2/child1", true, navigation{
			{Target: "/foo/child1"},
			{Target: "/foo/child2", Active: false, ActiveBelow: true},
			{Target: "/foo/child2/child1", Active: true, Child: true}}},
		{"/bar", "/bar", true, navigation{}},
		{"/cruz", "/cruz", true, navigation{
			{Target: "/cruz", Active: true, Order: -2},
			{Target: "/cruz/child1", Child: true}}}}
	for _, test := range tests {
		for i, _ := range test.Expected {
			test.Expected[i].Name = "Untitled"
		}
		ret, err := getNav(test.Path, test.Active, test.Public, getNodeFn, getChildrenFn)
		if err != nil || !(len(ret) == 0 && len(test.Expected) == 0 || reflect.DeepEqual(ret, test.Expected)) {
			t.Errorf("getNav(%q, %q, _) is\n%v, %v\nshould be\n%v, nil",
				test.Path, test.Active, ret, err, test.Expected)
		}
	}
}

func TestNavigationMakeAbsolute(t *testing.T) {
	nav := navigation{
		{Target: "/"},
		{Target: "foo"},
		{Target: "."},
		{Target: "../bar"}}
	nav.MakeAbsolute("/root")
	expected := navigation{
		{Target: "/"},
		{Target: "/root/foo/"},
		{Target: "/root/"},
		{Target: "/bar/"}}
	if !(reflect.DeepEqual(nav, expected)) {
		t.Errorf(`navigation.MakeAbsolute("/root") = %v, should be %v`,
			nav, expected)
	}
}
