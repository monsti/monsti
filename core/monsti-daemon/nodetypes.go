// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann <cneumann@datenkarussell.de>
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
	"pkg.monsti.org/monsti/api/util/i18n"
)
import "pkg.monsti.org/monsti/api/service"

var availableLocales = []string{"en", "de", "nl"}

func initNodeTypes(settings *settings, session *service.Session, logger *log.Logger) error {
	G := func(in string) string { return in }
	pathType := service.NodeType{
		Id:   "core.Path",
		Hide: true,
		Name: i18n.GenLanguageMap(G("Path"), availableLocales),
	}
	if err := session.Monsti().RegisterNodeType(&pathType); err != nil {
		return fmt.Errorf("Could not register path node type: %v", err)
	}

	documentType := service.NodeType{
		Id:        "core.Document",
		AddableTo: []string{"."},
		Name:      i18n.GenLanguageMap(G("Document"), availableLocales),
		Fields: []*service.FieldConfig{
			{
				Id:       "core.Title",
				Required: true,
				Name:     i18n.GenLanguageMap(G("Title"), availableLocales),
				Type:     new(service.TextFieldType),
			},
			{
				Id:   "core.Description",
				Name: i18n.GenLanguageMap(G("Description"), availableLocales),
				Type: new(service.TextFieldType),
			},
			{
				Id:   "core.Thumbnail",
				Name: i18n.GenLanguageMap(G("Thumbnail"), availableLocales),
				Type: new(service.RefFieldType),
			},
			{
				Id:       "core.Body",
				Required: true,
				Name:     i18n.GenLanguageMap(G("Body"), availableLocales),
				Type:     new(service.HTMLFieldType),
			},
			{
				Id:      "core.Categories",
				Hidden:  true,
				Name:    i18n.GenLanguageMap(G("Categories"), availableLocales),
				Classes: []string{"categories"},
				Type: &service.MapFieldType{
					ElementType: &service.ListFieldType{
						ElementType: new(service.TextFieldType),
						AddLabel:    i18n.GenLanguageMap(G("Add term"), availableLocales),
						RemoveLabel: i18n.GenLanguageMap(G("Remove term"), availableLocales),
					},
				},
			},
		},
	}
	if err := session.Monsti().RegisterNodeType(&documentType); err != nil {
		return fmt.Errorf("Could not register document node type: %v", err)
	}

	fileType := service.NodeType{
		Id:        "core.File",
		AddableTo: []string{"."},
		Name:      i18n.GenLanguageMap(G("File"), availableLocales),
		Fields: []*service.FieldConfig{
			{Id: "core.Title"},
			{
				Id:       "core.File",
				Required: true,
				Name:     i18n.GenLanguageMap(G("File"), availableLocales),
				Type:     new(service.FileFieldType),
			},
		},
	}
	if err := session.Monsti().RegisterNodeType(&fileType); err != nil {
		return fmt.Errorf("Could not register file node type: %v", err)
	}

	imageType := service.NodeType{
		Id:        "core.Image",
		Hide:      true,
		AddableTo: []string{"."},
		Name:      i18n.GenLanguageMap(G("Image"), availableLocales),
		Fields: []*service.FieldConfig{
			{Id: "core.Title"},
			{Id: "core.File"},
		},
	}
	if err := session.Monsti().RegisterNodeType(&imageType); err != nil {
		return fmt.Errorf("Could not register image node type: %v", err)
	}

	return nil
}
