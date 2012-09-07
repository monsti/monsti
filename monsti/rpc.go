package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/rpc/types"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"path/filepath"
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
	log.Println("%v", args)
	path := filepath.Join(m.Settings.Root, args.Path[1:], args.File)
	ret, err := ioutil.ReadFile(path)
	*reply = ret
	return err
}

func (m *NodeRPC) GetRequest(arg int, reply *client.Request) error {
	log.Println("RPC: GetRequest")
	log.Printf("RPC: GetRequest arg: %v \n", arg)
	if m.Ticket != nil {
		return errors.New("monsti: Still waiting for response to last request.")
	}
	m.Ticket = new(ticket)
	log.Println("Wating for ticket to send to worker.")
	(*m.Ticket) = <-m.Tickets
	log.Printf("%v", m.Ticket)
	*reply = client.Request{
		Method: m.Ticket.Request.Method,
		Node:   m.Ticket.Node}
	log.Printf("Reply %v", reply)
	log.Println("Got ticket, sent to worker.")
	return nil
}

func (m *NodeRPC) SendResponse(res client.Response, reply *int) error {
	log.Println("RPC: SendResponse")
	m.Ticket.ResponseChan <- res
	m.Ticket = nil
	return nil
}

func listenForRPC(tickets chan ticket) {
	ln, err := net.Listen("tcp", ":12345")
	log.Println("Waiting for worker on :12345.")
	if err != nil {
		panic("monsti: Could not setup connection to worker: " + err.Error())
	}
	conn, err := ln.Accept()
	log.Println("Connection to worker established.")
	if err != nil {
		panic("monsti: RPC setup error: " + err.Error())
	}
	server := rpc.NewServer()
	nodeRPC := NodeRPC{
		Tickets: tickets}
	server.Register(&nodeRPC)
	log.Println("Setting up RPC.")
	server.ServeConn(conn)
}
