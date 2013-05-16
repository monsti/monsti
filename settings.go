// This file is part of monsti/monsti-daemon.
// Copyright 2012-2013 Christian Neumann

package main

import (
	"fmt"
	"github.com/monsti/util"
	"path/filepath"
)

// Settings for the application and the sites.
type settings struct {
	Directories struct {
		// Config files
		Config string
	}
	// List of modules to be activated.
	Modules []string
}

// loadSettings loads daemon settings from the given configuration directory.
//
// The configuration directory path must be absolute or relative to the working
// directory.
func loadSettings(cfgPath string) (*settings, error) {
	settings := new(settings)
	err := util.ParseYAML(filepath.Join(cfgPath, "monsti.yaml"), settings)
	if err != nil {
		return nil, fmt.Errorf("Could not load main configuration file: %v", err)
	}
	settings.Directories.Config = cfgPath
	return settings, nil
}
