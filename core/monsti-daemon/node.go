// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann
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
	"bytes"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chrneumann/htmlwidgets"
	"github.com/nfnt/resize"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util/nodes"
	mtemplate "pkg.monsti.org/monsti/api/util/template"
)

// navLink represents a link in the navigation.
type navLink struct {
	Name, Target        string
	Active, ActiveBelow bool
	Order, Level        int
	Children            navigation
}

func (n navLink) Child() bool {
	return n.Level > 0
}

type navigation []navLink

// MakeAbsolute converts relative targets to absolute ones by adding the given
// root path. It also adds a trailing slash.
func (nav *navigation) MakeAbsolute(root string) {
	for i := range *nav {
		if (*nav)[i].Target[0] != '/' {
			(*nav)[i].Target = path.Join(root, (*nav)[i].Target)
		}
		if !strings.HasSuffix((*nav)[i].Target, "/") {
			(*nav)[i].Target = (*nav)[i].Target + "/"
		}
	}
}

// getNodeTitle tries to get a title of a node
func getNodeTitle(node *service.Node) string {
	title := "Untitled"
	if node.Fields["core.Title"] != nil {
		title = node.Fields["core.Title"].Value().(string)
	}
	return title
}

type getNodeFunc func(path string) (*service.Node, error)
type getChildrenFunc func(path string) ([]*service.Node, error)

func getNavLinks(node *service.Node, includeRoot, public bool, getNodeFn getNodeFunc,
	getChildrenFn getChildrenFunc, depth int) (
	navLinks navigation, err error) {
	return getNavLinksRec(node, includeRoot, public, getNodeFn, getChildrenFn, depth,
		0)
}

func getNavLinksRec(node *service.Node, includeRoot, public bool, getNodeFn getNodeFunc,
	getChildrenFn getChildrenFunc, depth int, level int) (
	navigation, error) {

	if depth < 0 || node.Hide || node.Type.Hide || public && !node.Public {
		return nil, nil
	}

	// Search children and sort them by their orders and names.
	children, err := getChildrenFn(node.Path)
	if err != nil {
		return nil, fmt.Errorf("Could not get children: %v", err)
	}
	sort.Sort(&nodes.Sorter{children, func(left, right *service.Node) bool {
		return left.Order < right.Order || (left.Order == right.Order &&
			getNodeTitle(left) < getNodeTitle(right))
	}})
	var childNavLinks navigation
	for _, child := range children {
		tmpNavLinks, err := getNavLinksRec(child, true, public, getNodeFn,
			getChildrenFn, depth-1, level+1)
		if err != nil {
			return nil, err
		}
		childNavLinks = append(childNavLinks, tmpNavLinks...)
	}

	if includeRoot {
		return navigation{
			navLink{
				Name:     getNodeTitle(node),
				Target:   node.Path,
				Order:    node.Order,
				Level:    level,
				Children: childNavLinks,
			}}, nil
	}
	return childNavLinks, nil
}

// getNav returns the navigation for the given node.
//
// If public is true, show only public pages.
// nodePath is the absolute path of the node for which to get the navigation.
// active is the absolute path to the currently active node.
// Descends until the given depth.
func getNav(nodePath, active string, public bool,
	getNodeFn getNodeFunc, getChildrenFn getChildrenFunc, depth int) (
	navLinks navigation, err error) {

	node, err := getNodeFn(nodePath)
	if err != nil {
		return nil, fmt.Errorf("Could not get node: %v", err)
	}

	// Search children
	children, err := getChildrenFn(nodePath)
	if err != nil {
		return nil, fmt.Errorf("Could not get children: %v", err)
	}

	// If there are no children, return the navigation for the parent
	// node unless this or the parent node is the root node. In the
	// first case, there is no parent, and in the second case, this
	// often makes no sense as the navigation would be similar to the
	// main navigation.
	//
	// In this case, the navigation of the parent node gives more
	// context than only listing this node with its siblings.
	if nodePath == "/" || path.Dir(nodePath) == "/" {
		navLinks, err = getNavLinks(node, nodePath == "/" || len(children) > 0,
			public, getNodeFn, getChildrenFn, depth)
		if nodePath == "/" && len(navLinks) > 0 {
			navLinks = append(navLinks, navLinks[0].Children...)
			navLinks[0].Children = nil
		}
	} else if len(children) == 0 {
		parent, err := getNodeFn(path.Dir(nodePath))
		if err != nil {
			return nil, fmt.Errorf("Could not get parent node: %v", err)
		}
		navLinks, err = getNavLinks(parent, true, public, getNodeFn,
			getChildrenFn, depth)
	} else {
		navLinks, err = getNavLinks(node, true, public, getNodeFn,
			getChildrenFn, depth)
	}
	if err != nil {
		return nil, err
	}

	// Compute node paths relative to active node and search and set the Active
	// link
	var relPaths func(navigation)
	relPaths = func(nav navigation) {
		for i, link := range nav {
			if active == link.Target {
				nav[i].Active = true
			} else if strings.HasPrefix(active, link.Target) {
				nav[i].ActiveBelow = true
			}
			relPaths(link.Children)
		}
	}
	relPaths(navLinks)

	return
}

// Add handles add requests.
func (h *nodeHandler) Add(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)
	nodeTypeIds, err := c.Serv.Monsti().GetAddableNodeTypes(c.Site,
		c.Node.Type.Id)
	if err != nil {
		return fmt.Errorf("Could not get addable node types: %v", err)
	}
	var nodeTypes []*service.NodeType
	for _, id := range nodeTypeIds {
		nodeType, err := c.Serv.Monsti().GetNodeType(id)
		if err != nil {
			return fmt.Errorf("Could not get node type: %v", err)
		}
		nodeTypes = append(nodeTypes, nodeType)
	}
	body, err := h.Renderer.Render("actions/add", mtemplate.Context{
		"Session":   c.UserSession,
		"NodeTypes": nodeTypes}, c.UserSession.Locale,
		h.Settings.Monsti.GetSiteTemplatesPath(c.Site))
	if err != nil {
		return fmt.Errorf("Can't render add action template: %v", err)
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession,
		Flags: EDIT_VIEW, Title: G("New child node")}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		c.Site, c.SiteSettings, c.UserSession.Locale, c.Serv)
	c.Res.Write(rendered)
	return nil
}

type removeFormData struct {
	Confirm string
}

// Remove handles remove requests.
func (h *nodeHandler) Remove(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)
	data := removeFormData{}
	form := htmlwidgets.NewForm(&data)
	form.AddWidget(new(htmlwidgets.HiddenWidget), "Confirm", G("Confirm"), "")
	switch c.Req.Method {
	case "GET":
		data.Confirm = "ok"
	case "POST":
		if form.Fill(c.Req.Form) && data.Confirm == "ok" {
			if err := c.Serv.Monsti().RemoveNode(c.Site, c.Node.Path); err != nil {
				return fmt.Errorf("Could not remove node: %v", err)
			}
			http.Redirect(c.Res, c.Req, path.Dir(c.Node.Path), http.StatusSeeOther)
			return nil
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	body, err := h.Renderer.Render("actions/removeform", mtemplate.Context{
		"Form": form.RenderData(), "Node": c.Node},
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site))
	if err != nil {
		panic("Can't render node remove formular: " + err.Error())
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession,
		Flags: EDIT_VIEW, Title: G("Remove node")}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		c.Site, c.SiteSettings, c.UserSession.Locale, c.Serv)
	c.Res.Write(rendered)
	return nil
}

type imageSize struct{ Width, Height uint }

func (s imageSize) String() string {
	return fmt.Sprintf("%vx%v", s.Width, s.Height)
}

// viewImage sends a possibly resized image
func (h *nodeHandler) viewImage(c *reqContext) error {
	sizeName := c.Req.FormValue("size")
	var size imageSize
	var body []byte
	var err error
	if sizeName != "" {
		var err error
		if sizeName == "core.ChooserThumbnail" {
			size.Width = 150
			size.Height = 150
		} else {
			style, ok := c.SiteSettings.Fields["core.ImageStyles"].(*service.MapField).
				Fields[sizeName]
			if ok {
				size.Width = uint(style.(*service.CombinedField).
					Fields["width"].Value().(int))
				size.Height = uint(style.(*service.CombinedField).
					Fields["height"].Value().(int))
			}
		}
		if err != nil || size.Width == 0 {
			if err != nil {
				h.Log.Printf("Could not get size config: %v", err)
			} else {
				h.Log.Printf("Could not find size %q for site %q: %v", sizeName,
					c.Site, err)
			}
		} else {
			cacheId := "core.image.thumbnail." + size.String()
			body, _, err = c.Serv.Monsti().FromCache(c.Site, c.Node.Path, cacheId)
			if err != nil {
				return fmt.Errorf("Could not get thumbnail from cache: %v", err)
			}
			if body == nil {
				body, err = c.Serv.Monsti().GetNodeData(c.Site, c.Node.Path,
					"__file_core.File")
				if err != nil {
					return fmt.Errorf("Could not get image data: %v", err)
				}
				image, format, err := image.Decode(bytes.NewBuffer(body))
				if err != nil {
					return fmt.Errorf("Could not decode image data: %v", err)
				}
				image = resize.Thumbnail(size.Width, size.Height, image,
					resize.Lanczos3)
				var out bytes.Buffer
				switch format {
				case "png", "gif":
					err = png.Encode(&out, image)
				default:
					err = jpeg.Encode(&out, image, nil)
				}
				body = out.Bytes()
				if err != nil {
					return fmt.Errorf("Could not encode resized image: %v", err)
				}
				if err := c.Serv.Monsti().ToCache(c.Site, c.Node.Path,
					cacheId, body,
					&service.CacheMods{Deps: []service.CacheDep{{Node: c.Node.Path}}}); err != nil {
					return fmt.Errorf("Could not cache resized image data: %v", err)
				}
			}
		}
	}
	if body == nil {
		body, err = c.Serv.Monsti().GetNodeData(c.Site, c.Node.Path,
			"__file_core.File")
		if err != nil {
			return fmt.Errorf("Could not read image: %v", err)
		}
	}
	c.Res.Write(body)
	return nil
}

// ViewNode handles node views.
func (h *nodeHandler) View(c *reqContext) error {
	// Redirect if trailing slash is missing and if this is not a file
	// node (in which case we write out the file's content).
	if c.Node.Path[len(c.Node.Path)-1] != '/' {
		if c.Node.Type.Id == "core.Image" || c.Node.Type.Id == "core.File" {
			c.Res.Header().Add("Last-Modified", c.Node.Changed.Format(time.RFC1123))
		}
		if c.Node.Type.Id == "core.Image" {
			return h.viewImage(c)
		} else if c.Node.Type.Id == "core.File" {
			content, err := c.Serv.Monsti().GetNodeData(c.Site, c.Node.Path,
				"__file_core.File")
			if err != nil {
				return fmt.Errorf("Could not read file: %v", err)
			}
			c.Res.Write(content)
		} else {
			newPath, err := url.Parse(c.Node.Path + "/")
			if err != nil {
				serveError("Could not parse request URL: %v", err)
			}
			url := c.Req.URL.ResolveReference(newPath)
			http.Redirect(c.Res, c.Req, url.String(), http.StatusSeeOther)
		}
		return nil
	}

	var rendered []byte
	var err error
	mods := new(service.CacheMods)
	if c.UserSession.User == nil && len(c.Req.Form) == 0 {
		rendered, mods, err = c.Serv.Monsti().FromCache(c.Site, c.Node.Path,
			"core.page.partial")
		if err != nil {
			return fmt.Errorf("Could not get partial cache: %v", err)
		}
	}
	if rendered == nil {
		rendered, mods, err = h.RenderNode(c, nil)
		if err != nil {
			return fmt.Errorf("Could not render node: %v", err)
		}
		if c.UserSession.User == nil && len(c.Req.Form) == 0 {
			if err := c.Serv.Monsti().ToCache(c.Site, c.Node.Path,
				"core.page.partial", rendered, mods); err != nil {
				return fmt.Errorf("Could not cache page: %v", err)
			}
		}
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession}
	content, renderMods := renderInMaster(h.Renderer, rendered, env, h.Settings,
		c.Site, c.SiteSettings, c.UserSession.Locale, c.Serv)
	mods.Join(renderMods)
	if c.UserSession.User == nil && len(c.Req.Form) == 0 {
		if err := c.Serv.Monsti().ToCache(c.Site, c.Node.Path,
			"core.page.full", content, mods); err != nil {
			return fmt.Errorf("Could not cache page: %v", err)
		}
	}
	c.Res.Write(content)
	return nil
}

// calcEmbedPath calculates the embed path for the given node path and
// embed URI.
func calcEmbedPath(nodePath, embedURI string) (string, error) {
	embedURL, err := url.Parse(embedURI)
	if err != nil {
		return "", fmt.Errorf("Could not parse embed URI: %v", err)
	}
	if path.IsAbs(embedURL.Path) {
		return embedURL.Path, nil
	}
	return path.Join(nodePath, embedURL.Path), nil
}

// RenderNode renders a requested node.
//
// If embedNode is not null, render the given node that is embedded
// into the node given by the request.
func (h *nodeHandler) RenderNode(c *reqContext, embedNode *service.EmbedNode) (
	[]byte, *service.CacheMods, error) {
	mods := &service.CacheMods{Deps: []service.CacheDep{{Node: c.Node.Path}}}
	reqNode := c.Node
	if embedNode != nil {
		embedPath, err := calcEmbedPath(reqNode.Path, embedNode.URI)
		if err != nil {
			return nil, nil, fmt.Errorf("Could not get calculate path: %v", err)
		}
		reqNode, err = c.Serv.Monsti().GetNode(c.Site, embedPath)
		if err != nil || reqNode == nil {
			return nil, nil, fmt.Errorf("Could not find node %q to embed: %v",
				embedPath, err)
		}
	}
	context := make(mtemplate.Context)
	context["Embed"] = make(map[string]template.HTML)
	// Embed nodes
	embedNodes := append(reqNode.Type.Embed, reqNode.Embed...)
	for _, embed := range embedNodes {
		rendered, renderMods, err := h.RenderNode(c, &embed)
		mods.Join(renderMods)
		if err != nil {
			return nil, nil, fmt.Errorf("Could not render embed node: %v", err)
		}
		context["Embed"].(map[string]template.HTML)[embed.Id] =
			template.HTML(rendered)
	}
	context["Node"] = reqNode
	switch reqNode.Type.Id {
	case "core.ContactForm":
		if err := renderContactForm(c, context, c.Req.Form, h); err != nil {
			return nil, nil, fmt.Errorf("Could not render contact form: %v", err)
		}
	}
	context["Embedded"] = embedNode != nil

	var ret []service.NodeContextRet
	err := c.Serv.Monsti().EmitSignal("monsti.NodeContext",
		service.NodeContextArgs{c.Id, reqNode.Type.Id, embedNode}, &ret)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not emit signal: %v", err)
	}
	for i, _ := range ret {
		mods.Join(ret[i].Mods)
		for key, value := range ret[i].Context {
			context[key] = template.HTML(value)
		}
	}

	template := strings.Replace(reqNode.Type.Id, ".", "/", 1) + "-view"

	context["Site"] = c.Site
	rendered, err := h.Renderer.Render(template, context,
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site))
	if err != nil {
		return nil, nil, fmt.Errorf("Could not render template: %v", err)
	}
	return rendered, mods, nil
}

type editFormData struct {
	NodeType string
	Name     string
	Node     service.Node
	Fields   service.NestedMap
}

// EditNode handles node edits.
func (h *nodeHandler) Edit(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)

	if err := c.Req.ParseMultipartForm(1024 * 1024); err != nil {
		if err != http.ErrNotMultipart {
			return fmt.Errorf("Could not parse form: %v", err)
		}
	}

	nodeType := c.Node.Type
	newNode := c.Req.Form.Get("NodeType") != ""
	if newNode {
		var err error
		nodeType, err = c.Serv.Monsti().GetNodeType(c.Req.FormValue("NodeType"))
		if err != nil {
			return fmt.Errorf("Could not get node type to add %q: %v",
				c.Req.FormValue("new"), err)
		}
		// TODO Check if node type may be added to this node
	}

	env := masterTmplEnv{Node: c.Node, Session: c.UserSession}

	if c.Action == service.EditAction {
		if newNode {
			env.Title = G("Add a new node")
		} else {
			env.Title = G("Edit node")
		}
		env.Flags = EDIT_VIEW
	}

	formData := editFormData{}
	formData.Fields = make(service.NestedMap)
	if newNode {
		formData.NodeType = nodeType.Id
		formData.Node.Type = nodeType
		err := formData.Node.InitFields(c.Serv.Monsti(), c.Site)
		if err != nil {
			return fmt.Errorf("Could not init node fields: %v", err)
		}
		formData.Node.PublishTime = time.Now().UTC()
		formData.Node.Public = true
	} else {
		formData.Node = *c.Node
	}
	form := htmlwidgets.NewForm(&formData)
	form.AddWidget(new(htmlwidgets.HiddenWidget), "NodeType", "", "")
	if !nodeType.Hide {
		form.AddWidget(new(htmlwidgets.BoolWidget), "Node.Hide", G("Hide"),
			G("Don't show node in navigation."))
	}
	form.AddWidget(new(htmlwidgets.BoolWidget), "Node.Public", G("Public"),
		G("Is the node accessible by every visitor?"))
	location, err := time.LoadLocation(c.SiteSettings.StringValue("core.Timezone"))
	if err != nil {
		location = time.UTC
	}
	form.AddWidget(&htmlwidgets.TimeWidget{
		Location: location}, "Node.PublishTime", G("Publish time"),
		G("The node won't be accessible to the public until it is published."))
	if newNode || c.Node.Name() != "" {
		form.AddWidget(&htmlwidgets.TextWidget{
			Regexp:          `^[-\w.]+$`,
			ValidationError: G("Please enter a name consisting only of the characters A-Z, a-z, 0-9, '.', and '-'")},
			"Name", G("Name"), G("The name as it should appear in the URL."))
	}
	if !newNode {
		formData.Name = c.Node.Name()
	}

	fileFields := make([]string, 0)
	for _, field := range nodeType.Fields {
		if field.Hidden {
			continue
		}
		formData.Node.Fields[field.Id].ToFormField(form, formData.Fields,
			field, c.UserSession.Locale)
		if _, ok := field.Type.(*service.FileFieldType); ok {
			fileFields = append(fileFields, field.Id)
		}
	}

	switch c.Req.Method {
	case "GET":
	case "POST":
		if form.Fill(c.Req.Form) {
			node := formData.Node
			node.Type = nodeType
			pathPrefix := node.GetPathPrefix()
			oldPath := c.Node.Path
			parentPath := c.Node.GetParentPath()
			if newNode {
				parentPath = c.Node.Path
			}
			node.Path = path.Join(parentPath, pathPrefix, formData.Name)
			renamed := !newNode && c.Node.Name() != "" && oldPath != node.Path
			writeNode := true
			if newNode || renamed {
				existing, err := c.Serv.Monsti().GetNode(c.Site, node.Path)
				if err != nil {
					return fmt.Errorf("Could not fetch possibly existing node: %v", err)
				}
				if existing != nil {
					form.AddError("Name", G("A node with this name does already exist"))
					writeNode = false
				}
				if err = node.InitFields(c.Serv.Monsti(), c.Site); err != nil {
					return fmt.Errorf("Could not init node fields: %v", err)
				}
			}

			// Check file format for image nodes.
			if nodeType.Id == "core.Image" {
				file, _, err := c.Req.FormFile("Fields.core.File")
				if err == nil {
					content, err := ioutil.ReadAll(file)
					if err != nil {
						return fmt.Errorf("Could not read multipart file: %v", err)
					}
					if _, _, err := image.Decode(bytes.NewBuffer(content)); err != nil {
						form.AddError("Fields.core.File",
							G("Unsupported image format. Try GIF, JPEG, or PNG."))
						writeNode = false
					}
				}

			}

			if writeNode {
				if renamed {
					err := c.Serv.Monsti().RenameNode(c.Site, c.Node.Path, node.Path)
					if err != nil {
						return fmt.Errorf("Could not move node: %v", err)
					}
				}
				for _, field := range nodeType.Fields {
					if !field.Hidden {
						node.Fields[field.Id].FromFormField(formData.Fields, field)
					}
				}
				err := c.Serv.Monsti().WriteNode(c.Site, node.Path, &node)
				if err != nil {
					return fmt.Errorf("Could not update node: %v", err)
				}

				// Save any attached files
				if len(fileFields) > 0 && c.Req.MultipartForm != nil {
					for _, name := range fileFields {
						file, _, err := c.Req.FormFile("Fields." + name)
						if err == nil {
							content, err := ioutil.ReadAll(file)
							if err != nil {
								return fmt.Errorf("Could not read multipart file: %v", err)
							}
							if err = c.Serv.Monsti().WriteNodeData(c.Site, node.Path,
								"__file_"+name, content); err != nil {
								return fmt.Errorf("Could not save file: %v", err)
							}
						}
					}
				}
				http.Redirect(c.Res, c.Req, node.Path+"/", http.StatusSeeOther)
				err = c.Serv.Monsti().MarkDep(
					c.Site, service.CacheDep{Node: path.Clean(node.Path)})
				if err != nil {
					return fmt.Errorf("Could not mark node: %v", err)
				}
				return nil
			}
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	rendered, err := h.Renderer.Render("edit",
		mtemplate.Context{"Form": form.RenderData()},
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site))

	if err != nil {
		return fmt.Errorf("Could not render template: %v", err)
	}

	content, _ := renderInMaster(h.Renderer, []byte(rendered), env, h.Settings,
		c.Site, c.SiteSettings, c.UserSession.Locale, c.Serv)

	c.Res.Write(content)
	return nil
}

type orderedNodes []*service.Node

func (n orderedNodes) Len() int {
	return len(n)
}

func (n orderedNodes) Less(i, j int) bool {
	return n[i].Order < n[j].Order
}

func (n orderedNodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// List handles list requests.
func (h *nodeHandler) List(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)
	m := c.Serv.Monsti()
	children, err := m.GetChildren(c.Site, c.Node.Path)
	if err != nil {
		return fmt.Errorf("Could not get children of node: %v", err)
	}
	switch c.Req.Method {
	case "GET":
	case "POST":
		for _, child := range children {
			if vals, ok := c.Req.Form["order-"+child.Name()]; ok && len(vals) == 1 {
				oldOrder := child.Order
				if order, err := strconv.Atoi(vals[0]); err == nil {
					child.Order = order
				}
				if oldOrder != child.Order {
					err := c.Serv.Monsti().WriteNode(c.Site, child.Path, child)
					if err != nil {
						return fmt.Errorf("Could not update node: %v", err)
					}
				}
			}
		}
		http.Redirect(c.Res, c.Req, path.Join(c.Node.Path, "/"), http.StatusSeeOther)
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	var parent interface{}
	parentPath := path.Dir(c.Node.Path)
	if parentPath != c.Node.Path {
		var err error
		if parent, err = m.GetNode(c.Site, parentPath); err != nil {
			return fmt.Errorf("Could not get parent of node: %v", err)
		}
	}
	sort.Sort(orderedNodes(children))
	body, err := h.Renderer.Render("actions/list", mtemplate.Context{
		"Parent":   parent,
		"Children": children,
		"Node":     c.Node},
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site))
	if err != nil {
		return fmt.Errorf("Can't render node list: %v", err)
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession,
		Flags: EDIT_VIEW, Title: G("List child nodes")}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		c.Site, c.SiteSettings, c.UserSession.Locale, c.Serv)
	c.Res.Write(rendered)
	return nil
}

type splittedPathElement struct {
	Name, Path string
	Parent     bool
}

// getSplittedPath returns the path splitted into its elements.
func getSplittedPath(nodePath string) (ret []splittedPathElement) {
	elements := strings.SplitAfter(nodePath, "/")
	if nodePath == "/" {
		elements = elements[:1]
	}
	path := ""
	for i, v := range elements {
		path = path + v
		if v != "/" && i != len(elements)-1 {
			v = v[:len(v)-1]
		}
		ret = append(ret, splittedPathElement{v, path, i != len(elements)-1})
	}
	return
}

// Chooser handles chooser requests.
func (h *nodeHandler) Chooser(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)
	m := c.Serv.Monsti()
	children, err := m.GetChildren(c.Site, c.Node.Path)
	if err != nil {
		return fmt.Errorf("Could not get children of node: %v", err)
	}
	switch c.Req.Method {
	case "GET":
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	var parent interface{}
	parentPath := path.Dir(c.Node.Path)
	if parentPath != c.Node.Path {
		var err error
		if parent, err = m.GetNode(c.Site, parentPath); err != nil {
			return fmt.Errorf("Could not get parent of node: %v", err)
		}
	}
	sort.Sort(orderedNodes(children))
	chooseType := c.Req.FormValue("type")
	context := mtemplate.Context{
		"SplittedPath": getSplittedPath(c.Node.Path),
		"Type":         chooseType,
		"Parent":       parent,
		"Children":     children,
		"Node":         c.Node}

	if chooseType == "image" {
		images := make([]*service.Node, 0)
		for _, node := range children {
			if node.Type.Id == "core.Image" {
				images = append(images, node)
			}
		}
		context["Images"] = images
	}

	body, err := h.Renderer.Render("actions/chooser", context,
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site))
	if err != nil {
		return fmt.Errorf("Can't render node chooser: %v", err)
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession,
		Flags: EDIT_VIEW | SLIM_VIEW,
		Title: G("Node chooser")}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		c.Site, c.SiteSettings, c.UserSession.Locale, c.Serv)
	c.Res.Write(rendered)
	return nil
}
