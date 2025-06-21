package bidengine

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ds "github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBidMarketContract is a mock implementation of BidMarketContract
type MockBidMarketContract struct {
	mock.Mock
}

func (m *MockBidMarketContract) GetOrder(ctx context.Context, orderID *big.Int) (*Order, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*Order), args.Error(1)
}

func (m *MockBidMarketContract) GetOrderCount(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockBidMarketContract) GetBids(ctx context.Context, orderID *big.Int) ([]Bid, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]Bid), args.Error(1)
}

func (m *MockBidMarketContract) IsBiddingOpen(ctx context.Context, orderID *big.Int) (bool, error) {
	args := m.Called(ctx, orderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockBidMarketContract) GetRemainingBidTime(ctx context.Context, orderID *big.Int) (*big.Int, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockBidMarketContract) SubmitBid(ctx context.Context, orderID *big.Int, pricePerSecond *big.Int, providerID *big.Int, machineID *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID, pricePerSecond, providerID, machineID)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) GetBidIndexFromTransaction(ctx context.Context, tx *types.Transaction, orderID *big.Int) (*big.Int, error) {
	args := m.Called(ctx, tx, orderID)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockBidMarketContract) CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID, bidIndex)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) AcceptBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID, bidIndex)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) CancelOrder(ctx context.Context, orderID *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) CloseOrder(ctx context.Context, orderID *big.Int, reason string) (*types.Transaction, error) {
	args := m.Called(ctx, orderID, reason)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) ExtendOrder(ctx context.Context, orderID *big.Int, amount *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID, amount)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) GetUsedResource(ctx context.Context, providerID *big.Int, machineID *big.Int) (*ResourceUsage, error) {
	args := m.Called(ctx, providerID, machineID)
	return args.Get(0).(*ResourceUsage), args.Error(1)
}

func (m *MockBidMarketContract) ReleaseOrderResource(ctx context.Context, orderID *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBidMarketContract) WatchOrderCreated(ctx context.Context, sink chan<- *OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) WatchOrderClosed(ctx context.Context, sink chan<- *OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) WatchOrderExpired(ctx context.Context, sink chan<- *OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) WatchBidSubmitted(ctx context.Context, sink chan<- *OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) WatchBidAccepted(ctx context.Context, sink chan<- *OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) WatchBidCancelled(ctx context.Context, sink chan<- *OrderEvent) error {
	args := m.Called(ctx, sink)
	return args.Error(0)
}

func (m *MockBidMarketContract) OrderBids(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*Bid, error) {
	args := m.Called(ctx, orderID, bidIndex)
	return args.Get(0).(*Bid), args.Error(1)
}

// MockLogger is a mock implementation of Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

// MockMetrics is a mock implementation of Metrics
type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) IncrementBidsSubmitted() {
	m.Called()
}

func (m *MockMetrics) IncrementBidsAccepted() {
	m.Called()
}

func (m *MockMetrics) IncrementBidsRejected() {
	m.Called()
}

func (m *MockMetrics) RecordBidLatency(duration float64) {
	m.Called(duration)
}

func (m *MockMetrics) IncrementOrdersTracked() {
	m.Called()
}

func (m *MockMetrics) IncrementOrdersCompleted() {
	m.Called()
}

func (m *MockMetrics) RecordOrderDuration(duration float64) {
	m.Called(duration)
}

func (m *MockMetrics) RecordResourceUtilization(usage *ResourceUsage) {
	m.Called(usage)
}

func (m *MockMetrics) RecordResourceAllocation(usage *ResourceUsage) {
	m.Called(usage)
}

func (m *MockMetrics) RecordRevenue(amount *big.Int) {
	m.Called(amount)
}

func (m *MockMetrics) RecordProfit(amount *big.Int) {
	m.Called(amount)
}

func (m *MockMetrics) RecordCost(amount *big.Int) {
	m.Called(amount)
}

// MockResourceManagerService is a mock implementation of ResourceManager
type MockResourceManagerService struct {
	mock.Mock
}

func (m *MockResourceManagerService) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockResourceManagerService) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockResourceManagerService) AllocateResources(ctx context.Context, orderID *big.Int, machine *Machine, usage *ResourceUsage) error {
	args := m.Called(ctx, orderID, machine, usage)
	return args.Error(0)
}

func (m *MockResourceManagerService) DeallocateResources(ctx context.Context, orderID *big.Int) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockResourceManagerService) GetCurrentUsage(ctx context.Context, machine *Machine) (*ResourceUsage, error) {
	args := m.Called(ctx, machine)
	return args.Get(0).(*ResourceUsage), args.Error(1)
}

func (m *MockResourceManagerService) GetAvailableResources(ctx context.Context, machine *Machine) (*ResourceUsage, error) {
	args := m.Called(ctx, machine)
	return args.Get(0).(*ResourceUsage), args.Error(1)
}

func (m *MockResourceManagerService) CanAllocateResources(ctx context.Context, machine *Machine, required *ResourceUsage) (bool, error) {
	args := m.Called(ctx, machine, required)
	return args.Bool(0), args.Error(1)
}

func (m *MockResourceManagerService) StartResource(ctx context.Context, orderID *big.Int, machine *Machine) error {
	args := m.Called(ctx, orderID, machine)
	return args.Error(0)
}

func (m *MockResourceManagerService) StopResource(ctx context.Context, orderID *big.Int) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockResourceManagerService) RegisterMachine(ctx context.Context, machine *Machine) error {
	args := m.Called(ctx, machine)
	return args.Error(0)
}

func (m *MockResourceManagerService) UnregisterMachine(ctx context.Context, machineID *big.Int) error {
	args := m.Called(ctx, machineID)
	return args.Error(0)
}

func (m *MockResourceManagerService) GetAllMachines(ctx context.Context) []*Machine {
	args := m.Called(ctx)
	return args.Get(0).([]*Machine)
}

func (m *MockResourceManagerService) GetMachine(ctx context.Context, machineID *big.Int) (*Machine, error) {
	args := m.Called(ctx, machineID)
	return args.Get(0).(*Machine), args.Error(1)
}

func TestSubmitBidUsesProviderIDFromConfig(t *testing.T) {
	// Create test configuration with a specific provider ID
	config := &BidEngineConfig{
		BidMarketAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ProviderAddress:  common.HexToAddress("0x0987654321098765432109876543210987654321"),
		ProviderID:       big.NewInt(42), // Use a specific provider ID for testing
		ProviderWallet:   common.HexToAddress("0x1111111111111111111111111111111111111111"),
		BidStrategy: BidStrategy{
			MinProfitMargin:   0.05,
			MaxProfitMargin:   0.20,
			CompetitiveFactor: 0.10,
			MarketAdjustment:  0.05,
			ResourceWeight: ResourceWeight{
				CPU:     0.25,
				GPU:     0.35,
				Memory:  0.20,
				Disk:    0.15,
				Network: 0.05,
			},
		},
		MaxConcurrentBids: 10,
		BidTimeout:        30 * time.Second,
		OrderSyncInterval: 30 * time.Second,
		BidCheckInterval:  60 * time.Second,
		LogLevel:          "INFO",
		LogFile:           "",
	}

	// Create mocks
	mockBidMarket := &MockBidMarketContract{}
	mockLogger := &MockLogger{}
	mockMetrics := &MockMetrics{}
	mockDatastore := ds.NewMapDatastore()
	mockResourceManager := &MockResourceManagerService{}

	// Create bid manager
	bidManager := NewBidManager(
		config,
		mockBidMarket,
		mockLogger,
		mockMetrics,
		mockDatastore,
		mockResourceManager,
	)

	// Test data
	orderID := big.NewInt(123)
	pricePerSecond := big.NewInt(1000000000000000) // 0.001 ETH per second
	machineID := big.NewInt(456)
	providerID := config.ProviderID

	// Create mock transaction
	mockTx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil)

	// Create mock order and machine for resource allocation
	order := &Order{
		ID:           orderID,
		CpuCores:     big.NewInt(4),
		GpuCores:     big.NewInt(1),
		GpuMemory:    big.NewInt(8192),
		MemoryMB:     big.NewInt(16384),
		DiskGB:       big.NewInt(100),
		UploadMbps:   big.NewInt(100),
		DownloadMbps: big.NewInt(100),
	}

	machine := &Machine{
		ID: machineID,
	}

	// Setup mocks
	mockBidMarket.On("SubmitBid", mock.Anything, orderID, pricePerSecond, providerID, machineID).Return(mockTx, nil)
	mockBidMarket.On("GetBidIndexFromTransaction", mock.Anything, mockTx, orderID).Return(big.NewInt(1), nil)
	mockBidMarket.On("GetOrder", mock.Anything, orderID).Return(order, nil)
	mockResourceManager.On("GetMachine", mock.Anything, machineID).Return(machine, nil)
	mockResourceManager.On("AllocateResources", mock.Anything, orderID, machine, mock.AnythingOfType("*bidengine.ResourceUsage")).Return(nil)
	mockResourceManager.On("StartResource", mock.Anything, orderID, machine).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockMetrics.On("IncrementBidsSubmitted").Return()

	// Call SubmitBid
	result, err := bidManager.SubmitBid(context.Background(), orderID, pricePerSecond, machineID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, orderID, result.OrderID)
	assert.Equal(t, mockTx.Hash(), result.TxHash)

	// Verify that SubmitBid was called with the correct provider ID from config
	mockBidMarket.AssertCalled(t, "SubmitBid", mock.Anything, orderID, pricePerSecond, config.ProviderID, machineID)

	// Verify that the provider ID used was 42 (from config), not 1 (hardcoded)
	mockBidMarket.AssertNotCalled(t, "SubmitBid", mock.Anything, orderID, pricePerSecond, big.NewInt(1), machineID)

	mockBidMarket.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestSubmitBidReturnsErrorWhenProviderIDNotConfigured(t *testing.T) {
	// Create test configuration without provider ID
	config := &BidEngineConfig{
		BidMarketAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ProviderAddress:  common.HexToAddress("0x0987654321098765432109876543210987654321"),
		ProviderID:       nil, // No provider ID configured
		ProviderWallet:   common.HexToAddress("0x1111111111111111111111111111111111111111"),
		BidStrategy: BidStrategy{
			MinProfitMargin:   0.05,
			MaxProfitMargin:   0.20,
			CompetitiveFactor: 0.10,
			MarketAdjustment:  0.05,
			ResourceWeight: ResourceWeight{
				CPU:     0.25,
				GPU:     0.35,
				Memory:  0.20,
				Disk:    0.15,
				Network: 0.05,
			},
		},
		MaxConcurrentBids: 10,
		BidTimeout:        30 * time.Second,
		OrderSyncInterval: 30 * time.Second,
		BidCheckInterval:  60 * time.Second,
		LogLevel:          "INFO",
		LogFile:           "",
	}

	// Create mocks
	mockBidMarket := &MockBidMarketContract{}
	mockLogger := &MockLogger{}
	mockMetrics := &MockMetrics{}
	mockDatastore := ds.NewMapDatastore()
	mockResourceManager := &MockResourceManagerService{}

	// Create bid manager
	bidManager := NewBidManager(
		config,
		mockBidMarket,
		mockLogger,
		mockMetrics,
		mockDatastore,
		mockResourceManager,
	)

	// Test data
	orderID := big.NewInt(1)
	pricePerSecond := big.NewInt(1000000000000000000) // 1 ETH per second
	machineID := big.NewInt(123)

	// Call SubmitBid
	result, err := bidManager.SubmitBid(context.Background(), orderID, pricePerSecond, machineID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "provider ID not configured")

	// Verify that SubmitBid was not called
	mockBidMarket.AssertNotCalled(t, "SubmitBid", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestCheckOrderExpiry(t *testing.T) {
	// Setup
	config := &BidEngineConfig{
		ProviderID: big.NewInt(1),
	}
	mockBidMarket := &MockBidMarketContract{}
	mockLogger := &MockLogger{}
	mockMetrics := &MockMetrics{}
	mockDatastore := ds.NewMapDatastore()
	mockResourceManager := &MockResourceManagerService{}

	// Create bid manager
	bidManager := NewBidManager(
		config,
		mockBidMarket,
		mockLogger,
		mockMetrics,
		mockDatastore,
		mockResourceManager,
	)

	// Test data
	orderID := big.NewInt(123)
	expiredTime := big.NewInt(time.Now().Unix() - 3600) // 1 hour ago
	order := &Order{
		ID:        orderID,
		ExpiredAt: expiredTime,
	}

	// Setup mocks
	mockBidMarket.On("GetOrder", mock.Anything, orderID).Return(order, nil)
	mockResourceManager.On("StopResource", mock.Anything, orderID).Return(nil)
	mockResourceManager.On("DeallocateResources", mock.Anything, orderID).Return(nil)
	mockBidMarket.On("CloseOrder", mock.Anything, orderID, "Order expired").Return(types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil), nil)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// Execute
	err := bidManager.CheckOrderExpiry(context.Background(), orderID)

	// Assert
	assert.NoError(t, err)
	mockBidMarket.AssertExpectations(t)
	mockResourceManager.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAllocateResourcesForBid(t *testing.T) {
	// Setup
	config := &BidEngineConfig{
		ProviderID: big.NewInt(1),
	}
	mockBidMarket := &MockBidMarketContract{}
	mockLogger := &MockLogger{}
	mockMetrics := &MockMetrics{}
	mockDatastore := ds.NewMapDatastore()
	mockResourceManager := &MockResourceManagerService{}

	// Create bid manager
	bidManager := NewBidManager(
		config,
		mockBidMarket,
		mockLogger,
		mockMetrics,
		mockDatastore,
		mockResourceManager,
	)

	// Test data
	orderID := big.NewInt(123)
	machineID := big.NewInt(456)

	order := &Order{
		ID:           orderID,
		CpuCores:     big.NewInt(4),
		GpuCores:     big.NewInt(1),
		GpuMemory:    big.NewInt(8192),
		MemoryMB:     big.NewInt(16384),
		DiskGB:       big.NewInt(100),
		UploadMbps:   big.NewInt(100),
		DownloadMbps: big.NewInt(100),
	}

	machine := &Machine{
		ID: machineID,
	}

	// Setup mocks
	mockBidMarket.On("GetOrder", mock.Anything, orderID).Return(order, nil)
	mockResourceManager.On("GetMachine", mock.Anything, machineID).Return(machine, nil)
	mockResourceManager.On("AllocateResources", mock.Anything, orderID, machine, mock.AnythingOfType("*bidengine.ResourceUsage")).Return(nil)
	mockResourceManager.On("StartResource", mock.Anything, orderID, machine).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// Execute
	err := bidManager.allocateResourcesForBid(context.Background(), orderID, machineID)

	// Assert
	assert.NoError(t, err)
	mockBidMarket.AssertExpectations(t)
	mockResourceManager.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestDeallocateResourcesForBid(t *testing.T) {
	// Setup
	config := &BidEngineConfig{
		ProviderID: big.NewInt(1),
	}
	mockBidMarket := &MockBidMarketContract{}
	mockLogger := &MockLogger{}
	mockMetrics := &MockMetrics{}
	mockDatastore := ds.NewMapDatastore()
	mockResourceManager := &MockResourceManagerService{}

	// Create bid manager
	bidManager := NewBidManager(
		config,
		mockBidMarket,
		mockLogger,
		mockMetrics,
		mockDatastore,
		mockResourceManager,
	)

	// Test data
	orderID := big.NewInt(123)

	// Setup mocks
	mockResourceManager.On("StopResource", mock.Anything, orderID).Return(nil)
	mockResourceManager.On("DeallocateResources", mock.Anything, orderID).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// Execute
	err := bidManager.deallocateResourcesForBid(context.Background(), orderID)

	// Assert
	assert.NoError(t, err)
	mockResourceManager.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
