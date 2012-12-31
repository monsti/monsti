package main

import (
	"github.com/monsti/rpc/client"
	"github.com/monsti/util/template"
	htmlT "html/template"
	"strings"
)

// Master template render flags.
type masterTmplFlags uint32

const (
	EDIT_VIEW masterTmplFlags = 1 << iota
)

// Environment/context for the master template.
type masterTmplEnv struct {
	Node               client.Node
	Session            *client.Session
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
	settings settings, site site, locale string) string {
	prinav, _ := getNav("/", splitFirstDir(env.Node.Path), false,
		site.Directories.Data)
	prinav.MakeAbsolute("/")
	var secnav navigation = nil
	if env.Node.Path != "/" {
		var root string
		secnav, root = getNav(env.Node.Path, env.Node.Path, true,
			site.Directories.Data)
		secnav.MakeAbsolute(root)
	}
	sidebarContent := getSidebar(env.Node.Path, site.Directories.Data)
	belowHeader := getBelowHeader(env.Node.Path, site.Directories.Data)
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
			"Node":             env.Node,
			"PrimaryNav":       prinav,
			"SecondaryNav":     secnav,
			"EditView":         env.Flags&EDIT_VIEW != 0,
			"ShowBelowHeader":  len(belowHeader) > 0 && (env.Flags&EDIT_VIEW == 0),
			"BelowHeader":      htmlT.HTML(belowHeader),
			"Footer":           htmlT.HTML(getFooter(site.Directories.Data)),
			"Sidebar":          htmlT.HTML(sidebarContent),
			"Title":            title,
			"Description":      description,
			"Content":          htmlT.HTML(content),
			"ShowSecondaryNav": len(secnav) > 0
		},
		"Session": env.Session}, locale, site.Directories.Templates)
}
