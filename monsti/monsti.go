/*
 Brassica is a simple and resource efficient CMS for low dynamic
 private and small business sites with mostly static pages and simple
 structure.

 monsti/monsti-serve contains a command to start a httpd.
*/
package main

import (
	"code.google.com/p/gorilla/schema"
	"io/ioutil"
	"launchpad.net/goyaml"
	"path/filepath"
)

var schemaDecoder = schema.NewDecoder()

// Settings for the application and the site.
type settings struct {
	Mail struct {
		Host, Username, Password string
	}

	// Path to the data directory.
	Root string

	// Path to the static files.
	Statics string

	// Path to the site specific static files.
	SiteStatics string

	// Path to the template directory.
	Templates string
}

// GetSettings loads application and site settings from given configuration
// directory.
func getSettings(path string) settings {
	path = filepath.Join(path, "monsti.yaml")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic("Could not load configuration: " + err.Error())
	}
	var s settings
	goyaml.Unmarshal(content, &s)
	return s
}
