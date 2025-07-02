package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	bidenginetypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	authchain "github.com/unicornultrafoundation/subnet-node/crypto/authchain"
	"k8s.io/client-go/tools/remotecommand"
)

// MockConfigProvider implements ConfigProvider for testing
type MockConfigProvider struct {
	mock.Mock
}

func (m *MockConfigProvider) GetString(key string, defaultValue string) string {
	args := m.Called(key, defaultValue)
	return args.String(0)
}

func (m *MockConfigProvider) GetBool(key string, defaultValue bool) bool {
	args := m.Called(key, defaultValue)
	return args.Bool(0)
}

// MockBidMarketContract implements BidMarketContract for testing
type MockBidMarketContract struct {
	mock.Mock
}

func (m *MockBidMarketContract) GetOrder(ctx context.Context, orderID *big.Int) (*bidenginetypes.Order, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*bidenginetypes.Order), args.Error(1)
}

// MockDeployerService implements DeployerService for testing
type MockDeployerService struct {
	mock.Mock
}

func (m *MockDeployerService) RequestDeployment(ctx context.Context, req *types.DeploymentRequest) (*types.DeploymentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*types.DeploymentResponse), args.Error(1)
}

func (m *MockDeployerService) GetDeployment(ctx context.Context, orderID string) (*types.DeploymentResponse, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*types.DeploymentResponse), args.Error(1)
}

func (m *MockDeployerService) CleanupDeployment(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockDeployerService) GetDeployments(ctx context.Context, requester string) ([]*types.DeploymentResponse, error) {
	args := m.Called(ctx, requester)
	return args.Get(0).([]*types.DeploymentResponse), args.Error(1)
}

func (m *MockDeployerService) GetDeploymentLogs(ctx context.Context, orderID string) ([]*types.ServiceLog, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]*types.ServiceLog), args.Error(1)
}

func (m *MockDeployerService) GetServiceStatus(ctx context.Context, orderID, serviceName string) (*types.ServiceStatus, error) {
	args := m.Called(ctx, orderID, serviceName)
	return args.Get(0).(*types.ServiceStatus), args.Error(1)
}

func (m *MockDeployerService) Exec(ctx context.Context, orderID, podName, serviceName string, cmd []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, tty bool, tsq remotecommand.TerminalSizeQueue) (types.ExecResult, error) {
	args := m.Called(ctx, orderID, podName, serviceName, cmd, stdin, stdout, stderr, tty, tsq)
	return args.Get(0).(types.ExecResult), args.Error(1)
}

// createTestAuthChain creates a deterministic authchain for testing using a seed
func createTestAuthChain(orderID string) string {
	// Use orderID as seed for deterministic key generation
	seed := sha256.Sum256([]byte(orderID))

	// Generate deterministic private keys from seed
	ownerPrivateKey := generateDeterministicPrivateKey(seed[:], "owner")
	ephemeralPrivateKey := generateDeterministicPrivateKey(seed[:], "ephemeral")

	// Get addresses
	ownerAddress := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)
	ephemeralAddress := crypto.PubkeyToAddress(ephemeralPrivateKey.PublicKey)

	// Create ephemeral message with RFC3339 time format
	expiration := time.Now().Add(time.Hour)
	ephemeralMessage := fmt.Sprintf("Subnet Node Login\nEphemeral address: %s\nExpiration: %s",
		ephemeralAddress.Hex(), expiration.Format(time.RFC3339))

	// Sign ephemeral message with owner's private key
	ephemeralSignature := authchain.CreateSignature(authchain.IdentityType{
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(ownerPrivateKey)),
		Address:    ownerAddress.Hex(),
	}, ephemeralMessage)

	// Create entity ID
	entityID := fmt.Sprintf("subnet_deployment:1:1:%s", orderID)

	// Sign entity ID with ephemeral private key
	entitySignature := authchain.CreateSignature(authchain.IdentityType{
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(ephemeralPrivateKey)),
		Address:    ephemeralAddress.Hex(),
	}, entityID)

	// Build auth chain
	chain := authchain.AuthChain{
		{
			Type:    authchain.AuthLinkTypeSIGNER,
			Payload: ownerAddress.Hex(),
		},
		{
			Type:      authchain.AuthLinkTypeECDSA_PERSONAL_EPHEMERAL,
			Payload:   ephemeralMessage,
			Signature: ephemeralSignature,
		},
		{
			Type:      authchain.AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY,
			Payload:   entityID,
			Signature: entitySignature,
		},
	}

	serialized, _ := authchain.SerializeAuthChain(chain)
	return serialized
}

// generateDeterministicPrivateKey creates a deterministic private key from a seed
func generateDeterministicPrivateKey(seed []byte, purpose string) *ecdsa.PrivateKey {
	// Create deterministic data for key generation
	data := append(seed, []byte(purpose)...)
	hash := sha256.Sum256(data)

	// Use hash as private key (ensure it's valid for secp256k1)
	privateKeyBytes := hash[:]

	// Ensure the private key is within the valid range for secp256k1
	curve := crypto.S256()
	curveOrder := curve.Params().N

	privateKeyInt := new(big.Int).SetBytes(privateKeyBytes)
	privateKeyInt.Mod(privateKeyInt, curveOrder)

	// Create private key
	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
		},
		D: privateKeyInt,
	}

	// Compute public key
	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(privateKeyInt.Bytes())

	return privateKey
}

// getOwnerAddressFromAuthChain extracts the owner address from a serialized authchain
func getOwnerAddressFromAuthChain(authChainStr string) string {
	chain, err := authchain.DeserializeAuthChain(authChainStr)
	if err != nil {
		panic(fmt.Sprintf("failed to deserialize authchain: %v", err))
	}

	if len(chain) == 0 {
		panic("empty authchain")
	}

	return chain[0].Payload
}

func TestDeploymentHandler_HealthHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}
	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "deployment-api", response["service"])
}

func TestDeploymentHandler_CreateDeploymentHandler_WithoutAuth(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}
	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	requestBody := types.DeploymentRequest{
		OrderID:   "123",
		Requester: "0x1234567890123456789012345678901234567890",
		TTL:       60,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/deployments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401 because no authchain header
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "X-AuthChain header required", response["error"])
}

func TestDeploymentHandler_GetDeploymentHandler_WithoutAuth(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}
	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	req := httptest.NewRequest("GET", "/api/v1/deployments/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401 because no authchain header
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "X-AuthChain header required", response["error"])
}

func TestDeploymentHandler_CleanupDeploymentHandler_WithoutAuth(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}
	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	req := httptest.NewRequest("DELETE", "/api/v1/deployments/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401 because no authchain header
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "X-AuthChain header required", response["error"])
}

func TestDeploymentHandler_CreateDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}

	// Mock config values
	mockConfig.On("GetString", "provider.id", "").Return("1")
	mockConfig.On("GetString", "provider.machine_id", "").Return("1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	// Create authchain and get owner address
	authChainStr := createTestAuthChain("123")
	ownerAddress := getOwnerAddressFromAuthChain(authChainStr)

	// Mock bid market to return order information
	mockOrder := &bidenginetypes.Order{
		ID:                 big.NewInt(123),
		Owner:              common.HexToAddress(ownerAddress),
		AcceptedProviderId: big.NewInt(1),
		AcceptedMachineId:  big.NewInt(1), // This should match the machine_id from config
		Status:             bidenginetypes.OrderStatusAccepted,
	}
	mockBidMarket.On("GetOrder", mock.Anything, big.NewInt(123)).Return(mockOrder, nil)

	// Mock successful deployment creation
	mockDeployer.On("RequestDeployment", mock.Anything, mock.Anything).Return(&types.DeploymentResponse{
		ID: "123",
		Status: &types.DeploymentStatus{
			State: types.DeploymentStateRunning,
		},
	}, nil)

	requestBody := types.DeploymentRequest{
		OrderID:   "123",
		Requester: ownerAddress,
		TTL:       60,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/deployments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AuthChain", authChainStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Debug: print response body
	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}

	assert.Equal(t, http.StatusCreated, w.Code)

	var response types.DeploymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "123", response.ID)

}

func TestDeploymentHandler_GetDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}

	// Mock config values
	mockConfig.On("GetString", "provider.id", "").Return("1")
	mockConfig.On("GetString", "provider.machine_id", "").Return("1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	// Create authchain and get owner address
	authChainStr := createTestAuthChain("123")
	ownerAddress := getOwnerAddressFromAuthChain(authChainStr)

	// Mock bid market to return order information
	mockOrder := &bidenginetypes.Order{
		ID:                 big.NewInt(123),
		Owner:              common.HexToAddress(ownerAddress),
		AcceptedProviderId: big.NewInt(1),
		AcceptedMachineId:  big.NewInt(1), // This should match the machine_id from config
		Status:             bidenginetypes.OrderStatusAccepted,
	}
	mockBidMarket.On("GetOrder", mock.Anything, big.NewInt(123)).Return(mockOrder, nil)

	// Mock successful deployment retrieval
	mockDeployer.On("GetDeployment", mock.Anything, "123").Return(&types.DeploymentResponse{
		ID: "123",
		Status: &types.DeploymentStatus{
			State: types.DeploymentStateRunning,
		},
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/deployments/123", nil)
	req.Header.Set("X-AuthChain", authChainStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.DeploymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "123", response.ID)

	mockDeployer.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
	mockBidMarket.AssertExpectations(t)
}

func TestDeploymentHandler_CleanupDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}

	// Mock config values
	mockConfig.On("GetString", "provider.id", "").Return("1")
	mockConfig.On("GetString", "provider.machine_id", "").Return("1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	// Create authchain and get owner address
	authChainStr := createTestAuthChain("123")
	ownerAddress := getOwnerAddressFromAuthChain(authChainStr)

	// Mock bid market to return order information
	mockOrder := &bidenginetypes.Order{
		ID:                 big.NewInt(123),
		Owner:              common.HexToAddress(ownerAddress),
		AcceptedProviderId: big.NewInt(1),
		AcceptedMachineId:  big.NewInt(1), // This should match the machine_id from config
		Status:             bidenginetypes.OrderStatusAccepted,
	}
	mockBidMarket.On("GetOrder", mock.Anything, big.NewInt(123)).Return(mockOrder, nil)

	// Mock successful deployment cleanup
	mockDeployer.On("CleanupDeployment", mock.Anything, "123").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/deployments/123", nil)
	req.Header.Set("X-AuthChain", authChainStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Deployment cleaned up successfully", response["message"])

	mockDeployer.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
	mockBidMarket.AssertExpectations(t)
}

func TestDeploymentHandler_CreateDeploymentHandler_OrderIDMismatch(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}

	// Mock config values
	mockConfig.On("GetString", "provider.id", "").Return("1")
	mockConfig.On("GetString", "provider.machine_id", "").Return("1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	// Create authchain and get owner address
	authChainStr := createTestAuthChain("123")
	ownerAddress := getOwnerAddressFromAuthChain(authChainStr)

	// Mock bid market to return order information for order 456 (the one in request body)
	mockOrder := &bidenginetypes.Order{
		ID:                 big.NewInt(456),
		Owner:              common.HexToAddress(ownerAddress),
		AcceptedProviderId: big.NewInt(1),
		AcceptedMachineId:  big.NewInt(1), // This should match the machine_id from config
		Status:             bidenginetypes.OrderStatusAccepted,
	}
	mockBidMarket.On("GetOrder", mock.Anything, big.NewInt(456)).Return(mockOrder, nil)

	requestBody := DeploymentRequest{
		OrderID:  "456", // Different from authchain
		Manifest: manifest.SDL{},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/deployments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AuthChain", authChainStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 because orderID doesn't match
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Unauthorized: invalid authchain")

	mockConfig.AssertExpectations(t)
	mockBidMarket.AssertExpectations(t)
}
