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
	"pkg.monsti.org/monsti/api/util"
	mtemplate "pkg.monsti.org/monsti/api/util/template"
)

// navLink represents a link in the navigation.
type navLink struct {
	Name, Target               string
	Active, ActiveBelow, Child bool
	Order                      int
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

// getNodeTitle tries to get a title of a node
func getNodeTitle(node *service.Node) string {
	title := "Untitled"
	if node.Fields["core.Title"] != nil {
		title = node.Fields["core.Title"].String()
	}
	return title
}

type getNodeFunc func(path string) (*service.Node, error)
type getChildrenFunc func(path string) ([]*service.Node, error)

// getNav returns the navigation for the given node.
//
// If public is true, show only public pages.
// nodePath is the absolute path of the node for which to get the navigation.
// active is the absolute path to the currently active node.
func getNav(nodePath, active string, public bool,
	getNodeFn getNodeFunc, getChildrenFn getChildrenFunc) (
	navLinks navigation, err error) {

	// Search children
	children, err := getChildrenFn(nodePath)
	if err != nil {
		return nil, fmt.Errorf("Could not get children: %v", err)
	}
	childrenNavLinks := navLinks[:]
	for _, child := range children {
		if child.Hide || child.Type.Hide || public && !child.Public {
			continue
		}
		childrenNavLinks = append(childrenNavLinks, navLink{
			Name:   getNodeTitle(child),
			Target: path.Join(nodePath, child.Name()),
			Child:  true, Order: child.Order})
	}
	if len(childrenNavLinks) == 0 {
		if nodePath == "/" || path.Dir(nodePath) == "/" {
			return nil, nil
		}
		return getNav(path.Dir(nodePath), active, public, getNodeFn, getChildrenFn)
	}
	sort.Sort(&childrenNavLinks)
	siblingsNavLinks := navLinks[:]
	// Search siblings
	if path.Dir(nodePath) == "/" {
		node, err := getNodeFn(nodePath)
		if err != nil {
			return nil, fmt.Errorf("Could not get node: %v", err)
		}
		siblingsNavLinks = append(siblingsNavLinks, navLink{
			Name:   getNodeTitle(node),
			Target: nodePath, Order: node.Order})
	} else if nodePath != "/" {
		parent := path.Dir(nodePath)
		siblings, err := getChildrenFn(parent)
		if err != nil {
			return nil, fmt.Errorf("Could not get siblings: %v", err)
		}
		for _, sibling := range siblings {
			if sibling.Hide || sibling.Type.Hide || public && !sibling.Public {
				continue
			}
			siblingsNavLinks = append(siblingsNavLinks, navLink{
				Name:   getNodeTitle(sibling),
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
		if active == link.Target {
			navLinks[i].Active = true
		} else if strings.HasPrefix(active, link.Target) {
			navLinks[i].ActiveBelow = true
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
		if !strings.HasSuffix((*nav)[i].Target, "/") {
			(*nav)[i].Target = (*nav)[i].Target + "/"
		}
	}
}

type addFormData struct {
	NodeType string
	New      string
}

// Add handles add requests.
func (h *nodeHandler) Add(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)
	data := addFormData{New: "1"}
	nodeTypeOptions := []htmlwidgets.SelectOption{}
	nodeTypes, err := c.Serv.Monsti().GetAddableNodeTypes(c.Site.Name,
		c.Node.Type.Id)
	if err != nil {
		return fmt.Errorf("Could not get addable node types: %v", err)
	}
	for _, id := range nodeTypes {
		nodeType, err := c.Serv.Monsti().GetNodeType(id)
		if err != nil {
			return fmt.Errorf("Could not get node type: %v", err)
		}
		nodeTypeOptions = append(nodeTypeOptions,
			htmlwidgets.SelectOption{nodeType.Id,
				nodeType.Name.Get(c.UserSession.Locale), false})
	}
	form := htmlwidgets.NewForm(&data)
	form.AddWidget(&htmlwidgets.SelectWidget{Options: nodeTypeOptions},
		"NodeType", G("Content type"), "")
	form.AddWidget(new(htmlwidgets.HiddenWidget), "New", "", "")
	form.Action = path.Join(c.Node.Path, "@@edit")
	body, err := h.Renderer.Render("actions/addform", mtemplate.Context{
		"Form": form.RenderData()}, c.UserSession.Locale,
		h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		return fmt.Errorf("Can't render node add formular: %v", err)
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession,
		Flags: EDIT_VIEW, Title: G("Add content")}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv)
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
			if err := c.Serv.Monsti().RemoveNode(c.Site.Name, c.Node.Path); err != nil {
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
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		panic("Can't render node remove formular: " + err.Error())
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession,
		Flags: EDIT_VIEW, Title: fmt.Sprintf(G("Remove \"%v\""), c.Node.Name())}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv)
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
			err = c.Serv.Monsti().GetSiteConfig(c.Site.Name,
				"core.image.sizes."+sizeName, &size)
		}
		if err != nil || size.Width == 0 {
			if err != nil {
				h.Log.Printf("Could not get size config: %v", err)
			} else {
				h.Log.Printf("Could not find size %q for site %q: %v", sizeName,
					c.Site.Name, err)
			}
		} else {
			cacheId := "core.image.thumbnail." + size.String()
			body, _, err = c.Serv.Monsti().FromCache(c.Site.Name, c.Node.Path, cacheId)
			if err != nil {
				return fmt.Errorf("Could not get thumbnail from cache: %v", err)
			}
			if body == nil {
				body, err = c.Serv.Monsti().GetNodeData(c.Site.Name, c.Node.Path,
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
				if err := c.Serv.Monsti().ToCache(c.Site.Name, c.Node.Path,
					cacheId, body,
					&service.CacheMods{Deps: []service.CacheDep{{Node: c.Node.Path}}}); err != nil {
					return fmt.Errorf("Could not cache resized image data: %v", err)
				}
			}
		}
	}
	if body == nil {
		body, err = c.Serv.Monsti().GetNodeData(c.Site.Name, c.Node.Path,
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
			content, err := c.Serv.Monsti().GetNodeData(c.Site.Name, c.Node.Path,
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
		rendered, mods, err = c.Serv.Monsti().FromCache(c.Site.Name, c.Node.Path,
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
			if err := c.Serv.Monsti().ToCache(c.Site.Name, c.Node.Path,
				"core.page.partial", rendered, mods); err != nil {
				return fmt.Errorf("Could not cache page: %v", err)
			}
		}
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession}
	content, renderMods := renderInMaster(h.Renderer, rendered, env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv)
	mods.Join(renderMods)
	if c.UserSession.User == nil && len(c.Req.Form) == 0 {
		if err := c.Serv.Monsti().ToCache(c.Site.Name, c.Node.Path,
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
		reqNode, err = c.Serv.Monsti().GetNode(c.Site.Name, embedPath)
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
	if overwrite, ok := reqNode.TemplateOverwrites[template]; ok {
		template = overwrite.Template
	}

	context["Site"] = c.Site
	rendered, err := h.Renderer.Render(template, context,
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		return nil, nil, fmt.Errorf("Could not render template: %v", err)
	}
	return rendered, mods, nil
}

type editFormData struct {
	NodeType string
	Name     string
	Node     service.Node
	Fields   util.NestedMap
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
	newNode := len(c.Req.FormValue("NodeType")) > 0
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
			env.Title = fmt.Sprintf(G("Add %v to \"%s\""),
				nodeType.Name.Get(c.UserSession.Locale), c.Node.Path)
		} else {
			env.Title = fmt.Sprintf(G("Edit \"%s\""), c.Node.Path)
		}
		env.Flags = EDIT_VIEW
	}

	formData := editFormData{}
	formData.Fields = make(util.NestedMap)
	if newNode {
		formData.NodeType = nodeType.Id
		formData.Node.Type = nodeType
		err := formData.Node.InitFields(c.Serv.Monsti(), c.Site.Name)
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
	var timezone string
	err := c.Serv.Monsti().GetSiteConfig(c.Site.Name, "core.timezone", &timezone)
	if err != nil {
		return fmt.Errorf("Could not get timezone: %v", err)
	}
	location, err := time.LoadLocation(timezone)
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
	nodeFields := nodeType.Fields
	if !newNode {
		nodeFields = append(nodeFields, c.Node.LocalFields...)
	}
	for _, field := range nodeFields {
		formData.Node.GetField(field.Id).ToFormField(form, formData.Fields,
			field, c.UserSession.Locale)
		if field.Type == "File" {
			fileFields = append(fileFields, field.Id)
		}
	}

	switch c.Req.Method {
	case "GET":
	case "POST":
		if len(c.Req.FormValue("New")) == 0 && form.Fill(c.Req.Form) {
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
				existing, err := c.Serv.Monsti().GetNode(c.Site.Name, node.Path)
				if err != nil {
					return fmt.Errorf("Could not fetch possibly existing node: %v", err)
				}
				if existing != nil {
					form.AddError("Name", G("A node with this name does already exist"))
					writeNode = false
				}
				if err = node.InitFields(c.Serv.Monsti(), c.Site.Name); err != nil {
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
					err := c.Serv.Monsti().RenameNode(c.Site.Name, c.Node.Path, node.Path)
					if err != nil {
						return fmt.Errorf("Could not move node: %v", err)
					}
				}
				for _, field := range nodeFields {
					node.GetField(field.Id).FromFormField(formData.Fields, field)
				}
				err := c.Serv.Monsti().WriteNode(c.Site.Name, node.Path, &node)
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
							if err = c.Serv.Monsti().WriteNodeData(c.Site.Name, node.Path,
								"__file_"+name, content); err != nil {
								return fmt.Errorf("Could not save file: %v", err)
							}
						}
					}
				}
				http.Redirect(c.Res, c.Req, node.Path+"/", http.StatusSeeOther)
				err = c.Serv.Monsti().MarkDep(
					c.Site.Name, service.CacheDep{Node: path.Clean(node.Path)})
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
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))

	if err != nil {
		return fmt.Errorf("Could not render template: %v", err)
	}

	content, _ := renderInMaster(h.Renderer, []byte(rendered), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv)

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
	children, err := m.GetChildren(c.Site.Name, c.Node.Path)
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
					err := c.Serv.Monsti().WriteNode(c.Site.Name, child.Path, child)
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
		if parent, err = m.GetNode(c.Site.Name, parentPath); err != nil {
			return fmt.Errorf("Could not get parent of node: %v", err)
		}
	}
	sort.Sort(orderedNodes(children))
	body, err := h.Renderer.Render("actions/list", mtemplate.Context{
		"Parent":   parent,
		"Children": children,
		"Node":     c.Node},
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		return fmt.Errorf("Can't render node list: %v", err)
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession,
		Flags: EDIT_VIEW, Title: fmt.Sprintf(G("List \"%v\""), c.Node.Name())}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv)
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
	children, err := m.GetChildren(c.Site.Name, c.Node.Path)
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
		if parent, err = m.GetNode(c.Site.Name, parentPath); err != nil {
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
		c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		return fmt.Errorf("Can't render node chooser: %v", err)
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession,
		Flags: EDIT_VIEW | SLIM_VIEW,
		Title: fmt.Sprintf(G("Chooser \"%v\""), c.Node.Name())}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv)
	c.Res.Write(rendered)
	return nil
}
