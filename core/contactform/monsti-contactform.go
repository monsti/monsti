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

/*
 Monsti is a simple and resource efficient CMS.

 This package implements the contactform node type.
*/
package main

import (
	"flag"
	"fmt"
	htmlT "html/template"
	"log"
	"os"

	"github.com/chrneumann/mimemail"
	"pkg.monsti.org/form"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util"
	"pkg.monsti.org/monsti/api/util/template"
)

var settings struct {
	Monsti util.MonstiSettings
}

var logger *log.Logger
var renderer template.Renderer

type contactFormData struct {
	Name, Email, Subject, Message string
}

func view(req service.Request, res *service.Response, s *service.Session) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", req.Session.Locale)
	data := contactFormData{}
	form := form.NewForm(&data, form.Fields{
		"Name":    form.Field{G("Name"), "", form.Required(G("Required.")), nil},
		"Email":   form.Field{G("Email"), "", form.Required(G("Required.")), nil},
		"Subject": form.Field{G("Subject"), "", form.Required(G("Required.")), nil},
		"Message": form.Field{G("Message"), "", form.Required(G("Required.")),
			new(form.TextArea)}})
	context := template.Context{}
	switch req.Method {
	case service.GetRequest:
		if _, submitted := req.Query["submitted"]; submitted {
			context["Submitted"] = 1
		}
	case service.PostRequest:
		if form.Fill(req.FormData) {
			mail := mimemail.Mail{
				From:    mimemail.Address{data.Name, data.Email},
				Subject: data.Subject,
				Body:    []byte(data.Message)}
			site := settings.Monsti.Sites[req.Site]
			owner := mimemail.Address{site.Owner.Name, site.Owner.Email}
			mail.To = []mimemail.Address{owner}
			err := s.Mail().SendMail(&mail)
			if err != nil {
				return fmt.Errorf("Could not send mail: %v", err)
			}
			res.Redirect = req.Node.Path + "/?submitted"
			return nil
		}
	default:
		return fmt.Errorf("Request method not supported: %v", req.Method)
	}
	res.Node = &req.Node
	body, err := s.Data().GetNodeData(req.Site, req.Node.Path, "body.html")
	if err != nil {
		return fmt.Errorf("Could not get node data: %v", err)
	}
	context["Body"] = htmlT.HTML(string(body))
	context["Form"] = form.RenderData()
	rendered, err := renderer.Render("contactform/view", context,
		req.Session.Locale, "")
	if err != nil {
		return fmt.Errorf("Could not render template: %v", err)
	}
	fmt.Fprint(res, rendered)
	return nil
}

type editFormData struct {
	Title, Body string
}

func edit(req service.Request, res *service.Response, s *service.Session) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", req.Session.Locale)
	data := editFormData{}
	form := form.NewForm(&data, form.Fields{
		"Title": form.Field{G("Title"), "", form.Required(G("Required.")), nil},
		"Body": form.Field{G("Body"), "", form.Required(G("Required.")),
			new(form.AlohaEditor)}})
	dataCli := s.Data()
	switch req.Method {
	case service.GetRequest:
		data.Title = req.Node.Title
		nodeData, err := dataCli.GetNodeData(req.Site, req.Node.Path, "body.html")
		if err != nil {
			return fmt.Errorf("Could not get node data: %v", err)
		}
		data.Body = string(nodeData)
	case service.PostRequest:
		if form.Fill(req.FormData) {
			var node struct{ service.NodeFields }
			node.NodeFields = req.Node
			node.Title = data.Title
			err := dataCli.WriteNode(req.Site, node.Path, node, "node")
			if err != nil {
				return fmt.Errorf("Could not update node: %v", err)
			}
			if err := dataCli.WriteNodeData(req.Site, req.Node.Path, "body.html",
				[]byte(data.Body)); err != nil {
				return fmt.Errorf("Could not update node data: %v", err)
			}
			res.Redirect = req.Node.Path
			return nil
		}
	default:
		return fmt.Errorf("Request method not supported: %v", req.Method)
	}
	rendered, err := renderer.Render("contactform/edit",
		template.Context{"Form": form.RenderData()},
		req.Session.Locale, "")
	if err != nil {
		return fmt.Errorf("Could not render template: %v", err)
	}
	fmt.Fprint(res, rendered)
	return nil
}

func main() {
	logger = log.New(os.Stderr, "contactform ", log.LstdFlags)
	// Load configuration
	flag.Parse()
	if flag.NArg() != 1 {
		logger.Fatal("Expecting configuration path.")
	}
	cfgPath := util.GetConfigPath(flag.Arg(0))
	if err := util.LoadModuleSettings("contactform", cfgPath, &settings); err != nil {
		logger.Fatal("Could not load settings: ", err)
	}
	if err := settings.Monsti.LoadSiteSettings(); err != nil {
		logger.Fatal("Could not load site settings: ", err)
	}

	infoPath := settings.Monsti.GetServicePath(service.Info.String())

	gettext.DefaultLocales.Domain = "monsti-document"
	gettext.DefaultLocales.LocaleDir = settings.Monsti.GetLocalePath()

	renderer.Root = settings.Monsti.GetTemplatesPath()

	provider := service.NewNodeProvider(logger, infoPath)
	contactform := service.NodeTypeHandler{
		Name:       "ContactForm",
		ViewAction: view,
		EditAction: edit,
	}
	provider.AddNodeType(&contactform)
	if err := provider.Serve(settings.Monsti.GetServicePath(
		service.Node.String() + "_contactform")); err != nil {
		panic("Could not setup node provider: " + err.Error())
	}
}
