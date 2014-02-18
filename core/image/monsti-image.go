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

	"github.com/quirkey/magick"
	"pkg.monsti.org/form"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util"
	"pkg.monsti.org/util/l10n"
	"pkg.monsti.org/monsti/api/util/template"
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

func edit(req service.Request, res *service.Response, s *service.Session) error {
	G := l10n.UseCatalog(req.Session.Locale)
	data := editFormData{}
	form := form.NewForm(&data, form.Fields{
		"Title": form.Field{G("Title"), "", form.Required(G("Required.")), nil},
		"Image": form.Field{G("Image"), "", nil, new(form.FileWidget)}})
	switch req.Method {
	case service.GetRequest:
		data.Title = req.Node.Title
	case service.PostRequest:
		var imageData []byte
		var err error
		if len(req.Files["Image"]) == 1 {
			imageData, err = req.Files["Image"][0].ReadFile()
			if err != nil {
				return fmt.Errorf("Could not read image data: %v", err)
			}
		}
		if form.Fill(req.FormData) {
			if len(imageData) > 0 {
				dataC := s.Data()
				var node struct{ service.NodeFields }
				node.NodeFields = req.Node
				node.Title = data.Title
				node.Hide = true
				if err := dataC.WriteNode(req.Site, node.Path, node,
					"node"); err != nil {
					return fmt.Errorf("Could not update node: %v", err)
				}
				if err := dataC.WriteNodeData(req.Site, req.Node.Path,
					"image.data", imageData); err != nil {
					return fmt.Errorf("Could not write image data: %v", err)
				}
				res.Redirect = req.Node.Path
				return nil
			}
			form.AddError("Image", G("There was a problem with your image upload."))
		}
	default:
		return fmt.Errorf("Request method not supported: %v", req.Method)
	}
	rendered, err := renderer.Render("image/edit",
		template.Context{"Form": form.RenderData()},
		req.Session.Locale, "")
	if err != nil {
		return fmt.Errorf("Could not render template: %v", err)
	}
	fmt.Fprint(res, rendered)
	return nil
}

type size struct{ Width, Height int }

func (s size) String() string {
	return fmt.Sprintf("%vx%v", s.Width, s.Height)
}

func view(req service.Request, res *service.Response, s *service.Session) error {
	sizeName := req.Query.Get("size")
	var size size
	var body []byte
	var err error
	if sizeName != "" {
		err = s.Data().GetConfig(req.Site, "image", "sizes."+sizeName, &size)
		if err != nil || size.Width == 0 {
			logger.Printf("Could not find size %q for site %q: %v", sizeName, req.Site,
				err)
		} else {
			sizePath := "image_" + size.String() + ".data"
			body, err = s.Data().GetNodeData(req.Site, req.Node.Path, sizePath)
			if err != nil || body == nil {
				body, err = s.Data().GetNodeData(req.Site, req.Node.Path, "image.data")
				if err != nil {
					return fmt.Errorf("Could not get image data: %v", err)
				}
				image, err := magick.NewFromBlob(body, "jpg")
				if err != nil {
					return fmt.Errorf("Could not open image data with magick: %v", err)
				}
				defer image.Destroy()
				err = image.Resize(size.String())
				if err != nil {
					return fmt.Errorf("Could not resize image: %v", err)
				}
				body, err = image.ToBlob("jpg")
				if err != nil {
					return fmt.Errorf("Could not dump image: %v", err)
				}
				if err := s.Data().WriteNodeData(req.Site, req.Node.Path,
					sizePath, body); err != nil {
					return fmt.Errorf("Could not write resized image data: %v", err)
				}
			}
		}
	}
	if body == nil {
		body, err = s.Data().GetNodeData(req.Site, req.Node.Path, "image.data")
		if err != nil {
			return fmt.Errorf("Could not get image data: %v", err)
		}
	}
	if req.Query.Get("raw") == "1" {
		res.Raw = true
		res.Write(body)
	} else {
		rendered, err := renderer.Render("image/view",
			template.Context{"Image": req.Node},
			req.Session.Locale, "")
		if err != nil {
			return fmt.Errorf("Could not render template: %v", err)
		}
		fmt.Fprint(res, rendered)
	}
	return nil
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
