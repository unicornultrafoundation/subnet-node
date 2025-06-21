package bidengine

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	ds "github.com/ipfs/go-datastore"
)

// ResourceManagerService manages resource allocation and machine registration
type ResourceManagerService struct {
	config   *BidEngineConfig
	provider ProviderContract
	logger   Logger
	metrics  Metrics
	storage  *Storage

	mu        sync.RWMutex
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc

	// Resource tracking
	machines           map[string]*Machine
	allocatedResources map[string]*ResourceUsage // orderID -> usage
}

// NewResourceManager creates a new ResourceManagerService instance
func NewResourceManager(
	config *BidEngineConfig,
	provider ProviderContract,
	logger Logger,
	metrics Metrics,
	datastore ds.Datastore,
) *ResourceManagerService {
	ctx, cancel := context.WithCancel(context.Background())

	return &ResourceManagerService{
		config:             config,
		provider:           provider,
		logger:             logger,
		metrics:            metrics,
		storage:            NewStorage(datastore, logger),
		ctx:                ctx,
		cancel:             cancel,
		machines:           make(map[string]*Machine),
		allocatedResources: make(map[string]*ResourceUsage),
	}
}

// Start starts the resource manager
func (rm *ResourceManagerService) Start(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.isRunning {
		return nil
	}

	rm.logger.Info("Starting ResourceManager")
	rm.isRunning = true

	// Load persisted machines from storage
	if err := rm.loadPersistedMachines(ctx); err != nil {
		rm.logger.Warn("Failed to load persisted machines", "error", err)
	}

	rm.logger.Info("ResourceManager started successfully")
	return nil
}

// Stop stops the resource manager
func (rm *ResourceManagerService) Stop(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.isRunning {
		return nil
	}

	rm.logger.Info("Stopping ResourceManager")
	rm.isRunning = false

	// Save current state to storage
	if err := rm.savePersistedMachines(ctx); err != nil {
		rm.logger.Warn("Failed to save persisted machines", "error", err)
	}

	// Cancel context
	rm.cancel()

	rm.logger.Info("ResourceManager stopped successfully")
	return nil
}

// RegisterMachine registers a new machine
func (rm *ResourceManagerService) RegisterMachine(ctx context.Context, machine *Machine) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	machineID := machine.ID.String()

	// Check if machine is already registered
	if _, exists := rm.machines[machineID]; exists {
		return fmt.Errorf("machine %s is already registered", machineID)
	}

	// Register machine
	rm.machines[machineID] = machine

	// Save machine to storage
	if err := rm.storage.SaveMachine(ctx, machine); err != nil {
		rm.logger.Warn("Failed to save machine to storage", "error", err)
	}

	rm.logger.Info("Machine registered successfully", "machineID", machineID)
	return nil
}

// UnregisterMachine unregisters a machine
func (rm *ResourceManagerService) UnregisterMachine(ctx context.Context, machineID *big.Int) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	machineIDStr := machineID.String()

	if _, exists := rm.machines[machineIDStr]; !exists {
		return fmt.Errorf("machine %s is not registered", machineIDStr)
	}

	delete(rm.machines, machineIDStr)
	rm.logger.Info("Machine unregistered", "machineID", machineID)
	return nil
}

// GetAllMachines returns all registered machines
func (rm *ResourceManagerService) GetAllMachines(ctx context.Context) []*Machine {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	machines := make([]*Machine, 0, len(rm.machines))
	for _, machine := range rm.machines {
		machines = append(machines, machine)
	}

	return machines
}

// GetMachine retrieves a machine by ID
func (rm *ResourceManagerService) GetMachine(ctx context.Context, machineID *big.Int) (*Machine, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	machineIDStr := machineID.String()
	machine, exists := rm.machines[machineIDStr]
	if !exists {
		return nil, fmt.Errorf("machine %s not found", machineIDStr)
	}

	return machine, nil
}

// AllocateResources allocates resources for an order
func (rm *ResourceManagerService) AllocateResources(ctx context.Context, orderID *big.Int, machine *Machine, usage *ResourceUsage) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	orderIDStr := orderID.String()

	// Check if resources are already allocated for this order
	if _, exists := rm.allocatedResources[orderIDStr]; exists {
		return ErrResourcesAlreadyAllocated
	}

	// Validate that machine can allocate the required resources
	canAllocate, err := rm.CanAllocateResources(ctx, machine, usage)
	if err != nil {
		return fmt.Errorf("failed to check resource availability: %w", err)
	}

	if !canAllocate {
		return ErrInsufficientResources
	}

	// Allocate resources
	rm.allocatedResources[orderIDStr] = usage

	rm.logger.Info("Resources allocated successfully",
		"orderID", orderID,
		"machineID", machine.ID,
		"usage", usage)

	rm.metrics.RecordResourceAllocation(usage)
	return nil
}

// DeallocateResources deallocates resources for an order
func (rm *ResourceManagerService) DeallocateResources(ctx context.Context, orderID *big.Int) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	orderIDStr := orderID.String()

	if _, exists := rm.allocatedResources[orderIDStr]; !exists {
		return ErrResourcesNotAllocated
	}

	delete(rm.allocatedResources, orderIDStr)
	rm.logger.Info("Resources deallocated", "orderID", orderID)
	return nil
}

// GetCurrentUsage gets the current resource usage for a machine
func (rm *ResourceManagerService) GetCurrentUsage(ctx context.Context, machine *Machine) (*ResourceUsage, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Calculate total usage from all allocated resources
	var totalUsage ResourceUsage

	for _, usage := range rm.allocatedResources {
		if totalUsage.CPUUsed == nil {
			totalUsage.CPUUsed = big.NewInt(0)
		}
		if totalUsage.GPUUsed == nil {
			totalUsage.GPUUsed = big.NewInt(0)
		}
		if totalUsage.MemoryUsed == nil {
			totalUsage.MemoryUsed = big.NewInt(0)
		}
		if totalUsage.DiskUsed == nil {
			totalUsage.DiskUsed = big.NewInt(0)
		}
		if totalUsage.NetworkUsed == nil {
			totalUsage.NetworkUsed = big.NewInt(0)
		}

		if usage.CPUUsed != nil {
			totalUsage.CPUUsed.Add(totalUsage.CPUUsed, usage.CPUUsed)
		}
		if usage.GPUUsed != nil {
			totalUsage.GPUUsed.Add(totalUsage.GPUUsed, usage.GPUUsed)
		}
		if usage.MemoryUsed != nil {
			totalUsage.MemoryUsed.Add(totalUsage.MemoryUsed, usage.MemoryUsed)
		}
		if usage.DiskUsed != nil {
			totalUsage.DiskUsed.Add(totalUsage.DiskUsed, usage.DiskUsed)
		}
		if usage.NetworkUsed != nil {
			totalUsage.NetworkUsed.Add(totalUsage.NetworkUsed, usage.NetworkUsed)
		}
	}

	return &totalUsage, nil
}

// GetAvailableResources gets the available resources for a machine
func (rm *ResourceManagerService) GetAvailableResources(ctx context.Context, machine *Machine) (*ResourceUsage, error) {
	currentUsage, err := rm.GetCurrentUsage(ctx, machine)
	if err != nil {
		return nil, err
	}

	// Calculate available resources
	available := &ResourceUsage{}

	if machine.CpuCores != nil && currentUsage.CPUUsed != nil {
		available.CPUUsed = new(big.Int).Sub(machine.CpuCores, currentUsage.CPUUsed)
	} else if machine.CpuCores != nil {
		available.CPUUsed = machine.CpuCores
	}

	if machine.GpuCores != nil && currentUsage.GPUUsed != nil {
		available.GPUUsed = new(big.Int).Sub(machine.GpuCores, currentUsage.GPUUsed)
	} else if machine.GpuCores != nil {
		available.GPUUsed = machine.GpuCores
	}

	if machine.MemoryMB != nil && currentUsage.MemoryUsed != nil {
		available.MemoryUsed = new(big.Int).Sub(machine.MemoryMB, currentUsage.MemoryUsed)
	} else if machine.MemoryMB != nil {
		available.MemoryUsed = machine.MemoryMB
	}

	if machine.DiskGB != nil && currentUsage.DiskUsed != nil {
		available.DiskUsed = new(big.Int).Sub(machine.DiskGB, currentUsage.DiskUsed)
	} else if machine.DiskGB != nil {
		available.DiskUsed = machine.DiskGB
	}

	if machine.UploadSpeed != nil && currentUsage.NetworkUsed != nil {
		available.NetworkUsed = new(big.Int).Sub(machine.UploadSpeed, currentUsage.NetworkUsed)
	} else if machine.UploadSpeed != nil {
		available.NetworkUsed = machine.UploadSpeed
	}

	return available, nil
}

// CanAllocateResources checks if a machine can allocate the required resources
func (rm *ResourceManagerService) CanAllocateResources(ctx context.Context, machine *Machine, required *ResourceUsage) (bool, error) {
	available, err := rm.GetAvailableResources(ctx, machine)
	if err != nil {
		return false, err
	}

	// Check if available resources meet requirements
	if required.CPUUsed != nil && (available.CPUUsed == nil || available.CPUUsed.Cmp(required.CPUUsed) < 0) {
		return false, nil
	}

	if required.GPUUsed != nil && (available.GPUUsed == nil || available.GPUUsed.Cmp(required.GPUUsed) < 0) {
		return false, nil
	}

	if required.MemoryUsed != nil && (available.MemoryUsed == nil || available.MemoryUsed.Cmp(required.MemoryUsed) < 0) {
		return false, nil
	}

	if required.DiskUsed != nil && (available.DiskUsed == nil || available.DiskUsed.Cmp(required.DiskUsed) < 0) {
		return false, nil
	}

	if required.NetworkUsed != nil && (available.NetworkUsed == nil || available.NetworkUsed.Cmp(required.NetworkUsed) < 0) {
		return false, nil
	}

	return true, nil
}

// StartResource starts resource allocation for an order
func (rm *ResourceManagerService) StartResource(ctx context.Context, orderID *big.Int, machine *Machine) error {
	rm.logger.Info("Starting resource allocation", "orderID", orderID, "machineID", machine.ID)
	return nil
}

// StopResource stops resource allocation for an order
func (rm *ResourceManagerService) StopResource(ctx context.Context, orderID *big.Int) error {
	rm.logger.Info("Stopping resource allocation", "orderID", orderID)
	return rm.DeallocateResources(ctx, orderID)
}

// loadPersistedMachines loads machines from storage on startup
func (rm *ResourceManagerService) loadPersistedMachines(ctx context.Context) error {
	rm.logger.Info("Loading persisted machines from storage")

	machines, err := rm.storage.ListMachines(ctx)
	if err != nil {
		return fmt.Errorf("failed to load machines: %w", err)
	}

	for _, machine := range machines {
		machineID := machine.ID.String()
		rm.machines[machineID] = machine
		rm.logger.Debug("Loaded machine from storage", "machineID", machineID)
	}

	rm.logger.Info("Loaded persisted machines", "count", len(machines))
	return nil
}

// savePersistedMachines saves current machine state to storage
func (rm *ResourceManagerService) savePersistedMachines(ctx context.Context) error {
	rm.logger.Info("Saving persisted machines to storage")

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for machineID, machine := range rm.machines {
		if err := rm.storage.SaveMachine(ctx, machine); err != nil {
			rm.logger.Warn("Failed to save machine to storage", "machineID", machineID, "error", err)
		}
	}

	return nil
}

// GetStats returns resource manager statistics
func (rm *ResourceManagerService) GetStats() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return map[string]interface{}{
		"registeredMachines": len(rm.machines),
		"allocatedResources": len(rm.allocatedResources),
		"isRunning":          rm.isRunning,
	}
}
