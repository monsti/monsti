/*
 HTTPd for Brassica.
*/
package main

import (
	"datenkarussell.de/brassica"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path) 
	settings := brassica.GetSettings()
	node, err := brassica.LookupNode(settings.Root, r.URL.Path)
	if err != nil {
		log.Println("Node not found.")
		http.Error(w, "Node not found: " + err.Error(), http.StatusNotFound)
		return
	}
	log.Printf("Node: %T %q\n", node, node.Title())
	renderer := brassica.NewRenderer("/home/cneumann/dev/brassica/templates")
	switch r.Method {
	case "GET":
		node.Get(w, r, renderer, settings)
	case "POST":
		node.Post(w, r, renderer, settings)
	default:
		http.Error(w, "Wrong method", http.StatusNotFound)
	}
}

func main() {
	http.Handle("/static/", http.FileServer(http.Dir("/home/cneumann/dev/brassica/")))
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}