package main

import (
	"bytes"
	"github.com/gorilla/sessions"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/template"
	"datenkarussell.de/monsti/worker"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

// nodeHandler is a net/http handler to process incoming HTTP requests.
type nodeHandler struct {
	Renderer template.Renderer
	Settings settings
	// Hosts is a map from hosts to site names.
	Hosts      map[string]string
	NodeQueues map[string]chan worker.Ticket
}

// QueueTicket adds a ticket to the ticket queue of the corresponding
// node type (ticket.Node.Type).
func (h *nodeHandler) QueueTicket(ticket worker.Ticket) {
	nodeType := ticket.Node.Type
	if _, ok := h.NodeQueues[nodeType]; !ok {
		panic("Missing queue for node type " + nodeType)
	}
	h.NodeQueues[nodeType] <- ticket
}

// splitAction splits and returns the path and @@action of the given URL.
func splitAction(path string) (string, string) {
	tokens := strings.Split(path, "/")
	last := tokens[len(tokens)-1]
	var action string
	if len(last) > 2 && last[:2] == "@@" {
		action = last[2:]
	}
	nodePath := path
	if len(action) > 0 {
		nodePath = path[:len(path)-(len(action)+3)]
		if len(nodePath) == 0 {
			nodePath = "/"
		}
	}
	return nodePath, action
}

// ServeHTTP handles incoming HTTP requests.
func (h *nodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "panic: %v\n", err)
			buf.Write(debug.Stack())
			log.Println(buf.String())
			http.Error(w, "Application error.",
				http.StatusInternalServerError)
		}
	}()
	nodePath, action := splitAction(r.URL.Path)
	site_name, ok := h.Hosts[r.Host]
	if !ok {
		panic("No site found for host " + r.Host)
	}
	site := h.Settings.Sites[site_name]
	site.Name = site_name
	session := getSession(r, site)
	cSession := getClientSession(session, h.Settings.Directories.Config)
	cSession.Locale = site.Locale
	node, err := lookupNode(site.Directories.Data, nodePath)
	if err != nil {
		log.Println("Node not found.")
		http.Error(w, "Node not found: "+err.Error(), http.StatusNotFound)
		return
	}

	if !checkPermission(action, cSession) {
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return
	}
	switch action {
	case "login":
		h.Login(w, r, node, session, cSession, site)
	case "logout":
		h.Logout(w, r, node, session)
	case "add":
		h.Add(w, r, node, session, cSession, site)
	case "remove":
		h.Remove(w, r, node, session, cSession, site)
	default:
		h.RequestNode(w, r, node, action, session, cSession, site)
	}
}

// RequestNode handles node requests.
func (h *nodeHandler) RequestNode(w http.ResponseWriter, r *http.Request,
	node client.Node, action string, session *sessions.Session,
	cSession *client.Session, site site) {
	// Setup ticket and send to workers.
	log.Println(site.Name, r.Method, r.URL.Path)
	c := make(chan client.Response)
	h.QueueTicket(worker.Ticket{
		Node:         node,
		Request:      r,
		ResponseChan: c,
		Session:      *cSession,
		Action:       action})

	// Process response received from a worker.
	res := <-c
	if res.Node != nil {
		oldPath := node.Path
		node = *res.Node
		node.Path = oldPath
	}
	if len(res.Redirect) > 0 {
		http.Redirect(w, r, res.Redirect, http.StatusSeeOther)
		return
	}
	env := masterTmplEnv{Node: node, Session: cSession}
	if action == "edit" {
		env.Title = fmt.Sprintf(G("Edit \"%s\""), node.Title)
		env.Flags = EDIT_VIEW
	}
	content := renderInMaster(h.Renderer, res.Body, env, h.Settings, site)
	err := session.Save(r, w)
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
