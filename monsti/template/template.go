package template

import (
	"github.com/drbawb/mustache"
	//	"github.com/hoisie/mustache"
	"path/filepath"
)

// A Renderer for mustache templates.
type Renderer struct {
	// Root is the absolute path to the template directory.
	Root string
}

func getText(foo string) string {
  return "yay!"
}

// Render the named template with given context. 
func (r Renderer) Render(name string, contexts ...interface{}) string {
	path := filepath.Join(r.Root, name)
	globalFuns := map[string]interface{}{
		"_": getText}
	content := mustache.RenderFile(path, append(contexts, globalFuns)...)
	return content
}
