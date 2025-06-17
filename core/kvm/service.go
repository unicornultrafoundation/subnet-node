package kvm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
)

// Service implements a simple KVM management service
type Service struct {
	config      *config.C
	logger      *logrus.Entry
	datastore   datastore.Datastore
	resourceSvc resource.Service

	// In-memory VM storage (in production, this would be libvirt)
	vms     map[string]*VM
	vmStats map[string]*VMStats
	mu      sync.RWMutex

	// Configuration
	enabled     bool
	maxVMs      int
	maxCPUCores int
	maxMemoryMB int
	maxDiskGB   int
}

// NewService creates a new KVM service
func NewService(
	cfg *config.C,
	logger *logrus.Entry,
	ds datastore.Datastore,
	resourceSvc resource.Service,
) *Service {
	return &Service{
		config:      cfg,
		logger:      logger.WithField("service", "kvm"),
		datastore:   ds,
		resourceSvc: resourceSvc,
		vms:         make(map[string]*VM),
		vmStats:     make(map[string]*VMStats),

		// Load configuration
		enabled:     cfg.GetBool("kvm.enabled", false),
		maxVMs:      cfg.GetInt("kvm.max_vms", 5),
		maxCPUCores: cfg.GetInt("kvm.max_cpu_cores", 4),
		maxMemoryMB: cfg.GetInt("kvm.max_memory_mb", 4096),
		maxDiskGB:   cfg.GetInt("kvm.max_disk_gb", 50),
	}
}

// IsEnabled returns whether the KVM service is enabled
func (s *Service) IsEnabled() bool {
	return s.enabled
}

// Start starts the KVM service
func (s *Service) Start(ctx context.Context) error {
	if !s.enabled {
		s.logger.Info("KVM service is disabled")
		return nil
	}

	s.logger.Info("Starting KVM service")

	// Load existing VMs from datastore
	if err := s.loadVMsFromDatastore(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to load VMs from datastore")
		return fmt.Errorf("failed to load VMs: %w", err)
	}

	// Start background monitoring
	go s.monitorVMs(ctx)

	s.logger.WithField("max_vms", s.maxVMs).Info("KVM service started")
	return nil
}

// Stop stops the KVM service
func (s *Service) Stop(ctx context.Context) error {
	if !s.enabled {
		return nil
	}

	s.logger.Info("Stopping KVM service")

	s.mu.Lock()
	defer s.mu.Unlock()

	// In a real implementation, we would gracefully stop all VMs
	for id, vm := range s.vms {
		if vm.Status == VMStatusRunning {
			s.logger.WithField("vm_id", id).Info("Stopping VM during service shutdown")
			vm.Status = VMStatusStopped
			vm.UpdatedAt = time.Now()
			// Save VM state
			s.saveVMToDatastore(ctx, vm)
		}
	}

	s.logger.Info("KVM service stopped")
	return nil
}

// CreateVM creates a new virtual machine
func (s *Service) CreateVM(ctx context.Context, req *CreateVMRequest) (*VM, error) {
	if !s.enabled {
		return nil, fmt.Errorf("KVM service is not enabled")
	}

	s.logger.WithFields(logrus.Fields{
		"name":      req.Name,
		"cpu_cores": req.CPUCores,
		"memory_mb": req.MemoryMB,
		"disk_gb":   req.DiskGB,
	}).Info("Creating VM")

	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check resource availability
	if err := s.checkResourceAvailability(ctx, req); err != nil {
		return nil, fmt.Errorf("insufficient resources: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check VM limit
	if len(s.vms) >= s.maxVMs {
		return nil, fmt.Errorf("maximum VM limit (%d) reached", s.maxVMs)
	}

	// Create VM
	now := time.Now()
	vm := &VM{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Status:    VMStatusStopped,
		CPUCores:  req.CPUCores,
		MemoryMB:  req.MemoryMB,
		DiskGB:    req.DiskGB,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  req.Metadata,
	}

	// In a real implementation, this would create actual VM resources
	// For now, we just simulate the creation
	vm.IPAddress = s.generateMockIP()

	// Store VM
	s.vms[vm.ID] = vm

	// Save to datastore
	if err := s.saveVMToDatastore(ctx, vm); err != nil {
		delete(s.vms, vm.ID)
		return nil, fmt.Errorf("failed to save VM: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"vm_id": vm.ID,
		"name":  vm.Name,
	}).Info("VM created successfully")

	return vm, nil
}

// GetVM retrieves a VM by ID
func (s *Service) GetVM(ctx context.Context, vmID string) (*VM, error) {
	if !s.enabled {
		return nil, fmt.Errorf("KVM service is not enabled")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	vm, exists := s.vms[vmID]
	if !exists {
		return nil, fmt.Errorf("VM not found: %s", vmID)
	}

	return vm, nil
}

// ListVMs returns all VMs
func (s *Service) ListVMs(ctx context.Context) ([]*VM, error) {
	if !s.enabled {
		return nil, fmt.Errorf("KVM service is not enabled")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	vms := make([]*VM, 0, len(s.vms))
	for _, vm := range s.vms {
		vms = append(vms, vm)
	}

	return vms, nil
}

// StartVM starts a virtual machine
func (s *Service) StartVM(ctx context.Context, vmID string) error {
	if !s.enabled {
		return fmt.Errorf("KVM service is not enabled")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	vm, exists := s.vms[vmID]
	if !exists {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	if vm.Status == VMStatusRunning {
		return fmt.Errorf("VM is already running")
	}

	s.logger.WithField("vm_id", vmID).Info("Starting VM")

	// Simulate VM startup
	vm.Status = VMStatusStarting
	vm.UpdatedAt = time.Now()

	// In a real implementation, this would start the actual VM
	go func() {
		time.Sleep(2 * time.Second) // Simulate startup time

		s.mu.Lock()
		vm.Status = VMStatusRunning
		vm.UpdatedAt = time.Now()
		s.mu.Unlock()

		s.saveVMToDatastore(context.Background(), vm)
		s.logger.WithField("vm_id", vmID).Info("VM started successfully")
	}()

	return s.saveVMToDatastore(ctx, vm)
}

// StopVM stops a virtual machine
func (s *Service) StopVM(ctx context.Context, vmID string) error {
	if !s.enabled {
		return fmt.Errorf("KVM service is not enabled")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	vm, exists := s.vms[vmID]
	if !exists {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	if vm.Status == VMStatusStopped {
		return fmt.Errorf("VM is already stopped")
	}

	s.logger.WithField("vm_id", vmID).Info("Stopping VM")

	// Simulate VM shutdown
	vm.Status = VMStatusStopping
	vm.UpdatedAt = time.Now()

	// In a real implementation, this would stop the actual VM
	go func() {
		time.Sleep(1 * time.Second) // Simulate shutdown time

		s.mu.Lock()
		vm.Status = VMStatusStopped
		vm.UpdatedAt = time.Now()
		s.mu.Unlock()

		s.saveVMToDatastore(context.Background(), vm)
		s.logger.WithField("vm_id", vmID).Info("VM stopped successfully")
	}()

	return s.saveVMToDatastore(ctx, vm)
}

// DeleteVM deletes a virtual machine
func (s *Service) DeleteVM(ctx context.Context, vmID string) error {
	if !s.enabled {
		return fmt.Errorf("KVM service is not enabled")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	vm, exists := s.vms[vmID]
	if !exists {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	if vm.Status == VMStatusRunning {
		return fmt.Errorf("cannot delete running VM, stop it first")
	}

	s.logger.WithField("vm_id", vmID).Info("Deleting VM")

	// Remove from memory
	delete(s.vms, vmID)
	delete(s.vmStats, vmID)

	// Remove from datastore
	key := datastore.NewKey("/kvm/vms/" + vmID)
	if err := s.datastore.Delete(ctx, key); err != nil {
		s.logger.WithError(err).Error("Failed to delete VM from datastore")
		return fmt.Errorf("failed to delete VM from datastore: %w", err)
	}

	s.logger.WithField("vm_id", vmID).Info("VM deleted successfully")
	return nil
}

// GetSystemResources returns current system resource usage
func (s *Service) GetSystemResources(ctx context.Context) (*SystemResources, error) {
	if !s.enabled {
		return nil, fmt.Errorf("KVM service is not enabled")
	}

	// Get system resources from resource service
	sysRes, err := s.resourceSvc.GetResource()
	if err != nil {
		return nil, fmt.Errorf("failed to get system resources: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate used resources from VMs
	usedCPU := 0
	usedMemoryMB := 0
	usedDiskGB := 0
	runningVMs := 0

	for _, vm := range s.vms {
		usedCPU += vm.CPUCores
		usedMemoryMB += vm.MemoryMB
		usedDiskGB += vm.DiskGB
		if vm.Status == VMStatusRunning {
			runningVMs++
		}
	}

	totalCPU := sysRes.CPU.Count
	totalMemoryMB := int(sysRes.Memory.Total / (1024 * 1024))
	totalDiskGB := int(sysRes.Storage.Total / (1024 * 1024 * 1024))

	return &SystemResources{
		TotalCPUCores:     totalCPU,
		AvailableCPUCores: totalCPU - usedCPU,
		TotalMemoryMB:     totalMemoryMB,
		AvailableMemoryMB: totalMemoryMB - usedMemoryMB,
		TotalDiskGB:       totalDiskGB,
		AvailableDiskGB:   totalDiskGB - usedDiskGB,
		RunningVMs:        runningVMs,
		MaxVMs:            s.maxVMs,
	}, nil
}

// GetVMStats returns statistics for a specific VM
func (s *Service) GetVMStats(ctx context.Context, vmID string) (*VMStats, error) {
	if !s.enabled {
		return nil, fmt.Errorf("KVM service is not enabled")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.vms[vmID]; !exists {
		return nil, fmt.Errorf("VM not found: %s", vmID)
	}

	stats, exists := s.vmStats[vmID]
	if !exists {
		// Return default stats if not collected yet
		return &VMStats{
			VMID:        vmID,
			CPUUsage:    0,
			MemoryUsage: 0,
			DiskUsage:   0,
			NetworkRxMB: 0,
			NetworkTxMB: 0,
			CollectedAt: time.Now(),
		}, nil
	}

	return stats, nil
}

// Helper methods

// validateCreateRequest validates the VM creation request
func (s *Service) validateCreateRequest(req *CreateVMRequest) error {
	if req.Name == "" {
		return fmt.Errorf("VM name is required")
	}

	if req.CPUCores < 1 || req.CPUCores > s.maxCPUCores {
		return fmt.Errorf("CPU cores must be between 1 and %d", s.maxCPUCores)
	}

	if req.MemoryMB < 512 || req.MemoryMB > s.maxMemoryMB {
		return fmt.Errorf("memory must be between 512 MB and %d MB", s.maxMemoryMB)
	}

	if req.DiskGB < 10 || req.DiskGB > s.maxDiskGB {
		return fmt.Errorf("disk size must be between 10 GB and %d GB", s.maxDiskGB)
	}

	// Check for duplicate names
	s.mu.RLock()
	for _, vm := range s.vms {
		if vm.Name == req.Name {
			s.mu.RUnlock()
			return fmt.Errorf("VM with name '%s' already exists", req.Name)
		}
	}
	s.mu.RUnlock()

	return nil
}

// checkResourceAvailability checks if system has enough resources
func (s *Service) checkResourceAvailability(ctx context.Context, req *CreateVMRequest) error {
	sysRes, err := s.resourceSvc.GetResource()
	if err != nil {
		return fmt.Errorf("failed to get system resources: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate currently used resources
	usedCPU := 0
	usedMemoryMB := 0
	usedDiskGB := 0

	for _, vm := range s.vms {
		usedCPU += vm.CPUCores
		usedMemoryMB += vm.MemoryMB
		usedDiskGB += vm.DiskGB
	}

	// Check CPU availability
	totalCPU := sysRes.CPU.Count
	if usedCPU+req.CPUCores > totalCPU {
		return fmt.Errorf("insufficient CPU cores: need %d, have %d available",
			req.CPUCores, totalCPU-usedCPU)
	}

	// Check memory availability
	totalMemoryMB := int(sysRes.Memory.Total / (1024 * 1024))
	if usedMemoryMB+req.MemoryMB > totalMemoryMB {
		return fmt.Errorf("insufficient memory: need %d MB, have %d MB available",
			req.MemoryMB, totalMemoryMB-usedMemoryMB)
	}

	// Check disk availability
	totalDiskGB := int(sysRes.Storage.Total / (1024 * 1024 * 1024))
	if usedDiskGB+req.DiskGB > totalDiskGB {
		return fmt.Errorf("insufficient disk space: need %d GB, have %d GB available",
			req.DiskGB, totalDiskGB-usedDiskGB)
	}

	return nil
}

// generateMockIP generates a mock IP address for simulation
func (s *Service) generateMockIP() string {
	// Simple mock IP generation - in real implementation this would be proper DHCP/network management
	return fmt.Sprintf("192.168.122.%d", 10+len(s.vms))
}

// saveVMToDatastore saves VM metadata to datastore
func (s *Service) saveVMToDatastore(ctx context.Context, vm *VM) error {
	key := datastore.NewKey("/kvm/vms/" + vm.ID)

	// Convert VM to JSON-like data
	data := map[string]interface{}{
		"id":         vm.ID,
		"name":       vm.Name,
		"status":     string(vm.Status),
		"cpu_cores":  vm.CPUCores,
		"memory_mb":  vm.MemoryMB,
		"disk_gb":    vm.DiskGB,
		"ip_address": vm.IPAddress,
		"created_at": vm.CreatedAt.Unix(),
		"updated_at": vm.UpdatedAt.Unix(),
		"metadata":   vm.Metadata,
	}

	// In a real implementation, this would serialize to JSON/protobuf
	// For simplicity, we'll just store the string representation
	value := fmt.Sprintf("%+v", data)

	return s.datastore.Put(ctx, key, []byte(value))
}

// loadVMsFromDatastore loads existing VMs from datastore
func (s *Service) loadVMsFromDatastore(ctx context.Context) error {
	q := query.Query{Prefix: "/kvm/vms/"}
	results, err := s.datastore.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("failed to query VMs: %w", err)
	}
	defer results.Close()

	count := 0
	for result := range results.Next() {
		if result.Error != nil {
			s.logger.WithError(result.Error).Error("Error loading VM from datastore")
			continue
		}

		// For simplicity, we'll just log that we found VMs
		// In a real implementation, this would deserialize and restore VM state
		vmID := result.Key[len("/kvm/vms/"):]
		s.logger.WithField("vm_id", vmID).Debug("Found VM in datastore")
		count++
	}

	s.logger.WithField("vm_count", count).Info("Loaded VMs from datastore")
	return nil
}

// monitorVMs runs background monitoring for VM statistics
func (s *Service) monitorVMs(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.collectVMStats()
		}
	}
}

// collectVMStats collects statistics for all running VMs
func (s *Service) collectVMStats() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for vmID, vm := range s.vms {
		if vm.Status == VMStatusRunning {
			// Simulate statistics collection
			stats := &VMStats{
				VMID:        vmID,
				CPUUsage:    float64(20 + (len(vmID) % 60)), // Mock CPU usage 20-80%
				MemoryUsage: float64(30 + (len(vmID) % 50)), // Mock memory usage 30-80%
				DiskUsage:   float64(10 + (len(vmID) % 30)), // Mock disk usage 10-40%
				NetworkRxMB: float64(len(vmID) % 100),       // Mock network usage
				NetworkTxMB: float64(len(vmID) % 80),
				CollectedAt: now,
			}
			s.vmStats[vmID] = stats
		}
	}

	if len(s.vms) > 0 {
		s.logger.WithField("running_vms", len(s.vmStats)).Debug("Collected VM statistics")
	}
}
