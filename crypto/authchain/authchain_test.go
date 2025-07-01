package authchain

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	service := NewService()
	assert.NotNil(t, service)
	assert.NotEmpty(t, service.validators)
}

func TestIsValidAuthChain(t *testing.T) {
	service := NewService()

	// Valid auth chain
	validChain := AuthChain{
		{Type: AuthLinkTypeSIGNER, Payload: "0x1234567890123456789012345678901234567890", Signature: ""},
		{Type: AuthLinkTypeECDSA_PERSONAL_EPHEMERAL, Payload: "test", Signature: "test"},
		{Type: AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY, Payload: "test", Signature: "test"},
	}
	assert.True(t, service.IsValidAuthChain(validChain))

	// Invalid: SIGNER not first
	invalidChain1 := AuthChain{
		{Type: AuthLinkTypeECDSA_PERSONAL_EPHEMERAL, Payload: "test", Signature: "test"},
		{Type: AuthLinkTypeSIGNER, Payload: "0x1234567890123456789012345678901234567890", Signature: ""},
	}
	assert.False(t, service.IsValidAuthChain(invalidChain1))

	// Invalid: Multiple SIGNER
	invalidChain2 := AuthChain{
		{Type: AuthLinkTypeSIGNER, Payload: "0x1234567890123456789012345678901234567890", Signature: ""},
		{Type: AuthLinkTypeSIGNER, Payload: "0x0987654321098765432109876543210987654321", Signature: ""},
	}
	assert.False(t, service.IsValidAuthChain(invalidChain2))
}

func TestCreateSimpleAuthChain(t *testing.T) {
	finalPayload := "test-payload"
	ownerAddress := "0x1234567890123456789012345678901234567890"
	signature := "0x" + strings.Repeat("a", 130) // 132 chars total

	authChain := CreateSimpleAuthChain(finalPayload, ownerAddress, signature)

	assert.Len(t, authChain, 2)
	assert.Equal(t, AuthLinkTypeSIGNER, authChain[0].Type)
	assert.Equal(t, ownerAddress, authChain[0].Payload)
	assert.Equal(t, "", authChain[0].Signature)

	assert.Equal(t, AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY, authChain[1].Type)
	assert.Equal(t, finalPayload, authChain[1].Payload)
	assert.Equal(t, signature, authChain[1].Signature)
}

func TestCreateAuthChain(t *testing.T) {
	// Generate private keys for testing
	ownerPrivateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	require.NoError(t, err)

	ephemeralPrivateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	require.NoError(t, err)

	ownerAddress := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)
	ephemeralAddress := crypto.PubkeyToAddress(ephemeralPrivateKey.PublicKey)

	ownerIdentity := IdentityType{
		Address:    ownerAddress.Hex(),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(ownerPrivateKey)),
	}

	ephemeralIdentity := IdentityType{
		Address:    ephemeralAddress.Hex(),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(ephemeralPrivateKey)),
	}

	entityID := "test-entity-123"
	ephemeralMinutesDuration := 60

	authChain := CreateAuthChain(ownerIdentity, ephemeralIdentity, ephemeralMinutesDuration, entityID)

	assert.Len(t, authChain, 3)
	assert.Equal(t, AuthLinkTypeSIGNER, authChain[0].Type)
	assert.Equal(t, ownerAddress.Hex(), authChain[0].Payload)
	assert.Equal(t, "", authChain[0].Signature)

	assert.Equal(t, AuthLinkTypeECDSA_PERSONAL_EPHEMERAL, authChain[1].Type)
	assert.NotEmpty(t, authChain[1].Payload)
	assert.NotEmpty(t, authChain[1].Signature)

	assert.Equal(t, AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY, authChain[2].Type)
	assert.Equal(t, entityID, authChain[2].Payload)
	assert.NotEmpty(t, authChain[2].Signature)
}

func TestInitializeAuthChain(t *testing.T) {
	// Generate private key for testing
	ephemeralPrivateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	require.NoError(t, err)

	ethAddress := "0x1234567890123456789012345678901234567890"
	ephemeralAddress := crypto.PubkeyToAddress(ephemeralPrivateKey.PublicKey)

	ephemeralIdentity := IdentityType{
		Address:    ephemeralAddress.Hex(),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(ephemeralPrivateKey)),
	}

	ephemeralMinutesDuration := 30

	// Mock signer function
	signer := func(message string) (string, error) {
		return "0x" + strings.Repeat("a", 130), nil
	}

	authIdentity, err := InitializeAuthChain(ethAddress, ephemeralIdentity, ephemeralMinutesDuration, signer)
	require.NoError(t, err)

	assert.NotNil(t, authIdentity)
	assert.Equal(t, ephemeralIdentity, authIdentity.EphemeralIdentity)
	assert.True(t, time.Now().Before(authIdentity.Expiration))
	assert.Len(t, authIdentity.AuthChain, 2)
}

func TestSignPayload(t *testing.T) {
	// Generate private key for testing
	ephemeralPrivateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	require.NoError(t, err)

	ephemeralAddress := crypto.PubkeyToAddress(ephemeralPrivateKey.PublicKey)

	ephemeralIdentity := IdentityType{
		Address:    ephemeralAddress.Hex(),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(ephemeralPrivateKey)),
	}

	authIdentity := &AuthIdentity{
		EphemeralIdentity: ephemeralIdentity,
		Expiration:        time.Now().Add(time.Hour),
		AuthChain: AuthChain{
			{Type: AuthLinkTypeSIGNER, Payload: "0x1234567890123456789012345678901234567890", Signature: ""},
			{Type: AuthLinkTypeECDSA_PERSONAL_EPHEMERAL, Payload: "test", Signature: "test"},
		},
	}

	entityID := "test-entity-456"
	result := SignPayload(authIdentity, entityID)

	assert.Len(t, result, 3)
	assert.Equal(t, AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY, result[2].Type)
	assert.Equal(t, entityID, result[2].Payload)
	assert.NotEmpty(t, result[2].Signature)
}

func TestCreateSignature(t *testing.T) {
	// Generate private key for testing
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	require.NoError(t, err)

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	identity := IdentityType{
		Address:    address.Hex(),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(privateKey)),
	}

	message := "test message"
	signature := CreateSignature(identity, message)

	assert.NotEmpty(t, signature)
	assert.Len(t, signature, 130) // 65 bytes = 130 hex chars
}

func TestOwnerAddress(t *testing.T) {
	ownerAddr := "0x1234567890123456789012345678901234567890"

	authChain := AuthChain{
		{Type: AuthLinkTypeSIGNER, Payload: ownerAddr, Signature: ""},
		{Type: AuthLinkTypeECDSA_PERSONAL_EPHEMERAL, Payload: "test", Signature: "test"},
	}

	result := OwnerAddress(authChain)
	assert.Equal(t, ownerAddr, result)

	// Test with empty chain
	emptyChain := AuthChain{}
	result = OwnerAddress(emptyChain)
	assert.Equal(t, "Invalid-Owner-Address", result)

	// Test with non-SIGNER first link
	invalidChain := AuthChain{
		{Type: AuthLinkTypeECDSA_PERSONAL_EPHEMERAL, Payload: "test", Signature: "test"},
	}
	result = OwnerAddress(invalidChain)
	assert.Equal(t, "Invalid-Owner-Address", result)
}

func TestGetEphemeralMessage(t *testing.T) {
	ephemeralAddress := "0x1234567890123456789012345678901234567890"
	expiration := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)

	message := GetEphemeralMessage(ephemeralAddress, expiration)

	expected := "Subnet Login\nEphemeral address: 0x1234567890123456789012345678901234567890\nExpiration: 2023-12-25T10:30:00Z"
	assert.Equal(t, expected, message)
}

func TestGetEphemeralSignatureType(t *testing.T) {
	// Personal signature (132 chars)
	personalSig := "0x" + strings.Repeat("a", 130)
	assert.Equal(t, AuthLinkTypeECDSA_PERSONAL_EPHEMERAL, GetEphemeralSignatureType(personalSig))

	// EIP-1654 signature (different length)
	eip1654Sig := "0x" + strings.Repeat("b", 100)
	assert.Equal(t, AuthLinkTypeECDSA_EIP_1654_EPHEMERAL, GetEphemeralSignatureType(eip1654Sig))
}

func TestGetSignedIdentitySignatureType(t *testing.T) {
	// Personal signature (132 chars)
	personalSig := "0x" + strings.Repeat("a", 130)
	assert.Equal(t, AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY, GetSignedIdentitySignatureType(personalSig))

	// EIP-1654 signature (different length)
	eip1654Sig := "0x" + strings.Repeat("b", 100)
	assert.Equal(t, AuthLinkTypeECDSA_EIP_1654_SIGNED_ENTITY, GetSignedIdentitySignatureType(eip1654Sig))
}

func TestParseEphemeralPayload(t *testing.T) {
	payload := "Subnet Login\nEphemeral address: 0x1234567890123456789012345678901234567890\nExpiration: 2023-12-25T10:30:00Z"

	result, err := ParseEphemeralPayload(payload)
	require.NoError(t, err)

	assert.Equal(t, "Subnet Login\nEphemeral address: 0x1234567890123456789012345678901234567890\nExpiration: 2023-12-25T10:30:00Z", result.Message)
	assert.Equal(t, "0x1234567890123456789012345678901234567890", result.EphemeralAddress)
	// Instead of hardcoding, check that expiration parses to the correct time
	exp, _ := time.Parse(time.RFC3339, "2023-12-25T10:30:00Z")
	assert.Equal(t, exp.UnixMilli(), result.Expiration)
}

func TestParseEphemeralPayloadInvalid(t *testing.T) {
	// Invalid format
	invalidPayload := "Invalid format"
	_, err := ParseEphemeralPayload(invalidPayload)
	assert.Error(t, err)

	// Invalid expiration format
	invalidExpirationPayload := "Subnet Login\nEphemeral address: 0x1234567890123456789012345678901234567890\nExpiration: invalid-date"
	_, err = ParseEphemeralPayload(invalidExpirationPayload)
	assert.Error(t, err)
}

func TestRecoverAddressFromEthSignature(t *testing.T) {
	// Generate private key for testing
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	require.NoError(t, err)

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	message := "test message"

	// Create signature using Ethereum personal sign format
	messageHash := createEthereumMessageHash(message)
	signature, err := crypto.Sign(messageHash, privateKey)
	require.NoError(t, err)

	// pubKey, err := crypto.SigToPub(messageHash, signature)
	// require.NoError(t, err)
	// recoveredAddress := crypto.PubkeyToAddress(*pubKey)

	signatureHex := hex.EncodeToString(signature)

	// Recover address
	recoveredAddress, err := RecoverAddressFromEthSignature(signatureHex, message)
	require.NoError(t, err)

	assert.Equal(t, address.Hex(), recoveredAddress)
}

func TestRecoverAddressFromEthSignatureInvalid(t *testing.T) {
	// Invalid hex
	_, err := RecoverAddressFromEthSignature("invalid-hex", "test")
	assert.Error(t, err)

	// Invalid length
	_, err = RecoverAddressFromEthSignature("0x1234567890abcdef", "test")
	assert.Error(t, err)
}

func TestValidatorImplementations(t *testing.T) {
	service := NewService()

	// Test SIGNER validator
	result, err := service.validators[AuthLinkTypeSIGNER]("", AuthLink{
		Type:      AuthLinkTypeSIGNER,
		Payload:   "0x1234567890123456789012345678901234567890",
		Signature: "",
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "0x1234567890123456789012345678901234567890", result.NextAuthority)
	assert.Empty(t, result.Error)

	// Test ECDSA_PERSONAL_SIGNED_ENTITY validator with missing signature
	result, err = service.validators[AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY]("0x1234567890123456789012345678901234567890", AuthLink{
		Type:      AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY,
		Payload:   "test",
		Signature: "",
	}, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Error)
	assert.Empty(t, result.NextAuthority)
}

func TestValidateAuthChain(t *testing.T) {
	service := NewService()

	// Generate private keys for testing
	ownerPrivateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	require.NoError(t, err)

	ephemeralPrivateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	require.NoError(t, err)

	ownerAddress := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)
	ephemeralAddress := crypto.PubkeyToAddress(ephemeralPrivateKey.PublicKey)

	ownerIdentity := IdentityType{
		Address:    ownerAddress.Hex(),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(ownerPrivateKey)),
	}

	ephemeralIdentity := IdentityType{
		Address:    ephemeralAddress.Hex(),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(ephemeralPrivateKey)),
	}

	entityID := "test-entity-123"
	authChain := CreateAuthChain(ownerIdentity, ephemeralIdentity, 60, entityID)

	// Test valid validation
	result, err := service.ValidateAuthChain(entityID, authChain, nil, time.Now().UnixMilli())
	require.NoError(t, err)
	assert.True(t, result.OK)
	assert.Empty(t, result.Message)

	// Test invalid final authority
	result, err = service.ValidateAuthChain("1", authChain, nil, time.Now().UnixMilli())
	require.NoError(t, err)
	assert.False(t, result.OK)
	assert.Contains(t, result.Message, "Invalid final authority")
}

func TestValidateAuthChainMalformed(t *testing.T) {
	service := NewService()

	// Malformed chain
	malformedChain := AuthChain{
		{Type: AuthLinkTypeECDSA_PERSONAL_EPHEMERAL, Payload: "test", Signature: "test"},
	}

	result, err := service.ValidateAuthChain("0x1234567890123456789012345678901234567890", malformedChain, nil, time.Now().UnixMilli())
	require.NoError(t, err)
	assert.False(t, result.OK)
	assert.Contains(t, result.Message, "Malformed authChain")
}

func TestGenerateSessionID(t *testing.T) {
	sessionID1 := generateSessionID()
	sessionID2 := generateSessionID()

	assert.NotEmpty(t, sessionID1)
	assert.NotEmpty(t, sessionID2)
	assert.NotEqual(t, sessionID1, sessionID2)
	assert.Len(t, sessionID1, 64) // 32 bytes = 64 hex chars
}
