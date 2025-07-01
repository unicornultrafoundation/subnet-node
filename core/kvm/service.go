package kvm

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
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

	// User context for consistent file operations
	userContext *libvirt.UserContext
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

		// Initialize user context
		userContext: libvirt.NewUserContext(logger),
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

	// Ensure consistent user context for all operations
	if err := s.ensureConsistentUserContext(); err != nil {
		return nil, fmt.Errorf("failed to ensure consistent user context: %w", err)
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
	diskPath, err := s.storageManager.CreateVMFromUbuntuImage(s.storagePool, sanitizedName, vm.DiskGB, s.ubuntuVersion, "arm64")
	if err != nil {
		return fmt.Errorf("failed to create VM disk: %w", err)
	}

	// Fix VM disk permissions after creation to ensure libvirt can access it
	if err := s.storageManager.FixVMFilePermissions(sanitizedName); err != nil {
		s.logger.WithError(err).Warn("Failed to fix VM disk permissions after creation")
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

	// Ensure cloud-init ISO has correct permissions
	if err := s.userContext.EnsureFileOwnership(actualISOPath); err != nil {
		s.logger.WithError(err).Warn("Failed to ensure cloud-init ISO permissions, but continuing")
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

	// Perform comprehensive permission check and fix for all VM files
	// This ensures the VM can be started successfully
	if err := s.ensureVMStartupPermissions(vm.Name, diskPath, actualISOPath); err != nil {
		s.logger.WithError(err).Warn("Failed to ensure VM startup permissions, but continuing")
		// Don't fail the entire operation, but log the warning
	}

	// Get real IP address from DHCP leases
	vm.IPAddress, err = s.getRealIPAddress(vm.Name)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get real IP address, will retry later")
		// Don't fail the VM creation, the IP will be retrieved when the VM starts
	}

	s.logger.WithFields(logrus.Fields{
		"vm_id":          vm.ID,
		"disk_path":      diskPath,
		"cloud_init_iso": actualISOPath,
		"ip_address":     vm.IPAddress,
	}).Info("Real VM created with libvirt")

	return nil
}

// ensureVMStartupPermissions ensures that all VM files have correct permissions for startup
func (s *Service) ensureVMStartupPermissions(vmName, diskPath, cloudInitISOPath string) error {
	s.logger.WithField("vm_name", vmName).Info("Ensuring VM startup permissions")

	// Ensure VM disk permissions
	if err := s.storageManager.FixVMFilePermissions(vmName); err != nil {
		s.logger.WithError(err).Warn("Failed to fix VM disk permissions")
	}

	// Ensure cloud-init ISO permissions
	if err := s.userContext.EnsureFileOwnership(cloudInitISOPath); err != nil {
		s.logger.WithError(err).Warn("Failed to ensure cloud-init ISO permissions")
	}

	// Ensure storage directory permissions
	// Use the storage path from service configuration
	if s.storagePath != "" {
		if err := s.userContext.EnsureDirectoryPermissions(s.storagePath); err != nil {
			s.logger.WithError(err).Warn("Failed to ensure storage directory permissions")
		}
	}

	// Try to fix domain permissions as well
	if err := s.domainManager.FixDomainPermissions(vmName); err != nil {
		s.logger.WithError(err).Warn("Failed to fix domain permissions")
	}

	s.logger.WithField("vm_name", vmName).Info("VM startup permissions check completed")
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

	// Ensure consistent user context for all operations
	if err := s.ensureConsistentUserContext(); err != nil {
		return fmt.Errorf("failed to ensure consistent user context: %w", err)
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
	s.logger.WithField("vm_name", vm.Name).Info("Starting real VM with libvirt")

	// Perform comprehensive permission check and fix before starting
	sanitizedName := s.sanitizeFileName(vm.Name)

	// Fix VM disk permissions before starting
	if err := s.storageManager.FixVMFilePermissions(sanitizedName); err != nil {
		s.logger.WithError(err).Warn("Failed to fix VM disk permissions, attempting to start anyway")
	}

	// Get domain by name
	domain, err := s.domainManager.GetDomain(vm.Name)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get domain")
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// Try to fix domain permissions before starting
	if err := s.domainManager.FixDomainPermissions(domain); err != nil {
		s.logger.WithError(err).Warn("Failed to fix domain permissions, attempting to start anyway")
	}

	// Start domain with retry logic
	var startErr error
	for attempt := 1; attempt <= 3; attempt++ {
		s.logger.WithFields(logrus.Fields{
			"vm_name": vm.Name,
			"attempt": attempt,
		}).Info("Attempting to start domain")

		startErr = s.domainManager.StartDomain(domain)
		if startErr == nil {
			s.logger.WithField("vm_name", vm.Name).Info("Domain started successfully")
			break
		}

		s.logger.WithError(startErr).WithFields(logrus.Fields{
			"vm_name": vm.Name,
			"attempt": attempt,
		}).Warn("Failed to start domain, attempting permission fix")

		// If start failed, try to fix permissions and retry
		if attempt < 3 {
			// Perform aggressive permission fix
			if err := s.performAggressivePermissionFix(vm.Name); err != nil {
				s.logger.WithError(err).Warn("Failed to perform aggressive permission fix")
			}

			// Wait a bit before retry
			time.Sleep(2 * time.Second)
		}
	}

	if startErr != nil {
		s.logger.WithError(startErr).Error("Failed to start domain after all attempts")
		return fmt.Errorf("failed to start domain after 3 attempts: %w", startErr)
	}

	// Update VM status
	vm.Status = VMStatusRunning
	vm.UpdatedAt = time.Now()

	// Try to get real IP address after starting
	go func() {
		// Wait a bit for the VM to boot and get an IP address
		time.Sleep(10 * time.Second)

		s.mu.Lock()
		defer s.mu.Unlock()

		if realIP, err := s.getRealIPAddress(vm.Name); err == nil && realIP != "" {
			vm.IPAddress = realIP
			vm.UpdatedAt = time.Now()
			s.logger.WithFields(logrus.Fields{
				"vm_id":      vm.ID,
				"ip_address": realIP,
			}).Info("Updated VM with real IP address")

			// Save updated VM to datastore
			s.saveVMToDatastore(context.Background(), vm)
		} else {
			s.logger.WithError(err).WithField("vm_id", vm.ID).Warn("Failed to get real IP address after VM start")
		}
	}()

	s.logger.WithField("vm_id", vm.ID).Info("Real VM started with libvirt")
	return nil
}

// performAggressivePermissionFix performs aggressive permission fixing for VM files
func (s *Service) performAggressivePermissionFix(vmName string) error {
	s.logger.WithField("vm_name", vmName).Info("Performing aggressive permission fix")

	// Fix VM disk permissions
	if err := s.storageManager.FixVMFilePermissions(vmName); err != nil {
		s.logger.WithError(err).Warn("Failed to fix VM disk permissions in aggressive fix")
	}

	// Fix domain permissions
	if err := s.domainManager.FixDomainPermissions(vmName); err != nil {
		s.logger.WithError(err).Warn("Failed to fix domain permissions in aggressive fix")
	}

	// Ensure storage directory permissions
	if s.storagePath != "" {
		if err := s.userContext.EnsureDirectoryPermissions(s.storagePath); err != nil {
			s.logger.WithError(err).Warn("Failed to ensure storage directory permissions in aggressive fix")
		}
	}

	s.logger.WithField("vm_name", vmName).Info("Aggressive permission fix completed")
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

// GetSystemArchitecture returns the detected system architecture information
func (s *Service) GetSystemArchitecture(ctx context.Context) (map[string]interface{}, error) {
	if !s.enabled {
		return nil, fmt.Errorf("KVM service is not enabled")
	}

	// If libvirt is available, use the domain manager to get detailed info
	if s.libvirtAvailable && s.domainManager != nil {
		return s.domainManager.GetSystemArchitectureInfo(), nil
	}

	// Fallback to basic runtime information
	arch := "x86_64"
	machine := "pc-q35-2.12"

	// Use runtime.GOARCH for basic detection
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
		machine = "pc-q35-2.12"
	case "arm64":
		arch = "aarch64"
		machine = "virt-8.2"
	case "ppc64le":
		arch = "ppc64le"
		machine = "pseries"
	case "s390x":
		arch = "s390x"
		machine = "s390-ccw-virtio"
	}

	// Determine domain type based on libvirt availability
	domainType := "qemu"
	if s.libvirtAvailable && s.client != nil && s.client.IsKVMAvailable() {
		domainType = "kvm"
	}

	return map[string]interface{}{
		"architecture":      arch,
		"machine":           machine,
		"cpu_model":         "Unknown (libvirt not available)",
		"vendor":            "Unknown (libvirt not available)",
		"features":          []string{},
		"go_arch":           runtime.GOARCH,
		"go_os":             runtime.GOOS,
		"kvm_available":     s.libvirtAvailable && s.client != nil && s.client.IsKVMAvailable(),
		"domain_type":       domainType,
		"libvirt_available": s.libvirtAvailable,
		"mode":              s.mode,
	}, nil
}

// DiagnosePermissionIssues performs a comprehensive diagnosis of permission issues
func (s *Service) DiagnosePermissionIssues(ctx context.Context) (map[string]interface{}, error) {
	if !s.enabled {
		return nil, fmt.Errorf("KVM service is not enabled")
	}

	diagnosis := map[string]interface{}{
		"timestamp":             time.Now().Unix(),
		"service_enabled":       s.enabled,
		"libvirt_available":     s.libvirtAvailable,
		"mode":                  s.mode,
		"storage_path":          s.storagePath,
		"storage_pool":          s.storagePool,
		"network_name":          s.networkName,
		"ssh_key_path":          s.sshKeyPath,
		"permission_issues":     []string{},
		"recommendations":       []string{},
		"file_permissions":      map[string]interface{}{},
		"directory_permissions": map[string]interface{}{},
	}

	// Check storage directory permissions
	if s.storagePath != "" {
		if info, err := os.Stat(s.storagePath); err == nil {
			stat := info.Sys().(*syscall.Stat_t)
			owner, _ := user.LookupId(fmt.Sprintf("%d", stat.Uid))
			group, _ := user.LookupId(fmt.Sprintf("%d", stat.Gid))

			diagnosis["directory_permissions"].(map[string]interface{})[s.storagePath] = map[string]interface{}{
				"exists":     true,
				"owner":      owner.Username,
				"group":      group.Username,
				"mode":       fmt.Sprintf("%o", stat.Mode&0777),
				"readable":   info.Mode()&0400 != 0,
				"writable":   info.Mode()&0200 != 0,
				"executable": info.Mode()&0100 != 0,
			}
		} else {
			diagnosis["directory_permissions"].(map[string]interface{})[s.storagePath] = map[string]interface{}{
				"exists": false,
				"error":  err.Error(),
			}
			diagnosis["permission_issues"] = append(diagnosis["permission_issues"].([]string),
				fmt.Sprintf("Storage directory does not exist: %s", s.storagePath))
			diagnosis["recommendations"] = append(diagnosis["recommendations"].([]string),
				fmt.Sprintf("Create storage directory: mkdir -p %s", s.storagePath))
		}
	}

	// Check VM disk files
	s.mu.RLock()
	for vmID, vm := range s.vms {
		sanitizedName := s.sanitizeFileName(vm.Name)
		diskPath := filepath.Join(s.storagePath, fmt.Sprintf("%s.qcow2", sanitizedName))

		if info, err := os.Stat(diskPath); err == nil {
			stat := info.Sys().(*syscall.Stat_t)
			owner, _ := user.LookupId(fmt.Sprintf("%d", stat.Uid))
			group, _ := user.LookupId(fmt.Sprintf("%d", stat.Gid))

			diagnosis["file_permissions"].(map[string]interface{})[diskPath] = map[string]interface{}{
				"vm_id":    vmID,
				"vm_name":  vm.Name,
				"exists":   true,
				"owner":    owner.Username,
				"group":    group.Username,
				"mode":     fmt.Sprintf("%o", stat.Mode&0777),
				"readable": info.Mode()&0400 != 0,
				"writable": info.Mode()&0200 != 0,
				"size":     info.Size(),
			}

			// Check if libvirt can access the file
			if info.Mode()&0006 == 0 { // No group or world permissions
				diagnosis["permission_issues"] = append(diagnosis["permission_issues"].([]string),
					fmt.Sprintf("VM disk file not accessible by libvirt: %s", diskPath))
				diagnosis["recommendations"] = append(diagnosis["recommendations"].([]string),
					fmt.Sprintf("Fix VM disk permissions: sudo chmod 660 %s", diskPath))
			}
		} else {
			diagnosis["file_permissions"].(map[string]interface{})[diskPath] = map[string]interface{}{
				"vm_id":   vmID,
				"vm_name": vm.Name,
				"exists":  false,
				"error":   err.Error(),
			}
		}
	}
	s.mu.RUnlock()

	// Check SSH key permissions
	if s.sshKeyPath != "" {
		if info, err := os.Stat(s.sshKeyPath); err == nil {
			stat := info.Sys().(*syscall.Stat_t)
			owner, _ := user.LookupId(fmt.Sprintf("%d", stat.Uid))
			group, _ := user.LookupId(fmt.Sprintf("%d", stat.Gid))

			diagnosis["file_permissions"].(map[string]interface{})[s.sshKeyPath] = map[string]interface{}{
				"type":     "ssh_key",
				"exists":   true,
				"owner":    owner.Username,
				"group":    group.Username,
				"mode":     fmt.Sprintf("%o", stat.Mode&0777),
				"readable": info.Mode()&0400 != 0,
				"writable": info.Mode()&0200 != 0,
			}

			// SSH keys should be 600 (user read/write only)
			if stat.Mode&0777 != 0600 {
				diagnosis["permission_issues"] = append(diagnosis["permission_issues"].([]string),
					fmt.Sprintf("SSH key has incorrect permissions: %s", s.sshKeyPath))
				diagnosis["recommendations"] = append(diagnosis["recommendations"].([]string),
					fmt.Sprintf("Fix SSH key permissions: chmod 600 %s", s.sshKeyPath))
			}
		} else {
			diagnosis["file_permissions"].(map[string]interface{})[s.sshKeyPath] = map[string]interface{}{
				"type":   "ssh_key",
				"exists": false,
				"error":  err.Error(),
			}
		}
	}

	// Check libvirt daemon status
	if s.libvirtAvailable {
		cmd := exec.Command("systemctl", "is-active", "libvirtd")
		if err := cmd.Run(); err == nil {
			diagnosis["libvirt_daemon_status"] = "running"
		} else {
			diagnosis["libvirt_daemon_status"] = "not_running"
			diagnosis["permission_issues"] = append(diagnosis["permission_issues"].([]string),
				"Libvirt daemon is not running")
			diagnosis["recommendations"] = append(diagnosis["recommendations"].([]string),
				"Start libvirt daemon: sudo systemctl start libvirtd")
		}
	}

	// Check KVM device
	if _, err := os.Stat("/dev/kvm"); err == nil {
		if info, err := os.Stat("/dev/kvm"); err == nil {
			stat := info.Sys().(*syscall.Stat_t)
			diagnosis["kvm_device"] = map[string]interface{}{
				"exists":   true,
				"mode":     fmt.Sprintf("%o", stat.Mode&0777),
				"readable": info.Mode()&0400 != 0,
				"writable": info.Mode()&0200 != 0,
			}

			if info.Mode()&0666 == 0 {
				diagnosis["permission_issues"] = append(diagnosis["permission_issues"].([]string),
					"KVM device not accessible")
				diagnosis["recommendations"] = append(diagnosis["recommendations"].([]string),
					"Fix KVM device permissions: sudo chmod 666 /dev/kvm")
			}
		}
	} else {
		diagnosis["kvm_device"] = map[string]interface{}{
			"exists": false,
			"error":  err.Error(),
		}
		diagnosis["permission_issues"] = append(diagnosis["permission_issues"].([]string),
			"KVM device not found")
		diagnosis["recommendations"] = append(diagnosis["recommendations"].([]string),
			"Enable KVM in BIOS or install KVM module: sudo modprobe kvm")
	}

	// Check user groups
	if currentUser, err := user.Current(); err == nil {
		groups, _ := currentUser.GroupIds()
		groupNames := []string{}
		for _, gid := range groups {
			if group, err := user.LookupGroupId(gid); err == nil {
				groupNames = append(groupNames, group.Name)
			}
		}

		diagnosis["user_groups"] = groupNames

		hasLibvirt := false
		hasKvm := false
		for _, group := range groupNames {
			if group == "libvirt" {
				hasLibvirt = true
			}
			if group == "kvm" {
				hasKvm = true
			}
		}

		if !hasLibvirt {
			diagnosis["permission_issues"] = append(diagnosis["permission_issues"].([]string),
				"User not in libvirt group")
			diagnosis["recommendations"] = append(diagnosis["recommendations"].([]string),
				fmt.Sprintf("Add user to libvirt group: sudo usermod -a -G libvirt %s", currentUser.Username))
		}

		if !hasKvm {
			diagnosis["permission_issues"] = append(diagnosis["permission_issues"].([]string),
				"User not in kvm group")
			diagnosis["recommendations"] = append(diagnosis["recommendations"].([]string),
				fmt.Sprintf("Add user to kvm group: sudo usermod -a -G kvm %s", currentUser.Username))
		}
	}

	// Test libvirt connection
	if s.libvirtAvailable {
		cmd := exec.Command("virsh", "-c", s.libvirtURI, "list", "--all")
		if err := cmd.Run(); err == nil {
			diagnosis["libvirt_connection"] = "successful"
		} else {
			diagnosis["libvirt_connection"] = "failed"
			diagnosis["permission_issues"] = append(diagnosis["permission_issues"].([]string),
				"Libvirt connection failed")
			diagnosis["recommendations"] = append(diagnosis["recommendations"].([]string),
				"Check libvirt configuration and user permissions")
		}
	}

	return diagnosis, nil
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

// getRealIPAddress gets the real IP address from libvirt DHCP leases
func (s *Service) getRealIPAddress(vmName string) (string, error) {
	if !s.libvirtAvailable || s.networkManager == nil {
		// Fallback to mock IP if libvirt is not available
		return s.generateMockIP(), nil
	}

	// Try to get IP from DHCP leases
	ip, err := s.networkManager.GetVMIPAddress(s.networkName, vmName)
	if err != nil {
		s.logger.WithError(err).WithField("vm_name", vmName).Warn("Failed to get real IP address, using mock IP")
		return s.generateMockIP(), nil
	}

	return ip, nil
}

// GetSSHConnectionInfo returns SSH connection information for a VM
func (s *Service) GetSSHConnectionInfo(ctx context.Context, vmID string) (*SSHConnectionInfo, error) {
	if !s.enabled {
		return nil, fmt.Errorf("KVM service is not enabled")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	vm, exists := s.vms[vmID]
	if !exists {
		return nil, fmt.Errorf("VM not found: %s", vmID)
	}

	if vm.Status != VMStatusRunning {
		return nil, fmt.Errorf("VM is not running: %s", vmID)
	}

	// Get the real IP address
	ipAddress, err := s.getRealIPAddress(vm.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get IP address: %w", err)
	}

	// Determine SSH key path
	sshKeyPath := s.sshKeyPath
	if sshKeyPath == "" {
		sshKeyPath = "/var/lib/libvirt/ssh/subnet-key"
	}

	// Check if SSH key exists
	if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("SSH key not found: %s", sshKeyPath)
	}

	return &SSHConnectionInfo{
		VMID:       vmID,
		VMName:     vm.Name,
		IPAddress:  ipAddress,
		Port:       22,
		Username:   "ubuntu",
		SSHKeyPath: sshKeyPath,
		SSHCommand: fmt.Sprintf("ssh -i %s ubuntu@%s", sshKeyPath, ipAddress),
	}, nil
}

// WaitForVMReady waits for a VM to be ready (IP address assigned and SSH accessible)
func (s *Service) WaitForVMReady(ctx context.Context, vmID string, timeout time.Duration) error {
	if !s.enabled {
		return fmt.Errorf("KVM service is not enabled")
	}

	s.logger.WithField("vm_id", vmID).Info("Waiting for VM to be ready")

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for VM to be ready: %s", vmID)
			}

			// Check if VM is running
			s.mu.RLock()
			vm, exists := s.vms[vmID]
			if !exists {
				s.mu.RUnlock()
				return fmt.Errorf("VM not found: %s", vmID)
			}
			s.mu.RUnlock()

			if vm.Status != VMStatusRunning {
				continue
			}

			// Try to get real IP address
			ipAddress, err := s.getRealIPAddress(vm.Name)
			if err != nil {
				s.logger.WithError(err).WithField("vm_id", vmID).Debug("Still waiting for IP address")
				continue
			}

			// Test SSH connectivity
			if s.TestSSHConnectivity(ipAddress) {
				s.logger.WithFields(logrus.Fields{
					"vm_id":      vmID,
					"ip_address": ipAddress,
				}).Info("VM is ready")
				return nil
			}

			s.logger.WithField("vm_id", vmID).Debug("VM is running but SSH not yet accessible")
		}
	}
}

// TestSSHConnectivity tests if SSH is accessible on the VM
func (s *Service) TestSSHConnectivity(ipAddress string) bool {
	// Use a simple TCP connection test to port 22
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:22", ipAddress), 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
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

// GetUserContext returns the user context for consistent file operations
func (s *Service) GetUserContext() *libvirt.UserContext {
	return s.userContext
}

// ensureConsistentUserContext ensures that all file operations use the same user context
func (s *Service) ensureConsistentUserContext() error {
	// Log current user context for debugging
	if effectiveUser := s.userContext.GetEffectiveUser(); effectiveUser != nil {
		s.logger.WithField("effective_user", effectiveUser.Username).Debug("Using consistent user context")
	} else {
		s.logger.Warn("No effective user available for file operations")
	}
	return nil
}
