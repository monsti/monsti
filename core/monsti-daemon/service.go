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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"pkg.monsti.org/monsti/api/service"
)

type subscription struct {
}

type emitRet struct {
	Ret   []byte
	Error string
}

type signal struct {
	Name string
	Args []byte
	Ret  chan emitRet
}

type MonstiService struct {
	// Services maps service names to service paths
	Services map[string][]string
	// Mutex to syncronize data access
	mutex         sync.RWMutex
	Settings      *settings
	Logger        *log.Logger
	Handler       *nodeHandler
	moduleInit    map[string]chan bool
	subscriptions map[string][]string
	subscriber    map[string]chan *signal
	subscriberRet map[string]chan emitRet
}

type PublishServiceArgs struct {
	Service, Path string
}

func (i *MonstiService) PublishService(args PublishServiceArgs,
	reply *int) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.Services == nil {
		i.Services = make(map[string][]string)
	}
	switch args.Service {
	default:
		return fmt.Errorf("Unknown service type %v", args.Service)
	}

	if i.Services[args.Service] == nil {
		i.Services[args.Service] = make([]string, 0)
	}
	i.Services[args.Service] = append(i.Services[args.Service], args.Path)

	return nil
}

func (i *MonstiService) ModuleInitDone(args string, reply *int) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.moduleInit[args] <- true
	return nil
}

type SendMailArgs struct {
	From string
	To   []string
	Msg  []byte
}

func (m *MonstiService) SendMail(args SendMailArgs, reply *int) error {
	if !m.Settings.Mail.Debug {
		auth := smtp.PlainAuth("", m.Settings.Mail.Username,
			m.Settings.Mail.Password, strings.Split(m.Settings.Mail.Host, ":")[0])
		if err := smtp.SendMail(m.Settings.Mail.Host, auth,
			args.From, args.To, args.Msg); err != nil {
			return fmt.Errorf("monsti: Could not send email: %v", err)
		}
	} else {
		m.Logger.Printf(`SendMail debug:
Mail from: %v
Recipients: %v
-- Msg Start --
%v
-- Msg End --`,
			args.From, args.To, string(args.Msg))
	}
	return nil
}

type ConnectSignalArgs struct {
	Id, Signal string
}

func (m *MonstiService) ConnectSignal(args *ConnectSignalArgs, ret *int) error {
	if m.subscriptions == nil {
		m.subscriptions = make(map[string][]string)
		m.subscriber = make(map[string]chan *signal)
	}
	m.subscriptions[args.Signal] = append(m.subscriptions[args.Signal], args.Id)
	if _, ok := m.subscriber[args.Id]; !ok {
		m.subscriber[args.Id] = make(chan *signal)
	}
	return nil
}

type Receive struct {
	Name string
	Args []byte
}

func (m *MonstiService) EmitSignal(args *Receive, ret *[][]byte) error {
	*ret = make([][]byte, len(m.subscriptions[args.Name]))
	for i, id := range m.subscriptions[args.Name] {
		retChan := make(chan emitRet)
		done := false
		go func() {
			time.Sleep(time.Second)
			for !done {
				time.Sleep(30 * time.Second)
				m.Logger.Printf(
					"Waiting for signal response. Signal: %v, Subscriber: %v",
					args.Name, id)
			}
		}()
		m.subscriber[id] <- &signal{args.Name, args.Args, retChan}
		emitRet := <-retChan
		if len(emitRet.Error) > 0 {
			return fmt.Errorf("Received error as signal response: %v", emitRet.Error)
		}
		(*ret)[i] = emitRet.Ret
		done = true
	}
	return nil
}

type WaitSignalRet struct {
	Name string
	Args []byte
}

func (m *MonstiService) WaitSignal(subscriber string, ret *WaitSignalRet) error {
	signal := <-m.subscriber[subscriber]
	ret.Name = signal.Name
	ret.Args = signal.Args
	if m.subscriberRet == nil {
		m.subscriberRet = make(map[string]chan emitRet)
	}
	m.subscriberRet[subscriber] = signal.Ret
	return nil
}

type FinishSignalArgs struct {
	Id  string
	Err string
	Ret []byte
}

func (m *MonstiService) FinishSignal(args *FinishSignalArgs, _ *int) error {
	m.subscriberRet[args.Id] <- emitRet{args.Ret, args.Err}
	return nil
}

// getNode looks up the given node.
// If no such node exists, return nil.
// If it's an existing path but not a regular note, a node of type
// core.Path will be returned.
// It adds a path attribute with the given path.
func getNode(root, path string) (node []byte, err error) {
	nodePath := filepath.Join(root, path[1:])
	node, err = ioutil.ReadFile(filepath.Join(nodePath, "node.json"))
	if os.IsNotExist(err) {
		_, err = os.Open(nodePath)
		if os.IsNotExist(err) {
			return nil, nil
		}
		if err != nil {
			return
		}
		node = []byte(`{"Type":"core.Path"}`)
	}
	if err != nil {
		return
	}
	pathJSON := fmt.Sprintf(`{"Path":%q,`, path)
	node = bytes.Replace(node, []byte("{"), []byte(pathJSON), 1)
	return
}

// getChildren looks up child nodes of the given node.
func getChildren(root, path string) (nodes [][]byte, err error) {
	files, err := ioutil.ReadDir(filepath.Join(root, path))
	if err != nil {
		return
	}
	for _, file := range files {
		node, _ := getNode(root, filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		} else if file.IsDir() {
			nodes = append(nodes,
				[]byte(fmt.Sprintf(`{"Path":%q,"Type":"core.Path"}`,
					filepath.Join(path, file.Name()))))
		}
	}
	return
}

type GetChildrenArgs struct {
	Site, Path string
}

func (i *MonstiService) GetChildren(args GetChildrenArgs,
	reply *[][]byte) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	ret, err := getChildren(site, args.Path)
	*reply = ret
	return err
}

type GetNodeArgs struct{ Site, Path string }

func (i *MonstiService) GetNode(args *GetNodeDataArgs,
	reply *[]byte) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	ret, err := getNode(site, args.Path)
	*reply = ret
	return err
}

type GetNodeDataArgs struct{ Site, Path, File string }

func (i *MonstiService) GetNodeData(args *GetNodeDataArgs,
	reply *[]byte) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	path := filepath.Join(site, args.Path[1:], filepath.Base(args.File))
	ret, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		*reply = nil
		return nil
	}
	*reply = ret
	return err
}

type WriteNodeDataArgs struct {
	Site, Path, File string
	Content          []byte
}

func (i *MonstiService) WriteNodeData(args *WriteNodeDataArgs,
	reply *int) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	path := filepath.Join(site, args.Path[1:], filepath.Base(args.File))
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return fmt.Errorf("Could not create node directory: %v", err)
	}
	err = ioutil.WriteFile(path, []byte(args.Content), 0600)
	if err != nil {
		return fmt.Errorf("Could not write node data: %v", err)
	}
	return nil
}

type RemoveNodeDataArgs struct {
	Site, Path, File string
}

func (i *MonstiService) RemoveNodeData(args *RemoveNodeDataArgs,
	reply *int) error {
	site := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	path := filepath.Join(site, args.Path[1:], filepath.Base(args.File))
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("Could not remove node data: %v", err)
	}
	return nil
}

type RemoveNodeArgs struct {
	Site, Node string
}

func (i *MonstiService) RemoveNode(args *RemoveNodeArgs, reply *int) error {
	root := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	cacheRoot := i.Settings.Monsti.GetSiteCachePath(args.Site)
	nodePath := filepath.Join(root, args.Node[1:])
	// Mark all reverse deps.
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "node.json" {
			cacheNodePath, err := filepath.Rel(root, filepath.Dir(path))
			if err != nil {
				return err
			}
			rdeps, err := readRdeps(cacheRoot, "/"+cacheNodePath)
			if err != nil {
				return err
			}
			for _, rdep := range rdeps {
				err := markDep(cacheRoot, rdep.Dep, 0)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := filepath.Walk(nodePath, walker); err != nil {
		return fmt.Errorf("Could not walk to be removed subtree: %v", err)
	}
	if err := os.RemoveAll(nodePath); err != nil {
		return fmt.Errorf("Can't remove node: %v", err)
	}
	return nil
}

type RenameNodeArgs struct {
	Site, Source, Target string
}

func (i *MonstiService) RenameNode(args *RenameNodeArgs, reply *int) error {
	root := i.Settings.Monsti.GetSiteNodesPath(args.Site)
	if err := os.MkdirAll(
		filepath.Dir(filepath.Join(root, args.Target)), 0700); err != nil {
		return fmt.Errorf("Can't create parent directory: %v", err)
	}
	if err := os.Rename(
		filepath.Join(root, args.Source),
		filepath.Join(root, args.Target)); err != nil {
		return fmt.Errorf("Can't move node: %v", err)
	}
	return nil
}

// getConfig returns the configuration value or section for the given name.
// If the file does not exist, it returns a nil slice.
func getConfig(path, name string) ([]byte, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Could not read configuration: %v", err)
	}
	var target interface{}
	err = json.Unmarshal(content, &target)
	if err != nil {
		return nil, fmt.Errorf("Could not parse configuration: %v", err)
	}
	subs := strings.Split(name, ".")
	for _, sub := range subs {
		if sub == "" {
			break
		}
		targetT := reflect.TypeOf(target)
		if targetT != reflect.TypeOf(map[string]interface{}{}) {
			target = nil
			break
		}
		var ok bool
		if target, ok = (target.(map[string]interface{}))[sub]; !ok {
			target = nil
			break
		}
	}
	target = map[string]interface{}{"Value": target}
	ret, err := json.Marshal(target)
	if err != nil {
		return nil, fmt.Errorf("Could not encode configuration: %v", err)
	}
	return ret, nil
}

type GetSiteConfigArgs struct{ Site, Name string }

func (i *MonstiService) GetSiteConfig(args *GetSiteConfigArgs,
	reply *[]byte) error {
	configPath := i.Settings.Monsti.GetSiteConfigPath(args.Site)
	parts := strings.SplitN(args.Name, ".", 2)
	module := parts[0]
	name := parts[1]
	config, err := getConfig(filepath.Join(configPath, module+".json"), name)
	if err != nil {
		reply = nil
		return err
	}
	*reply = config
	return nil
}

func findAddableNodeTypes(nodeType string,
	nodeTypes map[string]*service.NodeType) []string {
	types := make([]string, 0)
	for _, otherNodeType := range nodeTypes {
		for _, addableTo := range otherNodeType.AddableTo {
			if addableTo == "." ||
				addableTo == nodeType || (addableTo[len(addableTo)-1] == '.' &&
				nodeType[0:len(addableTo)] == addableTo) {
				types = append(types, otherNodeType.Id)
				break
			}
		}
	}
	return types
}

type GetAddableNodeTypesArgs struct{ Site, NodeType string }

func (i *MonstiService) GetAddableNodeTypes(args GetAddableNodeTypesArgs,
	types *[]string) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	*types = findAddableNodeTypes(args.NodeType, i.Settings.Config.NodeTypes)
	return nil
}

func (i *MonstiService) GetNodeType(nodeTypeID string,
	ret *service.NodeType) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	if nodeType, ok := i.Settings.Config.NodeTypes[nodeTypeID]; ok {
		*ret = *nodeType
		return nil
	}
	return fmt.Errorf("Unknown node type %q", nodeTypeID)
}

func (m *MonstiService) RegisterNodeType(nodeType *service.NodeType,
	reply *int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.Settings.Config.NodeTypes[nodeType.Id]; ok {
		return fmt.Errorf("Node type with id %v does already exist", nodeType.Id)
	}
	if m.Settings.Config.NodeTypes == nil {
		m.Settings.Config.NodeTypes = make(map[string]*service.NodeType)
		m.Settings.Config.NodeFields = make(map[string]*service.NodeField)
	}
	m.Settings.Config.NodeTypes[nodeType.Id] = nodeType
	for i, field := range nodeType.Fields {
		if existing, ok := m.Settings.Config.NodeFields[field.Id]; ok {
			nodeType.Fields[i] = existing
		} else {
			m.Settings.Config.NodeFields[field.Id] = field
		}
	}
	return nil
}

func (i *MonstiService) GetRequest(id uint, req *service.Request) error {
	if r := i.Handler.GetRequest(id); r != nil {
		*req = *r
	}
	return nil
}

type cacheData struct {
	CacheMods *service.CacheMods
	Data      []byte
}

func fromCache(root, node, id string) ([]byte, *service.CacheMods, error) {
	path := filepath.Join(root, node[1:], ".data",
		filepath.Base(id))
	raw, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	dec := gob.NewDecoder(bytes.NewReader(raw))
	var data cacheData
	if err = dec.Decode(&data); err != nil {
		return nil, nil, fmt.Errorf("Could not decode cache data: %v", err)
	}
	if !data.CacheMods.Expire.IsZero() &&
		data.CacheMods.Expire.Before(time.Now()) {
		return nil, nil, nil
	}
	return data.Data, data.CacheMods, err
}

type FromCacheArgs struct {
	Node, Site, Id string
}

type FromCacheRet struct {
	CacheMods *service.CacheMods
	Data      []byte
}

func (i *MonstiService) FromCache(args *FromCacheArgs,
	reply *FromCacheRet) error {
	cacheRoot := i.Settings.Monsti.GetSiteCachePath(args.Site)
	var err error
	content, mods, err := fromCache(cacheRoot, args.Node, args.Id)
	*reply = FromCacheRet{mods, content}
	return err
}

type CacheDepPair struct {
	Dep   service.CacheDep
	RDeps []service.CacheDep
}

type CacheDepMap []CacheDepPair

func readRdeps(root, node string) (CacheDepMap, error) {
	rdepsPath := filepath.Join(root, node[1:], ".rdeps.json")
	content, err := ioutil.ReadFile(rdepsPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("Could not read rdeps: %v", err)
	}
	if content == nil {
		return nil, nil
	}
	var depMap CacheDepMap
	err = json.Unmarshal(content, &depMap)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal rdeps: %v", err)
	}
	return depMap, nil
}

func writeRdeps(root, node string, rdeps CacheDepMap) error {
	content, err := json.MarshalIndent(rdeps, "", "  ")
	if err != nil {
		return fmt.Errorf("Could not marshal rdeps: %v", err)
	}
	rdepsPath := filepath.Join(root, node[1:], ".rdeps.json")
	if err := os.MkdirAll(filepath.Dir(rdepsPath), 0700); err != nil {
		return fmt.Errorf("Could not create node cache directory: %v", err)
	}
	if err := ioutil.WriteFile(rdepsPath, content, 0600); err != nil {
		return fmt.Errorf("Could not write rdeps: %v", err)
	}
	return nil
}

func appendRdeps(root string, dep service.CacheDep,
	rdeps []service.CacheDep) error {
	depMap, err := readRdeps(root, dep.Node)
	if err != nil {
		return fmt.Errorf("Could not read rdeps: %v", err)
	}
	depMap = append(depMap, CacheDepPair{dep, rdeps})
	if err := writeRdeps(root, dep.Node, depMap); err != nil {
		return fmt.Errorf("Could not write rdeps: %v", err)
	}
	return nil
}

func toCache(root, node, id string, content []byte,
	mods *service.CacheMods) error {
	// Write deps to filesystem.
	thisDep := service.CacheDep{Node: node, Cache: id}
	if mods != nil {
		for _, dep := range mods.Deps {
			err := appendRdeps(root, dep, []service.CacheDep{thisDep})
			if err != nil {
				return fmt.Errorf("Could not write rdeps: %v", err)
			}
		}
	}

	// Write cache to filesystem.
	nodePath := filepath.Join(root, node[1:])
	path := filepath.Join(nodePath, ".data", filepath.Base(id))
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("Could not create node cache directory: %v", err)
	}
	if mods != nil {
		mods.Deps = nil
	}
	var raw bytes.Buffer
	enc := gob.NewEncoder(&raw)
	data := cacheData{Data: content, CacheMods: mods}
	if err := enc.Encode(&data); err != nil {
		return fmt.Errorf("Could not encode cache data: %v", err)
	}
	if err := ioutil.WriteFile(path, raw.Bytes(), 0600); err != nil {
		return fmt.Errorf("Could not write node cache: %v", err)
	}

	return nil
}

type ToCacheArgs struct {
	Node, Site, Id string
	Content        []byte
	Mods           *service.CacheMods
}

func (i *MonstiService) ToCache(args *ToCacheArgs, reply *int) error {
	cacheRoot := i.Settings.Monsti.GetSiteCachePath(args.Site)
	return toCache(cacheRoot, args.Node, args.Id, args.Content, args.Mods)
}

func markDep(root string, dep service.CacheDep, level int) error {
	rdeps, err := readRdeps(root, dep.Node)
	if err != nil {
		return fmt.Errorf("Could not read rdeps: %v", err)
	}
	if dep.Cache != "" {
		path := filepath.Join(root, dep.Node[1:], ".data", dep.Cache)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("Could not remove cached data: %v", err)
		}
	}
	toBeMarked := make([]service.CacheDep, 0)
	var newDeps CacheDepMap
	dep.Node = filepath.Clean(dep.Node)
	for _, rdep := range rdeps {
		descend := rdep.Dep.Descend
		rdep.Dep.Descend = 0
		rdep.Dep.Node = filepath.Clean(rdep.Dep.Node)
		if descend == -1 || descend >= level && rdep.Dep == dep {
			toBeMarked = append(toBeMarked, rdep.RDeps...)
		} else {
			rdep.Dep.Descend = descend
			newDeps = append(newDeps, rdep)
		}
	}
	if err := writeRdeps(root, dep.Node, newDeps); err != nil {
		return fmt.Errorf("Could not write new rdeps: %v", err)
	}
	for _, dep := range toBeMarked {
		markDep(root, dep, 0)
	}

	if dep.Node != "/" {
		dep.Node = path.Dir(dep.Node)
		if err := markDep(root, dep, level+1); err != nil {
			return fmt.Errorf("Could not mark parent: %v", err)
		}
	}
	return nil
}

type MarkDepArgs struct {
	Site string
	Dep  service.CacheDep
}

func (i *MonstiService) MarkDep(args *MarkDepArgs, reply *int) error {
	cacheRoot := i.Settings.Monsti.GetSiteCachePath(args.Site)
	return markDep(cacheRoot, args.Dep, 0)
}
