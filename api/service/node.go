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

package service

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"strings"
)

// NodeClient represents the RPC connection to the Nodes service.
type NodeClient struct {
	*Client
}

// NewNodeClient returns a new Node Client.
func NewNodeClient() *NodeClient {
	var service_ NodeClient
	service_.Client = new(Client)
	return &service_
}

type NodeFields struct {
	Path string `json:"-"`
	// Content type of the node.
	Type  string
	Order int
	// Don't show the node in navigations if Hide is true.
	Hide   bool
	Fields map[string]interface{}
}

// GetField returns the named field (and true) of the node if present.
//
// If there is no such field, it returns nil.
func (n NodeFields) GetField(id string) interface{} {
	parts := strings.Split(id, ".")
	field := interface{}(n.Fields)
	for _, part := range parts {
		var ok bool
		field, ok = field.(map[string]interface{})[part]
		if !ok {
			return nil
		}
	}
	return field
}

// SetField sets the value of the named field.
func (n *NodeFields) SetField(id string, value interface{}) {
	parts := strings.Split(id, ".")
	if n.Fields == nil {
		n.Fields = make(map[string]interface{})
	}
	field := interface{}(n.Fields)
	for _, part := range parts[:len(parts)-1] {
		next := field.(map[string]interface{})[part]
		if next == nil {
			next = make(map[string]interface{})
			field.(map[string]interface{})[part] = next
		}
		field = next
	}
	field.(map[string]interface{})[parts[len(parts)-1]] = value
}

// PathToID returns an ID for the given node based on it's path.
//
// The ID is simply the path of the node with all slashes replaced by two
// underscores and the result prefixed with "node-".
//
// PathToID will panic if the path is not set.
//
// For example, a node with path "/foo/bar" will get the ID "node-__foo__bar".
func (n NodeFields) PathToID() string {
	if len(n.Path) == 0 {
		panic("Can't calculate ID of node with unset path.")
	}
	return "node-" + strings.Replace(n.Path, "/", "__", -1)
}

// Name returns the name of the node.
func (n NodeFields) Name() string {
	base := path.Base(n.Path)
	if base == "." || base == "/" {
		return ""
	}
	return base
}

// RequestFile stores the path or content of a multipart request's file.
type RequestFile struct {
	// TmpFile stores the path to a temporary file with the contents.
	TmpFile string
	// Content stores the file content if TmpFile is not set.
	Content []byte
}

// ReadFile returns the file's content. Uses io/ioutil ReadFile if the request
// file's content is in a temporary file.
func (r RequestFile) ReadFile() ([]byte, error) {
	if len(r.TmpFile) > 0 {
		return ioutil.ReadFile(r.TmpFile)
	}
	return r.Content, nil
}

type RequestMethod uint

const (
	GetRequest = iota
	PostRequest
)

type Action uint

const (
	ViewAction = iota
	EditAction
	LoginAction
	LogoutAction
	AddAction
	RemoveAction
)

// A request to be processed by a nodes service.
type Request struct {
	// Site name
	Site string
	// The requested node.
	Node NodeFields
	// The query values of the request URL.
	Query url.Values
	// Method of the request (GET,POST,...).
	Method RequestMethod
	// User session
	Session UserSession
	// Action to perform (e.g. "edit").
	Action Action
	// FormData stores the requests form data.
	FormData url.Values
	// Files stores files of multipart requests.
	Files map[string][]RequestFile
}

// Response to a node request.
type Response struct {
	// The html content to be embedded in the root template.
	Body []byte
	// Raw must be set to true if Body should not be embedded in the root
	// template. The content type will be automatically detected.
	Raw bool
	// If set, redirect to this target using error 303 'see other'.
	Redirect string
	// The node as received by GetRequest, possibly with some fields
	// updated (e.g. modified title).
	//
	// If nil, the original node data is used.
	Node *NodeFields
}

// Write appends the given bytes to the body of the response.
func (r *Response) Write(p []byte) (n int, err error) {
	r.Body = append(r.Body, p...)
	return len(p), nil
}

// Request performs the given request.
func (s *NodeClient) Request(req *Request) (*Response, error) {
	var res Response
	err := s.RPCClient.Call("Node.Request", req, &res)
	if err != nil {
		return nil, fmt.Errorf("service: RPC error for Request: %v", err)
	}
	return &res, nil
}

// GetNodeType returns all supported node types.
func (s *NodeClient) GetNodeTypes() ([]string, error) {
	var res []string
	err := s.RPCClient.Call("Node.GetNodeTypes", 0, &res)
	if err != nil {
		return nil, fmt.Errorf("service: RPC error for GetNodeTypes: %v", err)
	}
	return res, nil
}
