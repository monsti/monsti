/*
 Brassica is a simple and resource efficient CMS for low dynamic
 private and small business sites with mostly static pages and simple
 structure.

 brassica/brassica-serve contains a command to start a httpd.
*/
package brassica

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"net/http"
	"launchpad.net/goyaml"
	"github.com/hoisie/mustache"
)

type Renderer interface {
	// RenderInMaster renders the named template with the given context
	// in the master template.
	RenderInMaster(name, title string, context map[string]string) string
}

type renderer struct {
	// Root is the absolute path to the template directory.
	Root string
	// MasterTemplate holds the parsed master template.
	MasterTemplate *mustache.Template
}

// Render renders the named template with given context. 
func (r renderer) Render(name string, context map[string]string) string {
	path := filepath.Join(r.Root, name)
	content := mustache.RenderFile(path, context)
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
		panic("Could not load master template.")
	}
	return r
}

func (r renderer) RenderInMaster(name, title string,
	context map[string]string) string {
	content := r.Render(name, context)
	return r.MasterTemplate.Render(map[string]string{
		"title":   title,
		"content": content})
}

// Node is the interface implemented by the various node types
// (Documents, Images, ...).
type Node interface {
	// Path returns the node's path (e.g. "/node").
	Path() string

	// Title returns the node's title.
	Title() string

	// Description returns the node's description.
	Description() string

	// Get handles a GET request on the node.
	Get(w http.ResponseWriter, r *http.Request, renderer Renderer)

	// Post handles a POST request on the node.
	Post(w http.ResponseWriter, r *http.Request, renderer Renderer)
}

// node is the base implementation for nodes.
type node struct {
	path string
	data nodeData
}

type nodeData struct {
	Description string
	Title string
	Type string
}

func (n node) Path() string {
	return n.path
}

func (n node) Title() string {
	return n.data.Title
}

func (n node) Description() string {
	return n.data.Description
}

// Document is a node consisting of a html body.
type Document struct {
	node
	Body string
}

func (n Document) Get(w http.ResponseWriter, r *http.Request,
	renderer Renderer) {
	content := renderer.RenderInMaster("view/document.html", n.Title(), map[string]string{
		"body": n.Body})
	fmt.Fprint(w, content)
}

func (n Document) Post(w http.ResponseWriter, r *http.Request,
	renderer Renderer) {
	http.Error(w, "Implementation missing.", http.StatusInternalServerError)
}

// NodeFile is the filename of node description files.
const NodeFile = "node.yml"

// lookup_node look ups a node at the given path.
// If no such node exists, return nil.
func LookupNode(root, path string) (Node, error) {
	node_path := filepath.Join(root, path[1:], NodeFile)
  content, err := ioutil.ReadFile(node_path)
	if err != nil {
		return nil, err
	}
	var node = new(Document)
	body_path := filepath.Join(root, path[1:], "body.html")
  body, err := ioutil.ReadFile(body_path)
	if err != nil {
		panic("Body not found: " + body_path)
	}
	node.Body = string(body)
	var data nodeData
	goyaml.Unmarshal(content, &data)
	node.data = data
	return node, nil
}