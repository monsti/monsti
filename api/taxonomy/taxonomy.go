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

import (
	"pkg.monsti.org/monsti/api/service"
	"pkg.monsti.org/monsti/api/util/i18n"
)

var availableLocales = []string{"de", "en", "nl"}

// TermsFieldSettings holds the settings of a terms field that allows
// to specify terms.
type TermsFieldSettings struct {
	// Vocabulary is the path to the taxonomy's vocabulary.
	Vocabulary string
}

// FieldType returns a field type to be used in node and field type
// definitions.
func (f TermsFieldSettings) FieldType() *service.ListFieldType {
	G := func(in string) string { return in }
	return &service.ListFieldType{
		ElementType: new(service.TextFieldType),
		AddLabel:    i18n.GenLanguageMap(G("Add term"), availableLocales),
		RemoveLabel: i18n.GenLanguageMap(G("Remove term"), availableLocales),
		Classes:     []string{"terms-field"},
	}
}
