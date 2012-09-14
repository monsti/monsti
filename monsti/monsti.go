/*
 Brassica is a simple and resource efficient CMS for low dynamic
 private and small business sites with mostly static pages and simple
 structure.

 monsti/monsti-serve contains a command to start a httpd.
*/
package main

import (
	"code.google.com/p/gorilla/schema"
	"net/smtp"
	"os"
	"path/filepath"
)

var schemaDecoder = schema.NewDecoder()

// Settings for the application and the site.
type settings struct {
	MailAuth smtp.Auth

	MailServer string

	// Path to the data directory.
	Root string

	// Path to the static files.
	Statics string

	// Path to the site specific static files.
	SiteStatics string

	// Path to the template directory.
	Templates string
}

// GetSettings returns application and site settings.
func getSettings() settings {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	settings := settings{
		MailServer:  "localhost:12345",
		MailAuth:    smtp.PlainAuth("", "joe", "secret!", "host"),
		Root:        wd,
		Statics:     filepath.Join(filepath.Dir(wd), "static"),
		SiteStatics: filepath.Join(filepath.Dir(wd), "site-static"),
		Templates:   filepath.Join(filepath.Dir(wd), "templates")}
	return settings
}
