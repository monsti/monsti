package main

import (
	"flag"
	"fmt"
	"github.com/chrneumann/mimemail"
	"github.com/monsti/form"
	"github.com/monsti/rpc/client"
	"github.com/monsti/util"
	"github.com/monsti/util/l10n"
	"github.com/monsti/util/template"
	htmlT "html/template"
	"log"
	"os"
	"path/filepath"
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
	body := c.GetNodeData(req.Node.Path, "body.html")
	context["Body"] = htmlT.HTML(string(body))
	context["Form"] = form.RenderData()
	fmt.Fprint(res, renderer.Render("contactform/view", context,
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
	fmt.Fprint(res, renderer.Render("contactform/edit",
		template.Context{"Form": form.RenderData()},
		req.Session.Locale, ""))
}

func main() {
	logger := log.New(os.Stderr, "contactform", log.LstdFlags)
	flag.Parse()
	cfgPath := util.GetConfigPath("contactform", flag.Arg(0))
	if err := util.ParseYAML(cfgPath, &settings); err != nil {
		logger.Fatal("Could not load monsti-contactform configuration file: ", err)
	}
	util.MakeAbsolute(&settings.Directories.Templates, filepath.Dir(cfgPath))
	util.MakeAbsolute(&settings.Directories.Locales, filepath.Dir(cfgPath))
	l10n.Setup("monsti-contactform", settings.Directories.Locales)
	renderer.Root = settings.Directories.Templates
	client.NewConnection("contactform", logger).Serve(handle)
}
