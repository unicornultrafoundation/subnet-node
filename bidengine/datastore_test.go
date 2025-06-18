package bidengine

import (
	"context"
	"testing"
	"time"

	"github.com/holiman/uint256"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDatastore mocks the repo.Datastore interface
type MockDatastore struct {
	mock.Mock
}

func (m *MockDatastore) Put(ctx context.Context, key datastore.Key, value []byte) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockDatastore) Get(ctx context.Context, key datastore.Key) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockDatastore) Delete(ctx context.Context, key datastore.Key) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockDatastore) Query(ctx context.Context, q query.Query) (query.Results, error) {
	args := m.Called(ctx, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(query.Results), args.Error(1)
}

// Close mocks the Close method of the repo.Datastore interface
func (m *MockDatastore) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Batch mocks the Batch method of the repo.Datastore interface
func (m *MockDatastore) Batch(ctx context.Context) (datastore.Batch, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(datastore.Batch), args.Error(1)
}

// GetSize mocks the GetSize method of the repo.Datastore interface
func (m *MockDatastore) GetSize(ctx context.Context, key datastore.Key) (int, error) {
	args := m.Called(ctx, key)
	return args.Int(0), args.Error(1)
}

// Has mocks the Has method of the repo.Datastore interface
func (m *MockDatastore) Has(ctx context.Context, key datastore.Key) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

// Sync mocks the Sync method of the repo.Datastore interface
func (m *MockDatastore) Sync(ctx context.Context, prefix datastore.Key) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func TestStoreSaveBid(t *testing.T) {
	ctx := context.Background()
	mockDs := new(MockDatastore)
	log := logrus.New()

	store, err := NewStore(mockDs, log)
	require.NoError(t, err)

	bid := &Bid{
		ID:           uint256.NewInt(123),
		OrderId:      uint256.NewInt(123),
		ProviderId:   uint256.NewInt(1),
		MachineId:    uint256.NewInt(2),
		PricePerSec:  uint256.NewInt(1000000),
		Status:       BidStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpirationAt: time.Now().Add(time.Hour),
		Requirements: &BidRequirements{
			MinCPUCores: uint256.NewInt(4),
		},
		TxHash: "0xabc123",
	}

	// Expect the datastore to receive a Put call
	mockDs.On("Put", ctx, datastore.NewKey(bidKey(bid.ID.String()).String()), mock.Anything).Return(nil)

	err = store.SaveBid(ctx, bid)
	assert.NoError(t, err)
	mockDs.AssertExpectations(t)

	// Test error case - nil bid
	err = store.SaveBid(ctx, nil)
	assert.Error(t, err)
}

func TestStoreGetBidByID(t *testing.T) {
	ctx := context.Background()
	mockDs := new(MockDatastore)
	log := logrus.New()

	store, err := NewStore(mockDs, log)
	require.NoError(t, err)

	bidID := "123"

	// Test 1: Bid found
	mockDs.On("Get", ctx, bidKey(bidID)).Return([]byte(`{"id":"123","providerId":"1","machineId":"2","status":0}`), nil).Once()

	bid, err := store.GetBidByID(ctx, bidID)
	assert.NoError(t, err)
	assert.NotNil(t, bid)

	// Test 2: Bid not found
	mockDs.On("Get", ctx, bidKey("999")).Return(nil, datastore.ErrNotFound).Once()

	bid, err = store.GetBidByID(ctx, "999")
	assert.NoError(t, err) // No error, just nil bid
	assert.Nil(t, bid)

	// Test 3: Datastore error
	mockDs.On("Get", ctx, bidKey("error")).Return(nil, assert.AnError).Once()

	bid, err = store.GetBidByID(ctx, "error")
	assert.Error(t, err)
	assert.Nil(t, bid)

	mockDs.AssertExpectations(t)
}

func TestStoreDeleteBid(t *testing.T) {
	ctx := context.Background()
	mockDs := new(MockDatastore)
	log := logrus.New()

	store, err := NewStore(mockDs, log)
	require.NoError(t, err)

	bidID := "123"

	// Test successful deletion
	mockDs.On("Delete", ctx, bidKey(bidID)).Return(nil).Once()

	err = store.DeleteBid(ctx, bidID)
	assert.NoError(t, err)

	// Test deletion of non-existent bid (should not error)
	mockDs.On("Delete", ctx, bidKey("999")).Return(datastore.ErrNotFound).Once()

	err = store.DeleteBid(ctx, "999")
	assert.NoError(t, err)

	// Test datastore error
	mockDs.On("Delete", ctx, bidKey("error")).Return(assert.AnError).Once()

	err = store.DeleteBid(ctx, "error")
	assert.Error(t, err)

	mockDs.AssertExpectations(t)
}
