package validator

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	subnet_vpn "github.com/unicornultrafoundation/subnet-node/proto/subnet/vpn"
)

var log = logrus.WithField("service", "vpn-validator")

// VPNValidator validates VPN records in the DHT
type VPNValidator struct{}

// Validate checks if the key and value are valid for a VPN record
func (vpn VPNValidator) Validate(key string, value []byte) error {
	// Ensure the key starts with the correct prefix
	const prefix = "/vpn/mapping/"
	if !strings.HasPrefix(key, prefix) {
		return fmt.Errorf("invalid key: must start with '%s'", prefix)
	}

	// Parse the value as VirtualIPRecord
	var ipRecord subnet_vpn.VirtualIPRecord
	if err := proto.Unmarshal(value, &ipRecord); err != nil {
		return fmt.Errorf("invalid value: failed to unmarshal protobuf: %w", err)
	}

	// Validate the record fields
	if err := validateRecordFields(&ipRecord); err != nil {
		return err
	}

	// Validate the signature
	peerID := strings.TrimPrefix(key, prefix)
	valid, err := verifyVirtualIPSignature(&ipRecord, peerID)
	if err != nil {
		return fmt.Errorf("signature verification error: %w", err)
	}

	if !valid {
		return errors.New("invalid virtual IP: signature verification failed")
	}

	return nil
}

// validateRecordFields validates the fields of a VirtualIPRecord
func validateRecordFields(record *subnet_vpn.VirtualIPRecord) error {
	if record.Info == nil || record.Info.Ip == "" {
		return errors.New("invalid virtual IP: IP must be provided")
	}

	if record.Info.Timestamp <= 0 {
		return errors.New("invalid virtual IP: timestamp must be greater than zero")
	}

	if len(record.Signature) == 0 {
		return errors.New("invalid virtual IP: signature must be provided")
	}

	return nil
}

// Select resolves conflicts by selecting the first valid value
func (vpn VPNValidator) Select(k string, vals [][]byte) (int, error) {
	if len(vals) == 0 {
		return -1, errors.New("no values provided")
	}

	for i, v := range vals {
		// Attempt to validate each value
		if err := vpn.Validate(k, v); err == nil {
			return i, nil // Return the first valid value
		}
	}

	return -1, errors.New("no valid values found")
}

// Ensure VPNValidator implements record.Validator
var _ record.Validator = VPNValidator{}

// verifyVirtualIPSignature verifies the signature of a VirtualIPRecord
func verifyVirtualIPSignature(record *subnet_vpn.VirtualIPRecord, peerID string) (bool, error) {
	// Basic validation checks
	if len(record.Signature) == 0 {
		return false, errors.New("missing signature")
	}

	if record.Info == nil {
		return false, errors.New("missing IP info")
	}

	// Parse the peer ID
	parsedPeerID, err := peer.Decode(peerID)
	if err != nil {
		return false, fmt.Errorf("invalid peer ID format: %w", err)
	}

	// Get the public key from the peer ID
	pubKey, err := parsedPeerID.ExtractPublicKey()
	if err != nil {
		return false, fmt.Errorf("failed to extract public key from peer ID: %w", err)
	}

	// Create the hash that was signed
	hash := createVirtualIPHash(record.Info)
	if hash == nil {
		return false, errors.New("failed to create hash for verification")
	}

	// Verify the signature
	valid, err := pubKey.Verify(hash, record.Signature)
	if err != nil {
		return false, fmt.Errorf("signature verification error: %w", err)
	}

	return valid, nil
}

// createVirtualIPHash generates a hash of the timestamp and IP address
func createVirtualIPHash(info *subnet_vpn.VirtualIPInfo) []byte {
	if info == nil {
		log.Error("Cannot create hash for nil IP info")
		return nil
	}

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
