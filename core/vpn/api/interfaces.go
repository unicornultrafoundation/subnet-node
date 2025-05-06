package api

import (
	"context"

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

// StreamService handles stream creation and management
type StreamService interface {
	// CreateNewVPNStream creates a new stream to a peer
	CreateNewVPNStream(ctx context.Context, peerID peer.ID) (network.Stream, error)
}
