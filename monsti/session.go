package main

import (
	"code.google.com/p/gorilla/sessions"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"fmt"
	"net/http"
)

type loginFormData struct {
	Login, Password string
}

func (data *loginFormData) Check(e *template.FormErrors) {
	e.Check("Login", data.Login, template.Required())
	e.Check("Password", data.Password, template.Required())
}

// Login handles login requests.
func (h *nodeHandler) Login(w http.ResponseWriter, r *http.Request,
	nodePath string, session *sessions.Session) {
	var data loginFormData
	var errors template.FormErrors
	switch r.Method {
	case "GET":
	case "POST":
		r.ParseForm()
		var err error
		errors, err = template.Validate(r.Form, &data)
		if err != nil {
			panic("Could not parse form data: " + err.Error())
		}
		if len(errors) == 0 {
			user, err := getUser(data.Login, h.Settings.Directories.Config)
			if err == nil && user.Password == data.Password {
				session.Values["login"] = user.Login
				session.Save(r, w)
				http.Redirect(w, r, nodePath, http.StatusSeeOther)
				return
			}
			errors = make(template.FormErrors)
			errors[":error"] = "Wrong login or password."
		}
	default:
		panic("Request method not supported: " + r.Method)
	}
	data.Password = ""
	body := h.Renderer.Render("actions/loginform.html", errors, data)
	fmt.Fprint(w, renderInMaster(h.Renderer, []byte(body), new(masterTmplEnv),
		h.Settings))
}

// Logout handles logout requests.
func (h *nodeHandler) Logout(w http.ResponseWriter, r *http.Request,
	nodePath string, session *sessions.Session) {
	delete(session.Values, "login")
	session.Save(r, w)
	http.Redirect(w, r, nodePath, http.StatusSeeOther)
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
	user, err := getUser(login, configDir)
	if err != nil {
		delete(session.Values, "login")
		return
	}
	*cSession = client.Session{User: user}
	return
}

// getUser returns the user with the given login.
func getUser(login, configDir string) (*client.User, error) {
	return &client.User{
		Login:    login,
		Name:     "Administrator",
		Password: "foofoo"}, nil
}

// checkPermission checks if the session's user might perform the given action.
func checkPermission(action string, session *client.Session) bool {
	auth := session.User != nil
	switch action {
	case "edit", "add":
		if auth {
			return true
		}
	case "":
		return true
	}
	return false
}
