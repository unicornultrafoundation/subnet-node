package virtualbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	vbconfig "github.com/unicornultrafoundation/subnet-node/core/virtualbox/config"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/store"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/terraform"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/vbox"
)

// Service represents the VirtualBox service
type Service struct {
	config          *vbconfig.ServiceConfig
	Cfg             *config.C
	storeService    *store.Service
	vboxClient      *vbox.Client
	terraformClient *terraform.Client
	logger          *logrus.Logger
	mu              sync.RWMutex
	vmListCache     map[string]struct{} // Set of VM IDs running in VirtualBox
	monitorCancel   context.CancelFunc
}

// NewService creates a new VirtualBox service
func NewService(cfg *config.C, ds datastore.Datastore) (*Service, error) {
	serviceConfig, err := vbconfig.NewServiceConfigFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	logger := logrus.New().WithField("service", "virtualbox").Logger

	storeService := store.NewService(ds, logger)
	vboxClient := vbox.NewClient(serviceConfig, logger)
	terraformClient := terraform.NewClient(serviceConfig, logger)

	return &Service{
		config:          serviceConfig,
		Cfg:             cfg,
		storeService:    storeService,
		vboxClient:      vboxClient,
		terraformClient: terraformClient,
		logger:          logger,
		vmListCache:     make(map[string]struct{}),
	}, nil
}

// Start starts the VirtualBox service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting VirtualBox service")

	// Validate VirtualBox and Terraform
	if err := s.vboxClient.ValidateVirtualBox(ctx); err != nil {
		return fmt.Errorf("VirtualBox validation failed: %w", err)
	}

	if err := s.terraformClient.ValidateTerraform(ctx); err != nil {
		return fmt.Errorf("Terraform validation failed: %w", err)
	}

	// Start storage service
	if err := s.storeService.Start(ctx); err != nil {
		return err
	}

	// Load existing VMs into cache
	if err := s.loadExistingVMs(ctx); err != nil {
		s.logger.WithError(err).Warn("Failed to load existing VMs")
	}

	// Start VM monitor
	monitorCtx, cancel := context.WithCancel(ctx)
	s.monitorCancel = cancel
	go s.monitorVMs(monitorCtx)

	return nil
}

// Stop stops the VirtualBox service
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping VirtualBox service")

	// Stop monitor
	if s.monitorCancel != nil {
		s.monitorCancel()
	}

	// Stop storage service
	if err := s.storeService.Stop(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to stop storage service")
	}

	return nil
}

// CreateVM creates a new VM
func (s *Service) CreateVM(ctx context.Context, request *types.VMRequest) (*types.VMResponse, error) {
	s.logger.WithField("vmID", request.ID).Info("Creating VM")

	// Store VM request
	if err := s.storeService.StoreVMRequest(ctx, request); err != nil {
		return nil, fmt.Errorf("failed to store VM request: %w", err)
	}

	// Create VM using VirtualBox
	vmInfo, err := s.vboxClient.CreateVM(ctx, request.ID, request.Config)
	if err != nil {
		// Clean up request on failure
		s.storeService.DeleteVMRequest(ctx, request.ID)
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	// Update VM info with metadata
	vmInfo.ID = request.ID
	vmInfo.CreatedAt = request.CreatedAt
	vmInfo.Metadata = request.Metadata
	if vmInfo.Metadata == nil {
		vmInfo.Metadata = make(map[string]string)
	}
	vmInfo.Metadata["requester"] = request.Requester

	// Store VM info
	if err := s.storeService.StoreVMInfo(ctx, vmInfo); err != nil {
		s.logger.WithError(err).WithField("vmID", request.ID).Error("Failed to store VM info")
	}

	// Add to cache
	s.mu.Lock()
	s.vmListCache[request.ID] = struct{}{}
	s.mu.Unlock()

	// Create response
	response := &types.VMResponse{
		ID:        request.ID,
		Name:      vmInfo.Name,
		State:     vmInfo.State,
		UUID:      vmInfo.UUID,
		Requester: request.Requester,
		CreatedAt: request.CreatedAt,
		ExpiresAt: request.CreatedAt.Add(request.TTL),
		Config:    request.Config,
		Status:    &types.VMStatus{State: vmInfo.State},
		Metadata:  vmInfo.Metadata,
	}

	s.logger.WithField("vmID", request.ID).Info("Successfully created VM")
	return response, nil
}

// CreateVMWithTerraform creates a VM using Terraform
func (s *Service) CreateVMWithTerraform(ctx context.Context, request *types.VMRequest) (*types.VMResponse, error) {
	s.logger.WithField("vmID", request.ID).Info("Creating VM with Terraform")

	// Store VM request
	if err := s.storeService.StoreVMRequest(ctx, request); err != nil {
		return nil, fmt.Errorf("failed to store VM request: %w", err)
	}

	// Create Terraform configuration
	tfConfig := &types.TerraformConfig{
		VMName:        fmt.Sprintf("subnet-vm-%s", request.ID),
		OSType:        request.Config.OSType,
		MemoryMB:      request.Config.MemoryMB,
		CPUs:          request.Config.CPUs,
		DiskSizeGB:    request.Config.DiskSizeGB,
		NetworkType:   request.Config.NetworkType,
		BridgeName:    request.Config.BridgeName,
		BaseImagePath: request.Config.BaseImage,
		CustomVars:    request.Config.CustomSettings,
	}

	// Create VM using Terraform
	tfOutput, err := s.terraformClient.CreateVM(ctx, request.ID, tfConfig)
	if err != nil {
		// Clean up request on failure
		s.storeService.DeleteVMRequest(ctx, request.ID)
		return nil, fmt.Errorf("failed to create VM with Terraform: %w", err)
	}

	// Create VM info from Terraform output
	vmInfo := &types.VMInfo{
		ID:        request.ID,
		Name:      tfOutput.VMName,
		UUID:      tfOutput.VMUUID,
		State:     types.VMState(tfOutput.State),
		OSType:    request.Config.OSType,
		MemoryMB:  request.Config.MemoryMB,
		CPUs:      request.Config.CPUs,
		CreatedAt: request.CreatedAt,
		Config:    request.Config,
		Metadata:  request.Metadata,
	}
	if vmInfo.Metadata == nil {
		vmInfo.Metadata = make(map[string]string)
	}
	vmInfo.Metadata["requester"] = request.Requester
	vmInfo.Metadata["terraform_output"] = fmt.Sprintf("%+v", tfOutput)

	// Store VM info
	if err := s.storeService.StoreVMInfo(ctx, vmInfo); err != nil {
		s.logger.WithError(err).WithField("vmID", request.ID).Error("Failed to store VM info")
	}

	// Add to cache
	s.mu.Lock()
	s.vmListCache[request.ID] = struct{}{}
	s.mu.Unlock()

	// Create response
	response := &types.VMResponse{
		ID:        request.ID,
		Name:      vmInfo.Name,
		State:     vmInfo.State,
		UUID:      vmInfo.UUID,
		Requester: request.Requester,
		CreatedAt: request.CreatedAt,
		ExpiresAt: request.CreatedAt.Add(request.TTL),
		Config:    request.Config,
		Status:    &types.VMStatus{State: vmInfo.State},
		Metadata:  vmInfo.Metadata,
	}

	s.logger.WithField("vmID", request.ID).Info("Successfully created VM with Terraform")
	return response, nil
}

// GetVM gets VM information
func (s *Service) GetVM(ctx context.Context, vmID string) (*types.VMResponse, error) {
	// Get VM request
	request, err := s.storeService.GetVMRequest(ctx, vmID)
	if err != nil {
		return nil, fmt.Errorf("VM not found: %s", vmID)
	}

	// Get VM info
	vmInfo, err := s.storeService.GetVMInfo(ctx, vmID)
	if err != nil {
		return nil, fmt.Errorf("VM info not found: %s", vmID)
	}

	// Get current status
	status, err := s.vboxClient.GetVMStats(ctx, vmInfo.Name)
	if err != nil {
		s.logger.WithError(err).WithField("vmID", vmID).Warn("Failed to get VM stats")
		status = &types.VMStatus{State: vmInfo.State}
	}

	// Create response
	response := &types.VMResponse{
		ID:        vmID,
		Name:      vmInfo.Name,
		State:     vmInfo.State,
		UUID:      vmInfo.UUID,
		Requester: request.Requester,
		CreatedAt: request.CreatedAt,
		ExpiresAt: request.CreatedAt.Add(request.TTL),
		Config:    request.Config,
		Status:    status,
		Metadata:  vmInfo.Metadata,
	}

	return response, nil
}

// GetVMs gets all VMs for a requester
func (s *Service) GetVMs(ctx context.Context, requester string) ([]*types.VMResponse, error) {
	// Get VM requests
	requests, err := s.storeService.GetVMRequests(ctx, requester)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM requests: %w", err)
	}

	var responses []*types.VMResponse
	for _, request := range requests {
		response, err := s.GetVM(ctx, request.ID)
		if err != nil {
			s.logger.WithError(err).WithField("vmID", request.ID).Warn("Failed to get VM info")
			continue
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// StartVM starts a VM
func (s *Service) StartVM(ctx context.Context, vmID string) error {
	vmInfo, err := s.storeService.GetVMInfo(ctx, vmID)
	if err != nil {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	if err := s.vboxClient.StartVM(ctx, vmInfo.Name); err != nil {
		return fmt.Errorf("failed to start VM: %w", err)
	}

	s.logger.WithField("vmID", vmID).Info("Started VM")
	return nil
}

// StopVM stops a VM
func (s *Service) StopVM(ctx context.Context, vmID string) error {
	vmInfo, err := s.storeService.GetVMInfo(ctx, vmID)
	if err != nil {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	if err := s.vboxClient.StopVM(ctx, vmInfo.Name); err != nil {
		return fmt.Errorf("failed to stop VM: %w", err)
	}

	s.logger.WithField("vmID", vmID).Info("Stopped VM")
	return nil
}

// ShutdownVM gracefully shuts down a VM
func (s *Service) ShutdownVM(ctx context.Context, vmID string) error {
	vmInfo, err := s.storeService.GetVMInfo(ctx, vmID)
	if err != nil {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	if err := s.vboxClient.ShutdownVM(ctx, vmInfo.Name); err != nil {
		return fmt.Errorf("failed to shutdown VM: %w", err)
	}

	s.logger.WithField("vmID", vmID).Info("Shutdown VM")
	return nil
}

// DeleteVM deletes a VM
func (s *Service) DeleteVM(ctx context.Context, vmID string) error {
	vmInfo, err := s.storeService.GetVMInfo(ctx, vmID)
	if err != nil {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	// Delete VM from VirtualBox
	if err := s.vboxClient.DeleteVM(ctx, vmInfo.Name); err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	// Destroy Terraform resources if they exist
	if err := s.terraformClient.DestroyVM(ctx, vmID); err != nil {
		s.logger.WithError(err).WithField("vmID", vmID).Warn("Failed to destroy Terraform resources")
	}

	// Remove from storage
	if err := s.storeService.DeleteVMRequest(ctx, vmID); err != nil {
		s.logger.WithError(err).WithField("vmID", vmID).Warn("Failed to delete VM request")
	}
	if err := s.storeService.DeleteVMInfo(ctx, vmID); err != nil {
		s.logger.WithError(err).WithField("vmID", vmID).Warn("Failed to delete VM info")
	}

	// Remove from cache
	s.mu.Lock()
	delete(s.vmListCache, vmID)
	s.mu.Unlock()

	s.logger.WithField("vmID", vmID).Info("Deleted VM")
	return nil
}

// GetVMRequest gets the original VM request
func (s *Service) GetVMRequest(ctx context.Context, vmID string) (*types.VMRequest, error) {
	return s.storeService.GetVMRequest(ctx, vmID)
}

// loadExistingVMs loads existing VMs into the cache
func (s *Service) loadExistingVMs(ctx context.Context) error {
	requests, err := s.storeService.GetVMRequests(ctx, "")
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, request := range requests {
		s.vmListCache[request.ID] = struct{}{}
	}

	s.logger.WithField("count", len(requests)).Info("Loaded existing VMs into cache")
	return nil
}

// monitorVMs monitors VMs and cleans up expired ones
func (s *Service) monitorVMs(ctx context.Context) {
	ticker := time.NewTicker(s.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.cleanupExpiredVMs(ctx); err != nil {
				s.logger.WithError(err).Error("Failed to cleanup expired VMs")
			}
		}
	}
}

// cleanupExpiredVMs cleans up expired VMs
func (s *Service) cleanupExpiredVMs(ctx context.Context) error {
	requests, err := s.storeService.GetVMRequests(ctx, "")
	if err != nil {
		return err
	}

	now := time.Now()
	for _, request := range requests {
		expiresAt := request.CreatedAt.Add(request.TTL)
		if now.After(expiresAt) {
			s.logger.WithField("vmID", request.ID).Info("Cleaning up expired VM")
			if err := s.DeleteVM(ctx, request.ID); err != nil {
				s.logger.WithError(err).WithField("vmID", request.ID).Error("Failed to delete expired VM")
			}
		}
	}

	return nil
}
