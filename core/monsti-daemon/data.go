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

/*
 Monsti is a simple and resource efficient CMS.

 This package implements the data service.
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"pkg.monsti.org/monsti/api/service"
)

// getNode looks up the given node.
// If no such node exists, return nil.
// It adds a path attribute with the given path.
func getNode(root, path string) (node []byte, err error) {
	node_path := filepath.Join(root, path[1:], "node.json")
	node, err = ioutil.ReadFile(node_path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return
	}
	pathJSON := fmt.Sprintf(`{"Path":%q,`, path)
	node = bytes.Replace(node, []byte("{"), []byte(pathJSON), 1)
	return
}

// getChildren looks up child nodes of the given node.
func getChildren(root, path string) (nodes [][]byte, err error) {
	files, err := ioutil.ReadDir(filepath.Join(root, path))
	if err != nil {
		return
	}
	for _, file := range files {
		node, _ := getNode(root, filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return
}

type GetChildrenArgs struct {
	Site, Path string
}

func (i *MonstiService) GetChildren(args GetChildrenArgs,
	reply *[][]byte) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	ret, err := getChildren(site, args.Path)
	*reply = ret
	return err
}

type GetNodeArgs struct{ Site, Path string }

func (i *MonstiService) GetNode(args *GetNodeDataArgs,
	reply *[]byte) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	ret, err := getNode(site, args.Path)
	*reply = ret
	return err
}

type GetNodeDataArgs struct{ Site, Path, File string }

func (i *MonstiService) GetNodeData(args *GetNodeDataArgs,
	reply *[]byte) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	path := filepath.Join(site, args.Path[1:], args.File)
	ret, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		*reply = nil
		return nil
	}
	*reply = ret
	return err
}

type WriteNodeDataArgs struct {
	Site, Path, File string
	Content          []byte
}

func (i *MonstiService) WriteNodeData(args *WriteNodeDataArgs,
	reply *int) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	path := filepath.Join(site, args.Path[1:], args.File)
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return fmt.Errorf("Could not create node directory: %v", err)
	}
	err = ioutil.WriteFile(path, []byte(args.Content), 0600)
	if err != nil {
		return fmt.Errorf("Could not write node data: %v", err)
	}
	return nil
}

type RemoveNodeArgs struct {
	Site, Node string
}

func (i *MonstiService) RemoveNode(args *RemoveNodeArgs, reply *int) error {
	root := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	nodePath := filepath.Join(root, args.Node[1:])
	if err := os.RemoveAll(nodePath); err != nil {
		return fmt.Errorf("Can't remove node: %v", err)
	}
	return nil
}

type RenameNodeArgs struct {
	Site, Source, Target string
}

func (i *MonstiService) RenameNode(args *RenameNodeArgs, reply *int) error {
	root := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	if err := os.Rename(
		filepath.Join(root, args.Source),
		filepath.Join(root, args.Target)); err != nil {
		return fmt.Errorf("Can't move node: %v", err)
	}
	return nil
}

// getConfig returns the configuration value or section for the given name.
func getConfig(path, name string) ([]byte, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not read configuration: %v", err)
	}
	var target interface{}
	err = json.Unmarshal(content, &target)
	if err != nil {
		return nil, fmt.Errorf("Could not parse configuration: %v", err)
	}
	subs := strings.Split(name, ".")
	for _, sub := range subs {
		if sub == "" {
			break
		}
		targetT := reflect.TypeOf(target)
		if targetT != reflect.TypeOf(map[string]interface{}{}) {
			target = nil
			break
		}
		var ok bool
		if target, ok = (target.(map[string]interface{}))[sub]; !ok {
			target = nil
			break
		}
	}
	target = map[string]interface{}{"Value": target}
	ret, err := json.Marshal(target)
	if err != nil {
		return nil, fmt.Errorf("Could not encode configuration: %v", err)
	}
	return ret, nil
}

type GetConfigArgs struct{ Site, Module, Name string }

func (i *MonstiService) GetConfig(args *GetConfigArgs,
	reply *[]byte) error {
	configPath := i.Settings.Monsti.GetSiteConfigPath(args.Site)
	config, err := getConfig(filepath.Join(configPath, args.Module+".json"),
		args.Name)
	if err != nil {
		reply = nil
		return err
	}
	*reply = config
	return nil
}

func findAddableNodeTypes(nodeType string, nodeTypes map[string]service.NodeType,
) []string {
	types := make([]string, 0)
	for _, otherNodeType := range nodeTypes {
		isAddable := false
		for _, addableTo := range otherNodeType.AddableTo {
			if addableTo == "." ||
				addableTo == nodeType || (addableTo[len(addableTo)-1] == '.' &&
				nodeType[0:len(addableTo)] == addableTo) {
				isAddable = true
				break
			}
		}
		if otherNodeType.AddableTo == nil || isAddable {
			types = append(types, otherNodeType.Id)
		}
	}
	return types
}

type GetAddableNodeTypesArgs struct{ Site, NodeType string }

func (i *MonstiService) GetAddableNodeTypes(args GetAddableNodeTypesArgs,
	types *[]string) error {
	i.mutex.RLock()
	*types = findAddableNodeTypes(args.NodeType, i.Settings.Config.NodeTypes)
	defer i.mutex.RUnlock()
	return nil
}

func (i *MonstiService) GetNodeType(nodeTypeID string,
	ret *service.NodeType) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if nodeType, ok := i.Settings.Config.NodeTypes[nodeTypeID]; ok {
		*ret = nodeType
		return nil
	}
	return fmt.Errorf("Unknown node type %q", nodeTypeID)
}
