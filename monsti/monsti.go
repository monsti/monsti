/*
 Brassica is a simple and resource efficient CMS for low dynamic
 private and small business sites with mostly static pages and simple
 structure.

 monsti/monsti-serve contains a command to start a httpd.
*/
package monsti

import (
	"code.google.com/p/gorilla/schema"
	"errors"
	"fmt"
	"github.com/hoisie/mustache"
	"io/ioutil"
	"launchpad.net/goyaml"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
)

var schemaDecoder = schema.NewDecoder()

// Settings for the application and the site.
type Settings struct {
	MailAuth smtp.Auth

	MailServer string

	// Path to the data directory.
	Root string

	// Path to the static files.
	Statics string

	// Path to the site specific static files.
	SiteStatics string

	// Path to the template directory.
	Templates string
}

// GetSettings returns application and site settings.
func GetSettings() Settings {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	settings := Settings{
                MailServer:  "localhost:12345",
		MailAuth:    smtp.PlainAuth("", "joe", "secret!", "host"),
		Root:        wd,
		Statics:     filepath.Join(filepath.Dir(wd), "static"),
		SiteStatics: filepath.Join(filepath.Dir(wd), "site-static"),
		Templates:   filepath.Join(filepath.Dir(wd), "templates")}
	return settings
}

type masterTmplEnv struct {
	Node                     Node
	PrimaryNav, SecondaryNav []navLink
}

// Renderer represents a template renderer.
type Renderer interface {
	// RenderInMaster renders the named template with the given contexts
	// and master template environment in the master template.
	RenderInMaster(name string, env *masterTmplEnv, settings Settings,
		contexts ...interface{}) string
}

type renderer struct {
	// Root is the absolute path to the template directory.
	Root string
	// MasterTemplate holds the parsed master template.
	MasterTemplate *mustache.Template
}

// Render renders the named template with given context. 
func (r renderer) Render(name string, contexts ...interface{}) string {
	path := filepath.Join(r.Root, name)
	globalFuns := map[string]interface{}{
		"_": "implement me!"}
	content := mustache.RenderFile(path, append(contexts, globalFuns)...)
	return content
}

// NewRenderer returns a new Renderer.
//
// root is the absolute path to the template directory.
func NewRenderer(root string) Renderer {
	var r renderer
	r.Root = root
	path := filepath.Join(r.Root, "master.html")
	tmpl, err := mustache.ParseFile(path)
	r.MasterTemplate = tmpl
	if err != nil {
		panic("Could not load master template: " + err.Error())
	}
	return r
}

func (r renderer) RenderInMaster(name string, env *masterTmplEnv,
	settings Settings, contexts ...interface{}) string {
	content := r.Render(name, contexts...)
	sidebarContent := getSidebar(env.Node.Path(), settings.Root)
	showSidebar := (len(env.SecondaryNav) > 0 || len(sidebarContent) > 0) &&
		!env.Node.HideSidebar()
	belowHeader := getBelowHeader(env.Node.Path(), settings.Root)
	return r.MasterTemplate.Render(env, map[string]interface{}{
		"ShowBelowHeader":  len(belowHeader) > 0,
		"BelowHeader":      belowHeader,
		"Footer":           getFooter(settings.Root),
		"Sidebar":          sidebarContent,
		"Content":          content,
		"ShowSecondaryNav": len(env.SecondaryNav) > 0,
		"ShowSidebar":      showSidebar})
}

// getFooter retrieves the footer.
//
// root is the path to the data directory
//
// Returns an empty string if there is no footer.
func getFooter(root string) string {
	path := filepath.Join(root, "footer.html")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

// getBelowHeader retrieves the below header content for the given node.
//
// path is the node's path.
// root is the path to the data directory.
//
// Returns an empty string if there is no below header content.
func getBelowHeader(path, root string) string {
	file := filepath.Join(root, path, "below_header.html")
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(content)
}

// getSidebar retrieves the sidebar content for the given node.
//
// path is the node's path.
// root is the path to the data directory.
//
// It traverses up to the root until it finds a node with defined sidebar
// content.
//
// Returns an empty string if there is no sidebar content.
func getSidebar(path, root string) string {
	for {
		file := filepath.Join(root, path, "sidebar.html")
		content, err := ioutil.ReadFile(file)
		if err != nil {
			if path == filepath.Dir(path) {
				break
			}
			path = filepath.Dir(path)
			continue
		}
		return string(content)
	}
	return ""
}

// navLink represents a link in the navigation.
type navLink struct {
	Name, Target string
	Active       bool
}

// getNav returns the navigation for the given node.
// 
// node is the path of the node for which to get the navigation.
// active is the currently active node.
// root is the path of the data directory.
//
// The keys of the returned map are the link titles, the values are
// the link targets.
func getNav(node, active, root string) []navLink {
	path := filepath.Join(root, node, "navigation.yaml")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	var navLinks []navLink
	goyaml.Unmarshal(content, &navLinks)
	for i, link := range navLinks {
		if link.Target == active {
			navLinks[i].Active = true
			break
		}
	}
	return navLinks
}

// Node is the interface implemented by the various node types
// (Documents, Images, ...).
type Node interface {
	// Path returns the node's path (e.g. "/node").
	Path() string

	// Title returns the node's title.
	Title() string

	// Description returns the node's description.
	Description() string

	// HideSidebar returns if the node's sidebar should be hidden.
	HideSidebar() bool

	// Get handles a GET request on the node.
	Get(w http.ResponseWriter, r *http.Request, renderer Renderer,
		settings Settings)

	// Post handles a POST request on the node.
	Post(w http.ResponseWriter, r *http.Request, renderer Renderer,
		settings Settings)
}

// node is the base implementation for nodes.
type node struct {
	path string
	data nodeData
}

// nodeData is used for (un)marshaling from/to node.yaml.
type nodeData struct {
	Description  string
	Title        string
	Type         string
	Hide_sidebar bool
}

func (n node) Path() string {
	return n.path
}

func (n node) Title() string {
	return n.data.Title
}

func (n node) Description() string {
	return n.data.Description
}

func (n node) HideSidebar() bool {
	return n.data.Hide_sidebar
}

// Document is a node consisting of a html body.
type Document struct {
	node
	Body string
}

// ContactForm is a node including a body and a contact form.
type ContactForm struct {
	Document
}

func (n Document) Get(w http.ResponseWriter, r *http.Request,
	renderer Renderer, settings Settings) {
	prinav := getNav("/", n.Path(), settings.Root)
	var secnav []navLink = nil
	if n.Path() != "/" {
		secnav = getNav(n.Path(), n.Path(), settings.Root)
	}
	env := masterTmplEnv{
		Node:         n,
		PrimaryNav:   prinav,
		SecondaryNav: secnav}
	content := renderer.RenderInMaster("view/document.html",
		&env, settings, map[string]string{"body": n.Body})
	fmt.Fprint(w, content)
}

func (n Document) Post(w http.ResponseWriter, r *http.Request,
	renderer Renderer, settings Settings) {
	http.Error(w, "Implementation missing.", http.StatusInternalServerError)
}

func fetchDocument(data nodeData, path, root string) *Document {
	document := Document{node: node{path: path, data: data}}
	body_path := filepath.Join(root, path[1:], "body.html")
	body, err := ioutil.ReadFile(body_path)
	if err != nil {
		panic("Body not found: " + body_path)
	}
	document.Body = string(body)
	return &document
}

func (n ContactForm) Render(data *contactFormData, submitted bool,
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

func (n ContactForm) Get(w http.ResponseWriter, r *http.Request,
	renderer Renderer, settings Settings) {
	_, submitted := r.URL.Query()["submitted"]
	fmt.Fprint(w, n.Render(nil, submitted, nil, renderer, settings))
}

// formValidator is a function which validates a string.
type formValidator func(string) error

//  required is a formValidator to check for non empty values.
func required() formValidator {
	return func(value string) error {
		if len(value) == 0 {
			return errors.New("Required.")
		}
		return nil
	}
}

// formErrors holds errors for form fields.
//
// If field 'foo.bar' has an error err, then formErrors["foo.bar:error"] ==
// err.
type formErrors map[string]string

// check if the given field's value is valid.
//
// If it's not valid, add an error to the formErrors.
func (f *formErrors) Check(field string, value string, validators ...formValidator) {
	for _, validator := range validators {
		if err := validator(value); err != nil {
			(*f)[field+":error"] = err.Error()
		}
	}
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

func sendMail(from string, to []string, subject string, message []byte, settings Settings) {
	if err := smtp.SendMail(settings.MailServer, settings.MailAuth, from, to,
		message); err != nil {
		panic("monsti: Could not send email: " + err.Error())
	}
}

func (n ContactForm) Post(w http.ResponseWriter, r *http.Request,
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

// TemplateErrors converts a schema.MultiError to a string map.
//
// An error for the field Foo.Bar will be available under the key
// Foo.Bar:error
func toTemplateErrors(error schema.MultiError) map[string]string {
	vs := make(map[string]string)
	for field, msg := range error {
		vs[field+":error"] = msg.Error()
	}
	return vs
}

// NodeFile is the filename of node description files.
const NodeFile = "node.yaml"

// lookup_node look ups a node at the given path.
// If no such node exists, return nil.
func LookupNode(root, path string) (Node, error) {
	node_path := filepath.Join(root, path[1:], NodeFile)
	content, err := ioutil.ReadFile(node_path)
	if err != nil {
		return nil, err
	}
	var ret Node
	var data nodeData
	goyaml.Unmarshal(content, &data)
	switch data.Type {
	case "Document":
		document := fetchDocument(data, path, root)
		ret = document
	case "ContactForm":
		contactForm := ContactForm{*fetchDocument(data, path, root)}
		contactForm.data.Hide_sidebar = true
		ret = contactForm
	default:
		panic("Unknown node type: " + data.Type)
	}
	return ret, nil
}
