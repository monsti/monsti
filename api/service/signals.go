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

// SignalHandler wraps a handler for a specific signal.
type SignalHandler interface {
	// Name returns the name of the signal to handle.
	Name() string
	// Handle handles a signal with given arguments.
	Handle(args interface{}) (interface{}, error)
}

type nodeContextHandler struct {
	f func(Request int, NodeType string) string
}

func (r *nodeContextHandler) Name() string {
	return "monsti.NodeContext"
}

func (r *nodeContextHandler) Handle(args interface{}) (interface{}, error) {
	args_ := args.(string)
	return r.f(1, args_), nil
}

// NewNodeContextHandler consructs a signal handler that adds some
// template context for rendering a node.
func NewNodeContextHandler(
	cb func(Request int, NodeType string) string) SignalHandler {
	return &nodeContextHandler{cb}
}
