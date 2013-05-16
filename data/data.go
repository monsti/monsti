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

package data

import (
	"net/rpc"
	"os"
	"strings"
)

// GetNodeData requests data from some node.
//
// If the data does not exist, return null length []byte.
func (s *Service) GetNodeData(path, file string) []byte {
	args := &types.GetNodeDataArgs{path, file}
	var reply []byte
	err := s.Call("NodeRPC.GetNodeData", args, &reply)
	if err != nil {
		s.Logger.Fatal("master: RPC GetNodeData error:", err)
	}
	return reply
}

// WriteNodeData writes data for some node.
func (s *Service) WriteNodeData(path, file, content string) error {
	args := &types.WriteNodeDataArgs{path, file, content}
	return s.Call("NodeRPC.WriteNodeData", args, new(int))
}

// UpdateNode saves changes to given node.
func (s *Service) UpdateNode(node Node) error {
	return s.Call("NodeRPC.UpdateNode", node, new(int))
}
