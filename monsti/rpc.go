package main

import (
	"code.google.com/p/gorilla/sessions"
	"datenkarussell.de/monsti/rpc/client"
	"datenkarussell.de/monsti/rpc/types"
	"datenkarussell.de/monsti/worker"
	"errors"
	"fmt"
	"github.com/chrneumann/mimemail"
	"io/ioutil"
	"log"
	"net/smtp"
	"net/url"
	"path/filepath"
	"strings"
)

// NodeRPC provides RPC methods for workers.
//
// See monsti/rpc/client for more documentation on the metods.
type NodeRPC struct {
	Worker   *worker.Worker
	Settings settings
	Session  *sessions.Session
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
	err := m.Worker.Ticket.Request.ParseForm()
	if err != nil {
		return err
	}
	*reply = m.Worker.Ticket.Request.Form
	return nil
}

func (m *NodeRPC) GetRequest(arg int, reply *client.Request) error {
	log.Println("RPC: GetRequest")
	if m.Worker.Ticket != nil {
		return errors.New("monsti: Still waiting for response to last request.")
	}
	log.Println("Wating for ticket to send to worker.")
	ticket := <-m.Worker.Tickets
	m.Worker.Ticket = &ticket
	request := client.Request{
		Method:  m.Worker.Ticket.Request.Method,
		Node:    m.Worker.Ticket.Node,
		Query:   m.Worker.Ticket.Request.URL.Query(),
		Session: m.Worker.Ticket.Session}
	*reply = request
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
	m.Worker.Ticket.ResponseChan <- res
	m.Worker.Ticket = nil
	return nil
}
