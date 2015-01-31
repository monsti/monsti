// This file is part of monsti/util.
// Copyright 2012-2013 Christian Neumann

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
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"pkg.monsti.org/gettext"
)

// Context can be used to define a context for Render.
type Context map[string]interface{}

// A Renderer for mustache templates.
type Renderer struct {
	// Root is the absolute path to the template directory.
	Root string
}

// getIncludes searches for include and template.include files.
//
// roots are the template roots to search (results will be joined and duplicates
// removed).
// name is the name of the template (e.g. "blocks/sidebar").
//
// Returns a list of templates to be included.
func getIncludes(roots []string, name string) ([]string, error) {
	includes := make([]string, 0)
	if len(name) == 0 || name[0] == filepath.Separator {
		return nil, fmt.Errorf("Invalid template name: %q", name)
	}
	duplicateCheck := make(map[string]bool)
	paths := []string{
		name + ".include"}
	for path := name; path != "."; {
		path = filepath.Dir(path)
		paths = append(paths, filepath.Join(path, "include"))
	}
	for _, root := range roots {
		for _, path := range paths {
			contents, err := ioutil.ReadFile(filepath.Join(root, path))
			if err != nil {
				continue
			}
			for _, line := range bytes.Split(contents, []byte("\n")) {
				if len(line) == 0 {
					continue
				}
				if _, ok := duplicateCheck[string(line)]; ok {
					continue
				}
				duplicateCheck[string(line)] = true
				includes = append(includes, string(line))
			}
		}
	}
	return includes, nil
}

// Render the named template with given context.
//
// name is the name of the template (e.g. "blocks/sidebar").
// context is used as template context for rendering.
// locale is the locale to use for translation strings in templates.
// siteTemplates is the path to the site's overridden templates. If it's an empty
// string, Render will not search for overridden templates.
//
// Render searches for nested templates to include in these files:
// <dir_of_template>/<template>.include
// <dir_of_template>/include
// <any_parent_dir_of_template>/include
//
// Returns the rendered template.
func (r Renderer) Render(name string, context interface{},
	locale string, siteTemplates string) ([]byte, error) {
	tmpl := template.New(name)
	G, GN, GD, GDN := gettext.DefaultLocales.Use("", locale)
	funcs := template.FuncMap{
		"pathJoin": path.Join,
		"G":        G,
		"GN":       GN,
		"GD":       GD,
		"GDN":      GDN,
		"RawHTML": func(in interface{}) template.HTML {
			return template.HTML(fmt.Sprintf("%s", in))
		},
	}
	tmpl.Funcs(funcs)
	err := parseSiteTemplate(name, tmpl, r.Root, siteTemplates)
	if err != nil {
		return nil, err
	}
	includes, err := getIncludes([]string{r.Root, siteTemplates}, name)
	if err != nil {
		return nil, err
	}
	for _, v := range includes {
		err := parseSiteTemplate(v, tmpl.New(v), r.Root, siteTemplates)
		if err != nil {
			return nil, err
		}
	}
	out := bytes.Buffer{}
	if err := tmpl.Execute(&out, context); err != nil {
		return nil, fmt.Errorf("Could not execute template: %v", err)
	}
	return out.Bytes(), nil
}

// parseSiteTemplate tries to parse the named template for the given
// site and adds it to given template structure.
//
// It searches in the site's template root for the given template. If
// there is no template, it searches in Monsti's template directory.
//
// `name` is the name of the template (e.g. `blocks/sidebar`). `t` is
// the existing template structure. `root` is the path to Monsti's
// base templates. `siteRoot` is the path to the sites' templates.
func parseSiteTemplate(name string, t *template.Template, root string,
	siteRoot string) error {
	ok := false
	var err error
	if len(siteRoot) > 0 {
		if ok, err = parse(name, t, siteRoot); err != nil {
			return err
		}
	}
	if !ok {
		if _, err = parse(name, t, root); err != nil {
			return err
		}
	}
	return nil
}

// Parse the named template and add to the existing template structure.
//
// name is the name of the template (e.g. "blocks/sidebar").
// t is the existing template structure.
// root is the path to the templates.
//
// Returns false, nil if no template was found.
func parse(name string, t *template.Template, root string) (bool, error) {
	path := filepath.Join(root, name+".html")
	content, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("Could not read template file: %v", err)
	}
	if _, err = t.Parse(string(content)); err != nil {
		return false, fmt.Errorf("Could not parse template file: %v", err)
	}
	return true, nil
}
