// This file is part of Monsti, a web content management system.
// Copyright 2012-2014 Christian Neumann
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

import "fmt"

// MonstiClient represents the RPC connection to the Monsti service.
type MonstiClient struct {
	Client
}

// NewMonstiConnection establishes a new RPC connection to a Monsti service.
//
// path is the unix domain socket path to the service.
func NewMonstiConnection(path string) (*MonstiClient, error) {
	var service MonstiClient
	if err := service.Connect(path); err != nil {
		return nil,
			fmt.Errorf("service: Could not establish connection to Monsti service: %v",
				err)
	}
	return &service, nil
}

// ModuleInitDone tells Monsti that the given module has finished its
// initialization. Monsti won't finish its startup until all modules
// called this method.
func (s *MonstiClient) ModuleInitDone(module string) error {
	if s.Error != nil {
		return s.Error
	}
	err := s.RPCClient.Call("Monsti.ModuleInitDone", module, new(int))
	if err != nil {
		return fmt.Errorf("service: ModuleInitDone error: %v", err)
	}
	return nil
}
