package apps

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// Function to load an *ecdsa.PrivateKey from a hex string
func PrivateKeyFromHex(hexKey string) (*ecdsa.PrivateKey, error) {
	// Decode the hex string to bytes
	privateKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %w", err)
	}

	// Use the crypto package to convert the bytes to a private key
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bytes to ECDSA private key: %w", err)
	}

	return privateKey, nil
}
