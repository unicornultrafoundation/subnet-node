package vpn

import (
	"errors"
	"fmt"
	"strings"

	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/peer"
	subnet_vpn "github.com/unicornultrafoundation/subnet-node/proto/subnet/vpn"
)

type VPNValidator struct{}

// Validate checks if the key and value are valid for a resource record
func (vpn VPNValidator) Validate(key string, value []byte) error {
	// Ensure the key starts with the correct prefix
	if !strings.HasPrefix(key, "/vpn/mapping/") {
		return fmt.Errorf("invalid key: must start with '/vpn/mapping/'")
	}

	// Parse the value as VirtualIPRecord
	var ipRecord subnet_vpn.VirtualIPRecord

	if err := ipRecord.Unmarshal(value); err != nil {
		return fmt.Errorf("invalid value: failed to unmarshal JSON: %w", err)
	}

	// Additional validation for VirtualIPRecord fields
	if ipRecord.Info == nil || ipRecord.Info.Ip == "" {
		return errors.New("invalid virtual IP: IP must be provided")
	}

	if ipRecord.Info.Timestamp <= 0 {
		return errors.New("invalid virtual IP: timestamp must be greater than zero")
	}

	if len(ipRecord.Signature) == 0 {
		return errors.New("invalid virtual IP: signature must be provided")
	}

	// Validate the signature
	peerID := strings.TrimPrefix(key, "/vpn/mapping/")
	valid, err := verifyVirtualIPSignature(&ipRecord, peerID)

	if err != nil {
		return fmt.Errorf("signature verification error: %v", err)
	}

	if !valid {
		return errors.New("invalid virtual IP: signature verification failed")
	}

	return nil
}

// Select resolves conflicts by selecting the first valid value
func (vpn VPNValidator) Select(k string, vals [][]byte) (int, error) {
	for i, v := range vals {
		// Attempt to validate each value
		if err := vpn.Validate(k, v); err == nil {
			return i, nil // Return the first valid value
		}
	}
	return -1, errors.New("no valid values found")
}

// Ensure ResourceValidator implements record.Validator
var _ record.Validator = VPNValidator{}

// verifyVirtualIPSignature verifies the signature of a VirtualIPRecord
func verifyVirtualIPSignature(record *subnet_vpn.VirtualIPRecord, peerID string) (bool, error) {
	// If there's no signature, we can't verify
	if len(record.Signature) == 0 {
		return false, nil
	}

	// If there's no info, we can't verify
	if record.Info == nil {
		return false, nil
	}

	// Parse the peer ID
	parsedPeerID, err := peer.Decode(peerID)
	if err != nil {
		return false, fmt.Errorf("invalid peer ID format: %v", err)
	}

	// Get the public key from the peer ID
	pubKey, err := parsedPeerID.ExtractPublicKey()
	if err != nil {
		return false, fmt.Errorf("failed to extract public key from peer ID: %v", err)
	}

	// Create the hash that was signed
	hash := createVirtualIPHash(record.Info)
	if hash == nil {
		return false, fmt.Errorf("failed to create hash for verification")
	}

	// Verify the signature
	valid, err := pubKey.Verify(hash, record.Signature)
	if err != nil {
		return false, fmt.Errorf("signature verification error: %v", err)
	}

	return valid, nil
}
