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

/*
 Monsti is a simple and resource efficient CMS.

 This package implements the main daemon which starts and observes modules.
*/
package main

import (
	"flag"
	"github.com/monsti/service"
	"github.com/monsti/util"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type settings struct {
	Monsti util.MonstiSettings
	// List of modules to be activated.
	Modules []string
}

// moduleLog is a Writer used to log module messages on stderr.
type moduleLog struct {
	Type string
	Log  *log.Logger
}

func (s moduleLog) Write(p []byte) (int, error) {
	s.Log.Println(s.Type, "on stderr:", string(p))
	return len(p), nil
}

func main() {
	logger := log.New(os.Stderr, "monsti ", log.LstdFlags)

	// Load configuration
	flag.Parse()
	if flag.NArg() != 1 {
		logger.Fatalf("Usage: %v <config_directory>\n",
			filepath.Base(os.Args[0]))
	}
	cfgPath := util.GetConfigPath(flag.Arg(0))
	var settings settings
	if err := util.LoadModuleSettings("daemon", cfgPath, &settings); err != nil {
		logger.Fatal("Could not load settings: ", err)
	}

	// Start own INFO service
	var waitGroup sync.WaitGroup
	logger.Println("Starting INFO service")
	waitGroup.Add(1)
	infoPath := "monsti-info"
	go func() {
		defer waitGroup.Done()
		var provider service.Provider
		info := new(InfoService)
		provider.Logger = logger
		if err := provider.Serve(infoPath, "Info", info); err != nil {
			logger.Fatalf("Could not start INFO service: %v", err)
		}
	}()

	// Start modules
	for _, module := range settings.Modules {
		logger.Println("Starting module", module)
		executable := "monsti-" + module
		cmd := exec.Command(executable, settings.Monsti.Directories.Config,
			infoPath)
		cmd.Stderr = moduleLog{module, logger}
		go func() {
			if err := cmd.Run(); err != nil {
				logger.Fatalf("Module %q failed: %v", module, err)
			}
		}()
	}

	logger.Println("Monsti is up and running!")
	waitGroup.Wait()
	logger.Println("Monsti is shutting down.")
}
