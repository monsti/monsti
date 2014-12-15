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
Package module implements utility functions to setup modules.
*/

package module

import (
	"flag"
	"log"
	"os"

	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util"
	mtemplate "pkg.monsti.org/monsti/api/util/template"
)

// ModuleContext is used as argument for the StartModule setup
// function.
type ModuleContext struct {
	Settings *util.MonstiSettings
	Sessions *service.SessionPool
	// Session is one allocated session of the pool that must not be
	// freed.
	Session  *service.Session
	Logger   *log.Logger
	Renderer *mtemplate.Renderer
}

// StartModule sets up the module with the given name.
func StartModule(name string, setup func(context *ModuleContext) error) {
	logger := log.New(os.Stderr, name+" ", log.LstdFlags)
	// Load configuration
	flag.Parse()
	if flag.NArg() != 1 {
		logger.Fatal("Expecting configuration path.")
	}
	cfgPath := util.GetConfigPath(flag.Arg(0))
	settings, err := util.LoadMonstiSettings(cfgPath)
	if err != nil {
		logger.Fatal("Could not load settings: ", err)
	}
	gettext.DefaultLocales.Domain = "monsti-" + name
	gettext.DefaultLocales.LocaleDir = settings.Directories.Locale

	renderer := mtemplate.Renderer{
		Root: settings.GetTemplatesPath()}
	monstiPath := settings.GetServicePath(service.MonstiService.String())
	sessions := service.NewSessionPool(1, monstiPath)

	session, err := sessions.New()
	if err != nil {
		logger.Fatalf("Could not get session: %v", err)
	}
	defer sessions.Free(session)
	if err := setup(&ModuleContext{
		settings, sessions, session, logger, &renderer,
	}); err != nil {
		logger.Fatalf("Could not setup module: %v", err)
	}
	if err := session.Monsti().ModuleInitDone("example-module"); err != nil {
		logger.Fatalf("Could not finish initialization: %v", err)
	}
	for {
		if err := session.Monsti().WaitSignal(); err != nil {
			logger.Printf("Could not wait for signal: %v", err)
		}
	}
}
