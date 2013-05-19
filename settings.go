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
	"fmt"
	"github.com/monsti/util"
	"io/ioutil"
	"path/filepath"
)

// Site configuration.
type site struct {
	// Name of the site for internal use.
	Name string
	// Title as used in HTML head.
	Title string
	// The hosts which should deliver this site.
	Hosts []string
	// Name and email address of site owner.
	//
	// The owner's address is used as recipient of contact form submissions.
	Owner struct {
		Name, Email string
	}
	// Key to authenticate session cookies.
	SessionAuthKey string
	// Locale used to translate monsti's web interface.
	Locale string
}

// Settings for the application and the sites.
type settings struct {
	Monsti util.MonstiSettings
	// Listen is the host and port to listen for incoming HTTP connections.
	Listen string
	// Sites hosted by this monsti instance.
	Sites map[string]site
}

// LoadSiteSettings loads the configurated sites' settings.
func (s *settings) LoadSiteSettings() error {
	// Load site specific configuration files
	sitesPath := filepath.Join(s.Monsti.Directories.Config, "sites")
	siteDirs, err := ioutil.ReadDir(sitesPath)
	if err != nil {
		return fmt.Errorf("Could not read sites directory: %v", err)
	}
	s.Sites = make(map[string]site)
	for _, siteDir := range siteDirs {
		if !siteDir.IsDir() {
			continue
		}
		siteName := siteDir.Name()
		sitePath := filepath.Join(sitesPath, siteName)
		var siteSettings site
		err := util.ParseYAML(filepath.Join(sitePath, "site.yaml"),
			&siteSettings)
		if err != nil {
			return fmt.Errorf("Could not load settings for site %q: %v",
				siteName, err)
		}
		s.Sites[siteName] = siteSettings
	}
	return nil
}
