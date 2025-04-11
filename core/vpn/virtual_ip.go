package vpn

import (
	"context"
	"fmt"
)

// Save Virtual IP - PeerID mapping to DHT
func (s *Service) StoreMappingInDHT(ctx context.Context, peerID string) error {
	peerIDKey := "/vpn/mapping/" + peerID
	err := s.DHT.PutValue(ctx, peerIDKey, []byte(s.virtualIP))
	if err != nil {
		return fmt.Errorf("failed to store mapping: %v", err)
	}

	virtualIPKey := "/vpn/mapping/" + s.virtualIP
	err = s.DHT.PutValue(ctx, virtualIPKey, []byte(peerID))
	if err != nil {
		log.Errorln("Failed to store mapping:", err)
		err = s.DHT.PutValue(ctx, peerIDKey, []byte{}) // reset
		if err != nil {
			return fmt.Errorf("failed to store mapping: %v", err)
		}
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
	virtualIPDHT, err := s.DHT.GetValue(ctx, peerIDKey)
	if err != nil {
		return "", fmt.Errorf("failed to get virtual IP from DHT: %v", err)
	}

	return string(virtualIPDHT), nil
}

func (s *Service) GetPeerID(ctx context.Context, virtualIP string) (string, error) {
	virtualIPKey := "/vpn/mapping/" + virtualIP
	peerIDDHT, err := s.DHT.GetValue(ctx, virtualIPKey)
	if err != nil {
		return "", fmt.Errorf("failed to get peerID from DHT: %v", err)
	}

	return string(peerIDDHT), nil
}

// Sync current PeerID
func (s *Service) SyncPeerIDToDHT(ctx context.Context) error {
	peerID := s.PeerHost.ID().String()

	// Check if PeerID has already had a virtual IP on DHT
	// If yes, using that virtual IP, ignoring the one user provided
	virtualIP, err := s.GetVirtualIP(ctx, peerID)
	if err == nil {
		log.Infof("Use virtual IP %s for peer ID %s", virtualIP, peerID)
		s.virtualIP = virtualIP
		return nil
	}

	// Check if VirtualIP user provides is available on DHT
	if len(s.virtualIP) != 0 {
		mappingPeerID, err := s.GetPeerID(ctx, s.virtualIP)
		if err == nil {
			if mappingPeerID == peerID {
				log.Infof("Already registered peerID %s with virtual IP %s", peerID, s.virtualIP)
				return nil
			} else {
				log.Infof("Virtual IP %s user provided is used by peer ID %s, finding new available one...", s.virtualIP, mappingPeerID)
				s.virtualIP = ""
				for s.virtualIP == "" {
					s.virtualIP = s.findNextAvailableIP(ctx, peerID)
				}
			}
		} else {
			// Virtual IP user provided is available, use this
		}
	} else {
		// User doesn't provide virtual IP, find new available one...
		for s.virtualIP == "" {
			s.virtualIP = s.findNextAvailableIP(ctx, peerID)
		}
	}

	log.Infof("Registering peerID %s with virtual IP %s", peerID, s.virtualIP)

	// Store mappings
	return s.StoreMappingInDHT(ctx, peerID)
}

func (s *Service) findNextAvailableIP(ctx context.Context, peerID string) string {
	if len(s.routes) == 0 {
		return ""
	}
	for _, cidr := range s.routes {
		netIP, err := GenerateVirtualIP(cidr)
		if err != nil {
			log.Fatalf("the format of vpn.routes is not valid: %v", err)
		}
		ip := netIP.String()
		if exists, _ := s.IsMappingExistedInDHT(ctx, peerID, ip); !exists { // Ensure uniqueness
			return ip
		}
	}

	return "" // Handle edge case where all IPs are taken
}
