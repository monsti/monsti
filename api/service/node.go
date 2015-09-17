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
	"path"
	"strings"
	"time"

	"pkg.monsti.org/monsti/api/util/i18n"
)

// CoreFields contains the configurations for the fields that should be used
// by most node types.
var CoreFields = []*FieldConfig{
	{Id: "core.Title"}, {Id: "core.Description"}, {Id: "core.Thumbnail"},
	{Id: "core.Body"}, {Id: "core.Categories"}}

type Node struct {
	Path string `json:",omitempty"`
	// Content type of the node.
	Type  *NodeType `json:"-"`
	Order int
	// Don't show the node in navigations if Hide is true.
	Hide   bool
	Fields map[string]Field `json:"-"`
	Embed  []EmbedNode
	// Public controls wether the node or its content may be viewed by
	// unauthenticated users.
	Public bool
	// PublishTime holds the time the node has been or should be
	// published.
	PublishTime time.Time
	// Changed is updated with the current time on every write to the
	// database.
	Changed time.Time
}

func (n *Node) InitFields(m *MonstiClient, site string) error {
	n.Fields = make(map[string]Field)
	return initFields(n.Fields, n.Type.Fields, m, site)
}

// PathToID returns an ID for the given node based on it's path.
//
// The ID is simply the path of the node with all slashes replaced by two
// underscores and the result prefixed with "node-".
//
// PathToID will panic if the path is not set.
//
// For example, a node with path "/foo/bar" will get the ID "node-__foo__bar".
func (n Node) PathToID() string {
	if len(n.Path) == 0 {
		panic("Can't calculate ID of node with unset path.")
	}
	return "node-" + strings.Replace(n.Path, "/", "__", -1)
}

// DEPRECATED TypeToID returns an ID for the given node type.
//
// The ID is simply the type of the node with the namespace dot
// replaced by a hyphen and the result prefixed with "node-type-".
func (n Node) TypeToID() string {
	return "node-type-" + strings.Replace(n.Type.Id, ".", "-", 1)
}

// Name returns the name of the node.
func (n Node) Name() string {
	base := path.Base(n.Path)
	if base == "." || base == "/" {
		return ""
	}
	return base
}

// GetPathPrefix returns the calculated prefix path.
func (n Node) GetPathPrefix() string {
	if n.Type == nil {
		return ""
	}
	prefix := n.Type.PathPrefix
	prefix = strings.Replace(prefix, "$year", n.PublishTime.Format("2006"), -1)
	prefix = strings.Replace(prefix, "$month", n.PublishTime.Format("01"), -1)
	prefix = strings.Replace(prefix, "$day", n.PublishTime.Format("02"), -1)
	return prefix
}

// GetParentPath calculates the parent node's path respecting the
// node's path prefix.
func (n Node) GetParentPath() string {
	prefix := n.GetPathPrefix()
	nodePath := n.Path
	for prefix != "" && prefix != "." && prefix != "/" {
		prefix = path.Dir(prefix)
		nodePath = path.Dir(nodePath)
	}
	return path.Dir(nodePath)
}

type EmbedNode struct {
	Id  string
	URI string
}

type NodeQuery struct {
	Id string
}

type NodeType struct {
	// The Id of the node type including a namespace,
	// e.g. "namespace.somenodetype".
	Id string
	// Per default, nodes may not be added to any other node. This
	// behaviour can be overwritten with this option. Nodes of this type
	// may only be added to nodes of the specified types. You may
	// specify individual node types with their full id `namespace.id`
	// or all node types of a namespace using `namespace.` (i.e. the
	// namespace followed by a single dot). To specify all available
	// node types, use the single dot, i.e.`.`. It's always possible to
	// add nodes to any other node by directly manipulating the node
	// data on the file system. This option merely affects the web
	// interface.
	AddableTo []string
	// The name of the node type as shown in the web interface,
	// specified as a translation map (language -> msg).
	Name   i18n.LanguageMap
	Fields []*FieldConfig
	Embed  []EmbedNode
	// If true, never show nodes of this type in the navigation.
	Hide bool
	// PathPrefix defines a dynamic path that will be prepended to the
	// node name.
	//
	// Supported values: $year, $month, $day
	PathPrefix string
}
