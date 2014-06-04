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
	"fmt"
	"net/http"

	"path"
	"github.com/chrneumann/mimemail"
	"pkg.monsti.org/form"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/util/template"
)

type contactFormData struct {
	Name, Email, Subject, Message string
}

func renderContactForm(c *reqContext, context template.Context,
	h *nodeHandler) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.Site.Locale)
	data := contactFormData{}
	form := form.NewForm(&data, form.Fields{
		"Name":    form.Field{G("Name"), "", form.Required(G("Required.")), nil},
		"Email":   form.Field{G("Email"), "", form.Required(G("Required.")), nil},
		"Subject": form.Field{G("Subject"), "", form.Required(G("Required.")), nil},
		"Message": form.Field{G("Message"), "", form.Required(G("Required.")),
			new(form.TextArea)}})
	switch c.Req.Method {
	case "GET":
		if _, submitted := c.Req.Form["submitted"]; submitted {
			context["Submitted"] = 1
		}
	case "POST":
		if form.Fill(c.Req.Form) {
			mail := mimemail.Mail{
				From:    mimemail.Address{data.Name, data.Email},
				Subject: data.Subject,
				Body:    []byte(data.Message)}
			site := h.Settings.Monsti.Sites[c.Site.Name]
			owner := mimemail.Address{site.Owner.Name, site.Owner.Email}
			mail.To = []mimemail.Address{owner}
			err := c.Serv.Mail().SendMail(&mail)
			if err != nil {
				return fmt.Errorf("Could not send mail: %v", err)
			}
			http.Redirect(c.Res, c.Req, path.Dir(c.Node.Path)+"/?submitted", http.StatusSeeOther)
			return nil
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	context["Form"] = form.RenderData()
	return nil
}
