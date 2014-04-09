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
	htmlT "html/template"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"pkg.monsti.org/form"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util/template"
)

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
func getShortTitle(node *service.NodeFields) string {
	if len(node.ShortTitle) > 0 {
		return node.ShortTitle
	}
	return node.Title
}

type getNodeFunc func(path string) (*service.NodeFields, error)
type getChildrenFunc func(path string) ([]*service.NodeFields, error)

// getNav returns the navigation for the given node.
//
// nodePath is the absolute path of the node for which to get the navigation.
// active is the absolute path to the currently active node.
func getNav(nodePath, active string,
	getNodeFn getNodeFunc, getChildrenFn getChildrenFunc) (
	navLinks navigation, err error) {
	// Search children
	children, err := getChildrenFn(nodePath)
	if err != nil {
		return nil, fmt.Errorf("Could not get children: %v", err)
	}
	childrenNavLinks := navLinks[:]
	for _, child := range children {
		if child.Hide {
			continue
		}
		childrenNavLinks = append(childrenNavLinks, navLink{
			Name:   getShortTitle(child),
			Target: path.Join(nodePath, child.Name()),
			Child:  true, Order: child.Order})
	}
	if len(childrenNavLinks) == 0 {
		if nodePath == "/" || path.Dir(nodePath) == "/" {
			return nil, nil
		}
		return getNav(path.Dir(nodePath), active, getNodeFn, getChildrenFn)
	}
	sort.Sort(&childrenNavLinks)
	siblingsNavLinks := navLinks[:]
	// Search siblings
	if nodePath != "/" && path.Dir(nodePath) == "/" {
		node, err := getNodeFn(nodePath)
		if err != nil {
			return nil, fmt.Errorf("Could not get node: %v", err)
		}
		siblingsNavLinks = append(siblingsNavLinks, navLink{
			Name:   getShortTitle(node),
			Target: nodePath, Order: node.Order})
	} else if nodePath != "/" {
		parent := path.Dir(nodePath)
		siblings, err := getChildrenFn(parent)
		if err != nil {
			return nil, fmt.Errorf("Could not get siblings: %v", err)
		}
		for _, sibling := range siblings {
			if sibling.Hide {
				continue
			}
			siblingsNavLinks = append(siblingsNavLinks, navLink{
				Name:   getShortTitle(sibling),
				Target: path.Join(nodePath, "..", sibling.Name()), Order: sibling.Order})
		}
	}
	sort.Sort(&siblingsNavLinks)
	// Insert children at their parent
	for i, link := range siblingsNavLinks {
		if link.Target == nodePath {
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
		if strings.Contains(active, link.Target) && path.Dir(active) != link.Target {
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
func (h *nodeHandler) Add(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("monsti-httpd", c.UserSession.Locale)
	data := addFormData{}
	nodeTypeOptions := []form.Option{}
	nodeTypes, err := h.Info.GetAddableNodeTypes(c.Site.Name, c.Node.Type)
	if err != nil {
		return fmt.Errorf("Could not get addable node types: %v", err)
	}
	for _, nodeType := range nodeTypes {
		nodeTypeOptions = append(nodeTypeOptions,
			form.Option{nodeType, nodeType})
	}
	selectWidget := form.SelectWidget{nodeTypeOptions}
	form := form.NewForm(&data, form.Fields{
		"Type": form.Field{G("Content type"), "", form.Required(G("Required.")), selectWidget},
		"Name": form.Field{G("Name"),
			G("The name as it should appear in the URL."),
			form.And(form.Required(G("Required.")), form.Regex(`^[-\w]*$`,
				G("Contains	invalid characters."))), nil},
		"Title": form.Field{G("Title"), "", form.Required(G("Required.")), nil}})
	switch c.Req.Method {
	case "GET":
	case "POST":
		c.Req.ParseForm()
		if form.Fill(c.Req.Form) {
			data.Name = strings.ToLower(data.Name)
			if !inStringSlice(data.Type, nodeTypes) {
				return fmt.Errorf("Can't add this node type.")
			}
			newPath := filepath.Join(c.Node.Path, data.Name)
			var newNode struct{ service.NodeFields }
			newNode.NodeFields = service.NodeFields{
				Path:  newPath,
				Type:  data.Type,
				Title: data.Title}
			data, err := h.Info.FindDataService()
			if err != nil {
				return fmt.Errorf("Can't find data service: %v", err)
			}
			if err := data.WriteNode(c.Site.Name, newNode.Path, newNode,
				"node"); err != nil {
				return fmt.Errorf("Can't add node: %v", err)
			}
			http.Redirect(c.Res, c.Req, newPath+"/@@edit", http.StatusSeeOther)
			return nil
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	body, err := h.Renderer.Render("httpd/actions/addform", template.Context{
		"Form": form.RenderData()}, c.UserSession.Locale,
		h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		return fmt.Errorf("Can't render node add formular: %v", err)
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession,
		Flags: EDIT_VIEW, Title: G("Add content")}
	fmt.Fprint(c.Res, renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv))
	return nil
}

type removeFormData struct {
	Confirm int
}

// Remove handles remove requests.
func (h *nodeHandler) Remove(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("monsti-httpd", c.UserSession.Locale)
	data := removeFormData{}
	form := form.NewForm(&data, form.Fields{
		"Confirm": form.Field{G("Confirm"), "", form.Required(G("Required.")),
			new(form.HiddenWidget)}})
	switch c.Req.Method {
	case "GET":
	case "POST":
		c.Req.ParseForm()
		if form.Fill(c.Req.Form) {
			dataServ, err := h.Info.FindDataService()
			if err != nil {
				return fmt.Errorf("httpd: Could not connect to data service: %v", err)
			}
			if err := dataServ.RemoveNode(c.Site.Name, c.Node.Path); err != nil {
				return fmt.Errorf("Could not remove node: %v", err)
			}
			http.Redirect(c.Res, c.Req, path.Dir(c.Node.Path), http.StatusSeeOther)
			return nil
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	data.Confirm = 1489
	body, err := h.Renderer.Render("httpd/actions/removeform", template.Context{
		"Form": form.RenderData(), "Node": c.Node},
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		panic("Can't render node remove formular: " + err.Error())
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession,
		Flags: EDIT_VIEW, Title: fmt.Sprintf(G("Remove \"%v\""), c.Node.Title)}
	fmt.Fprint(c.Res, renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv))
	return nil
}

// ViewNode handles node views.
func (h *nodeHandler) View(c *reqContext) error {
	// Setup ticket and send to workers.
	h.Log.Printf("(%v) %v %v", c.Site.Name, c.Req.Method, c.Req.URL.Path)

	_, err := h.Info.GetNodeType(c.Node.Type)
	if err != nil {
		return fmt.Errorf("Could not get node type %q: %v",
			c.Node.Type, err)
	}
	/*
		if err = c.Req.ParseMultipartForm(1024 * 1024); err != nil {
			return fmt.Errorf("Could not parse form: %v", err)
		}
		method := map[string]service.RequestMethod{
			"GET":  service.GetRequest,
			"POST": service.PostRequest,
		}[c.Req.Method]
		req := service.Request{
			Site:     c.Site.Name,
			Method:   method,
			Node:     *c.Node,
			Query:    c.Req.URL.Query(),
			Session:  *c.UserSession,
			Action:   c.Action,
			FormData: c.Req.Form,
		}

		// Attach request files
		if c.Req.MultipartForm != nil {
			if len(c.Req.MultipartForm.File) > 0 {
				req.Files = make(map[string][]service.RequestFile)
			}
			for name, fileHeaders := range c.Req.MultipartForm.File {
				if _, ok := req.Files[name]; !ok {
					req.Files[name] = make([]service.RequestFile, 0)
				}
				for _, fileHeader := range fileHeaders {
					file, err := fileHeader.Open()
					if err != nil {
						return fmt.Errorf("Could not open multipart file header: %v", err)
					}
					if osFile, ok := file.(*os.File); ok {
						req.Files[name] = append(req.Files[name], service.RequestFile{
							TmpFile: osFile.Name()})
					} else {
						content, err := ioutil.ReadAll(file)
						if err != nil {
							return fmt.Errorf("Could not read multipart file: %v", err)
						}
						req.Files[name] = append(req.Files[name], service.RequestFile{
							Content: content})
					}
				}
			}
		}

		nodeType := req.Node.Type
	*/

	dataServ := c.Serv.Data()
	body, err := dataServ.GetNodeData(c.Site.Name, c.Node.Path, "body.html")
	if err != nil {
		return fmt.Errorf("Could not fetch node data: %v", err)
	}
	rendered, err := h.Renderer.Render(c.Node.Type+"/view",
		template.Context{"Body": htmlT.HTML(body)},
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		return fmt.Errorf("Could not render template: %v", err)
	}

	/*
		G, _, _, _ := gettext.DefaultLocales.Use("monsti-httpd", c.UserSession.Locale)
		if len(res.Body) == 0 && len(res.Redirect) == 0 {
			return fmt.Errorf("Got empty response.")
		}
		if res.Node != nil {
			oldPath := c.Node.Path
			c.Node = res.Node
			c.Node.Path = oldPath
		}
		if len(res.Redirect) > 0 {
			http.Redirect(c.Res, c.Req, res.Redirect, http.StatusSeeOther)
			return nil
		}
	*/
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession}
	/*
		if c.Action == service.EditAction {
			env.Title = fmt.Sprintf(G("Edit \"%s\""), c.Node.Title)
			env.Flags = EDIT_VIEW
		}
	*/
	var content []byte
	/*
		if res.Raw {
			content = res.Body
		} else {
	*/
	content = []byte(renderInMaster(h.Renderer, []byte(rendered), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv))
	/*
		}*/
	err = c.Session.Save(c.Req, c.Res)
	if err != nil {
		return fmt.Errorf("Could not save user session: %v", err)
	}
	c.Res.Write(content)
	return nil
}

/*
	// Setup ticket and send to workers.
	h.Log.Printf("(%v) %v %v", c.Site.Name, c.Req.Method, c.Req.URL.Path)

	nodeServ, err := h.Info.FindNodeService(c.Node.Type)
	if err != nil {
		return fmt.Errorf("Could not find node service for %q: %v",
			c.Node.Type, err)
	}
	defer func() {
		if err := nodeServ.Close(); err != nil {
			panic(fmt.Errorf("Could not close connection to node service: %v", err))
		}
	}()
	if err = c.Req.ParseMultipartForm(1024 * 1024); err != nil {
		return fmt.Errorf("Could not parse form: %v", err)
	}
	method := map[string]service.RequestMethod{
		"GET":  service.GetRequest,
		"POST": service.PostRequest,
	}[c.Req.Method]
	req := service.Request{
		Site:     c.Site.Name,
		Method:   method,
		Node:     *c.Node,
		Query:    c.Req.URL.Query(),
		Session:  *c.UserSession,
		Action:   c.Action,
		FormData: c.Req.Form,
	}

	// Attach request files
	if c.Req.MultipartForm != nil {
		if len(c.Req.MultipartForm.File) > 0 {
			req.Files = make(map[string][]service.RequestFile)
		}
		for name, fileHeaders := range c.Req.MultipartForm.File {
			if _, ok := req.Files[name]; !ok {
				req.Files[name] = make([]service.RequestFile, 0)
			}
			for _, fileHeader := range fileHeaders {
				file, err := fileHeader.Open()
				if err != nil {
					return fmt.Errorf("Could not open multipart file header: %v", err)
				}
				if osFile, ok := file.(*os.File); ok {
					req.Files[name] = append(req.Files[name], service.RequestFile{
						TmpFile: osFile.Name()})
				} else {
					content, err := ioutil.ReadAll(file)
					if err != nil {
						return fmt.Errorf("Could not read multipart file: %v", err)
					}
					req.Files[name] = append(req.Files[name], service.RequestFile{
						Content: content})
				}
			}
		}
	}

	res, err := nodeServ.Request(&req)
	if err != nil {
		return fmt.Errorf("Could not request node: %v", err)
	}

	G, _, _, _ := gettext.DefaultLocales.Use("monsti-httpd", c.UserSession.Locale)
	if len(res.Body) == 0 && len(res.Redirect) == 0 {
		return fmt.Errorf("Got empty response.")
	}
	if res.Node != nil {
		oldPath := c.Node.Path
		c.Node = res.Node
		c.Node.Path = oldPath
	}
	if len(res.Redirect) > 0 {
		http.Redirect(c.Res, c.Req, res.Redirect, http.StatusSeeOther)
		return nil
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession}
	if c.Action == service.EditAction {
		env.Title = fmt.Sprintf(G("Edit \"%s\""), c.Node.Title)
		env.Flags = EDIT_VIEW
	}
	var content []byte
	if res.Raw {
		content = res.Body
	} else {
		content = []byte(renderInMaster(h.Renderer, res.Body, env, h.Settings,
			*c.Site, c.UserSession.Locale, c.Serv))
	}
	err = c.Session.Save(c.Req, c.Res)
	if err != nil {
		return fmt.Errorf("Could not save user session: %v", err)
	}
	c.Res.Write(content)
	return nil
*/
