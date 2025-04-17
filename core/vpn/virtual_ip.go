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
		// Verify if the virtual IP is registered to the current peer ID
		err = s.VerifyVirtualIPHasRegistered(ctx, virtualIP)
		if err != nil {
			log.Warnf("Try using the virtual IP %s user provided since virtual IP %s on DHT is not registered to this peer ID: %v.", s.virtualIP, virtualIP, err)
		} else {
			log.Infof("Use virtual IP %s on DHT for peer ID %s", virtualIP, peerID)
			s.virtualIP = virtualIP
			return nil
		}
	}

	// Check if user has registered a virtual IP
	err = s.VerifyVirtualIPHasRegistered(ctx, s.virtualIP)
	if err == nil {
		// User has registered a virtual IP, store the mapping in DHT
		log.Infof("Use new virtual IP %s for peer ID %s", virtualIP, peerID)
		return s.StoreMappingInDHT(ctx, peerID)
	}

	// User hasn't registered a virtual IP yet, inform user to register one
	return fmt.Errorf("no virtual IP found for peer ID %s", peerID)
}

func (s *Service) GetPeerID(ctx context.Context, virtualIP string) (string, error) {
	// Get from cache first
	s.mu.Lock()
	peerID, found := s.peerIDCache.Get(virtualIP)
	s.mu.Unlock()
	if found {
		return peerID.(string), nil
	}

	peerID, err := s.GetPeerIDByRegistry(ctx, virtualIP)
	if err != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.peerIDCache.Set(virtualIP, "", 1*time.Minute)
		return "", fmt.Errorf("failed to get peer ID from registry: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.peerIDCache.Set(virtualIP, peerID, 1*time.Minute)
	return peerID.(string), nil
}

func (s *Service) GetPeerIDByRegistry(ctx context.Context, virtualIP string) (string, error) {
	tokenID := ConvertVirtualIPToNumber(virtualIP)
	if tokenID == 0 {
		return "", fmt.Errorf("IP %s is not within the range 10.0.0.0/8", virtualIP)
	}

	return s.accountService.IPRegistry().GetPeer(nil, big.NewInt(int64(tokenID)))
}

func (s *Service) VerifyVirtualIPHasRegistered(ctx context.Context, virtualIP string) error {
	if virtualIP == "" {
		return fmt.Errorf("virtual IP is not set")
	}

	peerID, err := s.GetPeerID(ctx, virtualIP)
	if err != nil {
		return fmt.Errorf("failed to get peer ID from registry: %v", err)
	}

	if peerID != s.PeerHost.ID().String() {
		return fmt.Errorf("virtual IP %s is not registered to this peer ID", virtualIP)
	}

	return nil
}
