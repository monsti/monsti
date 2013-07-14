// This file is part of Monsti, a web content management system.
// Copyright 2012-2013 Christian Neumann
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
// along with Monsti.  If not, see <http://www.gnu.org/licenses/>.

/*
 Monsti is a simple and resource efficient CMS for low dynamic
 private and small business sites with mostly static pages and simple
 structure.

 This package implements monsti's http daemon.
*/
package main

import (
	"flag"
	"pkg.monsti.org/service"
	"pkg.monsti.org/util"
	"pkg.monsti.org/util/template"
	"gitorious.org/monsti/gettext"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Settings for the application and the sites.
type settings struct {
	Monsti util.MonstiSettings
	// Listen is the host and port to listen for incoming HTTP connections.
	Listen string
}

func main() {
	logger := log.New(os.Stderr, "httpd ", log.LstdFlags)

	// Load configuration
	flag.Parse()
	if flag.NArg() != 1 {
		logger.Fatal("Expecting configuration path.")
	}
	cfgPath := util.GetConfigPath(flag.Arg(0))
	var settings settings
	if err := util.LoadModuleSettings("httpd", cfgPath, &settings); err != nil {
		logger.Fatal("Could not load settings: ", err)
	}
	if err := (&settings).Monsti.LoadSiteSettings(); err != nil {
		logger.Fatal("Could not load site settings: ", err)
	}

	gettext.DefaultLocales.Domain = "monsti-httpd"
	gettext.DefaultLocales.LocaleDir = settings.Monsti.GetLocalePath()

	// Connect to INFO service
	info, err := service.NewInfoConnection(
		settings.Monsti.GetServicePath(service.Info.String()))
	if err != nil {
		logger.Fatalf("Could not connect to INFO service: %v", err)
	}

	handler := nodeHandler{
		Info:     info,
		Renderer: template.Renderer{Root: settings.Monsti.GetTemplatesPath()},
		Settings: &settings,
		Log:      logger}
	http.Handle("/static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Monsti.GetStaticsPath()))))
	handler.Hosts = make(map[string]string)
	for site_title, site := range settings.Monsti.Sites {
		for _, host := range site.Hosts {
			handler.Hosts[host] = site_title
			http.Handle(host+"/site-static/", http.FileServer(http.Dir(
				filepath.Dir(settings.Monsti.GetSiteStaticsPath(site_title)))))
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
	logger.Printf("monsti-httpd is up and running. Listening on %q.", settings.Listen)
	<-c
}
