package worker

import (
	"github.com/monsti/rpc/client"
	"io"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"testing"
	"time"
)

type TestRPC struct {
	Worker  *Worker
	Tickets chan Ticket
}

func (t *TestRPC) Foo(arg int, ret *int) error {
	ticket := <-t.Tickets
	t.Worker.Ticket = &ticket
	return nil
}

func TestProcessDead(t *testing.T) {
	rpc := &TestRPC{
		Tickets: make(chan Ticket)}
	logger := log.New(os.Stderr, "", log.LstdFlags)
	worker := NewWorker("true", rpc.Tickets, rpc, "", logger)
	rpc.Worker = worker
	worker.cmd = exec.Command(os.Args[0], "-test.run", "TestDummyWorker")
	worker.cmd.Env = append([]string{"GO_WANT_DUMMY_WORKER=1"}, os.Environ()...)
	callbackCalled := false
	err := worker.Run(func() { callbackCalled = true })
	if err != nil {
		t.Error(err.Error())
	}
	ticket := Ticket{
		ResponseChan: make(chan client.Response)}
	worker.Tickets <- ticket
	select {
	case _, ok := <-ticket.ResponseChan:
		if ok {
			t.Error("Response chan should be closed")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for response")
	}
	if !callbackCalled {
		t.Error("Callback has not been called")
	}
}

type pipe struct {
	io.ReadCloser
	io.WriteCloser
}

func (pipe pipe) Close() (err error) {
	panic("Close() on pipe connection. RPC error?")
}

func TestDummyWorker(t *testing.T) {
	if os.Getenv("GO_WANT_DUMMY_WORKER") != "1" {
		return
	}
	pipe := &pipe{os.Stdin, os.Stdout}
	c := rpc.NewClient(pipe)
	var ret int
	err := c.Call("TestRPC.Foo", 0, &ret)
	if err != nil {
		t.Fatal(err.Error())
	}
}
