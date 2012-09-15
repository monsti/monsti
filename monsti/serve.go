/*
 HTTPd for Brassica.
*/
package main

import (
	"bytes"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"flag"
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
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
	node, err := lookupNode(h.Settings.Directories.Data, r.URL.Path)
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
	if res.Node != nil {
		oldPath := node.Path
		node = *res.Node
		node.Path = oldPath
	}
	if len(res.Redirect) > 0 {
		http.Redirect(w, r, res.Redirect, http.StatusSeeOther)
	}
	prinav := getNav("/", "/"+strings.SplitN(node.Path[1:], "/", 2)[0],
		h.Settings.Directories.Data)
	var secnav []navLink = nil
	if node.Path != "/" {
		secnav = getNav(node.Path, node.Path, h.Settings.Directories.Data)
	}
	env := masterTmplEnv{
		Node:         node,
		PrimaryNav:   prinav,
		SecondaryNav: secnav}
	content := renderInMaster(h.Renderer, res.Body, &env,
		h.Settings)
	fmt.Fprint(w, content)
}

func (h *nodeHandler) AddNodeProcess(nodeType string) {
	if _, ok := h.NodeQueues[nodeType]; !ok {
		h.NodeQueues[nodeType] = make(chan ticket)
	}
	go listenForRPC(h.Settings, h.NodeQueues[nodeType], nodeType)
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
	flag.Parse()
	cfgPath := flag.Arg(0)
	if !filepath.IsAbs(cfgPath) {
		wd, err := os.Getwd()
		if err != nil {
			panic("Could not get working directory: " + err.Error())
		}
		cfgPath = filepath.Join(wd, cfgPath)
	}
	settings := getSettings(cfgPath)
	handler := nodeHandler{
		Renderer:   template.Renderer{Root: settings.Directories.Templates},
		Settings:   settings,
		NodeQueues: make(map[string]chan ticket)}
	handler.AddNodeProcess("Document")
	handler.AddNodeProcess("ContactForm")
	http.Handle("/static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Directories.Statics))))
	http.Handle("/site-static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Directories.SiteStatics))))
	http.Handle("/", &handler)
	c := make(chan int)
	go func() {
		http.ListenAndServe(":8080", nil)
		c <- 1
	}()
	log.Println("Listening for http connections on :8080")
	<-c
}
