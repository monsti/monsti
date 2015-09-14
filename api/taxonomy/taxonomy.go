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

package taxonomy

import "pkg.monsti.org/monsti/api/service"

var availableLocales = []string{"de", "en", "nl"}

// FindNodesByTerm retrieves nodes which have the given vocabulary's
// term assigned.
func FindNodesByTerm(m *service.MonstiClient, vocabulary,
	term string) ([]*service.Node, error) {
	/*
		fields, err := m.GetFieldsByClass("term-field")
		if err != nil {
			return nil, fmt.Errorf("taxonomy: Could not retrieve term fields")
		}
		for _, field := range fields {
			// search fields where vocabulary matches
			if vocabulary == field.Data["core.Vocabulary"]
			// ...

		}
	*/
	/* TODO
	search nodes having this field with the given term
	return nodes
	*/
	return nil, nil
}
