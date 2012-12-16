// This file is part of monsti/l10n.
// Copyright 2012 Christian Neumann

// monsti/l10n is free software: you can redistribute it and/or modify it under
// the terms of the GNU Lesser General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.

// monsti/l10n is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more
// details.

// You should have received a copy of the GNU Lesser General Public License
// along with monsti/l10n. If not, see <http://www.gnu.org/licenses/>.

/*
 * This package provides localization services for monsti and monsti
 * content type workers.
 */
package l10n

import (
	"github.com/chrneumann/g5t"
)

// Catalogs maps locales to string catalogs.
var Catalogs map[string]g5t.CatalogType

// DefaultCatalogs is the default collection of catalogs.
var DefaultCatalogs = make(map[string]g5t.CatalogType)

// Settings for catalog retrieval.
type Settings struct {
	Domain string
	Directory string
}

var DefaultSettings Settings

// UseCatalog returns l10n functions for the given locale and the default
// catalogs.
//
// If no catalog exists for the given locale, return l10n functions which just
// return the untouched translation strings.
func UseCatalog(locale string) (func (string) string) {
	null_func := func(x string) string { return x }
	if len(locale) == 0 {
		return null_func
	}
	catalog, ok := DefaultCatalogs[locale]
	if !ok {
		catalog, err := g5t.LoadLang(DefaultSettings.Domain,
			DefaultSettings.Directory, locale, g5t.GettextParser)
		if err != nil {
			return null_func
		}
		DefaultCatalogs[locale] = catalog
	}
	return catalog.StringPartial()
}

// Setup sets up the default catalogs.
func Setup(domain, localesPath string) {
	DefaultSettings.Domain = domain
	DefaultSettings.Directory = localesPath
}
