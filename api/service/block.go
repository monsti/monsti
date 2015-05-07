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
	"strings"
	"time"

	"pkg.monsti.org/monsti/api/util/i18n"
)

type Block struct {
	Id string `json:",omitempty"`
	// Content type of the block.
	Type   *BlockType       `json:"-"`
	Fields map[string]Field `json:"-"`
	// Changed is updated with the current time on every write to the
	// database.
	Changed time.Time
}

func (n *Block) InitFields(m *MonstiClient, site string) error {
	n.Fields = make(map[string]Field)
	return initFields(n.Fields, n.Type.Fields, m, site)
}

// TypeToID returns an ID for the given block type.
//
// The ID is simply the type of the block with the namespace dot
// replaced by a hyphen and the result prefixed with "block-type-".
func (n Block) TypeToID() string {
	return "block-type-" + strings.Replace(n.Type.Id, ".", "-", 1)
}

type BlockType struct {
	// The Id of the block type including a namespace,
	// e.g. "namespace.someblocktype".
	Id string
	// The name of the block type as shown in the web interface,
	// specified as a translation map (language -> msg).
	Name   i18n.LanguageMap
	Fields []*FieldConfig
}
