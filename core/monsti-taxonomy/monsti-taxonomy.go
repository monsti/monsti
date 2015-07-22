// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann
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
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util/i18n"
	"pkg.monsti.org/monsti/api/util/module"
)

var availableLocales = []string{"de", "en"}

func setup(c *module.ModuleContext) error {
	gettext.DefaultLocales.Domain = "monsti-daemon"
	G := func(in string) string { return in }
	m := c.Session.Monsti()

	nodeType := &service.NodeType{
		Id:        "core.Vocabulary",
		AddableTo: []string{"."},
		Name:      i18n.GenLanguageMap(G("Vocabulary"), availableLocales),
		Fields: []*service.FieldConfig{
			{Id: "core.Title"},
			{Id: "core.Description"},
			{
				Id:     "core.VocabularyTerms",
				Hidden: true,
				Type: &service.ListFieldType{
					ElementType: &service.CombinedFieldType{map[string]service.FieldConfig{
						"Name":   {Type: new(service.TextFieldType)},
						"Title":  {Type: new(service.TextFieldType)},
						"Parent": {Type: new(service.TextFieldType)},
					},
					}}}}}
	if err := m.RegisterNodeType(nodeType); err != nil {
		c.Logger.Fatalf("Could not register %q node type: %v", nodeType.Id, err)
	}

	return nil
}

func main() {
	module.StartModule("taxonomy", setup)
}
