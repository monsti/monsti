// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann
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
	"pkg.monsti.org/monsti/api/util/i18n"
	"pkg.monsti.org/monsti/api/util/module"
)

var availableLocales = []string{"de", "en", "nl"}

func setup(c *module.ModuleContext) error {
	G := func(in string) string { return in }
	m := c.Session.Monsti()

	// Register a new node type
	nodeType := service.NodeType{
		Id:   "example.ExampleType",
		Name: i18n.GenLanguageMap(G("Example node type"), availableLocales),
		Fields: []*service.FieldConfig{
			// core.Title and core.Body are already known to the system,
			// just specify their IDs to include them.
			{Id: "core.Title"},
			{Id: "core.Body"},
			{
				Id:   "example.Foo",
				Name: i18n.GenLanguageMap(G("Foo"), availableLocales),
				Type: new(service.TextFieldType),
			},
			{
				Id:   "example.Bar",
				Name: i18n.GenLanguageMap(G("Bar"), availableLocales),
				Type: new(service.DateTimeFieldType),
			},
		},
	}
	if err := m.RegisterNodeType(&nodeType); err != nil {
		c.Logger.Fatalf("Could not register %q node type: %v", nodeType.Id, err)
	}

	nodeType = service.NodeType{
		Id:   "example.Embed",
		Name: i18n.GenLanguageMap(G("Embed example"), availableLocales),
		Fields: []*service.FieldConfig{
			// core.Title and core.Body are already known to the system,
			// just specify their IDs to include them.
			{Id: "core.Title"},
			{Id: "core.Body"},
		},
	}
	if err := m.RegisterNodeType(&nodeType); err != nil {
		c.Logger.Fatalf("Could not register %q node type: %v", nodeType.Id, err)
	}

	nodeType = service.NodeType{
		Id:   "example.Fields",
		Name: i18n.GenLanguageMap(G("Fields example"), availableLocales),
		Fields: []*service.FieldConfig{
			// core.Title and core.Body are already known to the system,
			// just specify their IDs to include them.
			{Id: "core.Title"},
			{Id: "core.Body"},
			{
				Id:   "example.Bool",
				Name: i18n.GenLanguageMap(G("Bool"), availableLocales),
				Type: new(service.BoolFieldType),
			},
			{
				Id:   "example.DateTime",
				Name: i18n.GenLanguageMap(G("DateTime"), availableLocales),
				Type: new(service.DateTimeFieldType),
			},
			{
				Id:   "example.HTMLArea",
				Name: i18n.GenLanguageMap(G("HTML"), availableLocales),
				Type: new(service.HTMLFieldType),
			},
			{
				Id:     "example.Hidden",
				Name:   i18n.GenLanguageMap(G("Hidden"), availableLocales),
				Hidden: true,
				Type:   new(service.TextFieldType),
			},
			{
				Id:   "example.Text",
				Name: i18n.GenLanguageMap(G("Text"), availableLocales),
				Type: new(service.TextFieldType),
			},
			{
				Id:   "example.TextList",
				Name: i18n.GenLanguageMap(G("Text list"), availableLocales),
				Type: &service.ListFieldType{
					ElementType: new(service.TextFieldType),
					AddLabel:    i18n.GenLanguageMap(G("Add text entry"), availableLocales),
					RemoveLabel: i18n.GenLanguageMap(G("Remove entry"), availableLocales),
				},
			},
			{
				Id:     "example.Map",
				Hidden: true,
				Name:   i18n.GenLanguageMap(G("Map"), availableLocales),
				Type:   &service.MapFieldType{new(service.TextFieldType)},
			},
			{
				Id:     "example.Combined",
				Hidden: true,
				Name:   i18n.GenLanguageMap(G("Combined"), availableLocales),
				Type: &service.CombinedFieldType{
					map[string]service.FieldConfig{
						"Text": {
							Name: i18n.GenLanguageMap(G("Text"), availableLocales),
							Type: new(service.TextFieldType),
						},
						"Bool": {
							Name: i18n.GenLanguageMap(G("Bool"), availableLocales),
							Type: new(service.BoolFieldType),
						},
					}},
			},
		},
	}
	if err := m.RegisterNodeType(&nodeType); err != nil {
		c.Logger.Fatalf("Could not register %q node type: %v", nodeType.Id, err)
	}

	// Add a signal handler
	handler := service.NewNodeContextHandler(c.Sessions,
		func(id uint, session *service.Session, nodeType string,
			embedNode *service.EmbedNode) (
			map[string][]byte, *service.CacheMods, error) {
			if nodeType == "example.ExampleType" {
				req, err := session.Monsti().GetRequest(id)
				if err != nil || req == nil {
					return nil, nil, fmt.Errorf("Could not get request: %v", err)
				}
				return map[string][]byte{
					"SignalFoo": []byte(fmt.Sprintf("Hello Signal! Site name: %v", req.Site)),
				}, nil, nil
			}
			return nil, nil, nil
		})
	if err := m.AddSignalHandler(handler); err != nil {
		c.Logger.Fatalf("Could not add signal handler: %v", err)
	}

	return nil
}

func main() {
	module.StartModule("example-module", setup)
}
