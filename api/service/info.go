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
func (s *MonstiClient) PublishService(service, path string) error {
	args := PublishServiceArgs{service, path}
	var reply int
	err := s.RPCClient.Call("Monsti.PublishService", args, &reply)
	if err != nil {
		return fmt.Errorf("service: Error calling PublishService: %v", err)
	}
	return nil
}

/*
// FindDataService requests a data client.
func (s *MonstiClient) FindDataService() (*MonstiClient, error) {
	var path string
	err := s.RPCClient.Call("Monsti.FindDataService", 0, &path)
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
*/
