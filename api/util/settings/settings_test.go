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

package settings

import (
	"os"
	"path/filepath"
	"testing"

	mtest "pkg.monsti.org/monsti/api/util/testing"
)

func TestLoadSiteSettings(t *testing.T) {
	files := map[string]string{
		"/etc/sites/example/site.yaml": `
title: "Monsti CMS Example Site"
hosts: ["localhost:8080"]
directories:
  data: ../../../bar
`,
		"/linked_site/site.yaml": `
title: "Linked Example Site"
`}
	root, cleanup, err := mtest.CreateDirectoryTree(files, "TestLoadSettings")
	if err != nil {
		t.Fatalf("Could not create test files: %v", err)
	}
	defer cleanup()
	err = os.Symlink(filepath.Join(root, "linked_site"),
		filepath.Join(root, "etc", "sites", "linked"))
	if err != nil {
		t.Fatalf("Could not create symlink to site config: %v", err)
	}
	sites, err := loadSiteSettings(filepath.Join(root, "etc", "sites"))
	if err != nil {
		t.Fatalf("Could not load site settings: %v", err)
	}
	if _, ok := sites["example"]; len(sites) != 2 || !ok {
		t.Fatalf("Should find two sites, but found %v", len(sites))
	}
	entry := sites["example"]
	if entry.Title != "Monsti CMS Example Site" {
		t.Errorf("settings.Sites[\"Example\"] should be "+
			`"Monsti CMS Example Site", but is %q`, entry.Title)
	}
	if entry.Locale != "en" {
		t.Errorf(`Default locale is not "en"`)
	}
	if len(entry.Hosts) != 1 || entry.Hosts[0] != "localhost:8080" {
		entry := sites["example"]
		if entry.Title != "Monsti CMS Example Site" {
			t.Errorf("settings.Sites[\"Example\"] should be "+
				`"Monsti CMS Example Site", but is %q`, entry.Title)
		}
		if len(entry.Hosts) != 1 || entry.Hosts[0] != "localhost:8080" {
			t.Errorf(`settings.Sites["example"].Hosts == %v, should be`+
				` ["localhost:8080"]`, entry.Hosts)
		}
		t.Errorf(`settings.Sites["example"].Hosts == %v, should be`+
			` ["localhost:8080"]`, entry.Hosts)
	}
}
