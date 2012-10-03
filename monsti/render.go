package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
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
	settings settings, contexts ...interface{}) string {
	prinav := getNav("/", "/"+strings.SplitN(env.Node.Path[1:], "/", 2)[0],
		settings.Directories.Data)
	var secnav []navLink = nil
	if env.Node.Path != "/" {
		secnav = getNav(env.Node.Path, env.Node.Path, settings.Directories.Data)
	}
	sidebarContent := getSidebar(env.Node.Path, settings.Directories.Data)
	showSidebar := (len(secnav) > 0 || len(sidebarContent) > 0) &&
		!env.Node.HideSidebar && (env.Flags&EDIT_VIEW == 0)
	belowHeader := getBelowHeader(env.Node.Path, settings.Directories.Data)
	title := env.Node.Title
	if env.Title != "" {
		title = env.Title
	}
	description := env.Node.Description
	if env.Title != "" {
		description = env.Description
	}
	return r.Render("master.html", map[string]interface{}{
		"Node":             env.Node,
		"PrimaryNav":       prinav,
		"SecondaryNav":     secnav,
		"Session":          env.Session,
		"ShowBelowHeader":  len(belowHeader) > 0 && (env.Flags&EDIT_VIEW == 0),
		"BelowHeader":      belowHeader,
		"Footer":           getFooter(settings.Directories.Data),
		"Sidebar":          sidebarContent,
		"SiteTitle":        settings.Title,
		"Title":            title,
		"Description":      description,
		"Content":          string(content),
		"ShowSecondaryNav": len(secnav) > 0,
		"ShowSidebar":      showSidebar})
}
