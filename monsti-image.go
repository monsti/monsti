package main

import (
	"flag"
	"fmt"
	"github.com/monsti/form"
	"github.com/monsti/util/l10n"
	"github.com/monsti/rpc/client"
	"github.com/monsti/util/template"
	"github.com/monsti/util"
	"log"
	"os"
)

type settings struct {
	// Absolute paths to used directories.
	Directories struct {
		// HTML Templates
		Templates string
		// Locales, i.e. the gettext machine objects (.mo)
		Locales string
	}
}

var renderer template.Renderer
var logger *log.Logger

func handle(req client.Request, res *client.Response, c client.Connection) {
	switch req.Action {
	case "edit":
		edit(req, res, c)
	default:
		view(req, res, c)
	}
}

type editFormData struct {
	Title string
	Image string
}

func edit(req client.Request, res *client.Response, c client.Connection) {
	G := l10n.UseCatalog(req.Session.Locale)
	data := editFormData{}
	form := form.NewForm(&data, form.Fields{
		"Title": form.Field{G("Title"), "", form.Required(G("Required.")), nil},
		"Image": form.Field{G("Image"), "", nil, new(form.FileWidget)}})
	switch req.Method {
	case "GET":
		data.Title = req.Node.Title
	case "POST":
		imageData, imgerr := c.GetFileData("Image")
		if form.Fill(c.GetFormData()) {
			if imgerr == nil {
				node := req.Node
				node.Title = data.Title
				c.UpdateNode(node)
				c.WriteNodeData(req.Node.Path, "image.data", string(imageData))
				res.Redirect = req.Node.Path
				return
			}
			logger.Println("Image upload failed: ", imgerr)
			form.AddError("Image", G("There was a problem with your image upload."))
		}
	default:
		panic("Request method not supported: " + req.Method)
	}
	fmt.Fprint(res, renderer.Render("edit/image",
		template.Context{"Form": form.RenderData()},
		req.Session.Locale, ""))
}

func view(req client.Request, res *client.Response, c client.Connection) {
	body := c.GetNodeData(req.Node.Path, "image.data")
	res.Raw = true
	res.Write(body)
}

func main() {
	logger = log.New(os.Stderr, "monsti-image", log.LstdFlags)
	flag.Parse()
	cfgPath := util.GetConfigPath("image", flag.Arg(0))
	var settings settings
	if err := util.ParseYAML(cfgPath, &settings); err != nil {
		logger.Fatal("Could not load monsti-image configuration file: ", err)
	}
	l10n.Setup("monsti-image", settings.Directories.Locales)
	renderer.Root = settings.Directories.Templates
	client.NewConnection("monsti-image", logger).Serve(handle)
}
