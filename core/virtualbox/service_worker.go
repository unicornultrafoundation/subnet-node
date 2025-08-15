package virtualbox

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/storage"
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
		s.handleCreateVMFromImageRequest(ctx, request)
	default:
		serviceLog.WithField("requestType", request.Type).Error("Unknown request type")
	}
}

func extractVMCreateFromImageRequest(data interface{}) (*vbtypes.VMCreateFromImageRequest, error) {
	switch v := data.(type) {
	case vbtypes.VMCreateFromImageRequest:
		return &v, nil
	case *vbtypes.VMCreateFromImageRequest:
		return v, nil
	case map[string]interface{}:
		// Convert map to JSON bytes, then unmarshal to struct
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
		}
		var req vbtypes.VMCreateFromImageRequest
		if err := json.Unmarshal(jsonBytes, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON to VMCreateFromImageRequest: %w", err)
		}
		return &req, nil
	default:
		return nil, fmt.Errorf("unsupported type for VMCreateFromImageRequest: %T", data)
	}
}

func (s *VirtualboxService) handleCreateVMFromImageRequest(ctx context.Context, request *vbtypes.VMRequest) {

	serviceLog.WithField("vmName", request.VMName).Info("Handling VM creation request")

	// Extract job ID from request data
	jobID, ok := request.Data["jobID"].(string)
	if !ok {
		serviceLog.WithField("vmName", request.VMName).Error("Failed to extract job ID from request data")
		return
	}

	// Start the job
	if err := s.jobManager.StartJob(ctx, jobID); err != nil {
		serviceLog.WithFields(logrus.Fields{
			"vmName": request.VMName,
			"jobID":  jobID,
			"error":  err,
		}).Error("Failed to start job")
		return
	}

	// Extract request data using JSON marshaling/unmarshaling
	reqData, err := extractVMCreateFromImageRequest(request.Data["request"])
	if err != nil {
		serviceLog.WithField("vmName", request.VMName).Error("Failed to extract create request from request data")
		// Fail the job
		if failErr := s.jobManager.FailJob(ctx, jobID, "Failed to extract create request from request data"); failErr != nil {
			serviceLog.WithError(failErr).Error("Failed to mark job as failed")
		}
		return
	}
	if !ok {
		serviceLog.WithField("vmName", request.VMName).Error("Failed to extract create request from request data")
		// Fail the job
		if failErr := s.jobManager.FailJob(ctx, jobID, "Failed to extract create request from request data"); failErr != nil {
			serviceLog.WithError(failErr).Error("Failed to mark job as failed")
		}
		return
	}

	// Perform the actual VM creation
	vm, err := s.performCreateVMFromImage(ctx, *reqData)
	if err != nil {
		serviceLog.WithFields(logrus.Fields{
			"vmName": request.VMName,
			"error":  err,
		}).Error("Failed to create VM")

		// Fail the job
		if failErr := s.jobManager.FailJob(ctx, jobID, err.Error()); failErr != nil {
			serviceLog.WithError(failErr).Error("Failed to mark job as failed")
		}
		return
	}

	// Update request with actual VM info
	request.VMID = vm.ID
	request.VMName = vm.Name
	request.VMStatus = vm.Status

	// Store the orderId to vmId mapping if orderId is provided
	if reqData.OrderId != "" {
		s.StoreOrderVMMapping(reqData.OrderId, vm.ID)
		serviceLog.WithFields(logrus.Fields{
			"orderId": reqData.OrderId,
			"vmId":    vm.ID,
		}).Info("Stored order to VM mapping after VM creation")
	}

	// Complete the job with success
	result := map[string]interface{}{
		"vm_id":        vm.ID,
		"vm_name":      vm.Name,
		"vm_status":    vm.Status,
		"cpu_cores":    vm.CPUCores,
		"memory_mb":    vm.MemoryMB,
		"disk_size_gb": vm.DiskSizeGB,
		"vm_folder":    vm.VMFolder,
		"ssh_port":     vm.SSHPort,
	}

	if err := s.jobManager.CompleteJob(ctx, jobID, vm.ID, result); err != nil {
		serviceLog.WithFields(logrus.Fields{
			"vmName": request.VMName,
			"jobID":  jobID,
			"error":  err,
		}).Error("Failed to mark job as completed")
	}

	serviceLog.WithFields(logrus.Fields{
		"vmID":   vm.ID,
		"vmName": vm.Name,
		"jobID":  jobID,
	}).Info("VM creation from image completed successfully")

}

func (s *VirtualboxService) performCreateVMFromImage(ctx context.Context, req vbtypes.VMCreateFromImageRequest) (*vbtypes.VM, error) {
	serviceLog.Infof("Creating VM from image: %s", req.Name)

	// Initialize storage manager
	osStorage := storage.NewStorageManager(s.vmDir)
	arch := runtime.GOARCH

	// Get or create VDI file
	vdiPath, err := osStorage.GetOrCreateVDI(ctx, req.Version, arch, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create VDI: %w", err)
	}

	// Generate cloud-init ISO
	cloudInitISO, err := osStorage.GenerateCloudInitISO(ctx, req.Name, req.Username, req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cloud-init ISO: %w", err)
	}

	serviceLog.Infof("VDI path: %s", vdiPath)
	serviceLog.Infof("Cloud-init ISO: %s", cloudInitISO)

	// Validate system resources
	serviceLog.Infof("Validating system resources...")
	if err := s.validateResources(ctx, req.CPUCores, req.MemoryMB, req.DiskSizeGB); err != nil {
		return nil, fmt.Errorf("resource validation failed: %w", err)
	}
	serviceLog.Infof("Resource validation passed")

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create VM without OS type
	serviceLog.Infof("Creating VM without OS type: %s", req.Name)
	if err := s.vboxExec.CreateVMWithoutOS(req.Name); err != nil {
		return nil, fmt.Errorf("failed to create VM without OS: %w", err)
	}

	// Get the actual UUID from VBoxManage
	output, err := s.vboxExec.executeCommand("showvminfo", req.Name, "--machinereadable")
	if err != nil {
		return nil, fmt.Errorf("failed to get VM info for UUID: %w", err)
	}
	vmInfo := s.parseMachineReadableOutput(output)
	vmUuid := vmInfo["UUID"]
	if vmUuid == "" {
		return nil, fmt.Errorf("could not retrieve VM UUID from VBoxManage output")
	}

	// Configure VM hardware
	serviceLog.Infof("Configuring VM hardware...")
	if err := s.vboxExec.ConfigureVMHardware(req.Name, req.CPUCores, req.MemoryMB); err != nil {
		serviceLog.Errorf("Failed to configure VM hardware: %v", err)
		// Clean up on failure
		serviceLog.Infof("Cleaning up failed VM...")
		if delErr := s.vboxExec.DeleteVM(req.Name); delErr != nil {
			serviceLog.Errorf("Failed to delete VM during cleanup: %v", delErr)
		}
		return nil, fmt.Errorf("failed to configure VM hardware: %w", err)
	}
	serviceLog.Infof("VM hardware configured successfully")

	// Configure network adapter
	serviceLog.Infof("Configuring network adapter...")
	if err := s.vboxExec.ConfigureNetwork(req.Name, "nat"); err != nil {
		serviceLog.Errorf("Failed to configure network adapter: %v", err)
		// Clean up on failure
		serviceLog.Infof("Cleaning up failed VM...")
		if delErr := s.vboxExec.DeleteVM(req.Name); delErr != nil {
			serviceLog.Errorf("Failed to delete VM during cleanup: %v", delErr)
		}
		return nil, fmt.Errorf("failed to configure network adapter: %w", err)
	}
	serviceLog.Infof("Network adapter configured successfully")

	// Resize VDI file to match requested disk size
	serviceLog.Infof("Resizing VDI file to %d GB: %s", req.DiskSizeGB, vdiPath)
	if err := s.vboxExec.ResizeVDI(vdiPath, req.DiskSizeGB); err != nil {
		serviceLog.Errorf("Failed to resize VDI file: %v", err)
		// Clean up on failure
		serviceLog.Infof("Cleaning up failed VM...")
		if delErr := s.vboxExec.DeleteVM(req.Name); delErr != nil {
			serviceLog.Errorf("Failed to delete VM during cleanup: %v", delErr)
		}
		return nil, fmt.Errorf("failed to resize VDI file: %w", err)
	}
	serviceLog.Infof("VDI file resized successfully to %d GB", req.DiskSizeGB)

	// Attach existing VDI file using VirtioSCSI
	serviceLog.Infof("Attaching VDI file using VirtioSCSI: %s", vdiPath)
	if err := s.vboxExec.AttachExistingVDI(req.Name, vdiPath); err != nil {
		serviceLog.Errorf("Failed to attach VDI file: %v", err)
		// Clean up on failure
		serviceLog.Infof("Cleaning up failed VM...")
		if delErr := s.vboxExec.DeleteVM(req.Name); delErr != nil {
			serviceLog.Errorf("Failed to delete VM during cleanup: %v", delErr)
		}
		return nil, fmt.Errorf("failed to attach VDI file: %w", err)
	}
	serviceLog.Infof("VDI file attached successfully using VirtioSCSI")

	// Attach cloud-init ISO using VirtioSCSI
	if cloudInitISO != "" {
		serviceLog.Infof("Attaching cloud-init ISO using VirtioSCSI: %s", cloudInitISO)
		if err := s.vboxExec.AttachCloudInitISOVirtioSCSI(req.Name, cloudInitISO); err != nil {
			serviceLog.Warnf("Failed to attach cloud-init ISO: %v", err)
			// Continue without cloud-init ISO - it's not critical for VM operation
		} else {
			serviceLog.Infof("Cloud-init ISO attached successfully using VirtioSCSI")
		}
	}

	// Create VM directory
	vmFolder := filepath.Join(s.vmDir, req.Name)
	serviceLog.Infof("VM directory: %s", vmFolder)

	// Set up SSH port forwarding
	serviceLog.Infof("Setting up SSH port forwarding for VM: %s", req.Name)
	hostPort, err := getAvailablePort()
	if err != nil {
		serviceLog.Errorf("Failed to find available port for SSH forwarding: %v", err)
		// Continue without port forwarding - it's not critical for VM operation
	} else {
		if err := s.vboxExec.SetupSSHPortForward(req.Name, hostPort, 22); err != nil {
			serviceLog.Errorf("Failed to set up SSH port forwarding: %v", err)
			// Continue without port forwarding - it's not critical for VM operation
		} else {
			serviceLog.Infof("Successfully set up SSH port forwarding: host port %d -> guest port 22", hostPort)
		}
	}

	// Create VM object
	vm := &vbtypes.VM{
		ID:         vmUuid,
		Name:       req.Name,
		Status:     vbtypes.Stopped, // Will be updated after starting
		CPUCores:   req.CPUCores,
		MemoryMB:   req.MemoryMB,
		DiskSizeGB: req.DiskSizeGB,
		VMFolder:   vmFolder,
		SSHPort:    hostPort,
	}

	// Store VM metadata in datastore
	if err := s.storeVMMetadata(ctx, vm); err != nil {
		serviceLog.Warnf("Failed to store VM metadata in datastore: %v", err)
	}

	serviceLog.Infof("Successfully created VM from image: %s", req.Name)
	return vm, nil
}
