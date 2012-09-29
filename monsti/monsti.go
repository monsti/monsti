/*
 Brassica is a simple and resource efficient CMS for low dynamic
 private and small business sites with mostly static pages and simple
 structure.

 This package implements the main application and http server.
*/
package main

import (
	"datenkarussell.de/monsti/template"
	"datenkarussell.de/monsti/util"
	"datenkarussell.de/monsti/worker"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Settings for the application and the site.
type settings struct {
	// Settings for sending mail (outgoing SMTP).
	Mail struct {
		// Host may be specified as address:port
		Host, Username, Password string
	}
	// Absolute paths to used directories.
	Directories struct {
		// Config files
		Config string
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
	// Key to authenticate session cookies.
	SessionAuthKey string
}

func main() {
	log.SetPrefix("monsti ")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Printf("Usage: %v <config_directory>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	cfgPath := util.GetConfigPath("monsti", flag.Arg(0))
	var settings settings
	err := util.ParseYAML(cfgPath, &settings)
	if err != nil {
		fmt.Println("Could not load configuration file: " + err.Error())
		os.Exit(1)
	}
	settings.Directories.Config = filepath.Dir(cfgPath)
	handler := nodeHandler{
		Renderer:   template.Renderer{Root: settings.Directories.Templates},
		Settings:   settings,
		NodeQueues: make(map[string]chan worker.Ticket)}
	for _, ntype := range settings.NodeTypes {
		handler.AddNodeProcess(ntype)
	}
	http.Handle("/static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Directories.Statics))))
	http.Handle("/site-static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Directories.SiteStatics))))
	http.Handle("/", &handler)
	http.ListenAndServe(":8080", nil)
}
