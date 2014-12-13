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

package service

import (
	"fmt"
	"html/template"
	"path"
	"strings"
	"time"

	"github.com/chrneumann/htmlwidgets"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/util"
)

type Field interface {
	// RenderHTML returns a string or template.HTML to be used in a html
	// template.
	RenderHTML() interface{}
	// String returns a raw string representation of the field.
	String() string
	// Load loads the field data (also see Dump).
	Load(interface{}) error
	// Dump dumps the field data.
	//
	// The dumped value must be something that can be marshalled into
	// JSON by encoding/json.
	Dump() interface{}
	// Adds a form field to the node edit form.
	ToFormField(*htmlwidgets.Form, util.NestedMap, *NodeField, string)
	// Load values from the form submission
	FromFormField(util.NestedMap, *NodeField)
}

// TextField is a basic unicode text field
type TextField string

func (t TextField) String() string {
	return string(t)
}

func (t TextField) RenderHTML() interface{} {
	return t
}

func (t *TextField) Load(in interface{}) error {
	*t = TextField(in.(string))
	return nil
}

func (t TextField) Dump() interface{} {
	return string(t)
}

func (t TextField) ToFormField(form *htmlwidgets.Form, data util.NestedMap,
	field *NodeField, locale string) {
	data.Set(field.Id, string(t))
	G, _, _, _ := gettext.DefaultLocales.Use("", locale)
	form.AddWidget(&htmlwidgets.TextWidget{
		MinLength: 1, ValidationError: G("Required.")}, "Fields."+field.Id,
		field.Name[locale], "")
}

func (t *TextField) FromFormField(data util.NestedMap, field *NodeField) {
	*t = TextField(data.Get(field.Id).(string))
}

// HTMLField is a text area containing HTML code
type HTMLField string

func (t HTMLField) String() string {
	return string(t)
}

func (t HTMLField) RenderHTML() interface{} {
	return template.HTML(t)
}

func (t *HTMLField) Load(in interface{}) error {
	*t = HTMLField(in.(string))
	return nil
}

func (t HTMLField) Dump() interface{} {
	return string(t)
}

func (t HTMLField) ToFormField(form *htmlwidgets.Form, data util.NestedMap,
	field *NodeField, locale string) {
	//G, _, _, _ := gettext.DefaultLocales.Use("", locale)
	data.Set(field.Id, string(t))
	widget := form.AddWidget(new(htmlwidgets.TextAreaWidget), "Fields."+field.Id,
		field.Name[locale], "")
	widget.Base().Classes = []string{"html-field"}
}

func (t *HTMLField) FromFormField(data util.NestedMap, field *NodeField) {
	*t = HTMLField(data.Get(field.Id).(string))
}

type FileField string

func (t FileField) String() string {
	return string(t)
}

func (t FileField) RenderHTML() interface{} {
	return template.HTML(t)
}

func (t *FileField) Load(in interface{}) error {
	*t = FileField(in.(string))
	return nil
}

func (t FileField) Dump() interface{} {
	return ""
}

func (t FileField) ToFormField(form *htmlwidgets.Form, data util.NestedMap,
	field *NodeField, locale string) {
	data.Set(field.Id, "")
	form.AddWidget(new(htmlwidgets.FileWidget), "Fields."+field.Id,
		field.Name[locale], "")
}

func (t *FileField) FromFormField(data util.NestedMap, field *NodeField) {
	*t = FileField(data.Get(field.Id).(string))
}

type DateTimeField struct {
	time.Time
}

func (t DateTimeField) RenderHTML() interface{} {
	return t.String()
}

func (t *DateTimeField) Load(in interface{}) error {
	date, ok := in.(string)
	if !ok {
		return fmt.Errorf("Data is not string")
	}
	val, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return fmt.Errorf("Could not parse the date value: %v", err)
	}
	*t = DateTimeField{Time: val}
	return nil
}

func (t DateTimeField) Dump() interface{} {
	return t.Format(time.RFC3339)
}

func (t DateTimeField) ToFormField(form *htmlwidgets.Form, data util.NestedMap,
	field *NodeField, locale string) {
	data.Set(field.Id, t.Time)
	form.AddWidget(&htmlwidgets.TimeWidget{}, "Fields."+field.Id,
		field.Name[locale], "")
}

func (t *DateTimeField) FromFormField(data util.NestedMap, field *NodeField) {
	*t = DateTimeField{Time: data.Get(field.Id).(time.Time)}
}

// TemplateOverwrite specifies a template that should be used instead
// of another.
type TemplateOverwrite struct {
	// The template to be used instead.
	Template string
}

type Node struct {
	Path string `json:",omitempty"`
	// Content type of the node.
	Type  *NodeType `json:"-"`
	Order int
	// Don't show the node in navigations if Hide is true.
	Hide               bool
	Fields             map[string]Field `json:"-"`
	TemplateOverwrites map[string]TemplateOverwrite
	Embed              []EmbedNode
	LocalFields        []*NodeField
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

func (n *Node) InitFields() {
	n.Fields = make(map[string]Field)
	nodeFields := append(n.Type.Fields, n.LocalFields...)
	for _, field := range nodeFields {
		var val Field
		switch field.Type {
		case "DateTime":
			val = new(DateTimeField)
		case "File":
			val = new(FileField)
		case "Text":
			val = new(TextField)
		case "HTMLArea":
			val = new(HTMLField)
		default:
			panic(fmt.Sprintf("Unknown field type %q for node %q", field.Type, n.Path))
		}
		n.Fields[field.Id] = val
	}
}

func (n Node) GetField(id string) Field {
	return n.Fields[id]
}

func (n Node) GetValue(id string) interface{} {
	return n.Fields[id]
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

// TypeToID returns an ID for the given node type.
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

type NodeField struct {
	// The Id of the field including a namespace,
	// e.g. "namespace.somefieldype".
	Id string
	// The name of the field as shown in the web interface,
	// specified as a translation map (language -> msg).
	Name     map[string]string
	Required bool
	Type     string
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
	Name   map[string]string
	Fields []*NodeField
	Embed  []EmbedNode
	// If true, never show nodes of this type in the navigation.
	Hide bool
	// PathPrefix defines a dynamic path that will be prepended to the
	// node name.
	//
	// Supported values: $year, $month, $day
	PathPrefix string
}

// GetLocalName returns the name of the node type in the given language.
//
// Fall backs to to the "en" locale or the id of the node type.
func (n NodeType) GetLocalName(locale string) string {
	name, ok := n.Name[locale]
	if !ok {
		name, ok = n.Name["en"]
	}
	if !ok {
		name = n.Id
	}
	return name
}
