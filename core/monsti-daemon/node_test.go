// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann
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
		Depth        int
		Public       bool
		Expected     navigation
	}{
		{"/", "/", 1, true, navigation{
			{Target: "/", Active: true},
			{Target: "/cruz", Level: 1, Order: -2},
			{Target: "/foo", Level: 1},
			{Target: "/bar", Level: 1, Order: 2}}},
		{"/", "/foo/child2/child1", 1, true, navigation{
			{Target: "/", Active: false, ActiveBelow: true},
			{Target: "/cruz", Level: 1, Order: -2},
			{Target: "/foo", Level: 1, ActiveBelow: true},
			{Target: "/bar", Level: 1, Order: 2}}},
		{"/foo", "/foo", 1, true, navigation{
			{Target: "/foo", Active: true, Children: navigation{
				{Target: "/foo/child1", Level: 1},
				{Target: "/foo/child2", Level: 1}}}}},
		{"/foo", "/foo", 2, true, navigation{
			{Target: "/foo", Active: true, Children: navigation{
				{Target: "/foo/child1", Level: 1},
				{Target: "/foo/child2", Level: 1, Children: navigation{
					{Target: "/foo/child2/child1", Level: 2}}}}}}},
		{"/foo/child1", "/foo/child1", 1, true, navigation{
			{Target: "/foo", Active: false, ActiveBelow: true, Children: navigation{
				{Target: "/foo/child1", Level: 1, Active: true},
				{Target: "/foo/child2", Level: 1}}}}},
		{"/foo/child2", "/foo/child2", 1, true, navigation{
			/*			{Target: "/foo/child1"},*/
			{Target: "/foo/child2", Active: true, Children: navigation{
				{Target: "/foo/child2/child1", Level: 1}}}}},
		{"/foo/child2/child1", "/foo/child2/child1", 1, true, navigation{
			/*			{Target: "/foo/child1"},*/
			{Target: "/foo/child2", Active: false, ActiveBelow: true, Children: navigation{
				{Target: "/foo/child2/child1", Active: true, Level: 1}}}}},
		{"/bar", "/bar", 1, true, navigation{}},
		{"/cruz", "/cruz", 1, true, navigation{
			{Target: "/cruz", Active: true, Order: -2, Children: navigation{
				{Target: "/cruz/child1", Level: 1}}}}}}
	for _, test := range tests {
		var makeTitle func(nav *navigation)
		makeTitle = func(nav *navigation) {
			for i, _ := range *nav {
				(*nav)[i].Name = "Untitled"
				makeTitle(&((*nav)[i].Children))
			}
		}
		makeTitle(&test.Expected)
		ret, err := getNav(test.Path, test.Active, test.Public, getNodeFn,
			getChildrenFn, test.Depth)
		if err != nil || !(len(ret) == 0 && len(test.Expected) == 0 ||
			reflect.DeepEqual(ret, test.Expected)) {
			t.Errorf("getNav(%q, %q, _) is\n%v, %v\nshould be\n%v, <nil> %v",
				test.Path, test.Active, ret, err, test.Expected, reflect.DeepEqual(ret, test.Expected))
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

func TestCalcEmbedPath(t *testing.T) {
	tests := []struct {
		Path, URI, Expected string
	}{
		{"/", "foo", "/foo"},
		{"/", "/foo", "/foo"},
		{"/foo/bar", "/foo", "/foo"},
		{"/foo/bar", "foo", "/foo/bar/foo"},
		{"/foo/bar/", "/foo", "/foo"},
		{"/foo/bar/", "foo", "/foo/bar/foo"},
	}
	for _, test := range tests {
		ret, err := calcEmbedPath(test.Path, test.URI)
		if err != nil || ret != test.Expected {
			t.Errorf(`calcEmbedPath(%q,%q) = (%q,%v), should be %q, nil`,
				test.Path, test.URI, ret, err, test.Expected)
		}
	}
}
