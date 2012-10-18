package main

import (
	"code.google.com/p/gorilla/sessions"
	"datenkarussell.de/monsti/form"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"fmt"
	"github.com/chrneumann/g5t"
	"io/ioutil"
	"launchpad.net/goyaml"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var G func(string) string = g5t.String

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
	Active       bool "active,omitempty"
}

// getNav returns the navigation for the given node.
// 
// node is the path of the node for which to get the navigation.
// active is the currently active node.
// root is the path of the data directory.
//
// The keys of the returned map are the link titles, the values are
// the link targets.
//
// If the node has no navigation defined (i.e. there exists no
// navigation.yaml), a navigation is searched recursively for the parent node up
// to the root.
func getNav(path, active, root string) []navLink {
	var content []byte
	for {
		file := filepath.Join(root, path, "navigation.yaml")
		var err error
		content, err = ioutil.ReadFile(file)
		if err != nil {
			if path == filepath.Dir(path) {
				break
			}
			path = filepath.Dir(path)
			continue
		}
		break
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

// dumpNav unmarshals the navigation and writes it to the node's directory.
func dumpNav(nav []navLink, nodePath, root string) {
	for i := range nav {
		nav[i].Active = false
	}
	content, err := goyaml.Marshal(&nav)
	if err != nil {
		panic("Could not marshal navigation: " + err.Error())
	}
	path := filepath.Join(root, nodePath[1:], "navigation.yaml")
	err = ioutil.WriteFile(path, content, 0600)
	if err != nil {
		panic("Could not write navigation: " + err.Error())
	}
}

type addFormData struct {
	Type, Name, Title string
}

// Add handles add requests.
func (h *nodeHandler) Add(w http.ResponseWriter, r *http.Request,
	node client.Node, session *sessions.Session, cSession *client.Session) {
	data := addFormData{}
	selectWidget := form.SelectWidget{[]form.Option{
		{"Document", G("Document")}}}
	form := form.NewForm(&data, form.Fields{
		"Type": form.Field{G("Content type"), "", form.Required(), selectWidget},
		"Name": form.Field{G("Name"),
			G("The name as it should appear in the URL."),
			form.And(form.Required(), form.Regex(`^\w*$`,
				G("Contains	invalid characters."))), nil},
		"Title": form.Field{G("Title"), "", form.Required(), nil}})
	switch r.Method {
	case "GET":
	case "POST":
		r.ParseForm()
		if form.Fill(r.Form) {
			data.Name = strings.ToLower(data.Name)
			// TODO Allow other content types.
			if data.Type != "Document" {
				panic("Can't add this content type.")
			}
			newPath := filepath.Join(node.Path, data.Name)
			newNode := client.Node{
				Path:  newPath,
				Type:  data.Type,
				Title: data.Title}
			if err := writeNode(newNode, h.Settings.Directories.Data); err != nil {
				panic("Can't add node: " + err.Error())
			}
			http.Redirect(w, r, newPath+"/@@edit", http.StatusSeeOther)
			return
		}
	default:
		panic("Request method not supported: " + r.Method)
	}
	body := h.Renderer.Render("actions/addform", template.Context{
		"Form": form.RenderData()})
	env := masterTmplEnv{Node: node, Session: cSession,
		Flags: EDIT_VIEW, Title: G("Add content")}
	fmt.Fprint(w, renderInMaster(h.Renderer, []byte(body), env, h.Settings))
}

type removeFormData struct {
	Confirm int
}

// Remove handles remove requests.
func (h *nodeHandler) Remove(w http.ResponseWriter, r *http.Request,
	node client.Node, session *sessions.Session, cSession *client.Session) {
	data := removeFormData{}
	form := form.NewForm(&data, form.Fields{
		"Confirm": form.Field{G("Confirm"), "", form.Required(),
			new(form.HiddenWidget)}})
	switch r.Method {
	case "GET":
	case "POST":
		r.ParseForm()
		if form.Fill(r.Form) {
			if err := removeNode(node, h.Settings.Directories.Data); err != nil {
				panic("Can't remove node: " + err.Error())
			}
			http.Redirect(w, r, path.Dir(node.Path), http.StatusSeeOther)
			return
		}
	default:
		panic("Request method not supported: " + r.Method)
	}
	data.Confirm = 1489
	body := h.Renderer.Render("actions/removeform", template.Context{
		"Form": form.RenderData(), "Node": node})
	env := masterTmplEnv{Node: node, Session: cSession,
		Flags: EDIT_VIEW, Title: fmt.Sprintf(G("Remove \"%v\""), node.Title)}
	fmt.Fprint(w, renderInMaster(h.Renderer, []byte(body), env, h.Settings))
}

// lookupNode look ups a node at the given path.
// If no such node exists, return nil.
func lookupNode(root, path string) (client.Node, error) {
	node_path := filepath.Join(root, path[1:], "node.yaml")
	content, err := ioutil.ReadFile(node_path)
	if err != nil {
		return client.Node{}, err
	}
	var node client.Node
	if err = goyaml.Unmarshal(content, &node); err != nil {
		return client.Node{}, err
	}
	node.Path = path
	return node, nil
}

// writeNode writes the given node to the data directory located at the given
// root.
func writeNode(node client.Node, root string) error {
	path := node.Path
	node.Path = ""
	content, err := goyaml.Marshal(&node)
	if err != nil {
		return err
	}
	node_path := filepath.Join(root, path[1:],
		"node.yaml")
	if err := os.Mkdir(filepath.Dir(node_path), 0700); err != nil {
		if !os.IsExist(err) {
			panic("Can't create directory for new node: " + err.Error())
		}
	}
	return ioutil.WriteFile(node_path, content, 0600)
}

// removeNode recursively removes the given node from the data directory located
// at the given root.
func removeNode(node client.Node, root string) error {
	nodePath := filepath.Join(root, node.Path[1:])
	return os.RemoveAll(nodePath)
}
