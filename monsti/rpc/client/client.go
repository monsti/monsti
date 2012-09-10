package client

import (
	"datenkarussell.de/monsti/rpc/types"
	"io"
	"log"
	"net/rpc"
	"os"
)

type Node struct {
	Path        string
	Type        string
	Title       string
	Description string

	// HideSidebar is true if the sidebar should be hidden on display.ñ
	HideSidebar bool
}

type Request struct {
	Node   Node
	Method string
}

type Response struct {
	Body []byte
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

// GetRequests blocks until it receives a request to process.
func (c Connection) GetRequest() Request {
	log.Println("Calling NodeRPC.GetRequest")
	var reply Request
	err := c.cli.Call("NodeRPC.GetRequest", 42, &reply)
	log.Printf("Reply %v", reply)
	if err != nil {
		log.Fatal("client: monsti.GetRequest error:", err)
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
		log.Fatal("client: monsti.GetNodeData error:", err)
	}
	log.Println("data:", string(reply))
	return reply
}

// SendResponse sends the repsonse of an earlier request.
func (c Connection) SendResponse(r *Response) {
	log.Println("Calling NodeRPC.SendResponse")
	var reply int
	log.Println(r)
	err := c.cli.Call("NodeRPC.SendResponse", r, &reply)
	if err != nil {
		log.Fatal("client: monsti.SendResponse error:", err)
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
