package main

import (
	"bytes"
	"code.google.com/p/gorilla/sessions"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"datenkarussell.de/monsti/worker"
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"net/http"
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

	// Get action if specified in URL.
	tokens := strings.Split(r.URL.Path, "/")
	last := tokens[len(tokens)-1]
	var action string
	if len(last) > 2 && last[:2] == "@@" {
		action = last[2:]
	}
	nodePath := r.URL.Path
	if len(action) > 0 {
		nodePath = r.URL.Path[:len(r.URL.Path)-(len(action)+3)]
		if len(nodePath) == 0 {
			nodePath = "/"
		}
	}
	session := getSession(r, h.Settings)
	switch action {
	case "login":
		h.Login(w, r, nodePath, session)
	case "logout":
		h.Logout(w, r, nodePath, session)
	default:
		h.RequestNode(w, r, nodePath, action, session)
	}
}

// RequestNode handles node requests.
func (h *nodeHandler) RequestNode(w http.ResponseWriter, r *http.Request,
	nodePath string, action string, session *sessions.Session) {
	// Setup ticket and send to workers.
	log.Println(r.Method, r.URL.Path)
	node, err := lookupNode(h.Settings.Directories.Data, nodePath)
	if err != nil {
		log.Println("Node not found.")
		http.Error(w, "Node not found: "+err.Error(), http.StatusNotFound)
		return
	}
	log.Printf("Node: %v %q\n", node.Type, node.Title)
	c := make(chan client.Response)
	clientSession := getClientSession(session, h.Settings.Directories.Config)
	h.QueueTicket(worker.Ticket{
		Node:         node,
		Request:      r,
		ResponseChan: c,
		Session:      *clientSession,
		Action:       action})
	log.Println("Sent ticket to node queue, wating to finish.")

	// Process response received from a worker.
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
		Session:      clientSession,
		Node:         node,
		PrimaryNav:   prinav,
		SecondaryNav: secnav}
	log.Println(clientSession)
	content := renderInMaster(h.Renderer, res.Body, &env, h.Settings)
	err = session.Save(r, w)
	if err != nil {
		panic(err.Error())
	}
	fmt.Fprint(w, content)
}

// AddNodeProcess starts a worker process to handle the given node type.
func (h *nodeHandler) AddNodeProcess(nodeType string) {
	if _, ok := h.NodeQueues[nodeType]; !ok {
		h.NodeQueues[nodeType] = make(chan worker.Ticket)
	}
	nodeRPC := NodeRPC{Settings: h.Settings}
	worker := worker.NewWorker(nodeType, h.NodeQueues[nodeType],
		&nodeRPC, h.Settings.Directories.Config)
	nodeRPC.Worker = worker
	callback := func() {
		h.AddNodeProcess(nodeType)
	}
	if err := worker.Run(callback); err != nil {
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
