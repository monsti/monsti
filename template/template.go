// This file is part of monsti/util.
// Copyright 2012 Christian Neumann

// monsti/util is free software: you can redistribute it and/or modify it under
// the terms of the GNU Lesser General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.

// monsti/util is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more
// details.

// You should have received a copy of the GNU Lesser General Public License
// along with monsti/util. If not, see <http://www.gnu.org/licenses/>.

/*
Package template implements template rendering services for Monsti
and Monsti content worker types.
*/
package template

import (
	"bytes"
	"github.com/monsti/util/l10n"
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
//
// name is the name of the template (e.g. "blocks/sidebar").
// context is used as template context for rendering.
// locale is the locale to use for translation strings in templates.
// siteTemplates is the path to the site's overridden templates. If it's an empty
// string, Render will not search for overridden templates.
//
// Returns the rendered template.
func (r Renderer) Render(name string, context interface{},
	locale string, siteTemplates string) string {
	tmpl := template.New(name)
	G := l10n.UseCatalog(locale)
	funcs := template.FuncMap{
		"pathJoin": path.Join,
		"G":        G}
	tmpl.Funcs(funcs)
	parse(name, tmpl, r.Root, siteTemplates)
	for _, v := range []string{
		"blocks/form-horizontal",
		"blocks/form-vertical",
		"blocks/primary-navigation",
		"blocks/headers-edit",
		"blocks/headers",
		"blocks/admin-bar",
		"blocks/sidebar",
		"blocks/below-header"} {
		parse(v, tmpl.New(v), r.Root, siteTemplates)
	}
	out := bytes.Buffer{}
	if err := tmpl.Execute(&out, context); err != nil {
		panic("Could not execute template: " + err.Error())
	}
	return out.String()
}

// Parse the named template and add to the existing template structure.
//
// name is the name of the template (e.g. "blocks/sidebar")
// t is the existing template structure.
// root is the path to monsti's template
// siteRoot is the path to the sites' overriden templates.
func parse(name string, t *template.Template, root string,
	siteRoot string) {
	if len(siteRoot) > 0 {
		path := filepath.Join(siteRoot, name+".html")
		content, err := ioutil.ReadFile(path)
		if err == nil {
			_, err = t.Parse(string(content))
			if err != nil {
				panic("Could not parse template:" + err.Error())
			}
			return
		}
	}
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
