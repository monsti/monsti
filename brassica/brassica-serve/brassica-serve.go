/*
 HTTPd for Brassica.
*/
package main

import (
	"datenkarussell.de/brassica"
	"log"
	"net/http"
	"path/filepath"
)

type nodeHandler struct {
	Settings brassica.Settings
}

func (h nodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path)
	node, err := brassica.LookupNode(h.Settings.Root, r.URL.Path)
	if err != nil {
		log.Println("Node not found.")
		http.Error(w, "Node not found: "+err.Error(), http.StatusNotFound)
		return
	}
	log.Printf("Node: %T %q\n", node, node.Title())
	renderer := brassica.NewRenderer(h.Settings.Templates)
	switch r.Method {
	case "GET":
		node.Get(w, r, renderer, h.Settings)
	case "POST":
		node.Post(w, r, renderer, h.Settings)
	default:
		http.Error(w, "Wrong method", http.StatusNotFound)
	}
}

func main() {
	settings := brassica.GetSettings()
	http.Handle("/static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Statics))))
	http.Handle("/site-static/", http.FileServer(http.Dir(
		filepath.Dir(settings.SiteStatics))))
	http.Handle("/", nodeHandler{settings})
	http.ListenAndServe(":8080", nil)
}
