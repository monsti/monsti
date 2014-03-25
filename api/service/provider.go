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
	"log"
	"net"
	"net/rpc"
	"os"
)

type Provider struct {
	Logger   *log.Logger
	listener net.Listener
	service  string
	rcvr     interface{}
}

// NewProvider returns a new Provider for the given service and using
// rcvr as RPC receiver.
func NewProvider(service string, rcvr interface{}) (p *Provider) {
	return &Provider{service: service, rcvr: rcvr}
}

// Listen starts listening on the given unix domain socket path for
// incoming rpc connections. Be sure to call Accept after that.
func (p *Provider) Listen(path string) error {
	os.Remove(path)
	var err error
	p.listener, err = net.Listen("unix", path)
	if err != nil {
		return fmt.Errorf("service: Could not listen on unix domain socket %q: %v",
			path, err)
	}
	return nil
}

// Accept starts accepting incoming connection and setting up RPC for the client.
func (p *Provider) Accept() error {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			return fmt.Errorf("service: Could not accept connection for %q: %v",
				p.service, err)
		}
		server := rpc.NewServer()
		if err = server.RegisterName(p.service, p.rcvr); err != nil {
			return fmt.Errorf("service: Could not register RPC methods: %v",
				err.Error())
		}
		go func() {
			server.ServeConn(conn)
			conn.Close()
		}()
	}
	return nil
}
