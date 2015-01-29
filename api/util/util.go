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
	"os"
	"path/filepath"
	"strings"
	"pkg.monsti.org/gettext"

	"launchpad.net/goyaml"
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

// Get the absolute config directory path using given command line path
// argument.
func GetConfigPath(arg string) (cfgPath string) {
	cfgPath = arg
	if !filepath.IsAbs(cfgPath) {
		wd, err := os.Getwd()
		if err != nil {
			panic("Could not get working directory: " + err.Error())
		}
		cfgPath = filepath.Join(wd, cfgPath)
	}
	return
}

type NestedMap map[string]interface{}

func (n NestedMap) Get(id string) interface{} {
	parts := strings.Split(id, ".")
	field := interface{}(map[string]interface{}(n))
	for _, part := range parts {
		var ok bool
		field, ok = field.(map[string]interface{})[part]
		if !ok {
			return nil
		}
	}
	return field
}

func (n NestedMap) Set(id string, value interface{}) {
	parts := strings.Split(id, ".")
	field := interface{}(map[string]interface{}(n))
	for _, part := range parts[:len(parts)-1] {
		next := field.(map[string]interface{})[part]
		if next == nil {
			next = make(map[string]interface{})
			field.(map[string]interface{})[part] = next
		}
		field = next
	}
	field.(map[string]interface{})[parts[len(parts)-1]] = value
}

// LanguageMap maps locales to translation strings.
type LanguageMap map[string]string

// Get returns the translation for the given locale. If the
// translation is not set, it returns the translation for the empty
// locale, i.e. "".
func (l LanguageMap) Get(locale string) string {
	if v, ok := l[locale]; ok {
		return v
	}
	return l[""]
}

// GenLanguageMap generates a language map for the given locales.
//
// The map will have an entry for each locale with the locale id being
// the key and the value being the gettext translation of msg (using
// pkg.monsti.org/gettext.DefaultLocales). The empty locale, i.e. "",
// is set to msg.
func GenLanguageMap(msg string, locales []string) LanguageMap {
	ret := make(LanguageMap)
	ret[""] = msg
	for _, lang := range locales {
		G, _, _, _ := gettext.DefaultLocales.Use("", lang)
		ret[lang] = G(msg)
	}
	return ret
}
