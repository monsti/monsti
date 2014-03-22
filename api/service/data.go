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
	"strings"
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

// lowerCaseEqualTo returns a function that checks for lowercase
// equality with the given value.
func lowerCaseEqualTo(value string) func(string) bool {
	return func(name string) bool {
		if strings.ToLower(name) == strings.ToLower(value) {
			return true
		}
		return false
	}
}

// nodeToData converts the node's fields of the given field namespaces
// to a JSON document.
func nodeToData(node interface{}, namespaces []string) ([][]byte, error) {
	nodeType := reflect.TypeOf(node)
	nodeValue := reflect.ValueOf(node)
	if nodeType.Kind() == reflect.Ptr {
		nodeType = nodeType.Elem()
		nodeValue = nodeValue.Elem()
	}
	if nodeType.Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"service: Node must be a struct or a pointer to a struct")
	}
	ret := make([][]byte, 0, len(namespaces))
	for _, ns := range namespaces {
		nsFields := nodeValue.FieldByNameFunc(lowerCaseEqualTo(ns + "fields"))
		if !nsFields.IsValid() {
			panic(fmt.Errorf("service: Invalid field namespace %q", ns))
		}
		data, err := json.Marshal(nsFields.Interface())
		if err != nil {
			return nil, fmt.Errorf(
				"service: Could not marshal fields of namespace %v: %v", ns, err)
		}
		ret = append(ret, data)
	}
	return ret, nil
}

// WriteNode writes the named fields of the given node.
//
// It panics if the node does not contain one of the named fields.
func (s *DataClient) WriteNode(site, path string, node interface{},
	fields ...string) error {
	if s.Error != nil {
		return nil
	}
	fieldsData, err := nodeToData(node, fields)
	if err != nil {
		return fmt.Errorf("service: Could not convert fields: %v", err)
	}
	for idx, field := range fields {
		err := s.WriteNodeData(site, path, field+".json", fieldsData[idx])
		if err != nil {
			return fmt.Errorf(
				"service: Could not write node fields for namespace %v: %v", field, err)
		}
	}
	return nil
}

// dataToNode fills gven node's fields.
func dataToNode(data [][]byte, node interface{}, namespaces []string) error {
	nodeType := reflect.TypeOf(node)
	nodeValue := reflect.ValueOf(node)
	if nodeType.Kind() != reflect.Ptr ||
		nodeType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("service: Node must be a pointer to a struct")
	}
	nodeValue = nodeValue.Elem()
	for idx, ns := range namespaces {
		if len(data[idx]) == 0 {
			continue
		}
		nsFields := nodeValue.FieldByNameFunc(lowerCaseEqualTo(ns + "fields"))
		err := json.Unmarshal(data[idx], nsFields.Addr().Interface())
		if err != nil {
			return fmt.Errorf(
				"service: Could not unmarshal fields of namespace %v: %v", ns, err)
		}
	}
	return nil
}

// ReadNode reads the named fields into the given node.
// Fields without any data present will be ignored.
func (s *DataClient) ReadNode(site, path string, node interface{},
	fields ...string) error {
	if s.Error != nil {
		return nil
	}
	fieldsData := make([][]byte, 0, len(fields))
	for _, field := range fields {
		fieldData, err := s.GetNodeData(site, path, field+".json")
		if err != nil {
			return fmt.Errorf(
				"service: Could not read node fields for namespace %v: %v", field, err)
		}
		fieldsData = append(fieldsData, fieldData)
	}
	err := dataToNode(fieldsData, node, fields)
	if err != nil {
		return fmt.Errorf("service: Could not fill fields: %v", err)
	}
	return nil
}

// GetChildren returns the children of the given node.
func (s *DataClient) GetChildren(site, path string) ([]*NodeFields, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	args := struct{ Site, Path string }{site, path}
	var reply [][]byte
	err := s.RPCClient.Call("Data.GetChildren", args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetChildren error: %v", err)
	}
	nodes := make([]*NodeFields, 0, len(reply))
	for _, entry := range reply {
		node := &NodeFields{}
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
