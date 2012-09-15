// Utilities for monsti and node implementations.
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
