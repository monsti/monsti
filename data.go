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

// GetNode returns the given node or nil if it does not exist.
func (s *DataClient) GetNode(site, path string) (*NodeInfo, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	args := struct{ Site, Path string }{site, path}
	var reply []byte
	err := s.RPCClient.Call("Data.GetNode", args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetNode error: %v", err)
	}
	node := &NodeInfo{}
	err = json.Unmarshal(reply, node)
	if err != nil {
		return nil, fmt.Errorf("service: Could not decode node: %v", err)
	}
	return node, nil
}

// FillFields loads the fields of the given nodes into target.
//
// If only one node is given, target must be a pointer to a struct.
// If more than one node is given, the target must be an initialized
// slice of structs.
//
// After loading the fields into the target, the node will be assigned
// to the target's (possibly embedded) NodeInfo field.
func (s *DataClient) FillFields(target interface{}, site string, nodes ...*NodeInfo) error {
	if s.Error != nil {
		return s.Error
	}
	targetType := reflect.TypeOf(target)
	targetValue := reflect.ValueOf(target)
	switch len(nodes) {
	case 0:
		return nil
	case 1:
		if targetType.Kind() != reflect.Ptr ||
			targetType.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("service: Target must be a pointer to a struct")
		}
		fields, err := s.GetNodeData(site, nodes[0].Path, "fields.json")
		if err != nil {
			return fmt.Errorf("Could not get node fields: %v", err)
		}
		err = json.Unmarshal(fields, target)
		if err != nil {
			return fmt.Errorf("Could not decode fields for %q: %v", nodes[0].Path, err)
		}
		info := targetValue.Elem().FieldByName("NodeInfo")
		if info.Type().Kind() == reflect.Ptr {
			info.Set(reflect.ValueOf(nodes[0]))
		} else {
			info.Set(reflect.ValueOf(nodes[0]).Elem())
		}
		return nil
	default:
		if targetType.Kind() != reflect.Ptr ||
			targetType.Elem().Kind() != reflect.Slice ||
			targetValue.Elem().IsNil() {
			return fmt.Errorf("service: Target must be a pointer to a non-nil slice")
		}
		for _, node := range nodes {
			singleTarget := reflect.New(targetType.Elem().Elem())
			if err := s.FillFields(singleTarget.Interface(), site, node); err != nil {
				return err
			}
			targetValue.Elem().Set(
				reflect.Append(targetValue.Elem(), singleTarget.Elem()))
		}
		return nil
	}
}

// GetChildren returns the children of the given node.
func (s *DataClient) GetChildren(site, path string) ([]*NodeInfo, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	args := struct{ Site, Path string }{site, path}
	var reply [][]byte
	err := s.RPCClient.Call("Data.GetChildren", args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetChildren error: %v", err)
	}
	nodes := make([]*NodeInfo, 0, len(reply))
	for _, entry := range reply {
		node := &NodeInfo{}
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

// UpdateNode saves changes to given node.
func (s *DataClient) UpdateNode(site string, node_ NodeInfo) error {
	if s.Error != nil {
		return nil
	}
	content, err := json.Marshal(node_)
	if err != nil {
		return fmt.Errorf("service: Could not marshal node: %v", err)
	}
	args := struct {
		Site, Path string
		Content    []byte
	}{
		site, node_.Path, content}
	if err := s.RPCClient.Call("Data.UpdateNode", &args, new(int)); err != nil {
		return fmt.Errorf("service: UpdateNode error: %v", err)
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
