package main

import (
	"github.com/monsti/form"
	"github.com/monsti/l10n"
	"github.com/monsti/rpc/client"
	"github.com/monsti/template"
	"github.com/monsti/util"
	"flag"
	"fmt"
	"github.com/chrneumann/mimemail"
	htmlT "html/template"
	"log"
	"os"
)

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
	G := l10n.UseCatalog(req.Session.Locale)
	data := contactFormData{}
	form := form.NewForm(&data, form.Fields{
		"Name":    form.Field{G("Name"), "", form.Required(G("Required.")), nil},
		"Email":   form.Field{G("Email"), "", form.Required(G("Required.")), nil},
		"Subject": form.Field{G("Subject"), "", form.Required(G("Required.")), nil},
		"Message": form.Field{G("Message"), "", form.Required(G("Required.")),
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
	fmt.Fprint(res, renderer.Render("view/contactform", context,
		req.Session.Locale, ""))
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
	fmt.Fprint(res, renderer.Render("edit/contactform",
		template.Context{"Form": form.RenderData()},
		req.Session.Locale, ""))
}

func main() {
	logger := log.New(os.Stderr, "contactform", log.LstdFlags)
	flag.Parse()
	cfgPath := util.GetConfigPath("contactform", flag.Arg(0))
	err := util.ParseYAML(cfgPath, &settings)
	if err != nil {
		panic("Could not load contactform configuration file: " + err.Error())
	}
	l10n.DefaultSettings.Domain = "monsti"
	l10n.DefaultSettings.Directory = settings.Directories.Locales
	renderer.Root = settings.Directories.Templates
	client.NewConnection("contactform", logger).Serve(handle)
}
