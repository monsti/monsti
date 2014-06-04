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
	"sync"

	"pkg.monsti.org/monsti/api/service"
)

type InfoService struct {
	// Services maps service names to service paths
	Services map[string][]string
	// Mutex to syncronize data access
	mutex  sync.RWMutex
	Config *Config
}

func (i *InfoService) PublishService(args service.PublishServiceArgs,
	reply *int) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.Services == nil {
		i.Services = make(map[string][]string)
	}
	switch args.Service {
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

type GetAddableNodeTypesArgs struct{ Site, NodeType string }

func (i *InfoService) GetAddableNodeTypes(args GetAddableNodeTypesArgs,
	types *[]string) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	*types = make([]string, 0)
	for nodeType, _ := range i.Config.NodeTypes {
		*types = append(*types, nodeType)
	}
	return nil
}

func (i *InfoService) GetNodeType(nodeTypeID string,
	ret *service.NodeType) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if nodeType, ok := i.Config.NodeTypes[nodeTypeID]; ok {
		*ret = nodeType
		return nil
	}
	return fmt.Errorf("Unknown node type %q", nodeTypeID)
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
