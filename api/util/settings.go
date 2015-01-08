package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Site configuration.
type SiteSettings struct {
	// Name of the site for internal use.
	Name string
	// Title as used in HTML head.
	Title string
	// The hosts which should deliver this site.
	Hosts []string
	// EmailName is used as name in the From header of outgoing site emails.
	// BaseURL is the URL to the root of the site. Used to generate
	// absolute URLs.
	BaseURL   string
	EmailName string
	// EmailAddress is used as address in the From header of outgoing site emails.
	EmailAddress string
	// Name and email address of site owner.
	//
	// The owner's address is used as recipient of contact form submissions.
	Owner struct {
		Name, Email string
	}
	// Key to authenticate session cookies.
	SessionAuthKey string
	// Key to authenticate password request tokens.
	PasswordTokenKey string
	// Locale used to translate monsti's web interface.
	Locale string
}

// MonstiSettings holds common Monsti settings.
type MonstiSettings struct {
	// Absolute paths to used directories.
	Directories struct {
		// Configuration directory
		Config string
		// Site data directory
		Data string
		// Shared data directory
		Share string
		// Locale directory
		Locale string
		// Runtime data directory
		Run string
	}
	// Sites hosted by this monsti instance.
	//
	// Load settings with *MonstiSettings.LoadSiteSettings()
	Sites map[string]SiteSettings
}

// GetServicePath returns the path to the given service's socket.
func (s MonstiSettings) GetServicePath(service string) string {
	return filepath.Join(s.Directories.Run, strings.ToLower(service)+".socket")
}

// GetSiteConfigPath returns the path to the given site's configuration
// directory.
func (s MonstiSettings) GetSiteConfigPath(site string) string {
	return filepath.Join(s.Directories.Config, "sites", site)
}

// GetSiteCachePath returns the path to the given site's cache directory.
func (s MonstiSettings) GetSiteCachePath(site string) string {
	return filepath.Join(s.Directories.Data, site, "cache")
}

// GetSiteNodesPath returns the path to the given site's node directory.
func (s MonstiSettings) GetSiteNodesPath(site string) string {
	return filepath.Join(s.Directories.Data, site, "nodes")
}

// GetSiteStaticsPath returns the path to the given site's site-static
// directory.
func (s MonstiSettings) GetSiteStaticsPath(site string) string {
	return filepath.Join(s.Directories.Data, site, "site-static")
}

// GetSiteDataPath returns the path to the given site's data
// directory.
func (s MonstiSettings) GetSiteDataPath(site string) string {
	return filepath.Join(s.Directories.Data, site)
}

// GetSiteTemplatesPath returns the path to the given site's templates
// directory.
func (s MonstiSettings) GetSiteTemplatesPath(site string) string {
	return filepath.Join(s.Directories.Data, site, "templates")
}

// GetStaticsPath returns the path to the global site-static directory.
func (s MonstiSettings) GetStaticsPath() string {
	return filepath.Join(s.Directories.Share, "static")
}

// GetTemplatesPath returns the path to the global templates directory.
func (s MonstiSettings) GetTemplatesPath() string {
	return filepath.Join(s.Directories.Share, "templates")
}

// loadSiteSettings returns the site settings in the given directory.
func loadSiteSettings(sitesDir string) (map[string]SiteSettings, error) {
	sitesPath := filepath.Join(sitesDir)
	siteDirs, err := ioutil.ReadDir(sitesPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read sites directory: %v", err)
	}
	sites := make(map[string]SiteSettings)
	for _, siteDir := range siteDirs {
		siteName := siteDir.Name()
		sitePath := filepath.Join(sitesPath, siteName)
		_, err := os.Stat(filepath.Join(sitePath, "site.yaml"))
		if err != nil {
			log.Println(err)
			continue
		}
		var siteSettings SiteSettings
		err = ParseYAML(filepath.Join(sitePath, "site.yaml"),
			&siteSettings)
		if err != nil {
			return nil, fmt.Errorf("Could not load settings for site %q: %v",
				siteName, err)
		}
		if len(siteSettings.Locale) == 0 {
			siteSettings.Locale = "en"
		}
		sites[siteName] = siteSettings
	}
	return sites, nil
}

// LoadSiteSettings loads the configurated sites' settings.
func (s *MonstiSettings) LoadSiteSettings() error {
	sites, err := loadSiteSettings(filepath.Join(s.Directories.Config, "sites"))
	if err != nil {
		return err
	}
	s.Sites = sites
	return nil
}

// LoadModuleSettings loads the given module's configuration.
//
// module is the name of the module, e.g. "data"
// cfgPath is the path to the configuration directory
// settings is a pointer to a struct to be filled with the module settings. It
// must contain a field named "Monsti" of type util.MonstiSettings to be filled
// with Monsti's common settings.
func LoadModuleSettings(module, cfgPath string, settings interface{}) error {
	// Value checking
	value := reflect.ValueOf(settings)
	if !value.IsValid() || value.Kind() != reflect.Ptr ||
		reflect.Indirect(value).Kind() != reflect.Struct {
		return fmt.Errorf("util: LoadModuleSettings expects its third " +
			"argument to be a pointer to a struct")
	}
	monstiSettings := reflect.Indirect(value).FieldByName("Monsti")
	if reflect.ValueOf(monstiSettings).Kind() != reflect.Struct {
		return fmt.Errorf("util: LoadModuleSettings expects its third " +
			`argument to contain a field named "Monsti" of type ` +
			`util.MonstiSettings`)
	}

	// Load module settings
	path := filepath.Join(cfgPath, module+".yaml")
	if err := ParseYAML(path, settings); err != nil {
		return fmt.Errorf("util: Could not parse module settings: %v", err)
	}

	// Load Monsti settings
	monstiSettingsRet, err := LoadMonstiSettings(cfgPath)
	if err != nil {
		return fmt.Errorf("util: Could not load monsti settings: %v", err)
	}
	monstiSettings.Set(reflect.ValueOf(*monstiSettingsRet))
	return nil
}

func LoadMonstiSettings(cfgPath string) (*MonstiSettings, error) {
	path := filepath.Join(cfgPath, "monsti.yaml")
	var settings MonstiSettings
	if err := ParseYAML(path, &settings); err != nil {
		return nil, fmt.Errorf("util: Could not parse Monsti settings: %v", err)
	}
	settings.Directories.Config = cfgPath
	MakeAbsolute(&settings.Directories.Data, cfgPath)
	MakeAbsolute(&settings.Directories.Share, cfgPath)
	MakeAbsolute(&settings.Directories.Locale, cfgPath)
	MakeAbsolute(&settings.Directories.Run, cfgPath)
	return &settings, nil
}
