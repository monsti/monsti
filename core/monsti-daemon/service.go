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
	"net/smtp"
	"strings"
	"sync"

	"pkg.monsti.org/monsti/api/service"

	"github.com/chrneumann/mimemail"
)

type MonstiService struct {
	// Services maps service names to service paths
	Services map[string][]string
	// Mutex to syncronize data access
	mutex    sync.RWMutex
	Settings *settings
}

func (i *MonstiService) PublishService(args service.PublishServiceArgs,
	reply *int) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.Services == nil {
		i.Services = make(map[string][]string)
	}
	switch args.Service {
	case "Data":
	case "Mail":
	default:
		return fmt.Errorf("Unknown service type %v", args.Service)
	}

	if i.Services[args.Service] == nil {
		i.Services[args.Service] = make([]string, 0)
	}
	i.Services[args.Service] = append(i.Services[args.Service], args.Path)

	return nil
}

/*
func (i *MonstiService) FindDataService(arg int, path *string) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if len(i.Services["Data"]) == 0 {
		return fmt.Errorf("Could not find any data services")
	}
	*path = i.Services["Data"][0]
	return nil
}
*/

func (m *MonstiService) SendMail(mail mimemail.Mail, reply *int) error {
	auth := smtp.PlainAuth("", m.Settings.Mail.Username,
		m.Settings.Mail.Password, strings.Split(m.Settings.Mail.Host, ":")[0])
	if err := smtp.SendMail(m.Settings.Mail.Host, auth,
		mail.Sender(), mail.Recipients(), mail.Message()); err != nil {
		return fmt.Errorf("monsti: Could not send email: %v", err)
	}
	return nil
}
