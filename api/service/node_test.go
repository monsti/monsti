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

package service

import (
	"testing"
	"time"
)

func TestNodeName(t *testing.T) {
	tests := []struct{ path, name string }{
		{"", ""},
		{"/", ""},
		{"/foo", "foo"},
		{"/foo/", "foo"},
		{"/foo/bar", "bar"},
		{"/foo/bar/", "bar"},
	}
	for _, test := range tests {
		node := Node{Path: test.path}
		name := node.Name()
		if name != test.name {
			t.Errorf(`%v.Name() = %q, should be %q`, node, name, test.name)
		}
	}
}

func TestFields(t *testing.T) {
	fields := []Field{
		new(TextField),
		new(HTMLField),
		new(FileField),
	}
	for _, field := range fields {
		out := field.Dump()
		field.Load(out)
		out2 := field.Dump()
		if out != out2 {
			t.Errorf("Dump/Load/Dump: %q != %q", out, out2)
		}
	}
}

func TestGetParent(t *testing.T) {
	tests := []struct {
		Path, Prefix, Parent string
	}{
		{"/", "", "/"},
		{"/foo", "", "/"},
		{"/foo/bar", "", "/foo"},
		{"/foo/bar/cruz", "bar", "/foo"},
		{"/foo/bar/cruz", "foo/bar", "/"},
	}
	for _, test := range tests {
		node := Node{
			Path: test.Path,
			Type: &NodeType{
				PathPrefix: test.Prefix,
			},
		}
		ret := node.GetParentPath()
		if ret != test.Parent {
			t.Errorf("GetParentPath for path %v, prefix %v should be %v, got %v",
				test.Path, test.Prefix, test.Parent, ret)
		}
	}
}

func TestGetPathPrefix(t *testing.T) {
	testTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	tests := []struct {
		Prefix string
		Node   Node
		Ret    string
	}{
		{"", Node{}, ""},
		{"$year", Node{PublishTime: testTime}, "2009"},
		{"$year/$month", Node{PublishTime: testTime}, "2009/11"},
		{"$year/$month/$day", Node{PublishTime: testTime}, "2009/11/10"},
	}
	for _, test := range tests {
		test.Node.Type = &NodeType{
			PathPrefix: test.Prefix,
		}
		ret := test.Node.GetPathPrefix()
		if ret != test.Ret {
			t.Errorf("GetPathPrefix for node %v should be %v, got %v",
				test.Node, test.Ret, ret)
		}
	}
}
