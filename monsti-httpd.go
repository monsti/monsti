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
	"github.com/monsti/service"
	"github.com/monsti/util"
	"github.com/monsti/util/l10n"
	"github.com/monsti/util/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

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
	if err := (&settings).LoadSiteSettings(); err != nil {
		logger.Fatal("Could not load site settings: ", err)
	}

	l10n.DefaultSettings.Domain = "monsti-httpd"
	l10n.DefaultSettings.Directory = settings.Monsti.GetLocalePath()

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
	for site_title, site := range settings.Sites {
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
