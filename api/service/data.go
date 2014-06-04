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
	"encoding/json"
	"fmt"
	"reflect"
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

// nodeToData converts the node to a JSON document.
func nodeToData(node *Node, indent bool) ([]byte, error) {
	var data []byte
	var err error
	if indent {
		data, err = json.MarshalIndent(node, "", "  ")
	} else {
		data, err = json.Marshal(node)
	}
	if err != nil {
		return nil, fmt.Errorf(
			"service: Could not marshal node: %v", err)
	}
	return data, nil
}

// WriteNode writes the given node.
func (s *DataClient) WriteNode(site, path string, node *Node) error {
	if s.Error != nil {
		return nil
	}
	data, err := nodeToData(node, true)
	if err != nil {
		return fmt.Errorf("service: Could not convert node: %v", err)
	}
	err = s.WriteNodeData(site, path, "node.json", data)
	if err != nil {
		return fmt.Errorf(
			"service: Could not write node: %v", err)
	}
	return nil
}

// dataToNode unmarshals given data
func dataToNode(data []byte) (*Node, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var node Node
	err := json.Unmarshal(data, &node)
	if err != nil {
		return nil, fmt.Errorf(
			"service: Could not unmarshal node: %v", err)
	}
	return &node, nil
}

// GetNode reads the given node.
func (s *DataClient) GetNode(site, path string) (*Node, error) {
	if s.Error != nil {
		return nil, nil
	}
	data, err := s.GetNodeData(site, path, "node.json")
	if err != nil {
		return nil, fmt.Errorf("service: Could not read node: %v", err)
	}
	node, err := dataToNode(data)
	if err != nil {
		return nil, fmt.Errorf("service: Could not convert node: %v", err)
	}
	node.Path = path
	return node, nil
}

// GetChildren returns the children of the given node.
func (s *DataClient) GetChildren(site, path string) ([]*Node, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	args := struct{ Site, Path string }{site, path}
	var reply [][]byte
	err := s.RPCClient.Call("Data.GetChildren", args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetChildren error: %v", err)
	}
	nodes := make([]*Node, 0, len(reply))
	for _, entry := range reply {
		node := &Node{}
		err = json.Unmarshal(entry, node)
		if err != nil {
			return nil, fmt.Errorf("service: Could not decode node: %v", err)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// GetNodeData requests data from some node.
//
// Returns a nil slice and nil error if the data does not exist.
func (s *DataClient) GetNodeData(site, path, file string) ([]byte, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	type GetNodeDataArgs struct {
	}
	args := struct{ Site, Path, File string }{
		site, path, file}
	var reply []byte
	err := s.RPCClient.Call("Data.GetNodeData", &args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetNodeData error:", err)
	}
	return reply, nil
}

// WriteNodeData writes data for some node.
func (s *DataClient) WriteNodeData(site, path, file string,
	content []byte) error {
	if s.Error != nil {
		return nil
	}
	args := struct {
		Site, Path, File string
		Content          []byte
	}{
		site, path, file, content}
	if err := s.RPCClient.Call("Data.WriteNodeData", &args, new(int)); err != nil {
		return fmt.Errorf("service: WriteNodeData error: %v", err)
	}
	return nil
}

// RemoveNode recursively removes the given site's node.
func (s *DataClient) RemoveNode(site string, node string) error {
	if s.Error != nil {
		return nil
	}
	args := struct {
		Site, Node string
	}{site, node}
	if err := s.RPCClient.Call("Data.RemoveNode", args, new(int)); err != nil {
		return fmt.Errorf("service: RemoveNode error: %v", err)
	}
	return nil
}

func getConfig(reply []byte, out interface{}) error {
	objectV := reflect.New(
		reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(out)))
	err := json.Unmarshal(reply, objectV.Interface())
	if err != nil {
		return fmt.Errorf("service: Could not decode configuration: %v", err)
	}
	value := objectV.Elem().MapIndex(
		objectV.Elem().MapKeys()[0])
	if !value.IsNil() {
		reflect.ValueOf(out).Elem().Set(value.Elem())
	}
	return nil
}

// GetConfig puts the named configuration into the variable out.
func (s *DataClient) GetConfig(site, module, name string,
	out interface{}) error {
	if s.Error != nil {
		return s.Error
	}
	args := struct{ Site, Module, Name string }{site, module, name}
	var reply []byte
	err := s.RPCClient.Call("Data.GetConfig", args, &reply)
	if err != nil {
		return fmt.Errorf("service: GetConfig error: %v", err)
	}
	return getConfig(reply, out)
}
