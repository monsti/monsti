// This file is part of Monsti, a web content management system.
// Copyright 2012-2013 Christian Neumann
// 
// Monsti is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// Monsti is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Monsti.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/monsti/form"
	"github.com/monsti/service/login"
	"github.com/monsti/service/node"
	"github.com/monsti/util/l10n"
	"github.com/monsti/util/template"
	"io/ioutil"
	"launchpad.net/goyaml"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

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
	Name, Target  string
	Active, Child bool
	Order         int
}

type navigation []navLink

// Len is the number of elements in the navigation.
func (n navigation) Len() int {
	return len(n)
}

// Less returns whether the element with index i should sort
// before the element with index j.
func (n navigation) Less(i, j int) bool {
	return n[i].Order < n[j].Order || (n[i].Order == n[j].Order &&
		n[i].Name < n[j].Name)
}

// Swap swaps the elements with indexes i and j.
func (n *navigation) Swap(i, j int) {
	(*n)[i], (*n)[j] = (*n)[j], (*n)[i]
}

// getShortTitle returns the given node's ShortTitle attribute, or, if the
// ShortTitle is of zero length, its Title attribute.
func getShortTitle(node node.Node) string {
	if len(node.ShortTitle) > 0 {
		return node.ShortTitle
	}
	return node.Title
}

// getNav returns the navigation for the given node.
// 
// nodePath is the absolute path of the node for which to get the navigation.
// active is the absolute path to the currently active node.
// root is the path of the data directory.
func getNav(nodePath, active string, root string) (navLinks navigation,
	err error) {
	// Search children
	children, err := ioutil.ReadDir(filepath.Join(root, nodePath))
	if err != nil {
		return nil, fmt.Errorf("Could not read node directory: %v", err)
	}
	anyChild := false
	childrenNavLinks := navLinks[:]
	for _, child := range children {
		if !child.IsDir() {
			continue
		}
		node, err := lookupNode(root, path.Join(nodePath,
			child.Name()))
		if err != nil || node.Hide {
			continue
		}
		anyChild = true
		childrenNavLinks = append(childrenNavLinks, navLink{
			Name:   getShortTitle(node),
			Target: child.Name(), Child: true, Order: node.Order})
	}
	if !anyChild {
		if nodePath == "/" || path.Dir(nodePath) == "/" {
			return nil, nil
		}
		return getNav(path.Dir(nodePath), active, root)
	}
	sort.Sort(&childrenNavLinks)
	siblingsNavLinks := navLinks[:]
	// Search siblings
	if nodePath != "/" && path.Dir(nodePath) == "/" {
		node, err := lookupNode(root, nodePath)
		if err != nil {
			return nil, fmt.Errorf("Could not find node: %v", err)
		}
		siblingsNavLinks = append(siblingsNavLinks, navLink{
			Name:   getShortTitle(node),
			Target: path.Join("..", path.Base(nodePath)), Order: node.Order})
	} else if nodePath != "/" {
		parent := path.Dir(nodePath)
		siblings, err := ioutil.ReadDir(filepath.Join(root, parent))
		if err != nil {
			return nil, fmt.Errorf("Could not read node directory: %v", err)
		}
		for _, sibling := range siblings {
			if !sibling.IsDir() {
				continue
			}
			node, err := lookupNode(root, path.Join(parent,
				sibling.Name()))
			if err != nil || node.Hide {
				continue
			}
			siblingsNavLinks = append(siblingsNavLinks, navLink{
				Name:   getShortTitle(node),
				Target: path.Join("..", sibling.Name()), Order: node.Order})
		}
	}
	sort.Sort(&siblingsNavLinks)
	// Insert children at their parent
	for i, link := range siblingsNavLinks {
		if link.Target == path.Join("..", path.Base(nodePath)) {
			navLinks = append(navLinks, siblingsNavLinks[:i+1]...)
			navLinks = append(navLinks, childrenNavLinks...)
			navLinks = append(navLinks, siblingsNavLinks[i+1:]...)
			break
		}
	}
	if len(navLinks) == 0 {
		navLinks = childrenNavLinks
	}
	// Compute node paths relative to active node and search and set the Active
	// link
	for i, link := range navLinks {
		rel, err := filepath.Rel(active, path.Join(nodePath, link.Target))
		if err != nil {
			panic(fmt.Sprint("Could not comute relative path:", err))
		}
		navLinks[i].Target = rel
		if rel == "." || (len(rel) >= 5 && rel[:2] == ".." && rel[len(rel)-2:] == "..") {
			navLinks[i].Active = true
		}
	}
	return
}

// MakeAbsolute converts relative targets to absolute ones by adding the given
// root path. It also adds a trailing slash.
func (nav *navigation) MakeAbsolute(root string) {
	for i := range *nav {
		if (*nav)[i].Target[0] != '/' {
			(*nav)[i].Target = path.Join(root, (*nav)[i].Target)
		}
		(*nav)[i].Target = (*nav)[i].Target + "/"
	}
}

type addFormData struct {
	Type, Name, Title string
}

// Add handles add requests.
func (h *nodeHandler) Add(w http.ResponseWriter, r *http.Request,
	reqnode node.Node, session *sessions.Session, cSession *login.Session,
	site site) {
	G := l10n.UseCatalog(cSession.Locale)
	data := addFormData{}
	nodeTypeOptions := []form.Option{}
	/*
		for _, nodeType := range h.Settings.NodeTypes {
			nodeTypeOptions = append(nodeTypeOptions,
				form.Option{nodeType, nodeType})
		}
	*/
	selectWidget := form.SelectWidget{nodeTypeOptions}
	form := form.NewForm(&data, form.Fields{
		"Type": form.Field{G("Content type"), "", form.Required(G("Required.")), selectWidget},
		"Name": form.Field{G("Name"),
			G("The name as it should appear in the URL."),
			form.And(form.Required(G("Required.")), form.Regex(`^[-\w]*$`,
				G("Contains	invalid characters."))), nil},
		"Title": form.Field{G("Title"), "", form.Required(G("Required.")), nil}})
	switch r.Method {
	case "GET":
	case "POST":
		r.ParseForm()
		if form.Fill(r.Form) {
			data.Name = strings.ToLower(data.Name)
			/*
				if !inStringSlice(data.Type, h.Settings.NodeTypes) {
					panic("Can't add this content type.")
				}
			*/
			newPath := filepath.Join(reqnode.Path, data.Name)
			newNode := node.Node{
				Path:  newPath,
				Type:  data.Type,
				Title: data.Title}
			data, err := h.Info.FindDataService()
			if err != nil {
				panic("Can't find data service: " + err.Error())
			}
			if err := data.UpdateNode(site.Name, newNode); err != nil {
				panic("Can't add node: " + err.Error())
			}
			http.Redirect(w, r, newPath+"/@@edit", http.StatusSeeOther)
			return
		}
	default:
		panic("Request method not supported: " + r.Method)
	}
	body := h.Renderer.Render("daemon/actions/addform", template.Context{
		"Form": form.RenderData()}, cSession.Locale, site.Directories.Templates)
	env := masterTmplEnv{Node: reqnode, Session: cSession,
		Flags: EDIT_VIEW, Title: G("Add content")}
	fmt.Fprint(w, renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		site, cSession.Locale))
}

type removeFormData struct {
	Confirm int
}

// Remove handles remove requests.
func (h *nodeHandler) Remove(w http.ResponseWriter, r *http.Request,
	node node.Node, session *sessions.Session, cSession *login.Session,
	site site) {
	G := l10n.UseCatalog(cSession.Locale)
	data := removeFormData{}
	form := form.NewForm(&data, form.Fields{
		"Confirm": form.Field{G("Confirm"), "", form.Required(G("Required.")),
			new(form.HiddenWidget)}})
	switch r.Method {
	case "GET":
	case "POST":
		r.ParseForm()
		if form.Fill(r.Form) {
			removeNode(node.Path, site.Directories.Data)
			http.Redirect(w, r, path.Dir(node.Path), http.StatusSeeOther)
			return
		}
	default:
		panic("Request method not supported: " + r.Method)
	}
	data.Confirm = 1489
	body := h.Renderer.Render("daemon/actions/removeform", template.Context{
		"Form": form.RenderData(), "Node": node},
		cSession.Locale, site.Directories.Templates)
	env := masterTmplEnv{Node: node, Session: cSession,
		Flags: EDIT_VIEW, Title: fmt.Sprintf(G("Remove \"%v\""), node.Title)}
	fmt.Fprint(w, renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		site, cSession.Locale))
}

// lookupNode look ups a node at the given path.
// If no such node exists, return nil.
func lookupNode(root, path string) (node.Node, error) {
	node_path := filepath.Join(root, path[1:], "node.yaml")
	content, err := ioutil.ReadFile(node_path)
	if err != nil {
		return node.Node{}, err
	}
	var reqnode node.Node
	if err = goyaml.Unmarshal(content, &reqnode); err != nil {
		return node.Node{}, err
	}
	reqnode.Path = path
	return reqnode, nil
}

// removeNode recursively removes the given node from the data directory located
// at the given root and from the navigation of the parent node.
func removeNode(path, root string) {
	nodePath := filepath.Join(root, path[1:])
	if err := os.RemoveAll(nodePath); err != nil {
		panic("Can't remove node: " + err.Error())
	}
}
