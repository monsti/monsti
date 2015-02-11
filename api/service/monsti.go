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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/smtp"
	"net/url"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"pkg.monsti.org/monsti/api/util/i18n"
)

// MonstiClient represents the RPC connection to the Monsti service.
type MonstiClient struct {
	Client
	SignalHandlers map[string]func(interface{}) (interface{}, error)
}

// NewMonstiConnection establishes a new RPC connection to a Monsti service.
//
// path is the unix domain socket path to the service.
func NewMonstiConnection(path string) (*MonstiClient, error) {
	var service MonstiClient
	if err := service.Connect(path); err != nil {
		return nil,
			fmt.Errorf("service: Could not establish connection to Monsti service: %v",
				err)
	}
	return &service, nil
}

// ModuleInitDone tells Monsti that the given module has finished its
// initialization. Monsti won't finish its startup until all modules
// called this method.
func (s *MonstiClient) ModuleInitDone(module string) error {
	if s.Error != nil {
		return s.Error
	}
	err := s.RPCClient.Call("Monsti.ModuleInitDone", module, new(int))
	if err != nil {
		return fmt.Errorf("service: ModuleInitDone error: %v", err)
	}
	return nil
}

// LoadSiteSettings loads settings for the given site.
func (s *MonstiClient) LoadSiteSettings(site string) (*Settings, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	var reply []byte
	err := s.RPCClient.Call("Monsti.LoadSiteSettings", site, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: LoadSiteSettings error: %v", err)
	}

	G := func(in string) string { return in }
	types := []*NodeField{
		{
			Id:       "core.SiteTitle",
			Required: true,
			Name:     i18n.GenLanguageMap(G("Site title"), []string{"de", "en"}),
			Type:     "Text",
		},
	}

	settings, err := newSettingsFromData(reply, types, s, site)
	if err != nil {
		return nil, fmt.Errorf("service: Could not convert node: %v", err)
	}
	return settings, nil
}

// WriteSiteSettings writes the given settings.
func (s *MonstiClient) WriteSiteSettings(site string, settings *Settings) error {
	if s.Error != nil {
		return s.Error
	}
	data, err := settings.toData(true)
	if err != nil {
		return fmt.Errorf("service: Could not convert settings: %v", err)
	}
	args := struct {
		Site     string
		Settings []byte
	}{site, data}
	err = s.RPCClient.Call("Monsti.WriteSiteSettings", &args, new(int))
	if err != nil {
		return fmt.Errorf("service: WriteSiteSettings error: %v", err)
	}
	return nil
}

// nodeToData converts the node to a JSON document.
// The Path field will be omitted.
func nodeToData(node *Node, indent bool) ([]byte, error) {
	var data []byte
	var err error
	path := node.Path
	node.Path = ""
	defer func() {
		node.Path = path
	}()

	var outNode nodeJSON
	outNode.Node = *node
	outNode.Type = node.Type.Id

	nodeFields := append(node.Type.Fields, node.LocalFields...)
	outNode.Fields, err = dumpFields(node.Fields, nodeFields)
	if err != nil {
		return nil, err
	}

	if indent {
		data, err = json.MarshalIndent(outNode, "", "  ")
	} else {
		data, err = json.Marshal(outNode)
	}
	if err != nil {
		return nil, fmt.Errorf(
			"service: Could not marshal node: %v", err)
	}
	return data, nil

}

// dumpFields converts the given fields to a two-dimensional map
// consisting of JSON raw messages.
func dumpFields(fields map[string]Field, types []*NodeField) (
	map[string]map[string]*json.RawMessage, error) {
	out := make(map[string]map[string]*json.RawMessage)
	for _, field := range types {
		parts := strings.SplitN(field.Id, ".", 2)
		dump, err := json.Marshal(fields[field.Id].Dump())
		if err != nil {
			return nil, fmt.Errorf("Could not marshal field: %v", err)
		}
		if out[parts[0]] == nil {
			out[parts[0]] = make(map[string]*json.RawMessage)
		}
		msg := json.RawMessage(dump)
		out[parts[0]][parts[1]] = &msg
	}
	return out, nil
}

// WriteNode writes the given node.
func (s *MonstiClient) WriteNode(site, path string, node *Node) error {
	if s.Error != nil {
		return s.Error
	}
	node.Changed = time.Now().UTC()
	data, err := nodeToData(node, true)
	if err != nil {
		return fmt.Errorf("service: Could not convert node: %v", err)
	}
	err = s.WriteNodeData(site, path, "node.json", data)
	if err != nil {
		return fmt.Errorf(
			"service: Could not write node: %v", err)
	}
	return nil
}

type nodeJSON struct {
	Node
	Type   string
	Fields map[string]map[string]*json.RawMessage
}

// restoreFields converts the given raw data to an array of already
// initialized fields.
func restoreFields(fields map[string]map[string]*json.RawMessage,
	types []*NodeField, out map[string]Field) error {
	for _, field := range types {
		parts := strings.SplitN(field.Id, ".", 2)
		value := fields[parts[0]][parts[1]]
		if value != nil {
			f := func(in interface{}) error {
				return json.Unmarshal(*value, in)
			}
			out[field.Id].Load(f)
		}
	}
	return nil
}

// dataToNode unmarshals given data
func dataToNode(data []byte,
	getNodeType func(id string) (*NodeType, error), m *MonstiClient, site string) (
	*Node, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var node nodeJSON
	err := json.Unmarshal(data, &node)
	if err != nil {
		return nil, fmt.Errorf(
			"service: Could not unmarshal node: %v", err)
	}
	ret := node.Node
	ret.Type, err = getNodeType(node.Type)
	if err != nil {
		return nil, fmt.Errorf("Could not get node type %q: %v",
			node.Type, err)
	}

	if err = ret.InitFields(m, site); err != nil {
		return nil, fmt.Errorf("Could not init node fields (node: %q): %v", ret, err)
	}
	nodeFields := append(ret.Type.Fields, ret.LocalFields...)
	if err = restoreFields(node.Fields, nodeFields, ret.Fields); err != nil {
		return nil, err
	}
	return &ret, nil
}

// GetNode reads the given node.
//
// If the node does not exist, it returns nil, nil.
func (s *MonstiClient) GetNode(site, path string) (*Node, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	args := struct{ Site, Path string }{site, path}
	var reply []byte
	err := s.RPCClient.Call("Monsti.GetNode", args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetNode error: %v", err)
	}
	node, err := dataToNode(reply, s.GetNodeType, s, site)
	if err != nil {
		return nil, fmt.Errorf("service: Could not convert node: %v", err)
	}
	return node, nil
}

// GetChildren returns the children of the given node.
func (s *MonstiClient) GetChildren(site, path string) ([]*Node, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	args := struct{ Site, Path string }{site, path}
	var reply [][]byte
	err := s.RPCClient.Call("Monsti.GetChildren", args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetChildren error: %v", err)
	}
	nodes := make([]*Node, 0, len(reply))
	for _, entry := range reply {

		node, err := dataToNode(entry, s.GetNodeType, s, site)
		if err != nil {
			return nil, fmt.Errorf("service: Could not convert node: %v", err)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// GetNodeData requests data from some node.
//
// Returns a nil slice and nil error if the data does not exist.
func (s *MonstiClient) GetNodeData(site, path, file string) ([]byte, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	type GetNodeDataArgs struct {
	}
	args := struct{ Site, Path, File string }{
		site, path, file}
	var reply []byte
	err := s.RPCClient.Call("Monsti.GetNodeData", &args, &reply)
	if err != nil {
		return nil, fmt.Errorf("service: GetNodeData error:", err)
	}
	return reply, nil
}

// WriteNodeData writes data for some node.
func (s *MonstiClient) WriteNodeData(site, path, file string,
	content []byte) error {
	if s.Error != nil {
		return s.Error
	}
	args := struct {
		Site, Path, File string
		Content          []byte
	}{
		site, path, file, content}
	if err := s.RPCClient.Call("Monsti.WriteNodeData", &args, new(int)); err != nil {
		return fmt.Errorf("service: WriteNodeData error: %v", err)
	}
	return nil
}

// RemoveNodeData removes data of some node.
func (s *MonstiClient) RemoveNodeData(site, path, file string) error {
	if s.Error != nil {
		return s.Error
	}
	args := struct {
		Site, Path, File string
	}{site, path, file}
	if err := s.RPCClient.Call("Monsti.RemoveNodeData", &args, new(int)); err != nil {
		return fmt.Errorf("service: RemoveNodeData error: %v", err)
	}
	return nil
}

// RemoveNode removes the given site's node and all its descendants.
//
// All reverse cache dependencies of removed nodes will be marked.
func (s *MonstiClient) RemoveNode(site string, node string) error {
	if s.Error != nil {
		return s.Error
	}
	args := struct {
		Site, Node string
	}{site, node}
	if err := s.RPCClient.Call("Monsti.RemoveNode", args, new(int)); err != nil {
		return fmt.Errorf("service: RemoveNode error: %v", err)
	}
	return nil
}

// RenameNode renames (moves) the given site's node.
//
// Source and target path must be absolute. TODO: All reverse cache
// dependencies of moved nodes will be marked.
func (s *MonstiClient) RenameNode(site, source, target string) error {
	if s.Error != nil {
		return s.Error
	}
	args := struct {
		Site, Source, Target string
	}{site, source, target}
	if err := s.RPCClient.Call("Monsti.RenameNode", args, new(int)); err != nil {
		return fmt.Errorf("service: RenameNode error: %v", err)
	}
	return nil
}

func getConfig(reply []byte, out interface{}) error {
	if len(reply) == 0 {
		return nil
	}
	objectV := reflect.New(
		reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(out)))
	err := json.Unmarshal(reply, objectV.Interface())
	if err != nil {
		return fmt.Errorf("service: Could not decode configuration: %v", err)
	}
	value := objectV.Elem().MapIndex(
		objectV.Elem().MapKeys()[0])
	if !value.IsNil() {
		reflect.ValueOf(out).Elem().Set(value.Elem())
	}
	return nil
}

// GetSiteConfig puts the named site local configuration into the
// variable out.
func (s *MonstiClient) GetSiteConfig(site, name string, out interface{}) error {
	if s.Error != nil {
		return s.Error
	}
	args := struct{ Site, Name string }{site, name}
	var reply []byte
	err := s.RPCClient.Call("Monsti.GetSiteConfig", args, &reply)
	if err != nil {
		return fmt.Errorf("service: GetSiteConfig error: %v", err)
	}
	return getConfig(reply, out)
}

/*

// GetConfig puts the named global configuration into the variable out.
func (s *MonstiClient) GetConfig(name string, out interface{}) error {
	if s.Error != nil {
		return s.Error
	}
	var reply []byte
	err := s.RPCClient.Call("Monsti.GetConfig", name, &reply)
	if err != nil {
		return fmt.Errorf("service: GetConfig error: %v", err)
	}
	return getConfig(reply, out)
}

*/

// RegisterNodeType registers a new node type.
//
// Known field types will be reused. Just specify the id. All other //
// attributes of the field type will be ignored in this case.
func (s *MonstiClient) RegisterNodeType(nodeType *NodeType) error {
	if s.Error != nil {
		return s.Error
	}
	err := s.RPCClient.Call("Monsti.RegisterNodeType", nodeType, new(int))
	if err != nil {
		return fmt.Errorf("service: Error calling RegisterNodeType: %v", err)
	}
	return nil
}

// GetNodeType requests information about the given node type.
func (s *MonstiClient) GetNodeType(nodeTypeID string) (*NodeType,
	error) {
	if s.Error != nil {
		return nil, s.Error
	}
	var nodeType NodeType
	err := s.RPCClient.Call("Monsti.GetNodeType", nodeTypeID, &nodeType)
	if err != nil {
		return nil, fmt.Errorf("service: Error calling GetNodeType: %v", err)
	}
	return &nodeType, nil
}

// GetAddableNodeTypes returns the node types that may be added as child nodes
// to the given node type at the given website.
func (s *MonstiClient) GetAddableNodeTypes(site, nodeType string) (types []string,
	err error) {
	if s.Error != nil {
		return nil, s.Error
	}
	args := struct{ Site, NodeType string }{site, nodeType}
	err = s.RPCClient.Call("Monsti.GetAddableNodeTypes", args, &types)
	if err != nil {
		err = fmt.Errorf("service: Error calling GetAddableNodeTypes: %v", err)
	}
	return
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
*/

type RequestMethod uint

const (
	GetRequest = iota
	PostRequest
)

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
	ListAction
	ChooserAction
	SettingsAction
)

// A request to be processed by a nodes service.
type Request struct {
	Id       uint
	NodePath string
	// Site name
	Site string
	// The query values of the request URL.
	Query url.Values
	// Method of the request (GET,POST,...).
	Method RequestMethod
	// User session
	Session *UserSession
	// Action to perform (e.g. "edit").
	Action Action
	// FormData stores the requests form data.
	FormData url.Values
	/*
			// The requested node.
			Node *Node
		// Files stores files of multipart requests.
				Files map[string][]RequestFile
	*/
}

// GetRequest returns the request with the given id.
//
// If there is no request with the given id, it returns nil.
func (s *MonstiClient) GetRequest(id uint) (*Request, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	var req Request
	if err := s.RPCClient.Call("Monsti.GetRequest", id, &req); err != nil {
		return nil, fmt.Errorf("service: Monsti.GetRequest error: %v", err)
	}
	if req.Id != id {
		return nil, nil
	}
	return &req, nil
}

// GetNodeType returns all supported node types.
func (s *MonstiClient) GetNodeTypes() ([]string, error) {
	if s.Error != nil {
		return nil, s.Error
	}
	var res []string
	err := s.RPCClient.Call("Monsti.GetNodeTypes", 0, &res)
	if err != nil {
		return nil, fmt.Errorf("service: RPC error for GetNodeTypes: %v", err)
	}
	return res, nil
}

// PublishService informs the INFO service about a new service.
//
// service is the identifier of the service
// path is the path to the unix domain socket of the service
//
// If the data does not exist, return null length []byte.
func (s *MonstiClient) PublishService(service, path string) error {
	args := struct{ Service, Path string }{service, path}
	if s.Error != nil {
		return s.Error
	}
	var reply int
	err := s.RPCClient.Call("Monsti.PublishService", args, &reply)
	if err != nil {
		return fmt.Errorf("service: Error calling PublishService: %v", err)
	}
	return nil
}

/*
// FindDataService requests a data client.
func (s *MonstiClient) FindDataService() (*MonstiClient, error) {
	var path string
	err := s.RPCClient.Call("Monsti.FindDataService", 0, &path)
	if err != nil {
		return nil, fmt.Errorf("service: Error calling FindDataService: %v", err)
	}
	service_ := NewDataClient()
	if err := service_.Connect(path); err != nil {
		return nil,
			fmt.Errorf("service: Could not establish connection to data service: %v",
				err)
	}
	return service_, nil
}
*/

// User represents a registered user of the site.
type User struct {
	Login string
	Name  string
	Email string
	// Hashed password.
	Password string
	// PasswordChanged keeps the time of the last password change.
	PasswordChanged time.Time
}

// UserSession is a session of an authenticated or anonymous user.
type UserSession struct {
	// Authenticaded user or nil
	User *User
	// Locale used for this session.
	Locale string
}

// SendMails sends the given mail.
func (s *MonstiClient) SendMail(from string, to []string, msg []byte) error {
	if s.Error != nil {
		return s.Error
	}
	args := struct {
		From string
		To   []string
		Msg  []byte
	}{from, to, msg}
	var reply int
	if err := s.RPCClient.Call("Monsti.SendMail", args, &reply); err != nil {
		return fmt.Errorf("service: Monsti.SendMail error: %v", err)
	}
	return nil
}

// SendMailFunc returns a function to send mails using SendMail.
//
// The function has a signature compatible to smtp.SendMail. It can be
// used with packages like `gomail`.
//
// The first two arguments (address and auth) will be ignored.
func (s *MonstiClient) SendMailFunc() func(
	string, smtp.Auth, string, []string, []byte) error {
	return func(_ string, _ smtp.Auth, from string, to []string,
		msg []byte) error {
		return s.SendMail(from, to, msg)
	}
}

// AddSignalHandler connects to a signal with the given signal handler.
//
// Currently, you can only set one handler per signal and MonstiClient.
//
// Be sure to wait for incoming signals by calling WaitSignal() on
// this MonstiClient!
func (s *MonstiClient) AddSignalHandler(handler SignalHandler) error {
	if s.Error != nil {
		return s.Error
	}
	args := struct{ Id, Signal string }{s.Id, handler.Name()}
	err := s.RPCClient.Call("Monsti.ConnectSignal", args, new(int))
	if err != nil {
		return fmt.Errorf("service: Monsti.ConnectSignal error: %v", err)
	}
	if s.SignalHandlers == nil {
		s.SignalHandlers = make(map[string]func(interface{}) (interface{}, error))
	}
	s.SignalHandlers[handler.Name()] = handler.Handle
	return nil
}

type argWrap struct{ Wrap interface{} }

// EmitSignal emits the named signal with given arguments and return
// value.
func (s *MonstiClient) EmitSignal(name string, args interface{},
	retarg interface{}) error {
	if s.Error != nil {
		return s.Error
	}
	gob.RegisterName(name+"Ret", reflect.Zero(
		reflect.TypeOf(retarg).Elem().Elem()).Interface())
	gob.RegisterName(name+"Args", args)
	var args_ struct {
		Name string
		Args []byte
	}
	buffer := &bytes.Buffer{}
	enc := gob.NewEncoder(buffer)
	err := enc.Encode(argWrap{args})
	if err != nil {
		return fmt.Errorf("service: Could not encode signal argumens: %v", err)
	}
	args_.Name = name
	args_.Args = buffer.Bytes()
	var ret [][]byte
	err = s.RPCClient.Call("Monsti.EmitSignal", args_, &ret)
	if err != nil {
		return fmt.Errorf("service: Monsti.EmitSignal error: %v", err)
	}
	reflect.ValueOf(retarg).Elem().Set(reflect.MakeSlice(
		reflect.TypeOf(retarg).Elem(), len(ret), len(ret)))
	for i, answer := range ret {
		buffer = bytes.NewBuffer(answer)
		dec := gob.NewDecoder(buffer)
		var ret_ argWrap
		err = dec.Decode(&ret_)
		if err != nil {
			return fmt.Errorf("service: Could not decode signal return value: %v", err)
		}
		reflect.ValueOf(retarg).Elem().Index(i).Set(reflect.ValueOf(ret_.Wrap))
	}
	return nil
}

// WaitSignal waits for the next emitted signal.
//
// You have to connect to some signals before. See AddSignalHandler.
// This method must not be called in parallel by the same client
// instance.
func (s *MonstiClient) WaitSignal() error {
	if s.Error != nil {
		return s.Error
	}
	signal := struct {
		Name string
		Args []byte
	}{}
	err := s.RPCClient.Call("Monsti.WaitSignal", s.Id, &signal)
	if err != nil {
		return fmt.Errorf("service: Monsti.WaitSignal error: %v", err)
	}
	buffer := bytes.NewBuffer(signal.Args)
	dec := gob.NewDecoder(buffer)
	var args_ argWrap
	err = dec.Decode(&args_)
	if err != nil {
		return fmt.Errorf("service: Could not decode signal argumens: %v", err)
	}
	var ret interface{}
	var reterr error
	func() {
		defer func() {
			if err := recover(); err != nil {
				var buf bytes.Buffer
				fmt.Fprintf(&buf, "error: %v\n", err)
				buf.Write(debug.Stack())
				reterr = errors.New(buf.String())
			}
		}()
		ret, reterr = s.SignalHandlers[signal.Name](args_.Wrap)
	}()
	signalRet := &struct {
		Id  string
		Err string
		Ret []byte
	}{Id: s.Id}
	buffer = &bytes.Buffer{}
	if reterr == nil {
		enc := gob.NewEncoder(buffer)
		err = enc.Encode(argWrap{ret})
		if err != nil {
			return fmt.Errorf("service: Could not encode signal return value: %v", err)
		}
	}
	signalRet.Ret = buffer.Bytes()
	if reterr != nil {
		signalRet.Err = reterr.Error()
	}
	err = s.RPCClient.Call("Monsti.FinishSignal", signalRet, new(int))
	if err != nil {
		return fmt.Errorf("service: Monsti.FinishSignal error: %v", err)
	}
	return nil
}

// CacheMods describes how a cache should be modified.
//
// It's usually returned from functions to modify caches that are
// written to higher in the call graph.
type CacheMods struct {
	// Dependencies.
	Deps []CacheDep
	// Don't write to the cache.
	Skip bool
	// The cache expires at this time (unless it's the zero value).
	Expire time.Time
}

// Join joins the given cache mods into the current cache mods.
//
// Deps will be appended, Skip will be ored, and Expire will be
// minimized. If right is nil, nothing will change.
func (c *CacheMods) Join(right *CacheMods) {
	if right != nil {
		c.Deps = append(c.Deps, right.Deps...)
		c.Skip = c.Skip || right.Skip
		if c.Expire.IsZero() ||
			!right.Expire.IsZero() && right.Expire.Before(c.Expire) {
			c.Expire = right.Expire
		}
	}
}

// CacheDep identifies something a cache may depend on and which can
// be marked.
type CacheDep struct {
	// Path to the node or subtree.
	Node string
	// Cache ID.
	Cache string `json:",omitempty"`
	// On how many child levels does this dependency apply? -1 for the
	// whole subtree. 0 for this node only.
	Descend int `json:",omitempty"`
}

// ToCache caches the given data.
//
// Each node has a cache where arbitrary data can be stored. The data
// may be retrieved later with the FromCache method. If the node or
// any other nodes as specified in the deps argument change, the
// cached data will be deleted. Each cached data is identified by an
// id which contains a namespace prefix,
// e.g. `mymodule.thumbnail_large`.
//
// On success, ToCache will replace the CacheMods deps with a CacheDep
// describing this cache.
func (s *MonstiClient) ToCache(site, node string, id string,
	content []byte, mods *CacheMods) error {
	if mods.Skip {
		return nil
	}
	if s.Error != nil {
		return s.Error
	}
	args := struct {
		Node, Site, Id string
		Content        []byte
		Mods           *CacheMods
	}{node, site, id, content, mods}
	if err := s.RPCClient.Call("Monsti.ToCache", &args, new(int)); err != nil {
		return fmt.Errorf("service: ToCache error: %v", err)
	}
	mods.Deps = []CacheDep{{Node: node, Cache: id}}
	return nil
}

// FromCache retrieves the given cached data or nil if the cache is empty.
//
// See ToCache for more information.
func (s *MonstiClient) FromCache(site string, node string,
	id string) ([]byte, *CacheMods, error) {
	if s.Error != nil {
		return nil, nil, s.Error
	}
	args := struct{ Node, Site, Id string }{node, site, id}
	var reply struct {
		CacheMods *CacheMods
		Data      []byte
	}
	err := s.RPCClient.Call("Monsti.FromCache", &args, &reply)
	if err != nil {
		return nil, nil, fmt.Errorf("service: FromCache error: %v", err)
	}
	return reply.Data, reply.CacheMods, nil
}

// MarkDep marks the given cache dependency as dirty.
//
// See ToCache for more information.
func (s *MonstiClient) MarkDep(site string, dep CacheDep) error {
	if s.Error != nil {
		return s.Error
	}
	args := struct {
		Site string
		Dep  CacheDep
	}{site, dep}
	if err := s.RPCClient.Call("Monsti.MarkDep", &args, new(int)); err != nil {
		return fmt.Errorf("service: MarkDep error: %v", err)
	}
	return nil
}
