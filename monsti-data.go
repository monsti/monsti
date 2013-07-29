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
	"flag"
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"os"
	"path/filepath"
	"pkg.monsti.org/service"
	"pkg.monsti.org/util"
	"sync"
)

// DataService implements RPC methods for the Data service.
type DataService struct {
	Info     *service.InfoClient
	Settings settings
}

// getNode looks up the given node.
// If no such node exists, return nil.
func getNode(root, path string) (node *service.NodeInfo, err error) {
	node_path := filepath.Join(root, path[1:], "node.yaml")
	content, err := ioutil.ReadFile(node_path)
	if err != nil {
		return
	}
	if err = goyaml.Unmarshal(content, &node); err != nil {
		node = nil
		return
	}
	node.Path = path
	return
}

// getChildren looks up child nodes of the given node.
func getChildren(root, path string) (nodes []service.NodeInfo, err error) {
	files, err := ioutil.ReadDir(filepath.Join(root, path))
	if err != nil {
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		node, _ := getNode(root, filepath.Join(path, file.Name()))
		if node != nil {
			nodes = append(nodes, *node)
		}
	}
	return
}

type GetChildrenArgs struct {
	Site, Node string
}

func (i *DataService) GetChildren(args GetChildrenArgs,
	reply *[]service.NodeInfo) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	ret, err := getChildren(site, args.Node[1:])
	*reply = ret
	return err
}

func (i *DataService) GetNodeData(args *service.GetNodeDataArgs,
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

func (i *DataService) WriteNodeData(args *service.WriteNodeDataArgs,
	reply *int) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	path := filepath.Join(site, args.Path[1:], args.File)
	err := ioutil.WriteFile(path, []byte(args.Content), 0600)
	return err
}

func (i *DataService) UpdateNode(args *service.UpdateNodeArgs, reply *int) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	return writeNode(args.Node, site)
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

// writeNode writes the given node to the data directory located at the given
// root.
func writeNode(reqnode service.NodeInfo, root string) error {
	path := reqnode.Path
	reqnode.Path = ""
	content, err := goyaml.Marshal(&reqnode)
	if err != nil {
		return err
	}
	node_path := filepath.Join(root, path[1:],
		"node.yaml")
	if err := os.Mkdir(filepath.Dir(node_path), 0700); err != nil {
		if !os.IsExist(err) {
			panic("Can't create directory for new node: " + err.Error())
		}
	}
	return ioutil.WriteFile(node_path, content, 0600)
}

type settings struct {
	Monsti util.MonstiSettings
}

func main() {
	logger := log.New(os.Stderr, "data ", log.LstdFlags)

	// Load configuration
	flag.Parse()
	cfgPath := util.GetConfigPath(flag.Arg(0))
	var settings settings
	if err := util.LoadModuleSettings("data", cfgPath, &settings); err != nil {
		logger.Fatal("Could not load settings: ", err)
	}

	// Connect to Info service
	info, err := service.NewInfoConnection(settings.Monsti.GetServicePath(
		service.Info.String()))
	if err != nil {
		logger.Fatalf("Could not connect to Info service: %v", err)
	}

	// Start own Data service
	var waitGroup sync.WaitGroup
	logger.Println("Starting Data service")
	waitGroup.Add(1)
	dataPath := settings.Monsti.GetServicePath(service.Data.String())
	go func() {
		defer waitGroup.Done()
		var provider service.Provider
		var data_ DataService
		data_.Info = info
		data_.Settings = settings
		provider.Logger = logger
		if err := provider.Serve(dataPath, "Data", &data_); err != nil {
			logger.Fatalf("Could not start Data service: %v", err)
		}
	}()

	if err := info.PublishService("Data", dataPath); err != nil {
		logger.Fatalf("Could not publish Data service: %v", err)
	}

	waitGroup.Wait()
}
