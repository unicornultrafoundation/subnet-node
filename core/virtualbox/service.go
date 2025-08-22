package virtualbox

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/ssh_connection"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var serviceLog = logrus.WithField("service", "virtualbox")

// VirtualboxService implements the Service interface
type VirtualboxService struct {
	vmDir      string
	mu         sync.RWMutex
	vboxExec   *VBoxManageExecutor
	datastore  datastore.Datastore
	syncTicker *time.Ticker

	sshServer *ssh_connection.SSHServer
	wsHandler *ssh_connection.WebSocketHandler

	stopChan chan struct{}

	jobManager *JobManager

	// OrderId to VMId mapping
	orderToVMMap map[string]string
	orderMapMu   sync.RWMutex
}

// IsVirtualBoxEnabled checks if VirtualBox service is enabled in the configuration
func IsVirtualBoxEnabled(cfg *config.C) bool {
	return cfg.GetBool("virtualbox.enable", false)
}

// NewService creates a new VirtualBox service
func NewService(ds datastore.Datastore) (*VirtualboxService, error) {

	// Get VM directory
	vmDir, err := fsutil.ExpandHome("~/VirtualBox VMs")
	if err != nil {
		return nil, fmt.Errorf("failed to expand VM directory path: %w", err)
	}

	// Ensure the VM directory exists
	if err := fsutil.DirWritable(vmDir); err != nil {
		return nil, fmt.Errorf("failed to create VM directory: %w", err)
	}

	// Ensure the ISOs directory exists
	isosDir := filepath.Join(vmDir, "ISOs")
	if err := fsutil.DirWritable(isosDir); err != nil {
		return nil, fmt.Errorf("failed to create ISOs directory: %w", err)
	}
	serviceLog.Infof("Ensured ISOs directory exists at: %s", isosDir)

	// Create SSH server and WebSocket handler
	sshServer := ssh_connection.NewSSHServer()
	wsHandler := ssh_connection.NewWebSocketHandler(sshServer, serviceLog)

	service := &VirtualboxService{
		vmDir:     vmDir,
		stopChan:  make(chan struct{}),
		vboxExec:  NewVBoxManageExecutor(vmDir),
		datastore: ds,

		sshServer: sshServer,
		wsHandler: wsHandler,

		jobManager: NewJobManager(),

		// Initialize order mapping maps
		orderToVMMap: make(map[string]string),
	}

	// Validate VirtualBox installation once during service creation
	if err := service.validateVirtualBoxInstallation(); err != nil {
		return nil, fmt.Errorf("VirtualBox validation failed: %w", err)
	}

	return service, nil
}

// Start starts the VirtualBox service
func (s *VirtualboxService) Start(ctx context.Context) error {
	serviceLog.Info("Starting VirtualBox service")

	// Load order mappings from datastore
	if err := s.loadAllOrderMappings(ctx); err != nil {
		serviceLog.Warnf("Failed to load order mappings from datastore: %v", err)
	}

	// Start the synchronization ticker to run every 15 seconds
	s.syncTicker = time.NewTicker(15 * time.Second)
	go s.syncVMsWithDatastore(ctx)

	// Start token cleanup routine
	s.wsHandler.StartCleanup(s.stopChan)

	// Sync existing running VMs with SSHServer
	if err := s.syncSSHServerWithRunningVMs(ctx); err != nil {
		serviceLog.Warnf("Failed to sync SSHServer with running VMs: %v", err)
	}

	s.StartWorker(ctx)

	serviceLog.Info("VirtualBox service started successfully")
	return nil
}

// Stop stops the VirtualBox service
func (s *VirtualboxService) Stop(ctx context.Context) error {
	serviceLog.Info("Stopping VirtualBox service")

	// Stop the synchronization ticker
	if s.syncTicker != nil {
		s.syncTicker.Stop()
	}

	// Stop token cleanup routine
	s.wsHandler.StopCleanup()

	close(s.stopChan)
	return nil
}

func (s *VirtualboxService) CreateVM(ctx context.Context, req vbtypes.VMCreateFromImageRequest) (*vbtypes.JobCreateResponse, error) {
	// Create a job for tracking progress
	requestData := map[string]interface{}{
		"name":         req.Name,
		"cpu_cores":    req.CPUCores,
		"memory_mb":    req.MemoryMB,
		"disk_size_gb": req.DiskSizeGB,
		"os":           req.OS,
		"version":      req.Version,
		"username":     req.Username,
		"password":     req.Password,
		"order_id":     req.OrderId,
	}

	job, err := s.jobManager.CreateJob(ctx, vbtypes.VMEventCreateVM, requestData, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return &vbtypes.JobCreateResponse{
		JobID: job.ID,
	}, nil
}

// GetVM gets a VM by ID using VBoxManage
func (s *VirtualboxService) GetVM(ctx context.Context, vmId string) (*vbtypes.VM, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try datastore first
	if vm, err := s.getVMMetadata(ctx, vmId); err == nil && vm != nil {
		// Always get the latest status from VBoxManage
		output, err := s.vboxExec.executeCommand("showvminfo", vmId, "--machinereadable")
		if err == nil {
			vmInfo := s.parseMachineReadableOutput(output)
			vm.Status = parseVMStatus(vmInfo["VMState"])
			// If SSHPort is not set, try to parse from NAT rules
			if vm.SSHPort == 0 {
				if port := parseSSHPortFromNAT(vmInfo); port > 0 {
					vm.SSHPort = port
					// Update the metadata with the new SSH port
					go func(vmCopy *vbtypes.VM) {
						if err := s.storeVMMetadata(context.Background(), vmCopy); err != nil {
							serviceLog.Warnf("Failed to update VM metadata with SSH port: %v", err)
						}
					}(vm)
				}
			}
		}
		return vm, nil
	}

	// Fallback to VBoxManage
	output, err := s.vboxExec.executeCommand("showvminfo", vmId, "--machinereadable")
	if err != nil {
		return nil, fmt.Errorf("VM not found: %w", err)
	}
	vmInfo := s.parseMachineReadableOutput(output)
	name := vmInfo["name"]
	if name == "" {
		return nil, fmt.Errorf("could not find VM name for UUID %s", vmId)
	}
	vm := &vbtypes.VM{
		ID:         vmId,
		Name:       name,
		Status:     parseVMStatus(vmInfo["VMState"]),
		CPUCores:   parseIntOrDefault(vmInfo["cpus"], 1),
		MemoryMB:   parseIntOrDefault(vmInfo["memory"], 1024),
		DiskSizeGB: parseIntOrDefault(vmInfo["storagebytes"], 0) / (1024 * 1024 * 1024),
		VMFolder:   filepath.Join(s.vmDir, name),
	}

	// Set SSH port if available
	if port := parseSSHPortFromNAT(vmInfo); port > 0 {
		vm.SSHPort = port
	}

	// Store in datastore for next time
	go func(vmCopy *vbtypes.VM) {
		if err := s.storeVMMetadata(context.Background(), vmCopy); err != nil {
			serviceLog.Warnf("Failed to store VM metadata: %v", err)
		}
	}(vm)

	return vm, nil
}

// GetVMs gets a list of all VMs
func (s *VirtualboxService) GetVMs(ctx context.Context) ([]*vbtypes.VM, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := time.Now()
	q := query.Query{Prefix: "virtualbox/vm/"}
	results, err := s.datastore.Query(ctx, q)
	if err == nil {
		var vms []*vbtypes.VM
		for result := range results.Next() {
			if result.Error != nil {
				continue
			}
			var vm vbtypes.VM
			if err := json.Unmarshal(result.Value, &vm); err == nil {
				vms = append(vms, &vm)
			}
		}
		_ = results.Close()
		serviceLog.Infof("GetVMs: fetched %d VMs from datastore in %s", len(vms), time.Since(start))
		return vms, len(vms), nil
	}
	serviceLog.Warnf("GetVMs: datastore query failed (%v), falling back to VBoxManage", err)

	// Fallback: Get all VMs using VBoxManageExecutor (legacy/first run)
	output, err := s.vboxExec.executeCommand("list", "vms")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list VMs: %w", err)
	}
	lines := strings.Split(output, "\n")
	var vms []*vbtypes.VM
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		start := strings.Index(line, "\"")
		end := strings.LastIndex(line, "\"")
		brace := strings.LastIndex(line, "{")
		braceEnd := strings.LastIndex(line, "}")
		if start != -1 && end != -1 && brace != -1 && braceEnd != -1 && brace < braceEnd {
			uuid := line[brace+1 : braceEnd]
			vm, err := s.GetVM(ctx, uuid)
			if err != nil {
				serviceLog.Warnf("Failed to get VM info for %s: %v", uuid, err)
				continue
			}
			vms = append(vms, vm)
		}
	}

	// Trigger a background sync to update the datastore for future calls
	go func() {
		if err := s.SyncVMs(context.Background()); err != nil {
			serviceLog.Warnf("Failed to sync VMs after fallback query: %v", err)
		}
	}()

	return vms, len(vms), nil
}

// UpdateVM updates an existing VM using VBoxManage
func (s *VirtualboxService) UpdateVM(ctx context.Context, vmId string, req vbtypes.VMUpdateRequest) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	serviceLog.Infof("Updating VM: %s", vmId)

	vm, err := s.GetVM(ctx, vmId)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM for update: %w", err)
	}

	updated := false

	if req.CPUCores > 0 && req.CPUCores != vm.CPUCores {
		serviceLog.Infof("Updating CPU cores from %d to %d", vm.CPUCores, req.CPUCores)
		if err := s.vboxExec.UpdateCPUCores(vm.ID, req.CPUCores); err != nil {
			return nil, fmt.Errorf("failed to update CPU cores: %w", err)
		}
		vm.CPUCores = req.CPUCores
		updated = true
	}

	if req.MemoryMB > 0 && req.MemoryMB != vm.MemoryMB {
		serviceLog.Infof("Updating memory from %d MB to %d MB", vm.MemoryMB, req.MemoryMB)
		if err := s.vboxExec.UpdateMemory(vm.ID, req.MemoryMB); err != nil {
			return nil, fmt.Errorf("failed to update memory: %w", err)
		}
		vm.MemoryMB = req.MemoryMB
		updated = true
	}

	if req.DiskSizeGB > 0 && req.DiskSizeGB != vm.DiskSizeGB {
		serviceLog.Infof("Updating disk size from %d GB to %d GB", vm.DiskSizeGB, req.DiskSizeGB)
		// Note: actual disk resize not implemented here
		vm.DiskSizeGB = req.DiskSizeGB
		updated = true
	}

	// Store updated VM metadata in datastore
	if updated {
		serviceLog.Infof("Storing updated VM metadata in datastore")
		if err := s.storeVMMetadata(ctx, vm); err != nil {
			serviceLog.Warnf("Failed to store updated VM metadata in datastore: %v", err)
		}
	} else {
		serviceLog.Infof("No changes detected, VM not updated")
	}

	serviceLog.Infof("Successfully updated VM: %s", vmId)
	return vm, nil
}

// DeleteVM deletes a VM using VBoxManage
func (s *VirtualboxService) DeleteVM(ctx context.Context, vmId string) error {
	vm, err := s.GetVM(ctx, vmId)

	// Stop the VM if running, before acquiring the write lock
	if err == nil && vm.Status == vbtypes.Running {
		if _, stopErr := s.StopVM(ctx, vm.ID); stopErr != nil {
			serviceLog.Warnf("Failed to stop VM before deletion: %v", stopErr)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	serviceLog.Infof("Deleting VM: %s", vm.ID)

	if err := s.vboxExec.DeleteVM(vm.ID); err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	// Remove VM from SSHServer
	s.sshServer.RemoveVMConfig(vm.ID)
	serviceLog.Infof("Removed VM %s from SSHServer", vm.ID)

	// Delete VM metadata from datastore
	if err := s.deleteVMMetadata(ctx, vm.ID); err != nil {
		serviceLog.Warnf("Failed to delete VM metadata from datastore: %v", err)
	} else {
		serviceLog.Infof("Successfully removed VM metadata from datastore")
	}

	// Recursively delete the VM's folder (including cloud-init files)
	if vm != nil && vm.VMFolder != "" {
		if removeErr := os.RemoveAll(vm.VMFolder); removeErr != nil {
			serviceLog.Warnf("Failed to delete VM folder %s: %v", vm.VMFolder, removeErr)
		} else {
			serviceLog.Infof("Deleted VM folder: %s", vm.VMFolder)
		}
	}

	serviceLog.Infof("Successfully deleted VM: %s", vmId)
	return nil
}

// StartVM starts a VM using VBoxManage
func (s *VirtualboxService) StartVM(ctx context.Context, vmId string) (*vbtypes.VM, error) {
	serviceLog.Infof("Starting VM: %s", vmId)

	// 1. Set up SSH port forwarding BEFORE starting the VM
	hostPort, err := getAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available port for SSH forwarding: %w", err)
	}
	if err := s.vboxExec.SetupSSHPortForward(vmId, hostPort, 22); err != nil {
		return nil, fmt.Errorf("failed to set up SSH port forwarding: %w", err)
	}

	// 2. Now start the VM
	if err := s.vboxExec.StartVM(vmId, true); err != nil {
		return nil, fmt.Errorf("failed to start VM: %w", err)
	}

	// Get the VM after starting, before acquiring the write lock
	vm, err := s.GetVM(ctx, vmId)
	if err != nil {
		return nil, err
	}
	vm.SSHPort = hostPort
	vm.Status = vbtypes.Running // Explicitly set status to Running

	// Register VM with SSHServer for SSH access
	s.sshServer.AddVMConfig(vm.ID, "localhost", strconv.Itoa(hostPort))
	serviceLog.Infof("Registered VM %s with SSHServer: localhost:%d", vm.ID, hostPort)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.storeVMMetadata(ctx, vm); err != nil {
		serviceLog.Warnf("Failed to store updated VM metadata in datastore: %v", err)
	}
	serviceLog.Infof("Successfully started VM: %s", vmId)
	return vm, nil
}

// StopVM stops a VM using VBoxManage
func (s *VirtualboxService) StopVM(ctx context.Context, vmId string) (*vbtypes.VM, error) {
	serviceLog.Infof("Stopping VM: %s", vmId)
	// Stop the VM before acquiring the write lock
	if err := s.vboxExec.StopVM(vmId); err != nil {
		return nil, fmt.Errorf("failed to stop VM: %w", err)
	}

	// Get the VM after stopping, before acquiring the write lock
	vm, err := s.GetVM(ctx, vmId)
	if err != nil {
		return nil, err
	}

	// Explicitly set status to Stopped
	vm.Status = vbtypes.Stopped

	// Remove VM from SSHServer when stopped
	s.sshServer.RemoveVMConfig(vm.ID)
	serviceLog.Infof("Removed VM %s from SSHServer", vm.ID)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.storeVMMetadata(ctx, vm); err != nil {
		serviceLog.Warnf("Failed to store updated VM metadata in datastore: %v", err)
	}

	serviceLog.Infof("Successfully stopped VM: %s", vmId)
	return vm, nil
}

// PauseVM pauses a VM using VBoxManage
func (s *VirtualboxService) PauseVM(ctx context.Context, vmId string) (*vbtypes.VM, error) {

	serviceLog.Infof("Pausing VM: %s", vmId)
	if err := s.vboxExec.PauseVM(vmId); err != nil {
		return nil, fmt.Errorf("failed to pause VM: %w", err)
	}

	// Get the VM after pausing
	vm, err := s.GetVM(ctx, vmId)
	if err != nil {
		return nil, err
	}

	// Explicitly set status to Paused
	vm.Status = vbtypes.Paused

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.storeVMMetadata(ctx, vm); err != nil {
		serviceLog.Warnf("Failed to store updated VM metadata in datastore: %v", err)
	}

	serviceLog.Infof("Successfully paused VM: %s", vmId)
	return vm, nil
}

// ResumeVM resumes a VM using VBoxManage
func (s *VirtualboxService) ResumeVM(ctx context.Context, vmId string) (*vbtypes.VM, error) {

	serviceLog.Infof("Resuming VM: %s", vmId)
	if err := s.vboxExec.ResumeVM(vmId); err != nil {
		return nil, fmt.Errorf("failed to resume VM: %w", err)
	}

	// Get the VM after resuming
	vm, err := s.GetVM(ctx, vmId)
	if err != nil {
		return nil, err
	}

	// Explicitly set status to Running
	vm.Status = vbtypes.Running

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.storeVMMetadata(ctx, vm); err != nil {
		serviceLog.Warnf("Failed to store updated VM metadata in datastore: %v", err)
	}

	serviceLog.Infof("Successfully resumed VM: %s", vmId)
	return vm, nil
}

// ResetVM resets a VM using VBoxManage
func (s *VirtualboxService) ResetVM(ctx context.Context, vmId string) (*vbtypes.VM, error) {

	serviceLog.Infof("Resetting VM: %s", vmId)
	if err := s.vboxExec.ResetVM(vmId); err != nil {
		return nil, fmt.Errorf("failed to reset VM: %w", err)
	}

	// Get the VM after resetting
	vm, err := s.GetVM(ctx, vmId)
	if err != nil {
		return nil, err
	}

	// Explicitly set status to Running
	vm.Status = vbtypes.Running

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.storeVMMetadata(ctx, vm); err != nil {
		serviceLog.Warnf("Failed to store updated VM metadata in datastore: %v", err)
	}

	serviceLog.Infof("Successfully reset VM: %s", vmId)
	return vm, nil
}

// validateResources checks if the system has sufficient resources to create the VM
func (s *VirtualboxService) validateResources(ctx context.Context, vmCPUCores int, vmMemoryMB int, vmDiskSizeGB int) error {
	serviceLog.Infof("Checking system resources for VM requirements...")

	resourceInfo, err := resource.GetResource()
	if err != nil {
		serviceLog.Warnf("Failed to get system resources, skipping validation: %v", err)
		return nil // Skip validation if we can't get resource info
	}
	// Sum resources of all actually running VMs
	vms, _, err := s.GetVMs(ctx)
	if err != nil {
		serviceLog.Warnf("Failed to get existing VMs, skipping overcommit check: %v", err)
	} else {
		totalCPUs := 0
		totalMem := 0
		totalDisk := 0
		for _, vm := range vms {
			status := vm.Status
			if status != vbtypes.Running {
				// Refresh status from VBoxManage for accuracy
				output, err := s.vboxExec.executeCommand("showvminfo", vm.ID, "--machinereadable")
				if err == nil {
					vmInfo := s.parseMachineReadableOutput(output)
					status = parseVMStatus(vmInfo["VMState"])
				}
			}
			if status == vbtypes.Running {
				totalCPUs += vm.CPUCores
				totalMem += vm.MemoryMB
				totalDisk += vm.DiskSizeGB
			}
		}
		// Add the new VM's requirements
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
	}

	serviceLog.Infof("Validating VM requirements against system resources...")
	serviceLog.Infof("VM Requirements: CPU=%d cores, Memory=%d MB, Disk=%d GB", vmCPUCores, vmMemoryMB, vmDiskSizeGB)
	serviceLog.Infof("System Resources: CPU=%d cores, Memory=%d MB, Disk=%d GB",
		resourceInfo.CPU.Count,
		resourceInfo.Memory.Total/(1024*1024),       // Convert bytes to MB
		resourceInfo.Storage.Total/(1024*1024*1024)) // Convert bytes to GB

	// Validate CPU cores
	if err := s.validateCPUResources(vmCPUCores, resourceInfo); err != nil {
		return fmt.Errorf("CPU validation failed: %w", err)
	}

	// Validate memory
	if err := s.validateMemoryResources(vmMemoryMB, resourceInfo); err != nil {
		return fmt.Errorf("memory validation failed: %w", err)
	}

	// Validate disk space
	if err := s.validateDiskResources(vmDiskSizeGB, resourceInfo); err != nil {
		return fmt.Errorf("disk validation failed: %w", err)
	}

	serviceLog.Infof("Resource validation passed successfully")
	return nil
}

// validateCPUResources validates CPU requirements
func (s *VirtualboxService) validateCPUResources(vmCPUCores int, resourceInfo *resource.ResourceInfo) error {
	// Check if requested CPU cores exceed available cores
	if vmCPUCores > resourceInfo.CPU.Count {
		return fmt.Errorf("insufficient CPU cores: requested %d, available %d", vmCPUCores, resourceInfo.CPU.Count)
	}

	// Check for reasonable CPU allocation (not more than 80% of available cores)
	maxRecommendedCores := int(float64(resourceInfo.CPU.Count) * 0.8)
	if vmCPUCores > maxRecommendedCores {
		serviceLog.Warnf("CPU allocation warning: requested %d cores exceeds recommended maximum of %d cores", vmCPUCores, maxRecommendedCores)
	}

	serviceLog.Infof("CPU validation passed: %d cores requested, %d available", vmCPUCores, resourceInfo.CPU.Count)
	return nil
}

// validateMemoryResources validates memory requirements
func (s *VirtualboxService) validateMemoryResources(vmMemoryMB int, resourceInfo *resource.ResourceInfo) error {
	// Convert memory from bytes to MB
	availableMemoryMB := int(resourceInfo.Memory.Total / (1024 * 1024))

	// Check if requested memory exceeds available memory
	if vmMemoryMB > availableMemoryMB {
		return fmt.Errorf("insufficient memory: requested %d MB, available %d MB", vmMemoryMB, availableMemoryMB)
	}

	// Check for reasonable memory allocation (not more than 80% of available memory)
	maxRecommendedMemoryMB := int(float64(availableMemoryMB) * 0.8)
	if vmMemoryMB > maxRecommendedMemoryMB {
		serviceLog.Warnf("Memory allocation warning: requested %d MB exceeds recommended maximum of %d MB", vmMemoryMB, maxRecommendedMemoryMB)
	}

	// Add 20% buffer for system overhead
	requiredMemoryWithBuffer := int(float64(vmMemoryMB) * 1.2)
	if requiredMemoryWithBuffer > availableMemoryMB {
		return fmt.Errorf("insufficient memory with system buffer: required %d MB, available %d MB", requiredMemoryWithBuffer, availableMemoryMB)
	}

	serviceLog.Infof("Memory validation passed: %d MB requested, %d MB available", vmMemoryMB, availableMemoryMB)
	return nil
}

// validateDiskResources validates disk space requirements
func (s *VirtualboxService) validateDiskResources(vmDiskSizeGB int, resourceInfo *resource.ResourceInfo) error {
	// Convert storage from bytes to GB
	availableDiskGB := int(resourceInfo.Storage.Total / (1024 * 1024 * 1024))

	// Check if requested disk space exceeds available disk space
	if vmDiskSizeGB > availableDiskGB {
		return fmt.Errorf("insufficient disk space: requested %d GB, available %d GB", vmDiskSizeGB, availableDiskGB)
	}

	// Check for reasonable disk allocation (not more than 90% of available disk space)
	maxRecommendedDiskGB := int(float64(availableDiskGB) * 0.9)
	if vmDiskSizeGB > maxRecommendedDiskGB {
		serviceLog.Warnf("Disk allocation warning: requested %d GB exceeds recommended maximum of %d GB", vmDiskSizeGB, maxRecommendedDiskGB)
	}

	// Add 10% buffer for overhead (file system, metadata, etc.)
	requiredDiskWithBuffer := int(float64(vmDiskSizeGB) * 1.1)
	if requiredDiskWithBuffer > availableDiskGB {
		return fmt.Errorf("insufficient disk space with overhead buffer: required %d GB, available %d GB", requiredDiskWithBuffer, availableDiskGB)
	}

	serviceLog.Infof("Disk validation passed: %d GB requested, %d GB available", vmDiskSizeGB, availableDiskGB)
	return nil
}

// SyncVMs manually triggers synchronization between VirtualBox and datastore
func (s *VirtualboxService) SyncVMs(ctx context.Context) error {
	serviceLog.Info("Manually triggering VM synchronization")
	return s.performVMSync(ctx)
}

// syncSSHServerWithRunningVMs syncs the SSHServer with currently running VMs
func (s *VirtualboxService) syncSSHServerWithRunningVMs(ctx context.Context) error {
	serviceLog.Info("Syncing SSHServer with running VMs...")

	// Get all VMs
	vms, _, err := s.GetVMs(ctx)

	if err != nil {
		return fmt.Errorf("failed to get VMs for SSHServer sync: %w", err)
	}

	// Clear existing configurations
	existingConfigs := s.sshServer.GetAllVMConfigs()
	for vmID := range existingConfigs {
		s.sshServer.RemoveVMConfig(vmID)
	}

	// Add configurations for running VMs
	for _, vm := range vms {
		if vm.Status == vbtypes.Running && vm.SSHPort > 0 {
			s.sshServer.AddVMConfig(vm.ID, "127.0.0.1", strconv.Itoa(vm.SSHPort))
			serviceLog.Infof("Synced running VM %s with SSHServer: localhost:%d", vm.ID, vm.SSHPort)
		}
	}

	serviceLog.Infof("SSHServer sync completed. Registered %d running VMs", len(vms))
	return nil
}

// GetSystemInfo gets system information
func (s *VirtualboxService) GetSystemInfo(ctx context.Context) (*vbtypes.VMSystemInfo, error) {
	// Get system resources using the direct resource function
	resourceInfo, err := resource.GetResource()
	if err != nil {
		serviceLog.Warnf("Failed to get system resources: %v", err)
		// Fallback to basic information
		info := &vbtypes.VMSystemInfo{
			HostOS:          runtime.GOOS,
			HostArch:        runtime.GOARCH,
			AvailableCPUs:   runtime.NumCPU(),
			AvailableRAMMB:  0,
			AvailableDiskGB: 0,
		}
		return info, nil
	}

	// Convert resource info to appropriate units
	availableRAMMB := int(resourceInfo.Memory.Total / (1024 * 1024))          // Convert bytes to MB
	availableDiskGB := int(resourceInfo.Storage.Total / (1024 * 1024 * 1024)) // Convert bytes to GB

	info := &vbtypes.VMSystemInfo{
		HostOS:          runtime.GOOS,
		HostArch:        runtime.GOARCH,
		AvailableCPUs:   resourceInfo.CPU.Count,
		AvailableRAMMB:  availableRAMMB,
		AvailableDiskGB: availableDiskGB,
	}

	return info, nil
}

// GetSSHServer returns the SSH server instance
func (s *VirtualboxService) GetSSHServer() *ssh_connection.SSHServer {
	return s.sshServer
}

// GetWebSocketHandler returns the WebSocket handler instance
func (s *VirtualboxService) GetWebSocketHandler() *ssh_connection.WebSocketHandler {
	return s.wsHandler
}

// StoreOrderVMMapping stores the mapping between orderId and vmId
func (s *VirtualboxService) StoreOrderVMMapping(orderId, vmId string) {
	s.orderMapMu.Lock()
	defer s.orderMapMu.Unlock()

	s.orderToVMMap[orderId] = vmId

	fmt.Println("s.orderToVMMap", s.orderToVMMap)

	// Persist to datastore
	go func() {
		if err := s.storeOrderMapping(context.Background(), orderId, vmId); err != nil {
			serviceLog.Warnf("Failed to store order mapping to datastore: %v", err)
		}
	}()

	serviceLog.WithFields(logrus.Fields{
		"orderId": orderId,
		"vmId":    vmId,
	}).Info("Stored order to VM mapping")
}

// GetVMIdByOrderId retrieves the vmId for a given orderId
func (s *VirtualboxService) GetVMIdByOrderId(orderId string) (string, bool) {
	s.orderMapMu.RLock()
	defer s.orderMapMu.RUnlock()

	vmId, exists := s.orderToVMMap[orderId]
	return vmId, exists
}

// RemoveOrderVMMapping removes the mapping for a given orderId
func (s *VirtualboxService) RemoveOrderVMMapping(orderId string) {
	s.orderMapMu.Lock()
	defer s.orderMapMu.Unlock()

	delete(s.orderToVMMap, orderId)

	// Remove from datastore
	go func() {
		if err := s.deleteOrderMapping(context.Background(), orderId); err != nil {
			serviceLog.Warnf("Failed to delete order mapping from datastore: %v", err)
		}
	}()

	serviceLog.WithField("orderId", orderId).Info("Removed order to VM mapping")
}

// GenerateSSHToken generates a one-time access token for SSH connections
func (s *VirtualboxService) GenerateSSHToken(ctx context.Context, vmID string, username string, password string) (*vbtypes.SSHTokenResponse, error) {
	// Validate VM exists and is running
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, fmt.Errorf("VM not found: %w", err)
	}

	if vm.Status != vbtypes.Running {
		return nil, fmt.Errorf("VM is not running")
	}

	if vm.SSHPort == 0 {
		return nil, fmt.Errorf("SSH port not configured for VM")
	}

	// Use the WebSocket handler to generate the token
	tokenResponse, err := s.wsHandler.GenerateSSHToken(vmID, username, password)
	if err != nil {
		return nil, err
	}

	return &vbtypes.SSHTokenResponse{
		Token:     tokenResponse.Token,
		ExpiresAt: tokenResponse.ExpiresAt,
	}, nil
}

// ValidateAndConsumeSSHToken validates a token and returns the credentials if valid
func (s *VirtualboxService) ValidateAndConsumeSSHToken(token string) (*vbtypes.SSHAccessToken, error) {
	accessToken, err := s.wsHandler.ValidateAndConsumeSSHToken(token)
	if err != nil {
		return nil, err
	}

	return &vbtypes.SSHAccessToken{
		Token:     accessToken.Token,
		VMID:      accessToken.VMID,
		Username:  accessToken.Username,
		Password:  accessToken.Password,
		CreatedAt: accessToken.CreatedAt,
		ExpiresAt: accessToken.ExpiresAt,
		Used:      accessToken.Used,
	}, nil
}

// GetJobProgress retrieves the progress of a job
func (s *VirtualboxService) GetJobProgress(ctx context.Context, jobID string) (*vbtypes.Job, error) {
	return s.jobManager.GetJob(ctx, jobID)
}

// ListJobs returns all jobs
func (s *VirtualboxService) ListJobs(ctx context.Context) ([]*vbtypes.Job, error) {
	return s.jobManager.ListJobs(ctx)
}

// CancelJob cancels a job
func (s *VirtualboxService) CancelJob(ctx context.Context, jobID string) error {
	return s.jobManager.CancelJob(ctx, jobID)
}

func generateVMID(name string) string {
	// Simple ID generation - in production, you might want a more sophisticated approach
	return fmt.Sprintf("vm-%s-%d", strings.ToLower(name), time.Now().Unix())
}

// validateVirtualBoxInstallation checks if VirtualBox is properly installed
func (s *VirtualboxService) validateVirtualBoxInstallation() error {
	serviceLog.Infof("Validating VirtualBox installation...")

	// Check if VBoxManage is available
	cmd := exec.Command("VBoxManage", "--version")
	output, err := cmd.Output()
	if err != nil {
		serviceLog.Errorf("VBoxManage not found or not accessible: %v", err)
		return fmt.Errorf("VBoxManage not found or not accessible: %w", err)
	}

	version := strings.TrimSpace(string(output))
	serviceLog.Infof("VirtualBox version: %s", version)

	// Check if VirtualBox is running
	cmd = exec.Command("VBoxManage", "list", "vms")
	if err := cmd.Run(); err != nil {
		serviceLog.Errorf("Failed to list VMs: %v", err)
		return fmt.Errorf("failed to list VMs: %w", err)
	}

	serviceLog.Infof("VirtualBox installation validated successfully")
	return nil
}

// Helper functions for parsing VBoxManage output
func (s *VirtualboxService) parseMachineReadableOutput(output string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse key=value format
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], `"`)
				value := strings.Trim(parts[1], `"`)
				result[key] = value
			}
		}
	}

	return result
}

// Store VM metadata in datastore
func (s *VirtualboxService) storeVMMetadata(ctx context.Context, vm *vbtypes.VM) error {
	data, err := json.Marshal(vm)
	if err != nil {
		return err
	}
	key := datastore.NewKey("virtualbox/vm/" + vm.ID)
	return s.datastore.Put(ctx, key, data)
}

// Store order mapping in datastore
func (s *VirtualboxService) storeOrderMapping(ctx context.Context, orderId, vmId string) error {
	mapping := map[string]string{
		"orderId": orderId,
		"vmId":    vmId,
	}
	data, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	key := datastore.NewKey("virtualbox/order_mapping/" + orderId)
	return s.datastore.Put(ctx, key, data)
}

// Load order mapping from datastore
func (s *VirtualboxService) loadOrderMapping(ctx context.Context, orderId string) (string, error) {
	key := datastore.NewKey("virtualbox/order_mapping/" + orderId)
	data, err := s.datastore.Get(ctx, key)
	if err != nil {
		return "", err
	}
	var mapping map[string]string
	if err := json.Unmarshal(data, &mapping); err != nil {
		return "", err
	}
	return mapping["vmId"], nil
}

// Load all order mappings from datastore
func (s *VirtualboxService) loadAllOrderMappings(ctx context.Context) error {
	q := query.Query{Prefix: "virtualbox/order_mapping/"}
	results, err := s.datastore.Query(ctx, q)
	if err != nil {
		return err
	}
	defer results.Close()

	for result := range results.Next() {
		if result.Error != nil {
			continue
		}
		var mapping map[string]string
		if err := json.Unmarshal(result.Value, &mapping); err != nil {
			continue
		}
		orderId := mapping["orderId"]
		vmId := mapping["vmId"]
		if orderId != "" && vmId != "" {
			s.orderToVMMap[orderId] = vmId
		}
	}

	serviceLog.Infof("Loaded %d order mappings from datastore", len(s.orderToVMMap))
	return nil
}

// Delete order mapping from datastore
func (s *VirtualboxService) deleteOrderMapping(ctx context.Context, orderId string) error {
	key := datastore.NewKey("virtualbox/order_mapping/" + orderId)
	return s.datastore.Delete(ctx, key)
}

// Retrieve VM metadata from datastore
func (s *VirtualboxService) getVMMetadata(ctx context.Context, uuid string) (*vbtypes.VM, error) {
	key := datastore.NewKey("virtualbox/vm/" + uuid)
	data, err := s.datastore.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var vm vbtypes.VM
	if err := json.Unmarshal(data, &vm); err != nil {
		return nil, err
	}
	return &vm, nil
}

// Delete VM metadata from datastore
func (s *VirtualboxService) deleteVMMetadata(ctx context.Context, uuid string) error {
	key := datastore.NewKey("virtualbox/vm/" + uuid)
	return s.datastore.Delete(ctx, key)
}

// syncVMsWithDatastore periodically synchronizes the VirtualBox VM data with the datastore
func (s *VirtualboxService) syncVMsWithDatastore(ctx context.Context) {
	serviceLog.Info("Starting VM synchronization with datastore")

	for {
		select {
		case <-s.syncTicker.C:
			if err := s.performVMSync(ctx); err != nil {
				serviceLog.Errorf("Error during VM synchronization: %v", err)
			}
		case <-s.stopChan:
			serviceLog.Info("Stopping VM synchronization")
			return
		case <-ctx.Done():
			serviceLog.Info("Context cancelled, stopping VM synchronization")
			return
		}
	}
}

// performVMSync does the actual synchronization between VirtualBox and the datastore
func (s *VirtualboxService) performVMSync(ctx context.Context) error {
	serviceLog.Debug("Synchronizing VirtualBox VMs with datastore")

	// Get all VMs directly from VBoxManage
	output, err := s.vboxExec.executeCommand("list", "vms")
	if err != nil {
		return fmt.Errorf("failed to list VMs from VBoxManage: %w", err)
	}

	lines := strings.Split(output, "\n")
	vboxVMs := make(map[string]string) // Map of UUID -> Name

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse VM name and UUID from line like: "VM Name" {uuid}
		nameStart := strings.Index(line, "\"")
		nameEnd := strings.LastIndex(line, "\"")
		uuidStart := strings.LastIndex(line, "{")
		uuidEnd := strings.LastIndex(line, "}")

		if nameStart != -1 && nameEnd != -1 && uuidStart != -1 && uuidEnd != -1 && nameStart < nameEnd && uuidStart < uuidEnd {
			name := line[nameStart+1 : nameEnd]
			uuid := line[uuidStart+1 : uuidEnd]
			vboxVMs[uuid] = name
		}
	}

	// Get all VMs from datastore
	q := query.Query{Prefix: "virtualbox/vm/"}
	results, err := s.datastore.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("failed to query datastore: %w", err)
	}

	datastoreVMs := make(map[string]*vbtypes.VM)
	for result := range results.Next() {
		if result.Error != nil {
			continue
		}

		var vm vbtypes.VM
		if err := json.Unmarshal(result.Value, &vm); err == nil {
			datastoreVMs[vm.ID] = &vm
		}
	}
	_ = results.Close()

	// Lock for writing to datastore
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update existing VMs in datastore with current status
	for uuid, name := range vboxVMs {
		// Get detailed VM info from VBoxManage
		vmOutput, err := s.vboxExec.executeCommand("showvminfo", uuid, "--machinereadable")
		if err != nil {
			serviceLog.Warnf("Failed to get info for VM %s: %v", uuid, err)
			continue
		}

		vmInfo := s.parseMachineReadableOutput(vmOutput)
		vmStatus := parseVMStatus(vmInfo["VMState"])

		// Check if VM exists in datastore
		if existingVM, exists := datastoreVMs[uuid]; exists {
			// Update VM status and other dynamic properties
			existingVM.Status = vmStatus

			// Update SSH port if available
			if port := parseSSHPortFromNAT(vmInfo); port > 0 && existingVM.SSHPort != port {
				existingVM.SSHPort = port
				serviceLog.Debugf("Updated SSH port for VM %s: %d", uuid, port)
			}

			// Store updated VM
			if err := s.storeVMMetadata(ctx, existingVM); err != nil {
				serviceLog.Warnf("Failed to update VM in datastore: %v", err)
			} else {
				serviceLog.Debugf("Updated VM in datastore: %s (%s)", existingVM.Name, existingVM.ID)
			}

			// Remove from datastoreVMs map to track processed VMs
			delete(datastoreVMs, uuid)
		} else {
			// VM exists in VirtualBox but not in datastore, create new entry
			cpuCores := parseIntOrDefault(vmInfo["cpus"], 1)
			memoryMB := parseIntOrDefault(vmInfo["memory"], 1024)

			newVM := &vbtypes.VM{
				ID:       uuid,
				Name:     name,
				Status:   vmStatus,
				CPUCores: cpuCores,
				MemoryMB: memoryMB,
				VMFolder: filepath.Join(s.vmDir, name),
			}

			// Set SSH port if available
			if port := parseSSHPortFromNAT(vmInfo); port > 0 {
				newVM.SSHPort = port
			}

			// Store new VM
			if err := s.storeVMMetadata(ctx, newVM); err != nil {
				serviceLog.Warnf("Failed to store new VM in datastore: %v", err)
			} else {
				serviceLog.Infof("Added new VM to datastore: %s (%s)", newVM.Name, newVM.ID)
			}
		}
	}

	// Remove VMs from datastore that no longer exist in VirtualBox
	for uuid, vm := range datastoreVMs {
		if _, exists := vboxVMs[uuid]; !exists {
			key := datastore.NewKey("virtualbox/vm/" + uuid)
			if err := s.datastore.Delete(ctx, key); err != nil {
				serviceLog.Warnf("Failed to remove VM %s from datastore: %v", uuid, err)
			} else {
				serviceLog.Infof("Removed non-existent VM from datastore: %s (%s)", vm.Name, vm.ID)
			}
		}
	}

	serviceLog.Debug("VM synchronization completed")
	return nil
}

func (s *VirtualboxService) CollectMetrics(ctx context.Context, vmId string, conn *websocket.Conn, period int) error {
	// Add panic recovery to the entire function
	defer func() {
		if r := recover(); r != nil {
			serviceLog.Errorf("Panic in CollectMetrics for VM %s: %v", vmId, r)
		}
	}()

	serviceLog.Infof("Starting metrics collection for VM: %s with period: %d seconds", vmId, period)

	// Check available metrics using VBoxManageExecutor
	checkOutput, err := s.vboxExec.ListAvailableMetrics(vmId)
	if err != nil {
		serviceLog.Errorf("Error checking available metrics: %v", err)
	} else {
		serviceLog.Debugf("Available metrics:\n%s", checkOutput)
	}

	// Define all metrics to collect - based on what we know works
	allMetrics := []string{
		"CPU/Load/User",
		"CPU/Load/Kernel",
		"RAM/Usage/Used",
		"Disk/Usage/Used",
		"Net/Rate/Rx",
		"Net/Rate/Tx",
	}

	// Start VBoxManage metrics collect command using VBoxManageExecutor
	cmd, err := s.vboxExec.StartMetricsCollection(ctx, vmId, allMetrics, period)
	if err != nil {
		return fmt.Errorf("error starting metrics collection: %w", err)
	}

	// Create pipes to capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting VBoxManage: %w", err)
	}

	// Ensure the process is cleaned up when the function returns
	defer func() {
		if cmd.Process != nil {
			serviceLog.Debugf("Cleaning up VBoxManage process for VM: %s", vmId)

			// Try graceful termination first
			if err := cmd.Process.Signal(os.Interrupt); err != nil {
				serviceLog.Debugf("Process already terminated or error sending interrupt: %v", err)
			} else {
				// Give it a moment to terminate gracefully
				time.Sleep(1 * time.Second)
			}

			// Force kill if still running
			if err := cmd.Process.Kill(); err != nil {
				serviceLog.Debugf("Process already terminated or error killing: %v", err)
			}

			// Wait a bit more to ensure process is fully terminated
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Create a context to signal when WebSocket disconnects
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up WebSocket close handler to detect disconnections
	originalCloseHandler := conn.CloseHandler()
	conn.SetCloseHandler(func(code int, text string) error {
		serviceLog.Debugf("WebSocket close handler triggered for VM: %s (code: %d, text: %s)", vmId, code, text)
		cancel() // Cancel context instead of sending to channel
		if originalCloseHandler != nil {
			return originalCloseHandler(code, text)
		}
		return nil
	})

	// Set up pong handler to detect disconnections
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	})

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Safe WebSocket write function with panic recovery
	safeWriteMessage := func(messageType int, data []byte) error {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in WebSocket write for VM %s: %v", vmId, r)
			}
		}()

		// Set a write deadline to prevent hanging
		if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return err
		}

		err := conn.WriteMessage(messageType, data)

		// Reset write deadline
		conn.SetWriteDeadline(time.Time{})

		return err
	}

	// Monitor WebSocket connection in a goroutine with better error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in VM status monitoring goroutine for VM %s: %v", vmId, r)
				cancel() // Cancel context on panic
			}
		}()

		ticker := time.NewTicker(1 * time.Second) // Check every 1 second for faster detection
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				serviceLog.Debugf("Context cancelled, stopping WebSocket monitoring for VM: %s", vmId)
				return
			case <-ticker.C:
				// Check if VM is still running
				vmState, err := s.vboxExec.GetVMStatus(vmId)
				if err != nil {
					serviceLog.Warnf("Failed to get VM status for %s: %v", vmId, err)
					// If we can't get VM status, assume it's stopped and terminate
					serviceLog.Infof("VM %s status check failed, terminating metrics collection", vmId)
					cancel() // Cancel context on error
					return
				} else if vmState != "running" {
					serviceLog.Infof("VM %s is no longer running (status: %s), terminating metrics collection", vmId, vmState)

					// Send error message to WebSocket client before disconnecting
					errorData := map[string]string{
						"type":    "error",
						"message": fmt.Sprintf("VM stopped during metrics collection. Status: %s", vmState),
					}
					jsonData, _ := json.Marshal(errorData)
					if err := safeWriteMessage(websocket.TextMessage, jsonData); err != nil {
						serviceLog.Debugf("Failed to send VM stopped error to WebSocket: %v", err)
					} else {
						// Small delay to ensure message is delivered
						time.Sleep(100 * time.Millisecond)
					}

					cancel() // Cancel context on VM stopped
					return
				}

				// Set a shorter write deadline to detect disconnection faster
				if err := conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
					serviceLog.Debugf("WebSocket disconnected, terminating VBoxManage process for VM: %s", vmId)
					cancel() // Cancel context on WebSocket disconnect
					return
				}

				// Send ping to check connection
				if err := safeWriteMessage(websocket.PingMessage, nil); err != nil {
					serviceLog.Debugf("WebSocket disconnected, terminating VBoxManage process for VM: %s", vmId)
					cancel() // Cancel context on WebSocket disconnect
					return
				}

				// Reset write deadline
				if err := conn.SetWriteDeadline(time.Time{}); err != nil {
					serviceLog.Debugf("WebSocket disconnected, terminating VBoxManage process for VM: %s", vmId)
					cancel() // Cancel context on WebSocket disconnect
					return
				}
			}
		}
	}()

	// Start a background goroutine to actively read from WebSocket to detect disconnections immediately
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in WebSocket read goroutine for VM %s: %v", vmId, r)
				cancel() // Cancel context on panic
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Set read deadline to prevent hanging
			if err := conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
				serviceLog.Debugf("Failed to set read deadline for VM %s: %v", vmId, err)
				cancel() // Cancel context on error
				return
			}

			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					serviceLog.Debugf("WebSocket read error (client disconnected): %v", err)
				} else {
					serviceLog.Debugf("WebSocket closed normally: %v", err)
				}
				cancel() // Cancel context on WebSocket disconnect
				return
			}
		}
	}()

	// Read stdout in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in stdout reading goroutine for VM %s: %v", vmId, r)
				cancel() // Cancel context on panic
			}
		}()

		scanner := bufio.NewScanner(stdout)
		var currentSnapshot *vbtypes.VMMetricsSnapshot
		var currentTimestamp string

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			serviceLog.Debugf("VBoxManage output: %s", line)

			// Parse the line into structured data
			metricData, err := parseVBoxManageOutput(line)
			if err != nil {
				serviceLog.Warnf("Error parsing line: %v", err)
				continue
			}

			// Skip if it's a header or empty line
			if metricData == nil {
				serviceLog.Debugf("Skipping line (header or empty): %s", line)
				continue
			}

			serviceLog.Debugf("Successfully parsed metric: %s = %f %s", metricData.Metric, metricData.Value, metricData.Unit)

			// Check if this is a new timestamp (new snapshot)
			if currentTimestamp != metricData.Timestamp {
				// Send previous snapshot if it exists
				if currentSnapshot != nil {
					jsonData, err := json.Marshal(currentSnapshot)
					if err != nil {
						serviceLog.Errorf("Error marshaling snapshot JSON: %v", err)
					} else {
						// Send complete snapshot to WebSocket client
						if err := safeWriteMessage(websocket.TextMessage, jsonData); err != nil {
							serviceLog.Debugf("WebSocket write error (client likely disconnected): %v", err)
							cancel() // Cancel context on error
							return
						}
						serviceLog.Debugf("Sent complete snapshot for timestamp %s with %d metrics", currentSnapshot.Timestamp, len(currentSnapshot.Metrics))
					}
				}

				// Start new snapshot
				currentTimestamp = metricData.Timestamp
				currentSnapshot = &vbtypes.VMMetricsSnapshot{
					Timestamp: metricData.Timestamp,
					VMId:      vmId,
					Metrics:   make(map[string]vbtypes.VMMetricValue),
				}
			}

			// Add metric to current snapshot (only value and unit, no redundant timestamp/metric name)
			currentSnapshot.Metrics[metricData.Metric] = vbtypes.VMMetricValue{
				Value: metricData.Value,
				Unit:  metricData.Unit,
			}
		}

		// Send the last snapshot if it exists
		if currentSnapshot != nil {
			jsonData, err := json.Marshal(currentSnapshot)
			if err != nil {
				serviceLog.Errorf("Error marshaling final snapshot JSON: %v", err)
			} else {
				if err := safeWriteMessage(websocket.TextMessage, jsonData); err != nil {
					serviceLog.Debugf("WebSocket write error on final snapshot (client likely disconnected): %v", err)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			serviceLog.Errorf("Error reading stdout: %v", err)
		}
	}()

	// Read stderr in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in stderr reading goroutine for VM %s: %v", vmId, r)
				cancel() // Cancel context on panic
			}
		}()

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()

			// Send error messages to WebSocket client as well
			errorData := map[string]string{
				"type":    "error",
				"message": line,
			}
			jsonData, _ := json.Marshal(errorData)
			if err := safeWriteMessage(websocket.TextMessage, jsonData); err != nil {
				serviceLog.Debugf("WebSocket write error on stderr (client likely disconnected): %v", err)
				cancel() // Cancel context on error
				return
			}
		}
		if err := scanner.Err(); err != nil {
			serviceLog.Errorf("Error reading stderr: %v", err)
		}
	}()

	// Wait for either completion or WebSocket disconnect
	select {
	case <-ctx.Done():
		serviceLog.Infof("WebSocket disconnected, terminating VBoxManage metrics collection for VM: %s", vmId)
	case <-time.After(24 * time.Hour): // Safety timeout
		serviceLog.Infof("Safety timeout reached, terminating VBoxManage metrics collection for VM: %s", vmId)
	}

	// Wait for the process to finish with timeout
	processDone := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in process wait goroutine for VM %s: %v", vmId, r)
			}
		}()
		processDone <- cmd.Wait()
	}()

	select {
	case err := <-processDone:
		if err != nil {
			serviceLog.Debugf("VBoxManage process finished with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		serviceLog.Debugf("Timeout waiting for VBoxManage process to finish for VM: %s", vmId)
	}

	serviceLog.Infof("Metrics collection stopped for VM: %s", vmId)
	return nil
}

// getAvailablePort finds an available TCP port on the host
func getAvailablePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	addr := l.Addr().String()
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, err
	}
	return port, nil
}

// Helper to parse SSH port from NAT rules in VBoxManage output
func parseSSHPortFromNAT(vmInfo map[string]string) int {
	for k, v := range vmInfo {
		if strings.HasPrefix(k, "Forwarding(") && strings.Contains(v, "ssh") {
			// Format: "ssh,tcp,,hostport,,guestport"
			parts := strings.Split(v, ",")
			if len(parts) >= 5 {
				hostPort, err := strconv.Atoi(parts[3])
				if err == nil {
					return hostPort
				}
			}
		}
	}
	return 0
}

func parseVMStatus(vmState string) vbtypes.VMStatus {
	switch strings.ToLower(vmState) {
	case "running":
		return vbtypes.Running
	case "paused":
		return vbtypes.Paused
	case "poweroff", "saved", "aborted":
		return vbtypes.Stopped
	default:
		return vbtypes.Unknown
	}
}

func parseIntOrDefault(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return num
}

// metricParseData is a temporary struct for parsing VBoxManage output
type metricParseData struct {
	Timestamp string
	Metric    string
	Value     float64
	Unit      string
}

func parseVBoxManageOutput(line string) (*metricParseData, error) {
	// Skip header lines and empty lines
	line = strings.TrimSpace(line)
	if line == "" || strings.Contains(line, "----") {
		return nil, nil
	}

	// Skip header lines that don't start with a timestamp
	// VBoxManage metrics output starts with timestamp like "04:19:01.615"
	if !regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3}`).MatchString(line) {
		// This is not a metric data line, skip it
		return nil, nil
	}

	// Parse the format: "04:19:01.615 vm-1       CPU/Load/User        5.11%" or "04:19:01.615 vm-1       RAM/Usage/Used       118816 kB"
	// Using regex to handle variable spacing and different units including kB, MB, B/s
	re := regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}\.\d{3})\s+(\S+)\s+([^\s]+(?:\s+[^\s]+)*)\s+([\d.]+)\s*([%]|[kK]?[bB]|[mM]?[bB]|[gG]?[bB]|[bB]/s|[kK][bB]/s|[mM][bB]/s|[gG][bB]/s)?$`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 6 {
		return nil, fmt.Errorf("could not parse line: %s", line)
	}

	timestamp := matches[1]
	metric := strings.TrimSpace(matches[3])
	valueStr := matches[4]
	unit := matches[5]

	// Convert value to float64
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse value '%s': %v", valueStr, err)
	}

	// Convert timestamp to full datetime
	now := time.Now()
	timeStr := fmt.Sprintf("%04d-%02d-%02d %s", now.Year(), now.Month(), now.Day(), timestamp)
	parsedTime, err := time.Parse("2006-01-02 15:04:05.000", timeStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse timestamp '%s': %v", timeStr, err)
	}

	// Normalize units for better consistency
	normalizedUnit := normalizeUnit(unit)

	return &metricParseData{
		Timestamp: parsedTime.Format(time.RFC3339),
		Metric:    metric,
		Value:     value,
		Unit:      normalizedUnit,
	}, nil
}

func normalizeUnit(unit string) string {
	switch strings.ToLower(unit) {
	case "%":
		return "%"
	case "kb", "k":
		return "KB"
	case "mb", "m":
		return "MB"
	case "gb", "g":
		return "GB"
	case "b/s":
		return "B/s"
	case "kb/s", "k/s":
		return "KB/s"
	case "mb/s", "m/s":
		return "MB/s"
	case "gb/s", "g/s":
		return "GB/s"
	default:
		return unit
	}
}
