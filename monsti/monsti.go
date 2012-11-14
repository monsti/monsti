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
	"github.com/chrneumann/g5t"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Site configuration.
type site struct {
	// Name of the site for internal use.
	Name string
	// Title as used in HTML head.
	Title string
	// The hosts which should deliver this site.
	Hosts []string
	// Name and email address of site owner.
	//
	// The owner's address is used as recipient of contact form submissions.
	Owner struct {
		Name, Email string
	}
	// Key to authenticate session cookies.
	SessionAuthKey string
	// Locale used to translate monsti's web interface.
	Locale string
	// Absolute paths to site specific directories.
	Directories struct {
		// Site content
		Data string
		// Site specific static files
		Statics string
	}
}

// Settings for the application and the sites.
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
		// Monsti's static files
		Statics string
		// HTML Templates
		Templates string
		// Locales, i.e. the gettext machine objects (.mo)
		Locales string
	}
	// List of node types to be activated.
	NodeTypes []string
	// Sites hosted by this monsti instance.
	Sites map[string]site
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
	err = g5t.Setup("monsti", settings.Directories.Locales, "de", g5t.GettextParser)
	if err != nil {
		panic("Could not setup gettext: " + err.Error())
	}
	handler := nodeHandler{
		Renderer:   template.Renderer{Root: settings.Directories.Templates},
		Settings:   settings,
		NodeQueues: make(map[string]chan worker.Ticket)}
	for _, ntype := range settings.NodeTypes {
		handler.AddNodeProcess(ntype)
	}
	http.Handle("/static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Directories.Statics))))
	handler.Hosts = make(map[string]string)
	for site_title, site := range settings.Sites {
		for _, host := range site.Hosts {
			handler.Hosts[host] = site_title
			http.Handle(host+"/site-static/", http.FileServer(http.Dir(
				filepath.Dir(site.Directories.Statics))))
		}
	}
	http.Handle("/", &handler)
	http.ListenAndServe(":8080", nil)
}
