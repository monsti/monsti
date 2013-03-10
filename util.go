// This file is part of monsti/util.
// Copyright 2012 Christian Neumann

// monsti/util is free software: you can redistribute it and/or modify it under
// the terms of the GNU Lesser General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.

// monsti/util is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more
// details.

// You should have received a copy of the GNU Lesser General Public License
// along with monsti/util. If not, see <http://www.gnu.org/licenses/>.

/*
Package util implements utility functions for Monsti and Monsti content type
workers.
*/
package util

import (
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"path/filepath"
)

// ParseYAML loads the given YAML file and unmarshals it into the given
// ojbect.
func ParseYAML(path string, out interface{}) error {
	path = filepath.Join(path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Could not load YAML file to be parsed: %v", err.Error())
	}
	err = goyaml.Unmarshal(content, out)
	if err != nil {
		return fmt.Errorf("Could not unmarshal YAML file: %v", err.Error())
	}
	return nil
}

// MakeAbsolute converts a possibly relative path to an absolute one using the
// given root.
func MakeAbsolute(path *string, root string) {
	if !filepath.IsAbs(*path) {
		*path = filepath.Join(root, *path)
	}
}

// Get config file path for the component using given command line path
// argument.
func GetConfigPath(component, arg string) (cfgPath string) {
	cfgPath = arg
	if !filepath.IsAbs(cfgPath) {
		wd, err := os.Getwd()
		if err != nil {
			panic("Could not get working directory: " + err.Error())
		}
		cfgPath = filepath.Join(wd, cfgPath)
	}
	cfgPath = filepath.Join(cfgPath, component+".yaml")
	return
}
