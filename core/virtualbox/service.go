package virtualbox

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/json"

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
func (s *VirtualboxService) GetVM(ctx context.Context, uuid string) (*vbtypes.VM, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try datastore first
	if vm, err := s.getVMMetadata(ctx, uuid); err == nil && vm != nil {
		// Always get the latest status from VBoxManage
		output, err := s.vboxExec.executeCommand("showvminfo", uuid, "--machinereadable")
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
	output, err := s.vboxExec.executeCommand("showvminfo", uuid, "--machinereadable")
	if err != nil {
		return nil, fmt.Errorf("VM not found: %w", err)
	}
	vmInfo := s.parseMachineReadableOutput(output)
	name := vmInfo["name"]
	if name == "" {
		return nil, fmt.Errorf("could not find VM name for UUID %s", uuid)
	}
	vm := &vbtypes.VM{
		ID:         uuid,
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
func (s *VirtualboxService) UpdateVM(ctx context.Context, uuid string, req vbtypes.VMUpdateRequest) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	serviceLog.Infof("Updating VM: %s", uuid)

	vm, err := s.GetVM(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM for update: %w", err)
	}

	updated := false

	if req.CPUCores > 0 && req.CPUCores != vm.CPUCores {
		serviceLog.Infof("Updating CPU cores from %d to %d", vm.CPUCores, req.CPUCores)
		if err := s.vboxExec.UpdateCPUCores(uuid, req.CPUCores); err != nil {
			return nil, fmt.Errorf("failed to update CPU cores: %w", err)
		}
		vm.CPUCores = req.CPUCores
		updated = true
	}

	if req.MemoryMB > 0 && req.MemoryMB != vm.MemoryMB {
		serviceLog.Infof("Updating memory from %d MB to %d MB", vm.MemoryMB, req.MemoryMB)
		if err := s.vboxExec.UpdateMemory(uuid, req.MemoryMB); err != nil {
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

	serviceLog.Infof("Successfully updated VM: %s", uuid)
	return vm, nil
}

// DeleteVM deletes a VM using VBoxManage
func (s *VirtualboxService) DeleteVM(ctx context.Context, uuid string) error {
	vm, err := s.GetVM(ctx, uuid)

	// Stop the VM if running, before acquiring the write lock
	if err == nil && vm.Status == vbtypes.Running {
		if _, stopErr := s.StopVM(ctx, uuid); stopErr != nil {
			serviceLog.Warnf("Failed to stop VM before deletion: %v", stopErr)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	serviceLog.Infof("Deleting VM: %s", uuid)

	if err := s.vboxExec.DeleteVM(uuid); err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	// Remove VM from SSHServer
	s.sshServer.RemoveVMConfig(uuid)
	serviceLog.Infof("Removed VM %s from SSHServer", uuid)

	// Delete VM metadata from datastore
	if err := s.deleteVMMetadata(ctx, uuid); err != nil {
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

	serviceLog.Infof("Successfully deleted VM: %s", uuid)
	return nil
}

// StartVM starts a VM using VBoxManage
func (s *VirtualboxService) StartVM(ctx context.Context, uuid string) (*vbtypes.VM, error) {
	serviceLog.Infof("Starting VM: %s", uuid)
	// 1. Set up SSH port forwarding BEFORE starting the VM
	hostPort, err := getAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available port for SSH forwarding: %w", err)
	}
	if err := s.vboxExec.SetupSSHPortForward(uuid, hostPort, 22); err != nil {
		return nil, fmt.Errorf("failed to set up SSH port forwarding: %w", err)
	}

	// 2. Now start the VM
	if err := s.vboxExec.StartVM(uuid, true); err != nil {
		return nil, fmt.Errorf("failed to start VM: %w", err)
	}

	// Get the VM after starting, before acquiring the write lock
	vm, err := s.GetVM(ctx, uuid)
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
	serviceLog.Infof("Successfully started VM: %s", uuid)
	return vm, nil
}

// StopVM stops a VM using VBoxManage
func (s *VirtualboxService) StopVM(ctx context.Context, uuid string) (*vbtypes.VM, error) {
	serviceLog.Infof("Stopping VM: %s", uuid)
	// Stop the VM before acquiring the write lock
	if err := s.vboxExec.StopVM(uuid); err != nil {
		return nil, fmt.Errorf("failed to stop VM: %w", err)
	}

	// Get the VM after stopping, before acquiring the write lock
	vm, err := s.GetVM(ctx, uuid)
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

	serviceLog.Infof("Successfully stopped VM: %s", uuid)
	return vm, nil
}

// PauseVM pauses a VM using VBoxManage
func (s *VirtualboxService) PauseVM(ctx context.Context, uuid string) (*vbtypes.VM, error) {

	serviceLog.Infof("Pausing VM: %s", uuid)
	if err := s.vboxExec.PauseVM(uuid); err != nil {
		return nil, fmt.Errorf("failed to pause VM: %w", err)
	}

	// Get the VM after pausing
	vm, err := s.GetVM(ctx, uuid)
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

	serviceLog.Infof("Successfully paused VM: %s", uuid)
	return vm, nil
}

// ResumeVM resumes a VM using VBoxManage
func (s *VirtualboxService) ResumeVM(ctx context.Context, uuid string) (*vbtypes.VM, error) {

	serviceLog.Infof("Resuming VM: %s", uuid)
	if err := s.vboxExec.ResumeVM(uuid); err != nil {
		return nil, fmt.Errorf("failed to resume VM: %w", err)
	}

	// Get the VM after resuming
	vm, err := s.GetVM(ctx, uuid)
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

	serviceLog.Infof("Successfully resumed VM: %s", uuid)
	return vm, nil
}

// ResetVM resets a VM using VBoxManage
func (s *VirtualboxService) ResetVM(ctx context.Context, uuid string) (*vbtypes.VM, error) {

	serviceLog.Infof("Resetting VM: %s", uuid)
	if err := s.vboxExec.ResetVM(uuid); err != nil {
		return nil, fmt.Errorf("failed to reset VM: %w", err)
	}

	// Get the VM after resetting
	vm, err := s.GetVM(ctx, uuid)
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

	serviceLog.Infof("Successfully reset VM: %s", uuid)
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
