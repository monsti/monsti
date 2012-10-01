package main

import (
	"testing"
)

func TestSplitAction(t *testing.T) {
	tests := []struct {
		Path, NodePath, Action string
	}{
		{"/", "/", ""},
		{"/@@action", "/", "action"},
		{"/foo", "/foo", ""},
		{"/foo/", "/foo/", ""},
		{"/foo/@@action", "/foo", "action"},
		{"/foo@@action", "/foo@@action", ""},
		{"/foo/@@action/foo", "/foo/@@action/foo", ""},
		{"/foo/bar", "/foo/bar", ""},
		{"/foo/bar/@@action", "/foo/bar", "action"},
		{"/foo/bar/@@action/", "/foo/bar/@@action/", ""}}
	for _, v := range tests {
		rnode, raction := splitAction(v.Path)
		if rnode != v.NodePath || raction != v.Action {
			t.Errorf("splitAction(%v) returns (%v,%v), expected (%v,%v)",
				v.Path, rnode, raction, v.NodePath, v.Action)
		}
	}
}
