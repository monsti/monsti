/*
 HTTPd for Brassica.
*/
package main

import (
	"bytes"
	"datenkarussell.de/monsti"
        "fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime/debug"
)

type nodeHandler struct {
	Settings monsti.Settings
}

func (h nodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "panic: %v\n", err)
			buf.Write(debug.Stack())
			log.Print(buf.String())
			http.Error(w, "Application error.",
				http.StatusInternalServerError)
		}
	}()
	log.Println(r.Method, r.URL.Path)
	node, err := monsti.LookupNode(h.Settings.Root, r.URL.Path)
	if err != nil {
		log.Println("Node not found.")
		http.Error(w, "Node not found: "+err.Error(), http.StatusNotFound)
		return
	}
	log.Printf("Node: %T %q\n", node, node.Title())
	renderer := monsti.NewRenderer(h.Settings.Templates)
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
	settings := monsti.GetSettings()
	http.Handle("/static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Statics))))
	http.Handle("/site-static/", http.FileServer(http.Dir(
		filepath.Dir(settings.SiteStatics))))
	http.Handle("/", nodeHandler{settings})
	http.ListenAndServe(":8080", nil)
}
