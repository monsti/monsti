package main

import (
	"code.google.com/p/gorilla/sessions"
	"datenkarussell.de/monsti/form"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"fmt"
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
	node client.Node, session *sessions.Session, cSession *client.Session) {
	data := loginFormData{}
	form := form.NewForm(&data, form.Fields{
		"Login":    form.Field{G("Login"), "", form.Required(), nil},
		"Password": form.Field{G("Password"), "", form.Required(), nil}})
	switch r.Method {
	case "GET":
	case "POST":
		r.ParseForm()
		if form.Fill(r.Form) {
			user := getUser(data.Login, h.Settings.Directories.Config)
			if user != nil && user.Password == data.Password {
				session.Values["login"] = user.Login
				session.Save(r, w)
				http.Redirect(w, r, node.Path, http.StatusSeeOther)
				return
			}
			form.AddError("", G("Wrong login or password."))
		}
	default:
		panic("Request method not supported: " + r.Method)
	}
	data.Password = ""
	body := h.Renderer.Render("actions/loginform", template.Context{
		"Form": form.RenderData()})
	env := masterTmplEnv{Node: node, Session: cSession, Title: G("Login"),
		Description: G("Login with your site account."),
		Flags:       EDIT_VIEW}
	fmt.Fprint(w, renderInMaster(h.Renderer, []byte(body), env, h.Settings))
}

// Logout handles logout requests.
func (h *nodeHandler) Logout(w http.ResponseWriter, r *http.Request,
	node client.Node, session *sessions.Session) {
	delete(session.Values, "login")
	session.Save(r, w)
	http.Redirect(w, r, node.Path, http.StatusSeeOther)
}

// getSession returns a currently active or new session.
func getSession(r *http.Request, settings settings) *sessions.Session {
	if len(settings.SessionAuthKey) == 0 {
		panic(`Missing "SessionAuthKey" setting.`)
	}
	store := sessions.NewCookieStore([]byte(settings.SessionAuthKey))
	session, _ := store.Get(r, "monsti-session")
	return session
}

// getClientSession returns the client session for the given session.
func getClientSession(session *sessions.Session,
	configDir string) (cSession *client.Session) {
	cSession = new(client.Session)
	loginData, ok := session.Values["login"]
	if !ok {
		return
	}
	login, ok := loginData.(string)
	if !ok {
		delete(session.Values, "login")
		return
	}
	user := getUser(login, configDir)
	if user != nil {
		delete(session.Values, "login")
		return
	}
	*cSession = client.Session{User: user}
	return
}

// getUser returns the user with the given login.
func getUser(login, configDir string) *client.User {
	path := filepath.Join(configDir, "users.yaml")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic("Could not load users.yaml: " + err.Error())
	}
	var users []client.User
	if err = goyaml.Unmarshal(content, &users); err != nil {
		panic("Could not unmarshal users.yaml: " + err.Error())
	}
	for _, user := range users {
		if user.Login == login {
			return &user
		}
	}
	return nil
}

// checkPermission checks if the session's user might perform the given action.
func checkPermission(action string, session *client.Session) bool {
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
