package discovery

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

var log = logrus.WithField("service", "vpn-discovery")

// Ensure PeerDiscovery implements the PeerDiscoveryService interface
var _ api.PeerDiscoveryService = (*PeerDiscovery)(nil)

// PeerDiscovery implements the api.PeerDiscoveryService interface
type PeerDiscovery struct {
	// Host for P2P communication
	host host.Host
	// DHT for peer discovery
	dht routing.ValueStore
	// Cache for peer ID lookups
	peerIDCache *cache.Cache
	// Virtual IP for this node
	virtualIP string
	// Account service for registry lookups
	accountService *account.AccountService
	// Mutex for thread safety
	mu sync.RWMutex
}

// NewPeerDiscovery creates a new peer discovery service
func NewPeerDiscovery(host host.Host, dht routing.ValueStore, virtualIP string, accountService *account.AccountService) *PeerDiscovery {
	return &PeerDiscovery{
		host:           host,
		dht:            dht,
		peerIDCache:    cache.New(1*time.Minute, 30*time.Second),
		virtualIP:      virtualIP,
		accountService: accountService,
	}
}

// GetPeerID gets the peer ID for a destination IP
func (p *PeerDiscovery) GetPeerID(ctx context.Context, destIP string) (string, error) {
	// Check cache first
	p.mu.RLock()
	peerID, found := p.peerIDCache.Get(destIP)
	p.mu.RUnlock()

	if found {
		if peerIDStr, ok := peerID.(string); ok && peerIDStr != "" {
			return peerIDStr, nil
		}
		// Empty string means negative cache hit
		return "", fmt.Errorf("no peer mapping found for IP %s (cached result)", destIP)
	}

	// Try to get from registry
	regPeerID, err := p.GetPeerIDByRegistry(ctx, destIP)
	if err != nil {
		p.mu.Lock()
		p.peerIDCache.Set(destIP, "", 1*time.Minute) // Cache negative result
		p.mu.Unlock()
		return "", fmt.Errorf("failed to get peer ID from registry for IP %s: %w", destIP, err)
	}

	// Cache the result
	p.mu.Lock()
	p.peerIDCache.Set(destIP, regPeerID, cache.DefaultExpiration)
	p.mu.Unlock()

	return regPeerID, nil
}

// SyncPeerIDToDHT syncs the peer ID to the DHT
func (p *PeerDiscovery) SyncPeerIDToDHT(ctx context.Context) error {
	// Get the current peer ID once to avoid multiple calls
	peerID := p.host.ID().String()

	currentVirtualIP := p.virtualIP

	// Check if PeerID has already had a virtual IP on DHT
	// If yes, using that virtual IP
	virtualIP, err := p.GetVirtualIP(ctx, peerID)
	if err == nil && virtualIP != "" {
		// Verify if the virtual IP is registered to the current peer ID
		verifyErr := p.VerifyVirtualIPHasRegistered(ctx, virtualIP)
		if verifyErr != nil {
			log.Warnf("Try using the virtual IP %s user provided since virtual IP %s on DHT is not registered to this peer ID: %v",
				currentVirtualIP, virtualIP, verifyErr)
		} else {
			log.Infof("Use virtual IP %s on DHT for peer ID %s", virtualIP, peerID)
			// Update the virtual IP atomically
			p.mu.Lock()
			p.virtualIP = virtualIP
			p.mu.Unlock()
			return nil
		}
	} else if err != nil {
		log.Debugf("Failed to get virtual IP from DHT for peer ID %s: %v", peerID, err)
	}

	// Check if user has registered a virtual IP
	verifyErr := p.VerifyVirtualIPHasRegistered(ctx, currentVirtualIP)
	if verifyErr == nil {
		// User has registered a virtual IP, store the mapping in DHT
		log.Infof("Use new virtual IP %s for peer ID %s", currentVirtualIP, peerID)
		if storeErr := p.StoreMappingInDHT(ctx, peerID); storeErr != nil {
			return fmt.Errorf("failed to store mapping in DHT for peer ID %s with virtual IP %s: %w",
				peerID, currentVirtualIP, storeErr)
		}
		return nil
	}

	// User hasn't registered a virtual IP yet, inform user to register one
	return fmt.Errorf("no virtual IP found for peer ID %s: %w", peerID, verifyErr)
}

// GetPeerIDByRegistry gets the peer ID from the registry
func (p *PeerDiscovery) GetPeerIDByRegistry(ctx context.Context, virtualIP string) (string, error) {
	tokenID := ConvertVirtualIPToNumber(virtualIP)
	if tokenID == 0 {
		return "", fmt.Errorf("IP %s is not within the range 10.0.0.0/8", virtualIP)
	}

	return p.accountService.IPRegistry().GetPeer(nil, big.NewInt(int64(tokenID)))
}

// VerifyVirtualIPHasRegistered verifies if the virtual IP is registered to the current peer ID
func (p *PeerDiscovery) VerifyVirtualIPHasRegistered(ctx context.Context, virtualIP string) error {
	if virtualIP == "" {
		return fmt.Errorf("virtual IP is not set")
	}

	// Get the current peer ID once to avoid multiple calls
	currentPeerID := p.host.ID().String()

	// Get the peer ID from the registry
	peerID, err := p.GetPeerID(ctx, virtualIP)
	if err != nil {
		return fmt.Errorf("failed to get peer ID from registry: %w", err)
	}

	// Compare with the current peer ID
	if peerID != currentPeerID {
		return fmt.Errorf("virtual IP %s is registered to peer ID %s, not to this peer ID %s",
			virtualIP, peerID, currentPeerID)
	}

	return nil
}

// NewPeerDiscoveryFromLibp2p creates a new peer discovery service from libp2p components
func NewPeerDiscoveryFromLibp2p(host host.Host, dht routing.ValueStore, virtualIP string, accountService *account.AccountService) *PeerDiscovery {
	return NewPeerDiscovery(
		host,
		dht,
		virtualIP,
		accountService,
	)
}
