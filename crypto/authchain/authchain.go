package authchain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// AuthLinkType represents the type of authentication link
type AuthLinkType string

const (
	// AuthLinkTypeSIGNER represents the owner identity link
	AuthLinkTypeSIGNER AuthLinkType = "SIGNER"
	// AuthLinkTypeECDSA_PERSONAL_EPHEMERAL represents ephemeral key delegation
	AuthLinkTypeECDSA_PERSONAL_EPHEMERAL AuthLinkType = "ECDSA_PERSONAL_EPHEMERAL"
	// AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY represents entity signature
	AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY AuthLinkType = "ECDSA_PERSONAL_SIGNED_ENTITY"
	// AuthLinkTypeECDSA_EIP_1654_EPHEMERAL represents EIP-1654 ephemeral key delegation
	AuthLinkTypeECDSA_EIP_1654_EPHEMERAL AuthLinkType = "ECDSA_EIP_1654_EPHEMERAL"
	// AuthLinkTypeECDSA_EIP_1654_SIGNED_ENTITY represents EIP-1654 entity signature
	AuthLinkTypeECDSA_EIP_1654_SIGNED_ENTITY AuthLinkType = "ECDSA_EIP_1654_SIGNED_ENTITY"
)

// AuthLink represents a single link in the authentication chain
type AuthLink struct {
	Type      AuthLinkType `json:"type"`
	Payload   string       `json:"payload"`
	Signature string       `json:"signature"`
}

// AuthChain represents a complete authentication chain
type AuthChain []AuthLink

// Context contains authentication and authorization information
type Context struct {
	UserAddress    common.Address
	EphemeralAddr  common.Address
	EntityID       string
	ExpiresAt      time.Time
	SessionID      string
	CreatedAt      time.Time
	ChainValidated bool
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

// ValidationOptions contains options for validation
type ValidationOptions struct {
	DateToValidateExpirationInMillis int64
	Provider                         interface{}
}

// ValidatorType represents a validator function
type ValidatorType func(authority string, authLink AuthLink, options *ValidationOptions) (*ValidationStepResult, error)

// ValidationStepResult represents the result of a validation step
type ValidationStepResult struct {
	Error         string `json:"error,omitempty"`
	NextAuthority string `json:"nextAuthority,omitempty"`
}

// EphemeralPayload represents parsed ephemeral payload
type EphemeralPayload struct {
	Message          string
	EphemeralAddress string
	Expiration       int64
}

// IdentityType represents an identity with address and private key
type IdentityType struct {
	Address    string
	PrivateKey string
}

// AuthIdentity represents an authentication identity
type AuthIdentity struct {
	EphemeralIdentity IdentityType
	Expiration        time.Time
	AuthChain         AuthChain
}

// Service provides authchain functionality
type Service struct {
	validators map[AuthLinkType]ValidatorType
}

// NewService creates a new authchain service
func NewService() *Service {
	service := &Service{
		validators: make(map[AuthLinkType]ValidatorType),
	}

	// Register default validators
	service.RegisterValidator(AuthLinkTypeSIGNER, SignerValidator)
	service.RegisterValidator(AuthLinkTypeECDSA_PERSONAL_EPHEMERAL, ECDSAPersonalEphemeralValidator)
	service.RegisterValidator(AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY, ECDSASignedEntityValidator)
	service.RegisterValidator(AuthLinkTypeECDSA_EIP_1654_EPHEMERAL, ECDSAEIP1654EphemeralValidator)
	service.RegisterValidator(AuthLinkTypeECDSA_EIP_1654_SIGNED_ENTITY, EIP1654SignedEntityValidator)

	return service
}

// RegisterValidator registers a validator for a specific auth link type
func (s *Service) RegisterValidator(authLinkType AuthLinkType, validator ValidatorType) {
	s.validators[authLinkType] = validator
}

// ValidateAuthChain validates the complete authentication chain
func (s *Service) ValidateAuthChain(
	expectedFinalAuthority string,
	authChain AuthChain,
	provider interface{},
	dateToValidateExpirationInMillis int64,
) (*ValidationResult, error) {
	if !s.IsValidAuthChain(authChain) {
		return &ValidationResult{
			OK:      false,
			Message: "ERROR: Malformed authChain",
		}, nil
	}

	currentAuthority := ""

	for _, authLink := range authChain {
		validator, exists := s.validators[authLink.Type]
		if !exists {
			return &ValidationResult{
				OK:      false,
				Message: fmt.Sprintf("ERROR: Unknown validator type %s", authLink.Type),
			}, nil
		}

		options := &ValidationOptions{
			DateToValidateExpirationInMillis: dateToValidateExpirationInMillis,
			Provider:                         provider,
		}

		result, err := validator(currentAuthority, authLink, options)
		if err != nil {
			return &ValidationResult{
				OK:      false,
				Message: fmt.Sprintf("ERROR. Link type: %s. %s.", authLink.Type, err.Error()),
			}, nil
		}

		if result.Error != "" {
			return &ValidationResult{
				OK:      false,
				Message: result.Error,
			}, nil
		}

		if result.NextAuthority != "" {
			currentAuthority = result.NextAuthority
		}
	}

	ok := currentAuthority == expectedFinalAuthority
	message := ""
	if !ok {
		message = fmt.Sprintf("ERROR: Invalid final authority. Expected: %s. Current: %s.", expectedFinalAuthority, currentAuthority)
	}

	return &ValidationResult{
		OK:      ok,
		Message: message,
	}, nil
}

// IsValidAuthChain checks if the auth chain is valid
func (s *Service) IsValidAuthChain(authChain AuthChain) bool {
	for index, authLink := range authChain {
		// SIGNER should be the first one
		if index == 0 && authLink.Type != AuthLinkTypeSIGNER {
			return false
		}

		// SIGNER should be unique
		if authLink.Type == AuthLinkTypeSIGNER && index != 0 {
			return false
		}
	}

	return true
}

// CreateSimpleAuthChain creates a simple auth chain
func CreateSimpleAuthChain(finalPayload string, ownerAddress string, signature string) AuthChain {
	return AuthChain{
		{
			Type:      AuthLinkTypeSIGNER,
			Payload:   ownerAddress,
			Signature: "",
		},
		{
			Type:      GetSignedIdentitySignatureType(signature),
			Payload:   finalPayload,
			Signature: signature,
		},
	}
}

// CreateAuthChain creates a complete auth chain
func CreateAuthChain(
	ownerIdentity IdentityType,
	ephemeralIdentity IdentityType,
	ephemeralMinutesDuration int,
	entityID string,
) AuthChain {
	expiration := time.Now().Add(time.Duration(ephemeralMinutesDuration) * time.Minute)

	ephemeralMessage := GetEphemeralMessage(ephemeralIdentity.Address, expiration)
	firstSignature := CreateSignature(ownerIdentity, ephemeralMessage)
	secondSignature := CreateSignature(ephemeralIdentity, ephemeralIdentity.Address)

	return AuthChain{
		{
			Type:      AuthLinkTypeSIGNER,
			Payload:   ownerIdentity.Address,
			Signature: "",
		},
		{
			Type:      AuthLinkTypeECDSA_PERSONAL_EPHEMERAL,
			Payload:   ephemeralMessage,
			Signature: firstSignature,
		},
		{
			Type:      AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY,
			Payload:   ephemeralIdentity.Address,
			Signature: secondSignature,
		},
	}
}

// InitializeAuthChain initializes an auth chain
func InitializeAuthChain(
	ethAddress string,
	ephemeralIdentity IdentityType,
	ephemeralMinutesDuration int,
	signer func(message string) (string, error),
) (*AuthIdentity, error) {
	expiration := time.Now().Add(time.Duration(ephemeralMinutesDuration) * time.Minute)

	ephemeralMessage := GetEphemeralMessage(ephemeralIdentity.Address, expiration)
	firstSignature, err := signer(ephemeralMessage)
	if err != nil {
		return nil, err
	}

	authChain := AuthChain{
		{
			Type:      AuthLinkTypeSIGNER,
			Payload:   ethAddress,
			Signature: "",
		},
		{
			Type:      GetEphemeralSignatureType(firstSignature),
			Payload:   ephemeralMessage,
			Signature: firstSignature,
		},
	}

	return &AuthIdentity{
		EphemeralIdentity: ephemeralIdentity,
		Expiration:        expiration,
		AuthChain:         authChain,
	}, nil
}

// SignPayload signs a payload with an auth identity
func SignPayload(authIdentity *AuthIdentity, entityID string) AuthChain {
	secondSignature := CreateSignature(authIdentity.EphemeralIdentity, entityID)

	result := make(AuthChain, len(authIdentity.AuthChain)+1)
	copy(result, authIdentity.AuthChain)

	result[len(result)-1] = AuthLink{
		Type:      AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY,
		Payload:   entityID,
		Signature: secondSignature,
	}

	return result
}

// CreateSignature creates a signature for a message
func CreateSignature(identity IdentityType, message interface{}) string {
	var messageBytes []byte

	switch msg := message.(type) {
	case string:
		messageBytes = []byte(msg)
	case []byte:
		messageBytes = msg
	default:
		return ""
	}

	// Hash the message
	hash := sha256.Sum256(messageBytes)

	// Convert private key from hex
	privateKeyBytes, err := hex.DecodeString(strings.TrimPrefix(identity.PrivateKey, "0x"))
	if err != nil {
		return ""
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return ""
	}

	// Sign the hash
	signature, err := crypto.Sign(hash[:], privateKey)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(signature)
}

// OwnerAddress extracts the owner address from an auth chain
func OwnerAddress(authChain AuthChain) string {
	if len(authChain) > 0 {
		if authChain[0].Type == AuthLinkTypeSIGNER {
			return authChain[0].Payload
		}
	}
	return "Invalid-Owner-Address"
}

// GetEphemeralMessage creates an ephemeral message
func GetEphemeralMessage(ephemeralAddress string, expiration time.Time) string {
	return fmt.Sprintf("Subnet Node Login\nEphemeral address: %s\nExpiration: %s",
		ephemeralAddress, expiration.Format(time.RFC3339))
}

// GetEphemeralSignatureType determines the signature type based on signature length
func GetEphemeralSignatureType(signature string) AuthLinkType {
	if len(signature) == 132 { // PERSONAL_SIGNATURE_LENGTH
		return AuthLinkTypeECDSA_PERSONAL_EPHEMERAL
	}
	return AuthLinkTypeECDSA_EIP_1654_EPHEMERAL
}

// GetSignedIdentitySignatureType determines the signature type based on signature length
func GetSignedIdentitySignatureType(signature string) AuthLinkType {
	if len(signature) == 132 { // PERSONAL_SIGNATURE_LENGTH
		return AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY
	}
	return AuthLinkTypeECDSA_EIP_1654_SIGNED_ENTITY
}

// ParseEphemeralPayload parses an ephemeral payload
func ParseEphemeralPayload(payload string) (*EphemeralPayload, error) {
	message := strings.ReplaceAll(payload, "\r", "")
	payloadParts := strings.Split(message, "\n")

	if len(payloadParts) < 3 {
		return nil, fmt.Errorf("invalid ephemeral payload format")
	}

	ephemeralAddress := strings.TrimPrefix(payloadParts[1], "Ephemeral address: ")
	expirationString := strings.TrimPrefix(payloadParts[2], "Expiration: ")

	expiration, err := time.Parse(time.RFC3339, expirationString)
	if err != nil {
		return nil, fmt.Errorf("invalid expiration format: %w", err)
	}

	return &EphemeralPayload{
		Message:          message,
		EphemeralAddress: ephemeralAddress,
		Expiration:       expiration.UnixMilli(),
	}, nil
}

// Validator implementations

// SignerValidator validates SIGNER type links
func SignerValidator(_ string, authLink AuthLink, _ *ValidationOptions) (*ValidationStepResult, error) {
	return &ValidationStepResult{
		NextAuthority: authLink.Payload,
	}, nil
}

// ECDSASignedEntityValidator validates ECDSA_PERSONAL_SIGNED_ENTITY type links
func ECDSASignedEntityValidator(authority string, authLink AuthLink, _ *ValidationOptions) (*ValidationStepResult, error) {
	if authLink.Signature == "" {
		return &ValidationStepResult{
			Error: "Invalid AuthLink. 'signature' must be present for type 'ECDSA_PERSONAL_SIGNED_ENTITY'",
		}, nil
	}

	signerAddress, err := RecoverAddressFromEthSignature(authLink.Signature, authLink.Payload)
	if err != nil {
		return &ValidationStepResult{
			Error: fmt.Sprintf("Failed to recover address: %s", err.Error()),
		}, nil
	}

	expectedSignedAddress := strings.ToLower(authority)
	actualSignedAddress := strings.ToLower(signerAddress)

	if expectedSignedAddress == actualSignedAddress {
		return &ValidationStepResult{
			NextAuthority: authLink.Payload,
		}, nil
	}

	return &ValidationStepResult{
		Error: fmt.Sprintf("Invalid signer address. Expected: %s. Actual: %s", expectedSignedAddress, actualSignedAddress),
	}, nil
}

// ECDSAPersonalEphemeralValidator validates ECDSA_PERSONAL_EPHEMERAL type links
func ECDSAPersonalEphemeralValidator(authority string, authLink AuthLink, options *ValidationOptions) (*ValidationStepResult, error) {
	if authLink.Signature == "" {
		return &ValidationStepResult{
			Error: "Invalid AuthLink. 'signature' must be present for type 'ECDSA_PERSONAL_EPHEMERAL'",
		}, nil
	}

	payload, err := ParseEphemeralPayload(authLink.Payload)
	if err != nil {
		return &ValidationStepResult{
			Error: fmt.Sprintf("Failed to parse ephemeral payload: %s", err.Error()),
		}, nil
	}

	dateToValidateExpirationInMillis := options.DateToValidateExpirationInMillis
	if dateToValidateExpirationInMillis == 0 {
		dateToValidateExpirationInMillis = time.Now().UnixMilli()
	}

	if payload.Expiration > dateToValidateExpirationInMillis {
		signerAddress, err := RecoverAddressFromEthSignature(authLink.Signature, payload.Message)
		if err != nil {
			return &ValidationStepResult{
				Error: fmt.Sprintf("Failed to recover address: %s", err.Error()),
			}, nil
		}

		expectedSignedAddress := strings.ToLower(authority)
		actualSignedAddress := strings.ToLower(signerAddress)

		if expectedSignedAddress == actualSignedAddress {
			return &ValidationStepResult{
				NextAuthority: payload.EphemeralAddress,
			}, nil
		}

		return &ValidationStepResult{
			Error: fmt.Sprintf("Invalid signer address. Expected: %s. Actual: %s", expectedSignedAddress, actualSignedAddress),
		}, nil
	}

	return &ValidationStepResult{
		Error: fmt.Sprintf("Ephemeral key expired. Expiration: %d. Test: %d", payload.Expiration, dateToValidateExpirationInMillis),
	}, nil
}

// ECDSAEIP1654EphemeralValidator validates ECDSA_EIP_1654_EPHEMERAL type links
func ECDSAEIP1654EphemeralValidator(authority string, authLink AuthLink, options *ValidationOptions) (*ValidationStepResult, error) {
	if authLink.Signature == "" {
		return &ValidationStepResult{
			Error: "Invalid AuthLink. 'signature' must be present for type 'ECDSA_EIP_1654_EPHEMERAL'",
		}, nil
	}

	payload, err := ParseEphemeralPayload(authLink.Payload)
	if err != nil {
		return &ValidationStepResult{
			Error: fmt.Sprintf("Failed to parse ephemeral payload: %s", err.Error()),
		}, nil
	}

	dateToValidateExpirationInMillis := options.DateToValidateExpirationInMillis
	if dateToValidateExpirationInMillis == 0 {
		dateToValidateExpirationInMillis = time.Now().UnixMilli()
	}

	if payload.Expiration > dateToValidateExpirationInMillis {
		// For now, we'll use the same validation as personal signature
		// In a real implementation, you would validate against EIP-1654 contract
		signerAddress, err := RecoverAddressFromEthSignature(authLink.Signature, payload.Message)
		if err != nil {
			return &ValidationStepResult{
				Error: fmt.Sprintf("Failed to recover address: %s", err.Error()),
			}, nil
		}

		expectedSignedAddress := strings.ToLower(authority)
		actualSignedAddress := strings.ToLower(signerAddress)

		if expectedSignedAddress == actualSignedAddress {
			return &ValidationStepResult{
				NextAuthority: payload.EphemeralAddress,
			}, nil
		}

		return &ValidationStepResult{
			Error: fmt.Sprintf("Invalid signer address. Expected: %s. Actual: %s", expectedSignedAddress, actualSignedAddress),
		}, nil
	}

	return &ValidationStepResult{
		Error: fmt.Sprintf("Ephemeral key expired. Expiration: %d. Test: %d", payload.Expiration, dateToValidateExpirationInMillis),
	}, nil
}

// EIP1654SignedEntityValidator validates ECDSA_EIP_1654_SIGNED_ENTITY type links
func EIP1654SignedEntityValidator(authority string, authLink AuthLink, options *ValidationOptions) (*ValidationStepResult, error) {
	if authLink.Signature == "" {
		return &ValidationStepResult{
			Error: "Invalid AuthLink. 'signature' must be present for type 'ECDSA_EIP_1654_SIGNED_ENTITY'",
		}, nil
	}

	// For now, we'll use the same validation as personal signature
	// In a real implementation, you would validate against EIP-1654 contract
	signerAddress, err := RecoverAddressFromEthSignature(authLink.Signature, authLink.Payload)
	if err != nil {
		return &ValidationStepResult{
			Error: fmt.Sprintf("Failed to recover address: %s", err.Error()),
		}, nil
	}

	expectedSignedAddress := strings.ToLower(authority)
	actualSignedAddress := strings.ToLower(signerAddress)

	if expectedSignedAddress == actualSignedAddress {
		return &ValidationStepResult{
			NextAuthority: authLink.Payload,
		}, nil
	}

	return &ValidationStepResult{
		Error: fmt.Sprintf("Invalid signer address. Expected: %s. Actual: %s", expectedSignedAddress, actualSignedAddress),
	}, nil
}

// RecoverAddressFromEthSignature recovers the address from an Ethereum signature
func RecoverAddressFromEthSignature(signature string, message string) (string, error) {
	// Hash the message
	hash := sha256.Sum256([]byte(message))

	// Decode the signature from hex
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return "", fmt.Errorf("invalid signature hex format: %w", err)
	}

	// Verify signature length
	if len(sigBytes) != 65 {
		return "", fmt.Errorf("invalid signature length, expected 65 bytes")
	}

	// Recover the public key from signature
	pubKey, err := crypto.SigToPub(hash[:], sigBytes)
	if err != nil {
		return "", fmt.Errorf("failed to recover public key: %w", err)
	}

	// Extract address from public key
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	return recoveredAddr.Hex(), nil
}

// generateSessionID generates a unique session identifier
func generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
