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

package main

import (
	mtest "github.com/monsti/util/testing"
	"path/filepath"
	"testing"
)

func TestLoadSettings(t *testing.T) {
	files := map[string]string{
		"/config/monsti.yaml": `
nodetypes: [Document]
listen: localhost:8080
directories:
  statics: ../foo
`,
		"/config/sites/example/site.yaml": `
title: "Monsti CMS Example Site"
hosts: ["localhost:8080"]
directories:
  data: ../../../bar
`}
	root, cleanup, err := mtest.CreateDirectoryTree(files, "TestLoadSettings")
	if err != nil {
		t.Fatalf("Could not create test files: %v", err)
	}
	defer cleanup()
	settings, err := loadSettings(filepath.Join(root, "config"))
	if err != nil {
		t.Fatalf("Could not load test settings: %v", err)
	}
	if settings.Listen != "localhost:8080" {
		t.Errorf("settings.Listen == %v, should be \"localhost:8080\"",
			settings.Listen)
	}
	if len(settings.NodeTypes) != 1 || settings.NodeTypes[0] != "Document" {
		t.Errorf("settings.NodeTypes == %v, should be [Document]",
			settings.NodeTypes)
	}
	if settings.Directories.Statics != filepath.Join(root, "/foo") {
		t.Errorf(`Statics directory should be %v, but is %v`,
			filepath.Join(root, "/foo"),
			settings.Directories.Statics)
	}
	if _, ok := settings.Sites["example"]; len(settings.Sites) != 1 || !ok {
		t.Fatalf("settings.Sites should consist of one key example, but is %v",
			settings.Sites)
	}
	entry := settings.Sites["example"]
	if entry.Title != "Monsti CMS Example Site" {
		t.Errorf("settings.Sites[\"Example\"] should be "+
			`"Monsti CMS Example Site", but is %q`, entry.Title)
	}
	if len(entry.Hosts) != 1 || entry.Hosts[0] != "localhost:8080" {
		entry := settings.Sites["example"]
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
		if entry.Directories.Data != filepath.Join(root, "/bar") {
			t.Errorf(`Site's data directory should be %v, but is %v`,
				filepath.Join(root, "/bar"),
				entry.Directories.Data)
		}
	}
}
