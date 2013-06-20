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
)

// DataClient represents the RPC connection to the Data service.
type DataClient struct {
	*Client
}

// NewDataClient returns a new Data Client.
func NewDataClient() *DataClient {
	var service_ DataClient
	service_.Client = new(Client)
	return &service_
}

type GetNodeDataArgs struct {
	Site, Path, File string
}

// GetNodeData requests data from some node.
//
// Returns a nil slice and nil error if the data does not exist.
func (s *DataClient) GetNodeData(site, path, file string) ([]byte, error) {
	args := &GetNodeDataArgs{site, path, file}
	var reply []byte
	err := s.RPCClient.Call("Data.GetNodeData", args, &reply)
	if err != nil {
		return nil, fmt.Errorf("info: GetNodeData error:", err)
	}
	return reply, nil
}

type WriteNodeDataArgs struct {
	Site, Path, File, Content string
}

// WriteNodeData writes data for some node.
func (s *DataClient) WriteNodeData(site, path, file, content string) error {
	args := &WriteNodeDataArgs{site, path, file, content}
	if err := s.RPCClient.Call("Data.WriteNodeData", args, new(int)); err != nil {
		return fmt.Errorf("info: WriteNodeData error:", err)
	}
	return nil
}

type UpdateNodeArgs struct {
	Site string
	Node NodeInfo
}

// UpdateNode saves changes to given node.
func (s *DataClient) UpdateNode(site string, node_ NodeInfo) error {
	args := &UpdateNodeArgs{site, node_}
	if err := s.RPCClient.Call("Data.UpdateNode", args, new(int)); err != nil {
		return fmt.Errorf("info: WriteNodeData error:", err)
	}
	return nil
}
