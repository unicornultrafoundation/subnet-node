package pool

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

// ConnectionState represents the state of a connection
type ConnectionState struct {
	Key           types.ConnectionKey
	PeerID        peer.ID
	StreamChannel *StreamChannel
	LastActivity  int64
}

// ConnectionStateV2 represents the state of a connection with V2 stream channel
type ConnectionStateV2 struct {
	Key           types.ConnectionKey
	PeerID        peer.ID
	StreamChannel *StreamChannelV2
	LastActivity  int64
}
