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
	"sync"
)

// ActionHandler processes incoming requests for some node action.
type ActionHandler func(req Request, res *Response,
	infoServ *InfoClient)

// NodeTypeHandler handles requests for some node type.
type NodeTypeHandler struct {
	Name                   string
	EditAction, ViewAction ActionHandler
}

// NodeProvider provides Node services for one ore more content types.
type NodeProvider struct {
	// Logger to be used
	Logger *log.Logger
	// Info service to be used
	Info  *InfoClient
	types map[string]*NodeTypeHandler
}

func NewNodeProvider(logger *log.Logger, info *InfoClient) *NodeProvider {
	p := NodeProvider{
		Logger: logger,
		Info:   info,
		types:  make(map[string]*NodeTypeHandler),
	}
	return &p
}

// AddNodeType adds a handler for a node type to be served.
func (p *NodeProvider) AddNodeType(h *NodeTypeHandler) {
	p.types[h.Name] = h
}

// Serve registers and serves the node types at the given Node service path.
func (p *NodeProvider) Serve(path string) error {
	// Start own NODE service
	var waitGroup sync.WaitGroup
	p.Logger.Println("Starting Node service")
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		var provider Provider
		var node_ nodeService
		node_.Provider = p
		node_.Info = p.Info
		provider.Logger = p.Logger
		if err := provider.Serve(path, "Node", &node_); err != nil {
			p.Logger.Printf("node: Could not start Node service: %v", err)
		}
	}()

	if err := p.Info.PublishService("Node", path); err != nil {
		return fmt.Errorf("node: Could not publish node service: %v", err)
	}
	waitGroup.Wait()
	return nil
}

type nodeService struct {
	Provider *NodeProvider
	Info     *InfoClient
}

func (i nodeService) Request(req Request,
	reply *Response) error {
	nodeType := req.Node.Type
	var f ActionHandler
	switch req.Action {
	case "edit":
		f = i.Provider.types[nodeType].EditAction
	default:
		f = i.Provider.types[nodeType].ViewAction
	}
	f(req, reply, i.Info)
	return nil
}

func (i *nodeService) GetNodeTypes(req int,
	reply *[]string) error {
	*reply = make([]string, 0)
	for key, _ := range i.Provider.types {
		*reply = append(*reply, key)
	}
	return nil
}
