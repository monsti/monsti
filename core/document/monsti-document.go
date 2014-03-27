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

 This package implements the document node type.
*/
package main

import (
	"flag"
	"fmt"
	htmlT "html/template"
	"log"
	"os"

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

type editFormData struct {
	Title, Body string
}

func edit(req service.Request, res *service.Response, s *service.Session) error {
	G, _, _, _ := gettext.DefaultLocales.Use("", req.Session.Locale)
	data := editFormData{}
	dataServ := s.Data()
	form := form.NewForm(&data, form.Fields{
		"Title": form.Field{G("Title"), "", form.Required(G("Required.")), nil},
		"Body": form.Field{G("Body"), "", form.Required(G("Required.")),
			new(form.AlohaEditor)}})
	switch req.Method {
	case service.GetRequest:
		data.Title = req.Node.Title
		body, err := dataServ.GetNodeData(req.Site, req.Node.Path,
			"body.html")
		if err != nil {
			return fmt.Errorf("document: Could not get node data")
		}
		data.Body = string(body)
	case service.PostRequest:
		if form.Fill(req.FormData) {
			var node struct{ service.NodeFields }
			node.NodeFields = req.Node
			node.Title = data.Title
			err := dataServ.WriteNode(req.Site, node.Path, node, "node")
			if err != nil {
				return fmt.Errorf("document: Could not update node: ", err)
			}
			if err := dataServ.WriteNodeData(req.Site, req.Node.Path,
				"body.html", []byte(data.Body)); err != nil {
				return fmt.Errorf("document: Could not update node: %v", err.Error())
			}
			res.Redirect = req.Node.Path
			return nil
		}
	default:
		return fmt.Errorf("Request method not supported: %v", req.Method)
	}
	rendered, err := renderer.Render("document/edit",
		template.Context{"Form": form.RenderData()},
		req.Session.Locale, settings.Monsti.GetSiteTemplatesPath(req.Site))
	if err != nil {
		return fmt.Errorf("Could not render template: %v", err)
	}
	fmt.Fprint(res, rendered)
	return nil
}

func view(req service.Request, res *service.Response, s *service.Session) error {
	dataServ := s.Data()
	body, err := dataServ.GetNodeData(req.Site, req.Node.Path, "body.html")
	if err != nil {
		logger.Fatalf("Could not fetch node data: %v", err)
	}
	rendered, err := renderer.Render("document/view",
		template.Context{"Body": htmlT.HTML(body)},
		req.Session.Locale, settings.Monsti.GetSiteTemplatesPath(req.Site))
	if err != nil {
		return fmt.Errorf("Could not render template: %v", err)
	}
	fmt.Fprint(res, rendered)
	return nil
}

func main() {
	logger = log.New(os.Stderr, "", 0)
	// Load configuration
	flag.Parse()
	if flag.NArg() != 1 {
		logger.Fatal("Expecting configuration path.")
	}
	cfgPath := util.GetConfigPath(flag.Arg(0))
	if err := util.LoadModuleSettings("document", cfgPath, &settings); err != nil {
		logger.Fatal("Could not load settings: ", err)
	}

	infoPath := settings.Monsti.GetServicePath(service.Info.String())

	gettext.DefaultLocales.Domain = "monsti-document"
	gettext.DefaultLocales.LocaleDir = settings.Monsti.GetLocalePath()

	renderer.Root = settings.Monsti.GetTemplatesPath()

	provider := service.NewNodeProvider(logger, infoPath)
	document := service.NodeTypeHandler{
		Name:       "Document",
		ViewAction: view,
		EditAction: edit,
	}
	provider.AddNodeType(&document)
	if err := provider.Serve(settings.Monsti.GetServicePath(
		service.Node.String() + "_document")); err != nil {
		panic(fmt.Sprintf("Could not setup node provider: %v", err))
	}
}
