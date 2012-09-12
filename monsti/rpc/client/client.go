package client

import (
	"datenkarussell.de/monsti/rpc/types"
	"io"
	"log"
	"net/rpc"
        "net/url"
	"os"
)

type Node struct {
	Path        string
        // Content type of the node.
	Type        string
	Title       string
	Description string
        // If true, hide the sidebar in the root template.
        HideSidebar bool
}

// A request to be processed by a worker.
type Request struct {
        // The requested node.
	Node   Node
        // The query values of the request URL.
        Query  url.Values
        // Method of the request (GET,POST,...).
	Method string
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
	log.Println("Setting up RPC...")
	client := rpc.NewClient(pipe)
	log.Println("Connection established.")
	return Connection{client}
}

// GetFormData retrieves form data of the request, i.e. query string values and
// possibly form data of POST and PUT requests.
func (c Connection) GetFormData() url.Values {
	log.Println("Calling NodeRPC.GetFormData")
	var reply url.Values
	log.Println("call in")
	err := c.cli.Call("NodeRPC.GetFormData", 0, &reply)
	log.Println("call out")
	if err != nil {
		log.Fatal("client: RPC GetFormData error:", err)
	}
	return reply
}

// GetRequests blocks until it receives a request to process.
func (c Connection) GetRequest() Request {
	log.Println("Calling NodeRPC.GetRequest")
	var reply Request
	err := c.cli.Call("NodeRPC.GetRequest", 0, &reply)
	if err != nil {
		log.Fatal("client: RPC GetRequest error:", err)
	}
	return reply
}

// GetNodeData requests data from some node.
func (c Connection) GetNodeData(path, file string) []byte {
	log.Println("Calling NodeRPC.GetNodeData ", path, file)
	args := &types.GetNodeDataArgs{path, file}
	var reply []byte
	err := c.cli.Call("NodeRPC.GetNodeData", args, &reply)
	if err != nil {
		log.Fatal("client: RPC GetNodeData error:", err)
	}
	return reply
}

// SendResponse sends the repsonse of an earlier request.
func (c Connection) SendResponse(r *Response) {
	log.Println("Calling NodeRPC.SendResponse")
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
func (c Connection) Serve(get, post Handler) {
	for {
		log.Println("Requesting some work...")
		req := c.GetRequest()
		res := Response{}
		log.Println("Got some work!")
		switch req.Method {
		case "GET":
			get(req, &res, c)
		case "POST":
			post(req, &res, c)
		default:
			panic("client: Unknown method: " + req.Method)
		}
		c.SendResponse(&res)
		log.Println("Processed request.")
	}
}
