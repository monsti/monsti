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

package main

import (
	"fmt"
	"github.com/monsti/service"
	"github.com/monsti/util/template"
	htmlT "html/template"
	"path"
	"strings"
)

// Master template render flags.
type masterTmplFlags uint32

const (
	EDIT_VIEW masterTmplFlags = 1 << iota
)

// Environment/context for the master template.
type masterTmplEnv struct {
	Node               service.NodeInfo
	Session            *service.UserSession
	Title, Description string
	Flags              masterTmplFlags
}

// splitFirstDir returns the first directory in the given path.
func splitFirstDir(path string) string {
	for len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	return strings.SplitN(path, "/", 2)[0]
}

// renderInMaster renders the content in the master template.
func renderInMaster(r template.Renderer, content []byte, env masterTmplEnv,
	settings *settings, site site, locale string) string {
	firstDir := splitFirstDir(env.Node.Path)
	prinav, err := getNav("/", path.Join("/", firstDir),
		settings.Monsti.GetSiteNodesPath(site.Name))
	prinav.MakeAbsolute(firstDir)
	if err != nil {
		panic(fmt.Sprint("Could not get primary navigation: ", err))
	}
	prinav.MakeAbsolute("/")
	var secnav navigation = nil
	if env.Node.Path != "/" {
		secnav, err = getNav(env.Node.Path, env.Node.Path,
			settings.Monsti.GetSiteNodesPath(site.Name))
		if err != nil {
			panic(fmt.Sprint("Could not get secondary navigation: ", err))
		}
		secnav.MakeAbsolute(env.Node.Path)
	}
	sidebarContent := getSidebar(env.Node.Path,
		settings.Monsti.GetSiteNodesPath(site.Name))
	belowHeader := getBelowHeader(env.Node.Path,
		settings.Monsti.GetSiteNodesPath(site.Name))
	title := env.Node.Title
	if env.Title != "" {
		title = env.Title
	}
	description := env.Node.Description
	if env.Title != "" {
		description = env.Description
	}
	return r.Render("master", template.Context{
		"Site": template.Context{
			"Title": site.Title,
		},
		"Page": template.Context{
			"Node":            env.Node,
			"PrimaryNav":      prinav,
			"SecondaryNav":    secnav,
			"EditView":        env.Flags&EDIT_VIEW != 0,
			"ShowBelowHeader": len(belowHeader) > 0 && (env.Flags&EDIT_VIEW == 0),
			"BelowHeader":     htmlT.HTML(belowHeader),
			"Footer": htmlT.HTML(getFooter(
				settings.Monsti.GetSiteNodesPath(site.Name))),
			"Sidebar":          htmlT.HTML(sidebarContent),
			"Title":            title,
			"Description":      description,
			"Content":          htmlT.HTML(content),
			"ShowSecondaryNav": len(secnav) > 0},
		"Session": env.Session}, locale,
		settings.Monsti.GetSiteTemplatesPath(site.Name))
}
