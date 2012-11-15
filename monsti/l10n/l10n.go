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
func UseCatalog(locale string) (func (string) string) {
	catalog, ok := DefaultCatalogs[locale]
	if !ok {
		catalog, err := g5t.LoadLang(DefaultSettings.Domain,
			DefaultSettings.Directory, locale, g5t.GettextParser)
		if err != nil {
			panic("Could not load catalog for locale " + locale)
		}
		DefaultCatalogs[locale] = catalog
	}
	return catalog.StringPartial()
}
