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
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util"
	"pkg.monsti.org/monsti/api/util/template"
)

// Context holds information about a request
type reqContext struct {
	Res         http.ResponseWriter
	Req         *http.Request
	Node        *service.Node
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
		h.Settings.Monsti.GetSiteDataPath(c.Site.Name))
	if err != nil {
		serveError("Could not get client session: %v", err)
	}
	c.UserSession.Locale = c.Site.Locale
	c.Node, err = c.Serv.Data().GetNode(c.Site.Name, nodePath)
	if err != nil {
		serveError("Error getting node: %v", err)
	}
	if c.Node == nil {
		h.Log.Printf("Node not found: %v @ %v", nodePath, c.Site.Name)
		c.Node = &service.Node{Path: nodePath}
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
	case service.EditAction:
		err = h.Edit(&c)
	default:
		err = h.View(&c)
	}
	if err != nil {
		serveError("Could not process request: %v", err)
	}
}
