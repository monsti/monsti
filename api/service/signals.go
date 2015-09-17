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

package service

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/chrneumann/htmlwidgets"
)

func init() {
	gob.RegisterName("monsti.NodeContextArgs", NodeContextArgs{})
	gob.RegisterName("monsti.NodeContextRet", NodeContextRet{})
	gob.RegisterName("monsti.RenderNodeArgs", RenderNodeArgs{})
	gob.RegisterName("monsti.RenderNodeRet", RenderNodeRet{})
	gob.Register(new(template.HTML))
	gob.Register(new(htmlwidgets.RenderData))
}

// SignalHandler wraps a handler for a specific signal.
type SignalHandler interface {
	// Name returns the name of the signal to handle.
	Name() string
	// Handle handles a signal with given arguments.
	Handle(args interface{}) (interface{}, error)
}

type NodeContextArgs struct {
	Request   uint
	NodeType  string
	EmbedNode *EmbedNode
}

type NodeContextRet struct {
	Context map[string][]byte
	Mods    *CacheMods
}

type nodeContextHandler struct {
	f func(request uint, session *Session, nodeType string, embedNode *EmbedNode) (
		map[string][]byte, *CacheMods, error)
	sessions *SessionPool
}

func (r *nodeContextHandler) Name() string {
	return "monsti.NodeContext"
}

func (r *nodeContextHandler) Handle(args interface{}) (interface{}, error) {
	session, err := r.sessions.New()
	if err != nil {
		return nil, fmt.Errorf("service: Could not get session: %v", err)
	}
	defer r.sessions.Free(session)
	args_ := args.(NodeContextArgs)
	context, mods, err := r.f(args_.Request, session, args_.NodeType,
		args_.EmbedNode)
	return NodeContextRet{context, mods}, err
}

// NewNodeContextHandler consructs a signal handler that adds some
// template context for rendering a node.
//
// DEPRECATED: Use the more powerful RenderNode signal.
func NewNodeContextHandler(
	sessions *SessionPool,
	cb func(request uint, session *Session, nodeType string,
		embedNode *EmbedNode) (map[string][]byte, *CacheMods, error)) SignalHandler {
	return &nodeContextHandler{cb, sessions}
}

type RenderNodeArgs struct {
	Request   uint
	NodeType  string
	EmbedNode *EmbedNode
}

// Redirect configures a HTTP redirect with the given URL and HTTP status.
type Redirect struct {
	URL    string
	Status int
}

type RenderNodeRet struct {
	// Raw context data. Set and get using the SetContext and Context methods.
	RawContext []byte
	// If set, perform this redirect.
	Redirect *Redirect
	Mods     *CacheMods
}

// SetContext sets the RawContext attribute using the given context data.
//
// The data will be marshaled using Go's JSON package.
func (r *RenderNodeRet) SetContext(in map[string]interface{}) error {
	var err error
	r.RawContext, err = json.Marshal(in)
	if err != nil {
		return fmt.Errorf("Could not marshal value: %v", err)
	}
	return nil
}

// Context retreives the context data stored in the RawContext attribute.
//
// The data will be unmarshaled using Go's JSON package.
func (r *RenderNodeRet) Context() (map[string]interface{}, error) {
	if r.RawContext == nil {
		return nil, nil
	}
	var unmarshaled map[string]interface{}
	err := json.Unmarshal(r.RawContext, &unmarshaled)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal value %v: %v", r.RawContext, err)
	}
	return unmarshaled, nil
}

type renderNodeHandler struct {
	f        func(signal *RenderNodeArgs, session *Session) (*RenderNodeRet, error)
	sessions *SessionPool
}

func (r *renderNodeHandler) Name() string {
	return "monsti.RenderNode"
}

func (r *renderNodeHandler) Handle(args interface{}) (interface{}, error) {
	session, err := r.sessions.New()
	if err != nil {
		return nil, fmt.Errorf("service: Could not get session: %v", err)
	}
	defer r.sessions.Free(session)
	args_ := args.(RenderNodeArgs)
	ret, err := r.f(&args_, session)
	if ret == nil {
		ret = new(RenderNodeRet)
	}
	return ret, err
}

// NewRenderNodeHandler consructs a signal handler that may add some
// template context for rendering a node and issue redirects.
func NewRenderNodeHandler(
	sessions *SessionPool,
	cb func(args *RenderNodeArgs, session *Session) (
		*RenderNodeRet, error)) SignalHandler {
	return &renderNodeHandler{cb, sessions}
}
