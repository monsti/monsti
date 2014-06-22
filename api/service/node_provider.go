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

/*

// ActionHandler processes incoming requests for some node action.
type ActionHandler func(req Request, res *Response, s *Session) error

// NodeTypeHandler handles requests for some node type.
type NodeTypeHandler struct {
	Name                                string
	EditAction, ViewAction, EmbedAction ActionHandler
}

// NodeProvider provides Node services for one ore more content types.
type NodeProvider struct {
	// Logger to be used
	Logger *log.Logger
	pool   *SessionPool
	types  map[string]*NodeTypeHandler
}

func NewNodeProvider(logger *log.Logger, infoPath string) *NodeProvider {
	p := NodeProvider{
		logger, NewSessionPool(1, infoPath),
		make(map[string]*NodeTypeHandler),
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
	p.Logger.Println("Setting up Node service")
	var node_ nodeService
	node_.Provider = p
	node_.Pool = p.pool
	provider := NewProvider("Node", &node_)
	provider.Logger = p.Logger
	if err := provider.Listen(path); err != nil {
		p.Logger.Printf("service: Could not start Node service: %v", err)
	}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		if err := provider.Accept(); err != nil {
			p.Logger.Printf("service: Could not accept at Node service: %v", err)
		}
	}()
	s, err := p.pool.New()
	if err != nil {
		return fmt.Errorf("service: Could not get session: %v", err)
	}
	defer p.pool.Free(s)
	if err := s.Info().PublishService("Node", path); err != nil {
		return fmt.Errorf("service: Could not publish Node service: %v", err)
	}
	waitGroup.Wait()
	return nil
}

type nodeService struct {
	Provider *NodeProvider
	Pool     *SessionPool
}

/*
func (i nodeService) Request(req Request, reply *Response) error {
	nodeType := req.Node.Type
	var f ActionHandler
	session, err := i.Pool.New()
	if err != nil {
		return fmt.Errorf("service: Could not get session: %v", err)
	}
	defer i.Pool.Free(session)
	switch req.Action {
	case EditAction:
		f = i.Provider.types[nodeType].EditAction
	}
	if f == nil {
		f = i.Provider.types[nodeType].ViewAction
	}
	return f(req, reply, session)
}

func (i *nodeService) GetNodeTypes(req int,
	reply *[]string) error {
	*reply = make([]string, 0)
	for key, _ := range i.Provider.types {
		*reply = append(*reply, key)
	}
	return nil
}

*/
