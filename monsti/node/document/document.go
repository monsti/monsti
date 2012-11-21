package main

import (
	"datenkarussell.de/monsti/form"
	"datenkarussell.de/monsti/l10n"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"datenkarussell.de/monsti/util"
	"flag"
	"fmt"
	htmlT "html/template"
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

func handle(req client.Request, res *client.Response, c client.Connection) {
	switch req.Action {
	case "edit":
		edit(req, res, c)
	default:
		view(req, res, c)
	}
}

type editFormData struct {
	Title, Body string
}

func edit(req client.Request, res *client.Response, c client.Connection) {
	G := l10n.UseCatalog(req.Session.Locale)
	data := editFormData{}
	form := form.NewForm(&data, form.Fields{
		"Title": form.Field{G("Title"), "", form.Required(G("Required.")), nil},
		"Body": form.Field{G("Body"), "", form.Required(G("Required.")),
			new(form.AlohaEditor)}})
	switch req.Method {
	case "GET":
		data.Title = req.Node.Title
		data.Body = string(c.GetNodeData(req.Node.Path, "body.html"))
	case "POST":
		if form.Fill(c.GetFormData()) {
			node := req.Node
			node.Title = data.Title
			c.UpdateNode(node)
			c.WriteNodeData(req.Node.Path, "body.html", data.Body)
			res.Redirect = req.Node.Path
			return
		}
	default:
		panic("Request method not supported: " + req.Method)
	}
	fmt.Fprint(res, renderer.Render("edit/document",
		template.Context{"Form": form.RenderData()},
		req.Session.Locale, ""))
}

func view(req client.Request, res *client.Response, c client.Connection) {
	body := c.GetNodeData(req.Node.Path, "body.html")
	content := renderer.Render("view/document",
		template.Context{"Body": htmlT.HTML(body)},
		req.Session.Locale, "")
	fmt.Fprint(res, content)
}

func main() {
	logger := log.New(os.Stderr, "document", log.LstdFlags)
	flag.Parse()
	cfgPath := util.GetConfigPath("document", flag.Arg(0))
	var settings settings
	err := util.ParseYAML(cfgPath, &settings)
	if err != nil {
		panic("Could not load document configuration file: " + err.Error())
	}
	l10n.DefaultSettings.Domain = "monsti"
	l10n.DefaultSettings.Directory = settings.Directories.Locales
	renderer.Root = settings.Directories.Templates
	client.NewConnection("document", logger).Serve(handle)
}
