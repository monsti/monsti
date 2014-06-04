// This file is part of Monsti, a web content management system.
// Copyright 2012-2014 Christian Neumann
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
// along with Monsti.  If not, see <http:#www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util"
)

type SingleConfig struct {
	Namespace string
	NodeTypes []service.NodeType
}

type Config struct {
	NodeTypes map[string]service.NodeType
}

// loadConfig parses all configuration files in the given directory
// and returns the application configuration.
func loadConfig(configDir string) (*Config, error) {
	configPath := filepath.Join(configDir)
	configFiles, err := ioutil.ReadDir(configPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read config directory: %v", err)
	}
	var config Config
	config.NodeTypes = make(map[string]service.NodeType, 0)
	for _, configFile := range configFiles {
		configName := configFile.Name()
		configPath := filepath.Join(configPath, configName)
		_, err := os.Stat(configPath)
		if err != nil {
			continue
		}
		var singleConfig SingleConfig
		err = util.ParseYAML(configPath, &singleConfig)
		if err != nil {
			return nil, fmt.Errorf("Could not load config %q: %v", configName, err)
		}
		for _, node := range singleConfig.NodeTypes {
			node.Id = singleConfig.Namespace + "." + node.Id
			for i, field := range node.Fields {
				node.Fields[i].Id = singleConfig.Namespace + "." + field.Id
			}
			config.NodeTypes[node.Id] = node
		}
	}
	return &config, nil
}
