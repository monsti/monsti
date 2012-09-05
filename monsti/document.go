package monsti

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
)


// Document is a node consisting of a html body.
type Document struct {
	node
	Body string
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


