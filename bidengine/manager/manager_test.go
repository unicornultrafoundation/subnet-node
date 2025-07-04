package manager

import (
	"context"
	"math/big"
	"testing"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// --- Mocks ---
type mockBidMarket struct{ types.BidMarketContract }

func (m *mockBidMarket) GetOrder(ctx context.Context, orderID *big.Int) (*types.Order, error) {
	return &types.Order{
		ID:          orderID,
		Status:      types.OrderStatusOpen,
		MachineType: big.NewInt(1),
		MaxBidPrice: big.NewInt(200),
		CpuCores:    big.NewInt(4),
		GpuCores:    big.NewInt(1),
		MemoryMB:    big.NewInt(8192),
		DiskGB:      big.NewInt(100),
		UploadMbps:  big.NewInt(100),
	}, nil
}

func (m *mockBidMarket) SubmitBid(ctx context.Context, orderID *big.Int, pricePerSecond *big.Int, providerID *big.Int, machineID *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *mockBidMarket) GetBidIndexFromTransaction(ctx context.Context, tx *ethtypes.Transaction, orderID *big.Int) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *mockBidMarket) CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *mockBidMarket) IsBiddingOpen(ctx context.Context, orderID *big.Int) (bool, error) {
	return true, nil
}

type mockResourceManager struct{ types.ResourceManager }

func (m *mockResourceManager) GetAllMachines(ctx context.Context) []*types.Machine {
	return []*types.Machine{
		{
			ID:          big.NewInt(1),
			Active:      true,
			MachineType: big.NewInt(1),
		},
	}
}

func (m *mockResourceManager) GetMachine(ctx context.Context, machineID *big.Int) (*types.Machine, error) {
	return &types.Machine{ID: machineID}, nil
}

func (m *mockResourceManager) CanAllocateResources(ctx context.Context, machine *types.Machine, required *types.ResourceUsage) (bool, error) {
	return true, nil
}

func (m *mockResourceManager) AllocateResources(ctx context.Context, orderID *big.Int, machine *types.Machine, usage *types.ResourceUsage) error {
	return nil
}

func (m *mockResourceManager) StartResource(ctx context.Context, orderID *big.Int, machine *types.Machine) error {
	return nil
}

func (m *mockResourceManager) StopResource(ctx context.Context, orderID *big.Int) error {
	return nil
}

func (m *mockResourceManager) DeallocateResources(ctx context.Context, orderID *big.Int) error {
	return nil
}

func (m *mockResourceManager) RegisterMachine(ctx context.Context, machine *types.Machine) error {
	return nil
}

func (m *mockResourceManager) UnregisterMachine(ctx context.Context, machineID *big.Int) error {
	return nil
}

func (m *mockResourceManager) GetCurrentUsage(ctx context.Context, machine *types.Machine) (*types.ResourceUsage, error) {
	return &types.ResourceUsage{}, nil
}

func (m *mockResourceManager) GetAvailableResources(ctx context.Context, machine *types.Machine) (*types.ResourceUsage, error) {
	return &types.ResourceUsage{}, nil
}

type mockPricingEngine struct{ types.PricingEngine }

func (m *mockPricingEngine) CalculateBidPrice(ctx context.Context, order *types.Order, machine *types.Machine, marketData *types.MarketData) (*big.Int, error) {
	return big.NewInt(100), nil
}

func (m *mockPricingEngine) AnalyzeMarket(ctx context.Context, orders []*types.Order) (*types.MarketData, error) {
	return &types.MarketData{}, nil
}

func (m *mockPricingEngine) CalculateResourcePrice(ctx context.Context, machine *types.Machine, usage *types.ResourceUsage) (*big.Int, error) {
	return big.NewInt(100), nil
}

func (m *mockPricingEngine) AdjustPriceForStrategy(ctx context.Context, basePrice *big.Int, strategy *types.BidStrategy) (*big.Int, error) {
	return basePrice, nil
}

type mockOrderMonitor struct{ types.OrderMonitor }

func (m *mockOrderMonitor) RegisterEventHandler(eventType types.OrderEventType, handler func(*types.OrderEvent)) {
}

func (m *mockOrderMonitor) TrackOrder(ctx context.Context, orderID *big.Int) error {
	return nil
}

func (m *mockOrderMonitor) UntrackOrder(ctx context.Context, orderID *big.Int) error {
	return nil
}

func (m *mockOrderMonitor) GetTrackedOrders(ctx context.Context) ([]*big.Int, error) {
	return []*big.Int{}, nil
}

type mockStorage struct{ types.Storage }

func (m *mockStorage) SaveBid(ctx context.Context, bid *types.Bid, orderID string, bidIndex int) error {
	return nil
}

func (m *mockStorage) UpdateBid(ctx context.Context, bid *types.Bid, orderID string, bidIndex int) error {
	return nil
}

func (m *mockStorage) DeleteBid(ctx context.Context, orderID string, bidIndex int) error {
	return nil
}

func (m *mockStorage) ListOrders(ctx context.Context) ([]*types.Order, error) {
	return []*types.Order{}, nil
}

func (m *mockStorage) GetBids(ctx context.Context, orderID string) ([]*types.Bid, error) {
	return []*types.Bid{}, nil
}

func (m *mockStorage) SaveOrder(ctx context.Context, order *types.Order) error {
	return nil
}

func (m *mockStorage) GetOrder(ctx context.Context, orderID string) (*types.Order, error) {
	return &types.Order{}, nil
}

func (m *mockStorage) UpdateOrder(ctx context.Context, order *types.Order) error {
	return nil
}

func (m *mockStorage) DeleteOrder(ctx context.Context, orderID string) error {
	return nil
}

func (m *mockStorage) SaveMachine(ctx context.Context, machine *types.Machine) error {
	return nil
}

func (m *mockStorage) GetMachine(ctx context.Context, machineID string) (*types.Machine, error) {
	return &types.Machine{}, nil
}

func (m *mockStorage) ListMachines(ctx context.Context) ([]*types.Machine, error) {
	return []*types.Machine{}, nil
}

func (m *mockStorage) SaveMarketData(ctx context.Context, marketData *types.MarketData) error {
	return nil
}

func (m *mockStorage) GetMarketData(ctx context.Context) (*types.MarketData, error) {
	return &types.MarketData{}, nil
}

func (m *mockStorage) SaveResourceAllocation(ctx context.Context, allocation *types.ResourceAllocation) error {
	return nil
}

func (m *mockStorage) ListResourceAllocations(ctx context.Context) ([]*types.ResourceAllocation, error) {
	return []*types.ResourceAllocation{}, nil
}

func (m *mockStorage) SaveLastOrderID(ctx context.Context, orderID *big.Int) error {
	return nil
}

func (m *mockStorage) GetLastOrderID(ctx context.Context) (*big.Int, error) {
	return nil, nil
}

type mockMetrics struct{ types.Metrics }

func (m *mockMetrics) IncrementBidsSubmitted() {}
func (m *mockMetrics) IncrementBidsAccepted()  {}
func (m *mockMetrics) IncrementBidsRejected()  {}

func (m *mockMetrics) RecordBidLatency(duration float64)                    {}
func (m *mockMetrics) IncrementOrdersTracked()                              {}
func (m *mockMetrics) IncrementOrdersCompleted()                            {}
func (m *mockMetrics) RecordOrderDuration(duration float64)                 {}
func (m *mockMetrics) RecordResourceUtilization(usage *types.ResourceUsage) {}
func (m *mockMetrics) RecordResourceAllocation(usage *types.ResourceUsage)  {}
func (m *mockMetrics) RecordRevenue(amount *big.Int)                        {}
func (m *mockMetrics) RecordProfit(amount *big.Int)                         {}
func (m *mockMetrics) RecordCost(amount *big.Int)                           {}

// Local struct for bidMarket mock for TestSubmitBid
type bidMarketMock struct{ mockBidMarket }

func (b *bidMarketMock) SubmitBid(ctx context.Context, orderID, pricePerSecond, providerID, machineID *big.Int) (*ethtypes.Transaction, error) {
	tx := ethtypes.NewTx(&ethtypes.LegacyTx{})
	return tx, nil
}
func (b *bidMarketMock) GetBidIndexFromTransaction(ctx context.Context, tx *ethtypes.Transaction, orderID *big.Int) (*big.Int, error) {
	return big.NewInt(0), nil
}

func TestTrackAndUntrackBid(t *testing.T) {
	logger := logrus.WithField("service", "bidengine")
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		nil, // datastore
		&mockResourceManager{},
		&mockPricingEngine{},
		&mockOrderMonitor{},
	)
	manager.storage = &mockStorage{}

	ctx := context.Background()
	orderID := big.NewInt(123)
	bidIndex := big.NewInt(0)

	bid := &types.Bid{
		Id:         bidIndex,
		ProviderId: big.NewInt(1),
		Status:     types.BidStatusActive,
	}
	err := manager.TrackBid(ctx, orderID, bid)
	assert.NoError(t, err)

	// Track lại bid đã tồn tại
	err = manager.TrackBid(ctx, orderID, bid)
	assert.Error(t, err)

	// Untrack bid
	err = manager.UntrackBid(ctx, orderID, bid)
	assert.NoError(t, err)

	// Untrack bid không tồn tại
	err = manager.UntrackBid(ctx, orderID, bid)
	assert.Error(t, err)
}

func TestGetTrackedBids(t *testing.T) {
	logger := logrus.WithField("service", "bidengine")
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		nil, // datastore
		&mockResourceManager{},
		&mockPricingEngine{},
		&mockOrderMonitor{},
	)
	manager.storage = &mockStorage{}

	ctx := context.Background()
	orderID := big.NewInt(456)
	bidIndex := big.NewInt(0)

	// Track the bid
	bid := &types.Bid{
		Id:         bidIndex,
		ProviderId: big.NewInt(1),
		Status:     types.BidStatusActive,
	}
	err := manager.TrackBid(ctx, orderID, bid)
	assert.NoError(t, err)

	// Debug: check if bid was actually tracked
	assert.Equal(t, 1, len(manager.pendingBids))

	// Get tracked bids
	bids, err := manager.GetTrackedBids(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(bids))

	// Debug: print the actual bids map
	t.Logf("Bids map: %+v", bids)
	t.Logf("OrderID: %s", orderID.String())

	// Find the correct key in the map (handle potential plus sign)
	var foundOrderID *big.Int
	var foundIds []*big.Int
	for key, ids := range bids {
		if key.Cmp(orderID) == 0 {
			foundOrderID = key
			foundIds = ids
			break
		}
	}

	assert.NotNil(t, foundOrderID, "OrderID should exist in bids map")
	assert.Greater(t, len(foundIds), 0, "Bid slice should not be empty")
	assert.Equal(t, bidIndex.String(), foundIds[0].String(), "Bid index should match")
}

func TestTryBidOnOrder(t *testing.T) {
	logger := logrus.WithField("service", "bidengine")
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		nil, // datastore
		&mockResourceManager{},
		&mockPricingEngine{},
		&mockOrderMonitor{},
	)
	manager.storage = &mockStorage{}

	ctx := context.Background()
	orderID := big.NewInt(123)

	// Use bidMarketMock with valid SubmitBid and GetBidIndexFromTransaction
	manager.bidMarket = &bidMarketMock{}

	order := &types.Order{
		ID:          orderID,
		Status:      types.OrderStatusOpen,
		MachineType: big.NewInt(1),
		MaxBidPrice: big.NewInt(200),
		CpuCores:    big.NewInt(4),
		GpuCores:    big.NewInt(1),
		MemoryMB:    big.NewInt(8192),
		DiskGB:      big.NewInt(100),
		UploadMbps:  big.NewInt(100),
	}

	err := manager.TryBidOnOrder(ctx, order)
	assert.NoError(t, err)
}

func TestHandleOrderCreate(t *testing.T) {
	logger := logrus.WithField("service", "bidengine")
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		nil, // datastore
		&mockResourceManager{},
		&mockPricingEngine{},
		&mockOrderMonitor{},
	)
	manager.storage = &mockStorage{}

	// Use bidMarketMock with valid SubmitBid and GetBidIndexFromTransaction
	manager.bidMarket = &bidMarketMock{}

	orderID := big.NewInt(123)
	order := &types.Order{
		ID:          orderID,
		Status:      types.OrderStatusOpen,
		MachineType: big.NewInt(1),
		MaxBidPrice: big.NewInt(200),
	}

	event := &types.OrderEvent{
		Type:    types.OrderEventNew,
		OrderID: orderID,
		Order:   order,
	}

	manager.handleOrderCreate(event)
}

func TestHandleOrderClosed(t *testing.T) {
	logger := logrus.WithField("service", "bidengine")
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		nil, // datastore
		&mockResourceManager{},
		&mockPricingEngine{},
		&mockOrderMonitor{},
	)
	manager.storage = &mockStorage{}

	orderID := big.NewInt(123)
	order := &types.Order{ID: orderID}

	event := &types.OrderEvent{
		Type:    types.OrderEventClosed,
		OrderID: orderID,
		Order:   order,
	}

	// Add a pending bid first
	manager.pendingBids[orderID.String()] = &ExtendedBid{
		Bid: &types.Bid{
			Id: big.NewInt(0),
		},
		OrderID: orderID.String(),
	}

	manager.handleOrderClosed(event)
}

func TestHandleOrderExpired(t *testing.T) {
	logger := logrus.WithField("service", "bidengine")
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		nil, // datastore
		&mockResourceManager{},
		&mockPricingEngine{},
		&mockOrderMonitor{},
	)
	manager.storage = &mockStorage{}

	orderID := big.NewInt(123)
	order := &types.Order{ID: orderID}

	event := &types.OrderEvent{
		Type:    types.OrderEventExpired,
		OrderID: orderID,
		Order:   order,
	}

	// Add a pending bid first
	manager.pendingBids[orderID.String()] = &ExtendedBid{
		Bid: &types.Bid{
			Id: big.NewInt(0),
		},
		OrderID: orderID.String(),
	}

	manager.handleOrderExpired(event)
}

func TestHandleOrderAccepted(t *testing.T) {
	logger := logrus.WithField("service", "bidengine")
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		nil, // datastore
		&mockResourceManager{},
		&mockPricingEngine{},
		&mockOrderMonitor{},
	)
	manager.storage = &mockStorage{}

	orderID := big.NewInt(123)
	providerID := big.NewInt(1)
	order := &types.Order{
		ID:                 orderID,
		AcceptedProviderId: providerID,
	}

	event := &types.OrderEvent{
		Type:    types.OrderEventAccepted,
		OrderID: orderID,
		Order:   order,
	}

	// Add a pending bid with matching provider ID
	manager.pendingBids[orderID.String()] = &ExtendedBid{
		Bid: &types.Bid{
			Id:         big.NewInt(0),
			ProviderId: providerID,
		},
		OrderID: orderID.String(),
	}

	manager.handleOrderAccepted(event)
}

func TestGetStats(t *testing.T) {
	logger := logrus.WithField("service", "bidengine")
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		nil, // datastore
		&mockResourceManager{},
		&mockPricingEngine{},
		&mockOrderMonitor{},
	)

	// Add some test data
	manager.pendingBids["123"] = &ExtendedBid{Bid: &types.Bid{Id: big.NewInt(0)}}
	manager.bidResults["123"] = &types.BidResult{OrderID: big.NewInt(123)}

	stats := manager.GetStats()
	assert.Equal(t, 1, stats["pendingBids"])
	assert.Equal(t, 1, stats["bidResults"])
	assert.False(t, stats["isRunning"].(bool))
}
