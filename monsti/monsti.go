/*
 Monsti is a simple and resource efficient CMS for low dynamic
 private and small business sites with mostly static pages and simple
 structure.

 This package implements the main application and http server.
*/
package main

import (
	"github.com/monsti/l10n"
	"github.com/monsti/template"
	"github.com/monsti/util"
	"github.com/monsti/worker"
	"flag"
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
		// HTML Templates to be used instead of monsti's ones.
		Templates string
	}
}

// Settings for the application and the sites.
type settings struct {
	// Settings for sending mail (outgoing SMTP).
	Mail struct {
		// Host may be specified as address:port
		Host, Username, Password string
	}
	// Listen is the host and port to listen for incoming HTTP connections.
	Listen string
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
	logger := log.New(os.Stderr, "monsti", log.LstdFlags)
	flag.Parse()
	if flag.NArg() != 1 {
		logger.Fatalf("Usage: %v <config_directory>\n", filepath.Base(os.Args[0]))
	}
	cfgPath := util.GetConfigPath("monsti", flag.Arg(0))
	var settings settings
	err := util.ParseYAML(cfgPath, &settings)
	if err != nil {
		logger.Fatal("Could not load configuration file: ", err)
	}
	l10n.DefaultSettings.Domain = "monsti"
	l10n.DefaultSettings.Directory = settings.Directories.Locales
	handler := nodeHandler{
		Renderer:   template.Renderer{Root: settings.Directories.Templates},
		Settings:   settings,
		NodeQueues: make(map[string]chan worker.Ticket),
		Log:        logger}
	for _, ntype := range settings.NodeTypes {
		handler.AddNodeProcess(ntype, logger)
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
	c := make(chan int)
	go func() {
		if err := http.ListenAndServe(settings.Listen, nil); err != nil {
			logger.Fatal("HTTP Listener failed: ", err)
		}
		c <- 1
	}()
	logger.Printf("Monsti is up and running. Listening on %q.", settings.Listen)
	<-c
}
