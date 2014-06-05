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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util"
)

// DataService implements RPC methods for the Data service.
type DataService struct {
	Settings settings
}

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

func (i *DataService) GetChildren(args GetChildrenArgs,
	reply *[][]byte) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	ret, err := getChildren(site, args.Path)
	*reply = ret
	return err
}

type GetNodeArgs struct{ Site, Path string }

func (i *DataService) GetNode(args *GetNodeDataArgs,
	reply *[]byte) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	ret, err := getNode(site, args.Path)
	*reply = ret
	return err
}

type GetNodeDataArgs struct{ Site, Path, File string }

func (i *DataService) GetNodeData(args *GetNodeDataArgs,
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

func (i *DataService) WriteNodeData(args *WriteNodeDataArgs,
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

func (i *DataService) RemoveNode(args *RemoveNodeArgs, reply *int) error {
	root := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	nodePath := filepath.Join(root, args.Node[1:])
	if err := os.RemoveAll(nodePath); err != nil {
		return fmt.Errorf("Can't remove node: %v", err)
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

func (i *DataService) GetConfig(args *GetConfigArgs,
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

type settings struct {
	Monsti util.MonstiSettings
}

func main() {
	logger := log.New(os.Stderr, "", 0)

	// Load configuration
	flag.Parse()
	cfgPath := util.GetConfigPath(flag.Arg(0))
	var settings settings
	if err := util.LoadModuleSettings("data", cfgPath, &settings); err != nil {
		logger.Fatal("Could not load settings: ", err)
	}

	// Connect to Info service
	info, err := service.NewInfoConnection(settings.Monsti.GetServicePath(
		service.InfoService.String()))
	if err != nil {
		logger.Fatalf("Could not connect to Info service: %v", err)
	}

	// Start own Data service
	var waitGroup sync.WaitGroup
	logger.Println("Setting up Data service")
	waitGroup.Add(1)
	dataPath := settings.Monsti.GetServicePath(service.DataService.String())
	go func() {
		defer waitGroup.Done()
		var data_ DataService
		data_.Settings = settings
		provider := service.NewProvider("Data", &data_)
		provider.Logger = logger
		if err := provider.Listen(dataPath); err != nil {
			logger.Fatalf("Could not start Data service: %v", err)
		}
		if err := provider.Accept(); err != nil {
			logger.Fatalf("Could not accept at Data service: %v", err)
		}
	}()

	if err := info.PublishService("Data", dataPath); err != nil {
		logger.Fatalf("Could not publish Data service: %v", err)
	}

	waitGroup.Wait()
}
