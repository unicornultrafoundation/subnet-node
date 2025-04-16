package vpn

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"time"

	"github.com/gogo/protobuf/proto"
	subnet_vpn "github.com/unicornultrafoundation/subnet-node/proto/subnet/vpn"
)

// createVirtualIPHash generates a hash of the timestamp and IP address
func createVirtualIPHash(info *subnet_vpn.VirtualIPInfo) []byte {
	// Marshal the data to protobuf
	data, err := proto.Marshal(info)
	if err != nil {
		log.Errorf("Failed to marshal IP info: %v", err)
		return nil
	}

	// Create a SHA-256 hash
	hash := sha256.Sum256(data)
	return hash[:]
}

// signVirtualIPHash signs the hash using the node's private key
func (s *Service) signVirtualIPHash(hash []byte) ([]byte, error) {
	// Get the private key from the host
	privKey := s.PeerHost.Peerstore().PrivKey(s.PeerHost.ID())
	if privKey == nil {
		return nil, fmt.Errorf("private key not found for peer ID %s", s.PeerHost.ID())
	}

	// Sign the hash
	signature, err := privKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %v", err)
	}

	return signature, nil
}

// Save Virtual IP - PeerID mapping to DHT
func (s *Service) StoreMappingInDHT(ctx context.Context, peerID string) error {
	if s.virtualIP == "" {
		return fmt.Errorf("virtual IP is not set")
	}

	// Create timestamp
	timestamp := time.Now().Unix()

	// Create the info
	info := &subnet_vpn.VirtualIPInfo{
		Ip:        s.virtualIP,
		Timestamp: timestamp,
	}

	// Create hash of timestamp and IP
	hash := createVirtualIPHash(info)
	if hash == nil {
		return fmt.Errorf("failed to create hash of IP and timestamp")
	}

	// Sign the hash
	signature, err := s.signVirtualIPHash(hash)
	if err != nil {
		log.Warnf("Failed to sign virtual IP hash: %v", err)
		// Continue without signature if signing fails
	}

	// Create the record
	ipRecord := &subnet_vpn.VirtualIPRecord{
		Info:      info,
		Signature: signature,
	}

	// Serialize the record
	recordData, err := proto.Marshal(ipRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal IP record: %v", err)
	}

	// Store in DHT
	peerIDKey := "/vpn/mapping/" + peerID
	err = s.DHT.PutValue(ctx, peerIDKey, recordData)
	if err != nil {
		return fmt.Errorf("failed to store mapping: %v", err)
	}

	return nil
}

// Check PeerID - Virtual Ip exists
func (s *Service) IsMappingExistedInDHT(ctx context.Context, peerID string, virtualIP string) (bool, error) {
	virtualIPDHT, err := s.GetVirtualIP(ctx, peerID)
	if err != nil {
		return false, err
	}

	peerIDDHT, err := s.GetPeerID(ctx, virtualIP)
	if err != nil {
		return false, err
	}

	if len(virtualIPDHT) == 0 || len(peerIDDHT) == 0 {
		return false, nil
	}

	return virtualIP == virtualIPDHT && peerID == peerIDDHT, nil
}

func (s *Service) GetVirtualIP(ctx context.Context, peerID string) (string, error) {
	peerIDKey := "/vpn/mapping/" + peerID
	recordData, err := s.DHT.GetValue(ctx, peerIDKey)
	if err != nil {
		return "", fmt.Errorf("failed to get virtual IP from DHT: %v", err)
	}

	// Try to parse as a protobuf VirtualIPRecord first
	record := &subnet_vpn.VirtualIPRecord{}
	if err := proto.Unmarshal(recordData, record); err == nil && record.Info != nil && record.Info.Ip != "" {
		// Successfully parsed as a protobuf record
		return record.Info.Ip, nil
	}

	// Fall back to treating it as a plain string (for backward compatibility)
	return string(recordData), nil
}

// Sync current PeerID
func (s *Service) SyncPeerIDToDHT(ctx context.Context) error {
	peerID := s.PeerHost.ID().String()

	// Check if PeerID has already had a virtual IP on DHT
	// If yes, using that virtual IP
	virtualIP, err := s.GetVirtualIP(ctx, peerID)
	if err == nil && virtualIP != "" {
		log.Infof("Use virtual IP %s for peer ID %s", virtualIP, peerID)
		s.virtualIP = virtualIP
		return nil
	}

	// Check if PeerID has already had a virtual IP on registry
	if s.virtualIP == "" {
		return fmt.Errorf("user needs to provide a virtual IP first")
	}

	// Check if the virtual IP is already mapped to the current peer ID
	peerIDMapped, err := s.GetPeerID(ctx, s.virtualIP)
	if err == nil && peerIDMapped == peerID {
		log.Infof("Use new virtual IP %s for peer ID %s", virtualIP, peerID)
		return s.StoreMappingInDHT(ctx, peerID)
	}

	// User hasn't registered a virtual IP yet, inform user to register one
	return fmt.Errorf("no virtual IP found for peer ID %s", peerID)
}

func (s *Service) GetPeerID(ctx context.Context, virtualIP string) (string, error) {
	// Get from cache first
	if peerID, ok := s.peerIDCache[virtualIP]; ok {
		return peerID, nil
	}

	peerID, err := s.GetPeerIDByRegistry(ctx, virtualIP)
	if err != nil {
		return "", fmt.Errorf("failed to get peer ID from registry: %v", err)
	}

	s.peerIDCache[virtualIP] = peerID
	return peerID, nil
}

func (s *Service) GetPeerIDByRegistry(ctx context.Context, virtualIP string) (string, error) {
	tokenID := ConvertVirtualIPToNumber(virtualIP)
	if tokenID == 0 {
		return "", fmt.Errorf("IP %s is not within the range 10.0.0.0/8", virtualIP)
	}

	return s.accountService.IPRegistry().GetPeer(nil, big.NewInt(int64(tokenID)))
}

// Auto clear Peer ID cache each 30 seconds
func (s *Service) startClearingPeerIDCache(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.peerIDCache = make(map[string]string)
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}
