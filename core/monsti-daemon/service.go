// This file is part of Monsti, a web content management system.
// Copyright 2012-2013 Christian Neumann
//
// Monsti is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// Monsti is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Monsti.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"sync"

	"github.com/chrneumann/mimemail"
)

type subscription struct {
}

type signal struct {
	Name string
	Args interface{}
	Ret  chan interface{}
}

type MonstiService struct {
	// Services maps service names to service paths
	Services map[string][]string
	// Mutex to syncronize data access
	mutex         sync.RWMutex
	Settings      *settings
	Logger        *log.Logger
	subscriptions map[string][]string
	subscriber    map[string]chan *signal
	subscriberRet map[string]chan interface{}
}

type PublishServiceArgs struct {
	Service, Path string
}

func (i *MonstiService) PublishService(args PublishServiceArgs,
	reply *int) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.Services == nil {
		i.Services = make(map[string][]string)
	}
	switch args.Service {
	default:
		return fmt.Errorf("Unknown service type %v", args.Service)
	}

	if i.Services[args.Service] == nil {
		i.Services[args.Service] = make([]string, 0)
	}
	i.Services[args.Service] = append(i.Services[args.Service], args.Path)

	return nil
}

func (i *MonstiService) ModuleInitDone(args string, reply *int) error {
	// TODO Implement me
	return nil
}

func (m *MonstiService) SendMail(mail mimemail.Mail, reply *int) error {
	if !m.Settings.Mail.Debug {
		auth := smtp.PlainAuth("", m.Settings.Mail.Username,
			m.Settings.Mail.Password, strings.Split(m.Settings.Mail.Host, ":")[0])
		if err := smtp.SendMail(m.Settings.Mail.Host, auth,
			mail.Sender(), mail.Recipients(), mail.Message()); err != nil {
			return fmt.Errorf("monsti: Could not send email: %v", err)
		}
	} else {
		m.Logger.Printf(`SendMail debug:
From: %v
To: %v
Cc: %v
Bcc: %v
Subject: %v
-- Body Start --
%v
-- Body End --`,
			mail.From, mail.To, mail.Cc, mail.Bcc, mail.Subject, string(mail.Body))
	}
	return nil
}

type ConnectSignalArgs struct {
	Id, Signal string
}

func (m *MonstiService) ConnectSignal(args *ConnectSignalArgs, ret *int) error {
	log.Printf("ConnectSignal %v", args)
	if m.subscriptions == nil {
		m.subscriptions = make(map[string][]string)
		m.subscriber = make(map[string]chan *signal)
	}
	m.subscriptions[args.Signal] = append(m.subscriptions[args.Signal], args.Id)
	if _, ok := m.subscriber[args.Id]; !ok {
		m.subscriber[args.Id] = make(chan *signal)
	}
	return nil
}

type Ret struct {
	Ret interface{}
}

type Receive struct {
	Name string
	Args interface{}
}

func (m *MonstiService) EmitSignal(args *Receive, ret *Ret) error {
	log.Printf("daemon received EmitSignal with args: %v ret: %t ", args, ret)
	for _, id := range m.subscriptions[args.Name] {
		log.Printf("send to %v", id)
		retChan := make(chan interface{})
		m.subscriber[id] <- &signal{args.Name, args.Args, retChan}
		ret.Ret = <-retChan
	}
	return nil
}

type WaitSignalRet struct {
	Name string
	Args interface{}
}

func (m *MonstiService) WaitSignal(subscriber string, ret *WaitSignalRet) error {
	log.Printf("WaitSignal %v", subscriber)
	signal := <-m.subscriber[subscriber]
	ret.Name = signal.Name
	ret.Args = signal.Args
	if m.subscriberRet == nil {
		m.subscriberRet = make(map[string]chan interface{})
	}
	m.subscriberRet[subscriber] = signal.Ret
	return nil
}

type FinishSignalArgs struct {
	Id  string
	Ret interface{}
}

func (m *MonstiService) FinishSignal(args *FinishSignalArgs, _ *int) error {
	m.subscriberRet[args.Id] <- args.Ret
	return nil
}
