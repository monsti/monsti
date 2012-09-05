package monsti

import (
	"code.google.com/p/gorilla/schema"
	"fmt"
	"net/http"
)

// ContactForm is a node including a body and a contact form.
type contactForm struct {
	Document
}

func (n contactForm) Render(data *contactFormData, submitted bool,
	errors formErrors, renderer Renderer, settings Settings) string {
	prinav := getNav("/", n.Path(), settings.Root)
	env := masterTmplEnv{
		Node:         n,
		PrimaryNav:   prinav,
		SecondaryNav: nil}
	context := map[string]string{"body": n.Body}
	if submitted {
		context["submitted"] = "1"
	}
	return renderer.RenderInMaster("view/contactform.html",
		&env, settings, context, errors, data)
}

func (n contactForm) Get(w http.ResponseWriter, r *http.Request,
	renderer Renderer, settings Settings) {
	_, submitted := r.URL.Query()["submitted"]
	fmt.Fprint(w, n.Render(nil, submitted, nil, renderer, settings))
}


type contactFormData struct {
	Name, Email, Subject, Message string
}

func (data *contactFormData) Check() (e formErrors) {
	e = make(formErrors)
	e.Check("Name", data.Name, required())
	e.Check("Email", data.Email, required())
	e.Check("Subject", data.Subject, required())
	e.Check("Message", data.Message, required())
	return
}

func (n contactForm) Post(w http.ResponseWriter, r *http.Request,
	renderer Renderer, settings Settings) {
	var form contactFormData
	if err := r.ParseForm(); err != nil {
		panic("monsti: Could not parse form.")
	}
	error := schemaDecoder.Decode(&form, r.Form)
	switch e := error.(type) {
	case nil:
		fe := form.Check()
		if len(fe) > 0 {
			fmt.Println(fe)
			fmt.Fprint(w, n.Render(&form, false, fe, renderer,
				settings))
			return
		}
		sendMail(form.Email, []string{"foo@bar.com"}, "foobar",
			[]byte("blabla"), settings)
		http.Redirect(w, r, n.Path()+"/?submitted", http.StatusSeeOther)
	case schema.MultiError:
		fmt.Fprint(w, n.Render(&form, false, toTemplateErrors(e), renderer,
			settings))
		return
	default:
		panic("monsti: Could not decode: " + e.Error())
	}
}
