package main

import (
	"bytes"
	"github.com/gorilla/sessions"
	"github.com/monsti/monsti-daemon/worker"
	"github.com/monsti/rpc/types"
	utesting "github.com/monsti/util/testing"
	"io/ioutil"
	"path/filepath"
	"testing"
)

// setupRPC creates a RPC environment for testing.
func setupRPC(t *testing.T, testName string) (NodeRPC, string, func()) {
	root, cleanup, err := utesting.CreateDirectoryTree(map[string]string{
		"/foo/__empty__": ""}, testName)
	if err != nil {
		t.Fatalf("Could not create directory tree: ", err)
	}
	site_ := site{Name: "FooSite"}
	site_.Directories.Data = root
	settings := settings{Sites: map[string]site{site_.Name: site_}}
	ticket := worker.Ticket{Site: site_.Name}
	worker := worker.Worker{Ticket: &ticket}
	session := sessions.Session{}
	return NodeRPC{&worker, &settings, &session, nil}, root, cleanup
}

func TestRPCWriteNodeData(t *testing.T) {
	rpc, root, cleanup := setupRPC(t, "TestRPCWriteNodeData")
	defer cleanup()
	var reply int
	rpc.WriteNodeData(&types.WriteNodeDataArgs{
		Path: "/foo", File: "test.txt", Content: "Hey World!"}, &reply)
	writtenData, err := ioutil.ReadFile(filepath.Join(root, "/foo/test.txt"))
	if err != nil {
		t.Fatalf("Could not read file which should have been written: ", err)
	}
	if !bytes.Equal(writtenData, []byte("Hey World!")) {
		t.Fatalf("Written data is %q, should be \"Hey World!\"", writtenData)
	}
}
