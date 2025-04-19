package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("service", "vpn-discovery")

// PeerDiscoveryService handles peer discovery and mapping
type PeerDiscoveryService interface {
	// GetPeerID gets the peer ID for a destination IP
	GetPeerID(ctx context.Context, destIP string) (string, error)
	// SyncPeerIDToDHT syncs the peer ID to the DHT
	SyncPeerIDToDHT(ctx context.Context) error
}

// PeerDiscovery implements the PeerDiscoveryService interface
type PeerDiscovery struct {
	// Host for P2P communication
	host p2phost.Host
	// DHT for peer discovery
	dht *ddht.DHT
	// Cache for peer ID lookups
	peerIDCache *cache.Cache
	// Virtual IP for this node
	virtualIP string
	// Mutex for thread safety
	mu sync.Mutex
}

// NewPeerDiscovery creates a new peer discovery service
func NewPeerDiscovery(host p2phost.Host, dht *ddht.DHT, virtualIP string) *PeerDiscovery {
	return &PeerDiscovery{
		host:        host,
		dht:         dht,
		peerIDCache: cache.New(1*time.Minute, 30*time.Second),
		virtualIP:   virtualIP,
	}
}

// GetPeerID gets the peer ID for a destination IP
func (p *PeerDiscovery) GetPeerID(ctx context.Context, destIP string) (string, error) {
	// Check cache first
	if cachedPeerID, found := p.peerIDCache.Get(destIP); found {
		return cachedPeerID.(string), nil
	}

	// Not in cache, look up in DHT
	peerID, err := p.lookupPeerIDInDHT(ctx, destIP)
	if err != nil {
		return "", err
	}

	// Cache the result
	p.peerIDCache.Set(destIP, peerID, cache.DefaultExpiration)

	return peerID, nil
}

// lookupPeerIDInDHT looks up a peer ID in the DHT
func (p *PeerDiscovery) lookupPeerIDInDHT(ctx context.Context, destIP string) (string, error) {
	// Create a key for the DHT lookup
	key := fmt.Sprintf("/vpn/ip/%s", destIP)

	// Look up the key in the DHT
	value, err := p.dht.GetValue(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get peer ID from DHT: %v", err)
	}

	if len(value) == 0 {
		return "", fmt.Errorf("no peer ID found for IP %s", destIP)
	}

	return string(value), nil
}

// SyncPeerIDToDHT syncs the peer ID to the DHT
func (p *PeerDiscovery) SyncPeerIDToDHT(ctx context.Context) error {
	if p.virtualIP == "" {
		return fmt.Errorf("virtual IP is not set")
	}

	// Create a key for the DHT
	key := fmt.Sprintf("/vpn/ip/%s", p.virtualIP)

	// Put the peer ID in the DHT
	err := p.dht.PutValue(ctx, key, []byte(p.host.ID().String()))
	if err != nil {
		return fmt.Errorf("failed to put peer ID in DHT: %v", err)
	}

	log.Infof("Synced peer ID %s for IP %s to DHT", p.host.ID().String(), p.virtualIP)
	return nil
}

// ParsePeerID parses a peer ID string
func ParsePeerID(peerIDStr string) (peer.ID, error) {
	return peer.Decode(peerIDStr)
}
