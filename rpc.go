package main

import (
	"errors"
	"fmt"
	"github.com/chrneumann/mimemail"
	"github.com/gorilla/sessions"
	"github.com/monsti/monsti-daemon/worker"
	"github.com/monsti/rpc/client"
	"github.com/monsti/rpc/types"
	"io/ioutil"
	"log"
	"net/smtp"
	"net/url"
	"os"
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
	Log      *log.Logger
}

func (m *NodeRPC) GetNodeData(args *types.GetNodeDataArgs, reply *[]byte) error {
	site := m.Settings.Sites[m.Worker.Ticket.Site]
	path := filepath.Join(site.Directories.Data, args.Path[1:], args.File)
	ret, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			*reply = make([]byte, 0)
			return nil
		}
	} else {
		*reply = ret
	}
	return err
}

func (m *NodeRPC) WriteNodeData(args *types.WriteNodeDataArgs,
	reply *int) error {
	site := m.Settings.Sites[m.Worker.Ticket.Site]
	path := filepath.Join(site.Directories.Data, args.Path[1:], args.File)
	err := ioutil.WriteFile(path, []byte(args.Content), 0600)
	return err
}

func (m *NodeRPC) GetFileData(key *string, reply *[]byte) error {
	if err := m.Worker.Ticket.Request.ParseMultipartForm(1024 * 1024); err != nil {
		return err
	}
	file, _, err := m.Worker.Ticket.Request.FormFile(*key)
	if err != nil {
		return err
	}
	defer file.Close()
	*reply, err = ioutil.ReadAll(file)
	return err
}

func (m *NodeRPC) GetFormData(arg int, reply *url.Values) error {
	err := m.Worker.Ticket.Request.ParseForm()
	if err != nil {
		return err
	}
	*reply = m.Worker.Ticket.Request.Form
	return nil
}

func (m *NodeRPC) GetRequest(arg int, reply *client.Request) error {
	if m.Worker.Ticket != nil {
		return errors.New("monsti: Still waiting for response to last request.")
	}
	ticket := <-m.Worker.Tickets
	m.Worker.Ticket = &ticket
	request := client.Request{
		Method:  m.Worker.Ticket.Request.Method,
		Node:    m.Worker.Ticket.Node,
		Query:   m.Worker.Ticket.Request.URL.Query(),
		Session: m.Worker.Ticket.Session,
		Action:  m.Worker.Ticket.Action}
	*reply = request
	return nil
}

func (m *NodeRPC) UpdateNode(node client.Node, reply *int) error {
	site := m.Settings.Sites[m.Worker.Ticket.Site]
	return writeNode(node, site.Directories.Data)
}

func (m *NodeRPC) SendMail(mail mimemail.Mail, reply *int) error {
	site := m.Settings.Sites[m.Worker.Ticket.Site]
	owner := mimemail.Address{site.Owner.Name, site.Owner.Email}
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
		m.Log.Println("monsti: Could not send email: " + err.Error())
		return fmt.Errorf("Could not send email.")
	}
	return nil
}

func (m *NodeRPC) SendResponse(res client.Response, reply *int) error {
	m.Worker.Ticket.ResponseChan <- res
	m.Worker.Ticket = nil
	return nil
}
