package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

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

func TestDeploymentHandler_HealthHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	handler := NewDeploymentHandler(mockDeployer)
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

func TestDeploymentHandler_CreateDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	handler := NewDeploymentHandler(mockDeployer)
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
	req := httptest.NewRequest("POST", "/api/v1/deployments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response types.DeploymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-order-123", response.ID)

	mockDeployer.AssertExpectations(t)
}

func TestDeploymentHandler_GetDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	handler := NewDeploymentHandler(mockDeployer)
	router := handler.Router()

	// Mock successful deployment retrieval
	mockDeployer.On("GetDeployment", mock.Anything, "test-order-123").Return(&types.DeploymentResponse{
		ID: "test-order-123",
		Status: &types.DeploymentStatus{
			State: types.DeploymentStateRunning,
		},
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/deployments/test-order-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.DeploymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-order-123", response.ID)

	mockDeployer.AssertExpectations(t)
}

func TestDeploymentHandler_ListDeploymentsHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	handler := NewDeploymentHandler(mockDeployer)
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

	req := httptest.NewRequest("GET", "/api/v1/deployments?requester=0x1234567890123456789012345678901234567890", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*types.DeploymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "test-order-123", response[0].ID)

	mockDeployer.AssertExpectations(t)
}

func TestDeploymentHandler_CleanupDeploymentHandler(t *testing.T) {
	mockDeployer := &MockDeployerService{}
	handler := NewDeploymentHandler(mockDeployer)
	router := handler.Router()

	// Mock successful deployment cleanup
	mockDeployer.On("CleanupDeployment", mock.Anything, "test-order-123").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/deployments/test-order-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Deployment cleaned up successfully", response["message"])

	mockDeployer.AssertExpectations(t)
}
