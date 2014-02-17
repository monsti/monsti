// This file is part of monsti/util.
// Copyright 2012-2013 Christian Neumann

// monsti/util is free software: you can redistribute it and/or modify it under
// the terms of the GNU Lesser General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.

// monsti/util is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License for more
// details.

// You should have received a copy of the GNU Lesser General Public License
// along with monsti/util. If not, see <http://www.gnu.org/licenses/>.

package l10n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUseCatalog(t *testing.T) {
	// Setup catalogs
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get working directory.")
	}
	Setup("monsti-contactform", filepath.Join(pwd, "test_locale"))
	f := UseCatalog("de")

	// Test translation
	translated := f("Subject")
	if translated != "Betreff" {
		t.Errorf(`Translation of "Subject" should be "Betreff", got %q`,
			translated)
	}
}
