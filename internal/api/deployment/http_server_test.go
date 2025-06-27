package deployment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	authchain "github.com/unicornultrafoundation/subnet-node/crypto/authchain"
)

// GenerateValidAuthChainBase64 creates a valid AuthChain string (base64) for testing purposes
// It generates a new private key, creates an entity ID following the schema,
// and returns a serialized auth chain that can be used in test requests
func GenerateValidAuthChainBase64(orderID string, expiryMinutes int) (string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", err
	}

	// Create entityID according to schema: "subnet_deployment:{providerId}:{machineId}:{order_id}"
	entityID := fmt.Sprintf("subnet_deployment:provider-1:machine-1:%s", orderID)

	// Convert private key to hex string
	privateKeyHex := fmt.Sprintf("0x%x", crypto.FromECDSA(privateKey))

	authChain, err := authchain.CreateAuthChainFromPrivateKey(privateKeyHex, entityID, time.Duration(expiryMinutes)*time.Minute)
	if err != nil {
		return "", err
	}
	authChainBase64, err := authchain.SerializeAuthChain(authChain)
	if err != nil {
		return "", err
	}
	return authChainBase64, nil
}

// TestNewHTTPServer tests the creation of a new HTTP server instance
// Verifies that all components are properly initialized
func TestNewHTTPServer(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := &config.C{}

	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	assert.NotNil(t, server)
	assert.Equal(t, logger, server.logger)
	assert.Equal(t, cfg, server.cfg)
	assert.Equal(t, mockDeployer, server.deployer)
	assert.NotNil(t, server.router)
}

// TestHTTPServer_HealthHandler tests the health check endpoint
// Verifies that the health endpoint returns the correct status and service name
func TestHTTPServer_HealthHandler(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := &config.C{}
	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", server.healthHandler)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "deployment-api", response["service"])
}

// TestHTTPServer_AuthMiddleware_NoAuthChainHeader tests authentication middleware
// when no X-AuthChain header is provided in the request
func TestHTTPServer_AuthMiddleware_NoAuthChainHeader(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := &config.C{}
	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(server.authMiddleware)
	router.GET("/api/v1/deployments/:orderID", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/deployments/order-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "X-AuthChain header required", response["error"])
}

// TestHTTPServer_AuthMiddleware_InvalidAuthChain tests authentication middleware
// when an invalid auth chain is provided in the X-AuthChain header
func TestHTTPServer_AuthMiddleware_InvalidAuthChain(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := &config.C{}
	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(server.authMiddleware)
	router.GET("/api/v1/deployments/:orderID", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/deployments/order-123", nil)
	req.Header.Set("X-AuthChain", "invalid-base64")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid authchain format")
}

// newTestConfigWithProviderAndMachine creates a test configuration
// with provider and machine IDs for testing authentication
func newTestConfigWithProviderAndMachine() *config.C {
	return &config.C{
		Settings: map[string]any{
			"deployment": map[string]any{
				"provider": map[string]any{
					"id": "provider-1",
				},
				"machine": map[string]any{
					"id": "machine-1",
				},
			},
		},
	}
}

// TestHTTPServer_AuthMiddleware_ValidAuthChain tests authentication middleware
// with a valid auth chain and verifies that user context is properly set
func TestHTTPServer_AuthMiddleware_ValidAuthChain(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := newTestConfigWithProviderAndMachine()
	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	authChainBase64, err := GenerateValidAuthChainBase64("order-123", 60)
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(server.authMiddleware)
	router.GET("/api/v1/deployments/:orderID", func(c *gin.Context) {
		userAddress, _ := c.Get(string(userAddressKey))
		orderID, _ := c.Get(string(orderIDKey))
		providerId, _ := c.Get(string(providerIDKey))
		machineId, _ := c.Get(string(machineIDKey))

		assert.NotEmpty(t, userAddress)
		assert.Equal(t, "order-123", orderID)
		assert.Equal(t, "provider-1", providerId)
		assert.Equal(t, "machine-1", machineId)

		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/deployments/order-123", nil)
	req.Header.Set("X-AuthChain", authChainBase64)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestHTTPServer_CreateDeploymentHandler tests the deployment creation endpoint
// Verifies that a deployment request is properly processed and returns the expected response
func TestHTTPServer_CreateDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := newTestConfigWithProviderAndMachine()
	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	// Mock successful deployment creation
	mockDeployer.On("RequestDeployment", mock.Anything, mock.Anything).Return(&types.DeploymentResponse{
		ID:        "order-123",
		Requester: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Status:    &types.DeploymentStatus{State: types.DeploymentStateRunning},
	}, nil)

	requestBody := types.DeploymentRequest{
		OrderID:   "order-123",
		Requester: "0x1234567890123456789012345678901234567890",
		TTL:       60,
	}

	bodyBytes, _ := json.Marshal(requestBody)

	authChainBase64, err := GenerateValidAuthChainBase64("order-123", 60)
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(server.authMiddleware)
	router.POST("/api/v1/deployments", server.createDeploymentHandler)

	req := httptest.NewRequest("POST", "/api/v1/deployments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AuthChain", authChainBase64)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "order-123", response["id"])
	assert.Equal(t, "running", response["status"].(map[string]interface{})["state"])

	mockDeployer.AssertExpectations(t)
}

// TestHTTPServer_GetDeploymentHandler tests the deployment retrieval endpoint
// Verifies that deployment information is properly returned
func TestHTTPServer_GetDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := &config.C{}
	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	// Mock successful deployment retrieval
	mockDeployer.On("GetDeployment", mock.Anything, "order-123").Return(&types.DeploymentResponse{
		ID:        "order-123",
		Requester: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Status:    &types.DeploymentStatus{State: types.DeploymentStateRunning},
	}, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(server.authMiddleware)
	router.GET("/api/v1/deployments/:orderID", server.getDeploymentHandler)

	req := httptest.NewRequest("GET", "/api/v1/deployments/order-123", nil)
	req.Header.Set("X-AuthChain", "dummy-authchain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "order-123", response["id"])
	assert.Equal(t, "running", response["status"].(map[string]interface{})["state"])

	mockDeployer.AssertExpectations(t)
}

// TestHTTPServer_CleanupDeploymentHandler tests the deployment cleanup endpoint
// Verifies that deployment cleanup is properly handled
func TestHTTPServer_CleanupDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := &config.C{}
	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	// Mock successful deployment cleanup
	mockDeployer.On("CleanupDeployment", mock.Anything, "order-123").Return(nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(server.authMiddleware)
	router.DELETE("/api/v1/deployments/:orderID", server.cleanupDeploymentHandler)

	req := httptest.NewRequest("DELETE", "/api/v1/deployments/order-123", nil)
	req.Header.Set("X-AuthChain", "dummy-authchain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Deployment cleaned up successfully", response["message"])

	mockDeployer.AssertExpectations(t)
}

// TestHTTPServer_GetServiceStatusHandler tests the service status retrieval endpoint
// Verifies that service status information is properly returned
func TestHTTPServer_GetServiceStatusHandler(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := &config.C{}
	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	// Mock successful service status retrieval
	mockDeployer.On("GetServiceStatus", mock.Anything, "order-123", "web").Return(&types.ServiceStatus{
		Name:     "web",
		State:    types.ServiceStateRunning,
		Replicas: 3,
	}, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(server.authMiddleware)
	router.GET("/api/v1/deployments/:orderID/services/:serviceName/status", server.getServiceStatusHandler)

	req := httptest.NewRequest("GET", "/api/v1/deployments/order-123/services/web/status", nil)
	req.Header.Set("X-AuthChain", "dummy-authchain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "web", response["name"])
	assert.Equal(t, "running", response["state"])
	assert.Equal(t, float64(3), response["replicas"])

	mockDeployer.AssertExpectations(t)
}

// TestHTTPServer_GetDeploymentLogsHandler tests the deployment logs retrieval endpoint
// Verifies that deployment logs are properly returned
func TestHTTPServer_GetDeploymentLogsHandler(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := &config.C{}
	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	// Mock successful logs retrieval - using ServiceLog as expected by interface
	mockDeployer.On("GetDeploymentLogs", mock.Anything, "order-123").Return([]*types.ServiceLog{
		{
			Name: "web",
		},
		{
			Name: "web",
		},
	}, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(server.authMiddleware)
	router.GET("/api/v1/deployments/:orderID/logs", server.getDeploymentLogsHandler)

	req := httptest.NewRequest("GET", "/api/v1/deployments/order-123/logs", nil)
	req.Header.Set("X-AuthChain", "dummy-authchain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	mockDeployer.AssertExpectations(t)
}

// TestHTTPServer_ExecHandler tests the exec endpoint for running commands in pods
// Verifies that the exec request is properly handled and returns WebSocket connection info
func TestHTTPServer_ExecHandler(t *testing.T) {
	mockDeployer := &MockDeployer{}
	logger := logrus.New()
	cfg := &config.C{}
	server := NewHTTPServer(logger, 8080, cfg, mockDeployer)

	execRequest := map[string]interface{}{
		"podName":     "web-pod-1",
		"serviceName": "web",
		"command":     []string{"ls", "-la"},
		"tty":         false,
	}

	bodyBytes, _ := json.Marshal(execRequest)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(server.authMiddleware)
	router.POST("/api/v1/deployments/:orderID/exec", server.execHandler)

	req := httptest.NewRequest("POST", "/api/v1/deployments/order-123/exec", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AuthChain", "dummy-authchain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "order-123", response["orderID"])
	assert.Contains(t, response["message"], "WebSocket connection")
}

// MockDeployer implements the Deployer interface for testing purposes
// It uses testify/mock to provide mock implementations of all required methods
type MockDeployer struct {
	mock.Mock
}

// RequestDeployment mocks the deployment request functionality
func (m *MockDeployer) RequestDeployment(ctx context.Context, req *types.DeploymentRequest) (*types.DeploymentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*types.DeploymentResponse), args.Error(1)
}

// GetDeployment mocks the deployment retrieval functionality
func (m *MockDeployer) GetDeployment(ctx context.Context, orderID string) (*types.DeploymentResponse, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*types.DeploymentResponse), args.Error(1)
}

// GetDeployments mocks the deployments list retrieval functionality
func (m *MockDeployer) GetDeployments(ctx context.Context, requester string) ([]*types.DeploymentResponse, error) {
	args := m.Called(ctx, requester)
	return args.Get(0).([]*types.DeploymentResponse), args.Error(1)
}

// CleanupDeployment mocks the deployment cleanup functionality
func (m *MockDeployer) CleanupDeployment(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

// GetServiceStatus mocks the service status retrieval functionality
func (m *MockDeployer) GetServiceStatus(ctx context.Context, orderID, serviceName string) (*types.ServiceStatus, error) {
	args := m.Called(ctx, orderID, serviceName)
	return args.Get(0).(*types.ServiceStatus), args.Error(1)
}

// GetDeploymentLogs mocks the deployment logs retrieval functionality
func (m *MockDeployer) GetDeploymentLogs(ctx context.Context, orderID string) ([]*types.ServiceLog, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]*types.ServiceLog), args.Error(1)
}
