package main

import (
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/rpc/types"
	"errors"
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
		panic("monsti: Could not send email: " + err.Error())
	}
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

// workerLog is a Writer used to log incoming worker errors.
type workerLog struct {
	Type string
}

func (w workerLog) Write(p []byte) (int, error) {
	log.Printf("from %v: %v", w.Type, string(p))
	return len(p), nil
}

func listenForRPC(settings settings, tickets chan ticket, nodeType string) {
	cmd := exec.Command(strings.ToLower(nodeType), settings.Directories.Config)
	inPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic("monsti: Could not setup stdout pipe of worker: " + err.Error())
	}
	outPipe, err := cmd.StdinPipe()
	if err != nil {
		panic("monsti: Could not setup stdin pipe of worker: " + err.Error())
	}
	cmd.Stderr = workerLog{nodeType}
	pipe := pipeConnection{inPipe, outPipe}
	log.Println("Starting worker for " + nodeType)
	err = cmd.Start()
	if err != nil {
		panic("monsti: Could not setup connection to worker: " + err.Error())
	}
	server := rpc.NewServer()
	nodeRPC := NodeRPC{
		Settings: settings,
		Tickets:  tickets}
	server.Register(&nodeRPC)
	log.Println("Setting up RPC.")
	server.ServeConn(pipe)
}
