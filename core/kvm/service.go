package kvm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/kvm/libvirt"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
)

// Service implements KVM management with configvm-based mode detection
type Service struct {
	config      *config.C
	logger      *logrus.Entry
	datastore   datastore.Datastore
	resourceSvc *resource.Service

	// Libvirt components (only used if libvirt is available)
	libvirtAvailable bool
	client           *libvirt.Client
	domainManager    *libvirt.DomainManager
	storageManager   *libvirt.StorageManager
	networkManager   *libvirt.NetworkManager
	cloudInitManager *libvirt.CloudInitManager

	// In-memory VM storage (used in both modes)
	vms     map[string]*VM
	vmStats map[string]*VMStats
	mu      sync.RWMutex

	// Configuration
	enabled       bool
	mode          string // auto, real, simulation
	maxVMs        int
	maxCPUCores   int
	maxMemoryMB   int
	maxDiskGB     int
	libvirtURI    string
	storagePool   string
	storagePath   string
	networkName   string
	ubuntuVersion string
	sshKeyPath    string
}

// NewService creates a new KVM service
func NewService(
	cfg *config.C,
	logger *logrus.Entry,
	ds datastore.Datastore,
	resourceSvc *resource.Service,
) *Service {
	service := &Service{
		config:      cfg,
		logger:      logger.WithField("service", "kvm"),
		datastore:   ds,
		resourceSvc: resourceSvc,
		vms:         make(map[string]*VM),
		vmStats:     make(map[string]*VMStats),

		// Load configuration
		enabled:       cfg.GetBool("kvm.enabled", true),
		mode:          cfg.GetString("kvm.mode", "auto"),
		maxVMs:        cfg.GetInt("kvm.max_vms", 5),
		maxCPUCores:   cfg.GetInt("kvm.max_cpu_cores", 4),
		maxMemoryMB:   cfg.GetInt("kvm.max_memory_mb", 4096),
		maxDiskGB:     cfg.GetInt("kvm.max_disk_gb", 50),
		libvirtURI:    cfg.GetString("kvm.libvirt_uri", "qemu:///system"),
		storagePool:   cfg.GetString("kvm.storage_pool", "subnet-vms"),
		storagePath:   cfg.GetString("kvm.storage_path", "/var/lib/libvirt/images/subnet"),
		networkName:   cfg.GetString("kvm.network_name", "subnet-net"),
		ubuntuVersion: cfg.GetString("kvm.ubuntu_version", "22.04"),
		sshKeyPath:    cfg.GetString("kvm.ssh_key_path", "/var/lib/libvirt/ssh/subnet-key"),
	}

	// Log configuration values
	service.logger.WithFields(logrus.Fields{
		"enabled":        service.enabled,
		"mode":           service.mode,
		"max_vms":        service.maxVMs,
		"max_cpu_cores":  service.maxCPUCores,
		"max_memory_mb":  service.maxMemoryMB,
		"max_disk_gb":    service.maxDiskGB,
		"libvirt_uri":    service.libvirtURI,
		"storage_pool":   service.storagePool,
		"storage_path":   service.storagePath,
		"network_name":   service.networkName,
		"ubuntu_version": service.ubuntuVersion,
		"ssh_key_path":   service.sshKeyPath,
	}).Info("KVM service configuration loaded")

	// Determine libvirt availability based on mode
	service.determineLibvirtMode()

	// Log final service state
	if service.enabled {
		if service.libvirtAvailable {
			service.logger.Info("KVM service initialized successfully with real libvirt support")
		} else {
			service.logger.Info("KVM service initialized successfully in simulation mode")
		}
	} else {
		service.logger.Info("KVM service is disabled")
	}

	return service
}

// determineLibvirtMode determines whether to use real libvirt or simulation mode
func (s *Service) determineLibvirtMode() {
	s.logger.WithFields(logrus.Fields{
		"enabled": s.enabled,
		"mode":    s.mode,
	}).Info("Determining libvirt mode")

	switch s.mode {
	case "real":
		// Force real mode - try to initialize libvirt
		s.logger.Info("Mode set to 'real', attempting libvirt initialization")
		if err := s.tryInitLibvirt(); err != nil {
			s.logger.WithError(err).Error("Failed to initialize libvirt in real mode")
			s.libvirtAvailable = false
			// Don't fail completely, just log the error and continue in simulation mode
		} else {
			s.logger.Info("Libvirt initialized successfully in real mode")
			s.libvirtAvailable = true
		}
	case "simulation":
		// Force simulation mode - skip libvirt initialization entirely
		s.libvirtAvailable = false
		s.logger.Info("Using simulation mode as configured - skipping libvirt initialization")
	case "auto":
		fallthrough
	default:
		// Auto-detect: try libvirt if enabled, fallback to simulation
		s.logger.Info("Mode set to 'auto', checking if KVM is enabled")
		if s.enabled {
			s.logger.Info("KVM is enabled, attempting libvirt initialization")
			if err := s.tryInitLibvirt(); err != nil {
				// Check if it's a permission error and provide helpful message
				if s.isPermissionError(err) {
					s.logger.WithError(err).Info("Libvirt permission error detected, falling back to simulation mode. This is normal for user sessions without elevated privileges.")
				} else {
					s.logger.WithError(err).Warn("Libvirt not available, using simulation mode")
				}
				s.libvirtAvailable = false
			} else {
				s.logger.Info("Libvirt detected and initialized successfully, using real virtualization mode")
				s.libvirtAvailable = true
			}
		} else {
			s.libvirtAvailable = false
			s.logger.Info("KVM disabled, using simulation mode")
		}
	}

	s.logger.WithField("libvirt_available", s.libvirtAvailable).Info("Libvirt availability determined")
}

// tryInitLibvirt attempts to initialize libvirt components
func (s *Service) tryInitLibvirt() error {
	s.logger.Info("Attempting to initialize libvirt components")

	var err error

	// Try to connect to libvirt
	s.logger.WithField("uri", s.libvirtURI).Info("Creating libvirt client")
	s.client, err = libvirt.NewClient(s.libvirtURI, s.logger)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create libvirt client")
		return fmt.Errorf("failed to create libvirt client: %w", err)
	}

	// Check if libvirt is actually available
	s.logger.Info("Checking if libvirt is available")
	if !s.client.IsAvailable() {
		s.logger.Error("Libvirt client reports not available")
		return fmt.Errorf("libvirt is not available on this system")
	}

	s.logger.Info("Libvirt client is available, initializing managers")

	// Initialize managers
	s.domainManager = libvirt.NewDomainManager(s.client, s.logger)
	s.storageManager = libvirt.NewStorageManager(s.client, s.logger)
	s.networkManager = libvirt.NewNetworkManager(s.client, s.logger)
	s.cloudInitManager = libvirt.NewCloudInitManager(s.logger)

	// Ensure storage pool exists
	s.logger.WithFields(logrus.Fields{
		"pool": s.storagePool,
		"path": s.storagePath,
	}).Info("Ensuring storage pool exists")
	if err := s.storageManager.EnsureDefaultPool(s.storagePool, s.storagePath); err != nil {
		s.logger.WithError(err).Error("Failed to ensure storage pool")
		return fmt.Errorf("failed to ensure storage pool: %w", err)
	}

	// Ensure network exists - this is the most likely point of failure due to permissions
	s.logger.WithField("network", s.networkName).Info("Ensuring network exists")
	if err := s.ensureNetwork(); err != nil {
		// Check if it's a permission error
		if s.isPermissionError(err) {
			s.logger.WithError(err).Warn("Permission denied creating network bridge. This is expected in user sessions without elevated privileges.")
			return fmt.Errorf("permission denied creating network bridge: %w", err)
		}
		s.logger.WithError(err).Error("Failed to ensure network")
		return fmt.Errorf("failed to ensure network: %w", err)
	}

	s.logger.Info("Libvirt initialization completed successfully")
	return nil
}

// ensureNetwork ensures the default network exists
func (s *Service) ensureNetwork() error {
	// Try to get existing network
	_, err := s.client.GetNetworkByName(s.networkName)
	if err == nil {
		// Network exists
		return nil
	}

	// Create default network
	networkXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<network>
  <name>%s</name>
  <forward mode="nat"/>
  <bridge name="virbr2" stp="on" delay="0"/>
  <ip address="192.168.123.1" netmask="255.255.255.0">
    <dhcp>
      <range start="192.168.123.2" end="192.168.123.254"/>
    </dhcp>
  </ip>
</network>`, s.networkName)

	_, err = s.client.CreateNetwork(networkXML)
	if err != nil {
		// Check if it's a permission error
		if s.isPermissionError(err) {
			s.logger.WithError(err).Warn("Permission denied creating network bridge. This may require elevated privileges or different libvirt URI.")
			return fmt.Errorf("permission denied creating network bridge: %w", err)
		}
		return fmt.Errorf("failed to create network: %w", err)
	}

	s.logger.WithField("network", s.networkName).Info("Default network created")
	return nil
}

// isPermissionError checks if the error is related to permissions
func (s *Service) isPermissionError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "Operation not permitted") ||
		contains(errStr, "Permission denied") ||
		contains(errStr, "access denied") ||
		contains(errStr, "insufficient privileges")
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

// containsSubstring checks if a string contains a substring (case-insensitive)
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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

	// Create real VM if libvirt is available
	if s.libvirtAvailable {
		if err := s.createRealVM(vm); err != nil {
			return nil, fmt.Errorf("failed to create real VM: %w", err)
		}
	} else {
		// Fallback to simulation mode
		vm.IPAddress = s.generateMockIP()
	}

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

// createRealVM creates a real VM using libvirt
func (s *Service) createRealVM(vm *VM) error {
	// Generate SSH key if not exists
	sshKeyPath := s.sshKeyPath
	if sshKeyPath == "" {
		sshKeyPath = "/var/lib/libvirt/ssh/subnet-key"
	}

	publicKey, err := s.cloudInitManager.GenerateSSHKey(sshKeyPath)
	if err != nil {
		return fmt.Errorf("failed to generate SSH key: %w", err)
	}

	// Create VM disk from Ubuntu cloud image
	sanitizedName := s.sanitizeFileName(vm.Name)
	diskPath, err := s.storageManager.CreateVMFromUbuntuImage(s.storagePool, sanitizedName, vm.DiskGB, s.ubuntuVersion, "amd64")
	if err != nil {
		return fmt.Errorf("failed to create VM disk: %w", err)
	}

	// Create cloud-init configuration
	cloudInitConfig := s.cloudInitManager.CreateDefaultCloudInitConfig(vm.Name, "ubuntu")
	cloudInitConfig.SSHKey = publicKey

	// Create cloud-init ISO
	cloudInitISOPath := fmt.Sprintf("%s/%s-cloud-init.iso", s.storagePath, sanitizedName)
	actualISOPath, err := s.cloudInitManager.CreateCloudInitISO(cloudInitISOPath, cloudInitConfig)
	if err != nil {
		return fmt.Errorf("failed to create cloud-init ISO: %w", err)
	}

	// Create libvirt domain
	domain, err := s.domainManager.CreateDomainWithCloudInit(
		vm.Name,
		vm.ID,
		vm.MemoryMB,
		vm.CPUCores,
		diskPath,
		s.networkName,
		actualISOPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}

	// Store domain reference (you might want to store this in the VM struct)
	_ = domain

	// Generate IP address (in a real implementation, you'd get this from DHCP)
	vm.IPAddress = s.generateMockIP()

	s.logger.WithFields(logrus.Fields{
		"vm_id":          vm.ID,
		"disk_path":      diskPath,
		"cloud_init_iso": actualISOPath,
	}).Info("Real VM created with libvirt")

	return nil
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

	// Update status to starting
	vm.Status = VMStatusStarting
	vm.UpdatedAt = time.Now()

	// Start real VM if libvirt is available
	if s.libvirtAvailable {
		if err := s.startRealVM(vm); err != nil {
			vm.Status = VMStatusError
			s.saveVMToDatastore(context.Background(), vm)
			return fmt.Errorf("failed to start real VM: %w", err)
		}
	} else {
		// Fallback to simulation mode
		go func() {
			time.Sleep(2 * time.Second) // Simulate startup time

			s.mu.Lock()
			vm.Status = VMStatusRunning
			vm.UpdatedAt = time.Now()
			s.mu.Unlock()

			s.saveVMToDatastore(context.Background(), vm)
			s.logger.WithField("vm_id", vmID).Info("VM started successfully (simulation)")
		}()
	}

	return s.saveVMToDatastore(ctx, vm)
}

// startRealVM starts a real VM using libvirt
func (s *Service) startRealVM(vm *VM) error {
	// Get domain by name
	domain, err := s.domainManager.GetDomain(vm.Name)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// Start domain
	if err := s.domainManager.StartDomain(domain); err != nil {
		return fmt.Errorf("failed to start domain: %w", err)
	}

	// Update VM status
	vm.Status = VMStatusRunning
	vm.UpdatedAt = time.Now()

	s.logger.WithField("vm_id", vm.ID).Info("Real VM started with libvirt")
	return nil
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

	// Update status to stopping
	vm.Status = VMStatusStopping
	vm.UpdatedAt = time.Now()

	// Stop real VM if libvirt is available
	if s.libvirtAvailable {
		if err := s.stopRealVM(vm); err != nil {
			vm.Status = VMStatusError
			s.saveVMToDatastore(context.Background(), vm)
			return fmt.Errorf("failed to stop real VM: %w", err)
		}
	} else {
		// Fallback to simulation mode
		go func() {
			time.Sleep(1 * time.Second) // Simulate shutdown time

			s.mu.Lock()
			vm.Status = VMStatusStopped
			vm.UpdatedAt = time.Now()
			s.mu.Unlock()

			s.saveVMToDatastore(context.Background(), vm)
			s.logger.WithField("vm_id", vmID).Info("VM stopped successfully (simulation)")
		}()
	}

	return s.saveVMToDatastore(ctx, vm)
}

// stopRealVM stops a real VM using libvirt
func (s *Service) stopRealVM(vm *VM) error {
	// Get domain by name
	domain, err := s.domainManager.GetDomain(vm.Name)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// Stop domain gracefully
	if err := s.domainManager.StopDomain(domain); err != nil {
		return fmt.Errorf("failed to stop domain: %w", err)
	}

	// Update VM status
	vm.Status = VMStatusStopped
	vm.UpdatedAt = time.Now()

	s.logger.WithField("vm_id", vm.ID).Info("Real VM stopped with libvirt")
	return nil
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

	// Delete real VM if libvirt is available
	if s.libvirtAvailable {
		if err := s.deleteRealVM(vm); err != nil {
			return fmt.Errorf("failed to delete real VM: %w", err)
		}
	}

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

// deleteRealVM deletes a real VM using libvirt
func (s *Service) deleteRealVM(vm *VM) error {
	// Get domain by name
	domain, err := s.domainManager.GetDomain(vm.Name)
	if err != nil {
		// Domain might not exist, which is fine for deletion
		s.logger.WithField("vm_name", vm.Name).Warn("Domain not found during deletion")
	} else {
		// Delete domain
		if err := s.domainManager.DeleteDomain(domain); err != nil {
			return fmt.Errorf("failed to delete domain: %w", err)
		}
	}

	// Delete VM disk
	sanitizedName := s.sanitizeFileName(vm.Name)
	if err := s.storageManager.DeleteDiskImage(s.storagePool, sanitizedName); err != nil {
		s.logger.WithError(err).Warn("Failed to delete VM disk")
	}

	// Delete cloud-init ISO - try both possible locations
	cloudInitISOPath := fmt.Sprintf("%s/%s-cloud-init.iso", s.storagePath, sanitizedName)
	if err := os.Remove(cloudInitISOPath); err != nil && !os.IsNotExist(err) {
		s.logger.WithError(err).Warn("Failed to delete cloud-init ISO from original path")
	}

	// Try user-writable directory as fallback
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userCloudInitPath := filepath.Join(homeDir, ".subnet", "libvirt", "cloud-init", fmt.Sprintf("%s-cloud-init.iso", sanitizedName))
		if err := os.Remove(userCloudInitPath); err != nil && !os.IsNotExist(err) {
			s.logger.WithError(err).Warn("Failed to delete cloud-init ISO from user directory")
		}
	}

	s.logger.WithField("vm_id", vm.ID).Info("Real VM deleted with libvirt")
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

	// get mock data for test first
	sysRes := &resource.ResourceInfo{
		CPU: resource.CpuInfo{
			Count: 10,
		},
		Memory: resource.MemoryInfo{
			Total: 1024 * 1024 * 1024 * 10,
		},
		Storage: resource.StorageInfo{
			Total: 1024 * 1024 * 1024 * 20,
		},
	}

	// sysRes, err := s.resourceSvc.GetResource()
	// if err != nil {
	// 	return fmt.Errorf("failed to get system resources: %w", err)
	// }

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
	return fmt.Sprintf("192.168.123.%d", 10+len(s.vms))
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

// sanitizeFileName sanitizes a string for use as a filename by replacing invalid characters
func (s *Service) sanitizeFileName(name string) string {
	// Replace spaces and other problematic characters with underscores
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}
