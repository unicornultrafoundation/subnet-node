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

// MockProviderContract for testing
type MockProviderContract struct {
	mock.Mock
}

func (m *MockProviderContract) GetMachines(ctx context.Context, providerID *big.Int) ([]Machine, error) {
	args := m.Called(ctx, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Machine), args.Error(1)
}

func (m *MockProviderContract) GetProvider(ctx context.Context, providerID *big.Int) (*Provider, error) {
	args := m.Called(ctx, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Provider), args.Error(1)
}

func (m *MockProviderContract) GetProviderOwner(ctx context.Context, providerID *big.Int) (common.Address, error) {
	args := m.Called(ctx, providerID)
	return args.Get(0).(common.Address), args.Error(1)
}

func (m *MockProviderContract) IsProviderOperatorOrOwner(ctx context.Context, providerID *big.Int, account common.Address) (bool, error) {
	args := m.Called(ctx, providerID, account)
	return args.Bool(0), args.Error(1)
}

func (m *MockProviderContract) IsVerified(ctx context.Context, providerID *big.Int) (bool, error) {
	args := m.Called(ctx, providerID)
	return args.Bool(0), args.Error(1)
}

func (m *MockProviderContract) GetActiveMachinesPaginated(ctx context.Context, providerID *big.Int, start *big.Int, limit *big.Int) ([]Machine, error) {
	args := m.Called(ctx, providerID, start, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Machine), args.Error(1)
}

func (m *MockProviderContract) IsMachineActive(ctx context.Context, providerID *big.Int, machineID *big.Int) (bool, error) {
	args := m.Called(ctx, providerID, machineID)
	return args.Bool(0), args.Error(1)
}

func (m *MockProviderContract) GetMachineResourcePrice(ctx context.Context, providerID *big.Int, machineID *big.Int) (*ResourceUsage, error) {
	args := m.Called(ctx, providerID, machineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ResourceUsage), args.Error(1)
}

func (m *MockProviderContract) AddMachine(ctx context.Context, providerID *big.Int, machine Machine) (*types.Transaction, error) {
	args := m.Called(ctx, providerID, machine)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockProviderContract) UpdateMachine(ctx context.Context, providerID *big.Int, machineID *big.Int, machine Machine) (*types.Transaction, error) {
	args := m.Called(ctx, providerID, machineID, machine)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockProviderContract) RemoveMachine(ctx context.Context, providerID *big.Int, machineID *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, providerID, machineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockProviderContract) SetMachineResourcePrice(ctx context.Context, providerID *big.Int, machineID *big.Int, prices *ResourceUsage) (*types.Transaction, error) {
	args := m.Called(ctx, providerID, machineID, prices)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockProviderContract) ValidateMachineRequirements(ctx context.Context, machineType *big.Int, providerID *big.Int, machineID *big.Int, requirements *ResourceUsage) (bool, error) {
	args := m.Called(ctx, machineType, providerID, machineID, requirements)
	return args.Bool(0), args.Error(1)
}

func TestSyncMachinesFromContract(t *testing.T) {
	// Create test config
	config := &BidEngineConfig{
		ProviderID: big.NewInt(1),
	}

	// Create mocks
	mockProvider := &MockProviderContract{}
	mockLogger := &MockLogger{}
	mockMetrics := &MockMetrics{}
	mockDatastore := ds.NewMapDatastore()

	// Create resource manager
	rm := NewResourceManager(config, mockProvider, mockLogger, mockMetrics, mockDatastore)

	// Create test machines from contract
	contractMachines := []Machine{
		{
			ID:                   big.NewInt(0),
			Active:               true,
			MachineType:          big.NewInt(1),
			Region:               big.NewInt(1),
			CpuCores:             big.NewInt(4),
			GpuCores:             big.NewInt(1),
			GpuMemory:            big.NewInt(8192),
			MemoryMB:             big.NewInt(16384),
			DiskGB:               big.NewInt(100),
			UploadSpeed:          big.NewInt(100),
			DownloadSpeed:        big.NewInt(100),
			CreatedAt:            big.NewInt(time.Now().Unix()),
			UpdatedAt:            big.NewInt(time.Now().Unix()),
			StakeAmount:          big.NewInt(1000),
			RemovedAt:            big.NewInt(0),
			UnlockTime:           big.NewInt(0),
			WithdrawalProcessed:  false,
			Metadata:             "test machine 1",
			CpuPricePerSecond:    big.NewInt(100),
			GpuPricePerSecond:    big.NewInt(200),
			MemoryPricePerSecond: big.NewInt(50),
			DiskPricePerSecond:   big.NewInt(25),
		},
		{
			ID:                   big.NewInt(1),
			Active:               true,
			MachineType:          big.NewInt(2),
			Region:               big.NewInt(2),
			CpuCores:             big.NewInt(8),
			GpuCores:             big.NewInt(2),
			GpuMemory:            big.NewInt(16384),
			MemoryMB:             big.NewInt(32768),
			DiskGB:               big.NewInt(200),
			UploadSpeed:          big.NewInt(200),
			DownloadSpeed:        big.NewInt(200),
			CreatedAt:            big.NewInt(time.Now().Unix()),
			UpdatedAt:            big.NewInt(time.Now().Unix()),
			StakeAmount:          big.NewInt(2000),
			RemovedAt:            big.NewInt(0),
			UnlockTime:           big.NewInt(0),
			WithdrawalProcessed:  false,
			Metadata:             "test machine 2",
			CpuPricePerSecond:    big.NewInt(150),
			GpuPricePerSecond:    big.NewInt(300),
			MemoryPricePerSecond: big.NewInt(75),
			DiskPricePerSecond:   big.NewInt(35),
		},
		{
			ID:                   big.NewInt(2),
			Active:               false, // Inactive machine
			MachineType:          big.NewInt(1),
			Region:               big.NewInt(1),
			CpuCores:             big.NewInt(2),
			GpuCores:             big.NewInt(0),
			GpuMemory:            big.NewInt(0),
			MemoryMB:             big.NewInt(8192),
			DiskGB:               big.NewInt(50),
			UploadSpeed:          big.NewInt(50),
			DownloadSpeed:        big.NewInt(50),
			CreatedAt:            big.NewInt(time.Now().Unix()),
			UpdatedAt:            big.NewInt(time.Now().Unix()),
			StakeAmount:          big.NewInt(500),
			RemovedAt:            big.NewInt(time.Now().Unix()),
			UnlockTime:           big.NewInt(time.Now().Unix() + 86400),
			WithdrawalProcessed:  false,
			Metadata:             "inactive machine",
			CpuPricePerSecond:    big.NewInt(50),
			GpuPricePerSecond:    big.NewInt(0),
			MemoryPricePerSecond: big.NewInt(25),
			DiskPricePerSecond:   big.NewInt(15),
		},
	}

	// Set up mock expectations
	mockProvider.On("GetMachines", mock.Anything, big.NewInt(1)).Return(contractMachines, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Info", "Machine sync completed", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// Test sync
	ctx := context.Background()
	err := rm.syncMachinesFromContract(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, rm.machines, 2) // Only active machines should be synced

	// Check that active machines are synced
	machine0, exists := rm.machines["0"]
	assert.True(t, exists)
	assert.Equal(t, big.NewInt(0), machine0.ID)
	assert.True(t, machine0.Active)
	assert.Equal(t, "test machine 1", machine0.Metadata)

	machine1, exists := rm.machines["1"]
	assert.True(t, exists)
	assert.Equal(t, big.NewInt(1), machine1.ID)
	assert.True(t, machine1.Active)
	assert.Equal(t, "test machine 2", machine1.Metadata)

	// Check that inactive machine is not synced
	_, exists = rm.machines["2"]
	assert.False(t, exists)

	// Verify mock expectations
	mockProvider.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSyncMachinesFromContractNoProvider(t *testing.T) {
	// Create test config
	config := &BidEngineConfig{
		ProviderID: big.NewInt(1),
	}

	// Create mocks
	mockLogger := &MockLogger{}
	mockMetrics := &MockMetrics{}
	mockDatastore := ds.NewMapDatastore()

	// Create resource manager without provider contract
	rm := NewResourceManager(config, nil, mockLogger, mockMetrics, mockDatastore)

	// Set up mock expectations
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

	// Test sync
	ctx := context.Background()
	err := rm.syncMachinesFromContract(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, rm.machines, 0)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
}

func TestSyncMachinesFromContractNoProviderID(t *testing.T) {
	// Create test config without provider ID
	config := &BidEngineConfig{}

	// Create mocks
	mockProvider := &MockProviderContract{}
	mockLogger := &MockLogger{}
	mockMetrics := &MockMetrics{}
	mockDatastore := ds.NewMapDatastore()

	// Create resource manager
	rm := NewResourceManager(config, mockProvider, mockLogger, mockMetrics, mockDatastore)

	// Set up mock expectations
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

	// Test sync
	ctx := context.Background()
	err := rm.syncMachinesFromContract(ctx)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider ID not configured")
	assert.Len(t, rm.machines, 0)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
}

func TestMachineNeedsUpdate(t *testing.T) {
	// Create test config
	config := &BidEngineConfig{
		ProviderID: big.NewInt(1),
	}

	// Create mocks
	mockProvider := &MockProviderContract{}
	mockLogger := &MockLogger{}
	mockMetrics := &MockMetrics{}
	mockDatastore := ds.NewMapDatastore()

	// Create resource manager
	rm := NewResourceManager(config, mockProvider, mockLogger, mockMetrics, mockDatastore)

	// Create test machines
	existingMachine := &Machine{
		ID:                   big.NewInt(0),
		Active:               true,
		UpdatedAt:            big.NewInt(1000),
		CpuPricePerSecond:    big.NewInt(100),
		GpuPricePerSecond:    big.NewInt(200),
		MemoryPricePerSecond: big.NewInt(50),
		DiskPricePerSecond:   big.NewInt(25),
	}

	updatedMachine := &Machine{
		ID:                   big.NewInt(0),
		Active:               true,
		UpdatedAt:            big.NewInt(2000), // Newer timestamp
		CpuPricePerSecond:    big.NewInt(100),
		GpuPricePerSecond:    big.NewInt(200),
		MemoryPricePerSecond: big.NewInt(50),
		DiskPricePerSecond:   big.NewInt(25),
	}

	// Test that machine needs update due to newer timestamp
	needsUpdate := rm.machineNeedsUpdate(existingMachine, updatedMachine)
	assert.True(t, needsUpdate)

	// Test that machine doesn't need update with same timestamp
	needsUpdate = rm.machineNeedsUpdate(existingMachine, existingMachine)
	assert.False(t, needsUpdate)

	// Test that machine needs update due to price changes
	priceChangedMachine := &Machine{
		ID:                   big.NewInt(0),
		Active:               true,
		UpdatedAt:            big.NewInt(1000), // Same timestamp
		CpuPricePerSecond:    big.NewInt(150),  // Different price
		GpuPricePerSecond:    big.NewInt(200),
		MemoryPricePerSecond: big.NewInt(50),
		DiskPricePerSecond:   big.NewInt(25),
	}

	needsUpdate = rm.machineNeedsUpdate(existingMachine, priceChangedMachine)
	assert.True(t, needsUpdate)
}
