package deployments

import (
	"context"
	"crypto/ecdsa"
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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	bidenginetypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// MockBidMarketContract is a mock implementation of BidMarketContract
type MockBidMarketContract struct {
	mock.Mock
}

func (m *MockBidMarketContract) GetOrder(ctx context.Context, orderID *big.Int) (*bidenginetypes.Order, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*bidenginetypes.Order), args.Error(1)
}

// Implement other methods as needed for the interface
func (m *MockBidMarketContract) GetOrderCount(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockBidMarketContract) OrderCount(ctx context.Context) (*big.Int, error) {
	return m.GetOrderCount(ctx)
}

func (m *MockBidMarketContract) Orders(ctx context.Context, orderID *big.Int) (*bidenginetypes.Order, error) {
	return m.GetOrder(ctx, orderID)
}

func (m *MockBidMarketContract) GetBids(ctx context.Context, orderID *big.Int) ([]bidenginetypes.Bid, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]bidenginetypes.Bid), args.Error(1)
}

func (m *MockBidMarketContract) IsBiddingOpen(ctx context.Context, orderID *big.Int) (bool, error) {
	args := m.Called(ctx, orderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockBidMarketContract) GetRemainingBidTime(ctx context.Context, orderID *big.Int) (*big.Int, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockBidMarketContract) SubmitBid(ctx context.Context, orderID *big.Int, pricePerSecond *big.Int, providerID *big.Int, machineID *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID, pricePerSecond, providerID, machineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) GetBidIndexFromTransaction(ctx context.Context, tx *types.Transaction, orderID *big.Int) (*big.Int, error) {
	args := m.Called(ctx, tx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockBidMarketContract) CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID, bidIndex)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) AcceptBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID, bidIndex)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) CancelOrder(ctx context.Context, orderID *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) CloseOrder(ctx context.Context, orderID *big.Int, reason string) (*types.Transaction, error) {
	args := m.Called(ctx, orderID, reason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) ExtendOrder(ctx context.Context, orderID *big.Int, amount *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) GetUsedResource(ctx context.Context, providerID *big.Int, machineID *big.Int) (*bidenginetypes.ResourceUsage, error) {
	args := m.Called(ctx, providerID, machineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*bidenginetypes.ResourceUsage), args.Error(1)
}

func (m *MockBidMarketContract) ReleaseOrderResource(ctx context.Context, orderID *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) WatchOrderCreated(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) WatchOrderClosed(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) WatchOrderExpired(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

// ConsoleSession represents an execution session
type ConsoleSession interface {
	// Execute executes a command and returns the result
	Execute(ctx context.Context, command []string) (*ConsoleResult, error)
	// Close closes the session
	Close() error
}

// ConsoleResult represents the result of a console execution
type ConsoleResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// MockBidEvent is a mock implementation of bidenginetypes.BidEvent
type MockBidEvent struct {
	OrderID        *big.Int
	BidIndex       *big.Int
	Provider       common.Address
	PricePerSecond *big.Int
}

func (m *MockBidMarketContract) WatchBidSubmitted(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) WatchBidAccepted(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) WatchBidCancelled(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) OrderBids(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*bidenginetypes.Bid, error) {
	args := m.Called(ctx, orderID, bidIndex)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*bidenginetypes.Bid), args.Error(1)
}

// MockServiceManager is a mock implementation of ServiceManager
type MockServiceManager struct {
	mock.Mock
}

func (m *MockServiceManager) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockServiceManager) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockServiceManager) CreateDeployment(ctx context.Context, deployment *Deployment) error {
	args := m.Called(ctx, deployment)
	return args.Error(0)
}

func (m *MockServiceManager) GetDeployment(ctx context.Context, id string) (*Deployment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Deployment), args.Error(1)
}

func (m *MockServiceManager) ListDeployments(ctx context.Context) ([]*Deployment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Deployment), args.Error(1)
}

func (m *MockServiceManager) UpdateDeployment(ctx context.Context, deployment *Deployment) error {
	args := m.Called(ctx, deployment)
	return args.Error(0)
}

func (m *MockServiceManager) DeleteDeployment(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServiceManager) StartDeployment(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServiceManager) StopDeployment(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServiceManager) RestartDeployment(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServiceManager) InspectDeployment(ctx context.Context, id string) (*DeploymentInspection, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DeploymentInspection), args.Error(1)
}

func (m *MockServiceManager) InspectService(ctx context.Context, deploymentID, serviceName string) (*ServiceInspection, error) {
	args := m.Called(ctx, deploymentID, serviceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ServiceInspection), args.Error(1)
}

func (m *MockServiceManager) GetDeploymentMetrics(ctx context.Context, id string, duration time.Duration) (*DeploymentMetrics, error) {
	args := m.Called(ctx, id, duration)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DeploymentMetrics), args.Error(1)
}

func (m *MockServiceManager) GetServiceMetrics(ctx context.Context, deploymentID string, serviceName string, duration time.Duration) (*ServiceMetrics, error) {
	args := m.Called(ctx, deploymentID, serviceName, duration)
	return args.Get(0).(*ServiceMetrics), args.Error(1)
}

func (m *MockServiceManager) GetDeploymentLogs(ctx context.Context, id, serviceName string, tail int) (io.ReadCloser, error) {
	args := m.Called(ctx, id, serviceName, tail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockServiceManager) StreamDeploymentLogs(ctx context.Context, id, serviceName string, follow bool) (<-chan LogEntry, error) {
	args := m.Called(ctx, id, serviceName, follow)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan LogEntry), args.Error(1)
}

func (m *MockServiceManager) ExecConsole(ctx context.Context, deploymentID, serviceName string, command []string, tty bool) (ExecSession, error) {
	args := m.Called(ctx, deploymentID, serviceName, command, tty)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(ExecSession), args.Error(1)
}

func (m *MockServiceManager) UpdateDeploymentImage(ctx context.Context, deploymentID string, serviceName string, image string) error {
	args := m.Called(ctx, deploymentID, serviceName, image)
	return args.Error(0)
}

func (m *MockServiceManager) GetDeploymentEvents(ctx context.Context, deploymentID string, limit int) ([]*DeploymentEvent, error) {
	args := m.Called(ctx, deploymentID, limit)
	return args.Get(0).([]*DeploymentEvent), args.Error(1)
}

func (m *MockServiceManager) StreamDeploymentEvents(ctx context.Context, deploymentID string) (<-chan *DeploymentEvent, error) {
	args := m.Called(ctx, deploymentID)
	return args.Get(0).(<-chan *DeploymentEvent), args.Error(1)
}

// MockConsoleSession is a mock implementation of ConsoleSession
type MockConsoleSession struct {
	mock.Mock
}

func (m *MockConsoleSession) Execute(ctx context.Context, command []string) (*ConsoleResult, error) {
	args := m.Called(ctx, command)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ConsoleResult), args.Error(1)
}

func (m *MockConsoleSession) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Helper function to create a proper signature for testing
func createTestSignature(privateKey *ecdsa.PrivateKey, message string) string {
	prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	messageHash := crypto.Keccak256Hash([]byte(prefixed))
	signature, err := crypto.Sign(messageHash.Bytes(), privateKey)
	if err != nil {
		panic(err)
	}
	// Do NOT adjust signature[64] (recovery ID)
	return "0x" + hex.EncodeToString(signature)
}

// Helper function to create authorization header for testing
func createAuthHeader(privateKey *ecdsa.PrivateKey, deploymentID, action string, timestamp int64) string {
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	message := fmt.Sprintf("deployment:%s:%s:%d", deploymentID, action, timestamp)
	signature := createTestSignature(privateKey, message)

	authData := AuthorizationRequest{
		Signature: signature,
		Address:   address.Hex(),
		Message:   message,
		Timestamp: timestamp,
	}
	authJSON, _ := json.Marshal(authData)
	return string(authJSON)
}

func TestAPI_validateDeploymentAuthorization(t *testing.T) {
	// Create test private key
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// Get the public address
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Create mock bid market contract
	mockBidMarket := &MockBidMarketContract{}

	// Create mock service manager
	mockService := &MockServiceManager{}

	// Create logger
	logger := logrus.New()

	// Create API instance
	api := NewAPI(mockService, mockBidMarket, logger)

	tests := []struct {
		name           string
		deploymentID   string
		authHeader     string
		setupMock      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing authorization header",
			deploymentID:   "123",
			authHeader:     "",
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Missing authorization header",
		},
		{
			name:           "Invalid authorization format",
			deploymentID:   "123",
			authHeader:     "invalid json",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid authorization format",
		},
		{
			name:           "Authorization timestamp expired",
			deploymentID:   "123",
			authHeader:     `{"signature":"0x","address":"0x","message":"","timestamp":0}`,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization timestamp expired",
		},
		{
			name:           "Invalid deployment ID format",
			deploymentID:   "invalid",
			authHeader:     createAuthHeader(privateKey, "invalid", "get", time.Now().Unix()),
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid deployment ID format",
		},
		{
			name:         "Order not found",
			deploymentID: "123",
			authHeader:   createAuthHeader(privateKey, "123", "get", time.Now().Unix()),
			setupMock: func() {
				mockBidMarket.On("GetOrder", mock.Anything, big.NewInt(123)).Return(nil, fmt.Errorf("order not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to get order information",
		},
		{
			name:         "Not the deployment owner",
			deploymentID: "123",
			authHeader:   createAuthHeader(privateKey, "123", "get", time.Now().Unix()),
			setupMock: func() {
				order := &bidenginetypes.Order{
					Owner: common.HexToAddress("0x0987654321098765432109876543210987654321"),
				}
				mockBidMarket.On("GetOrder", mock.Anything, big.NewInt(123)).Return(order, nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Not authorized: not the deployment owner",
		},
		{
			name:         "Valid authorization",
			deploymentID: "123",
			authHeader:   createAuthHeader(privateKey, "123", "get", time.Now().Unix()),
			setupMock: func() {
				order := &bidenginetypes.Order{
					Owner: address,
				}
				mockBidMarket.On("GetOrder", mock.Anything, big.NewInt(123)).Return(order, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockBidMarket.ExpectedCalls = nil
			mockService.ExpectedCalls = nil

			// Setup mock
			tt.setupMock()

			// Create request
			req := httptest.NewRequest("GET", "/api/v1/deployments/"+tt.deploymentID, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the validation function
			_, authorized := api.validateDeploymentAuthorization(w, req, tt.deploymentID)

			// Check response
			if tt.expectedStatus != http.StatusOK {
				assert.False(t, authorized)
				assert.Equal(t, tt.expectedStatus, w.Code)

				var errorResp ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResp.Error)
			} else {
				assert.True(t, authorized)
			}

			// Verify mocks
			mockBidMarket.AssertExpectations(t)
		})
	}
}

func TestAPI_validateDeploymentAuthorization_TimestampValidation(t *testing.T) {
	// Create test private key
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// Get the public address
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Create mock bid market contract
	mockBidMarket := &MockBidMarketContract{}

	// Create mock service manager
	mockService := &MockServiceManager{}

	// Create logger
	logger := logrus.New()

	// Create API instance
	api := NewAPI(mockService, mockBidMarket, logger)

	// Setup mock for valid order
	order := &bidenginetypes.Order{
		Owner: address,
	}
	mockBidMarket.On("GetOrder", mock.Anything, big.NewInt(123)).Return(order, nil)

	tests := []struct {
		name           string
		timestamp      int64
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Timestamp too old (6 minutes ago)",
			timestamp:      time.Now().Add(-6 * time.Minute).Unix(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization timestamp expired",
		},
		{
			name:           "Timestamp too future (6 minutes ahead)",
			timestamp:      time.Now().Add(6 * time.Minute).Unix(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization timestamp expired",
		},
		{
			name:           "Timestamp exactly at limit (5 minutes ago)",
			timestamp:      time.Now().Add(-5 * time.Minute).Unix(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization timestamp expired",
		},
		{
			name:           "Timestamp exactly at limit (5 minutes ahead)",
			timestamp:      time.Now().Add(5 * time.Minute).Unix(),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization timestamp expired",
		},
		{
			name:           "Timestamp within tolerance (4 minutes ago)",
			timestamp:      time.Now().Add(-4 * time.Minute).Unix(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Timestamp within tolerance (4 minutes ahead)",
			timestamp:      time.Now().Add(4 * time.Minute).Unix(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Timestamp current",
			timestamp:      time.Now().Unix(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Timestamp zero (epoch)",
			timestamp:      0,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization timestamp expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockBidMarket.ExpectedCalls = nil
			mockService.ExpectedCalls = nil

			// Only set up mock if timestamp is within tolerance
			if tt.expectedStatus == http.StatusOK {
				mockBidMarket.On("GetOrder", mock.Anything, big.NewInt(123)).Return(order, nil)
			}

			// Create request with specific timestamp
			authHeader := createAuthHeader(privateKey, "123", "get", tt.timestamp)

			req := httptest.NewRequest("GET", "/api/v1/deployments/123", nil)
			req.Header.Set("Authorization", authHeader)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the validation function
			_, authorized := api.validateDeploymentAuthorization(w, req, "123")

			// Check response
			if tt.expectedStatus != http.StatusOK {
				assert.False(t, authorized)
				assert.Equal(t, tt.expectedStatus, w.Code)

				var errorResp ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResp.Error)
			} else {
				assert.True(t, authorized)
				mockBidMarket.AssertExpectations(t)
			}
		})
	}
}

func TestAPI_verifySignature(t *testing.T) {
	// Create test private key
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// Get the public address
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Create API instance
	api := NewAPI(nil, nil, logrus.New())

	tests := []struct {
		name           string
		message        string
		signature      string
		expectedAddr   common.Address
		expectedResult bool
	}{
		{
			name:           "Valid signature",
			message:        "test message",
			expectedAddr:   address,
			expectedResult: true,
		},
		{
			name:           "Invalid signature",
			message:        "test message",
			signature:      "0xinvalid",
			expectedAddr:   address,
			expectedResult: false,
		},
		{
			name:           "Wrong address",
			message:        "test message",
			expectedAddr:   common.HexToAddress("0x1234567890123456789012345678901234567890"),
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var signature string
			if tt.name == "Valid signature" {
				// Create valid signature using helper function
				signature = createTestSignature(privateKey, tt.message)
			} else if tt.name == "Invalid signature" {
				signature = tt.signature
			} else {
				// Wrong address test - use valid signature but wrong address
				signature = createTestSignature(privateKey, tt.message)
			}

			result := api.verifySignature(tt.message, signature, tt.expectedAddr)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestAPI_getDeployment_WithAuthorization(t *testing.T) {
	// Create test private key
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// Get the public address
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Create mock bid market contract
	mockBidMarket := &MockBidMarketContract{}

	// Create mock service manager
	mockService := &MockServiceManager{}

	// Create logger
	logger := logrus.New()

	// Create API instance
	api := NewAPI(mockService, mockBidMarket, logger)

	// Setup mocks
	deploymentID := "123"
	order := &bidenginetypes.Order{
		Owner: address,
	}
	mockBidMarket.On("GetOrder", mock.Anything, big.NewInt(123)).Return(order, nil)

	expectedDeployment := &Deployment{
		ID:   deploymentID,
		Name: "test-deployment",
	}
	mockService.On("GetDeployment", mock.Anything, deploymentID).Return(expectedDeployment, nil)

	// Create request with valid authorization using helper function
	authHeader := createAuthHeader(privateKey, deploymentID, "get", time.Now().Unix())

	req := httptest.NewRequest("GET", "/api/v1/deployments/"+deploymentID, nil)
	req.Header.Set("Authorization", authHeader)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create router and register routes
	router := mux.NewRouter()
	api.RegisterRoutes(router)

	// Serve the request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Verify mocks
	mockBidMarket.AssertExpectations(t)
	mockService.AssertExpectations(t)
}

func TestAPI_getDeployment_WithoutAuthorization(t *testing.T) {
	// Create mock service manager
	mockService := &MockServiceManager{}

	// Create logger
	logger := logrus.New()

	// Create API instance
	api := NewAPI(mockService, nil, logger)

	// Create request without authorization
	deploymentID := "123"
	req := httptest.NewRequest("GET", "/api/v1/deployments/"+deploymentID, nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create router and register routes
	router := mux.NewRouter()
	api.RegisterRoutes(router)

	// Serve the request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Missing authorization header", response.Error)
}
