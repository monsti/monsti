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
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/monsti/service/info"
	"github.com/monsti/service/login"
	"github.com/monsti/service/node"
	"github.com/monsti/util/l10n"
	"github.com/monsti/util/template"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
)

// nodeHandler is a net/http handler to process incoming HTTP requests.
type nodeHandler struct {
	Renderer template.Renderer
	Settings *settings
	// Hosts is a map from hosts to site names.
	Hosts map[string]string
	// Log is the logger used by the node handler.
	Log *log.Logger
	// Info is a connection to an INFO service.
	Info *info.Service
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

// ServeHTTP handles incoming HTTP requests.
func (h *nodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "panic: %v\n", err)
			buf.Write(debug.Stack())
			h.Log.Println(buf.String())
			http.Error(w, "Application error.",
				http.StatusInternalServerError)
		}
	}()
	nodePath, action := splitAction(r.URL.Path)
	if len(action) == 0 && nodePath[len(nodePath)-1] != '/' {
		newPath, err := url.Parse(nodePath + "/")
		if err != nil {
			panic("Could not parse request URL:" + err.Error())
		}
		url := r.URL.ResolveReference(newPath)
		http.Redirect(w, r, url.String(), http.StatusSeeOther)
		return
	}
	site_name, ok := h.Hosts[r.Host]
	if !ok {
		panic("No site found for host " + r.Host)
	}
	site := h.Settings.Sites[site_name]
	site.Name = site_name
	session := getSession(r, site)
	defer context.Clear(r)
	cSession := getClientSession(session, h.Settings.Monsti.GetSiteConfigPath(
		site.Name))
	cSession.Locale = site.Locale
	node, err := lookupNode(h.Settings.Monsti.GetSiteNodesPath(site.Name),
		nodePath)
	if err != nil {
		h.Log.Println("Node not found.")
		http.Error(w, "Node not found: "+err.Error(), http.StatusNotFound)
		return
	}

	if !checkPermission(action, cSession) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}
	switch action {
	case "login":
		h.Login(w, r, node, session, cSession, site)
	case "logout":
		h.Logout(w, r, node, session)
	case "add":
		h.Add(w, r, node, session, cSession, site)
	case "remove":
		h.Remove(w, r, node, session, cSession, site)
	default:
		h.RequestNode(w, r, node, action, session, cSession, site)
	}
}

// RequestNode handles node requests.
func (h *nodeHandler) RequestNode(w http.ResponseWriter, r *http.Request,
	reqnode node.Node, action string, session *sessions.Session,
	cSession *login.Session, site site) {
	// Setup ticket and send to workers.
	h.Log.Println(site.Name, r.Method, r.URL.Path)

	nodeServ, err := h.Info.FindNodeService(reqnode.Type)
	if err != nil {
		panic(err)
		http.Error(w, "Application error.",
			http.StatusInternalServerError)
		return
	}
	req := node.Request{
		Method:  r.Method,
		Node:    reqnode,
		Query:   r.URL.Query(),
		Session: *cSession,
		Action:  action}
	res, err := nodeServ.Request(&req)
	if err != nil {
		panic(err)
		http.Error(w, "Application error.",
			http.StatusInternalServerError)
		return
	}

	G := l10n.UseCatalog(cSession.Locale)
	if len(res.Body) == 0 && len(res.Redirect) == 0 {
		http.Error(w, "Application error.",
			http.StatusInternalServerError)
		return
	}
	if res.Node != nil {
		oldPath := reqnode.Path
		reqnode = *res.Node
		reqnode.Path = oldPath
	}
	if len(res.Redirect) > 0 {
		http.Redirect(w, r, res.Redirect, http.StatusSeeOther)
		return
	}
	env := masterTmplEnv{Node: reqnode, Session: cSession}
	if action == "edit" {
		env.Title = fmt.Sprintf(G("Edit \"%s\""), reqnode.Title)
		env.Flags = EDIT_VIEW
	}
	var content []byte
	if res.Raw {
		content = res.Body
	} else {
		content = []byte(renderInMaster(h.Renderer, res.Body, env, h.Settings,
			site, cSession.Locale))
	}
	err = session.Save(r, w)
	if err != nil {
		panic(err.Error())
	}
	w.Write(content)
}
