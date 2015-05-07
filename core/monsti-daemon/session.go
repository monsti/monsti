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
	"encoding/base32"
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
	"github.com/gorilla/sessions"
	gomail "gopkg.in/gomail.v1"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/service"
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
		if form.Fill(c.Req.Form) {
			user, err := getUser(data.Login,
				h.Settings.Monsti.GetSiteDataPath(c.Site))
			if err != nil {
				return fmt.Errorf("Could not get user: %v", err)
			}
			if user != nil && passwordEqual(user.Password, data.Password) {
				c.Session.Values["login"] = user.Login
				c.Session.Save(c.Req, c.Res)
				http.Redirect(c.Res, c.Req, c.Node.Path+"/", http.StatusSeeOther)
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
		h.Settings.Monsti.GetSiteTemplatesPath(c.Site))
	if err != nil {
		return fmt.Errorf("Can't render login form: %v", err)
	}
	env := masterTmplEnv{Node: c.Node, Session: c.UserSession, Title: G("Login"),
		Description: G("Login with your site account."),
		Flags:       EDIT_VIEW}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		c.Site, c.SiteSettings, c.UserSession.Locale, c.Serv)
	c.Res.Write(rendered)
	return nil
}

// Logout handles logout requests.
func (h *nodeHandler) Logout(c *reqContext) error {
	delete(c.Session.Values, "login")
	c.Session.Save(c.Req, c.Res)
	http.Redirect(c.Res, c.Req, c.Node.Path, http.StatusSeeOther)
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
	switch c.Req.Method {
	case "GET":
		if _, ok := c.Req.Form["sent"]; ok {
			sent = true
		}
	case "POST":
		if form.Fill(c.Req.Form) {
			user, err := getUser(data.User,
				h.Settings.Monsti.GetSiteDataPath(c.Site))
			if err != nil {
				return fmt.Errorf("Could not get user: %v", err)
			}
			if user != nil {
				link := getRequestPasswordToken(c.Site, data.User,
					c.SiteSettings.StringValue("core.PasswordTokenKey"))

				// Send email to user
				mail := gomail.NewMessage()
				mail.SetAddressHeader("From",
					c.SiteSettings.StringValue("core.EmailAddress"),
					c.SiteSettings.StringValue("core.EmailName"))
				mail.SetAddressHeader("To", user.Email, user.Login)
				mail.SetHeader("Subject", G("Password request"))

				body, err := h.Renderer.Render("mails/change_password",
					template.Context{
						"SiteSettings": c.SiteSettings,
						"Account":      data.User,
						"ChangeLink": c.SiteSettings.StringValue("core.BaseURL") +
							"/@@change-password?token=" + link,
					}, c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site))
				if err != nil {
					return fmt.Errorf("Can't render password change mail: %v", err)
				}
				mail.SetBody("text/plain", string(body))
				mailer := gomail.NewCustomMailer("", nil, gomail.SetSendMail(
					c.Serv.Monsti().SendMailFunc()))
				err = mailer.Send(mail)
				if err != nil {
					return fmt.Errorf("Could not send mail: %v", err)
				}

				http.Redirect(c.Res, c.Req, "@@request-password-token?sent",
					http.StatusSeeOther)
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
		h.Settings.Monsti.GetSiteTemplatesPath(c.Site))
	if err != nil {
		return fmt.Errorf("Can't render login form: %v", err)
	}
	env := masterTmplEnv{
		Node:    c.Node,
		Session: c.UserSession,
		Title:   G("Request new password"),
		Flags:   EDIT_VIEW}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		c.Site, c.SiteSettings, c.UserSession.Locale, c.Serv)
	c.Res.Write(rendered)
	return nil
}

type changePasswordFormData struct {
	OldPassword, Password string
}

// ChangePassword allows to change the user's password.
func (h *nodeHandler) ChangePassword(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)
	authenticated := c.UserSession.User != nil
	data := changePasswordFormData{}
	form := htmlwidgets.NewForm(&data)
	if authenticated {
		form.AddWidget(&htmlwidgets.PasswordWidget{}, "OldPassword",
			G("Old Password"), "")
	}
	form.AddWidget(&htmlwidgets.PasswordWidget{
		VerifyLabel: G("Please repeat the password."),
		VerifyError: G("Passwords do not match."),
	}, "Password", G("New Password"), "")
	var token string
	tokenInvalid := false
	var user *service.User
	if !authenticated {
		if tokens, ok := c.Req.Form["token"]; ok {
			token = tokens[0]
		}
		if len(token) == 0 {
			http.Redirect(c.Res, c.Req, "@@login", http.StatusSeeOther)
			return nil
		}
		getUserFn := func(login string) (*service.User, error) {
			return getUser(login, h.Settings.Monsti.GetSiteDataPath(c.Site))
		}
		var err error
		user, err = verifyRequestPasswordToken(
			c.Site, getUserFn, c.SiteSettings.StringValue("core.PasswordTokenKey"),
			token)
		if err != nil {
			return fmt.Errorf("Could not verify request password token: %v", err)
		}
		if user == nil {
			tokenInvalid = true
		}
	} else {
		user = c.UserSession.User
	}
	changed := false
	switch c.Req.Method {
	case "GET":
		if _, ok := c.Req.Form["changed"]; ok {
			changed = true
		}
	case "POST":
		if authenticated || !tokenInvalid {
			if form.Fill(c.Req.Form) {
				changePassword := true
				if authenticated {
					changePassword = passwordEqual(user.Password,
						data.OldPassword)
					if !changePassword {
						form.AddError("Password", G("Wrong password."))
					}
				}
				if changePassword {
					hashed, err := bcrypt.GenerateFromPassword([]byte(data.Password), 0)
					if err != nil {
						return fmt.Errorf("Could not hash user password: %v", err)
					}
					user.PasswordChanged = time.Now().UTC()
					user.Password = string(hashed)
					err = writeUser(user, h.Settings.Monsti.GetSiteDataPath(c.Site))
					if err != nil {
						return fmt.Errorf("Could not change user password: %v", err)
					}
					http.Redirect(c.Res, c.Req, "@@change-password?changed",
						http.StatusSeeOther)
					return nil
				}
			}
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}

	body, err := h.Renderer.Render("actions/change_password",
		template.Context{
			"TokenInvalid": tokenInvalid,
			"Changed":      changed,
			"Form":         form.RenderData()}, c.UserSession.Locale,
		h.Settings.Monsti.GetSiteTemplatesPath(c.Site))
	if err != nil {
		return fmt.Errorf("Can't render ChangePassword form: %v", err)
	}
	env := masterTmplEnv{
		Node:    c.Node,
		Session: c.UserSession,
		Title:   G("Change password"),
		Flags:   EDIT_VIEW}
	rendered, _ := renderInMaster(h.Renderer, []byte(body), env, h.Settings,
		c.Site, c.SiteSettings, c.UserSession.Locale, c.Serv)
	c.Res.Write(rendered)
	return nil
}

// getSession returns a currently active or new session.
func getSession(r *http.Request, key string) (
	*sessions.Session, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("Missing session auth key")
	}
	store := sessions.NewCookieStore([]byte(key))
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

// getUserDatabase reads the user database from the given site data directory.
func getUserDatabase(dataDir string) (map[string]service.User, error) {
	path := filepath.Join(dataDir, "users.json")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not load user database: %v", err)
	}
	users := make(map[string]service.User)
	if err = json.Unmarshal(content, &users); err != nil {
		return nil, fmt.Errorf("Could not unmarshal user database: %v", err)
	}
	return users, nil
}

// writeUserDatabase writes the given user database to the given site
// data directory.
func writeUserDatabase(users map[string]service.User, dataDir string) error {
	path := filepath.Join(dataDir, "users.json")
	content, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("Could not marshal user database: %v", err)
	}
	if err = ioutil.WriteFile(path, content, 0660); err != nil {
		return fmt.Errorf("Could not write user database: %v", err)
	}
	return nil
}

// getUser returns the user with the given login. If there is no such
// user, returns nil.
func getUser(login_, dataDir string) (*service.User, error) {
	users, err := getUserDatabase(dataDir)
	if err != nil {
		return nil, fmt.Errorf("Could not get user database: %v", err)
	}
	if user, ok := users[login_]; ok {
		user.Login = login_
		return &user, nil
	}
	return nil, nil
}

// writeUser saves the given user in the user database.
//
// An existing entry for the given user login will be overwritten.
func writeUser(user *service.User, dataDir string) error {
	users, err := getUserDatabase(dataDir)
	if err != nil {
		return fmt.Errorf("Could not get user database: %v", err)
	}
	users[user.Login] = *user
	if err = writeUserDatabase(users, dataDir); err != nil {
		return fmt.Errorf("Could not write user database: %v", err)
	}
	return nil
}

// checkPermission checks if the session's user might perform the given action.
func checkPermission(action service.Action, session *service.UserSession) bool {
	auth := session.User != nil
	switch action {
	case service.RemoveAction, service.EditAction, service.AddAction,
		service.LogoutAction, service.ListAction, service.ChooserAction,
		service.SettingsAction:
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
	return base32.StdEncoding.EncodeToString(hash[:])
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
// given site and returns the user who requested the password
// change. If the token is invalid, returns nil.
func verifyRequestPasswordToken(site string,
	getUserFn func(login string) (*service.User, error),
	secret string, token string) (*service.User, error) {
	if len(secret) == 0 {
		panic("Secret passed to verifyRequestPasswordToken must not be empty")
	}
	parts := strings.Split(token, "-")
	if len(parts) < 3 {
		return nil, nil
	}
	userPartsCount := len(parts) - 2
	userSubstring := strings.Join(parts[:userPartsCount], "-")
	user, err := getUserFn(userSubstring)
	if err != nil {
		return nil, fmt.Errorf("Could not get user: %v", err)
	}
	timeSubstring := parts[userPartsCount]
	generated, err := strconv.Atoi(timeSubstring)
	if err != nil || int64(generated) < user.PasswordChanged.Unix() {
		return nil, nil
	}
	hashSubstring := parts[userPartsCount+1]
	calculated := generateToken(site, user.Login, timeSubstring, secret)
	if calculated == hashSubstring {
		return user, nil
	}
	return nil, nil
}
