package discovery

import (
	"context"
	"fmt"
	"math/big"
	"net"
)

// ConvertVirtualIPToNumber converts a virtual IP to a token ID
func ConvertVirtualIPToNumber(virtualIP string) uint32 {
	// Parse the IP address
	ip := net.ParseIP(virtualIP)
	if ip == nil {
		return 0
	}

	// Ensure it's an IPv4 address
	ipv4 := ip.To4()
	if ipv4 == nil {
		return 0
	}

	// Check if it's in the 10.0.0.0/8 range
	if ipv4[0] != 10 {
		return 0
	}

	// Convert to a token ID (last 3 octets as a number)
	tokenID := uint32(ipv4[1])<<16 | uint32(ipv4[2])<<8 | uint32(ipv4[3])
	return tokenID
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
