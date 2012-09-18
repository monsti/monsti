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
}

// pipeConnection is a bidirectional pipe to a worker process used for RPC
// communication.
type pipeConnection struct {
	io.ReadCloser
	io.WriteCloser
}

func (pipe pipeConnection) Close() (err error) {
	panic("monsti: Close() on pipe connection. RPC error?")
}

// workerLog is a Writer used to log incoming worker errors.
type workerLog struct {
	Type string
}

func (w workerLog) Write(p []byte) (int, error) {
	log.Printf("from %v: %v", w.Type, string(p))
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
	cmd exec.Cmd
	// Pipe to the worker process.
	pipe pipeConnection
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
	configDir string) (w *Worker, err error) {
	w = &Worker{
		Tickets:  tickets,
		NodeType: nodeType,
		rcvr:     rcvr}
	w.cmd = *exec.Command(strings.ToLower(nodeType), configDir)
	inPipe, err := w.cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Could not setup stdout pipe of worker: %v",
			err.Error())
	}
	outPipe, err := w.cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("Could not setup stdin pipe of worker: %v",
			err.Error())
	}
	w.cmd.Stderr = workerLog{nodeType}
	w.pipe = pipeConnection{inPipe, outPipe}
	return
}

// Run starts the worker process in a goroutine.
func (w *Worker) Run() error {
	log.Println("Starting worker for " + w.NodeType)
	err := w.cmd.Start()
	if err != nil {
		return fmt.Errorf("Could not setup connection to worker: %v",
			err.Error())
	}
	server := rpc.NewServer()
	if err = server.Register(w.rcvr); err != nil {
		return fmt.Errorf("Could not register RPC methods: %v",
			err.Error())
	}
	log.Println("Setting up RPC.")
	go server.ServeConn(w.pipe)
	return nil
}
