package main

import (
	"code.google.com/p/gorilla/schema"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"fmt"
	"log"
)

var renderer template.Renderer
var schemaDecoder *schema.Decoder

type contactFormData struct {
	Name, Email, Subject, Message string
}

func render(c client.Connection, node client.Node, data *contactFormData, submitted bool, errors template.FormErrors) string {
	body := c.GetNodeData(node.Path, "body.html")
	context := map[string]string{"body": string(body)}
	if submitted {
		context["submitted"] = "1"
	}
	return renderer.Render("view/contactform.html",
		context, errors, data)
}

func (data *contactFormData) Check() (e template.FormErrors) {
	e = make(template.FormErrors)
	e.Check("Name", data.Name, template.Required())
	e.Check("Email", data.Email, template.Required())
	e.Check("Subject", data.Subject, template.Required())
	e.Check("Message", data.Message, template.Required())
	return
}

func get(req client.Request, res *client.Response, c client.Connection) {
	_, submitted := req.Query["submitted"]
	res.HideSidebar = true
	fmt.Fprint(res, render(c, req.Node, nil, submitted, nil))
}

func post(req client.Request, res *client.Response, c client.Connection) {
	var form contactFormData
	data := c.GetFormData()
        log.Println(data)
	error := schemaDecoder.Decode(&form, data)
	switch e := error.(type) {
	case nil:
		fe := form.Check()
		if len(fe) > 0 {
			fmt.Fprint(res, render(c, req.Node, &form, false, fe))
			return
		}
		/*sendMail(form.Email, []string{"foo@bar.com"}, "foobar",
		[]byte("blabla"), settings)*/
		res.Redirect = req.Node.Path + "/?submitted"
	case schema.MultiError:
		fmt.Fprint(res, render(c, req.Node, &form, false,
			template.ToTemplateErrors(e)))
		return
	default:
		panic("contactform: Could not decode: " + e.Error())
	}
}

func main() {
        schemaDecoder = schema.NewDecoder()
	renderer.Root = "/home/cneumann/dev/monsti/templates/"
	log.Println("Setting up contactform.")
	client.NewConnection("contactform").Serve(get, post)
}
