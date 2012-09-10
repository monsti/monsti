package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"fmt"
        "log"
)

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
        renderer.Root = "/home/cneumann/dev/monsti/templates/"
        log.Println("Setting up document.")
	client.NewConnection("contactform").Serve(get, post)
}
