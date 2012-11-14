package template

import (
	"bytes"
	"datenkarussell.de/monsti/l10n"
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
)

// Context can be used to define a context for Render.
type Context map[string]interface{}

// A Renderer for mustache templates.
type Renderer struct {
	// Root is the absolute path to the template directory.
	Root string
}

// Render the named template with given context. 
func (r Renderer) Render(name string, context interface{},
	locale string) string {
	tmpl := template.New(name)
	G := l10n.UseCatalog(locale)
	funcs := template.FuncMap{
		"pathJoin": path.Join,
		"G":        G}
	tmpl.Funcs(funcs)
	parse(name, tmpl, r.Root)
	parse("blocks/form-horizontal", tmpl.New("blocks/form-horizontal"), r.Root)
	parse("blocks/form-vertical", tmpl.New("blocks/form-vertical"), r.Root)
	out := bytes.Buffer{}
	if err := tmpl.Execute(&out, context); err != nil {
		panic("Could not execute template: " + err.Error())
	}
	return out.String()
}

func parse(name string, t *template.Template, root string) {
	path := filepath.Join(root, name+".html")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic("Could not load template:" + err.Error())
	}
	_, err = t.Parse(string(content))
	if err != nil {
		panic("Could not parse template:" + err.Error())
	}
}
