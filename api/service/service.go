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
	"net"
	"net/rpc"
	"os"
	"sync"
)

var lastConnectionId uint
var lastConnectionMutex sync.Mutex

// getConnectionId returns a new connection ID.
func getConnectionId() string {
	lastConnectionMutex.Lock()
	defer lastConnectionMutex.Unlock()
	lastConnectionId += 1
	return fmt.Sprintf("%v#%v", os.Getpid(), lastConnectionId)
}

type Type uint

// Monsti service types.
const (
	MonstiService Type = iota
	/*
		InfoService Type = iota
		DataService
		LoginService
		NodeService
		MailService
	*/
)

func (t Type) String() string {
	serviceNames := [...]string{
		"Monsti"}
	return serviceNames[t]
}

// Client represents the rpc connection to a service.
type Client struct {
	RPCClient *rpc.Client
	// Error holds the last error if any.
	Error error
	// Id is a unique identifier for this client.
	Id string
}

// Connect establishes a new RPC connection to the given service.
//
// path is the unix domain socket path to the service.
func (s *Client) Connect(path string) error {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return err
	}
	s.Id = getConnectionId()
	// TODO Fix id
	s.RPCClient = rpc.NewClient(conn)
	return nil
}

// Close closes the client's RPC connection.
func (s *Client) Close() error {
	return s.RPCClient.Close()
}
