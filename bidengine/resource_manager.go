package bidengine

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

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

	// Contract sync
	lastSyncTime time.Time
	syncInterval time.Duration
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
		syncInterval:       5 * time.Minute, // Sync every 5 minutes
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

	// Load persisted allocations from storage
	if err := rm.loadPersistedAllocations(ctx); err != nil {
		rm.logger.Warn("Failed to load persisted allocations", "error", err)
	}

	// Initial sync with contract
	if err := rm.syncMachinesFromContract(ctx); err != nil {
		rm.logger.Warn("Failed to sync machines from contract", "error", err)
	}

	// Start sync loop
	go rm.syncLoop(ctx)

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

	if err := rm.savePersistedAllocations(ctx); err != nil {
		rm.logger.Warn("Failed to save persisted allocations", "error", err)
	}

	// Cancel context
	rm.cancel()

	rm.logger.Info("ResourceManager stopped successfully")
	return nil
}

// syncLoop runs the periodic sync with the contract
func (rm *ResourceManagerService) syncLoop(ctx context.Context) {
	ticker := time.NewTicker(rm.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := rm.syncMachinesFromContract(ctx); err != nil {
				rm.logger.Warn("Failed to sync machines from contract", "error", err)
			}
		}
	}
}

// syncMachinesFromContract syncs machines from the smart contract
func (rm *ResourceManagerService) syncMachinesFromContract(ctx context.Context) error {
	if rm.provider == nil {
		rm.logger.Debug("No provider contract available, skipping sync")
		return nil
	}

	rm.logger.Debug("Starting machine sync from contract")

	// Get provider ID from config
	providerID := rm.config.ProviderID
	if providerID == nil {
		rm.logger.Warn("Provider ID not configured, cannot sync machines")
		return fmt.Errorf("provider ID not configured")
	}

	// Get machines from contract
	contractMachines, err := rm.provider.GetMachines(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to get machines from contract: %w", err)
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Track which machines are still active in contract
	activeMachineIDs := make(map[string]bool)

	for i, contractMachine := range contractMachines {
		// Only sync active machines
		if !contractMachine.Active {
			continue
		}

		// Use index as machine ID since contract returns machines in order
		machineID := big.NewInt(int64(i))
		machineIDStr := machineID.String()
		activeMachineIDs[machineIDStr] = true

		// Convert contract machine to our Machine type
		machine := &Machine{
			ID:                   machineID,
			Active:               contractMachine.Active,
			MachineType:          contractMachine.MachineType,
			Region:               contractMachine.Region,
			CpuCores:             contractMachine.CpuCores,
			GpuCores:             contractMachine.GpuCores,
			GpuMemory:            contractMachine.GpuMemory,
			MemoryMB:             contractMachine.MemoryMB,
			DiskGB:               contractMachine.DiskGB,
			UploadSpeed:          contractMachine.UploadSpeed,
			DownloadSpeed:        contractMachine.DownloadSpeed,
			CreatedAt:            contractMachine.CreatedAt,
			UpdatedAt:            contractMachine.UpdatedAt,
			StakeAmount:          contractMachine.StakeAmount,
			RemovedAt:            contractMachine.RemovedAt,
			UnlockTime:           contractMachine.UnlockTime,
			WithdrawalProcessed:  contractMachine.WithdrawalProcessed,
			Metadata:             contractMachine.Metadata,
			CpuPricePerSecond:    contractMachine.CpuPricePerSecond,
			GpuPricePerSecond:    contractMachine.GpuPricePerSecond,
			MemoryPricePerSecond: contractMachine.MemoryPricePerSecond,
			DiskPricePerSecond:   contractMachine.DiskPricePerSecond,
		}

		// Check if machine already exists and needs update
		existingMachine, exists := rm.machines[machineIDStr]
		if !exists {
			// New machine from contract
			rm.machines[machineIDStr] = machine
			rm.logger.Info("Added new machine from contract", "machineID", machineIDStr)

			// Save to storage
			if err := rm.storage.SaveMachine(ctx, machine); err != nil {
				rm.logger.Warn("Failed to save machine to storage", "machineID", machineIDStr, "error", err)
			}
		} else {
			// Update existing machine if needed
			if rm.machineNeedsUpdate(existingMachine, machine) {
				rm.machines[machineIDStr] = machine
				rm.logger.Info("Updated machine from contract", "machineID", machineIDStr)

				// Update in storage
				if err := rm.storage.SaveMachine(ctx, machine); err != nil {
					rm.logger.Warn("Failed to update machine in storage", "machineID", machineIDStr, "error", err)
				}
			}
		}
	}

	// Remove machines that are no longer active in contract
	for machineID := range rm.machines {
		if !activeMachineIDs[machineID] {
			delete(rm.machines, machineID)
			rm.logger.Info("Removed inactive machine from contract", "machineID", machineID)
		}
	}

	rm.lastSyncTime = time.Now()
	rm.logger.Info("Machine sync completed",
		"totalMachines", len(rm.machines),
		"contractMachines", len(contractMachines),
		"activeMachines", len(activeMachineIDs))

	return nil
}

// machineNeedsUpdate checks if a machine needs to be updated
func (rm *ResourceManagerService) machineNeedsUpdate(existing, updated *Machine) bool {
	// Compare key fields that might change
	if existing.UpdatedAt.Cmp(updated.UpdatedAt) < 0 {
		return true
	}

	if existing.CpuPricePerSecond.Cmp(updated.CpuPricePerSecond) != 0 ||
		existing.GpuPricePerSecond.Cmp(updated.GpuPricePerSecond) != 0 ||
		existing.MemoryPricePerSecond.Cmp(updated.MemoryPricePerSecond) != 0 ||
		existing.DiskPricePerSecond.Cmp(updated.DiskPricePerSecond) != 0 {
		return true
	}

	return false
}

// ForceSync forces an immediate sync with the contract
func (rm *ResourceManagerService) ForceSync(ctx context.Context) error {
	return rm.syncMachinesFromContract(ctx)
}

// GetLastSyncTime returns the last sync time
func (rm *ResourceManagerService) GetLastSyncTime() time.Time {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.lastSyncTime
}

// RegisterMachine registers a new machine locally (for backward compatibility)
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

// UnregisterMachine unregisters a machine locally (for backward compatibility)
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

// AllocateResources allocates resources for an order and persists to database
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

	// Allocate resources in memory
	rm.allocatedResources[orderIDStr] = usage

	// Persist allocation to database
	allocation := &ResourceAllocation{
		OrderID:   orderID,
		Usage:     usage,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := rm.storage.SaveResourceAllocation(ctx, allocation); err != nil {
		// Rollback memory allocation if database save fails
		delete(rm.allocatedResources, orderIDStr)
		return fmt.Errorf("failed to persist resource allocation: %w", err)
	}

	rm.logger.Info("Resources allocated successfully",
		"orderID", orderID,
		"machineID", machine.ID,
		"usage", usage)

	rm.metrics.RecordResourceAllocation(usage)
	return nil
}

// DeallocateResources deallocates resources for an order and removes from database
func (rm *ResourceManagerService) DeallocateResources(ctx context.Context, orderID *big.Int) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	orderIDStr := orderID.String()

	if _, exists := rm.allocatedResources[orderIDStr]; !exists {
		return ErrResourcesNotAllocated
	}

	// Remove from memory
	delete(rm.allocatedResources, orderIDStr)

	// Remove from database
	allocation := &ResourceAllocation{
		OrderID: orderID,
	}

	// Note: Storage doesn't have a delete method, but we can mark as deallocated
	// by setting usage to nil or removing the allocation
	allocation.Usage = nil
	allocation.UpdatedAt = time.Now()

	if err := rm.storage.SaveResourceAllocation(ctx, allocation); err != nil {
		rm.logger.Warn("Failed to update deallocation in storage", "orderID", orderIDStr, "error", err)
	}

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

// loadPersistedAllocations loads allocated resources from storage on startup
func (rm *ResourceManagerService) loadPersistedAllocations(ctx context.Context) error {
	rm.logger.Info("Loading persisted resource allocations from storage")

	allocations, err := rm.storage.ListResourceAllocations(ctx)
	if err != nil {
		return fmt.Errorf("failed to load resource allocations: %w", err)
	}

	for _, allocation := range allocations {
		// Only load allocations that have actual usage (not deallocated)
		if allocation.Usage != nil {
			orderIDStr := allocation.OrderID.String()
			rm.allocatedResources[orderIDStr] = allocation.Usage
			rm.logger.Debug("Loaded resource allocation from storage", "orderID", orderIDStr)
		}
	}

	rm.logger.Info("Loaded persisted resource allocations", "count", len(allocations))
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

// savePersistedAllocations saves current resource allocations to storage
func (rm *ResourceManagerService) savePersistedAllocations(ctx context.Context) error {
	rm.logger.Info("Saving persisted resource allocations to storage")

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for orderIDStr, usage := range rm.allocatedResources {
		orderID, ok := new(big.Int).SetString(orderIDStr, 10)
		if !ok {
			rm.logger.Warn("Invalid order ID format", "orderID", orderIDStr)
			continue
		}

		allocation := &ResourceAllocation{
			OrderID:   orderID,
			Usage:     usage,
			UpdatedAt: time.Now(),
		}

		if err := rm.storage.SaveResourceAllocation(ctx, allocation); err != nil {
			rm.logger.Warn("Failed to save resource allocation to storage", "orderID", orderIDStr, "error", err)
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
		"lastSyncTime":       rm.lastSyncTime,
	}
}
