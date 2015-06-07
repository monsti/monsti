// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann
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
	htmlT "html/template"
	"path"
	"strings"

	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util/template"
)

// Master template render flags.
type masterTmplFlags uint32

const (
	EDIT_VIEW masterTmplFlags = 1 << iota
	SLIM_VIEW
)

// Environment/context for the master template.
type masterTmplEnv struct {
	Node               *service.Node
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

type renderedBlock struct {
	Block    *service.Block
	Rendered htmlT.HTML
}

// renderInMaster renders the content in the master template.
func renderInMaster(r template.Renderer, content []byte, env masterTmplEnv,
	settings *settings, site string, siteSettings *service.Settings,
	userLocale string,
	s *service.Session) ([]byte, *service.CacheMods) {
	mods := &service.CacheMods{Deps: []service.CacheDep{{Node: "/", Descend: -1}}}
	if env.Flags&EDIT_VIEW != 0 {
		ret, err := r.Render("admin/master", template.Context{
			"Site":         site,
			"SiteSettings": siteSettings,
			"Page": template.Context{
				"Title":    env.Title,
				"Node":     env.Node,
				"EditView": true,
				"SlimView": env.Flags&SLIM_VIEW != 0,
				"Content":  htmlT.HTML(content),
			},
			"Session": env.Session}, userLocale,
			settings.Monsti.GetSiteTemplatesPath(site))
		if err != nil {
			panic("Can't render admin template: " + err.Error())
		}
		return ret, nil
	}
	firstDir := splitFirstDir(env.Node.Path)
	getNodeFn := func(path string) (*service.Node, error) {
		node, err := s.Monsti().GetNode(site, path)
		return node, err
	}
	getChildrenFn := func(path string) ([]*service.Node, error) {
		return s.Monsti().GetChildren(site, path)
	}

	priNavDepth := 1
	if nav, ok := siteSettings.Fields["core.Navigations"].(*service.MapField).
		Fields["core.Main"]; ok {
		priNavDepth = nav.(*service.CombinedField).
			Fields["depth"].Value().(int)
	}
	prinav, err := getNav("/", path.Join("/", firstDir), env.Session.User == nil,
		getNodeFn, getChildrenFn, priNavDepth)
	if err != nil {
		panic(fmt.Sprint("Could not get primary navigation: ", err))
	}
	prinav.MakeAbsolute("/")

	var secnav navigation = nil
	if env.Node.Path != "/" {
		secnav, err = getNav(env.Node.Path, env.Node.Path, env.Session.User == nil,
			getNodeFn, getChildrenFn, 1)
		if err != nil {
			panic(fmt.Sprint("Could not get secondary navigation: ", err))
		}
		secnav.MakeAbsolute(env.Node.Path)
	}

	blocks := make(map[string][]renderedBlock)

	// EXPERIMENTAL Render blocks
	if _, ok := siteSettings.Fields["core.RegionBlocks"].(*service.MapField).
		Fields["core.PrimaryNavigation"].(*service.ListField); ok {
		renderedNav, err := r.Render("blocks/core/Navigation", template.Context{
			"Id":    "core.PrimaryNavigation",
			"Links": prinav,
		}, userLocale, settings.Monsti.GetSiteTemplatesPath(site))
		if err != nil {
			panic(fmt.Sprintf("Could not render navigation: %v", err))
		}
		blocks["core.PrimaryNavigation"] = append(blocks["core.PrimaryNavigation"],
			renderedBlock{
				Rendered: htmlT.HTML(renderedNav),
			})
	}

	title := getNodeTitle(env.Node)
	ret, err := r.Render("master", template.Context{
		"Site":         site,
		"SiteSettings": siteSettings,
		"Page": template.Context{
			"Node":             env.Node,
			"PrimaryNav":       prinav, // TODO DEPRECATED
			"SecondaryNav":     secnav, // TODO DEPRECATED
			"EditView":         env.Flags&EDIT_VIEW != 0,
			"Title":            title,
			"Content":          htmlT.HTML(content),
			"ShowSecondaryNav": len(secnav) > 0, // TODO DEPRECATED
			"Blocks":           blocks,
		},
		"Session": env.Session}, userLocale,
		settings.Monsti.GetSiteTemplatesPath(site))
	if err != nil {
		panic(fmt.Sprintf(
			"Can't render master template for site %v: %v",
			site, err))
	}
	return ret, mods
}
