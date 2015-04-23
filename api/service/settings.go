// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann
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

package service

import (
	"encoding/json"
	"fmt"
)

// Settings contains definition and data of a subset of site settings.
type Settings struct {
	FieldConfigs []*FieldConfig   `json:"-"`
	Fields       map[string]Field `json:"-"`
}

// StringValue returns the value of the given string field.
func (n *Settings) StringValue(field string) string {
	return n.Fields[field].Value().(string)
}

func (n *Settings) InitFields(m *MonstiClient, site string) error {
	n.Fields = make(map[string]Field)
	return initFields(n.Fields, n.FieldConfigs, m, site)
}

// toData converts the settings to a JSON document.
func (s *Settings) toData(indent bool) ([]byte, error) {
	out, err := dumpFields(s.Fields, s.FieldConfigs)
	if err != nil {
		return nil, err
	}
	var data []byte
	if indent {
		data, err = json.MarshalIndent(out, "", "  ")
	} else {
		data, err = json.Marshal(out)
	}
	if err != nil {
		return nil, fmt.Errorf(
			"service: Could not marshal settings: %v", err)
	}
	return data, nil
}

// newSettingsFromData unmarshals given settings data.
func newSettingsFromData(data []byte, fieldConfigs []*FieldConfig,
	m *MonstiClient, site string) (
	*Settings, error) {
	ret := &Settings{FieldConfigs: fieldConfigs}
	if err := ret.InitFields(m, site); err != nil {
		return nil, fmt.Errorf("Could not init settings fields: %v", err)
	}
	if len(data) == 0 {
		return ret, nil
	}
	var fields map[string]map[string]*json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf(
			"service: Could not unmarshal settings: %v", err)
	}
	if err := restoreFields(fields, fieldConfigs, ret.Fields); err != nil {
		return nil, err
	}
	return ret, nil
}
