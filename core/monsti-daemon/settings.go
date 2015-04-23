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
	"fmt"
	"net/http"

	"path"
	"github.com/chrneumann/htmlwidgets"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/service"
	mtemplate "pkg.monsti.org/monsti/api/util/template"
)

type settingsFormData struct {
	Fields service.NestedMap
}

func (h *nodeHandler) SettingsAction(c *reqContext) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", c.UserSession.Locale)
	m := c.Serv.Monsti()

	settings, err := m.LoadSiteSettings(c.Site)
	if err != nil {
		return fmt.Errorf("Could not load site settings: %v", err)
	}

	formData := settingsFormData{}
	formData.Fields = make(service.NestedMap)

	form := htmlwidgets.NewForm(&formData)

	for _, field := range settings.FieldConfigs {
		if field.Hidden {
			continue
		}
		settings.Fields[field.Id].ToFormField(form, formData.Fields,
			field, c.UserSession.Locale)
	}

	switch c.Req.Method {
	case "GET":
	case "POST":
		if form.Fill(c.Req.Form) {
			for _, field := range settings.FieldConfigs {
				if !field.Hidden {
					settings.Fields[field.Id].FromFormField(formData.Fields, field)
				}
			}
			if err := m.WriteSiteSettings(c.Site, settings); err != nil {
				return fmt.Errorf("Could not update settings: %v", err)
			}
			/*
				err = m.MarkDep(
					c.Site.Name, service.CacheDep{Settings: path.Clean(settings.Path)})
				if err != nil {
					return fmt.Errorf("Could not mark settings: %v", err)
				}
			*/
			http.Redirect(c.Res, c.Req, path.Join(c.Node.Path, "@@settings?saved=1"),
				http.StatusSeeOther)
			return nil
		}
	default:
		return fmt.Errorf("Request method not supported: %v", c.Req.Method)
	}
	rendered, err := h.Renderer.Render("actions/settings",
		mtemplate.Context{
			"Form":  form.RenderData(),
			"Saved": c.Req.FormValue("saved"),
		}, c.UserSession.Locale, h.Settings.Monsti.GetSiteTemplatesPath(c.Site))

	if err != nil {
		return fmt.Errorf("Could not render settings template: %v", err)
	}

	content, _ := renderInMaster(h.Renderer, []byte(rendered),
		masterTmplEnv{Node: c.Node, Session: c.UserSession,
			Title: G("Settings"), Flags: EDIT_VIEW},
		h.Settings, c.Site, c.UserSession.Locale, c.Serv)

	c.Res.Write(content)
	return nil
}
