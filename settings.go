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
