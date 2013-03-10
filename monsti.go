/*
 Monsti is a simple and resource efficient CMS for low dynamic
 private and small business sites with mostly static pages and simple
 structure.

 This package implements the main application and http server.
*/
package main

import (
	"flag"
	"github.com/monsti/monsti-daemon/worker"
	"github.com/monsti/util/l10n"
	"github.com/monsti/util/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	logger := log.New(os.Stderr, "monsti", log.LstdFlags)
	flag.Parse()
	if flag.NArg() != 1 {
		logger.Fatalf("Usage: %v <config_directory>\n", filepath.Base(os.Args[0]))
	}
	cfgPath := flag.Arg(0)
	if !filepath.IsAbs(cfgPath) {
		wd, err := os.Getwd()
		if err != nil {
			panic("Could not get working directory: " + err.Error())
		}
		cfgPath = filepath.Join(wd, cfgPath)
	}
	settings, err := loadSettings(cfgPath)
	if err != nil {
		logger.Fatal("Could not load settings: ", err)
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
