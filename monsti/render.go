package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
)

type masterTmplEnv struct {
	Node                     client.Node
	PrimaryNav, SecondaryNav []navLink
	Session                  *client.Session
}

func renderInMaster(r template.Renderer, content []byte, env *masterTmplEnv,
	settings settings, contexts ...interface{}) string {
	sidebarContent := getSidebar(env.Node.Path, settings.Directories.Data)
	showSidebar := (len(env.SecondaryNav) > 0 || len(sidebarContent) > 0) &&
		!env.Node.HideSidebar
	belowHeader := getBelowHeader(env.Node.Path, settings.Directories.Data)
	return r.Render("master.html", env, map[string]interface{}{
		"ShowBelowHeader":  len(belowHeader) > 0,
		"BelowHeader":      belowHeader,
		"Footer":           getFooter(settings.Directories.Data),
		"Sidebar":          sidebarContent,
		"SiteTitle":        settings.Title,
		"Content":          string(content),
		"ShowSecondaryNav": len(env.SecondaryNav) > 0,
		"ShowSidebar":      showSidebar})
}
