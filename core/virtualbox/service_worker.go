package virtualbox

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/hardware_detector"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

func (s *VirtualboxService) StartWorker(ctx context.Context) {
	go s.processRequests(ctx)
}

// processRequests processes VM requests from the channel
func (s *VirtualboxService) processRequests(ctx context.Context) {
	for {
		select {
		case request := <-s.jobManager.GetRequestChannel():
			go s.handleVMRequest(ctx, request)
		case <-s.stopChan:
			serviceLog.Info("Request processing stopped")
			return
		case <-ctx.Done():
			serviceLog.Info("Context cancelled, stopping request processing")
			return
		}
	}
}

// handleVMRequest handles VM operation requests from the channel
func (s *VirtualboxService) handleVMRequest(ctx context.Context, request *vbtypes.VMRequest) {
	serviceLog.WithFields(logrus.Fields{
		"requestType": request.Type,
		"vmID":        request.VMID,
		"vmName":      request.VMName,
		"timestamp":   request.Timestamp,
	}).Info("Handling VM request")

	switch request.Type {
	case vbtypes.VMEventCreateVM:
		s.handleCreateVMRequest(ctx, request)
	case vbtypes.VMEventCreateTemplateVM:
		s.handleCreateTemplateVMRequest(ctx, request)
	default:
		serviceLog.WithField("requestType", request.Type).Error("Unknown request type")
	}
}

// handleCreateVMRequest handles VM creation requests
func (s *VirtualboxService) handleCreateVMRequest(ctx context.Context, request *vbtypes.VMRequest) {
	serviceLog.WithField("vmName", request.VMName).Info("Handling VM creation request")

	// Extract request data
	reqData, ok := request.Data["request"].(vbtypes.VMCreateRequest)
	if !ok {
		serviceLog.WithField("vmName", request.VMName).Error("Failed to extract create request from request data")
		return
	}

	// Perform the actual VM creation
	vm, err := s.performCreateVM(ctx, reqData)
	if err != nil {
		serviceLog.WithFields(logrus.Fields{
			"vmName": request.VMName,
			"error":  err,
		}).Error("Failed to create VM")
		return
	}

	// Update request with actual VM info
	request.VMID = vm.ID
	request.VMName = vm.Name
	request.VMStatus = vm.Status

	serviceLog.WithFields(logrus.Fields{
		"vmID":   vm.ID,
		"vmName": vm.Name,
	}).Info("VM creation completed successfully")
}

func (s *VirtualboxService) handleCreateTemplateVMRequest(ctx context.Context, request *vbtypes.VMRequest) {
	serviceLog.WithField("vmName", request.VMName).Info("Handling VM template creation request")

	// Extract request data
	reqData, ok := request.Data["request"].(vbtypes.VMCreateRequest)
	if !ok {
		serviceLog.WithField("vmName", request.VMName).Error("Failed to extract create request from request data")
		return
	}

	// Perform the actual VM template creation
	vm, err := s.performCreateTemplateVM(ctx, reqData)
	if err != nil {
		serviceLog.WithFields(logrus.Fields{
			"vmName": request.VMName,
			"error":  err,
		}).Error("Failed to create VM template")
		return
	}

	// Update request with actual VM info
	request.VMID = vm.ID
	request.VMName = vm.Name
	request.VMStatus = vm.Status

	serviceLog.WithFields(logrus.Fields{
		"vmID":   vm.ID,
		"vmName": vm.Name,
	}).Info("VM template creation completed successfully")
}

// performCreateTemplateVM performs the actual VM template creation (moved from CreateTemplateVM)
func (s *VirtualboxService) performCreateTemplateVM(ctx context.Context, req vbtypes.VMCreateRequest) (*vbtypes.VM, error) {
	// Validate system resources before creating VM
	serviceLog.Infof("Validating system resources...")
	if err := s.validateResources(ctx, req); err != nil {
		return nil, fmt.Errorf("resource validation failed: %w", err)
	}
	serviceLog.Infof("Resource validation passed")

	s.mu.Lock()
	defer s.mu.Unlock()

	serviceLog.Infof("Creating VM: %s", req.Name)

	// Validate OS type compatibility with hardware if provided
	if req.OSType != "" {
		serviceLog.Infof("Validating OS type compatibility: %s", req.OSType)
		if err := hardware_detector.ValidateOSTypeCompatibility(req.OSType); err != nil {
			return nil, fmt.Errorf("OS type validation failed: %w", err)
		}
		serviceLog.Infof("OS type validation passed")
	}

	// Generate unique VM ID
	vmID := generateVMID(req.Name)
	serviceLog.Infof("Generated VM ID: %s", vmID)

	// Determine OS type and ISO URL based on the request
	serviceLog.Infof("Determining OS type and ISO URL...")
	osType, isoURL, err := s.storageMgr.DetermineOSTypeAndISO(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to determine OS type: %w", err)
	}
	serviceLog.Infof("OS Type: %s, ISO URL: %s", osType, isoURL)

	// Download ISO if needed
	serviceLog.Infof("Ensuring ISO is available...")
	isoPath, err := s.ensureISO(ctx, isoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure ISO: %w", err)
	}
	serviceLog.Infof("ISO Path: %s", isoPath)

	// Create VM directory
	vmFolder := filepath.Join(s.vmDir, req.Name)
	serviceLog.Infof("Creating VM directory: %s", vmFolder)
	if err := fsutil.DirWritable(vmFolder); err != nil {
		return nil, fmt.Errorf("failed to create VM directory: %w", err)
	}

	// Create VM using VBoxManage
	serviceLog.Infof("Creating VM with VBoxManage...")
	if err := s.vboxExec.CreateVM(req.Name, req.OSType); err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}
	serviceLog.Infof("VM created successfully")
	// Get the actual UUID from VBoxManage
	output, err := s.vboxExec.executeCommand("showvminfo", req.Name, "--machinereadable")
	if err != nil {
		return nil, fmt.Errorf("failed to get VM info for UUID: %w", err)
	}
	vmInfo := s.parseMachineReadableOutput(output)
	vmUuid := vmInfo["UUID"]
	vmName := req.Name
	if vmUuid == "" {
		return nil, fmt.Errorf("could not retrieve VM UUID from VBoxManage output")
	}

	// Configure VM hardware using VBoxManage
	serviceLog.Infof("Configuring VM hardware...")
	if err := s.vboxExec.ConfigureVMHardware(vmName, req.CPUCores, req.MemoryMB); err != nil {
		serviceLog.Errorf("Failed to configure VM hardware: %v", err)
		// Clean up on failure
		serviceLog.Infof("Cleaning up failed VM...")
		if delErr := s.vboxExec.DeleteVM(vmName); delErr != nil {
			serviceLog.Errorf("Failed to delete VM during cleanup: %v", delErr)
		}
		return nil, fmt.Errorf("failed to configure VM hardware: %w", err)
	}
	serviceLog.Infof("VM hardware configured successfully")

	// Configure network adapter
	serviceLog.Infof("Configuring network adapter...")
	if err := s.vboxExec.ConfigureNetwork(vmName, "nat"); err != nil {
		serviceLog.Errorf("Failed to configure network adapter: %v", err)
		// Clean up on failure
		serviceLog.Infof("Cleaning up failed VM...")
		if delErr := s.vboxExec.DeleteVM(vmName); delErr != nil {
			serviceLog.Errorf("Failed to delete VM during cleanup: %v", delErr)
		}
		return nil, fmt.Errorf("failed to configure network adapter: %w", err)
	}
	serviceLog.Infof("Network adapter configured successfully")

	// Create and attach storage using VBoxManage
	serviceLog.Infof("Setting up VM storage...")

	// Generate cloud-init ISO
	cloudInitISO := ""
	cloudInitISO, err = s.generateCloudInitISO(req.Name, req.Username, req.Password)
	if err != nil {
		serviceLog.Errorf("Failed to generate cloud-init ISO: %v", err)
		// Continue without cloud-init ISO - it's not critical for VM creation
	}

	if err := s.vboxExec.SetupStorage(vmName, req, isoPath, cloudInitISO); err != nil {
		serviceLog.Errorf("Failed to setup VM storage: %v", err)
		// Clean up on failure
		serviceLog.Infof("Cleaning up failed VM...")
		if delErr := s.vboxExec.DeleteVM(vmUuid); delErr != nil {
			serviceLog.Errorf("Failed to delete VM during cleanup: %v", delErr)
		}
		return nil, fmt.Errorf("failed to setup VM storage: %w", err)
	}
	serviceLog.Infof("VM storage setup completed")

	// Create VM object (no CreatedAt/UpdatedAt)
	vm := &vbtypes.VM{
		ID:         vmUuid,
		Name:       req.Name,
		Status:     vbtypes.Stopped,
		CPUCores:   req.CPUCores,
		MemoryMB:   req.MemoryMB,
		DiskSizeGB: req.DiskSizeGB,
		VMFolder:   vmFolder,
	}

	// Store VM metadata in datastore
	if err := s.storeVMMetadata(ctx, vm); err != nil {
		serviceLog.Warnf("Failed to store VM metadata in datastore: %v", err)
	}

	serviceLog.Infof("Successfully created VM: %s", req.Name)
	return vm, nil
}
