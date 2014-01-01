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
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/service"
	"pkg.monsti.org/util"
	"pkg.monsti.org/util/template"
)

// Context holds information about a request
type reqContext struct {
	Res         http.ResponseWriter
	Req         *http.Request
	Node        *service.NodeInfo
	Action      service.Action
	Session     *sessions.Session
	UserSession *service.UserSession
	Site        *util.SiteSettings
	Serv        *service.Session
}

// nodeHandler is a net/http handler to process incoming HTTP requests.
type nodeHandler struct {
	Renderer template.Renderer
	Settings *settings
	// Hosts is a map from hosts to site names.
	Hosts map[string]string
	// Log is the logger used by the node handler.
	Log *log.Logger
	// Info is a connection to an INFO service.
	Info     *service.InfoClient
	Sessions *service.SessionPool
}

// splitAction splits and returns the path and @@action of the given URL.
func splitAction(path string) (string, string) {
	tokens := strings.Split(path, "/")
	last := tokens[len(tokens)-1]
	var action string
	if len(last) > 2 && last[:2] == "@@" {
		action = last[2:]
	}
	nodePath := path
	if len(action) > 0 {
		nodePath = path[:len(path)-(len(action)+3)]
		if len(nodePath) == 0 {
			nodePath = "/"
		}
	}
	return nodePath, action
}

type ServeError string

func (err ServeError) Error() string {
	return string(err)
}

func serveError(args ...interface{}) {
	panic(ServeError(fmt.Sprintf(args[0].(string), args[1:]...)))
}

// ServeHTTP handles incoming HTTP requests.
func (h *nodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := reqContext{Res: w, Req: r}
	defer func() {
		if err := recover(); err != nil {
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "error: %v\n", err)
			if _, ok := err.(ServeError); !ok {
				buf.Write(debug.Stack())
			}
			h.Log.Println(buf.String())
			http.Error(c.Res, "Application error.",
				http.StatusInternalServerError)
		}
	}()
	var err error
	c.Serv, err = h.Sessions.New()
	if err != nil {
		serveError("Could not get session: %v", err)
	}
	defer h.Sessions.Free(c.Serv)
	var nodePath string
	nodePath, action := splitAction(c.Req.URL.Path)
	if len(action) == 0 && nodePath[len(nodePath)-1] != '/' {
		newPath, err := url.Parse(nodePath + "/")
		if err != nil {
			serveError("Could not parse request URL: %v", err)
		}
		url := c.Req.URL.ResolveReference(newPath)
		http.Redirect(c.Res, c.Req, url.String(), http.StatusSeeOther)
		return
	}
	c.Action = map[string]service.Action{
		"view":   service.ViewAction,
		"edit":   service.EditAction,
		"login":  service.LoginAction,
		"logout": service.LogoutAction,
		"add":    service.AddAction,
		"remove": service.RemoveAction,
	}[action]
	site_name, ok := h.Hosts[c.Req.Host]
	if !ok {
		serveError("No site found for host %v", c.Req.Host)
	}
	site := h.Settings.Monsti.Sites[site_name]
	c.Site = &site
	c.Site.Name = site_name
	c.Session, err = getSession(c.Req, *c.Site)
	if err != nil {
		serveError("Could not get session: %v", err)
	}
	defer context.Clear(c.Req)
	c.UserSession, err = getClientSession(c.Session,
		h.Settings.Monsti.GetSiteConfigPath(c.Site.Name))
	if err != nil {
		serveError("Could not get client session: %v", err)
	}
	c.UserSession.Locale = c.Site.Locale
	c.Node, err = c.Serv.Data().GetNode(c.Site.Name, nodePath)
	if err != nil {
		h.Log.Printf("Node not found: %v", err)
		c.Node = &service.NodeInfo{Path: nodePath}
		http.Error(c.Res, "Document not found", http.StatusNotFound)
		return
	}
	if !checkPermission(c.Action, c.UserSession) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}
	switch c.Action {
	case service.LoginAction:
		err = h.Login(&c)
	case service.LogoutAction:
		err = h.Logout(&c)
	case service.AddAction:
		err = h.Add(&c)
	case service.RemoveAction:
		err = h.Remove(&c)
	default:
		err = h.RequestNode(&c)
	}
	if err != nil {
		serveError("Could not process request: %v", err)
	}
}

// RequestNode handles node requests.
func (h *nodeHandler) RequestNode(c *reqContext) error {
	// Setup ticket and send to workers.
	h.Log.Print(c.Site.Name, c.Req.Method, c.Req.URL.Path)

	nodeServ, err := h.Info.FindNodeService(c.Node.Type)
	defer func() {
		if err := nodeServ.Close(); err != nil {
			panic(fmt.Errorf("Could not close connection to node service: %v", err))
		}
	}()
	if err != nil {
		return fmt.Errorf("Could not find node service for %q at %q: %v", c.Node.Type, err)
	}
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
}
