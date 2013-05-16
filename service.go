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
	"github.com/monsti/service/info"
	"github.com/monsti/service/node"
)

type InfoService struct {
	// Services maps service names to service paths
	Services map[string][]string
	// NodeTypes maps node types to service paths
	NodeTypes map[string][]string
}

func (i *InfoService) PublishService(args info.PublishServiceArgs,
	reply *int) error {
	if i.Services == nil {
		i.Services = make(map[string][]string)
	}
	switch args.Service {
	case "Node":
		if i.Services[args.Service] == nil {
			i.Services[args.Service] = make([]string, 0)
		}
		i.Services[args.Service] = append(i.Services[args.Service], args.Path)

		if i.NodeTypes == nil {
			i.NodeTypes = make(map[string][]string)
		}
		nodeServ := node.NewService()
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
	default:
		return fmt.Errorf("Unknown service type %v", args.Service)
	}
	return nil
}

func (i *InfoService) FindNodeService(nodeType string, path *string) error {
	if len(i.NodeTypes[nodeType]) == 0 {
		return fmt.Errorf("Unknown node type %v", nodeType)
	}
	*path = i.NodeTypes[nodeType][0]
	return nil
}
