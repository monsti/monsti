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
	"github.com/monsti/service"
	"github.com/monsti/util"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// DataService implements RPC methods for the Data service.
type DataService struct {
	Info     *service.InfoClient
	Settings settings
}

func (i *DataService) GetNodeData(args *service.GetNodeDataArgs,
	reply *[]byte) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	path := filepath.Join(site, args.Path[1:], args.File)
	ret, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			*reply = make([]byte, 0)
			return nil
		}
	} else {
		*reply = ret
	}
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
