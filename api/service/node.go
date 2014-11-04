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
			panic(fmt.Sprintf("Unknown field type %v", field.Type))
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

/*

// RequestFile stores the path or content of a multipart request's file.
type RequestFile struct {
	// TmpFile stores the path to a temporary file with the contents.
	TmpFile string
	// Content stores the file content if TmpFile is not set.
	Content []byte
}

// ReadFile returns the file's content. Uses io/ioutil ReadFile if the request
// file's content is in a temporary file.
func (r RequestFile) ReadFile() ([]byte, error) {
	if len(r.TmpFile) > 0 {
		return ioutil.ReadFile(r.TmpFile)
	}
	return r.Content, nil
}

type RequestMethod uint

const (
	GetRequest = iota
	PostRequest
)
*/

type Action uint

const (
	ViewAction = iota
	EditAction
	LoginAction
	LogoutAction
	AddAction
	RemoveAction
	RequestPasswordTokenAction
	ChangePasswordAction
)

/*
// A request to be processed by a nodes service.
type Request struct {
	// Site name
	Site string
	// The requested node.
	Node Node
	// The query values of the request URL.
	Query url.Values
	// Method of the request (GET,POST,...).
	Method RequestMethod
	// User session
	Session UserSession
	// Action to perform (e.g. "edit").
	Action Action
	// FormData stores the requests form data.
	FormData url.Values
	// Files stores files of multipart requests.
	Files map[string][]RequestFile
}
*/

/*
// Response to a node request.
type Response struct {
	// The html content to be embedded in the root template.
	Body []byte
	// Raw must be set to true if Body should not be embedded in the root
	// template. The content type will be automatically detected.
	Raw bool
	// If set, redirect to this target using error 303 'see other'.
	Redirect string
	// The node as received by GetRequest, possibly with some fields
	// updated (e.g. modified title).
	//
	// If nil, the original node data is used.
	Node *Node
}
*/

/*
// Write appends the given bytes to the body of the response.
func (r *Response) Write(p []byte) (n int, err error) {
	r.Body = append(r.Body, p...)
	return len(p), nil
}
*/

/*
// Request performs the given request.
func (s *MonstiClient) Request(req *Request) (*Response, error) {
	var res Response
	err := s.RPCClient.Call("Monsti.Request", req, &res)
	if err != nil {
		return nil, fmt.Errorf("service: RPC error for Request: %v", err)
	}
	return &res, nil
}
*/

// GetNodeType returns all supported node types.
func (s *MonstiClient) GetNodeTypes() ([]string, error) {
	var res []string
	err := s.RPCClient.Call("Monsti.GetNodeTypes", 0, &res)
	if err != nil {
		return nil, fmt.Errorf("service: RPC error for GetNodeTypes: %v", err)
	}
	return res, nil
}
