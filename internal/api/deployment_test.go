package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	bidenginetypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
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

// createTestAuthChain creates a simple authchain for testing
func createTestAuthChain(orderID string) string {
	chain := authchain.AuthChain{
		{
			Type:    authchain.AuthLinkTypeSIGNER,
			Payload: "0x1234567890123456789012345678901234567890",
		},
		{
			Type:      authchain.AuthLinkTypeECDSA_PERSONAL_EPHEMERAL,
			Payload:   "delegate:0x1234567890123456789012345678901234567890123456789012345678901234:1234567890",
			Signature: "0xdummysignature",
		},
		{
			Type:    authchain.AuthLinkTypeECDSA_PERSONAL_SIGNED_ENTITY,
			Payload: fmt.Sprintf("subnet_deployment:provider-1:machine-1:%s", orderID),
		},
	}
	serialized, _ := authchain.SerializeAuthChain(chain)
	return serialized
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
		OrderID:   "test-order-123",
		Requester: "0x1234567890123456789012345678901234567890",
		TTL:       60,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/deployments", bytes.NewBuffer(bodyBytes))
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

	req := httptest.NewRequest("GET", "/deployments/test-order-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 401 because no authchain header
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "X-AuthChain header required", response["error"])
}

func TestDeploymentHandler_ListDeploymentsHandler_WithoutAuth(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}
	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	req := httptest.NewRequest("GET", "/deployments?requester=0x1234567890123456789012345678901234567890", nil)
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

	req := httptest.NewRequest("DELETE", "/deployments/test-order-123", nil)
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
	mockConfig.On("GetString", "deployment.provider.id", "").Return("provider-1")
	mockConfig.On("GetString", "deployment.machine.id", "").Return("machine-1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	// Mock successful deployment creation
	mockDeployer.On("RequestDeployment", mock.Anything, mock.Anything).Return(&types.DeploymentResponse{
		ID: "test-order-123",
		Status: &types.DeploymentStatus{
			State: types.DeploymentStateRunning,
		},
	}, nil)

	requestBody := types.DeploymentRequest{
		OrderID:   "test-order-123",
		Requester: "0x1234567890123456789012345678901234567890",
		TTL:       60,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/deployments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AuthChain", createTestAuthChain("test-order-123"))
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
	assert.Equal(t, "test-order-123", response.ID)

	mockDeployer.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
}

func TestDeploymentHandler_GetDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}

	// Mock config values
	mockConfig.On("GetString", "deployment.provider.id", "").Return("provider-1")
	mockConfig.On("GetString", "deployment.machine.id", "").Return("machine-1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	// Mock successful deployment retrieval
	mockDeployer.On("GetDeployment", mock.Anything, "test-order-123").Return(&types.DeploymentResponse{
		ID: "test-order-123",
		Status: &types.DeploymentStatus{
			State: types.DeploymentStateRunning,
		},
	}, nil)

	req := httptest.NewRequest("GET", "/deployments/test-order-123", nil)
	req.Header.Set("X-AuthChain", createTestAuthChain("test-order-123"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.DeploymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-order-123", response.ID)

	mockDeployer.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
}

func TestDeploymentHandler_ListDeploymentsHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}

	// Mock config values
	mockConfig.On("GetString", "deployment.provider.id", "").Return("provider-1")
	mockConfig.On("GetString", "deployment.machine.id", "").Return("machine-1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	// Mock successful deployments list
	mockDeployer.On("GetDeployments", mock.Anything, "0x1234567890123456789012345678901234567890").Return([]*types.DeploymentResponse{
		{
			ID: "test-order-123",
			Status: &types.DeploymentStatus{
				State: types.DeploymentStateRunning,
			},
		},
	}, nil)

	req := httptest.NewRequest("GET", "/deployments?requester=0x1234567890123456789012345678901234567890", nil)
	req.Header.Set("X-AuthChain", createTestAuthChain("test-order-123"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*types.DeploymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "test-order-123", response[0].ID)

	mockDeployer.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
}

func TestDeploymentHandler_CleanupDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}

	// Mock config values
	mockConfig.On("GetString", "deployment.provider.id", "").Return("provider-1")
	mockConfig.On("GetString", "deployment.machine.id", "").Return("machine-1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	// Mock successful deployment cleanup
	mockDeployer.On("CleanupDeployment", mock.Anything, "test-order-123").Return(nil)

	req := httptest.NewRequest("DELETE", "/deployments/test-order-123", nil)
	req.Header.Set("X-AuthChain", createTestAuthChain("test-order-123"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Deployment cleaned up successfully", response["message"])

	mockDeployer.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
}

func TestDeploymentHandler_CreateDeploymentHandler_RequesterMismatch(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}

	// Mock config values
	mockConfig.On("GetString", "deployment.provider.id", "").Return("provider-1")
	mockConfig.On("GetString", "deployment.machine.id", "").Return("machine-1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	requestBody := types.DeploymentRequest{
		OrderID:   "test-order-123",
		Requester: "0x9876543210987654321098765432109876543210", // Different from authchain
		TTL:       60,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/deployments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AuthChain", createTestAuthChain("test-order-123"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 403 because requester doesn't match authenticated user
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "requester does not match authenticated user")

	mockConfig.AssertExpectations(t)
}

func TestDeploymentHandler_CreateDeploymentHandler_OrderIDMismatch(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}

	// Mock config values
	mockConfig.On("GetString", "deployment.provider.id", "").Return("provider-1")
	mockConfig.On("GetString", "deployment.machine.id", "").Return("machine-1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	requestBody := types.DeploymentRequest{
		OrderID:   "different-order-456", // Different from authchain
		Requester: "0x1234567890123456789012345678901234567890",
		TTL:       60,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/deployments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AuthChain", createTestAuthChain("test-order-123"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 because orderID doesn't match
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "OrderID mismatch")

	mockConfig.AssertExpectations(t)
}

func TestDeploymentHandler_GetDeploymentHandler_MissingOrderID(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	mockConfig := &MockConfigProvider{}
	mockBidMarket := &MockBidMarketContract{}

	// Mock config values
	mockConfig.On("GetString", "deployment.provider.id", "").Return("provider-1")
	mockConfig.On("GetString", "deployment.machine.id", "").Return("machine-1")

	handler := NewDeploymentHandler(mockDeployer, mockConfig, mockBidMarket)
	router := handler.Router()

	req := httptest.NewRequest("GET", "/deployments", nil) // No orderID in path
	req.Header.Set("X-AuthChain", createTestAuthChain("test-order-123"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 because orderID is required for GET operations
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "orderID is required")

	mockConfig.AssertExpectations(t)
}
