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

func get(req client.Request, res *client.Response, c client.Connection) {
	body := c.GetNodeData(req.Node.Path, "body.html")
	content := renderer.Render("view/document.html",
		map[string]string{"body": string(body)}, req.Node)
	fmt.Fprint(res, content)
}

func post(req client.Request, res *client.Response, c client.Connection) {
	panic("document: Implementation missing.")
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
	log.Println("Setting up document.")
	client.NewConnection("document").Serve(get, post)
}
