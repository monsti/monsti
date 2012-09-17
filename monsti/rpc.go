package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/rpc/types"
	"errors"
	"fmt"
	"github.com/chrneumann/mimemail"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"net/smtp"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

// ticket represents an incoming request to be processed by the workers.
type ticket struct {
	Node    client.Node
	Request *http.Request
	// The request handler waits until the response gets over this channel.
	ResponseChan chan client.Response
}

// NodeRPC provides RPC methods for workers.
//
// See monsti/rpc/client for more documentation on the metods.
type NodeRPC struct {
	Settings settings
	Ticket   *ticket
	Tickets  chan ticket
}

func (m *NodeRPC) GetNodeData(args *types.GetNodeDataArgs, reply *[]byte) error {
	log.Println("RPC: GetNodeData")
	path := filepath.Join(m.Settings.Directories.Data, args.Path[1:], args.File)
	ret, err := ioutil.ReadFile(path)
	*reply = ret
	return err
}

func (m *NodeRPC) GetFormData(arg int, reply *url.Values) error {
	log.Println("RPC: GetFormData")
	err := m.Ticket.Request.ParseForm()
	if err != nil {
		return err
	}
	*reply = m.Ticket.Request.Form
	return nil
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
		Node:   m.Ticket.Node,
		Query:  m.Ticket.Request.URL.Query()}
	log.Println("Got ticket, sent to worker.")
	return nil
}

func (m *NodeRPC) SendMail(mail mimemail.Mail, reply *int) error {
	log.Println("RPC: SendMail")
	owner := mimemail.Address{m.Settings.Owner.Name, m.Settings.Owner.Email}
	if len(mail.From.Email) == 0 {
		mail.From = owner
	}
	if mail.To == nil {
		mail.To = []mimemail.Address{owner}
	}
	auth := smtp.PlainAuth("", m.Settings.Mail.Username,
		m.Settings.Mail.Password, strings.Split(m.Settings.Mail.Host, ":")[0])
	if err := smtp.SendMail(m.Settings.Mail.Host, auth,
		mail.Sender(), mail.Recipients(), mail.Message()); err != nil {
		log.Println("monsti: Could not send email: " + err.Error())
		return fmt.Errorf("Could not send email.")
	}
	return nil
}

func (m *NodeRPC) SendResponse(res client.Response, reply *int) error {
	log.Println("RPC: SendResponse")
	m.Ticket.ResponseChan <- res
	m.Ticket = nil
	return nil
}

// pipeConnection is a bidirectional pipe to a worker process.
type pipeConnection struct {
	io.ReadCloser
	io.WriteCloser
}

func (pipe pipeConnection) Close() (err error) {
	panic("monsti: Close() on pipe connection. RPC error?")
}

// workerLog is a Writer used to log incoming worker errors.
type workerLog struct {
	Type string
}

func (w workerLog) Write(p []byte) (int, error) {
	log.Printf("from %v: %v", w.Type, string(p))
	return len(p), nil
}

// listenForRPC starts a worker process and listens for RPC request on a pipe to
// the process.
func listenForRPC(settings settings, tickets chan ticket, nodeType string) error {
	cmd := exec.Command(strings.ToLower(nodeType), settings.Directories.Config)
	inPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Could not setup stdout pipe of worker: %v",
			err.Error())
	}
	outPipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Could not setup stdin pipe of worker: %v",
			err.Error())
	}
	cmd.Stderr = workerLog{nodeType}
	pipe := pipeConnection{inPipe, outPipe}
	log.Println("Starting worker for " + nodeType)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Could not setup connection to worker: %v",
			err.Error())
	}
	server := rpc.NewServer()
	nodeRPC := NodeRPC{
		Settings: settings,
		Tickets:  tickets}
	server.Register(&nodeRPC)
	log.Println("Setting up RPC.")
	go server.ServeConn(pipe)
	return nil
}
