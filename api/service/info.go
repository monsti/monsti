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
	"fmt"
)

type InfoClient struct {
	Client
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
func (s *InfoClient) PublishService(service, path string) error {
	args := PublishServiceArgs{service, path}
	var reply int
	err := s.RPCClient.Call("Info.PublishService", args, &reply)
	if err != nil {
		return fmt.Errorf("service: Error calling PublishService: %v", err)
	}
	return nil
}

// FindDataService requests a data client.
func (s *InfoClient) FindDataService() (*DataClient, error) {
	var path string
	err := s.RPCClient.Call("Info.FindDataService", 0, &path)
	if err != nil {
		return nil, fmt.Errorf("service: Error calling FindDataService: %v", err)
	}
	service_ := NewDataClient()
	if err := service_.Connect(path); err != nil {
		return nil,
			fmt.Errorf("service: Could not establish connection to data service: %v",
				err)
	}
	return service_, nil
}

// FindMailService requests a mail client.
func (s *InfoClient) FindMailService() (*MailClient, error) {
	var path string
	err := s.RPCClient.Call("Info.FindMailService", 0, &path)
	if err != nil {
		return nil, fmt.Errorf("service: Error calling FindMailService: %v", err)
	}
	service_ := NewMailClient()
	if err := service_.Connect(path); err != nil {
		return nil,
			fmt.Errorf("service: Could not establish connection to mail service: %v",
				err)
	}
	return service_, nil
}

// FindNodeService requests a node client for the given node type.
func (s *InfoClient) FindNodeService(nodeType string) (*NodeClient,
	error) {
	var path string
	err := s.RPCClient.Call("Info.FindNodeService", nodeType, &path)
	if err != nil {
		return nil, fmt.Errorf("service: Error calling FindNodeService: %v", err)
	}
	service_ := NewNodeClient()
	if err := service_.Connect(path); err != nil {
		return nil,
			fmt.Errorf("service: Could not establish connection to node service: %v",
				err)
	}
	return service_, nil
}

// GetAddableNodeTypes returns the node types that may be added as child nodes
// to the given node type at the given website.
func (s *InfoClient) GetAddableNodeTypes(site, nodeType string) (types []string,
	err error) {
	args := struct{ Site, NodeType string }{site, nodeType}
	err = s.RPCClient.Call("Info.GetAddableNodeTypes", args, &types)
	if err != nil {
		err = fmt.Errorf("service: Error calling GetAddableNodeTypes: %v", err)
	}
	return
}

// NewInfoConnection establishes a new RPC connection to an INFO service.
//
// path is the unix domain socket path to the service.
func NewInfoConnection(path string) (*InfoClient, error) {
	var service InfoClient
	if err := service.Connect(path); err != nil {
		return nil,
			fmt.Errorf("service: Could not establish connection to info service: %v",
				err)
	}
	return &service, nil
}

func connectService() {

}
