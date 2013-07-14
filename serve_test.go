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
	"bytes"
	"pkg.monsti.org/rpc/client"
	"net/http"
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

type responseWriter struct {
	Body []byte
}

func (r *responseWriter) Header() (h http.Header) {
	return make(http.Header, 0)
}

func (r *responseWriter) Write(data []byte) (int, error) {
	r.Body = data
	return len(data), nil
}

func (r *responseWriter) WriteHeader(code int) {
}

func TestProcessNodeResponse(t *testing.T) {
	tests := []struct {
		res  client.Response
		body []byte
	}{
		{
			res: client.Response{
				Body: []byte("foo"),
				Raw:  true},
			body: []byte("foo")}}
	for i, v := range tests {
		w := responseWriter{}
		h := nodeHandler{}
		req := http.Request{}
		node := client.Node{}
		site := site{SessionAuthKey: "foobar"}
		session := getSession(&req, site)
		h.ProcessNodeResponse(v.res, &w, &req, node, "action",
			session, &client.Session{}, site)
		if !bytes.Equal(w.Body, v.body) {
			t.Errorf("Test %v failed: Body should be %q, was %q.", i, v.body,
				w.Body)
		}

	}
}
