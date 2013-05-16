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

package info

import (
	"fmt"
	"github.com/monsti/service"
	"github.com/monsti/service/node"
	"log"
)

type Service struct {
	service.Service
}

// PublishServiceArgs are the arguments provided by the caller of
// PublishService.
type PublishServiceArgs struct {
	Service, Path string
}

// PublishService informs the INFO service about a new service.
//
// service is the identifier of the service
// path is the path to the unix domain socket of the service
//
// If the data does not exist, return null length []byte.
func (s *Service) PublishService(service, path string) error {
	args := PublishServiceArgs{service, path}
	var reply int
	err := s.Client.Call("Info.PublishService", args, &reply)
	if err != nil {
		return fmt.Errorf("info: Could not call PublishService on service: %v", err)
	}
	return nil
}

// FindNodeService requests a node service for the given node type.
func (s *Service) FindNodeService(nodeType string) (*node.Service,
	error) {
	var path string
	err := s.Client.Call("Info.FindNodeService", nodeType, &path)
	if err != nil {
		return nil, fmt.Errorf("info: Could not call FindNodeService on service: %v", err)
	}
	service_ := node.NewService()
	if err := service_.Connect(path); err != nil {
		return nil,
			fmt.Errorf("Could not establish connection to node service: %v",
				err)
	}
	return service_, nil
}

// NewConnection establishes a new RPC connection to an INFO service.
//
// path is the unix domain socket path to the service.
// logger is a Logger to be used for logging messages by the connection.
func NewConnection(path string, logger *log.Logger) (*Service, error) {
	var service Service
	if err := service.Connect(path); err != nil {
		return nil,
			fmt.Errorf("Could not establish connection to info service: %v",
				err)
	}
	return &service, nil
}

func connectService() {

}
