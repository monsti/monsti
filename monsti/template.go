package monsti

import (
	"path/filepath"
        "github.com/hoisie/mustache"
)


type masterTmplEnv struct {
	Node                     Node
	PrimaryNav, SecondaryNav []navLink
}

// Renderer represents a template renderer.
type Renderer interface {
	// RenderInMaster renders the named template with the given contexts
	// and master template environment in the master template.
	RenderInMaster(name string, env *masterTmplEnv, settings Settings,
		contexts ...interface{}) string
}

type renderer struct {
	// Root is the absolute path to the template directory.
	Root string
	// MasterTemplate holds the parsed master template.
	MasterTemplate *mustache.Template
}

// Render renders the named template with given context. 
func (r renderer) Render(name string, contexts ...interface{}) string {
	path := filepath.Join(r.Root, name)
	globalFuns := map[string]interface{}{
		"_": "implement me!"}
	content := mustache.RenderFile(path, append(contexts, globalFuns)...)
	return content
}

// NewRenderer returns a new Renderer.
//
// root is the absolute path to the template directory.
func NewRenderer(root string) Renderer {
	var r renderer
	r.Root = root
	path := filepath.Join(r.Root, "master.html")
	tmpl, err := mustache.ParseFile(path)
	r.MasterTemplate = tmpl
	if err != nil {
		panic("Could not load master template: " + err.Error())
	}
	return r
}

func (r renderer) RenderInMaster(name string, env *masterTmplEnv,
	settings Settings, contexts ...interface{}) string {
	content := r.Render(name, contexts...)
	sidebarContent := getSidebar(env.Node.Path(), settings.Root)
	showSidebar := (len(env.SecondaryNav) > 0 || len(sidebarContent) > 0) &&
		!env.Node.HideSidebar()
	belowHeader := getBelowHeader(env.Node.Path(), settings.Root)
	return r.MasterTemplate.Render(env, map[string]interface{}{
		"ShowBelowHeader":  len(belowHeader) > 0,
		"BelowHeader":      belowHeader,
		"Footer":           getFooter(settings.Root),
		"Sidebar":          sidebarContent,
		"Content":          content,
		"ShowSecondaryNav": len(env.SecondaryNav) > 0,
		"ShowSidebar":      showSidebar})
}


