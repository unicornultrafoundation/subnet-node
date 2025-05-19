package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	subnet_vpn "github.com/unicornultrafoundation/subnet-node/proto/subnet/vpn"
)

// PeerIDMappingPrefix is the DHT key prefix for peer ID mappings
const (
	PeerIDMappingPrefix = "/vpn/mapping/"
)

// GetVirtualIP gets the virtual IP for a peer ID from the DHT
func (p *PeerDiscovery) GetVirtualIP(ctx context.Context, peerID string) (string, error) {
	peerIDKey := PeerIDMappingPrefix + peerID
	recordData, err := p.dht.GetValue(ctx, peerIDKey)
	if err != nil {
		return "", fmt.Errorf("failed to get virtual IP from DHT for peer ID %s: %w", peerID, err)
	}

	if len(recordData) == 0 {
		return "", fmt.Errorf("empty data returned from DHT for peer ID %s", peerID)
	}

	// Try to parse as a protobuf record first
	record := &subnet_vpn.VirtualIPRecord{}
	if err := proto.Unmarshal(recordData, record); err == nil && record.Info != nil && record.Info.Ip != "" {
		// Successfully parsed as a protobuf record
		return record.Info.Ip, nil
	}

	// Fall back to treating it as a plain string
	return string(recordData), nil
}

// StoreMappingInDHT stores the virtual IP to peer ID mapping in the DHT
func (p *PeerDiscovery) StoreMappingInDHT(ctx context.Context, peerID string) error {
	currentVirtualIP := p.virtualIP

	if currentVirtualIP == "" {
		return fmt.Errorf("virtual IP is not set for peer ID %s", peerID)
	}

	// Create the record
	ipRecord, err := p.createSignedIPRecord(currentVirtualIP)
	if err != nil {
		return fmt.Errorf("failed to create signed IP record for peer ID %s: %w", peerID, err)
	}

	// Serialize the record
	recordData, err := proto.Marshal(ipRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal IP record for peer ID %s: %w", peerID, err)
	}

	// Store in DHT
	peerIDKey := PeerIDMappingPrefix + peerID
	err = p.dht.PutValue(ctx, peerIDKey, recordData)
	if err != nil {
		return fmt.Errorf("failed to store mapping in DHT for peer ID %s: %w", peerID, err)
	}

	log.Infof("Synced peer ID %s for IP %s to DHT", peerID, currentVirtualIP)
	return nil
}

// createSignedIPRecord creates a signed record
func (p *PeerDiscovery) createSignedIPRecord(virtualIP string) (*subnet_vpn.VirtualIPRecord, error) {
	// Create timestamp for the record
	timestamp := time.Now().Unix()

	// Create the info for the record
	info := &subnet_vpn.VirtualIPInfo{
		Ip:        virtualIP,
		Timestamp: timestamp,
	}

	// Create hash of timestamp and IP for the record
	hash := p.createVirtualIPHash(info)
	if hash == nil {
		return nil, fmt.Errorf("failed to create hash of IP %s and timestamp %d", virtualIP, timestamp)
	}

	// Sign the hash for the record
	signature, err := p.signVirtualIPHash(hash)
	if err != nil {
		log.Warnf("Failed to sign virtual IP hash for IP %s: %v", virtualIP, err)
	}

	// Create the record
	ipRecord := &subnet_vpn.VirtualIPRecord{
		Info:      info,
		Signature: signature,
	}

	return ipRecord, nil
}
