package main

import (
	"bytes"
	"github.com/monsti/rpc/client"
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
