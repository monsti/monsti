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

package main

import (
	"fmt"
	"github.com/monsti/service"
	"sync"
)

type InfoService struct {
	// Services maps service names to service paths
	Services map[string][]string
	// NodeTypes maps node types to service paths
	NodeTypes map[string][]string
	// Mutex to syncronize data access
	mutex sync.RWMutex
}

func (i *InfoService) PublishService(args service.PublishServiceArgs,
	reply *int) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.Services == nil {
		i.Services = make(map[string][]string)
	}
	switch args.Service {
	case "Node":
		if i.NodeTypes == nil {
			i.NodeTypes = make(map[string][]string)
		}
		nodeServ := service.NewNodeClient()
		if err := nodeServ.Connect(args.Path); err != nil {
			return fmt.Errorf("Could not connect to your node service: %v", err)
		}
		nodeTypes, err := nodeServ.GetNodeTypes()
		if err != nil {
			return fmt.Errorf("Could not retrieve your node types: %v", err)
		}
		for _, nodeType := range nodeTypes {
			if i.NodeTypes[nodeType] == nil {
				i.NodeTypes[nodeType] = make([]string, 0)
			}
			i.NodeTypes[nodeType] = append(i.NodeTypes[nodeType], args.Path)
		}
	case "Data":
	case "Mail":
	default:
		return fmt.Errorf("Unknown service type %v", args.Service)
	}

	if i.Services[args.Service] == nil {
		i.Services[args.Service] = make([]string, 0)
	}
	i.Services[args.Service] = append(i.Services[args.Service], args.Path)

	return nil
}

func (i *InfoService) FindNodeService(nodeType string, path *string) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if len(i.NodeTypes[nodeType]) == 0 {
		return fmt.Errorf("Unknown node type %v", nodeType)
	}
	*path = i.NodeTypes[nodeType][0]
	return nil
}

type GetAddableNodeTypesArgs struct{ Site, NodeType string }

func (i *InfoService) GetAddableNodeTypes(args GetAddableNodeTypesArgs,
	types *[]string) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	*types = make([]string, 0)
	for nodeType, _ := range i.NodeTypes {
		*types = append(*types, nodeType)
	}
	return nil
}

func (i *InfoService) FindDataService(arg int, path *string) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if len(i.Services["Data"]) == 0 {
		return fmt.Errorf("Could not find any data services")
	}
	*path = i.Services["Data"][0]
	return nil
}

func (i *InfoService) FindMailService(arg int, path *string) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if len(i.Services["Mail"]) == 0 {
		return fmt.Errorf("Could not find any mail services")
	}
	*path = i.Services["Mail"][0]
	return nil
}
