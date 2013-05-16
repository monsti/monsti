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
	"code.google.com/p/go.crypto/bcrypt"
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
	"path/filepath"
)

type loginFormData struct {
	Login, Password string
}

// Login handles login requests.
func (h *nodeHandler) Login(w http.ResponseWriter, r *http.Request,
	reqnode node.Node, session *sessions.Session, cSession *login.Session,
	site site) {
	G := l10n.UseCatalog(cSession.Locale)
	data := loginFormData{}
	form := form.NewForm(&data, form.Fields{
		"Login": form.Field{G("Login"), "", form.Required(G("Required.")),
			nil},
		"Password": form.Field{G("Password"), "", form.Required(G("Required.")),
			new(form.PasswordWidget)}})
	switch r.Method {
	case "GET":
	case "POST":
		r.ParseForm()
		if form.Fill(r.Form) {
			user := getUser(data.Login, site.Directories.Config)
			if user != nil && passwordEqual(user.Password, data.Password) {
				session.Values["login"] = user.Login
				session.Save(r, w)
				http.Redirect(w, r, reqnode.Path, http.StatusSeeOther)
				return
			}
			form.AddError("", G("Wrong login or password."))
		}
	default:
		panic("Request method not supported: " + r.Method)
	}
	data.Password = ""
	body := h.Renderer.Render("daemon/actions/loginform", template.Context{
		"Form": form.RenderData()}, cSession.Locale, site.Directories.Templates)
	env := masterTmplEnv{Node: reqnode, Session: cSession, Title: G("Login"),
		Description: G("Login with your site account."),
		Flags:       EDIT_VIEW}
	fmt.Fprint(w, renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		site, cSession.Locale))
}

// Logout handles logout requests.
func (h *nodeHandler) Logout(w http.ResponseWriter, r *http.Request,
	reqnode node.Node, session *sessions.Session) {
	delete(session.Values, "login")
	session.Save(r, w)
	http.Redirect(w, r, reqnode.Path, http.StatusSeeOther)
}

// getSession returns a currently active or new session.
func getSession(r *http.Request, site site) *sessions.Session {
	if len(site.SessionAuthKey) == 0 {
		panic(`Missing "SessionAuthKey" setting.`)
	}
	store := sessions.NewCookieStore([]byte(site.SessionAuthKey))
	session, _ := store.Get(r, "monsti-session")
	return session
}

// getClientSession returns the client session for the given session.
//
// configDir is the site's configuration directory.
func getClientSession(session *sessions.Session,
	configDir string) (cSession *login.Session) {
	cSession = new(login.Session)
	loginData, ok := session.Values["login"]
	if !ok {
		return
	}
	login_, ok := loginData.(string)
	if !ok {
		delete(session.Values, "login")
		return
	}
	user := getUser(login_, configDir)
	if user == nil {
		delete(session.Values, "login")
		return
	}
	*cSession = login.Session{User: user}
	return
}

// getUser returns the user with the given login.
func getUser(login_, configDir string) *login.User {
	path := filepath.Join(configDir, "users.yaml")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic("Could not load users.yaml: " + err.Error())
	}
	var users []login.User
	if err = goyaml.Unmarshal(content, &users); err != nil {
		panic("Could not unmarshal users.yaml: " + err.Error())
	}
	for _, user := range users {
		if user.Login == login_ {
			return &user
		}
	}
	return nil
}

// checkPermission checks if the session's user might perform the given action.
func checkPermission(action string, session *login.Session) bool {
	auth := session.User != nil
	switch action {
	case "remove", "edit", "add", "logout":
		if auth {
			return true
		}
	case "", "login":
		return true
	}
	return false
}

// passwordEqual returns true iff the hash matches the password.
func passwordEqual(hash, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash),
		[]byte(password)); err != nil {
		return false
	}
	return true
}
