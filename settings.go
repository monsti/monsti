package util

import (
	"fmt"
	"github.com/monsti/service"
	"path/filepath"
	"reflect"
)

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
		// Runtime data directory
		Run string
	}
}

// GetServicePath returns the path to the given service's socket.
func (s MonstiSettings) GetServicePath(service service.Type) string {
	return filepath.Join(s.Directories.Run, service.String()+".socket")
}

// GetSiteConfigPath returns the path to the given site's configuration
// directory.
func (s MonstiSettings) GetSiteConfigPath(site string) string {
	return filepath.Join(s.Directories.Config, "sites", site)
}

// GetLocalePath returns the path to the locale directory.
func (s MonstiSettings) GetLocalePath() string {
	return filepath.Join(s.Directories.Share, "locale")
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

// GetSiteTemplatesPath returns the path to the given site's templates
// directory.
func (s MonstiSettings) GetSiteTemplatesPath(site string) string {
	return filepath.Join(s.Directories.Data, site, "templates")
}

// GetStaticsPath returns the path to the global site-static directory.
func (s MonstiSettings) GetStaticsPath() string {
	return filepath.Join(s.Directories.Share, "site-static")
}

// GetTemplatesPath returns the path to the global templates directory.
func (s MonstiSettings) GetTemplatesPath() string {
	return filepath.Join(s.Directories.Share, "templates")
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
	monstiSettings := value.FieldByName("Monsti")
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
	monstiSettings.Set(reflect.Zero(monstiSettings.Type()))
	path = filepath.Join(cfgPath, "monsti.yaml")
	if err := ParseYAML(path,
		monstiSettings.Addr().Interface()); err != nil {
		return fmt.Errorf("util: Could not parse Monsti settings: %v", err)
	}
	return nil
}
