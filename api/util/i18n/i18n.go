// This file is part of Monsti.
// Copyright 2012 Christian Neumann

// Monsti is free software: you can redistribute it and/or modify it under
// the terms of the GNU Lesser General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.

// Monsti is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more
// details.

// You should have received a copy of the GNU Lesser General Public License
// along with Monsti. If not, see <http://www.gnu.org/licenses/>.

/*
Package i18n implements utility functions for working with translations.
*/
package i18n

import "pkg.monsti.org/gettext"

// LanguageMap maps locales to translation strings.
type LanguageMap map[string]string

// Get returns the translation for the given locale. If the
// translation is not set, it returns the translation for the empty
// locale, i.e. "".
func (l LanguageMap) Get(locale string) string {
	if v, ok := l[locale]; ok {
		return v
	}
	return l[""]
}

// GenLanguageMap generates a language map for the given locales.
//
// The map will have an entry for each locale with the locale id being
// the key and the value being the gettext translation of msg (using
// pkg.monsti.org/gettext.DefaultLocales). The empty locale, i.e. "",
// is set to msg.
func GenLanguageMap(msg string, locales []string) LanguageMap {
	ret := make(LanguageMap)
	ret[""] = msg
	for _, lang := range locales {
		G, _, _, _ := gettext.DefaultLocales.Use("", lang)
		ret[lang] = G(msg)
	}
	return ret
}
