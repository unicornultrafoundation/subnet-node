package p2p

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	pbstream "github.com/unicornultrafoundation/subnet-node/common/io"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
)

// SignProtocolListener handles requests for signing resource usage data
type SignProtocolListener struct {
	protoID       protocol.ID
	listenAddr    ma.Multiaddr
	targetAddr    ma.Multiaddr
	signHandler   func(stream network.Stream) ([]byte, error)
	activeStreams sync.Map // Tracks active streams
}

// SignatureResponse represents the response containing the signature
type SignatureResponse struct {
	Signature string `json:"signature"`
}

// NewSignProtocolListener creates a new listener for the signing protocol
func NewSignProtocolListener(protoID protocol.ID, signHandler func(stream network.Stream) ([]byte, error)) *SignProtocolListener {
	return &SignProtocolListener{
		protoID:     protoID,
		signHandler: signHandler,
	}
}

// Protocol returns the protocol ID
func (l *SignProtocolListener) Protocol() protocol.ID {
	return l.protoID
}

// ListenAddress returns the listener's address
func (l *SignProtocolListener) ListenAddress() ma.Multiaddr {
	return l.listenAddr
}

// TargetAddress returns the target address
func (l *SignProtocolListener) TargetAddress() ma.Multiaddr {
	return l.targetAddr
}

// key returns the protocol ID as the listener's key
func (l *SignProtocolListener) key() protocol.ID {
	return l.protoID
}

// close stops the listener
func (l *SignProtocolListener) close() {
	log.Printf("Closing listener for protocol: %s", l.protoID)
	l.activeStreams.Range(func(key, value interface{}) bool {
		stream := value.(network.Stream)
		stream.Reset()
		return true
	})
	l.activeStreams = sync.Map{}
}

// handleStream processes incoming streams for the signing protocol
func (l *SignProtocolListener) handleStream(stream network.Stream) {
	defer stream.Close()

	// Add stream to active streams
	l.activeStreams.Store(stream.Conn().RemotePeer(), stream)

	// Process the signing request
	signature, err := l.signHandler(stream)
	if err != nil {
		log.Printf("Failed to sign resource usage: %v", err)
		return
	}

	// Send back the signature as a response
	response := pbapp.SignatureResponse{Signature: signature}
	if err := pbstream.WriteProtoBuffered(stream, &response); err != nil {
		log.Printf("Failed to send signature response: %v", err)
	}
}
