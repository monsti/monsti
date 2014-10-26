// This file is part of Monsti, a web content management system.
// Copyright 2012-2014 Christian Neumann
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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/chrneumann/htmlwidgets"
	"github.com/chrneumann/mimemail"
	"github.com/gorilla/sessions"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util"
	"pkg.monsti.org/monsti/api/util/template"
)

type loginFormData struct {
	Login, Password string
}

// Login handles login requests.
func (h *nodeHandler) Login(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)
	data := loginFormData{}

	form := htmlwidgets.NewForm(&data)
	form.AddWidget(new(htmlwidgets.TextWidget), "Login", G("Login"), "")
	form.AddWidget(new(htmlwidgets.PasswordWidget), "Password", G("Password"), "")

	switch c.Req.Method {
	case "GET":
	case "POST":
		c.Req.ParseForm()
		if form.Fill(c.Req.Form) {
			user, err := getUser(data.Login,
				h.Settings.Monsti.GetSiteDataPath(c.Site.Name))
			if err != nil {
				return fmt.Errorf("Could not get user: %v", err)
			}
			if user != nil && passwordEqual(user.Password, data.Password) {
				c.Session.Values["login"] = user.Login
				c.Session.Save(c.Req, c.Res)
				http.Redirect(c.Res, c.Req, c.Node.Path, http.StatusSeeOther)
				return nil
			}
			form.AddError("", G("Wrong login or password."))
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	data.Password = ""
	body, err := h.Renderer.Render("actions/loginform", template.Context{
		"Form": form.RenderData()}, c.UserSession.Locale,
		h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		return fmt.Errorf("Can't render login form: %v", err)
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession, Title: G("Login"),
		Description: G("Login with your site account."),
		Flags:       EDIT_VIEW}
	fmt.Fprint(c.Res, renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv))
	return nil
}

// Logout handles logout requests.
func (h *nodeHandler) Logout(c *reqContext) error {
	delete(c.Session.Values, "login")
	c.Session.Save(c.Req, c.Res)
	http.Redirect(c.Res, c.Req, c.Node.Path, http.StatusSeeOther)
	return nil
}

func (h *nodeHandler) ChangePassword(c *reqContext) error {
	return nil
}

type requestPasswordTokenFormData struct {
	User string
}

// RequestPasswordToken sends the user a token to be able to change
// the login password.
func (h *nodeHandler) RequestPasswordToken(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)
	data := requestPasswordTokenFormData{}
	form := htmlwidgets.NewForm(&data)
	form.AddWidget(new(htmlwidgets.TextWidget), "User", G("Login"), "")

	sent := false
	c.Req.ParseForm()
	switch c.Req.Method {
	case "GET":
		if _, ok := c.Req.Form["sent"]; ok {
			sent = true
		}
	case "POST":
		if form.Fill(c.Req.Form) {
			user, err := getUser(data.User,
				h.Settings.Monsti.GetSiteDataPath(c.Site.Name))
			if err != nil {
				return fmt.Errorf("Could not get user: %v", err)
			}
			if user != nil {
				site := h.Settings.Monsti.Sites[c.Site.Name]
				link := getRequestPasswordToken(c.Site.Name, data.User, site.PasswordTokenKey)
				mail := mimemail.Mail{
					From:    mimemail.Address{site.EmailName, site.EmailAddress},
					Subject: G("Password request"),
					Body: []byte(fmt.Sprintf(`Hello,

someone, possibly you, requested a new password for your account %v at
"%v".

To change your password, visit the following link within 24 hours.
If you did not request a new password, you may ignore this email.
%v

This is an automatically generated email. Please don't reply to it.
`, data.User, site.Title, site.BaseURL+"/@@change-password?token="+link))}
				mail.To = []mimemail.Address{mimemail.Address{user.Login, user.Email}}
				err := c.Serv.Monsti().SendMail(&mail)
				if err != nil {
					return fmt.Errorf("Could not send mail: %v", err)
				}
				http.Redirect(c.Res, c.Req, "@@request-password-token?sent", http.StatusSeeOther)
				return nil
			} else {
				form.AddError("User", G("Unkown user."))
			}
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}

	body, err := h.Renderer.Render("actions/request_password_token_form",
		template.Context{
			"Sent": sent,
			"Form": form.RenderData()}, c.UserSession.Locale,
		h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		return fmt.Errorf("Can't render login form: %v", err)
	}
	env := masterTmplEnv{
		Node:    c.Node,
		Session: c.UserSession,
		Title:   G("Request new password"),
		Flags:   EDIT_VIEW}
	fmt.Fprint(c.Res, renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv))
	return nil
}

/*
type forgotPasswordFormData struct {
	Token, Login, Password string
}

// ForgotPassword handles new password requests.
func (h *nodeHandler) ForgotPassword(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)
	data := forgotPasswordFormData{}

	hasToken := len(c.Req.FormValue("Token")) > 0

	form := htmlwidgets.NewForm(&data)
	form.AddWidget(new(htmlwidgets.HiddenWidget), "Token", "", "")
	if hasToken {
		form.AddWidget(&htmlwidgets.PasswordWidget{Confirm: true}), "Password",
	} else {
		form.AddWidget(new(htmlwidgets.TextWidget), "Login", G("Login"), "")
	}
	G("New password"), "")

	switch c.Req.Method {
	case "GET":
	case "POST":
		c.Req.ParseForm()
		if form.Fill(c.Req.Form) {
			user, err := getUser(data.Login,
				h.Settings.Monsti.GetSiteDataPath(c.Site.Name))
			if err != nil {
				return fmt.Errorf("Could not get user: %v", err)
			}
			if user != nil {
			} else {
				form.AddError("Login", G("Unkown user."))
			}
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	data.Password = ""
	body, err := h.Renderer.Render("actions/loginform", template.Context{
		"Form": form.RenderData()}, c.UserSession.Locale,
		h.Settings.Monsti.GetSiteTemplatesPath(c.Site.Name))
	if err != nil {
		return fmt.Errorf("Can't render login form: %v", err)
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession, Title: G("Login"),
		Description: G("Login with your site account."),
		Flags:       EDIT_VIEW}
	fmt.Fprint(c.Res, renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		*c.Site, c.UserSession.Locale, c.Serv))
	return nil
}
*/

// getSession returns a currently active or new session.
func getSession(r *http.Request, site util.SiteSettings) (
	*sessions.Session, error) {
	if len(site.SessionAuthKey) == 0 {
		return nil, fmt.Errorf(`Missing "SessionAuthKey" setting.`)
	}
	store := sessions.NewCookieStore([]byte(site.SessionAuthKey))
	session, _ := store.Get(r, "monsti-session")
	return session, nil
}

// getClientSession returns the client session for the given session.
//
// configDir is the site's configuration directory.
func getClientSession(session *sessions.Session,
	dataDir string) (uSession *service.UserSession, err error) {
	uSession = new(service.UserSession)
	loginData, ok := session.Values["login"]
	if !ok {
		return
	}
	login_, ok := loginData.(string)
	if !ok {
		delete(session.Values, "login")
		return
	}
	user, err := getUser(login_, dataDir)
	if err != nil {
		err = fmt.Errorf("Could not get user: %v", err)
		return
	}
	if user == nil {
		delete(session.Values, "login")
		return
	}
	*uSession = service.UserSession{User: user}
	return
}

// getUser returns the user with the given login.
func getUser(login_, dataDir string) (*service.User, error) {
	path := filepath.Join(dataDir, "users.json")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not load user database: %v", err)
	}
	var users map[string]service.User
	if err = json.Unmarshal(content, &users); err != nil {
		return nil, fmt.Errorf("Could not unmarshal user database: %v", err)
	}
	if user, ok := users[login_]; ok {
		user.Login = login_
		return &user, nil
	}
	return nil, nil
}

// checkPermission checks if the session's user might perform the given action.
func checkPermission(action service.Action, session *service.UserSession) bool {
	auth := session.User != nil
	switch action {
	case service.RemoveAction, service.EditAction, service.AddAction,
		service.LogoutAction:
		if auth {
			return true
		}
	default:
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

func generateToken(args ...string) string {
	hash := sha256.Sum256([]byte(strings.Join(args, "#")))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// genPasswordToken generates a password token
func getRequestPasswordToken(site, login, secret string) string {
	if len(secret) == 0 {
		panic("Secret passed to getRequestPasswordToken must not be empty")
	}
	generated := time.Now().Unix()
	return fmt.Sprintf("%v-%v-%v", login, generated, generateToken(
		site, login, fmt.Sprint(generated), secret))
}

// verifyRequestPasswordToken verifies the password token for the
// given site and user.
func verifyRequestPasswordToken(site string, user service.User, secret string,
	token string) bool {
	if len(secret) == 0 {
		panic("Secret passed to verifyRequestPasswordToken must not be empty")
	}
	parts := strings.Split(token, "-")
	if len(parts) < 3 {
		return false
	}
	userPartsCount := len(parts) - 2
	userSubstring := strings.Join(parts[:userPartsCount], "-")
	if userSubstring != user.Login {
		return false
	}
	timeSubstring := parts[userPartsCount]
	generated, err := strconv.Atoi(timeSubstring)
	if err != nil || int64(generated) < user.PasswordChanged.Unix() {
		return false
	}
	hashSubstring := parts[userPartsCount+1]
	calculated := generateToken(site, user.Login, timeSubstring, secret)
	return calculated == hashSubstring
}
