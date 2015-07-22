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
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"path/filepath"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util/template"
)

// Context holds information about a request
type reqContext struct {
	Id           uint
	Res          http.ResponseWriter
	Req          *http.Request
	Node         *service.Node
	Action       service.Action
	Session      *sessions.Session
	UserSession  *service.UserSession
	Site         string
	SiteSettings *service.Settings
	Serv         *service.Session
}

// nodeHandler is a net/http handler to process incoming HTTP requests.
type nodeHandler struct {
	Renderer         template.Renderer
	Settings         *settings
	InitializedSites map[string]bool
	// Log is the logger used by the node handler.
	Log *log.Logger
	// Info is a connection to an INFO service.
	Monsti        *service.MonstiClient
	Sessions      *service.SessionPool
	requests      map[uint]*reqContext
	lastRequestID uint
	mutex         sync.RWMutex
}

func (n *nodeHandler) GetRequest(id uint) *service.Request {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	req, ok := n.requests[id]
	if !ok {
		return nil
	}
	return &service.Request{
		Id:       id,
		NodePath: req.Node.Path,
		Site:     req.Site,
		Query:    req.Req.URL.Query(),
		Method:   req.Req.Method,
		Session:  req.UserSession,
		Action:   req.Action,
		Form:     req.Req.Form,
		PostForm: req.Req.PostForm,
		/*
			Node:  req.Node,
		*/
	}
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
	h.mutex.Lock()
	c.Id = h.lastRequestID
	h.lastRequestID += 1
	if h.requests == nil {
		h.requests = make(map[uint]*reqContext)
	}
	h.requests[c.Id] = &c
	h.mutex.Unlock()
	defer func() {
		h.mutex.Lock()
		delete(h.requests, c.Id)
		h.mutex.Unlock()
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
		"view":                   service.ViewAction,
		"edit":                   service.EditAction,
		"list":                   service.ListAction,
		"chooser":                service.ChooserAction,
		"settings":               service.SettingsAction,
		"login":                  service.LoginAction,
		"logout":                 service.LogoutAction,
		"add":                    service.AddAction,
		"remove":                 service.RemoveAction,
		"request-password-token": service.RequestPasswordTokenAction,
		"change-password":        service.ChangePasswordAction,
	}[action]
	c.Site = strings.SplitN(c.Req.Host, ":", 2)[0]
	if v, ok := h.InitializedSites[c.Site]; !(ok && v) {
		ok, err := c.Serv.Monsti().InitSite(c.Site)
		if err != nil {
			serveError("Error initializing site for host %v: %v", c.Site, err)
		}
		if !ok {
			serveError("No site found for host %v", c.Site)
		}
		http.Handle(c.Req.Host+"/site-static/", http.FileServer(http.Dir(
			filepath.Dir(h.Settings.Monsti.GetSiteStaticsPath(c.Site)))))
		h.InitializedSites[c.Site] = true
	}
	c.SiteSettings, err = c.Serv.Monsti().LoadSiteSettings(c.Site)
	if err != nil {
		serveError("Could not load site settings: %v", err)
	}
	c.Session, err = getSession(c.Req, c.SiteSettings.StringValue("core.SessionAuthKey"))
	if err != nil {
		serveError("Could not get session: %v", err)
	}
	defer context.Clear(c.Req)
	c.UserSession, err = getClientSession(c.Session,
		h.Settings.Monsti.GetSiteDataPath(c.Site))
	if err != nil {
		serveError("Could not get client session: %v", err)
	}
	c.UserSession.Locale = c.SiteSettings.Fields["core.Locale"].Value().(string)

	h.Log.Printf("(%v) %v %v", c.Site, c.Req.Method, c.Req.URL.Path)

	if err := c.Req.ParseForm(); err != nil {
		serveError("Could not parse form: %v", err)
	}

	// Try to serve page from cache
	if c.UserSession.User == nil && c.Action == service.ViewAction &&
		nodePath[len(nodePath)-1] == '/' &&
		len(c.Req.Form) == 0 {
		content, _, err := c.Serv.Monsti().FromCache(c.Site, nodePath,
			"core.page.full")
		if err == nil && content != nil {
			c.Res.Write(content)
			return
		}
	}

	c.Node, err = c.Serv.Monsti().GetNode(c.Site, nodePath)
	if err != nil {
		serveError("Error getting node %v of site %v: %v",
			nodePath, c.Site, err)
	}
	if c.Node == nil ||
		(c.Action == service.ViewAction && c.UserSession.User == nil &&
			(c.Node.Public == false || c.Node.PublishTime.After(time.Now()))) {
		h.Log.Printf("Node not found: %v @ %v", nodePath, c.Site)
		c.Node = &service.Node{Path: nodePath}
		http.Error(c.Res, "Document not found", http.StatusNotFound)
		return
	}
	if c.Node.Type.Id == "core.Path" {
		switch c.Action {
		case service.ViewAction, service.EditAction,
			service.AddAction, service.RemoveAction:
			http.Redirect(c.Res, c.Req, filepath.Join(c.Node.Path, "@@list"),
				http.StatusSeeOther)
			return
		}
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
	case service.ListAction:
		err = h.List(&c)
	case service.ChooserAction:
		err = h.Chooser(&c)
	case service.SettingsAction:
		err = h.SettingsAction(&c)
	case service.RequestPasswordTokenAction:
		err = h.RequestPasswordToken(&c)
	case service.ChangePasswordAction:
		err = h.ChangePassword(&c)
	default:
		err = h.View(&c)
	}
	if err != nil {
		serveError("Could not process request: %v", err)
	}
}
