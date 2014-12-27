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

package service

import "encoding/gob"

type NodeContextArgs struct {
	Request   uint
	NodeType  string
	EmbedNode *EmbedNode
}

type NodeContextRet struct {
	Context map[string]string
	Mods    *CacheMods
}

func init() {
	gob.RegisterName("monsti.NodeContextArgs", NodeContextArgs{})
	gob.RegisterName("monsti.NodeContextRet", NodeContextRet{})
}

// SignalHandler wraps a handler for a specific signal.
type SignalHandler interface {
	// Name returns the name of the signal to handle.
	Name() string
	// Handle handles a signal with given arguments.
	Handle(args interface{}) (interface{}, error)
}

type nodeContextHandler struct {
	f func(Request uint, NodeType string, embedNode *EmbedNode) (
		map[string]string, *CacheMods, error)
}

func (r *nodeContextHandler) Name() string {
	return "monsti.NodeContext"
}

func (r *nodeContextHandler) Handle(args interface{}) (interface{}, error) {
	args_ := args.(NodeContextArgs)
	context, mods, err := r.f(args_.Request, args_.NodeType, args_.EmbedNode)
	return NodeContextRet{context, mods}, err
}

// NewNodeContextHandler consructs a signal handler that adds some
// template context for rendering a node.
func NewNodeContextHandler(
	cb func(Request uint, NodeType string,
		embedNode *EmbedNode) (map[string]string, *CacheMods, error)) SignalHandler {
	return &nodeContextHandler{cb}
}
