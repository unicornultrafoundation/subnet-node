package job_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/ipfs/go-datastore"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"

	"github.com/sirupsen/logrus"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var jobLog = logrus.WithField("service", "virtualbox-job-manager")

func (jm *JobManager) startWorker(ctx context.Context) {
	go jm.processRequests(ctx)
}

func (jm *JobManager) processRequests(ctx context.Context) {
	for {
		select {
		case request := <-jm.requestChannel:
			go jm.handleRequest(ctx, request)
		case <-jm.stopChan:
			jobLog.Info("Request processing stopped")
			return
		case <-ctx.Done():
			jobLog.Info("Context cancelled, stopping request processing")
			return
		}
	}
}

func (jm *JobManager) handleRequest(ctx context.Context, request *vbtypes.VMRequest) {

	jobLog.WithField("requestType", request.Type).Info("Handling VM Job request")
	switch request.Type {
	case vbtypes.VMEventCreateVM:
		jm.handleCreateVMRequest(ctx, request)
	default:
		jobLog.WithField("requestType", request.Type).Error("Unknown request type")
	}
}

// Start the job
// Handle the create VM request
func (jm *JobManager) handleCreateVMRequest(ctx context.Context, request *vbtypes.VMRequest) {
	if err := jm.StartJob(ctx, request.Data["jobID"].(string)); err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("Failed to start job")
		return
	}

	reqData, err := extractVMCreateFromImageRequest(request.Data["request"])
	if err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("Failed to extract create request from request data")
		return
	}

	// Validate resources
	if err := jm.validateResources(ctx, reqData.CPUCores, reqData.MemoryMB, reqData.DiskSizeGB); err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("Resource validation failed: " + err.Error())
		return
	}

	// Perform the actual VM creation
	// Note: Do NOT hold the lock during time-consuming VM operations
	// This allows GetJobProgress to work concurrently

	// create VM without OS type
	vm, err := jm.vboxService.CreateVMWithoutOS(reqData.Name)

	if err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("VM creation failed: " + err.Error())
		_ = jm.MarkJobAsFailed(ctx, request.Data["jobID"].(string), "VM creation failed: "+err.Error())
		return
	}

	vdiPath, err := jm.storageManager.GetOrCreateVDI(ctx, reqData.Version, runtime.GOARCH, reqData.Name)
	if err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("VDI creation failed: " + err.Error())
		_ = jm.MarkJobAsFailed(ctx, request.Data["jobID"].(string), "VDI creation failed: "+err.Error())
		return
	}

	cloudInitISO, err := jm.storageManager.GenerateCloudInitISO(ctx, reqData.Name, reqData.Username, reqData.Password)
	if err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("Cloud init ISO generation failed: " + err.Error())
		_ = jm.MarkJobAsFailed(ctx, request.Data["jobID"].(string), "Cloud init ISO generation failed: "+err.Error())
		return
	}

	// Configure VM hardware
	if err := jm.vboxService.ConfigureVMHardware(vm.Name, reqData.CPUCores, reqData.MemoryMB); err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("VM hardware configuration failed: " + err.Error())
		_ = jm.MarkJobAsFailed(ctx, request.Data["jobID"].(string), "VM hardware configuration failed: "+err.Error())
		return
	}

	// Configure network adapter
	if err := jm.vboxService.ConfigureNetwork(vm.Name, "nat"); err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("Network configuration failed: " + err.Error())
		_ = jm.MarkJobAsFailed(ctx, request.Data["jobID"].(string), "Network configuration failed: "+err.Error())
		return
	}

	// Resize VDI file to match requested disk size
	if err := jm.vboxService.ResizeVDI(vdiPath, reqData.DiskSizeGB); err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("VDI resizing failed: " + err.Error())
		_ = jm.MarkJobAsFailed(ctx, request.Data["jobID"].(string), "VDI resizing failed: "+err.Error())
		return
	}

	// Attach existing VDI file using VirtioSCSI
	if err := jm.vboxService.AttachExistingVDI(vm.Name, vdiPath); err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("VDI attachment failed: " + err.Error())
		_ = jm.MarkJobAsFailed(ctx, request.Data["jobID"].(string), "VDI attachment failed: "+err.Error())
		return
	}

	// attach cloud-init ISO using VirtioSCSI
	if cloudInitISO != "" {
		if err := jm.vboxService.AttachCloudInitISOVirtioSCSI(vm.Name, cloudInitISO); err != nil {
			jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("Cloud init ISO attachment failed: " + err.Error())
			_ = jm.MarkJobAsFailed(ctx, request.Data["jobID"].(string), "Cloud init ISO attachment failed: "+err.Error())
			return
		}
	}

	// Store the orderId to vmId mapping
	if err := jm.storeOrderVMMapping(reqData.OrderId, vm.ID); err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("Failed to store order to VM mapping: " + err.Error())
		_ = jm.MarkJobAsFailed(ctx, request.Data["jobID"].(string), "Failed to store order to VM mapping: "+err.Error())
		return
	}

	// update job status to completed
	// Now CompleteJob can acquire the lock without waiting, and GetJobProgress can work concurrently
	if err := jm.CompleteJob(ctx, request.Data["jobID"].(string), vm.ID, map[string]interface{}{}); err != nil {
		jobLog.WithField("jobID", request.Data["jobID"].(string)).Error("Failed to update job status: " + err.Error())
		return
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

func (jm *JobManager) validateResources(_ context.Context, vmCPUCores int, vmMemoryMB int, vmDiskSizeGB int) error {
	resourceInfo, err := resource.GetResource()
	if err != nil {
		return fmt.Errorf("failed to get resource info: %w", err)
	}

	// resourceInfo take time to run
	// mock data to pass it
	// resourceInfo := &resource.ResourceInfo{
	// 	CPU: resource.CpuInfo{
	// 		Count: 10,
	// 	},
	// 	Memory: resource.MemoryInfo{
	// 		Total: 1024 * 1024 * 1024,
	// 	},
	// 	Storage: resource.StorageInfo{
	// 		Total: 200 * 1024 * 1024 * 1024, // 200GB
	// 	},
	// }

	vms, err := jm.vboxService.ListVMs()
	if err != nil {
		return fmt.Errorf("failed to get VMs: %w", err)
	}

	totalCPUs, totalMem, totalDisk := 0, 0, 0
	for _, vm := range vms {
		vmDetail, err := jm.vboxService.GetVM(vm.ID)
		if err != nil {
			return fmt.Errorf("failed to get VM details: %w", err)
		}
		totalCPUs += vmDetail.CPUCores
		totalMem += vmDetail.MemoryMB
		totalDisk += vmDetail.DiskSizeGB
	}

	totalCPUs += vmCPUCores
	totalMem += vmMemoryMB
	totalDisk += vmDiskSizeGB

	availableCPUs := resourceInfo.CPU.Count
	availableMem := int(resourceInfo.Memory.Total / (1024 * 1024))
	availableDisk := int(resourceInfo.Storage.Total / (1024 * 1024 * 1024))

	if totalCPUs > availableCPUs {
		return fmt.Errorf("insufficient CPU cores: total required %d, available %d", totalCPUs, availableCPUs)
	}
	if totalMem > availableMem {
		return fmt.Errorf("insufficient memory: total required %d MB, available %d MB", totalMem, availableMem)
	}
	if totalDisk > availableDisk {
		return fmt.Errorf("insufficient disk space: total required %d GB, available %d GB", totalDisk, availableDisk)
	}

	if err := jm.validateCPUResources(vmCPUCores, resourceInfo); err != nil {
		return fmt.Errorf("failed to validate CPU resources: %w", err)
	}
	if err := jm.validateMemoryResources(vmMemoryMB, resourceInfo); err != nil {
		return fmt.Errorf("failed to validate memory resources: %w", err)
	}
	if err := jm.validateDiskResources(vmDiskSizeGB, resourceInfo); err != nil {
		return fmt.Errorf("failed to validate disk resources: %w", err)
	}
	return nil
}

// validateCPUResources validates CPU requirements
func (jm *JobManager) validateCPUResources(vmCPUCores int, resourceInfo *resource.ResourceInfo) error {
	// Check if requested CPU cores exceed available cores
	if vmCPUCores > resourceInfo.CPU.Count {
		return fmt.Errorf("insufficient CPU cores: requested %d, available %d", vmCPUCores, resourceInfo.CPU.Count)
	}

	// Check for reasonable CPU allocation (not more than 80% of available cores)
	maxRecommendedCores := int(float64(resourceInfo.CPU.Count) * 0.8)
	if vmCPUCores > maxRecommendedCores {
		jobLog.Warnf("CPU allocation warning: requested %d cores exceeds recommended maximum of %d cores", vmCPUCores, maxRecommendedCores)
	}

	jobLog.Infof("CPU validation passed: %d cores requested, %d available", vmCPUCores, resourceInfo.CPU.Count)
	return nil
}

// validateMemoryResources validates memory requirements
func (jm *JobManager) validateMemoryResources(vmMemoryMB int, resourceInfo *resource.ResourceInfo) error {
	// Convert memory from bytes to MB
	availableMemoryMB := int(resourceInfo.Memory.Total / (1024 * 1024))

	// Check if requested memory exceeds available memory
	if vmMemoryMB > availableMemoryMB {
		return fmt.Errorf("insufficient memory: requested %d MB, available %d MB", vmMemoryMB, availableMemoryMB)
	}

	// Check for reasonable memory allocation (not more than 80% of available memory)
	maxRecommendedMemoryMB := int(float64(availableMemoryMB) * 0.8)
	if vmMemoryMB > maxRecommendedMemoryMB {
		jobLog.Warnf("Memory allocation warning: requested %d MB exceeds recommended maximum of %d MB", vmMemoryMB, maxRecommendedMemoryMB)
	}

	// Add 20% buffer for system overhead
	requiredMemoryWithBuffer := int(float64(vmMemoryMB) * 1.2)
	if requiredMemoryWithBuffer > availableMemoryMB {
		return fmt.Errorf("insufficient memory with system buffer: required %d MB, available %d MB", requiredMemoryWithBuffer, availableMemoryMB)
	}

	jobLog.Infof("Memory validation passed: %d MB requested, %d MB available", vmMemoryMB, availableMemoryMB)
	return nil
}

// validateDiskResources validates disk space requirements
func (jm *JobManager) validateDiskResources(vmDiskSizeGB int, resourceInfo *resource.ResourceInfo) error {
	// Convert storage from bytes to GB
	availableDiskGB := int(resourceInfo.Storage.Total / (1024 * 1024 * 1024))

	// Check if requested disk space exceeds available disk space
	if vmDiskSizeGB > availableDiskGB {
		return fmt.Errorf("insufficient disk space: requested %d GB, available %d GB", vmDiskSizeGB, availableDiskGB)
	}

	// Check for reasonable disk allocation (not more than 90% of available disk space)
	maxRecommendedDiskGB := int(float64(availableDiskGB) * 0.9)
	if vmDiskSizeGB > maxRecommendedDiskGB {
		jobLog.Warnf("Disk allocation warning: requested %d GB exceeds recommended maximum of %d GB", vmDiskSizeGB, maxRecommendedDiskGB)
	}

	// Add 10% buffer for overhead (file system, metadata, etc.)
	requiredDiskWithBuffer := int(float64(vmDiskSizeGB) * 1.1)
	if requiredDiskWithBuffer > availableDiskGB {
		return fmt.Errorf("insufficient disk space with overhead buffer: required %d GB, available %d GB", requiredDiskWithBuffer, availableDiskGB)
	}

	jobLog.Infof("Disk validation passed: %d GB requested, %d GB available", vmDiskSizeGB, availableDiskGB)
	return nil
}

func (jm *JobManager) storeOrderVMMapping(orderId, vmId string) error {

	jm.orderMapMu.Lock()
	defer jm.orderMapMu.Unlock()

	(*jm.orderToVMMap)[orderId] = vmId

	// Persist to datastore
	mapping := map[string]string{
		"orderId": orderId,
		"vmId":    vmId,
	}
	data, err := json.Marshal(mapping)
	if err != nil {
		jobLog.WithError(err).Error("Failed to marshal mapping to JSON")
		return err
	}

	key := datastore.NewKey("virtualbox/order_mapping/" + orderId)
	return jm.datastore.Put(context.Background(), key, data)

}
