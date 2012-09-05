package monsti

import (
	"io/ioutil"
	"launchpad.net/goyaml"
	"net/http"
	"path/filepath"
)


// NodeFile is the filename of node description files.
const NodeFile = "node.yaml"



// Node is the interface implemented by the various node types
// (Documents, Images, ...).
type Node interface {
	// Path returns the node's path (e.g. "/node").
	Path() string

	// Title returns the node's title.
	Title() string

	// Description returns the node's description.
	Description() string

	// HideSidebar returns if the node's sidebar should be hidden.
	HideSidebar() bool

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
	Description  string
	Title        string
	Type         string
	Hide_sidebar bool
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

func (n node) HideSidebar() bool {
	return n.data.Hide_sidebar
}


// lookup_node look ups a node at the given path.
// If no such node exists, return nil.
func LookupNode(root, path string) (Node, error) {
	node_path := filepath.Join(root, path[1:], NodeFile)
	content, err := ioutil.ReadFile(node_path)
	if err != nil {
		return nil, err
	}
	var ret Node
	var data nodeData
	goyaml.Unmarshal(content, &data)
	switch data.Type {
	case "Document":
		document := fetchDocument(data, path, root)
		ret = document
	case "ContactForm":
		contactForm := contactForm{*fetchDocument(data, path, root)}
		contactForm.data.Hide_sidebar = true
		ret = contactForm
	default:
		panic("Unknown node type: " + data.Type)
	}
	return ret, nil
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

