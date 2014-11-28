// This file is part of Monsti, a web content management system.
// Copyright 2014 Christian Neumann
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

	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util"
	"pkg.monsti.org/monsti/api/util/module"
)

var availableLocales = []string{"de", "en"}

func setup(c *module.ModuleContext) error {
	G := func(in string) string { return in }
	m := c.Session.Monsti()

	// Register a new node type
	nodeType := service.NodeType{
		Id:        "example.ExampleType",
		AddableTo: []string{"."},
		Name:      util.GenLanguageMap(G("Example node type"), availableLocales),
		Fields: []*service.NodeField{
			// core.Title and core.Body are already known to the system,
			// just specify their IDs to include them.
			{Id: "core.Title"},
			{Id: "core.Body"},
			{
				Id:   "example.Foo",
				Name: util.GenLanguageMap(G("Foo"), availableLocales),
				Type: "Text",
			},
			{
				Id:   "example.Bar",
				Name: util.GenLanguageMap(G("Bar"), availableLocales),
				Type: "DateTime",
			},
		},
	}
	if err := m.RegisterNodeType(&nodeType); err != nil {
		c.Logger.Fatalf("Could not register %q node type: %v", nodeType.Id, err)
	}

	// Add a signal handler
	handler := service.NewNodeContextHandler(
		func(id uint, nodeType string, embedNode *service.EmbedNode) map[string]string {
			session, err := c.Sessions.New()
			if err != nil {
				c.Logger.Fatalf("Could not get session: %v", err)
				return nil
			}
			defer c.Sessions.Free(session)
			if nodeType == "example.ExampleType" {
				req, err := session.Monsti().GetRequest(id)
				if err != nil || req == nil {
					c.Logger.Fatalf("Could not get request: %v", err)
				}
				return map[string]string{
					"SignalFoo": fmt.Sprintf("Hello Signal! Site name: %v", req.Site),
				}
			}
			return nil
		})
	if err := m.AddSignalHandler(handler); err != nil {
		c.Logger.Fatalf("Could not add signal handler: %v", err)
	}

	return nil
}

func main() {
	module.StartModule("example-module", setup)
}
