package api

import (
	"context"
	"io"
	"math/big"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// PeerDiscoveryService handles peer discovery and mapping
type PeerDiscoveryService interface {
	// GetPeerID gets the peer ID for a destination IP
	GetPeerID(ctx context.Context, destIP string) (string, error)
	// SyncPeerIDToDHT syncs the peer ID to the DHT
	SyncPeerIDToDHT(ctx context.Context) error
	// GetVirtualIP gets the virtual IP for a peer ID
	GetVirtualIP(ctx context.Context, peerID string) (string, error)
	// StoreMappingInDHT stores a mapping in the DHT
	StoreMappingInDHT(ctx context.Context, peerID string) error
	// VerifyVirtualIPHasRegistered verifies if a virtual IP is registered to the current peer ID
	VerifyVirtualIPHasRegistered(ctx context.Context, virtualIP string) error
	// GetPeerIDByRegistry gets a peer ID from the registry
	GetPeerIDByRegistry(ctx context.Context, destIP string) (string, error)
}

// VPNStream is an interface for the network.Stream to make it easier to mock
type VPNStream interface {
	io.ReadWriteCloser
	Reset() error
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Ensure network.Stream implements VPNStream
var _ VPNStream = (network.Stream)(nil)

// StreamService handles stream creation and management
type StreamService interface {
	// CreateNewVPNStream creates a new stream to a peer
	CreateNewVPNStream(ctx context.Context, peerID peer.ID) (VPNStream, error)
}

// DHTService is an interface for DHT operations
type DHTService interface {
	// GetValue gets a value from the DHT
	GetValue(ctx context.Context, key string) ([]byte, error)
	// PutValue puts a value in the DHT
	PutValue(ctx context.Context, key string, value []byte) error
}

// PeerstoreService is an interface for the peerstore service
type PeerstoreService interface {
	// PrivKey returns the private key for a peer ID
	PrivKey(p peer.ID) crypto.PrivKey
}

// HostService is an interface for the host service
type HostService interface {
	// ID returns the peer ID of the host
	ID() peer.ID
	// Peerstore returns the peerstore of the host
	Peerstore() PeerstoreService
}

// AccountService is an interface for the account service
type AccountService interface {
	// IPRegistry returns the IP registry contract
	IPRegistry() IPRegistry
}

// IPRegistry is an interface for the IP registry contract
type IPRegistry interface {
	// GetPeer gets the peer ID for a token ID
	GetPeer(opts any, tokenID *big.Int) (string, error)
}
