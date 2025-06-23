package storage

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jbenet/goprocess"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// MockDatastore is a mock implementation of ds.Datastore
type MockDatastore struct {
	mock.Mock
	data map[string][]byte
}

func NewMockDatastore() *MockDatastore {
	return &MockDatastore{
		data: make(map[string][]byte),
	}
}

func (m *MockDatastore) Put(ctx context.Context, key ds.Key, value []byte) error {
	args := m.Called(ctx, key, value)
	m.data[key.String()] = value
	return args.Error(0)
}

func (m *MockDatastore) Get(ctx context.Context, key ds.Key) ([]byte, error) {
	if value, exists := m.data[key.String()]; exists {
		return value, nil
	}
	return nil, ds.ErrNotFound
}

func (m *MockDatastore) Has(ctx context.Context, key ds.Key) (exists bool, err error) {
	args := m.Called(ctx, key)
	_, exists = m.data[key.String()]
	return exists, args.Error(0)
}

func (m *MockDatastore) Delete(ctx context.Context, key ds.Key) error {
	args := m.Called(ctx, key)
	delete(m.data, key.String())
	return args.Error(0)
}

func (m *MockDatastore) Query(ctx context.Context, q query.Query) (query.Results, error) {
	args := m.Called(ctx, q)

	// Simple mock implementation for query
	var results []query.Entry
	for key, value := range m.data {
		if q.Prefix != "" && !startsWith(key, q.Prefix) {
			continue
		}
		results = append(results, query.Entry{
			Key:   key,
			Value: value,
		})
	}

	return &MockQueryResults{results: results, index: 0, query: q}, args.Error(0)
}

func (m *MockDatastore) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatastore) Sync(ctx context.Context, prefix ds.Key) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func (m *MockDatastore) GetSize(ctx context.Context, key ds.Key) (size int, err error) {
	args := m.Called(ctx, key)
	if value, exists := m.data[key.String()]; exists {
		return len(value), args.Error(0)
	}
	return 0, ds.ErrNotFound
}

// MockQueryResults implements query.Results
type MockQueryResults struct {
	results []query.Entry
	index   int
	query   query.Query
}

func (m *MockQueryResults) Query() query.Query {
	return m.query
}

func (m *MockQueryResults) Next() <-chan query.Result {
	ch := make(chan query.Result)
	go func() {
		defer close(ch)
		for i, entry := range m.results {
			if i >= m.index {
				ch <- query.Result{Entry: entry}
			}
		}
	}()
	return ch
}

func (m *MockQueryResults) NextSync() (query.Result, bool) {
	if m.index >= len(m.results) {
		return query.Result{}, false
	}
	result := query.Result{Entry: m.results[m.index]}
	m.index++
	return result, true
}

func (m *MockQueryResults) Close() error {
	return nil
}

func (m *MockQueryResults) Rest() ([]query.Entry, error) {
	var entries []query.Entry
	for i := m.index; i < len(m.results); i++ {
		entries = append(entries, m.results[i])
	}
	return entries, nil
}

func (m *MockQueryResults) Process() goprocess.Process {
	return goprocess.WithParent(nil)
}

// Helper function
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func testStorage() (*Storage, *MockDatastore) {
	mockDS := NewMockDatastore()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	storage := NewStorage(mockDS, logger)
	return storage, mockDS
}

func testOrder() *Order {
	return &Order{
		ID:          big.NewInt(1),
		Status:      types.OrderStatusOpen,
		CpuCores:    big.NewInt(2),
		GpuCores:    big.NewInt(1),
		MemoryMB:    big.NewInt(1024),
		DiskGB:      big.NewInt(100),
		UploadMbps:  big.NewInt(10),
		MinBidPrice: big.NewInt(100),
		MaxBidPrice: big.NewInt(1000),
		CreatedAt:   big.NewInt(time.Now().Unix()),
		ExpiredAt:   big.NewInt(time.Now().Add(24 * time.Hour).Unix()),
	}
}

func testBid() *Bid {
	return &Bid{
		Id:             big.NewInt(0),
		Provider:       common.HexToAddress("0x1234567890123456789012345678901234567890"),
		PricePerSecond: big.NewInt(150),
		Status:         types.BidStatusActive,
		CreatedAt:      big.NewInt(time.Now().Unix()),
		ProviderId:     big.NewInt(1),
		MachineId:      big.NewInt(1),
	}
}

func testMachine() *Machine {
	return &Machine{
		ID:                   big.NewInt(1),
		MachineType:          big.NewInt(1),
		Region:               big.NewInt(1),
		Active:               true,
		CpuCores:             big.NewInt(4),
		GpuCores:             big.NewInt(2),
		MemoryMB:             big.NewInt(2048),
		DiskGB:               big.NewInt(200),
		CpuPricePerSecond:    big.NewInt(10),
		GpuPricePerSecond:    big.NewInt(20),
		MemoryPricePerSecond: big.NewInt(5),
		DiskPricePerSecond:   big.NewInt(2),
	}
}

func testMarketData() *MarketData {
	return &MarketData{
		AveragePricePerSecond: big.NewInt(150),
		MinPricePerSecond:     big.NewInt(100),
		MaxPricePerSecond:     big.NewInt(200),
		TotalOrders:           big.NewInt(10),
		ActiveOrders:          big.NewInt(5),
		LastUpdated:           time.Now(),
	}
}

func testResourceAllocation() *ResourceAllocation {
	return &ResourceAllocation{
		OrderID: big.NewInt(1),
		Usage: &types.ResourceUsage{
			CPUUsed:    big.NewInt(2),
			GPUUsed:    big.NewInt(1),
			MemoryUsed: big.NewInt(1024),
			DiskUsed:   big.NewInt(100),
		},
	}
}

func TestSaveOrder(t *testing.T) {
	storage, mockDS := testStorage()
	order := testOrder()

	mockDS.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := storage.SaveOrder(context.Background(), order)
	assert.NoError(t, err)
	mockDS.AssertExpectations(t)
}

func TestGetOrder(t *testing.T) {
	storage, mockDS := testStorage()
	order := testOrder()

	// Save order first
	orderStorage := &OrderStorage{
		Order:      order,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastSynced: time.Now(),
		BidCount:   0,
		Status:     order.Status,
	}
	data, _ := json.Marshal(orderStorage)
	mockDS.data["/bidengine/orders/1"] = data

	retrievedOrder, err := storage.GetOrder(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, order.ID.String(), retrievedOrder.ID.String())
}

func TestGetOrderNotFound(t *testing.T) {
	storage, _ := testStorage()

	_, err := storage.GetOrder(context.Background(), "999")
	assert.Error(t, err)
	assert.Equal(t, ErrOrderNotFound, err)
}

func TestListOrders(t *testing.T) {
	storage, mockDS := testStorage()
	order := testOrder()

	// Save order
	orderStorage := &OrderStorage{
		Order:      order,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastSynced: time.Now(),
		BidCount:   0,
		Status:     order.Status,
	}
	data, _ := json.Marshal(orderStorage)
	mockDS.data["/bidengine/orders/1"] = data

	mockDS.On("Query", mock.Anything, mock.Anything).Return(nil, nil)

	orders, err := storage.ListOrders(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, 1)
	mockDS.AssertExpectations(t)
}

func TestSaveBid(t *testing.T) {
	storage, mockDS := testStorage()
	bid := testBid()

	mockDS.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := storage.SaveBid(context.Background(), bid, "1", 0)
	assert.NoError(t, err)
	mockDS.AssertExpectations(t)
}

func TestGetBids(t *testing.T) {
	storage, mockDS := testStorage()
	bid := testBid()

	// Save bid
	bidStorage := &BidStorage{
		Bid:       bid,
		OrderID:   "1",
		BidIndex:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    bid.Status,
	}
	data, _ := json.Marshal(bidStorage)
	mockDS.data["/bidengine/bids/1/0"] = data

	mockDS.On("Query", mock.Anything, mock.Anything).Return(nil, nil)

	bids, err := storage.GetBids(context.Background(), "1")
	assert.NoError(t, err)
	assert.Len(t, bids, 1)
	mockDS.AssertExpectations(t)
}

func TestSaveMachine(t *testing.T) {
	storage, mockDS := testStorage()
	machine := testMachine()

	mockDS.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := storage.SaveMachine(context.Background(), machine)
	assert.NoError(t, err)
	mockDS.AssertExpectations(t)
}

func TestGetMachine(t *testing.T) {
	storage, mockDS := testStorage()
	machine := testMachine()

	// Save machine
	machineStorage := &MachineStorage{
		Machine:    machine,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastUsed:   time.Now(),
		UsageCount: 0,
	}
	data, _ := json.Marshal(machineStorage)
	mockDS.data["/bidengine/machines/1"] = data

	retrievedMachine, err := storage.GetMachine(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, machine.ID.String(), retrievedMachine.ID.String())
}

func TestGetMachineNotFound(t *testing.T) {
	storage, _ := testStorage()

	_, err := storage.GetMachine(context.Background(), "999")
	assert.Error(t, err)
	assert.Equal(t, ErrMachineNotFound, err)
}

func TestListMachines(t *testing.T) {
	storage, mockDS := testStorage()
	machine := testMachine()

	// Save machine
	machineStorage := &MachineStorage{
		Machine:    machine,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastUsed:   time.Now(),
		UsageCount: 0,
	}
	data, _ := json.Marshal(machineStorage)
	mockDS.data["/bidengine/machines/1"] = data

	mockDS.On("Query", mock.Anything, mock.Anything).Return(nil, nil)

	machines, err := storage.ListMachines(context.Background())
	assert.NoError(t, err)
	assert.Len(t, machines, 1)
	mockDS.AssertExpectations(t)
}

func TestSaveMarketData(t *testing.T) {
	storage, mockDS := testStorage()
	marketData := testMarketData()

	mockDS.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := storage.SaveMarketData(context.Background(), marketData)
	assert.NoError(t, err)
	mockDS.AssertExpectations(t)
}

func TestGetMarketData(t *testing.T) {
	storage, mockDS := testStorage()
	marketData := testMarketData()

	// Save market data
	marketDataStorage := &MarketDataStorage{
		MarketData: marketData,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	data, _ := json.Marshal(marketDataStorage)
	mockDS.data["/bidengine/market/current"] = data

	retrievedMarketData, err := storage.GetMarketData(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, marketData.AveragePricePerSecond.String(), retrievedMarketData.AveragePricePerSecond.String())
}

func TestGetMarketDataNotFound(t *testing.T) {
	storage, _ := testStorage()

	marketData, err := storage.GetMarketData(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, marketData)
}

func TestSaveResourceAllocation(t *testing.T) {
	storage, mockDS := testStorage()
	allocation := testResourceAllocation()

	mockDS.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := storage.SaveResourceAllocation(context.Background(), allocation)
	assert.NoError(t, err)
	mockDS.AssertExpectations(t)
}

func TestListResourceAllocations(t *testing.T) {
	storage, mockDS := testStorage()
	allocation := testResourceAllocation()

	// Save allocation
	data, _ := json.Marshal(allocation)
	mockDS.data["/bidengine/allocations/1"] = data

	mockDS.On("Query", mock.Anything, mock.Anything).Return(nil, nil)

	allocations, err := storage.ListResourceAllocations(context.Background())
	assert.NoError(t, err)
	assert.Len(t, allocations, 1)
	mockDS.AssertExpectations(t)
}

func TestUpdateOrder(t *testing.T) {
	storage, mockDS := testStorage()
	order := testOrder()

	// First save the order
	orderStorage := &OrderStorage{
		Order:      order,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastSynced: time.Now(),
		BidCount:   0,
		Status:     order.Status,
	}
	existingData, _ := json.Marshal(orderStorage)
	mockDS.data["/bidengine/orders/1"] = existingData

	mockDS.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Update order status
	order.Status = types.OrderStatusAccepted
	err := storage.UpdateOrder(context.Background(), order)
	assert.NoError(t, err)
	mockDS.AssertExpectations(t)
}

func TestUpdateBid(t *testing.T) {
	storage, mockDS := testStorage()
	bid := testBid()

	// First save the bid
	bidStorage := &BidStorage{
		Bid:       bid,
		OrderID:   "1",
		BidIndex:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    bid.Status,
	}
	existingData, _ := json.Marshal(bidStorage)
	mockDS.data["/bidengine/bids/1/0"] = existingData

	mockDS.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Update bid status
	bid.Status = types.BidStatusAccepted
	err := storage.UpdateBid(context.Background(), bid, "1", 0)
	assert.NoError(t, err)
	mockDS.AssertExpectations(t)
}

func TestDeleteOrder(t *testing.T) {
	storage, mockDS := testStorage()

	mockDS.On("Delete", mock.Anything, mock.Anything).Return(nil)

	err := storage.DeleteOrder(context.Background(), "1")
	assert.NoError(t, err)
	mockDS.AssertExpectations(t)
}
