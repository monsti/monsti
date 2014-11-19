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

/*
 Monsti is a simple and resource efficient CMS.

 This package implements the document node type.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util"
	"pkg.monsti.org/monsti/api/util/template"
)

var logger *log.Logger
var renderer template.Renderer

var availableLocales = []string{"de", "en"}

func main() {
	// Setup module

	logger = log.New(os.Stderr, "example-modu]e ", log.LstdFlags)
	flag.Parse()
	if flag.NArg() != 1 {
		logger.Fatal("Expecting configuration path.")
	}
	cfgPath := util.GetConfigPath(flag.Arg(0))
	settings, err := util.LoadMonstiSettings(cfgPath)
	if err != nil {
		logger.Fatal("Could not load settings: ", err)
	}

	gettext.DefaultLocales.Domain = "monsti-example-module"
	gettext.DefaultLocales.LocaleDir = settings.Directories.Locale

	// Setup API
	monstiPath := settings.GetServicePath(service.MonstiService.String())
	sessions := service.NewSessionPool(1, monstiPath)
	session, err := sessions.New()
	if err != nil {
		logger.Fatalf("Could not get session: %v", err)
	}
	defer sessions.Free(session)

	// Register a new node type
	G := func(in string) string { return in }
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
	if err := session.Monsti().RegisterNodeType(&nodeType); err != nil {
		logger.Fatalf("Could not register %q node type: %v", nodeType.Id, err)
	}

	// Add a signal handler
	handler := service.NewNodeContextHandler(
		func(id uint, nodeType string, embedNode *service.EmbedNode) map[string]string {
			if nodeType == "example.ExampleType" {
				req, err := session.Monsti().GetRequest(id)
				if err != nil || req == nil {
					logger.Fatalf("Could not get request: %v", err)
				}
				return map[string]string{
					"SignalFoo": fmt.Sprintf("Hello Signal! Site name: %v", req.Site),
				}
			}
			return nil
		})
	if err := session.Monsti().AddSignalHandler(handler); err != nil {
		logger.Fatalf("Could not add signal handler: %v", err)
	}

	// At the end of the initialization, every module has to call
	// ModuleInitDone. Monsti won't complete its startup until all
	// modules have called this method.
	if err := session.Monsti().ModuleInitDone("example-module"); err != nil {
		logger.Fatalf("Could not finish initialization: %v", err)
	}

	for {
		if err := session.Monsti().WaitSignal(); err != nil {
			logger.Fatalf("Could not wait for signal: %v", err)
		}
	}
}
