package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
)

type masterTmplEnv struct {
	Node                     client.Node
	PrimaryNav, SecondaryNav []navLink
}

func renderInMaster(r template.Renderer, content []byte, env *masterTmplEnv,
	settings settings, contexts ...interface{}) string {
	sidebarContent := getSidebar(env.Node.Path, settings.Root)
	showSidebar := (len(env.SecondaryNav) > 0 || len(sidebarContent) > 0) &&
		!env.Node.HideSidebar
	belowHeader := getBelowHeader(env.Node.Path, settings.Root)
	return r.Render("master.html", env, map[string]interface{}{
		"ShowBelowHeader":  len(belowHeader) > 0,
		"BelowHeader":      belowHeader,
		"Footer":           getFooter(settings.Root),
		"Sidebar":          sidebarContent,
		"SiteTitle":        settings.Title,
		"Content":          string(content),
		"ShowSecondaryNav": len(env.SecondaryNav) > 0,
		"ShowSidebar":      showSidebar})
}
