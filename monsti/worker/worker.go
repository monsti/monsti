package worker

import (
	"datenkarussell.de/monsti/rpc/client"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"os/exec"
	"strings"
)

// Ticket represents an incoming request to be processed by the worker.
type Ticket struct {
	// The requested node.
	Node client.Node
	// Request is the original HTTP request.
	Request *http.Request
	// ResponseChan is a channel over which the built respsonse can be send
	// back to the client.
	ResponseChan chan client.Response
	Session      client.Session
	// Action as specified in the URL (/path/to/node/@@some_action).
	Action      string
}

// pipeConnection is a bidirectional pipe to a worker process used for RPC
// communication.
type pipeConnection struct {
	io.ReadCloser
	io.WriteCloser
	Worker *Worker
}

func (pipe pipeConnection) Close() (err error) {
	return nil
}

// workerLog is a Writer used to log incoming worker errors.
type workerLog struct {
	Type string
}

func (w workerLog) Write(p []byte) (int, error) {
	return len(p), nil
}

// Worker represents a process which communicates via RPC over a bidirectional
// pipe with Monsti to process incoming requests for some node type.
type Worker struct {
	// The currently processed ticket (if any).
	Ticket *Ticket
	// Tickets is the channel where new tickets can be fetched.
	Tickets chan Ticket
	// The node type for which this worker processes requests. 
	NodeType string
	// Command of the worker process.
	cmd *exec.Cmd
	// Pipe to the worker process.
	pipe *pipeConnection
	// Receiver for RPC.
	rcvr interface{}
}

// NewWorker creates a new worker for the node type that fetches new tickets
// from the given channele.
//
// The worker process fetches its configuration from the given configuration
// directory. RPC methods are provided to the worker process by the given
// receiver.
func NewWorker(nodeType string, tickets chan Ticket, rcvr interface{},
	configDir string) (w *Worker) {
	w = &Worker{
		Tickets:  tickets,
		NodeType: nodeType,
		rcvr:     rcvr}
	w.cmd = exec.Command(strings.ToLower(nodeType), configDir)
	return
}

// Run starts the worker process in a goroutine and calls the given function
// after the process died.
func (w *Worker) Run(callback func()) error {
	inPipe, err := w.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Could not setup stdout pipe of worker: %v",
			err.Error())
	}
	outPipe, err := w.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Could not setup stdin pipe of worker: %v",
			err.Error())
	}
	w.cmd.Stderr = workerLog{w.NodeType}
	w.pipe = &pipeConnection{inPipe, outPipe, w}
	err = w.cmd.Start()
	if err != nil {
		return fmt.Errorf("Could not setup connection to worker: %v",
			err.Error())
	}
	server := rpc.NewServer()
	if err = server.Register(w.rcvr); err != nil {
		return fmt.Errorf("Could not register RPC methods: %v",
			err.Error())
	}
	go server.ServeConn(w.pipe)
	go func() {
		w.cmd.Wait()
		log.Println("Worker process died.")
		w.postMortem()
		callback()
	}()
	return nil
}

// postMortem gets called after the worker process died. It performs some
// cleanup actions.
func (w *Worker) postMortem() {
	if w.cmd.ProcessState == nil || !w.cmd.ProcessState.Exited() {
		panic("worker: postMortem() called on living worker")
	}
	if w.Ticket != nil {
		close(w.Ticket.ResponseChan)
	}
	w.Ticket = nil
}
