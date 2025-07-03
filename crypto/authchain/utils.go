package authchain

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// SerializeAuthChain converts an AuthChain to JSON string
func SerializeAuthChain(authChain AuthChain) (string, error) {
	data, err := json.Marshal(authChain)
	if err != nil {
		return "", fmt.Errorf("failed to serialize auth chain: %w", err)
	}
	return string(data), nil
}

// DeserializeAuthChain converts a JSON string to AuthChain
func DeserializeAuthChain(data string) (AuthChain, error) {
	var authChain AuthChain
	err := json.Unmarshal([]byte(data), &authChain)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize auth chain: %w", err)
	}
	return authChain, nil
}

// ValidateAuthChainFormat validates the basic format of an auth chain
func ValidateAuthChainFormat(authChain AuthChain) error {
	if len(authChain) == 0 {
		return fmt.Errorf("auth chain cannot be empty")
	}

	// Check if first link is SIGNER
	if authChain[0].Type != AuthLinkTypeSIGNER {
		return fmt.Errorf("first link must be SIGNER type")
	}

	// Check for multiple SIGNER links
	signerCount := 0
	for _, link := range authChain {
		if link.Type == AuthLinkTypeSIGNER {
			signerCount++
		}
	}
	if signerCount > 1 {
		return fmt.Errorf("auth chain cannot have multiple SIGNER links")
	}

	// Validate owner address format
	ownerAddr := common.HexToAddress(authChain[0].Payload)
	if ownerAddr == (common.Address{}) {
		return fmt.Errorf("invalid owner address format")
	}

	return nil
}

// GetOwnerAddress extracts the owner address from an auth chain
func GetOwnerAddress(authChain AuthChain) (common.Address, error) {
	if len(authChain) == 0 {
		return common.Address{}, fmt.Errorf("empty auth chain")
	}

	if authChain[0].Type != AuthLinkTypeSIGNER {
		return common.Address{}, fmt.Errorf("first link is not SIGNER type")
	}

	ownerAddr := common.HexToAddress(authChain[0].Payload)
	if ownerAddr == (common.Address{}) {
		return common.Address{}, fmt.Errorf("invalid owner address")
	}

	return ownerAddr, nil
}

// GetEphemeralAddress extracts the ephemeral address from an auth chain
func GetEphemeralAddress(authChain AuthChain) (common.Address, error) {
	if len(authChain) < 2 {
		return common.Address{}, fmt.Errorf("auth chain too short")
	}

	if authChain[1].Type != AuthLinkTypeECDSA_PERSONAL_EPHEMERAL {
		return common.Address{}, fmt.Errorf("second link is not ECDSA_PERSONAL_EPHEMERAL type")
	}

	// Support both legacy and new format
	payload := authChain[1].Payload
	if strings.HasPrefix(payload, "delegate:") {
		// Legacy format: delegate:<pubkey>:<timestamp>
		parts := strings.Split(payload, ":")
		if len(parts) != 3 {
			return common.Address{}, fmt.Errorf("invalid ephemeral message format")
		}
		ephemeralPubKeyHex := parts[1]
		if !strings.HasPrefix(ephemeralPubKeyHex, "0x") {
			ephemeralPubKeyHex = "0x" + ephemeralPubKeyHex
		}
		ephemeralPubKeyBytes, err := hex.DecodeString(ephemeralPubKeyHex[2:])
		if err != nil {
			return common.Address{}, fmt.Errorf("invalid ephemeral public key: %w", err)
		}
		if len(ephemeralPubKeyBytes) != 64 {
			return common.Address{}, fmt.Errorf("invalid ephemeral public key length")
		}
		pubKey := &ecdsa.PublicKey{
			Curve: crypto.S256(),
			X:     new(big.Int).SetBytes(ephemeralPubKeyBytes[:32]),
			Y:     new(big.Int).SetBytes(ephemeralPubKeyBytes[32:]),
		}
		return crypto.PubkeyToAddress(*pubKey), nil
	} else {
		// New format: "Subnet Login\nEphemeral address: ...\nExpiration: ..."
		parsed, err := ParseEphemeralPayload(payload)
		if err != nil {
			return common.Address{}, err
		}
		return common.HexToAddress(parsed.EphemeralAddress), nil
	}
}

// GetEntityID extracts the entity ID from an auth chain
func GetEntityID(authChain AuthChain) (string, error) {
	if len(authChain) < 3 {
		return "", fmt.Errorf("auth chain too short")
	}

	if authChain[2].Type != AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY {
		return "", fmt.Errorf("third link is not ECDSA_PERSONAL_SIGNED_ENTITY type")
	}

	parts := strings.Split(authChain[2].Payload, ":")
	if len(parts) != 3 || parts[0] != "request" {
		return "", fmt.Errorf("invalid entity payload format")
	}

	return parts[1], nil
}

// GetExpiryTime extracts the expiry time from an auth chain
func GetExpiryTime(authChain AuthChain) (time.Time, error) {
	if len(authChain) < 3 {
		return time.Time{}, fmt.Errorf("auth chain too short")
	}

	if authChain[2].Type != AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY {
		return time.Time{}, fmt.Errorf("third link is not ECDSA_PERSONAL_SIGNED_ENTITY type")
	}

	parts := strings.Split(authChain[2].Payload, ":")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid entity payload format")
	}

	timestamp, ok := new(big.Int).SetString(parts[2], 10)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid timestamp in entity payload")
	}

	return time.Unix(timestamp.Int64(), 0), nil
}

// IsExpired checks if an auth chain has expired
func IsExpired(authChain AuthChain) (bool, error) {
	expiry, err := GetExpiryTime(authChain)
	if err != nil {
		return false, err
	}

	return time.Now().After(expiry), nil
}

// GetTimeUntilExpiry returns the time remaining until expiry
func GetTimeUntilExpiry(authChain AuthChain) (time.Duration, error) {
	expiry, err := GetExpiryTime(authChain)
	if err != nil {
		return 0, err
	}

	return expiry.Sub(time.Now()), nil
}

// CreateAuthChainFromPrivateKey creates an auth chain using a private key string
func CreateAuthChainFromPrivateKey(privateKeyHex, entityID string, expiry time.Duration) (AuthChain, error) {
	// Parse private key
	if !strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = "0x" + privateKeyHex
	}

	privateKeyBytes, err := hex.DecodeString(privateKeyHex[2:])
	if err != nil {
		return nil, fmt.Errorf("invalid private key format: %w", err)
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	// Create ephemeral key
	ephemeralPrivateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	ownerAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	ephemeralAddress := crypto.PubkeyToAddress(ephemeralPrivateKey.PublicKey)

	ownerIdentity := IdentityType{
		Address:    ownerAddress.Hex(),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(privateKey)),
	}

	ephemeralIdentity := IdentityType{
		Address:    ephemeralAddress.Hex(),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(ephemeralPrivateKey)),
	}

	ephemeralMinutesDuration := int(expiry.Minutes())
	return CreateAuthChain(ownerIdentity, ephemeralIdentity, ephemeralMinutesDuration, entityID), nil
}

// ValidateAuthChain validates an auth chain using the service
func ValidateAuthChain(authChain AuthChain, entityID string) (*ValidationResult, error) {
	service := NewService()

	return service.ValidateAuthChain(entityID, authChain, nil, time.Now().UnixMilli())
}

// ValidateAuthChainWithDefaultExpiry validates an auth chain with default 60s expiry tolerance
func ValidateAuthChainWithDefaultExpiry(authChain AuthChain, entityID string) (*ValidationResult, error) {
	service := NewService()

	// Default 60s expiry tolerance
	expiryTolerance := time.Now().Add(60 * time.Second).UnixMilli()

	return service.ValidateAuthChain(entityID, authChain, nil, expiryTolerance)
}

// AuthChainInfo contains information about an auth chain
type AuthChainInfo struct {
	OwnerAddress     common.Address
	EphemeralAddress common.Address
	EntityID         string
	ExpiryTime       time.Time
	IsExpired        bool
	TimeUntilExpiry  time.Duration
	IsValid          bool
	ValidationError  string
}

// GetAuthChainInfo extracts comprehensive information from an auth chain
func GetAuthChainInfo(authChain AuthChain) *AuthChainInfo {
	info := &AuthChainInfo{}

	// Try to extract owner address
	if ownerAddr, err := GetOwnerAddress(authChain); err == nil {
		info.OwnerAddress = ownerAddr
	} else {
		info.ValidationError = err.Error()
	}

	// Try to extract ephemeral address
	if ephemeralAddr, err := GetEphemeralAddress(authChain); err == nil {
		info.EphemeralAddress = ephemeralAddr
	} else if info.ValidationError == "" {
		info.ValidationError = err.Error()
	}

	// Try to extract entity ID
	if entityID, err := GetEntityID(authChain); err == nil {
		info.EntityID = entityID
	} else if info.ValidationError == "" {
		info.ValidationError = err.Error()
	}

	// Try to extract expiry time
	if expiry, err := GetExpiryTime(authChain); err == nil {
		info.ExpiryTime = expiry
		info.IsExpired = time.Now().After(expiry)
		info.TimeUntilExpiry = expiry.Sub(time.Now())
	} else if info.ValidationError == "" {
		info.ValidationError = err.Error()
	}

	// Check if valid
	info.IsValid = info.ValidationError == "" && !info.IsExpired

	return info
}
