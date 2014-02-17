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

/*
 Monsti is a simple and resource efficient CMS.

 This package implements the mail service.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"sync"

	"github.com/chrneumann/mimemail"
	"pkg.monsti.org/service"
	"pkg.monsti.org/util"
)

type settings struct {
	Monsti   util.MonstiSettings
	Host     string
	Username string
	Password string
}

// MailService implements RPC methods for the Mail service.
type MailService struct {
	Info     *service.InfoClient
	Settings settings
}

func (m *MailService) SendMail(mail mimemail.Mail, reply *int) error {
	auth := smtp.PlainAuth("", m.Settings.Username,
		m.Settings.Password, strings.Split(m.Settings.Host, ":")[0])
	if err := smtp.SendMail(m.Settings.Host, auth,
		mail.Sender(), mail.Recipients(), mail.Message()); err != nil {
		return fmt.Errorf("monsti: Could not send email: %v", err)
	}
	return nil
}

func main() {
	logger := log.New(os.Stderr, "mail ", log.LstdFlags)

	// Load configuration
	flag.Parse()
	cfgPath := util.GetConfigPath(flag.Arg(0))
	var settings settings
	if err := util.LoadModuleSettings("mail", cfgPath, &settings); err != nil {
		logger.Fatal("Could not load settings: ", err)
	}

	// Connect to Info service
	info, err := service.NewInfoConnection(settings.Monsti.GetServicePath(
		service.Info.String()))
	if err != nil {
		logger.Fatalf("Could not connect to Info service: %v", err)
	}

	// Start own Mail service
	var waitGroup sync.WaitGroup
	logger.Println("Starting Mail service")
	waitGroup.Add(1)
	mailPath := settings.Monsti.GetServicePath(service.Mail.String())
	go func() {
		defer waitGroup.Done()
		var provider service.Provider
		var mail_ MailService
		mail_.Info = info
		mail_.Settings = settings
		provider.Logger = logger
		if err := provider.Serve(mailPath, "Mail", &mail_); err != nil {
			logger.Fatalf("Could not start Mail service: %v", err)
		}
	}()

	if err := info.PublishService("Mail", mailPath); err != nil {
		logger.Fatalf("Could not publish Mail service: %v", err)
	}

	waitGroup.Wait()
}
