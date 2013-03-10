// This file is part of monsti/monsti-daemon.
// Copyright 2012-2013 Christian Neumann

// monsti/monsti-daemon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or (at your
// option) any later version.

// monsti/monsti-daemon is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License
// for more details.

// You should have received a copy of the GNU Lesser General Public License
// along with monsti/monsti-daemon. If not, see <http://www.gnu.org/licenses/>.

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
	// Absolute paths to site specific directories.
	Directories struct {
		// Configuration directory
		Config string
		// Site content
		Data string
		// Site specific static files
		Statics string
		// HTML Templates to be used instead of monsti's ones.
		Templates string
	}
}

// Settings for the application and the sites.
type settings struct {
	// Settings for sending mail (outgoing SMTP).
	Mail struct {
		// Host may be specified as address:port
		Host, Username, Password string
	}
	// Listen is the host and port to listen for incoming HTTP connections.
	Listen string
	// Absolute paths to used directories.
	Directories struct {
		// Config files
		Config string
		// Monsti's static files
		Statics string
		// HTML Templates
		Templates string
		// Locales, i.e. the gettext machine objects (.mo)
		Locales string
	}
	// List of node types to be activated.
	NodeTypes []string
	// Sites hosted by this monsti instance.
	Sites map[string]site
}

// loadSettings loads daemon and site settings from the given configuration
// directory.
//
// The configuration directory path must be absolute or relative to the working
// directory.
func loadSettings(cfgPath string) (*settings, error) {
	// Load main configuration file
	settings := new(settings)
	err := util.ParseYAML(filepath.Join(cfgPath, "monsti.yaml"), settings)
	if err != nil {
		return nil, fmt.Errorf("Could not load main configuration file: %v", err)
	}
	settings.Directories.Config = cfgPath
	util.MakeAbsolute(&settings.Directories.Statics, cfgPath)
	util.MakeAbsolute(&settings.Directories.Templates, cfgPath)
	util.MakeAbsolute(&settings.Directories.Locales, cfgPath)

	// Load site specific configuration files
	sitesPath := filepath.Join(settings.Directories.Config, "sites")
	siteDirs, err := ioutil.ReadDir(sitesPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read sites directory: %v", err)
	}
	settings.Sites = make(map[string]site)
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
			return nil, fmt.Errorf("Could not load settings for site %q: %v",
				siteName, err)
		}
		siteSettings.Directories.Config = sitePath
		util.MakeAbsolute(&siteSettings.Directories.Config, sitePath)
		util.MakeAbsolute(&siteSettings.Directories.Data, sitePath)
		util.MakeAbsolute(&siteSettings.Directories.Statics, sitePath)
		util.MakeAbsolute(&siteSettings.Directories.Templates, sitePath)
		settings.Sites[siteName] = siteSettings
	}
	return settings, nil
}
