package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"datenkarussell.de/monsti/util"
	"flag"
	"fmt"
	"github.com/chrneumann/g5t"
	"github.com/chrneumann/mimemail"
	"log"
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

func (data *contactFormData) Check(e *template.FormErrors) {
	e.Check("Name", data.Name, template.Required())
	e.Check("Email", data.Email, template.Required())
	e.Check("Subject", data.Subject, template.Required())
	e.Check("Message", data.Message, template.Required())
}

func handle(req client.Request, res *client.Response, c client.Connection) {
	var data contactFormData
	var errors template.FormErrors
	context := make(map[string]interface{})
	switch req.Method {
	case "GET":
		if _, submitted := req.Query["submitted"]; submitted {
			context["submitted"] = 1
		}
	case "POST":
		var err error
		errors, err = template.Validate(c.GetFormData(), &data)
		if err != nil {
			panic("Could not parse form data: " + err.Error())
		}
		if len(errors) == 0 {
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
		context, errors, data))
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
