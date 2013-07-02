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
	"github.com/monsti/form"
	"github.com/monsti/service"
	"github.com/monsti/util"
	"github.com/monsti/util/l10n"
	"github.com/monsti/util/template"
	htmlT "html/template"
	"log"
	"os"
)

var settings struct {
	Monsti util.MonstiSettings
}

var logger *log.Logger
var renderer template.Renderer

type editFormData struct {
	Title, Body string
}

func edit(req service.Request, res *service.Response, s *service.Session) {
	G := l10n.UseCatalog(req.Session.Locale)
	data := editFormData{}
	dataServ := s.Data()
	form := form.NewForm(&data, form.Fields{
		"Title": form.Field{G("Title"), "", form.Required(G("Required.")), nil},
		"Body": form.Field{G("Body"), "", form.Required(G("Required.")),
			new(form.AlohaEditor)}})
	switch req.Method {
	case "GET":
		data.Title = req.Node.Title
		body, err := dataServ.GetNodeData(req.Site, req.Node.Path,
			"body.html")
		if err != nil {
			panic("document: Could not get node data")
		}
		data.Body = string(body)
	case "POST":
		if form.Fill(req.FormData) {
			node := req.Node
			node.Title = data.Title
			if err := dataServ.UpdateNode(req.Site, node); err != nil {
				panic("document: Could not update node: " + err.Error())
			}
			if err := dataServ.WriteNodeData(req.Site, req.Node.Path,
				"body.html", data.Body); err != nil {
				panic("document: Could not update node: " + err.Error())
			}
			res.Redirect = req.Node.Path
			return
		}
	default:
		panic("Request method not supported: " + req.Method)
	}
	fmt.Fprint(res, renderer.Render("document/edit",
		template.Context{"Form": form.RenderData()},
		req.Session.Locale, settings.Monsti.GetSiteTemplatesPath(req.Site)))
}

func view(req service.Request, res *service.Response, s *service.Session) {
	dataServ := s.Data()
	body, err := dataServ.GetNodeData(req.Site, req.Node.Path, "body.html")
	if err != nil {
		logger.Fatalf("Could not fetch node data: %v", err)
	}
	content := renderer.Render("document/view",
		template.Context{"Body": htmlT.HTML(body)},
		req.Session.Locale, settings.Monsti.GetSiteTemplatesPath(req.Site))
	fmt.Fprint(res, content)
}

func main() {
	logger = log.New(os.Stderr, "document ", log.LstdFlags)
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

	l10n.Setup("monsti-document", settings.Monsti.GetLocalePath())
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
		panic("Could not setup node provider: " + err.Error())
	}
}
