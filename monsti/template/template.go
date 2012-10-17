package template

import (
	"bytes"
	"github.com/chrneumann/g5t"
	"html/template"
	"io/ioutil"
	"path/filepath"
)

// Context can be used to define a context for Render.
type Context map[string]interface{}

// A Renderer for mustache templates.
type Renderer struct {
	// Root is the absolute path to the template directory.
	Root string
}

// getText looks up the translation for the given string.
func getText(msg string) string {
	return g5t.String(msg)
}

// pathJoin joins both url path segments so that they are separated by exactly
// one slash.
func pathJoin(first, last string) string {
	for len(first) > 0 && first[len(first)-1:] == "/" {
		first = first[:len(first)-1]
	}
	return first + "/" + last
}

// Render the named template with given context. 
func (r Renderer) Render(name string, context interface{}) string {
	tmpl := template.New(name)
	funcs := template.FuncMap{
		"pathJoin": pathJoin,
		"G":        getText}
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
