package vpn

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// Save Virtual IP - PeerID mapping to DHT
func (s *Service) StoreMappingInDHT(peerID string, virtualIP string) error {
	ctx := context.Background()
	peerIDKey := "/vpn/mapping/" + peerID
	err := s.DHT.PutValue(ctx, peerIDKey, []byte(virtualIP))
	if err != nil {
		return fmt.Errorf("failed to store mapping: %v", err)
	}

	virtualIPKey := "/vpn/mapping/" + virtualIP
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
func (s *Service) IsMappingExistedInDHT(peerID string, virtualIP string) (bool, error) {
	virtualIPDHT, err := s.GetVirtualIP(peerID)
	if err != nil {
		return false, err
	}

	peerIDDHT, err := s.GetPeerID(virtualIP)
	if err != nil {
		return false, err
	}

	if len(virtualIPDHT) == 0 || len(peerIDDHT) == 0 {
		return false, nil
	}

	return true, nil
}

func (s *Service) GetVirtualIP(peerID string) (string, error) {
	ctx := context.Background()
	peerIDKey := "/vpn/mapping/" + peerID
	virtualIPDHT, err := s.DHT.GetValue(ctx, peerIDKey)
	if err != nil {
		return "", fmt.Errorf("failed to get virtual IP from DHT: %v", err)
	}

	return string(virtualIPDHT), nil
}

func (s *Service) GetPeerID(virtualIP string) (string, error) {
	ctx := context.Background()

	virtualIPKey := "/vpn/mapping/" + virtualIP
	peerIDDHT, err := s.DHT.GetValue(ctx, virtualIPKey)
	if err != nil {
		return "", fmt.Errorf("failed to get peerID from DHT: %v", err)
	}

	return string(peerIDDHT), nil
}

// Sync current PeerID
func (s *Service) SyncPeerIDToDHT() (string, error) {
	peerID := s.PeerHost.ID().String()

	// Generate a unique IP using a hash
	hash := sha256.Sum256([]byte(peerID))
	lastOctet := (hash[2] % 253) + 2 // Ensures range 2-254

	virtualIP := fmt.Sprintf("%d.%d.%d.%d", s.firstOctet, hash[0], hash[1], lastOctet)

	// Resolve collisions (ensure uniqueness)
	exist, _ := s.IsMappingExistedInDHT(peerID, virtualIP)
	if exist { // Already have virtual IP in DHT
		return virtualIP, nil
	}

	virtualIP = ""
	for virtualIP == "" {
		virtualIP = s.findNextAvailableIP(peerID)
	}

	log.Printf("Registering peerID %s with virtual IP %s", peerID, virtualIP)

	// Store mappings
	return virtualIP, s.StoreMappingInDHT(peerID, virtualIP)
}

func (s *Service) findNextAvailableIP(peerID string) string {
	for i := 1; i < 255; i++ { // Iterate over last octet 10.x.x.1 to 10.x.x.254
		ip := fmt.Sprintf("%d.%d.%d.%d", s.firstOctet, rand.Intn(256), rand.Intn(256), i)
		if exists, _ := s.IsMappingExistedInDHT(peerID, ip); !exists { // Ensure uniqueness
			return ip
		}
	}
	log.Println("No available IPs found!")
	return "" // Handle edge case where all IPs are taken
}

func GetAppIDFromVirtualIP(virtualIP string) (int64, error) {
	parts := strings.Split(virtualIP, ".")
	if len(parts) != 4 {
		return 0, fmt.Errorf("invalid IP format: %s", virtualIP)
	}

	octet2, err1 := strconv.Atoi(parts[1])
	octet3, err2 := strconv.Atoi(parts[2])
	octet4, err3 := strconv.Atoi(parts[3])

	if err1 != nil || err2 != nil || err3 != nil {
		return 0, fmt.Errorf("invalid octet in IP: %s", virtualIP)
	}

	// Ensure we don't accidentally map special addresses
	if octet4 == 0 || octet4 == 255 {
		return 0, fmt.Errorf("invalid reserved IP: 10.%d.%d.%d", octet2, octet3, octet4)
	}

	appID := (int64(octet2) * 254 * 254) + (int64(octet3) * 254) + int64(octet4)
	return int64(appID), nil
}

func GetVirtualIPFromAppID(appID int64) (string, error) {
	if appID < 1 || appID > (255*255*254) {
		return "", fmt.Errorf("invalid AppID: %d", appID)
	}

	appID-- // Adjust since IDs start from 1
	octet2 := int(appID / (254 * 254))
	octet3 := int((appID % (254 * 254)) / 254)
	octet4 := int((appID % 254)) + 1 // Ensure we never get .0

	// Ensure no special-case IPs are generated
	if octet2 > 255 || octet3 > 255 || octet4 > 255 {
		return "", fmt.Errorf("generated a reserved IP: 10.%d.%d.%d", octet2, octet3, octet4)
	}

	return fmt.Sprintf("10.%d.%d.%d", octet2, octet3, octet4), nil
}
