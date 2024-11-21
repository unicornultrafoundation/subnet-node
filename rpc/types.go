package rpc

import "context"

// ServerCodec implements reading, parsing and writing RPC messages for the server side of
// an RPC session. Implementations must be go-routine safe since the codec can be called in
// multiple go-routines concurrently.
type ServerCodec interface {
	peerInfo() PeerInfo
	readBatch() (msgs []*jsonrpcMessage, isBatch bool, err error)
	close()

	jsonWriter
}

// jsonWriter can write JSON messages to its underlying connection.
// Implementations must be safe for concurrent use.
type jsonWriter interface {
	// writeJSON writes a message to the connection.
	writeJSON(ctx context.Context, msg interface{}, isError bool) error

	// Closed returns a channel which is closed when the connection is closed.
	closed() <-chan interface{}
	// RemoteAddr returns the peer address of the connection.
	remoteAddr() string
}
