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

// MockPricingEngine is a mock implementation of PricingEngine
type MockPricingEngine struct {
	mock.Mock
}

func (m *MockPricingEngine) CalculateBidPrice(ctx context.Context, order *Order, machine *Machine, marketData *MarketData) (*big.Int, error) {
	args := m.Called(ctx, order, machine, marketData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockPricingEngine) AnalyzeMarket(ctx context.Context, orders []*Order) (*MarketData, error) {
	args := m.Called(ctx, orders)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MarketData), args.Error(1)
}

func (m *MockPricingEngine) CalculateResourcePrice(ctx context.Context, machine *Machine, usage *ResourceUsage) (*big.Int, error) {
	args := m.Called(ctx, machine, usage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockPricingEngine) AdjustPriceForStrategy(ctx context.Context, basePrice *big.Int, strategy *BidStrategy) (*big.Int, error) {
	args := m.Called(ctx, basePrice, strategy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func createTestBidManager(config *BidEngineConfig) (*BidManagerService, *MockBidMarketContract, *MockLogger, *MockMetrics, *MockResourceManagerService, *MockPricingEngine) {
	mockBidMarket := &MockBidMarketContract{}
	mockLogger := &MockLogger{}
	mockMetrics := &MockMetrics{}
	mockDatastore := ds.NewMapDatastore()
	mockResourceManager := &MockResourceManagerService{}
	mockPricingEngine := &MockPricingEngine{}

	bidManager := NewBidManager(
		config,
		mockBidMarket,
		mockLogger,
		mockMetrics,
		mockDatastore,
		mockResourceManager,
		mockPricingEngine,
	)

	return bidManager, mockBidMarket, mockLogger, mockMetrics, mockResourceManager, mockPricingEngine
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
	testManager, testBidMarket, testLogger, testMetrics, testResourceManager, _ := createTestBidManager(config)

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
	testBidMarket.On("SubmitBid", mock.Anything, orderID, pricePerSecond, providerID, machineID).Return(mockTx, nil)
	testBidMarket.On("GetBidIndexFromTransaction", mock.Anything, mockTx, orderID).Return(big.NewInt(1), nil)
	testBidMarket.On("GetOrder", mock.Anything, orderID).Return(order, nil)
	testResourceManager.On("GetMachine", mock.Anything, machineID).Return(machine, nil)
	testResourceManager.On("AllocateResources", mock.Anything, orderID, machine, mock.AnythingOfType("*bidengine.ResourceUsage")).Return(nil)
	testResourceManager.On("StartResource", mock.Anything, orderID, machine).Return(nil)
	testLogger.On("Info", mock.Anything, mock.Anything).Return()
	testLogger.On("Warn", mock.Anything, mock.Anything).Return()
	testLogger.On("Debug", mock.Anything, mock.Anything).Return()
	testMetrics.On("IncrementBidsSubmitted").Return()

	// Call SubmitBid
	result, err := testManager.SubmitBid(context.Background(), orderID, pricePerSecond, machineID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, orderID, result.OrderID)
	assert.Equal(t, mockTx.Hash(), result.TxHash)

	// Verify that SubmitBid was called with the correct provider ID from config
	testBidMarket.AssertCalled(t, "SubmitBid", mock.Anything, orderID, pricePerSecond, config.ProviderID, machineID)

	// Verify that the provider ID used was 42 (from config), not 1 (hardcoded)
	testBidMarket.AssertNotCalled(t, "SubmitBid", mock.Anything, orderID, pricePerSecond, big.NewInt(1), machineID)

	testBidMarket.AssertExpectations(t)
	testLogger.AssertExpectations(t)
	testMetrics.AssertExpectations(t)
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
	bidManager, mockBidMarket, _, _, _, _ := createTestBidManager(config)

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
	bidManager, mockBidMarket, mockLogger, _, mockResourceManager, _ := createTestBidManager(config)

	// Test data - order expired more than 1 day ago (should close)
	orderID := big.NewInt(123)
	expiredTime := big.NewInt(time.Now().Unix() - 90000) // More than 1 day ago (86400 + 3600 seconds)
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

func TestCheckOrderExpiryNotYetReady(t *testing.T) {
	// Setup
	config := &BidEngineConfig{
		ProviderID: big.NewInt(1),
	}
	bidManager, mockBidMarket, mockLogger, _, _, _ := createTestBidManager(config)

	// Test data - order expired but less than 1 day ago (should not close)
	orderID := big.NewInt(123)
	expiredTime := big.NewInt(time.Now().Unix() - 3600) // 1 hour ago (less than 1 day)
	order := &Order{
		ID:        orderID,
		ExpiredAt: expiredTime,
	}

	// Setup mocks
	mockBidMarket.On("GetOrder", mock.Anything, orderID).Return(order, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	// Execute
	err := bidManager.CheckOrderExpiry(context.Background(), orderID)

	// Assert
	assert.NoError(t, err)
	mockBidMarket.AssertExpectations(t)
	mockLogger.AssertExpectations(t)

	// Verify that CloseOrder was not called
	mockBidMarket.AssertNotCalled(t, "CloseOrder", mock.Anything, mock.Anything, mock.Anything)
}

func TestAllocateResourcesForBid(t *testing.T) {
	// Setup
	config := &BidEngineConfig{
		ProviderID: big.NewInt(1),
	}
	bidManager, mockBidMarket, mockLogger, _, mockResourceManager, _ := createTestBidManager(config)

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
	bidManager, _, mockLogger, _, mockResourceManager, _ := createTestBidManager(config)

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

func TestCancelBid(t *testing.T) {
	// Setup
	config := &BidEngineConfig{
		ProviderID: big.NewInt(1),
	}
	bidManager, mockBidMarket, mockLogger, _, mockResourceManager, _ := createTestBidManager(config)

	// Test data
	orderID := big.NewInt(123)
	bidIndex := big.NewInt(0)
	machineID := big.NewInt(456)

	// Setup a pending bid first
	bidManager.mu.Lock()
	bidManager.pendingBids[orderID.String()] = &ExtendedBid{
		Bid: &Bid{
			Id:        bidIndex,
			MachineId: machineID,
			Status:    BidStatusActive,
		},
		OrderID:     orderID.String(),
		SubmittedAt: time.Now(),
	}
	bidManager.mu.Unlock()

	// Setup mocks
	mockBidMarket.On("CancelBid", mock.Anything, orderID, bidIndex).Return(
		types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil), nil)
	mockResourceManager.On("StopResource", mock.Anything, orderID).Return(nil)
	mockResourceManager.On("DeallocateResources", mock.Anything, orderID).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	// Execute
	err := bidManager.CancelBid(context.Background(), orderID, bidIndex)

	// Assert
	assert.NoError(t, err)
	mockBidMarket.AssertExpectations(t)
	mockResourceManager.AssertExpectations(t)
	mockLogger.AssertExpectations(t)

	// Verify bid is removed from pending bids
	bidManager.mu.RLock()
	_, exists := bidManager.pendingBids[orderID.String()]
	bidManager.mu.RUnlock()
	assert.False(t, exists, "Bid should be removed from pending bids")
}

func TestCancelBidWithResourceDeallocationError(t *testing.T) {
	// Setup
	config := &BidEngineConfig{
		ProviderID: big.NewInt(1),
	}
	bidManager, mockBidMarket, mockLogger, _, mockResourceManager, _ := createTestBidManager(config)

	// Test data
	orderID := big.NewInt(123)
	bidIndex := big.NewInt(0)
	machineID := big.NewInt(456)

	// Setup a pending bid first
	bidManager.mu.Lock()
	bidManager.pendingBids[orderID.String()] = &ExtendedBid{
		Bid: &Bid{
			Id:        bidIndex,
			MachineId: machineID,
			Status:    BidStatusActive,
		},
		OrderID:     orderID.String(),
		SubmittedAt: time.Now(),
	}
	bidManager.mu.Unlock()

	// Setup mocks - resource deallocation fails
	mockBidMarket.On("CancelBid", mock.Anything, orderID, bidIndex).Return(
		types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil), nil)
	mockResourceManager.On("StopResource", mock.Anything, orderID).Return(nil)
	mockResourceManager.On("DeallocateResources", mock.Anything, orderID).Return(
		assert.AnError)
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	// Execute
	err := bidManager.CancelBid(context.Background(), orderID, bidIndex)

	// Assert - should still succeed even if resource deallocation fails
	assert.NoError(t, err)
	mockBidMarket.AssertExpectations(t)
	mockResourceManager.AssertExpectations(t)
	mockLogger.AssertExpectations(t)

	// Verify bid is still removed from pending bids
	bidManager.mu.RLock()
	_, exists := bidManager.pendingBids[orderID.String()]
	bidManager.mu.RUnlock()
	assert.False(t, exists, "Bid should be removed from pending bids even if resource deallocation fails")
}
