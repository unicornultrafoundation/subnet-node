package kvm

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
)

var log = logrus.WithField("service", "kvm")

// Service implements the main KVM service
type Service struct {
	cfg             *config.C
	kvmConfig       *KVMConfig
	resourceService *resource.Service
	datastore       datastore.Datastore

	// Components
	vmManager       VMManager
	storageManager  StorageManager
	networkManager  NetworkManager
	resourceChecker ResourceChecker
	registry        VMRegistry
	libvirtClient   LibvirtClient

	// State
	running  bool
	stopChan chan struct{}
	mu       sync.RWMutex
	logger   *logrus.Logger
}

// NewService creates a new KVM service instance
func NewService(cfg *config.C, resourceService *resource.Service, ds datastore.Datastore) (*Service, error) {
	logger := log.WithField("component", "kvm-service").Logger

	// Parse KVM configuration
	kvmConfig, err := parseKVMConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse KVM config: %w", err)
	}

	if !kvmConfig.Enabled {
		logger.Info("KVM service is disabled in configuration")
		return &Service{
			cfg:       cfg,
			kvmConfig: kvmConfig,
			running:   false,
			logger:    logger,
		}, nil
	}

	service := &Service{
		cfg:             cfg,
		kvmConfig:       kvmConfig,
		resourceService: resourceService,
		datastore:       ds,
		stopChan:        make(chan struct{}),
		logger:          logger,
	}

	// Initialize components
	if err := service.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize KVM components: %w", err)
	}

	return service, nil
}

// initializeComponents initializes all KVM service components
func (s *Service) initializeComponents() error {
	var err error

	// Initialize libvirt client
	s.libvirtClient, err = NewLibvirtClientImpl(s.kvmConfig.LibvirtURI, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create libvirt client: %w", err)
	}

	// Initialize registry
	s.registry, err = NewVMRegistryImpl(s.datastore, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create VM registry: %w", err)
	}

	// Initialize resource checker
	s.resourceChecker, err = NewResourceCheckerImpl(s.resourceService, s.kvmConfig, s.registry, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create resource checker: %w", err)
	}

	// Initialize storage manager
	s.storageManager, err = NewStorageManagerImpl(s.kvmConfig, s.libvirtClient, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create storage manager: %w", err)
	}

	// Initialize network manager
	s.networkManager, err = NewNetworkManagerImpl(s.kvmConfig, s.libvirtClient, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create network manager: %w", err)
	}

	// Initialize VM manager
	s.vmManager, err = NewVMManagerImpl(s.libvirtClient, s.storageManager, s.networkManager, s.registry, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create VM manager: %w", err)
	}

	return nil
}

// Start starts the KVM service
func (s *Service) Start(ctx context.Context) error {
	if !s.kvmConfig.Enabled {
		s.logger.Info("KVM service is disabled, skipping start")
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("KVM service is already running")
	}

	s.logger.Info("Starting KVM service")

	// Connect to libvirt
	if err := s.libvirtClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}

	// Validate system capabilities
	if err := s.validateSystemCapabilities(); err != nil {
		return fmt.Errorf("system validation failed: %w", err)
	}

	// Initialize storage pools and networks
	if err := s.initializeInfrastructure(); err != nil {
		return fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Start background tasks
	go s.runBackgroundTasks(ctx)

	s.running = true
	s.logger.Info("KVM service started successfully")

	return nil
}

// Stop stops the KVM service
func (s *Service) Stop(ctx context.Context) error {
	if !s.kvmConfig.Enabled {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.logger.Info("Stopping KVM service")

	// Signal stop to background tasks
	close(s.stopChan)

	// Disconnect from libvirt
	if s.libvirtClient != nil {
		if err := s.libvirtClient.Disconnect(); err != nil {
			s.logger.WithError(err).Error("Failed to disconnect from libvirt")
		}
	}

	s.running = false
	s.logger.Info("KVM service stopped")

	return nil
}

// IsHealthy checks if the KVM service is healthy
func (s *Service) IsHealthy(ctx context.Context) error {
	if !s.kvmConfig.Enabled {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return fmt.Errorf("KVM service is not running")
	}

	// Check libvirt connection
	if !s.libvirtClient.IsConnected() {
		return fmt.Errorf("libvirt connection is not healthy")
	}

	return nil
}

// CreateVM creates a new virtual machine
func (s *Service) CreateVM(ctx context.Context, req *CreateVMRequest) (*VMDetails, error) {
	if !s.kvmConfig.Enabled {
		return nil, &KVMError{
			Code:    ErrCodeServiceUnavailable,
			Message: "KVM service is disabled",
		}
	}

	s.logger.WithFields(logrus.Fields{
		"name":     req.Name,
		"template": req.Template,
		"cpu":      req.CPU,
		"memory":   req.Memory,
		"disk":     req.Disk,
	}).Info("Creating new VM")

	// Validate request
	if err := s.validateCreateVMRequest(req); err != nil {
		return nil, &KVMError{
			Code:    ErrCodeValidationError,
			Message: "Invalid VM creation request",
			Details: err.Error(),
		}
	}

	// Check resource availability
	resourceReq := s.createResourceRequirement(req)
	if err := s.resourceChecker.CheckAvailableResources(resourceReq); err != nil {
		return nil, &KVMError{
			Code:    ErrCodeResourceInsufficient,
			Message: "Insufficient resources",
			Details: err.Error(),
		}
	}

	// Create VM specification
	vmSpec := &VMSpec{
		Name:        req.Name,
		Template:    req.Template,
		CPU:         req.CPU,
		Memory:      req.Memory,
		Disk:        req.Disk,
		NetworkType: NetworkTypeNAT, // Default to NAT
		CloudInit:   req.CloudInit,
		Metadata:    req.Metadata,
	}

	// Set network type if specified
	if req.Network != "" {
		networkType, err := s.parseNetworkType(req.Network)
		if err != nil {
			return nil, &KVMError{
				Code:    ErrCodeValidationError,
				Message: "Invalid network type",
				Details: err.Error(),
			}
		}
		vmSpec.NetworkType = networkType
	}

	// Provision the VM
	vm, err := s.vmManager.Provision(ctx, vmSpec)
	if err != nil {
		return nil, &KVMError{
			Code:    ErrCodeLibvirtError,
			Message: "Failed to provision VM",
			Details: err.Error(),
		}
	}

	// Convert to VMDetails
	vmDetails := s.convertVMToDetails(vm)

	s.logger.WithField("vm_id", vm.ID).Info("VM created successfully")

	return vmDetails, nil
}

// DeleteVM deletes a virtual machine
func (s *Service) DeleteVM(ctx context.Context, vmID string) error {
	if !s.kvmConfig.Enabled {
		return &KVMError{
			Code:    ErrCodeServiceUnavailable,
			Message: "KVM service is disabled",
		}
	}

	s.logger.WithField("vm_id", vmID).Info("Deleting VM")

	// Check if VM exists
	if !s.registry.VMExists(vmID) {
		return &KVMError{
			Code:    ErrCodeVMNotFound,
			Message: "VM not found",
			Details: vmID,
		}
	}

	// Stop VM if running
	status, err := s.vmManager.GetStatus(ctx, vmID)
	if err != nil {
		s.logger.WithError(err).WithField("vm_id", vmID).Warn("Failed to get VM status before deletion")
	} else if status == VMStatusRunning {
		if err := s.vmManager.Stop(ctx, vmID); err != nil {
			s.logger.WithError(err).WithField("vm_id", vmID).Warn("Failed to stop VM before deletion")
		}
	}

	// Destroy the VM
	if err := s.vmManager.Destroy(ctx, vmID); err != nil {
		return &KVMError{
			Code:    ErrCodeLibvirtError,
			Message: "Failed to destroy VM",
			Details: err.Error(),
		}
	}

	s.logger.WithField("vm_id", vmID).Info("VM deleted successfully")

	return nil
}

// ListVMs lists all virtual machines
func (s *Service) ListVMs(ctx context.Context, filters map[string]string) ([]*VMDetails, error) {
	if !s.kvmConfig.Enabled {
		return nil, &KVMError{
			Code:    ErrCodeServiceUnavailable,
			Message: "KVM service is disabled",
		}
	}

	vmMetadataList, err := s.registry.ListVMs(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs from registry: %w", err)
	}

	var vmDetailsList []*VMDetails
	for _, metadata := range vmMetadataList {
		vmDetails := s.convertMetadataToDetails(metadata)
		vmDetailsList = append(vmDetailsList, vmDetails)
	}

	return vmDetailsList, nil
}

// GetVM gets details of a specific virtual machine
func (s *Service) GetVM(ctx context.Context, vmID string) (*VMDetails, error) {
	if !s.kvmConfig.Enabled {
		return nil, &KVMError{
			Code:    ErrCodeServiceUnavailable,
			Message: "KVM service is disabled",
		}
	}

	metadata, err := s.registry.GetVM(vmID)
	if err != nil {
		return nil, &KVMError{
			Code:    ErrCodeVMNotFound,
			Message: "VM not found",
			Details: vmID,
		}
	}

	vmDetails := s.convertMetadataToDetails(metadata)

	return vmDetails, nil
}

// StartVM starts a virtual machine
func (s *Service) StartVM(ctx context.Context, vmID string) error {
	if !s.kvmConfig.Enabled {
		return &KVMError{
			Code:    ErrCodeServiceUnavailable,
			Message: "KVM service is disabled",
		}
	}

	s.logger.WithField("vm_id", vmID).Info("Starting VM")

	if !s.registry.VMExists(vmID) {
		return &KVMError{
			Code:    ErrCodeVMNotFound,
			Message: "VM not found",
			Details: vmID,
		}
	}

	if err := s.vmManager.Start(ctx, vmID); err != nil {
		return &KVMError{
			Code:    ErrCodeLibvirtError,
			Message: "Failed to start VM",
			Details: err.Error(),
		}
	}

	s.logger.WithField("vm_id", vmID).Info("VM started successfully")

	return nil
}

// StopVM stops a virtual machine
func (s *Service) StopVM(ctx context.Context, vmID string) error {
	if !s.kvmConfig.Enabled {
		return &KVMError{
			Code:    ErrCodeServiceUnavailable,
			Message: "KVM service is disabled",
		}
	}

	s.logger.WithField("vm_id", vmID).Info("Stopping VM")

	if !s.registry.VMExists(vmID) {
		return &KVMError{
			Code:    ErrCodeVMNotFound,
			Message: "VM not found",
			Details: vmID,
		}
	}

	if err := s.vmManager.Stop(ctx, vmID); err != nil {
		return &KVMError{
			Code:    ErrCodeLibvirtError,
			Message: "Failed to stop VM",
			Details: err.Error(),
		}
	}

	s.logger.WithField("vm_id", vmID).Info("VM stopped successfully")

	return nil
}

// RestartVM restarts a virtual machine
func (s *Service) RestartVM(ctx context.Context, vmID string) error {
	if !s.kvmConfig.Enabled {
		return &KVMError{
			Code:    ErrCodeServiceUnavailable,
			Message: "KVM service is disabled",
		}
	}

	s.logger.WithField("vm_id", vmID).Info("Restarting VM")

	if !s.registry.VMExists(vmID) {
		return &KVMError{
			Code:    ErrCodeVMNotFound,
			Message: "VM not found",
			Details: vmID,
		}
	}

	if err := s.vmManager.Restart(ctx, vmID); err != nil {
		return &KVMError{
			Code:    ErrCodeLibvirtError,
			Message: "Failed to restart VM",
			Details: err.Error(),
		}
	}

	s.logger.WithField("vm_id", vmID).Info("VM restarted successfully")

	return nil
}

// Helper methods

func (s *Service) validateSystemCapabilities() error {
	capabilities, err := s.resourceChecker.GetSystemCapabilities()
	if err != nil {
		return fmt.Errorf("failed to get system capabilities: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"total_cpu":     capabilities.TotalCPU,
		"total_memory":  capabilities.TotalMemory,
		"total_storage": capabilities.TotalStorage,
		"max_vms":       capabilities.MaxVMs,
	}).Info("System capabilities validated")

	return nil
}

func (s *Service) initializeInfrastructure() error {
	// This would initialize storage pools, networks, etc.
	// Implementation depends on the specific infrastructure setup
	s.logger.Info("Infrastructure initialized")
	return nil
}

func (s *Service) runBackgroundTasks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.performMaintenanceTasks()
		}
	}
}

func (s *Service) performMaintenanceTasks() {
	// Update VM statuses, cleanup orphaned resources, etc.
	s.logger.Debug("Performing maintenance tasks")
}

func (s *Service) validateCreateVMRequest(req *CreateVMRequest) error {
	if req.Name == "" {
		return fmt.Errorf("VM name is required")
	}

	if req.Template == "" {
		return fmt.Errorf("template is required")
	}

	if req.CPU < 1 || req.CPU > 16 {
		return fmt.Errorf("CPU must be between 1 and 16")
	}

	if req.Memory < 512 {
		return fmt.Errorf("memory must be at least 512 MB")
	}

	if req.Disk < 1 {
		return fmt.Errorf("disk size must be at least 1 GB")
	}

	return nil
}

func (s *Service) createResourceRequirement(req *CreateVMRequest) *ResourceRequirement {
	// Convert MB to bytes for memory, GB to bytes for disk
	memoryBytes := int64(req.Memory) * 1024 * 1024
	diskBytes := int64(req.Disk) * 1024 * 1024 * 1024

	return &ResourceRequirement{
		CPU:    big.NewInt(int64(req.CPU)),
		Memory: big.NewInt(memoryBytes),
		Disk:   big.NewInt(diskBytes),
	}
}

func (s *Service) parseNetworkType(network string) (NetworkType, error) {
	switch network {
	case "bridge":
		return NetworkTypeBridge, nil
	case "nat":
		return NetworkTypeNAT, nil
	case "isolated":
		return NetworkTypeIsolated, nil
	default:
		return NetworkTypeNAT, fmt.Errorf("unknown network type: %s", network)
	}
}

func (s *Service) convertVMToDetails(vm *VM) *VMDetails {
	return &VMDetails{
		ID:          vm.ID,
		Name:        vm.Name,
		Status:      vm.Status,
		IPAddress:   vm.Network.IPAddress.String(),
		MACAddress:  vm.Network.MACAddress,
		Resources:   vm.Resources,
		CreatedAt:   vm.CreatedAt,
		UpdatedAt:   time.Now(),
		Template:    vm.Metadata["template"],
		NetworkType: vm.Network.Type,
		Metadata:    vm.Metadata,
	}
}

func (s *Service) convertMetadataToDetails(metadata *VMMetadata) *VMDetails {
	return &VMDetails{
		ID:         metadata.ID,
		Name:       metadata.Name,
		Status:     metadata.Status,
		IPAddress:  metadata.IPAddress,
		MACAddress: metadata.MACAddress,
		Resources: &ResourceInfo{
			CPU:    int(metadata.Resources.CPU.Int64()),
			Memory: int(metadata.Resources.Memory.Int64() / (1024 * 1024)),      // Convert to MB
			Disk:   int(metadata.Resources.Disk.Int64() / (1024 * 1024 * 1024)), // Convert to GB
		},
		CreatedAt:   metadata.CreatedAt,
		UpdatedAt:   metadata.UpdatedAt,
		Template:    metadata.Template,
		NetworkType: metadata.NetworkType,
		Metadata:    metadata.Metadata,
	}
}

// parseKVMConfig parses KVM configuration from the main config
func parseKVMConfig(cfg *config.C) (*KVMConfig, error) {
	kvmConfig := &KVMConfig{
		Enabled:        cfg.GetBool("kvm.enabled", false),
		LibvirtURI:     cfg.GetString("kvm.libvirt_uri", "qemu:///system"),
		StoragePath:    cfg.GetString("kvm.storage_path", "/var/lib/subnet-node/kvm/storage"),
		TemplatePath:   cfg.GetString("kvm.template_path", "/var/lib/subnet-node/kvm/templates"),
		MaxVMs:         cfg.GetInt("kvm.max_vms", 10),
		ReservedCPU:    0.2, // TODO: Parse from config
		ReservedMemory: 0.2, // TODO: Parse from config
		DefaultNetwork: cfg.GetString("kvm.default_network", "default"),
		StoragePool:    cfg.GetString("kvm.storage_pool", "default"),
		Networks:       make(map[string]*NetworkConfig),
		DefaultResources: &DefaultResources{
			CPU:    cfg.GetInt("kvm.default_resources.cpu", 1),
			Memory: cfg.GetInt("kvm.default_resources.memory", 1024),
			Disk:   cfg.GetInt("kvm.default_resources.disk", 10),
		},
	}

	return kvmConfig, nil
}
