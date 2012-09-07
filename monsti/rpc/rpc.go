package rcp

// Connection represents the rpc connection.
type Connection int

// NewConnection establishes a new rpc connection and registers the content
// type.
func NewConnection(nodeType string) Connection {
}

// GetRequests blocks until it receives a request to process.
func (c Connection) GetRequest() Request {
}

// GetNodeData requests data from some node.
func (c Connection) GetNodeData(string path, string file) []byte {
}

/// SendResponse sends the repsonse of an earlier request.
func (c Connection) SendResponse(r Response) {
}


