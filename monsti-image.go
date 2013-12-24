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

 This package implements the image node type.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"pkg.monsti.org/form"
	"pkg.monsti.org/service"
	"pkg.monsti.org/util"
	"pkg.monsti.org/util/l10n"
	"pkg.monsti.org/util/template"
)

var settings struct {
	Monsti util.MonstiSettings
}

var logger *log.Logger
var renderer template.Renderer

type editFormData struct {
	Title string
	Image string
}

func edit(req service.Request, res *service.Response, s *service.Session) {
	G := l10n.UseCatalog(req.Session.Locale)
	data := editFormData{}
	form := form.NewForm(&data, form.Fields{
		"Title": form.Field{G("Title"), "", form.Required(G("Required.")), nil},
		"Image": form.Field{G("Image"), "", nil, new(form.FileWidget)}})
	switch req.Method {
	case "GET":
		data.Title = req.Node.Title
	case "POST":
		var imageData []byte
		var err error
		if len(req.Files["Image"]) == 1 {
			imageData, err = req.Files["Image"][0].ReadFile()
			if err != nil {
				panic("Could not read image data: " + err.Error())
			}
		}
		if form.Fill(req.FormData) {
			if len(imageData) > 0 {
				dataC := s.Data()
				node := req.Node
				node.Title = data.Title
				node.Hide = true
				if err := dataC.UpdateNode(req.Site, node); err != nil {
					panic("Could not update node: " + err.Error())
				}
				if err := dataC.WriteNodeData(req.Site, req.Node.Path,
					"image.data", string(imageData)); err != nil {
					panic("Could not write image data: " + err.Error())
				}
				res.Redirect = req.Node.Path
				return
			}
			form.AddError("Image", G("There was a problem with your image upload."))
		}
	default:
		panic("Request method not supported: " + req.Method)
	}
	rendered, err := renderer.Render("image/edit",
		template.Context{"Form": form.RenderData()},
		req.Session.Locale, "")
	if err != nil {
		panic("Could not render template: " + err.Error())
	}
	fmt.Fprint(res, rendered)
}

func view(req service.Request, res *service.Response, s *service.Session) {
	body, err := s.Data().GetNodeData(req.Site, req.Node.Path,
		"image.data")
	if err != nil {
		panic("Could not get image data: " + err.Error())
	}
	res.Raw = true
	res.Write(body)
}

func main() {
	logger = log.New(os.Stderr, "image ", log.LstdFlags)
	// Load configuration
	flag.Parse()
	if flag.NArg() != 1 {
		logger.Fatal("Expecting configuration path.")
	}
	cfgPath := util.GetConfigPath(flag.Arg(0))
	if err := util.LoadModuleSettings("image", cfgPath, &settings); err != nil {
		logger.Fatal("Could not load settings: ", err)
	}

	infoPath := settings.Monsti.GetServicePath(service.Info.String())

	l10n.Setup("monsti-image", settings.Monsti.GetLocalePath())
	renderer.Root = settings.Monsti.GetTemplatesPath()

	provider := service.NewNodeProvider(logger, infoPath)
	image := service.NodeTypeHandler{
		Name:       "Image",
		ViewAction: view,
		EditAction: edit,
	}
	provider.AddNodeType(&image)
	if err := provider.Serve(settings.Monsti.GetServicePath(
		service.Node.String() + "_image")); err != nil {
		panic("Could not setup node provider: " + err.Error())
	}
}
