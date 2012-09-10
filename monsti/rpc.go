package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/rpc/types"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"os/exec"
	"path/filepath"
	"strings"
)

// ticket represents an incoming request to be processed by the workers.
type ticket struct {
	// The request handlers waits until some value gets over this channel.
	Node         client.Node
	Request      *http.Request
	ResponseChan chan client.Response
}

type NodeRPC struct {
	Settings settings
	Ticket   *ticket
	Tickets  chan ticket
}

func (m *NodeRPC) GetNodeData(args *types.GetNodeDataArgs, reply *[]byte) error {
	log.Println("RPC: GetNodeData")
	path := filepath.Join(m.Settings.Root, args.Path[1:], args.File)
	ret, err := ioutil.ReadFile(path)
	*reply = ret
	return err
}

func (m *NodeRPC) GetRequest(arg int, reply *client.Request) error {
	log.Println("RPC: GetRequest")
	if m.Ticket != nil {
		return errors.New("monsti: Still waiting for response to last request.")
	}
	m.Ticket = new(ticket)
	log.Println("Wating for ticket to send to worker.")
	(*m.Ticket) = <-m.Tickets
	*reply = client.Request{
		Method: m.Ticket.Request.Method,
		Node:   m.Ticket.Node}
	log.Println("Got ticket, sent to worker.")
	return nil
}

func (m *NodeRPC) SendResponse(res client.Response, reply *int) error {
	log.Println("RPC: SendResponse")
	m.Ticket.ResponseChan <- res
	m.Ticket = nil
	return nil
}

type pipeConnection struct {
	io.ReadCloser
	io.WriteCloser
}

func (pipe pipeConnection) Close() (err error) {
	panic("monsti: Close() on pipe connection. RPC error?")
}

func listenForRPC(tickets chan ticket, nodeType string) {
	cmd := exec.Command(strings.ToLower(nodeType))
	inPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic("monsti: Could not setup stdout pipe of worker: " + err.Error())
	}
	outPipe, err := cmd.StdinPipe()
	if err != nil {
		panic("monsti: Could not setup stdin pipe of worker: " + err.Error())
	}
	pipe := pipeConnection{inPipe, outPipe}
	log.Println("Starting worker for " + nodeType)
	err = cmd.Start()
	if err != nil {
		panic("monsti: Could not setup connection to worker: " + err.Error())
	}
	server := rpc.NewServer()
	nodeRPC := NodeRPC{
		Tickets: tickets}
	server.Register(&nodeRPC)
	log.Println("Setting up RPC.")
	server.ServeConn(pipe)
}
