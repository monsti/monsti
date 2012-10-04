package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"datenkarussell.de/monsti/util"
	"flag"
	"fmt"
	"log"
)

type settings struct {
	// Absolute paths to used directories.
	Directories struct {
		// HTML Templates
		Templates string
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

func (data *editFormData) Check(e *template.FormErrors) {
	e.Check("Title", data.Title, template.Required())
}

func edit(req client.Request, res *client.Response, c client.Connection) {
	var data editFormData
	var errors template.FormErrors
	switch req.Method {
	case "GET":
		data.Title = req.Node.Title
		data.Body = string(c.GetNodeData(req.Node.Path, "body.html"))
	case "POST":
		var err error
		errors, err = template.Validate(c.GetFormData(), &data)
		if err != nil {
			panic("Could not parse form data: " + err.Error())
		}
		if len(errors) == 0 {
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
	fmt.Fprint(res, renderer.Render("edit/document.html",
		errors, data))
}

func view(req client.Request, res *client.Response, c client.Connection) {
	body := c.GetNodeData(req.Node.Path, "body.html")
	content := renderer.Render("view/document.html",
		map[string]string{"body": string(body)}, req.Node)
	fmt.Fprint(res, content)
}

func main() {
	log.SetPrefix("document")
	flag.Parse()
	cfgPath := util.GetConfigPath("document", flag.Arg(0))
	var settings settings
	err := util.ParseYAML(cfgPath, &settings)
	if err != nil {
		panic("Could not load document configuration file: " + err.Error())
	}
	renderer.Root = settings.Directories.Templates
	client.NewConnection("document").Serve(handle)
}
