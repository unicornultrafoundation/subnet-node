package virtualbox

import (
	"context"
	"fmt"
	"math/big"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var serviceLog = logrus.WithField("service", "virtualbox")

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	mu         sync.RWMutex
	storageMgr StorageManager
	vmDir      string
	stopChan   chan struct{}
	vboxExec   *VBoxManageExecutor
}

// NewService creates a new VirtualBox service
func NewService() (*ServiceImpl, error) {
	// Create storage manager
	storageMgr, err := NewStorageManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create storage manager: %w", err)
	}

	// Get VM directory
	vmDir, err := fsutil.ExpandHome("~/VirtualBox VMs")
	if err != nil {
		return nil, fmt.Errorf("failed to expand VM directory path: %w", err)
	}

	service := &ServiceImpl{
		storageMgr: storageMgr,
		vmDir:      vmDir,
		stopChan:   make(chan struct{}),
		vboxExec:   NewVBoxManageExecutor(vmDir),
	}

	// Validate VirtualBox installation once during service creation
	if err := service.validateVirtualBoxInstallation(); err != nil {
		return nil, fmt.Errorf("VirtualBox validation failed: %w", err)
	}

	// Validate cloud-init templates
	if err := service.vboxExec.templateMgr.ValidateTemplates(); err != nil {
		return nil, fmt.Errorf("cloud-init template validation failed: %w", err)
	}

	return service, nil
}

// IsVirtualBoxEnabled checks if VirtualBox service is enabled in the configuration
func IsVirtualBoxEnabled(cfg *config.C) bool {
	return cfg.GetBool("virtualbox.enable", false)
}

// Start starts the VirtualBox service
func (s *ServiceImpl) Start(ctx context.Context) error {
	serviceLog.Info("Starting VirtualBox service")
	serviceLog.Info("VirtualBox service started successfully")
	return nil
}

// Stop stops the VirtualBox service
func (s *ServiceImpl) Stop(ctx context.Context) error {
	serviceLog.Info("Stopping VirtualBox service")
	close(s.stopChan)
	return nil
}

// CreateVM creates a new VM with the specified configuration using VBoxManage
func (s *ServiceImpl) CreateVM(ctx context.Context, req vbtypes.VMCreateRequest) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	serviceLog.Infof("Creating VM: %s", req.Name)

	// Validate system resources before creating VM
	serviceLog.Infof("Validating system resources...")
	if err := s.validateResources(ctx, req); err != nil {
		return nil, fmt.Errorf("resource validation failed: %w", err)
	}
	serviceLog.Infof("Resource validation passed")

	// Validate OS type compatibility with hardware if provided
	if req.OSType != "" {
		serviceLog.Infof("Validating OS type compatibility: %s", req.OSType)
		if err := s.validateOSTypeCompatibility(req.OSType); err != nil {
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

	// Configure VM hardware using VBoxManage
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

	// Create and attach storage using VBoxManage
	serviceLog.Infof("Setting up VM storage...")

	// Generate cloud-init ISO
	cloudInitISO := ""
	cloudInitISO, err = s.generateCloudInitISO(req.Name, req.Username, req.Password)
	if err != nil {
		serviceLog.Errorf("Failed to generate cloud-init ISO: %v", err)
		// Continue without cloud-init ISO - it's not critical for VM creation
	}

	if err := s.vboxExec.SetupStorage(req.Name, req, isoPath, cloudInitISO); err != nil {
		serviceLog.Errorf("Failed to setup VM storage: %v", err)
		// Clean up on failure
		serviceLog.Infof("Cleaning up failed VM...")
		if delErr := s.vboxExec.DeleteVM(req.Name); delErr != nil {
			serviceLog.Errorf("Failed to delete VM during cleanup: %v", delErr)
		}
		return nil, fmt.Errorf("failed to setup VM storage: %w", err)
	}
	serviceLog.Infof("VM storage setup completed")

	// Create VM object
	vm := &vbtypes.VM{
		ID:         vmID,
		Name:       req.Name,
		Status:     vbtypes.Stopped,
		CPUCores:   req.CPUCores,
		MemoryMB:   req.MemoryMB,
		DiskSizeGB: req.DiskSizeGB,
		ISOURL:     isoURL,
		ISOPath:    isoPath,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		VBoxPath:   "VBoxManage",
		VMFolder:   vmFolder,
	}

	serviceLog.Infof("Successfully created VM: %s", req.Name)
	return vm, nil
}

// validateResources checks if the system has sufficient resources to create the VM
func (s *ServiceImpl) validateResources(ctx context.Context, req vbtypes.VMCreateRequest) error {
	serviceLog.Infof("Checking system resources for VM requirements...")

	// Get detailed resource information
	resourceInfo, err := resource.GetResource()
	if err != nil {
		serviceLog.Warnf("Failed to get system resources, skipping validation: %v", err)
		return nil // Skip validation if we can't get resource info
	}

	serviceLog.Infof("Validating VM requirements against system resources...")
	serviceLog.Infof("VM Requirements: CPU=%d cores, Memory=%d MB, Disk=%d GB", req.CPUCores, req.MemoryMB, req.DiskSizeGB)
	serviceLog.Infof("System Resources: CPU=%d cores, Memory=%d MB, Disk=%d GB",
		resourceInfo.CPU.Count,
		resourceInfo.Memory.Total/(1024*1024),       // Convert bytes to MB
		resourceInfo.Storage.Total/(1024*1024*1024)) // Convert bytes to GB

	// Validate CPU cores
	if err := s.validateCPUResources(req, resourceInfo); err != nil {
		return fmt.Errorf("CPU validation failed: %w", err)
	}

	// Validate memory
	if err := s.validateMemoryResources(req, resourceInfo); err != nil {
		return fmt.Errorf("memory validation failed: %w", err)
	}

	// Validate disk space
	if err := s.validateDiskResources(req, resourceInfo); err != nil {
		return fmt.Errorf("disk validation failed: %w", err)
	}

	// TODO: Check existing VMs to ensure we don't overcommit resources

	serviceLog.Infof("Resource validation passed successfully")
	return nil
}

// validateCPUResources validates CPU requirements
func (s *ServiceImpl) validateCPUResources(req vbtypes.VMCreateRequest, resourceInfo *resource.ResourceInfo) error {
	// Check if requested CPU cores exceed available cores
	if req.CPUCores > resourceInfo.CPU.Count {
		return fmt.Errorf("insufficient CPU cores: requested %d, available %d", req.CPUCores, resourceInfo.CPU.Count)
	}

	// Check for reasonable CPU allocation (not more than 80% of available cores)
	maxRecommendedCores := int(float64(resourceInfo.CPU.Count) * 0.8)
	if req.CPUCores > maxRecommendedCores {
		serviceLog.Warnf("CPU allocation warning: requested %d cores exceeds recommended maximum of %d cores", req.CPUCores, maxRecommendedCores)
	}

	serviceLog.Infof("CPU validation passed: %d cores requested, %d available", req.CPUCores, resourceInfo.CPU.Count)
	return nil
}

// validateMemoryResources validates memory requirements
func (s *ServiceImpl) validateMemoryResources(req vbtypes.VMCreateRequest, resourceInfo *resource.ResourceInfo) error {
	// Convert memory from bytes to MB
	availableMemoryMB := int(resourceInfo.Memory.Total / (1024 * 1024))

	// Check if requested memory exceeds available memory
	if req.MemoryMB > availableMemoryMB {
		return fmt.Errorf("insufficient memory: requested %d MB, available %d MB", req.MemoryMB, availableMemoryMB)
	}

	// Check for reasonable memory allocation (not more than 80% of available memory)
	maxRecommendedMemoryMB := int(float64(availableMemoryMB) * 0.8)
	if req.MemoryMB > maxRecommendedMemoryMB {
		serviceLog.Warnf("Memory allocation warning: requested %d MB exceeds recommended maximum of %d MB", req.MemoryMB, maxRecommendedMemoryMB)
	}

	// Add 20% buffer for system overhead
	requiredMemoryWithBuffer := int(float64(req.MemoryMB) * 1.2)
	if requiredMemoryWithBuffer > availableMemoryMB {
		return fmt.Errorf("insufficient memory with system buffer: required %d MB, available %d MB", requiredMemoryWithBuffer, availableMemoryMB)
	}

	serviceLog.Infof("Memory validation passed: %d MB requested, %d MB available", req.MemoryMB, availableMemoryMB)
	return nil
}

// validateDiskResources validates disk space requirements
func (s *ServiceImpl) validateDiskResources(req vbtypes.VMCreateRequest, resourceInfo *resource.ResourceInfo) error {
	// Convert storage from bytes to GB
	availableDiskGB := int(resourceInfo.Storage.Total / (1024 * 1024 * 1024))

	// Check if requested disk space exceeds available disk space
	if req.DiskSizeGB > availableDiskGB {
		return fmt.Errorf("insufficient disk space: requested %d GB, available %d GB", req.DiskSizeGB, availableDiskGB)
	}

	// Check for reasonable disk allocation (not more than 90% of available disk space)
	maxRecommendedDiskGB := int(float64(availableDiskGB) * 0.9)
	if req.DiskSizeGB > maxRecommendedDiskGB {
		serviceLog.Warnf("Disk allocation warning: requested %d GB exceeds recommended maximum of %d GB", req.DiskSizeGB, maxRecommendedDiskGB)
	}

	// Add 10% buffer for overhead (file system, metadata, etc.)
	requiredDiskWithBuffer := int(float64(req.DiskSizeGB) * 1.1)
	if requiredDiskWithBuffer > availableDiskGB {
		return fmt.Errorf("insufficient disk space with overhead buffer: required %d GB, available %d GB", requiredDiskWithBuffer, availableDiskGB)
	}

	serviceLog.Infof("Disk validation passed: %d GB requested, %d GB available", req.DiskSizeGB, availableDiskGB)
	return nil
}

// validateOSTypeCompatibility validates if the provided OS type is compatible with the detected hardware
func (s *ServiceImpl) validateOSTypeCompatibility(requestedOSType string) error {
	serviceLog.Infof("Validating OS type compatibility for: %s", requestedOSType)

	// Detect hardware to determine appropriate OS type
	detector := NewHardwareDetector()
	hardware, err := detector.DetectHardware()
	if err != nil {
		serviceLog.Warnf("Hardware detection failed, skipping OS type validation: %v", err)
		return nil // Skip validation if we can't detect hardware
	}

	// Get the appropriate OS type for the detected hardware
	appropriateOSType := detector.determineOSType(hardware)
	serviceLog.Infof("Detected hardware architecture: %s", hardware.Architecture)
	serviceLog.Infof("Appropriate OS type for hardware: %s", appropriateOSType)
	serviceLog.Infof("Requested OS type: %s", requestedOSType)

	// Check if the requested OS type is compatible with the hardware architecture
	if !s.isOSTypeCompatible(requestedOSType, hardware.Architecture) {
		return fmt.Errorf("OS type '%s' is not compatible with hardware architecture '%s'. Recommended OS type: '%s'",
			requestedOSType, hardware.Architecture, appropriateOSType)
	}

	serviceLog.Infof("OS type validation passed: %s is compatible with %s architecture", requestedOSType, hardware.Architecture)
	return nil
}

// isOSTypeCompatible checks if an OS type is compatible with a given architecture
func (s *ServiceImpl) isOSTypeCompatible(osType, architecture string) bool {
	// Define compatibility matrix
	compatibilityMap := map[string][]string{
		"Ubuntu_ARM64": {"arm64", "aarch64"},
		"Ubuntu_64":    {"amd64", "x86_64"},
		"Ubuntu":       {"arm", "386", "i386", "amd64", "x86_64", "arm64", "aarch64"},
		"Debian_ARM64": {"arm64", "aarch64"},
		"Debian_64":    {"amd64", "x86_64"},
		"Debian":       {"arm", "386", "i386", "amd64", "x86_64", "arm64", "aarch64"},
		// "Windows_ARM64": {"arm64", "aarch64"},
		// "Windows_64":    {"amd64", "x86_64"},
		// "Windows":       {"amd64", "x86_64"},
	}

	// Check if the OS type is in our compatibility map
	supportedArchitectures, exists := compatibilityMap[osType]
	if !exists {
		serviceLog.Warnf("Unknown OS type: %s, allowing it to pass validation", osType)
		return true // Allow unknown OS types to pass validation
	}

	// Check if the architecture is supported by this OS type
	for _, supportedArch := range supportedArchitectures {
		if supportedArch == architecture {
			return true
		}
	}

	return false
}

// GetVM gets a VM by ID using VBoxManage
func (s *ServiceImpl) GetVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// For now, we'll use the VM name as ID
	vmName := vmID

	// Get VM info from VirtualBox using VBoxManageExecutor
	output, err := s.vboxExec.executeCommand("showvminfo", vmName, "--machinereadable")
	if err != nil {
		return nil, fmt.Errorf("VM not found: %w", err)
	}

	// Parse the machine-readable output
	vmInfo := s.parseMachineReadableOutput(output)

	// Create VM object from machine info
	vm := &vbtypes.VM{
		ID:       vmID,
		Name:     vmName,
		Status:   s.parseVMStatus(vmInfo["VMState"]),
		CPUCores: s.parseIntOrDefault(vmInfo["cpus"], 1),
		MemoryMB: s.parseIntOrDefault(vmInfo["memory"], 1024),
		VBoxPath: "VBoxManage",
		VMFolder: filepath.Join(s.vmDir, vmName),
	}

	return vm, nil
}

// GetVMs gets a list of all VMs
func (s *ServiceImpl) GetVMs(ctx context.Context) ([]*vbtypes.VM, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all VMs using VBoxManageExecutor
	vmNames, err := s.vboxExec.ListVMs()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list VMs: %w", err)
	}

	var vms []*vbtypes.VM
	for _, vmName := range vmNames {
		// Get detailed info for each VM
		vm, err := s.GetVM(ctx, vmName)
		if err != nil {
			serviceLog.Warnf("Failed to get VM info for %s: %v", vmName, err)
			continue
		}

		vms = append(vms, vm)
	}

	return vms, len(vms), nil
}

// GetVMCount gets the count of VMs with optional filtering
func (s *ServiceImpl) GetVMCount(ctx context.Context, filter *vbtypes.VMFilter) (*big.Int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vmNames, err := s.vboxExec.ListVMs()
	if err != nil {
		return big.NewInt(0), fmt.Errorf("failed to list VMs: %w", err)
	}

	if filter == nil {
		return big.NewInt(int64(len(vmNames))), nil
	}

	count := 0
	for _, vmName := range vmNames {
		vm, err := s.GetVM(ctx, vmName)
		if err != nil {
			continue
		}

		// Apply filters
		if filter.Status != "" && string(vm.Status) != filter.Status {
			continue
		}
		if filter.Query != "" && !strings.Contains(strings.ToLower(vm.Name), strings.ToLower(filter.Query)) {
			continue
		}

		count++
	}

	return big.NewInt(int64(count)), nil
}

// UpdateVM updates an existing VM using VBoxManage
func (s *ServiceImpl) UpdateVM(ctx context.Context, vmID string, req vbtypes.VMUpdateRequest) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	// Get current VM
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	// Update VM parameters using VBoxManage
	if req.CPUCores > 0 && req.CPUCores != vm.CPUCores {
		if err := s.vboxExec.UpdateCPUCores(vmName, req.CPUCores); err != nil {
			return nil, err
		}
		vm.CPUCores = req.CPUCores
	}

	if req.MemoryMB > 0 && req.MemoryMB != vm.MemoryMB {
		if err := s.vboxExec.UpdateMemory(vmName, req.MemoryMB); err != nil {
			return nil, err
		}
		vm.MemoryMB = req.MemoryMB
	}

	vm.UpdatedAt = time.Now()
	return vm, nil
}

// DeleteVM deletes a VM using VBoxManage
func (s *ServiceImpl) DeleteVM(ctx context.Context, vmID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Deleting VM: %s", vmName)

	// Get the machine
	vm, err := s.GetVM(ctx, vmID)
	if err == nil && vm.Status == vbtypes.Running {
		if _, stopErr := s.StopVM(ctx, vmID); stopErr != nil {
			serviceLog.Warnf("Failed to stop VM before deletion: %v", stopErr)
		}
	}

	// Delete VM using VBoxManage executor
	if err := s.vboxExec.DeleteVM(vmName); err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	serviceLog.Infof("Successfully deleted VM: %s", vmName)
	return nil
}

// StartVM starts a VM using VBoxManage
func (s *ServiceImpl) StartVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Starting VM: %s", vmName)

	// Start the VM using VBoxManage executor
	if err := s.vboxExec.StartVM(vmName, true); err != nil {
		return nil, fmt.Errorf("failed to start VM: %w", err)
	}

	// Get updated VM info
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	serviceLog.Infof("Successfully started VM: %s", vmName)
	return vm, nil
}

// StopVM stops a VM using VBoxManage
func (s *ServiceImpl) StopVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Stopping VM: %s", vmName)

	// Stop the VM using VBoxManage executor
	if err := s.vboxExec.StopVM(vmName); err != nil {
		return nil, fmt.Errorf("failed to stop VM: %w", err)
	}

	// Get updated VM info
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	serviceLog.Infof("Successfully stopped VM: %s", vmName)
	return vm, nil
}

// PauseVM pauses a VM using VBoxManage
func (s *ServiceImpl) PauseVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Pausing VM: %s", vmName)

	// Pause the VM using VBoxManage executor
	if err := s.vboxExec.PauseVM(vmName); err != nil {
		return nil, fmt.Errorf("failed to pause VM: %w", err)
	}

	// Get updated VM info
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	serviceLog.Infof("Successfully paused VM: %s", vmName)
	return vm, nil
}

// ResumeVM resumes a VM using VBoxManage
func (s *ServiceImpl) ResumeVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Resuming VM: %s", vmName)

	// Resume the VM using VBoxManage executor
	if err := s.vboxExec.ResumeVM(vmName); err != nil {
		return nil, fmt.Errorf("failed to resume VM: %w", err)
	}

	// Get updated VM info
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	serviceLog.Infof("Successfully resumed VM: %s", vmName)
	return vm, nil
}

// ResetVM resets a VM using VBoxManage
func (s *ServiceImpl) ResetVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Resetting VM: %s", vmName)

	// Reset the VM using VBoxManage executor
	if err := s.vboxExec.ResetVM(vmName); err != nil {
		return nil, fmt.Errorf("failed to reset VM: %w", err)
	}

	// Get updated VM info
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	serviceLog.Infof("Successfully reset VM: %s", vmName)
	return vm, nil
}

// GetVMUsage gets resource usage for a specific VM
func (s *ServiceImpl) GetVMUsage(ctx context.Context, vmID string) (*vbtypes.VMUsage, error) {
	// This is a placeholder implementation
	// In a real implementation, you would collect actual resource usage data
	return &vbtypes.VMUsage{
		VMID:          vmID,
		CPUPerc:       0.0,
		MemoryMB:      0,
		DiskUsageGB:   0.0,
		NetworkRxMB:   0.0,
		NetworkTxMB:   0.0,
		UptimeSeconds: 0,
		Timestamp:     time.Now(),
	}, nil
}

// GetAllVMUsage gets resource usage for all VMs
func (s *ServiceImpl) GetAllVMUsage(ctx context.Context) (*vbtypes.VMUsage, error) {
	// This is a placeholder implementation
	return &vbtypes.VMUsage{
		VMID:          "all",
		CPUPerc:       0.0,
		MemoryMB:      0,
		DiskUsageGB:   0.0,
		NetworkRxMB:   0.0,
		NetworkTxMB:   0.0,
		UptimeSeconds: 0,
		Timestamp:     time.Now(),
	}, nil
}

// GetSystemInfo gets system information
func (s *ServiceImpl) GetSystemInfo(ctx context.Context) (*vbtypes.VMSystemInfo, error) {
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

// DownloadISO downloads an ISO file
func (s *ServiceImpl) DownloadISO(ctx context.Context, isoURL string) (*vbtypes.ISOInfo, error) {
	isoPath := filepath.Join(s.vmDir, "ISOs", filepath.Base(isoURL))

	// Check if already exists
	if s.storageMgr.FileExists(isoPath) {
		return s.storageMgr.GetFileInfo(isoPath)
	}

	// Download the ISO
	if err := s.storageMgr.DownloadFile(ctx, isoURL, isoPath); err != nil {
		return nil, err
	}

	return s.storageMgr.GetFileInfo(isoPath)
}

// GetISOInfo gets information about an ISO file
func (s *ServiceImpl) GetISOInfo(ctx context.Context, isoPath string) (*vbtypes.ISOInfo, error) {
	return s.storageMgr.GetFileInfo(isoPath)
}

// ListISOs lists all available ISO files
func (s *ServiceImpl) ListISOs(ctx context.Context) ([]*vbtypes.ISOInfo, error) {
	isoDir := filepath.Join(s.vmDir, "ISOs")
	files, err := s.storageMgr.ListFiles(isoDir)
	if err != nil {
		return nil, err
	}

	var isos []*vbtypes.ISOInfo
	for _, file := range files {
		if strings.HasSuffix(file, ".iso") {
			isoInfo, err := s.storageMgr.GetFileInfo(file)
			if err == nil {
				isos = append(isos, isoInfo)
			}
		}
	}

	return isos, nil
}

// DeleteISO deletes an ISO file
func (s *ServiceImpl) DeleteISO(ctx context.Context, isoPath string) error {
	return s.storageMgr.DeleteFile(isoPath)
}

// ListOSTypes lists all available OS types that can run on the current machine
func (s *ServiceImpl) ListOSTypes(ctx context.Context) ([]string, error) {
	return s.storageMgr.GetSupportedOSTypes(), nil
}

// ensureISO ensures an ISO file is available locally
func (s *ServiceImpl) ensureISO(ctx context.Context, isoURL string) (string, error) {
	if isoURL == "" {
		return "", nil
	}

	// Check if ISO already exists
	isoDir := filepath.Join(s.vmDir, "ISOs")
	isoName := filepath.Base(isoURL)
	isoPath := filepath.Join(isoDir, isoName)

	if s.storageMgr.FileExists(isoPath) {
		serviceLog.Infof("ISO already exists: %s", isoPath)
		return isoPath, nil
	}

	// Download ISO
	serviceLog.Infof("Downloading ISO: %s", isoPath)
	if err := s.storageMgr.DownloadFile(ctx, isoURL, isoPath); err != nil {
		return "", err
	}

	return isoPath, nil
}

// configureVMHardwareWithVBoxManage configures VM hardware using VBoxManage with dynamic hardware detection
func (s *ServiceImpl) configureVMHardwareWithVBoxManage(vmName string, req vbtypes.VMCreateRequest, osType string) error {
	serviceLog.Infof("Starting VM hardware configuration with VBoxManage...")

	// Detect hardware and get appropriate settings
	detector := NewHardwareDetector()
	hardware, err := detector.DetectHardware()
	if err != nil {
		serviceLog.Warnf("Hardware detection failed, using fallback settings: %v", err)
		return s.configureVMHardwareWithVBoxManageFallback(vmName, req, osType)
	}

	settings := detector.GetVirtualBoxSettings(hardware)
	serviceLog.Infof("Detected hardware: %+v", hardware)
	serviceLog.Infof("Using VirtualBox settings: %+v", settings)

	// Set OS type
	serviceLog.Infof("Setting OS type: %s", osType)
	if err := s.vboxExec.setOSType(vmName, osType); err != nil {
		return fmt.Errorf("failed to set OS type: %w", err)
	}

	// Set CPU count
	serviceLog.Infof("Setting CPU count to %d", req.CPUCores)
	if err := s.vboxExec.setCPUs(vmName, req.CPUCores); err != nil {
		return fmt.Errorf("failed to set CPU count: %w", err)
	}

	// Set memory
	serviceLog.Infof("Setting memory to %d MB", req.MemoryMB)
	if err := s.vboxExec.setMemory(vmName, req.MemoryMB); err != nil {
		return fmt.Errorf("failed to set memory: %w", err)
	}

	// Set VRAM based on detected hardware
	serviceLog.Infof("Setting VRAM to %d MB", settings.VRAMMB)
	if err := s.vboxExec.setVRAM(vmName, settings.VRAMMB); err != nil {
		return fmt.Errorf("failed to set VRAM: %w", err)
	}

	// Set chipset based on detected hardware
	serviceLog.Infof("Setting chipset to %s", settings.Chipset)
	if err := s.vboxExec.setChipset(vmName, settings.Chipset); err != nil {
		return fmt.Errorf("failed to set chipset: %w", err)
	}

	// Set firmware based on detected hardware
	serviceLog.Infof("Setting firmware to %s", settings.Firmware)
	if err := s.vboxExec.setFirmware(vmName, settings.Firmware); err != nil {
		return fmt.Errorf("failed to set firmware: %w", err)
	}

	// Set graphics controller based on detected hardware
	serviceLog.Infof("Setting graphics controller to %s", settings.GraphicsController)
	if err := s.vboxExec.setGraphicsController(vmName, settings.GraphicsController); err != nil {
		return fmt.Errorf("failed to set graphics controller: %w", err)
	}

	// Set IOAPIC based on detected hardware
	ioapicStatus := settings.IOAPICEnabled
	serviceLog.Infof("Setting IOAPIC to %v", ioapicStatus)
	if err := s.vboxExec.setIOAPIC(vmName, ioapicStatus); err != nil {
		return fmt.Errorf("failed to set IOAPIC: %w", err)
	}

	// Set boot order
	serviceLog.Infof("Setting boot order")
	if err := s.vboxExec.setBootOrder(vmName); err != nil {
		return fmt.Errorf("failed to set boot order: %w", err)
	}

	// Configure input devices
	serviceLog.Infof("Configuring input devices...")
	if err := s.vboxExec.configureInputDevices(vmName); err != nil {
		return fmt.Errorf("failed to configure input devices: %w", err)
	}

	// Configure USB based on detected hardware
	serviceLog.Infof("Configuring USB with controller: %s", settings.USBController)
	if err := s.vboxExec.configureUSBWithSettings(vmName, settings.USBController); err != nil {
		return fmt.Errorf("failed to configure USB: %w", err)
	}

	// Configure audio based on detected hardware
	serviceLog.Infof("Configuring audio with controller: %s, output: %s, input: %s", settings.AudioController, settings.AudioOutput, settings.AudioInput)
	if err := s.vboxExec.configureAudioWithSettings(vmName, settings.AudioController, settings.AudioOutput, settings.AudioInput); err != nil {
		return fmt.Errorf("failed to configure audio: %w", err)
	}

	serviceLog.Infof("VM hardware configuration completed successfully")
	return nil
}

// configureVMHardwareWithVBoxManageFallback provides fallback hardware configuration when detection fails
func (s *ServiceImpl) configureVMHardwareWithVBoxManageFallback(vmName string, req vbtypes.VMCreateRequest, osType string) error {
	serviceLog.Infof("Using fallback hardware configuration for: %s", vmName)

	// Set OS type
	serviceLog.Infof("Setting OS type: %s", osType)
	if err := s.vboxExec.setOSType(vmName, osType); err != nil {
		return fmt.Errorf("failed to set OS type: %w", err)
	}

	// Set CPU count
	serviceLog.Infof("Setting CPU count to %d", req.CPUCores)
	if err := s.vboxExec.setCPUs(vmName, req.CPUCores); err != nil {
		return fmt.Errorf("failed to set CPU count: %w", err)
	}

	// Set memory
	serviceLog.Infof("Setting memory to %d MB", req.MemoryMB)
	if err := s.vboxExec.setMemory(vmName, req.MemoryMB); err != nil {
		return fmt.Errorf("failed to set memory: %w", err)
	}

	// Set VRAM (video memory) - fallback to 16MB
	serviceLog.Infof("Setting VRAM to 16 MB")
	if err := s.vboxExec.setVRAM(vmName, 16); err != nil {
		return fmt.Errorf("failed to set VRAM: %w", err)
	}

	// Set chipset based on runtime architecture
	chipset := "ich9"
	if runtime.GOARCH == "arm64" || runtime.GOARCH == "aarch64" {
		chipset = "armv8virtual"
	}
	serviceLog.Infof("Setting chipset to %s", chipset)
	if err := s.vboxExec.setChipset(vmName, chipset); err != nil {
		return fmt.Errorf("failed to set chipset: %w", err)
	}

	// Set firmware - fallback to EFI for most architectures
	firmware := "efi"
	if runtime.GOARCH == "386" || runtime.GOARCH == "i386" {
		firmware = "bios"
	}
	serviceLog.Infof("Setting firmware to %s", firmware)
	if err := s.vboxExec.setFirmware(vmName, firmware); err != nil {
		return fmt.Errorf("failed to set firmware: %w", err)
	}

	// Set graphics controller - fallback to vmsvga
	serviceLog.Infof("Setting graphics controller to vmsvga")
	if err := s.vboxExec.setGraphicsController(vmName, "vmsvga"); err != nil {
		return fmt.Errorf("failed to set graphics controller: %w", err)
	}

	// Set IOAPIC based on architecture
	ioapicEnabled := true
	if runtime.GOARCH == "arm64" || runtime.GOARCH == "aarch64" || runtime.GOARCH == "arm" {
		ioapicEnabled = false
	}
	ioapicStatus := ioapicEnabled
	serviceLog.Infof("Setting IOAPIC to %v", ioapicStatus)
	if err := s.vboxExec.setIOAPIC(vmName, ioapicStatus); err != nil {
		return fmt.Errorf("failed to set IOAPIC: %w", err)
	}

	// Set boot order
	serviceLog.Infof("Setting boot order")
	if err := s.vboxExec.setBootOrder(vmName); err != nil {
		return fmt.Errorf("failed to set boot order: %w", err)
	}

	// Configure input devices
	serviceLog.Infof("Configuring input devices...")
	if err := s.vboxExec.configureInputDevices(vmName); err != nil {
		return fmt.Errorf("failed to configure input devices: %w", err)
	}

	// Configure USB
	serviceLog.Infof("Configuring USB...")
	if err := s.vboxExec.configureUSB(vmName); err != nil {
		return fmt.Errorf("failed to configure USB: %w", err)
	}

	// Configure audio
	serviceLog.Infof("Configuring audio...")
	if err := s.vboxExec.configureAudio(vmName); err != nil {
		return fmt.Errorf("failed to configure audio: %w", err)
	}

	serviceLog.Infof("Fallback VM hardware configuration completed for: %s", vmName)
	return nil
}

// setupVMStorageWithVBoxManage sets up VM storage including disk and ISO attachment
func (s *ServiceImpl) setupVMStorageWithVBoxManage(vmName string, req vbtypes.VMCreateRequest, isoPath string) error {
	serviceLog.Infof("Starting VM storage setup with VBoxManage...")

	// Create virtual disk using VBoxManage
	serviceLog.Infof("Creating virtual disk...")
	diskPath := filepath.Join(s.vmDir, req.Name, fmt.Sprintf("%s.vdi", req.Name))
	serviceLog.Infof("Disk path: %s", diskPath)

	if _, err := s.vboxExec.createVirtualDisk(vmName, diskPath, req.DiskSizeGB); err != nil {
		return fmt.Errorf("failed to create virtual disk: %w", err)
	}
	serviceLog.Infof("Virtual disk created successfully")

	// Add VirtioSCSI controller using VBoxManage (better for ARM64)
	serviceLog.Infof("Adding VirtioSCSI controller...")
	if err := s.vboxExec.addVirtioSCSIController(vmName); err != nil {
		return fmt.Errorf("failed to add VirtioSCSI controller: %w", err)
	}
	serviceLog.Infof("VirtioSCSI controller added successfully")

	// Attach disk to VirtioSCSI controller
	serviceLog.Infof("Attaching disk to VirtioSCSI controller...")
	if err := s.vboxExec.attachDisk(vmName, diskPath); err != nil {
		return fmt.Errorf("failed to attach disk: %w", err)
	}
	serviceLog.Infof("Disk attached successfully")

	// Attach ISO to VirtioSCSI controller if available
	if isoPath != "" {
		serviceLog.Infof("Attaching ISO to VirtioSCSI controller...")
		if err := s.vboxExec.attachISO(vmName, isoPath); err != nil {
			return fmt.Errorf("failed to attach ISO: %w", err)
		}
		serviceLog.Infof("ISO attached successfully")
	}

	serviceLog.Infof("VM storage setup completed successfully")
	return nil
}

// generateCloudInitISO generates cloud-init files and ISO for a VM
func (s *ServiceImpl) generateCloudInitISO(vmName string, username string, password string) (string, error) {
	serviceLog.Infof("Generating cloud-init ISO for VM: %s", vmName)

	// Generate VM-specific cloud-init configuration
	hostname := vmName

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create VM-specific cloud-init configuration
	_, _, cloudInitDir, err := s.vboxExec.GenerateCloudInitFiles(vmName, hostname, username, hashedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to generate cloud-init files: %w", err)
	}

	cloudInitISO, err := s.vboxExec.GenerateCloudInitISO(cloudInitDir)
	if err != nil {
		return "", fmt.Errorf("failed to generate cloud-init ISO: %w", err)
	}

	serviceLog.Infof("Successfully generated cloud-init ISO for VM %s: %s", vmName, cloudInitISO)
	serviceLog.Infof("VM %s cloud-init credentials - Username: %s, Password: %s", vmName, username, password)
	return cloudInitISO, nil
}

func hashPassword(password string) (string, error) {
	// openssl passwd -1  <password>
	cmd := exec.Command("openssl", "passwd", "-1", password)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(output), nil
}

func generateVMID(name string) string {
	// Simple ID generation - in production, you might want a more sophisticated approach
	return fmt.Sprintf("vm-%s-%d", strings.ToLower(name), time.Now().Unix())
}

// validateVirtualBoxInstallation checks if VirtualBox is properly installed
func (s *ServiceImpl) validateVirtualBoxInstallation() error {
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
func (s *ServiceImpl) parseMachineReadableOutput(output string) map[string]string {
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

func (s *ServiceImpl) parseVMListOutput(output string) []string {
	var vmNames []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse "VMName" {uuid} format
		if strings.Contains(line, `"`) {
			start := strings.Index(line, `"`)
			end := strings.LastIndex(line, `"`)
			if start != -1 && end != -1 && start != end {
				vmName := line[start+1 : end]
				vmNames = append(vmNames, vmName)
			}
		}
	}

	return vmNames
}

func (s *ServiceImpl) parseVMStatus(status string) vbtypes.VMStatus {
	switch strings.ToLower(status) {
	case "running":
		return vbtypes.Running
	case "poweroff":
		return vbtypes.Stopped
	case "paused":
		return vbtypes.Paused
	case "saved":
		return vbtypes.Stopped
	case "aborted":
		return vbtypes.Stopped
	default:
		return vbtypes.Unknown
	}
}

func (s *ServiceImpl) parseIntOrDefault(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	if parsed, err := fmt.Sscanf(value, "%d", &defaultValue); err != nil || parsed != 1 {
		return defaultValue
	}

	return defaultValue
}
