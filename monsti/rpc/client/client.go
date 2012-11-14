package client

import (
	"datenkarussell.de/monsti/rpc/types"
	"github.com/chrneumann/mimemail"
	"io"
	"log"
	"net/rpc"
	"net/url"
	"os"
)

type Node struct {
	Path string "path,omitempty"
	// Content type of the node.
	Type        string
	Title       string
	Description string
	// If true, hide the sidebar in the root template.
	HideSidebar bool
}

// User represents a registered user of the site.
type User struct {
	Login string
	Name  string
	Email string
	// Hashed password.
	Password string
}

// Session of an authenticated or anonymous user.
type Session struct {
	// Authenticaded user or nil
	User *User
	// Locale used for this session.
	Locale string
}

// A request to be processed by a worker.
type Request struct {
	// The requested node.
	Node Node
	// The query values of the request URL.
	Query url.Values
	// Method of the request (GET,POST,...).
	Method string
	// User session
	Session Session
	// Action to perform (e.g. "edit").
	Action string
}

// Response to a node request.
type Response struct {
	// The html content to be embedded in the root template.
	Body []byte
	// If set, redirect to this target using error 303 'see other'.
	Redirect string
	// The node as received by GetRequest, possibly with some fields
	// updated (e.g. modified title).
	//
	// If nil, the original node data is used.
	Node *Node
}

// Write appends the given bytes to the body of the response.
func (r *Response) Write(p []byte) (n int, err error) {
	r.Body = append(r.Body, p...)
	return len(p), nil
}

// Connection represents the rpc connection.
type Connection struct {
	cli *rpc.Client
}

// Bidirectional pipe used for RPC communication.
type pipeConnection struct {
	io.ReadCloser
	io.WriteCloser
}

func (pipe pipeConnection) Close() (err error) {
	panic("client: Close() on pipe connection. RPC error?")
}

// NewConnection establishes a new rpc connection and registers the content
// type.
func NewConnection(nodeType string) Connection {
	pipe := &pipeConnection{os.Stdin, os.Stdout}
	client := rpc.NewClient(pipe)
	return Connection{client}
}

// GetFormData retrieves form data of the request, i.e. query string values and
// possibly form data of POST and PUT requests.
func (c Connection) GetFormData() url.Values {
	var reply url.Values
	err := c.cli.Call("NodeRPC.GetFormData", 0, &reply)
	if err != nil {
		log.Fatal("client: RPC GetFormData error:", err)
	}
	return reply
}

// GetRequests blocks until it receives a request to process.
func (c Connection) GetRequest() Request {
	var reply Request
	err := c.cli.Call("NodeRPC.GetRequest", 0, &reply)
	if err != nil {
		log.Fatal("client: RPC GetRequest error:", err)
	}
	return reply
}

// GetNodeData requests data from some node.
//
// If the data does not exist, return null length []byte.
func (c Connection) GetNodeData(path, file string) []byte {
	args := &types.GetNodeDataArgs{path, file}
	var reply []byte
	err := c.cli.Call("NodeRPC.GetNodeData", args, &reply)
	if err != nil {
		log.Fatal("client: RPC GetNodeData error:", err)
	}
	return reply
}

// WriteNodeData writes data for some node.
func (c Connection) WriteNodeData(path, file, content string) error {
	args := &types.WriteNodeDataArgs{path, file, content}
	return c.cli.Call("NodeRPC.WriteNodeData", args, new(int))
}

// UpdateNode saves changes to given node.
func (c Connection) UpdateNode(node Node) error {
	return c.cli.Call("NodeRPC.UpdateNode", node, new(int))
}

// Send given Mail.
//
// An empty From or To field will be filled with the site owner's name and
// address.
func (c Connection) SendMail(m mimemail.Mail) {
	var reply int
	err := c.cli.Call("NodeRPC.SendMail", m, &reply)
	if err != nil {
		log.Fatal("client: RPC SendMail error:", err)
	}
}

// SendResponse sends the repsonse of an earlier request.
func (c Connection) SendResponse(r *Response) {
	var reply int
	err := c.cli.Call("NodeRPC.SendResponse", r, &reply)
	if err != nil {
		log.Fatal("client: RPC SendResponse error:", err)
	}
}

// Handler is a function that handles requests fetched via
// Connection.GetRequest.
type Handler func(Request, *Response, Connection)

// Serve waits in a loop for incoming requests and processes them with the given
// get or post handler.
func (c Connection) Serve(handle Handler) {
	for {
		req := c.GetRequest()
		res := Response{}
		handle(req, &res, c)
		c.SendResponse(&res)
	}
}
