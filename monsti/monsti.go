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
	// Settings for sending mail (outgoing SMTP).
	Mail struct {
		// Host may be specified as address:port
		Host, Username, Password string
	}

	// Absolute paths to used directories.
	Directories struct {
		// Site content
		Data string
		// Monsti's static files
		Statics string
		// Site specific static files
		SiteStatics string
		// HTML Templates
		Templates string
		// Locales, i.e. the gettext machine objects (.mo)
		Locales string
	}

	// Site title
	Title string

	// Name and email address of site owner.
	//
	// The owner's address is used as recipient of contact form submissions.
	Owner struct {
		Name, Email string
	}

	// List of node types to be activated.
	NodeTypes []string
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
