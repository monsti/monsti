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
	data := contactFormData{}
	form := form.NewForm(&data, form.Fields{
		"Name":     form.Field{G("Name"), "", form.Required(), nil},
		"Email":    form.Field{G("Email"), "", form.Required(), nil},
		"Subject:": form.Field{G("Subject"), "", form.Required(), nil},
		"Message":  form.Field{G("Message"), "", form.Required(), nil}})
	context := make(map[string]interface{})
	switch req.Method {
	case "GET":
		if _, submitted := req.Query["submitted"]; submitted {
			context["submitted"] = 1
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
	context["body"] = string(body)
	fmt.Fprint(res, renderer.Render("view/contactform.html",
		context, form.RenderData()))
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
