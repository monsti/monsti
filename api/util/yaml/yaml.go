// This file is part of Monsti.
// Copyright 2012-2015 Christian Neumann

// Monsti is free software: you can redistribute it and/or modify it under
// the terms of the GNU Lesser General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.

// Monsti is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more
// details.

// You should have received a copy of the GNU Lesser General Public License
// along with Monsti. If not, see <http://www.gnu.org/licenses/>.

/*
Package yaml implements utility functions to work with YAML.
*/
package yaml

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"launchpad.net/goyaml"
)

// Parse loads the given YAML file and unmarshals it into the given
// ojbect.
func Parse(path string, out interface{}) error {
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
