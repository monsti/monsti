package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
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

// renderInMaster renders the content in the master template.
func renderInMaster(r template.Renderer, content []byte, env masterTmplEnv,
	settings settings, site site, locale string) string {
	prinav := getNav("/", "/"+strings.SplitN(env.Node.Path[1:], "/", 2)[0],
		true, site.Directories.Data)
	var secnav []navLink = nil
	if env.Node.Path != "/" {
		secnav = getNav(env.Node.Path, env.Node.Path, true,
			site.Directories.Data)
	}
	sidebarContent := getSidebar(env.Node.Path, site.Directories.Data)
	showSidebar := (len(secnav) > 0 || len(sidebarContent) > 0) &&
		!env.Node.HideSidebar && (env.Flags&EDIT_VIEW == 0)
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
			"ShowSecondaryNav": len(secnav) > 0,
			"ShowSidebar":      showSidebar,
		},
		"Session": env.Session}, locale)
}
