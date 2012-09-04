/*
 Brassica is a simple and resource efficient CMS for low dynamic
 private and small business sites with mostly static pages and simple
 structure.

 brassica/brassica-serve contains a command to start a httpd.
*/
package brassica

import (
	"fmt"
	"github.com/hoisie/mustache"
	"io/ioutil"
	"launchpad.net/goyaml"
	"net/http"
	"os"
	"path/filepath"
)

// Settings for the application and the site.
type Settings struct {
	// Path to the data directory.
	Root string

	// Path to the template directory.
	Templates string

	// Path to the static files.
	Statics string

	// Path to the site specific static files.
	SiteStatics string
}

// GetSettings returns application and site settings.
func GetSettings() Settings {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	settings := Settings{
		Root:        wd,
		Templates:   filepath.Join(filepath.Dir(wd), "templates"),
		Statics:     filepath.Join(filepath.Dir(wd), "static"),
		SiteStatics: filepath.Join(filepath.Dir(wd), "site-static")}
	return settings
}

type masterTmplEnv struct {
	Node                     Node
	PrimaryNav, SecondaryNav []navLink
}

// Renderer represents a template renderer.
type Renderer interface {
	// RenderInMaster renders the named template with the given context
	// and master template environment in the master template.
	RenderInMaster(name string, context map[string]string,
		env *masterTmplEnv, settings Settings) string
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
		panic("Could not load master template: " + err.Error())
	}
	return r
}

func (r renderer) RenderInMaster(name string, context map[string]string,
	env *masterTmplEnv, settings Settings) string {
	content := r.Render(name, context)
	sidebarContent := getSidebar(env.Node.Path(), settings.Root)
	showSidebar := len(env.SecondaryNav) > 0 || len(sidebarContent) > 0
	return r.MasterTemplate.Render(env, map[string]interface{}{
		"BelowHeader":      getBelowHeader(env.Node.Path(), settings.Root),
		"Footer":           getFooter(settings.Root),
		"Sidebar":          sidebarContent,
		"Content":          content,
		"ShowSecondaryNav": len(env.SecondaryNav) > 0,
		"ShowSidebar":      showSidebar})
}

// getFooter retrieves the footer.
//
// root is the path to the data directory
//
// Returns an empty string if there is no footer.
func getFooter(root string) string {
	path := filepath.Join(root, "footer.html")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

// getBelowHeader retrieves the below header content for the given node.
//
// path is the node's path.
// root is the path to the data directory.
//
// Returns an empty string if there is no below header content.
func getBelowHeader(path, root string) string {
	file := filepath.Join(root, path, "below_header.html")
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(content)
}

// getSidebar retrieves the sidebar content for the given node.
//
// path is the node's path.
// root is the path to the data directory.
//
// It traverses up to the root until it finds a node with defined sidebar
// content.
//
// Returns an empty string if there is no sidebar content.
func getSidebar(path, root string) string {
	for {
		file := filepath.Join(root, path, "sidebar.html")
		content, err := ioutil.ReadFile(file)
		if err != nil {
			if path == filepath.Dir(path) {
				break
			}
			path = filepath.Dir(path)
			continue
		}
		return string(content)
	}
	return ""
}

// navLink represents a link in the navigation.
type navLink struct {
	Name, Target string
	Active       bool
}

// getNav returns the navigation for the given node.
// 
// node is the path of the node for which to get the navigation.
// active is the currently active node.
// root is the path of the data directory.
//
// The keys of the returned map are the link titles, the values are
// the link targets.
func getNav(node, active, root string) []navLink {
	path := filepath.Join(root, node, "navigation.yaml")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	var navLinks []navLink
	goyaml.Unmarshal(content, &navLinks)
	for i, link := range navLinks {
		if link.Target == active {
			navLinks[i].Active = true
			break
		}
	}
	return navLinks
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
	Get(w http.ResponseWriter, r *http.Request, renderer Renderer,
		settings Settings)

	// Post handles a POST request on the node.
	Post(w http.ResponseWriter, r *http.Request, renderer Renderer,
		settings Settings)
}

// node is the base implementation for nodes.
type node struct {
	path string
	data nodeData
}

// nodeData is used for (un)marshaling from/to node.yaml.
type nodeData struct {
	Description string
	Title       string
	Type        string
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
	renderer Renderer, settings Settings) {
	prinav := getNav("/", n.Path(), settings.Root)
	var secnav []navLink = nil
	if n.Path() != "/" {
		secnav = getNav(n.Path(), n.Path(), settings.Root)
	}
	env := masterTmplEnv{
		Node:         n,
		PrimaryNav:   prinav,
		SecondaryNav: secnav}
	content := renderer.RenderInMaster("view/document.html",
		map[string]string{"body": n.Body}, &env, settings)
	fmt.Fprint(w, content)
}

func (n Document) Post(w http.ResponseWriter, r *http.Request,
	renderer Renderer, settings Settings) {
	http.Error(w, "Implementation missing.", http.StatusInternalServerError)
}

// NodeFile is the filename of node description files.
const NodeFile = "node.yaml"

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
	node.path = path
	return node, nil
}
