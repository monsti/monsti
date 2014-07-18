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

	"pkg.monsti.org/monsti/api/util"
)

// nodeToData converts the node to a JSON document.
// The Path field will be omitted.
func nodeToData(node *Node, indent bool) ([]byte, error) {
	var data []byte
	var err error
	path := node.Path
	node.Path = ""
	defer func() {
		node.Path = path
	}()

	var outNode nodeJSON
	outNode.Node = *node
	outNode.Type = node.Type.Id
	outNode.Fields = make(util.NestedMap)

	for _, field := range node.Type.Fields {
		outNode.Fields.Set(field.Id, node.Fields[field.Id].Dump())
	}

	if indent {
		data, err = json.MarshalIndent(outNode, "", "  ")
	} else {
		data, err = json.Marshal(outNode)
	}
	if err != nil {
		return nil, fmt.Errorf(
			"service: Could not marshal node: %v", err)
	}
	return data, nil
}

// WriteNode writes the given node.
func (s *MonstiClient) WriteNode(site, path string, node *Node) error {
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

type nodeJSON struct {
	Node
	Type   string
	Fields util.NestedMap
}

// dataToNode unmarshals given data
func dataToNode(data []byte,
	getNodeType func(id string) (*NodeType, error)) (
	*Node, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var node nodeJSON
	err := json.Unmarshal(data, &node)
	if err != nil {
		return nil, fmt.Errorf(
			"service: Could not unmarshal node: %v", err)
	}
	ret := node.Node
	ret.Type, err = getNodeType(node.Type)
	if err != nil {
		return nil, fmt.Errorf("Could not get node type %q: %v",
			node.Type)
	}

	ret.InitFields()
	for _, field := range ret.Type.Fields {
		value := node.Fields.Get(field.Id)
		if value != nil {
			ret.Fields[field.Id].Load(node.Fields.Get(field.Id))
		}
	}
	return &ret, nil
}

// GetNode reads the given node.
//
// If the node does not exist, it returns nil, nil.
func (s *MonstiClient) GetNode(site, path string) (*Node, error) {
	if s.Error != nil {
		return nil, nil
	}
	args := struct{ Site, Path string }{site, path}
	var reply []byte
	err := s.RPCClient.Call("Monsti.GetNode", args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetNode error: %v", err)
	}
	node, err := dataToNode(reply, s.GetNodeType)
	if err != nil {
		return nil, fmt.Errorf("service: Could not convert node: %v", err)
	}
	return node, nil
}

// GetChildren returns the children of the given node.
func (s *MonstiClient) GetChildren(site, path string) ([]*Node, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	args := struct{ Site, Path string }{site, path}
	var reply [][]byte
	err := s.RPCClient.Call("Monsti.GetChildren", args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetChildren error: %v", err)
	}
	nodes := make([]*Node, 0, len(reply))
	for _, entry := range reply {

		node, err := dataToNode(entry, s.GetNodeType)
		if err != nil {
			return nil, fmt.Errorf("service: Could not convert node: %v", err)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// GetNodeData requests data from some node.
//
// Returns a nil slice and nil error if the data does not exist.
func (s *MonstiClient) GetNodeData(site, path, file string) ([]byte, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	type GetNodeDataArgs struct {
	}
	args := struct{ Site, Path, File string }{
		site, path, file}
	var reply []byte
	err := s.RPCClient.Call("Monsti.GetNodeData", &args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetNodeData error:", err)
	}
	return reply, nil
}

// WriteNodeData writes data for some node.
func (s *MonstiClient) WriteNodeData(site, path, file string,
	content []byte) error {
	if s.Error != nil {
		return nil
	}
	args := struct {
		Site, Path, File string
		Content          []byte
	}{
		site, path, file, content}
	if err := s.RPCClient.Call("Monsti.WriteNodeData", &args, new(int)); err != nil {
		return fmt.Errorf("service: WriteNodeData error: %v", err)
	}
	return nil
}

// RemoveNode recursively removes the given site's node.
func (s *MonstiClient) RemoveNode(site string, node string) error {
	if s.Error != nil {
		return nil
	}
	args := struct {
		Site, Node string
	}{site, node}
	if err := s.RPCClient.Call("Monsti.RemoveNode", args, new(int)); err != nil {
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
func (s *MonstiClient) GetConfig(site, module, name string,
	out interface{}) error {
	if s.Error != nil {
		return s.Error
	}
	args := struct{ Site, Module, Name string }{site, module, name}
	var reply []byte
	err := s.RPCClient.Call("Monsti.GetConfig", args, &reply)
	if err != nil {
		return fmt.Errorf("service: GetConfig error: %v", err)
	}
	return getConfig(reply, out)
}

type NodeField struct {
	Id       string
	Name     map[string]string
	Required bool
	Type     string
}

type EmbedNode struct {
	Id  string
	URI string
}

type NodeType struct {
	Id     string
	Name   map[string]string
	Fields []NodeField
	Embed  []EmbedNode
}

// GetNodeType requests information about the given node type.
func (s *MonstiClient) GetNodeType(nodeTypeID string) (*NodeType,
	error) {
	var nodeType NodeType
	err := s.RPCClient.Call("Monsti.GetNodeType", nodeTypeID, &nodeType)
	if err != nil {
		return nil, fmt.Errorf("service: Error calling GetNodeType: %v", err)
	}
	return &nodeType, nil
}

// GetAddableNodeTypes returns the node types that may be added as child nodes
// to the given node type at the given website.
func (s *MonstiClient) GetAddableNodeTypes(site, nodeType string) (types []string,
	err error) {
	args := struct{ Site, NodeType string }{site, nodeType}
	err = s.RPCClient.Call("Monsti.GetAddableNodeTypes", args, &types)
	if err != nil {
		err = fmt.Errorf("service: Error calling GetAddableNodeTypes: %v", err)
	}
	return
}
