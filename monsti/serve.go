/*
 HTTPd for Brassica.
*/
package main

import (
	"bytes"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"net/http"
	"path/filepath"
	"runtime/debug"
)

type nodeHandler struct {
	Renderer   template.Renderer
	Settings   settings
	NodeQueues map[string]chan ticket
}

func (h *nodeHandler) QueueTicket(ticket ticket) {
	nodeType := ticket.Node.Type
	log.Println("Queuing ticket for node type " + nodeType)
	if _, ok := h.NodeQueues[nodeType]; !ok {
		panic("Missing queue for node type " + nodeType)
	}
	h.NodeQueues[nodeType] <- ticket
}

func (h *nodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	node, err := lookupNode(h.Settings.Root, r.URL.Path)
	if err != nil {
		log.Println("Node not found.")
		http.Error(w, "Node not found: "+err.Error(), http.StatusNotFound)
		return
	}
	log.Printf("Node: %v %q\n", node.Type, node.Title)
	c := make(chan client.Response)
	h.QueueTicket(ticket{
		ResponseChan: c,
		Node:         node,
		Request:      r})
	log.Println("Sent ticket to node queue, wating to finish.")
	res := <-c
	prinav := getNav("/", node.Path, h.Settings.Root)
	var secnav []navLink = nil
	if node.Path != "/" {
		secnav = getNav(node.Path, node.Path, h.Settings.Root)
	}
	env := masterTmplEnv{
		Node:         node,
		PrimaryNav:   prinav,
		SecondaryNav: secnav}
	content := renderInMaster(h.Renderer, res.Body, &env,
		h.Settings)
	fmt.Fprint(w, content)
}

// lookupNode look ups a node at the given path.
// If no such node exists, return nil.
func lookupNode(root, path string) (client.Node, error) {
	node_path := filepath.Join(root, path[1:], "node.yaml")
	content, err := ioutil.ReadFile(node_path)
	if err != nil {
		return client.Node{}, err
	}
	var node client.Node
	goyaml.Unmarshal(content, &node)
	node.Path = path
	return node, nil
}

func main() {
	log.SetPrefix("monsti ")
	settings := getSettings()
	handler := nodeHandler{
		Renderer:   template.Renderer{Root: settings.Templates},
		Settings:   settings,
		NodeQueues: make(map[string]chan ticket)}
	handler.NodeQueues["Document"] = make(chan ticket)
	go listenForRPC(handler.NodeQueues["Document"])
	http.Handle("/static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Statics))))
	http.Handle("/site-static/", http.FileServer(http.Dir(
		filepath.Dir(settings.SiteStatics))))
	http.Handle("/", &handler)
	c := make(chan int)
	go func() {
		http.ListenAndServe(":8080", nil)
		c <- 1
	}()
	log.Println("Listening for http connections on :8080")
	<-c
}
