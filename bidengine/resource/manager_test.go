package resource

import (
	"context"
	"math/big"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// --- Mocks ---
type mockProviderContract struct{ types.ProviderContract }

func (m *mockProviderContract) GetMachines(ctx context.Context, providerID *big.Int) ([]*types.Machine, error) {
	return []*types.Machine{}, nil
}

// Minimal mock storage
type mockStorage struct{ types.Storage }

func (m *mockStorage) SaveMachine(ctx context.Context, machine *types.Machine) error { return nil }
func (m *mockStorage) ListMachines(ctx context.Context) ([]*types.Machine, error) {
	return []*types.Machine{}, nil
}
func (m *mockStorage) SaveResourceAllocation(ctx context.Context, allocation *types.ResourceAllocation) error {
	return nil
}
func (m *mockStorage) ListResourceAllocations(ctx context.Context) ([]*types.ResourceAllocation, error) {
	return []*types.ResourceAllocation{}, nil
}

// Minimal mock metrics
type mockMetrics struct{ types.Metrics }

func (m *mockMetrics) RecordResourceAllocation(usage *types.ResourceUsage) {}

func TestRegisterAndUnregisterMachine(t *testing.T) {
	logger := logrus.New()
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockProviderContract{},
		logger,
		&mockMetrics{},
		&mockStorage{},
	)
	ctx := context.Background()
	machine := &types.Machine{ID: big.NewInt(42), Active: true}

	err := manager.RegisterMachine(ctx, machine)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(manager.GetAllMachines(ctx)))

	err = manager.UnregisterMachine(ctx, big.NewInt(42))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(manager.GetAllMachines(ctx)))
}

func TestAllocateAndDeallocateResources(t *testing.T) {
	logger := logrus.New()
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockProviderContract{},
		logger,
		&mockMetrics{},
		&mockStorage{},
	)
	ctx := context.Background()

	// Create a machine with sufficient resources
	machine := &types.Machine{
		ID:       big.NewInt(1),
		CpuCores: big.NewInt(8),
		GpuCores: big.NewInt(2),
		MemoryMB: big.NewInt(16384),
		DiskGB:   big.NewInt(100),
		Active:   true,
	}

	// Register the machine first
	err := manager.RegisterMachine(ctx, machine)
	assert.NoError(t, err)

	// Create resource usage for allocation
	usage := &types.ResourceUsage{
		CPUUsed:    big.NewInt(2),
		GPUUsed:    big.NewInt(1),
		MemoryUsed: big.NewInt(4096),
		DiskUsed:   big.NewInt(20),
	}
	orderID := big.NewInt(100)

	// Allocate resources
	err = manager.AllocateResources(ctx, orderID, machine, usage)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(manager.allocatedResources))

	// Deallocate resources
	err = manager.DeallocateResources(ctx, orderID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(manager.allocatedResources))
}

func TestCanAllocateResources(t *testing.T) {
	logger := logrus.New()
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockProviderContract{},
		logger,
		&mockMetrics{},
		&mockStorage{},
	)
	ctx := context.Background()

	machine := &types.Machine{
		ID:       big.NewInt(1),
		CpuCores: big.NewInt(4),
		Active:   true,
	}

	err := manager.RegisterMachine(ctx, machine)
	assert.NoError(t, err)

	// Test allocation within limits
	usage := &types.ResourceUsage{CPUUsed: big.NewInt(2)}
	can, err := manager.CanAllocateResources(ctx, machine, usage)
	assert.NoError(t, err)
	assert.True(t, can)

	// Test allocation exceeding limits
	usage2 := &types.ResourceUsage{CPUUsed: big.NewInt(10)}
	can, err = manager.CanAllocateResources(ctx, machine, usage2)
	assert.NoError(t, err)
	assert.False(t, can)
}

func TestGetMachine(t *testing.T) {
	logger := logrus.New()
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockProviderContract{},
		logger,
		&mockMetrics{},
		&mockStorage{},
	)
	ctx := context.Background()

	machine := &types.Machine{ID: big.NewInt(123), Active: true}
	err := manager.RegisterMachine(ctx, machine)
	assert.NoError(t, err)

	// Test getting existing machine
	retrieved, err := manager.GetMachine(ctx, big.NewInt(123))
	assert.NoError(t, err)
	assert.Equal(t, machine.ID, retrieved.ID)

	// Test getting non-existent machine
	_, err = manager.GetMachine(ctx, big.NewInt(999))
	assert.Error(t, err)
}

func TestDuplicateRegistration(t *testing.T) {
	logger := logrus.New()
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockProviderContract{},
		logger,
		&mockMetrics{},
		&mockStorage{},
	)
	ctx := context.Background()

	machine := &types.Machine{ID: big.NewInt(456), Active: true}

	// First registration should succeed
	err := manager.RegisterMachine(ctx, machine)
	assert.NoError(t, err)

	// Second registration should fail
	err = manager.RegisterMachine(ctx, machine)
	assert.Error(t, err)
}

func TestUnregisterNonExistentMachine(t *testing.T) {
	logger := logrus.New()
	manager := NewManager(
		&types.BidEngineConfig{ProviderID: big.NewInt(1)},
		&mockProviderContract{},
		logger,
		&mockMetrics{},
		&mockStorage{},
	)
	ctx := context.Background()

	// Try to unregister a machine that doesn't exist
	err := manager.UnregisterMachine(ctx, big.NewInt(999))
	assert.Error(t, err)
}
