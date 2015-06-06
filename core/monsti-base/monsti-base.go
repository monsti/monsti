// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann
//
// Monsti is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// Monsti is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Monsti.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"net/http"

	"path"
	"github.com/chrneumann/htmlwidgets"
	gomail "gopkg.in/gomail.v1"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util/i18n"
	"pkg.monsti.org/monsti/api/util/module"
)

var availableLocales = []string{"de", "en"}

func setup(c *module.ModuleContext) error {
	gettext.DefaultLocales.Domain = "monsti-daemon"
	G := func(in string) string { return in }
	m := c.Session.Monsti()

	nodeType := &service.NodeType{
		Id:        "core.ContactForm",
		AddableTo: []string{"."},
		Name:      i18n.GenLanguageMap(G("Contact form"), availableLocales),
		Fields: append(service.CoreFields, []*service.FieldConfig{
			{
				Id:     "core.ContactFormFields",
				Hidden: true,
				Type: &service.ListFieldType{
					&service.CombinedFieldType{map[string]service.FieldConfig{
						"Name":     {Type: new(service.TextFieldType)},
						"Required": {Type: new(service.BoolFieldType)},
						"Field": {Type: &service.DynamicTypeFieldType{
							Fields: []service.FieldConfig{
								{
									Id:   "text",
									Type: new(service.DummyFieldType),
								},
								{
									Id:   "textarea",
									Type: new(service.DummyFieldType),
								},
							}}}}}}}}...)}
	if err := m.RegisterNodeType(nodeType); err != nil {
		c.Logger.Fatalf("Could not register %q node type: %v", nodeType.Id, err)
	}

	handler := service.NewRenderNodeHandler(c.Sessions,
		func(args *service.RenderNodeArgs, session *service.Session) (
			*service.RenderNodeRet, error) {
			if args.NodeType == "core.ContactForm" {
				req, err := session.Monsti().GetRequest(args.Request)
				if err != nil || req == nil {
					return nil, fmt.Errorf("Could not get request: %v", err)
				}
				return renderContactForm(req, session)
			}
			return nil, nil
		})
	if err := m.AddSignalHandler(handler); err != nil {
		c.Logger.Fatalf("Could not add signal handler: %v", err)
	}

	return nil
}

type dataField struct {
	Id, Name string
}

func renderContactForm(req *service.Request, session *service.Session) (
	*service.RenderNodeRet, error) {

	m := session.Monsti()
	siteSettings, err := m.LoadSiteSettings(req.Site)
	if err != nil {
		return nil, fmt.Errorf("Could not get site settings: %v", err)
	}
	G, _, _, _ := gettext.DefaultLocales.Use("",
		siteSettings.Fields["core.Locale"].Value().(string))

	node, err := m.GetNode(req.Site, req.NodePath)
	if err != nil {
		return nil, fmt.Errorf("Could not get contact form node: %v", err)
	}

	data := make(service.NestedMap)
	var dataFields []dataField
	form := htmlwidgets.NewForm(data)

	formFields := node.Fields["core.ContactFormFields"].(*service.ListField)

	for i, field := range formFields.Fields {
		combinedField := field.(*service.CombinedField)
		name := combinedField.Fields["Name"].Value().(string)
		required := combinedField.Fields["Required"].Value().(bool)
		fieldId := fmt.Sprintf("field_%d", i)
		dataFields = append(dataFields, dataField{fieldId, name})
		data[fieldId] = ""
		innerFieldType := combinedField.Fields["Field"].(*service.DynamicTypeField).DynamicType
		var widget htmlwidgets.Widget
		switch innerFieldType {
		case "text":
			textWidget := &htmlwidgets.TextWidget{ValidationError: G("Required.")}
			if required {
				textWidget.MinLength = 1
			}
			widget = textWidget
		case "textarea":
			areaWidget := &htmlwidgets.TextAreaWidget{ValidationError: G("Required.")}
			if required {
				areaWidget.MinLength = 1
			}
			widget = areaWidget
		default:
			panic(fmt.Sprintf("Unknow inner field type <%v>", innerFieldType))
		}
		form.AddWidget(widget, fieldId, name, "")
	}

	context := make(map[string]interface{})
	switch req.Method {
	case "GET":
		if _, submitted := req.Form["submitted"]; submitted {
			context["Submitted"] = 1
		}
	case "POST":
		if form.Fill(req.PostForm) {
			mail := gomail.NewMessage()
			mail.SetAddressHeader("From",
				siteSettings.StringValue("core.EmailAddress"),
				siteSettings.StringValue("core.EmailName"))
			mail.SetAddressHeader("To",
				siteSettings.StringValue("core.OwnerEmail"),
				siteSettings.StringValue("core.OwnerName"))
			// mail.SetAddressHeader("Reply-To", data.Email, data.Name)
			mail.SetHeader("Subject", "Contact form submit")
			var fieldValues string
			for _, v := range dataFields {
				fieldValues += fmt.Sprintf("%v: %v\n", v.Name, data[v.Id])
			}
			body := fmt.Sprintf("%v\n\n%v",
				fmt.Sprintf(G("Received from contact form at %v"),
					siteSettings.StringValue("core.Title")), fieldValues)
			mail.SetBody("text/plain", body)
			mailer := gomail.NewCustomMailer("", nil, gomail.SetSendMail(
				m.SendMailFunc()))
			err := mailer.Send(mail)
			if err != nil {
				return nil, fmt.Errorf("Could not send mail: %v", err)
			}
			return &service.RenderNodeRet{
				Redirect: &service.Redirect{
					path.Dir(node.Path) + "/?submitted", http.StatusSeeOther}}, nil
		}
	default:
		return nil, fmt.Errorf("Request method not supported: %v", req.Method)
	}
	context["Form"] = form.RenderData()
	return &service.RenderNodeRet{Context: context}, err
}

func main() {
	module.StartModule("base", setup)
}
