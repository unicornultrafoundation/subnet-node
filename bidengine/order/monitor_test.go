package order

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// --- Mocks ---
type mockBidMarket struct{ types.BidMarketContract }

func (m *mockBidMarket) GetOrder(ctx context.Context, orderID *big.Int) (*types.Order, error) {
	return &types.Order{
		ID:        orderID,
		Status:    types.OrderStatusOpen,
		CreatedAt: big.NewInt(time.Now().Unix()),
	}, nil
}
func (m *mockBidMarket) WatchOrderCreated(ctx context.Context, sink chan<- *types.OrderEvent) error {
	return nil
}

// Minimal mock storage
type mockStorage struct{ types.Storage }

func (m *mockStorage) SaveOrder(ctx context.Context, order *types.Order) error { return nil }
func (m *mockStorage) DeleteOrder(ctx context.Context, orderID string) error   { return nil }
func (m *mockStorage) ListOrders(ctx context.Context) ([]*types.Order, error) {
	return []*types.Order{}, nil
}
func (m *mockStorage) UpdateOrder(ctx context.Context, order *types.Order) error { return nil }

// Minimal mock metrics
type mockMetrics struct{ types.Metrics }

func (m *mockMetrics) IncrementOrdersTracked() {}

func TestTrackAndUntrackOrder(t *testing.T) {
	logger := logrus.New()
	monitor := NewMonitor(
		&types.BidEngineConfig{},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		&mockStorage{},
	)
	ctx := context.Background()
	orderID := big.NewInt(123)

	err := monitor.TrackOrder(ctx, orderID)
	assert.NoError(t, err)

	// Track existing order again
	err = monitor.TrackOrder(ctx, orderID)
	assert.Error(t, err)

	// Untrack order
	err = monitor.UntrackOrder(ctx, orderID)
	assert.NoError(t, err)

	// Untrack non-existent order
	err = monitor.UntrackOrder(ctx, orderID)
	assert.Error(t, err)
}

func TestGetTrackedOrders(t *testing.T) {
	logger := logrus.New()
	monitor := NewMonitor(
		&types.BidEngineConfig{},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		&mockStorage{},
	)
	ctx := context.Background()
	orderID := big.NewInt(456)
	_ = monitor.TrackOrder(ctx, orderID)

	ids, err := monitor.GetTrackedOrders(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ids))
	assert.Equal(t, orderID.String(), ids[0].String())
}

func TestMonitorOrderStatus(t *testing.T) {
	logger := logrus.New()
	monitor := NewMonitor(
		&types.BidEngineConfig{},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		&mockStorage{},
	)
	ctx := context.Background()
	orderID := big.NewInt(789)
	_ = monitor.TrackOrder(ctx, orderID)

	err := monitor.MonitorOrderStatus(ctx, orderID)
	assert.NoError(t, err)

	// Monitor non-existent order
	err = monitor.MonitorOrderStatus(ctx, big.NewInt(999))
	assert.Error(t, err)
}

func TestRegisterEventHandlerAndEmit(t *testing.T) {
	logger := logrus.New()
	monitor := NewMonitor(
		&types.BidEngineConfig{},
		&mockBidMarket{},
		logger,
		&mockMetrics{},
		&mockStorage{},
	)
	called := false
	handler := func(ev *types.OrderEvent) {
		called = true
	}
	monitor.RegisterEventHandler(types.OrderEventNew, handler)

	event := &types.OrderEvent{
		Type:    types.OrderEventNew,
		OrderID: big.NewInt(1),
		Order:   &types.Order{ID: big.NewInt(1)},
	}
	monitor.emitEvent(event)
	time.Sleep(100 * time.Millisecond) // Wait for goroutine
	assert.True(t, called)
}
