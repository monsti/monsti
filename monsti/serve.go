package main

import (
	"bytes"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"datenkarussell.de/monsti/util"
	"datenkarussell.de/monsti/worker"
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

// nodeHandler is a net/http handler to process incoming HTTP requests.
type nodeHandler struct {
	Renderer   template.Renderer
	Settings   settings
	NodeQueues map[string]chan worker.Ticket
}

// QueueTicket adds a ticket to the ticket queue of the corresponding
// node type (ticket.Node.Type).
func (h *nodeHandler) QueueTicket(ticket worker.Ticket) {
	nodeType := ticket.Node.Type
	log.Println("Queuing ticket for node type " + nodeType)
	if _, ok := h.NodeQueues[nodeType]; !ok {
		panic("Missing queue for node type " + nodeType)
	}
	h.NodeQueues[nodeType] <- ticket
}

// ServeHTTP handles incoming HTTP requests.
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
	h.QueueTicket(worker.Ticket{
		ResponseChan: c,
		Node:         node,
		Request:      r})
	log.Println("Sent ticket to node queue, wating to finish.")
	h.processResponse(<-c, node, w, r)
}

// Process response received via RPC from a worker.
func (h *nodeHandler) processResponse(res client.Response, node client.Node,
	w http.ResponseWriter, r *http.Request) {
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
	content := renderInMaster(h.Renderer, res.Body, &env, h.Settings)
	fmt.Fprint(w, content)
}

// AddNodeProcess starts a worker process to handle the given node type.
func (h *nodeHandler) AddNodeProcess(nodeType string) {
	if _, ok := h.NodeQueues[nodeType]; !ok {
		h.NodeQueues[nodeType] = make(chan worker.Ticket)
	}
	nodeRPC := NodeRPC{Settings: h.Settings}
	worker, err := worker.NewWorker(nodeType, h.NodeQueues[nodeType],
		&nodeRPC, h.Settings.Directories.Config)
	if err != nil {
		panic("Could not create worker: " + err.Error())
	}
	nodeRPC.Worker = worker
	if err = worker.Run(); err != nil {
		panic("Could not run worker: " + err.Error())
	}
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
	if flag.NArg() != 1 {
		fmt.Printf("Usage: %v <config_directory>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	cfgPath := util.GetConfigPath("monsti", flag.Arg(0))
	var settings settings
	err := util.ParseYAML(cfgPath, &settings)
	if err != nil {
		fmt.Println("Could not load configuration file: " + err.Error())
		os.Exit(1)
	}
	settings.Directories.Config = filepath.Dir(cfgPath)
	handler := nodeHandler{
		Renderer:   template.Renderer{Root: settings.Directories.Templates},
		Settings:   settings,
		NodeQueues: make(map[string]chan worker.Ticket)}
	for _, ntype := range settings.NodeTypes {
		handler.AddNodeProcess(ntype)
	}
	http.Handle("/static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Directories.Statics))))
	http.Handle("/site-static/", http.FileServer(http.Dir(
		filepath.Dir(settings.Directories.SiteStatics))))
	http.Handle("/", &handler)
	log.Println("Listening for http connections on :8080")
	http.ListenAndServe(":8080", nil)
}
