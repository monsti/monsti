package main

import (
	"datenkarussell.de/monsti/form"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"datenkarussell.de/monsti/util"
	"flag"
	"fmt"
	"github.com/chrneumann/g5t"
	"github.com/chrneumann/mimemail"
	htmlT "html/template"
	"log"
)

var G func(string) string = g5t.String

type cfsettings struct {
	// Absolute paths to used directories.
	Directories struct {
		// HTML Templates
		Templates string
		// Locales, i.e. the gettext machine objects (.mo)
		Locales string
	}
}

var renderer template.Renderer
var settings cfsettings

type contactFormData struct {
	Name, Email, Subject, Message string
}

func handle(req client.Request, res *client.Response, c client.Connection) {
	switch req.Action {
	case "edit":
		edit(req, res, c)
	default:
		view(req, res, c)
	}
}

func view(req client.Request, res *client.Response, c client.Connection) {
	data := contactFormData{}
	form := form.NewForm(&data, form.Fields{
		"Name":    form.Field{G("Name"), "", form.Required(), nil},
		"Email":   form.Field{G("Email"), "", form.Required(), nil},
		"Subject": form.Field{G("Subject"), "", form.Required(), nil},
		"Message": form.Field{G("Message"), "", form.Required(),
			new(form.TextArea)}})
	context := template.Context{}
	switch req.Method {
	case "GET":
		if _, submitted := req.Query["submitted"]; submitted {
			context["Submitted"] = 1
		}
	case "POST":
		if form.Fill(c.GetFormData()) {
			c.SendMail(mimemail.Mail{
				From:    mimemail.Address{data.Name, data.Email},
				Subject: data.Subject,
				Body:    []byte(data.Message)})
			res.Redirect = req.Node.Path + "/?submitted"
			return
		}
	default:
		panic("Request method not supported: " + req.Method)
	}
	res.Node = &req.Node
	res.Node.HideSidebar = true
	body := c.GetNodeData(req.Node.Path, "body.html")
	context["Body"] = htmlT.HTML(string(body))
	context["Form"] = form.RenderData()
	fmt.Fprint(res, renderer.Render("view/contactform", context))
}

type editFormData struct {
	Title, Body string
}

func edit(req client.Request, res *client.Response, c client.Connection) {
	data := editFormData{}
	form := form.NewForm(&data, form.Fields{
		"Title": form.Field{G("Title"), "", form.Required(), nil},
		"Body": form.Field{G("Body"), "", form.Required(),
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
	fmt.Fprint(res, renderer.Render("edit/contactform",
		template.Context{"Form": form.RenderData()}))
}

func main() {
	log.SetPrefix("contactform ")
	flag.Parse()
	cfgPath := util.GetConfigPath("contactform", flag.Arg(0))
	err := util.ParseYAML(cfgPath, &settings)
	if err != nil {
		panic("Could not load contactform configuration file: " + err.Error())
	}
	err = g5t.Setup("monsti", settings.Directories.Locales, "de", g5t.GettextParser)
	if err != nil {
		panic("Could not setup gettext: " + err.Error())
	}
	renderer.Root = settings.Directories.Templates
	client.NewConnection("contactform").Serve(handle)
}
