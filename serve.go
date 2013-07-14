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
	"github.com/monsti/service"
	"github.com/monsti/util"
	"github.com/monsti/util/template"
	"gitorious.org/monsti/gettext"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
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
	Info *service.InfoClient
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
	site := h.Settings.Monsti.Sites[site_name]
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
	reqnode service.NodeInfo, action string, session *sessions.Session,
	cSession *service.UserSession, site util.SiteSettings) {
	// Setup ticket and send to workers.
	h.Log.Println(site.Name, r.Method, r.URL.Path)

	nodeServ, err := h.Info.FindNodeService(reqnode.Type)
	if err != nil {
		panic(fmt.Sprintf("Could not find node service: %v", err))
	}
	if err = r.ParseMultipartForm(1024 * 1024); err != nil {
		panic(fmt.Sprintf("Could not parse form: %v", err))
	}
	req := service.Request{
		Site:     site.Name,
		Method:   r.Method,
		Node:     reqnode,
		Query:    r.URL.Query(),
		Session:  *cSession,
		Action:   action,
		FormData: r.Form,
	}

	// Attach request files
	if r.MultipartForm != nil {
		if len(r.MultipartForm.File) > 0 {
			req.Files = make(map[string][]service.RequestFile)
		}
		for name, fileHeaders := range r.MultipartForm.File {
			h.Log.Println(name)
			if _, ok := req.Files[name]; !ok {
				req.Files[name] = make([]service.RequestFile, 0)
			}
			for _, fileHeader := range fileHeaders {
				file, err := fileHeader.Open()
				if err != nil {
					panic("Could not open multipart file header: " + err.Error())
				}
				if osFile, ok := file.(*os.File); ok {
					req.Files[name] = append(req.Files[name], service.RequestFile{
						TmpFile: osFile.Name()})
				} else {
					content, err := ioutil.ReadAll(file)
					if err != nil {
						panic("Could not read multipart file: " + err.Error())
					}
					req.Files[name] = append(req.Files[name], service.RequestFile{
						Content: content})
				}
			}
		}
	}

	res, err := nodeServ.Request(&req)
	if err != nil {
		panic(fmt.Sprintf("Could not request node: %v", err))
	}

	G, _, _, _ := gettext.DefaultLocales.Use("monsti-httpd", cSession.Locale)
	if len(res.Body) == 0 && len(res.Redirect) == 0 {
		panic("Got empty response.")
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
