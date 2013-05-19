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
	"github.com/monsti/service/info"
	"github.com/monsti/service/node"
	"github.com/monsti/util"
	"github.com/monsti/util/l10n"
	"github.com/monsti/util/template"
	htmlT "html/template"
	"log"
	"os"
	"sync"
)

type settings struct {
	Monsti util.MonstiSettings
}

var renderer template.Renderer

func handle(req node.Request, res *node.Response, infoServ *info.Service) {
	switch req.Action {
	case "edit":
		edit(req, res, infoServ)
	default:
		view(req, res, infoServ)
	}
}

type editFormData struct {
	Title, Body string
}

func edit(req node.Request, res *node.Response, infoServ *info.Service) {
	G := l10n.UseCatalog(req.Session.Locale)
	data := editFormData{}
	dataServ, err := infoServ.FindDataService()
	if err != nil {
		panic("document: Could not connect to data service.")
	}
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
		req.Session.Locale, ""))
}

func view(req node.Request, res *node.Response, infoServ *info.Service) {
	body := "yay!" //c.GetNodeData(req.Node.Path, "body.html")
	content := renderer.Render("document/view",
		template.Context{"Body": htmlT.HTML(body)},
		req.Session.Locale, "")
	fmt.Fprint(res, content)
}

type NodeService struct {
	Info *info.Service
}

func (i *NodeService) Request(req node.Request,
	reply *node.Response) error {
	handle(req, reply, i.Info)
	return nil
}

func (i *NodeService) GetNodeTypes(req int,
	reply *[]string) error {
	*reply = []string{
		"Document"}
	return nil
}

func main() {
	logger := log.New(os.Stderr, "document ", log.LstdFlags)

	// Load configuration
	flag.Parse()
	cfgPath := util.GetConfigPath(flag.Arg(0))
	var settings settings
	if err := util.LoadModuleSettings("document", cfgPath, &settings); err != nil {
		logger.Fatal("Could not load settings: ", err)
	}

	l10n.Setup("monsti", settings.Monsti.GetLocalePath())
	renderer.Root = settings.Monsti.GetTemplatesPath()

	// Connect to Info service
	info, err := info.NewConnection(settings.Monsti.GetServicePath(
		service.Info))
	if err != nil {
		logger.Fatalf("Could not connect to INFO service: %v", err)
	}

	// Start own NODE service
	var waitGroup sync.WaitGroup
	logger.Println("Starting NODE service")
	waitGroup.Add(1)
	nodePath := "monsti-document"
	go func() {
		defer waitGroup.Done()
		var provider service.Provider
		var node_ NodeService
		node_.Info = info
		provider.Logger = logger
		if err := provider.Serve(nodePath, "Node", &node_); err != nil {
			logger.Fatalf("Could not start NODE service: %v", err)
		}
	}()

	if err := info.PublishService("Node", nodePath); err != nil {
		logger.Fatalf("Could not publish node service: %v", err)
	}
	waitGroup.Wait()
}
