package virtualbox

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

var serviceLog = logrus.WithField("service", "virtualbox")

// ServiceConfig holds the configuration for the VirtualBox service
type ServiceConfig struct {
	Enable bool

	// VM defaults
	DefaultMemoryMB    int
	DefaultCPUs        int
	DefaultDiskSizeGB  int
	DefaultNetworkType string
	DefaultOSType      string

	// Timeouts
	VMStartTimeout  time.Duration
	VMStopTimeout   time.Duration
	VMDeleteTimeout time.Duration

	// Monitoring
	MonitorInterval time.Duration
	Headless        bool

	// Network configuration
	Network NetworkConfig

	// Storage configuration
	Storage StorageConfig

	// Advanced settings
	Advanced AdvancedConfig
}

// NetworkConfig holds network-related configuration
type NetworkConfig struct {
	DefaultBridgeName string
	EnableNAT         bool
	EnableBridged     bool
	EnableHostOnly    bool
	EnableInternal    bool
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	DefaultController string
	DefaultType       string
	EnableTrim        bool
	EnableCompression bool
}

// AdvancedConfig holds advanced VirtualBox settings
type AdvancedConfig struct {
	EnableAudio        bool
	EnableUSB          bool
	EnableVRDE         bool
	VRDEPort           int
	EnablePAE          bool
	EnableNestedPaging bool
	EnableHWVirt       bool
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	mu         sync.RWMutex
	storageMgr StorageManager
	vmDir      string
	stopChan   chan struct{}
	config     *ServiceConfig
}

// NewService creates a new VirtualBox service
func NewService(config *ServiceConfig) (*ServiceImpl, error) {
	// Create storage manager
	storageMgr, err := NewStorageManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create storage manager: %w", err)
	}

	// Get VM directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	vmDir := filepath.Join(homeDir, "VirtualBox VMs")

	return &ServiceImpl{
		storageMgr: storageMgr,
		vmDir:      vmDir,
		stopChan:   make(chan struct{}),
		config:     config,
	}, nil
}

// LoadServiceConfig loads VirtualBox service configuration from the main config
func LoadServiceConfig(cfg *config.C) *ServiceConfig {
	return &ServiceConfig{
		Enable: cfg.GetBool("virtualbox.enable", false),

		// VM defaults
		DefaultMemoryMB:    cfg.GetInt("virtualbox.default_memory_mb", 2048),
		DefaultCPUs:        cfg.GetInt("virtualbox.default_cpus", 2),
		DefaultDiskSizeGB:  cfg.GetInt("virtualbox.default_disk_size_gb", 20),
		DefaultNetworkType: cfg.GetString("virtualbox.default_network_type", "nat"),
		DefaultOSType:      cfg.GetString("virtualbox.default_os_type", "Ubuntu_arm64"),

		// Timeouts
		VMStartTimeout:  cfg.GetDuration("virtualbox.vm_start_timeout", 60*time.Second),
		VMStopTimeout:   cfg.GetDuration("virtualbox.vm_stop_timeout", 30*time.Second),
		VMDeleteTimeout: cfg.GetDuration("virtualbox.vm_delete_timeout", 60*time.Second),

		// Monitoring
		MonitorInterval: cfg.GetDuration("virtualbox.monitor_interval", 30*time.Second),
		Headless:        cfg.GetBool("virtualbox.headless", true),

		// Network configuration
		Network: NetworkConfig{
			DefaultBridgeName: cfg.GetString("virtualbox.network.default_bridge_name", "en0"),
			EnableNAT:         cfg.GetBool("virtualbox.network.enable_nat", true),
			EnableBridged:     cfg.GetBool("virtualbox.network.enable_bridged", true),
			EnableHostOnly:    cfg.GetBool("virtualbox.network.enable_hostonly", false),
			EnableInternal:    cfg.GetBool("virtualbox.network.enable_internal", false),
		},

		// Storage configuration
		Storage: StorageConfig{
			DefaultController: cfg.GetString("virtualbox.storage.default_controller", "SATA"),
			DefaultType:       cfg.GetString("virtualbox.storage.default_type", "vdi"),
			EnableTrim:        cfg.GetBool("virtualbox.storage.enable_trim", true),
			EnableCompression: cfg.GetBool("virtualbox.storage.enable_compression", false),
		},

		// Advanced settings
		Advanced: AdvancedConfig{
			EnableAudio:        cfg.GetBool("virtualbox.advanced.enable_audio", false),
			EnableUSB:          cfg.GetBool("virtualbox.advanced.enable_usb", false),
			EnableVRDE:         cfg.GetBool("virtualbox.advanced.enable_vrde", true),
			VRDEPort:           cfg.GetInt("virtualbox.advanced.vrde_port", 3389),
			EnablePAE:          cfg.GetBool("virtualbox.advanced.enable_pae", false),
			EnableNestedPaging: cfg.GetBool("virtualbox.advanced.enable_nested_paging", true),
			EnableHWVirt:       cfg.GetBool("virtualbox.advanced.enable_hw_virt", true),
		},
	}
}

// Start starts the VirtualBox service
func (s *ServiceImpl) Start(ctx context.Context) error {
	serviceLog.Info("Starting VirtualBox service")

	// Verify VirtualBox is available by trying to list machines
	if err := s.validateVirtualBoxInstallation(); err != nil {
		return fmt.Errorf("VirtualBox validation failed: %w", err)
	}

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

	// Validate VirtualBox installation first
	if err := s.validateVirtualBoxInstallation(); err != nil {
		return nil, fmt.Errorf("VirtualBox validation failed: %w", err)
	}

	// Generate unique VM ID
	vmID := generateVMID(req.Name)
	serviceLog.Infof("Generated VM ID: %s", vmID)

	// Determine OS type and ISO URL based on the request
	serviceLog.Infof("Determining OS type and ISO URL...")
	osType, isoURL, err := s.determineOSTypeAndISOFromRequest(req)
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
	if err := os.MkdirAll(vmFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create VM directory: %w", err)
	}

	// Create VM using VBoxManage
	serviceLog.Infof("Creating VM with VBoxManage...")
	if err := s.createVMWithVBoxManage(req.Name, vmFolder); err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}
	serviceLog.Infof("VM created successfully")

	// Configure VM hardware using VBoxManage
	serviceLog.Infof("Configuring VM hardware...")
	if err := s.configureVMHardwareWithVBoxManage(req.Name, req, osType); err != nil {
		serviceLog.Errorf("Failed to configure VM hardware: %v", err)
		// Clean up on failure
		serviceLog.Infof("Cleaning up failed VM...")
		if delErr := s.deleteVMWithVBoxManage(req.Name); delErr != nil {
			serviceLog.Errorf("Failed to delete VM during cleanup: %v", delErr)
		}
		return nil, fmt.Errorf("failed to configure VM hardware: %w", err)
	}
	serviceLog.Infof("VM hardware configured successfully")

	// Create and attach storage using VBoxManage
	serviceLog.Infof("Setting up VM storage...")
	if err := s.setupVMStorageWithVBoxManage(req.Name, req, isoPath); err != nil {
		serviceLog.Errorf("Failed to setup VM storage: %v", err)
		// Clean up on failure
		serviceLog.Infof("Cleaning up failed VM...")
		if delErr := s.deleteVMWithVBoxManage(req.Name); delErr != nil {
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

// createVMWithVBoxManage creates a VM using VBoxManage
func (s *ServiceImpl) createVMWithVBoxManage(name, baseFolder string) error {
	args := []string{"createvm", "--name", name, "--register"}
	if baseFolder != "" {
		args = append(args, "--basefolder", baseFolder)
	}

	cmd := exec.Command("VBoxManage", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("VBoxManage createvm failed: %w, output: %s", err, string(output))
	}

	serviceLog.Infof("VBoxManage createvm output: %s", string(output))
	return nil
}

// GetVM gets a VM by ID using VBoxManage
func (s *ServiceImpl) GetVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// For now, we'll use the VM name as ID
	vmName := vmID

	// Get VM info from VirtualBox using VBoxManage
	cmd := exec.Command("VBoxManage", "showvminfo", vmName, "--machinereadable")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("VM not found: %w", err)
	}

	// Parse the machine-readable output
	vmInfo := s.parseMachineReadableOutput(string(output))

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

// GetVMs gets a list of VMs with pagination and filtering using VBoxManage
func (s *ServiceImpl) GetVMs(ctx context.Context, start, end *big.Int, filter vbtypes.VMFilter) ([]*vbtypes.VM, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all VMs using VBoxManage
	cmd := exec.Command("VBoxManage", "list", "vms")
	output, err := cmd.Output()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list VMs: %w", err)
	}

	// Parse the output to get VM names
	vmNames := s.parseVMListOutput(string(output))

	var vms []*vbtypes.VM
	for _, vmName := range vmNames {
		// Get detailed info for each VM
		vm, err := s.GetVM(ctx, vmName)
		if err != nil {
			serviceLog.Warnf("Failed to get VM info for %s: %v", vmName, err)
			continue
		}

		// Apply filters
		if filter.Status != "" && string(vm.Status) != filter.Status {
			continue
		}
		if filter.Query != "" && !strings.Contains(strings.ToLower(vm.Name), strings.ToLower(filter.Query)) {
			continue
		}

		vms = append(vms, vm)
	}

	// Apply pagination
	total := len(vms)
	startIdx := int(start.Int64())
	endIdx := int(end.Int64())

	if startIdx >= total {
		return []*vbtypes.VM{}, total, nil
	}

	if endIdx > total {
		endIdx = total
	}

	return vms[startIdx:endIdx], total, nil
}

// GetVMCount gets the count of VMs with optional filtering
func (s *ServiceImpl) GetVMCount(ctx context.Context, filter *vbtypes.VMFilter) (*big.Int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cmd := exec.Command("VBoxManage", "list", "vms")
	output, err := cmd.Output()
	if err != nil {
		return big.NewInt(0), fmt.Errorf("failed to list VMs: %w", err)
	}

	vmNames := s.parseVMListOutput(string(output))

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
		cmd := exec.Command("VBoxManage", "modifyvm", vmName, "--cpus", fmt.Sprintf("%d", req.CPUCores))
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("failed to update CPU cores: %w, output: %s", err, string(output))
		}
		vm.CPUCores = req.CPUCores
	}

	if req.MemoryMB > 0 && req.MemoryMB != vm.MemoryMB {
		cmd := exec.Command("VBoxManage", "modifyvm", vmName, "--memory", fmt.Sprintf("%d", req.MemoryMB))
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("failed to update memory: %w, output: %s", err, string(output))
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

	// Delete VM using VBoxManage
	if err := s.deleteVMWithVBoxManage(vmName); err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	serviceLog.Infof("Successfully deleted VM: %s", vmName)
	return nil
}

// deleteVMWithVBoxManage deletes a VM using VBoxManage
func (s *ServiceImpl) deleteVMWithVBoxManage(vmName string) error {
	cmd := exec.Command("VBoxManage", "unregistervm", vmName, "--delete")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("VBoxManage unregistervm failed: %w, output: %s", err, string(output))
	}
	return nil
}

// StartVM starts a VM using VBoxManage
func (s *ServiceImpl) StartVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Starting VM: %s", vmName)

	// Start the VM using VBoxManage
	cmd := exec.Command("VBoxManage", "startvm", vmName)
	if s.config.Headless {
		cmd = exec.Command("VBoxManage", "startvm", vmName, "--type", "headless")
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to start VM: %w, output: %s", err, string(output))
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

	// Stop the VM using VBoxManage
	cmd := exec.Command("VBoxManage", "controlvm", vmName, "poweroff")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to stop VM: %w, output: %s", err, string(output))
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

	// Pause the VM using VBoxManage
	cmd := exec.Command("VBoxManage", "controlvm", vmName, "pause")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to pause VM: %w, output: %s", err, string(output))
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

	// Resume the VM using VBoxManage
	cmd := exec.Command("VBoxManage", "controlvm", vmName, "resume")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to resume VM: %w, output: %s", err, string(output))
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

	// Reset the VM using VBoxManage
	cmd := exec.Command("VBoxManage", "controlvm", vmName, "reset")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to reset VM: %w, output: %s", err, string(output))
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
	// Get VirtualBox version using VBoxManage
	cmd := exec.Command("VBoxManage", "--version")
	output, err := cmd.Output()
	vboxVersion := "VBoxManage"
	if err == nil {
		vboxVersion = strings.TrimSpace(string(output))
	}

	info := &vbtypes.VMSystemInfo{
		VBoxVersion:     vboxVersion,
		HostOS:          runtime.GOOS,
		HostArch:        runtime.GOARCH,
		AvailableCPUs:   runtime.NumCPU(),
		AvailableRAMMB:  0, // TODO: Implement RAM detection
		AvailableDiskGB: 0, // TODO: Implement disk space detection
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
	// Detect the current machine architecture
	arch := runtime.GOARCH
	os := runtime.GOOS

	// Define OS types based on architecture
	var osTypes []string

	switch arch {
	case "arm64", "aarch64":
		// ARM64 architecture - Apple Silicon, ARM servers, etc.
		osTypes = []string{
			"Ubuntu_ARM64",
			"Debian_ARM64",
			"Fedora_ARM64",
			"OpenSUSE_ARM64",
			"ArchLinux_ARM64",
			"RedHat_ARM64",
			"Oracle_ARM64",
			"Linux_ARM64",
			"FreeBSD_ARM64",
			"NetBSD_ARM64",
			"BSD_ARM64",
			"Other_ARM64",
		}
	case "amd64", "x86_64":
		// x86_64 architecture - Intel/AMD 64-bit
		osTypes = []string{
			"Ubuntu_64",
			"Debian_64",
			"Fedora_64",
			"OpenSUSE_64",
			"ArchLinux_64",
			"RedHat_64",
			"Oracle_64",
			"Linux_64",
			"Windows10_64",
			"Windows11_64",
			"Windows2019_64",
			"Windows2022_64",
			"FreeBSD_64",
			"NetBSD_64",
			"BSD_64",
			"Other_64",
		}
	case "arm":
		// 32-bit ARM architecture
		osTypes = []string{
			"Ubuntu",
			"Debian",
			"Fedora",
			"OpenSUSE",
			"ArchLinux",
			"RedHat",
			"Oracle",
			"Linux",
			"Other",
		}
	case "386", "i386":
		// 32-bit x86 architecture
		osTypes = []string{
			"Ubuntu",
			"Debian",
			"Fedora",
			"OpenSUSE",
			"ArchLinux",
			"RedHat",
			"Oracle",
			"Linux",
			"Windows10",
			"Windows7",
			"WindowsXP",
			"Other",
		}
	default:
		// Unknown architecture - return basic types
		osTypes = []string{
			"Other",
			"Linux",
		}
	}

	// Add OS-specific types based on the host OS
	switch os {
	case "darwin":
		// macOS host - add macOS guest types
		if arch == "arm64" {
			osTypes = append(osTypes, "MacOS_ARM64", "Darwin_ARM64")
		} else if arch == "amd64" {
			osTypes = append(osTypes, "MacOS_64", "Darwin_64")
		}
	case "linux":
		// Linux host - already covered above
	case "windows":
		// Windows host - already covered above
	}

	// Sort and remove duplicates
	seen := make(map[string]bool)
	var uniqueOSTypes []string
	for _, osType := range osTypes {
		if !seen[osType] {
			seen[osType] = true
			uniqueOSTypes = append(uniqueOSTypes, osType)
		}
	}

	return uniqueOSTypes, nil
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
		return isoPath, nil
	}

	// Download ISO
	if err := s.storageMgr.DownloadFile(ctx, isoURL, isoPath); err != nil {
		return "", err
	}

	return isoPath, nil
}

// configureVMHardwareWithVBoxManage configures VM hardware using VBoxManage
func (s *ServiceImpl) configureVMHardwareWithVBoxManage(vmName string, req vbtypes.VMCreateRequest, osType string) error {
	serviceLog.Infof("Starting VM hardware configuration with VBoxManage...")

	// Set OS type
	serviceLog.Infof("Setting OS type: %s", osType)
	cmd := exec.Command("VBoxManage", "modifyvm", vmName, "--ostype", osType)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set OS type: %w, output: %s", err, string(output))
	}

	// Set CPU count
	serviceLog.Infof("Setting CPU count to %d", req.CPUCores)
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--cpus", fmt.Sprintf("%d", req.CPUCores))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set CPU count: %w, output: %s", err, string(output))
	}

	// Set memory
	serviceLog.Infof("Setting memory to %d MB", req.MemoryMB)
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--memory", fmt.Sprintf("%d", req.MemoryMB))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set memory: %w, output: %s", err, string(output))
	}

	// Set VRAM (video memory) - use 16MB for ARM64 as per working script
	serviceLog.Infof("Setting VRAM to 16 MB")
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--vram", "16")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set VRAM: %w, output: %s", err, string(output))
	}

	// Set chipset to armv8virtual for ARM64
	serviceLog.Infof("Setting chipset to armv8virtual")
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--chipset", "armv8virtual")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set chipset: %w, output: %s", err, string(output))
	}

	// Set firmware to EFI for ARM64
	serviceLog.Infof("Setting firmware to EFI")
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--firmware", "efi")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set firmware: %w, output: %s", err, string(output))
	}

	// Set graphics controller to VMSVGA
	serviceLog.Infof("Setting graphics controller to VMSVGA")
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--graphicscontroller", "vmsvga")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set graphics controller: %w, output: %s", err, string(output))
	}

	// Disable IOAPIC for ARM64
	serviceLog.Infof("Disabling IOAPIC")
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--ioapic", "off")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to disable IOAPIC: %w, output: %s", err, string(output))
	}

	// Set boot order
	serviceLog.Infof("Setting boot order")
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--boot1", "dvd", "--boot2", "disk", "--boot3", "none", "--boot4", "none")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set boot order: %w, output: %s", err, string(output))
	}

	// Configure input devices
	serviceLog.Infof("Configuring input devices...")
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--mouse", "usbtablet", "--keyboard", "usb")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure input devices: %w, output: %s", err, string(output))
	}

	// Configure USB
	serviceLog.Infof("Configuring USB...")
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--usbohci", "off", "--usbehci", "off", "--usbxhci", "on")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure USB: %w, output: %s", err, string(output))
	}

	// Configure audio
	serviceLog.Infof("Configuring audio...")
	cmd = exec.Command("VBoxManage", "modifyvm", vmName, "--audio-controller", "hda", "--audio-out", "on", "--audio-in", "off")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure audio: %w, output: %s", err, string(output))
	}

	// Configure network adapter
	serviceLog.Infof("Configuring network adapter...")
	if err := s.configureNetworkAdapterWithVBoxManage(vmName); err != nil {
		return fmt.Errorf("failed to configure network adapter: %w", err)
	}

	serviceLog.Infof("VM hardware configuration completed successfully")
	return nil
}

// configureNetworkAdapterWithVBoxManage configures the network adapter using VBoxManage
func (s *ServiceImpl) configureNetworkAdapterWithVBoxManage(vmName string) error {
	// Set network adapter to NAT
	serviceLog.Infof("Setting network adapter to NAT...")
	cmd := exec.Command("VBoxManage", "modifyvm", vmName, "--nic1", "nat", "--cableconnected1", "on")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set network adapter: %w, output: %s", err, string(output))
	}

	serviceLog.Infof("Network adapter configuration completed")
	return nil
}

// setupVMStorageWithVBoxManage sets up VM storage including disk and ISO attachment
func (s *ServiceImpl) setupVMStorageWithVBoxManage(vmName string, req vbtypes.VMCreateRequest, isoPath string) error {
	serviceLog.Infof("Starting VM storage setup with VBoxManage...")

	// Create virtual disk using VBoxManage
	serviceLog.Infof("Creating virtual disk...")
	diskPath := filepath.Join(s.vmDir, req.Name, fmt.Sprintf("%s.vdi", req.Name))
	serviceLog.Infof("Disk path: %s", diskPath)

	cmd := exec.Command("VBoxManage", "createhd", "--filename", diskPath, "--size", fmt.Sprintf("%d", req.DiskSizeGB*1024), "--format", "VDI")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create virtual disk: %w, output: %s", err, string(output))
	}
	serviceLog.Infof("Virtual disk created successfully")

	// Add VirtioSCSI controller using VBoxManage (better for ARM64)
	serviceLog.Infof("Adding VirtioSCSI controller...")
	cmd = exec.Command("VBoxManage", "storagectl", vmName, "--name", "VirtioSCSI", "--add", "virtio-scsi", "--bootable", "on")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add VirtioSCSI controller: %w, output: %s", err, string(output))
	}
	serviceLog.Infof("VirtioSCSI controller added successfully")

	// Attach disk to VirtioSCSI controller
	serviceLog.Infof("Attaching disk to VirtioSCSI controller...")
	cmd = exec.Command("VBoxManage", "storageattach", vmName, "--storagectl", "VirtioSCSI", "--port", "0", "--device", "0", "--type", "hdd", "--medium", diskPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to attach disk: %w, output: %s", err, string(output))
	}
	serviceLog.Infof("Disk attached successfully")

	// Attach ISO to VirtioSCSI controller if available
	if isoPath != "" {
		serviceLog.Infof("Attaching ISO to VirtioSCSI controller...")
		cmd = exec.Command("VBoxManage", "storageattach", vmName, "--storagectl", "VirtioSCSI", "--port", "1", "--device", "0", "--type", "dvddrive", "--medium", isoPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to attach ISO: %w, output: %s", err, string(output))
		}
		serviceLog.Infof("ISO attached successfully")
	}

	serviceLog.Infof("VM storage setup completed successfully")
	return nil
}

func generateVMID(name string) string {
	// Simple ID generation - in production, you might want a more sophisticated approach
	return fmt.Sprintf("vm-%s-%d", strings.ToLower(name), time.Now().Unix())
}

// determineOSTypeAndISO determines the appropriate OS type and ISO URL based on system architecture
func (s *ServiceImpl) determineOSTypeAndISO(ctx context.Context) (string, string, error) {
	// Detect the current machine architecture
	arch := runtime.GOARCH

	// Map architecture to appropriate OS type and ISO URL
	switch arch {
	case "arm64", "aarch64":
		return "Ubuntu_ARM64", "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-arm64.iso", nil
	case "amd64", "x86_64":
		return "Ubuntu_64", "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-amd64.iso", nil
	case "arm":
		return "Ubuntu", "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-armhf.iso", nil
	case "386", "i386":
		return "Ubuntu", "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-i386.iso", nil
	default:
		// Fallback to Ubuntu_64 for unknown architectures
		return "Ubuntu_64", "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-amd64.iso", nil
	}
}

// determineOSTypeAndISOFromRequest determines the appropriate OS type and ISO URL based on the request
func (s *ServiceImpl) determineOSTypeAndISOFromRequest(req vbtypes.VMCreateRequest) (string, string, error) {
	// Use provided OS type or determine from architecture
	osType := req.OSType
	if osType == "" {
		// Fallback to architecture-based detection
		arch := runtime.GOARCH
		switch arch {
		case "arm64", "aarch64":
			osType = "Ubuntu_ARM64"
		case "amd64", "x86_64":
			osType = "Ubuntu_64"
		case "arm":
			osType = "Ubuntu"
		case "386", "i386":
			osType = "Ubuntu"
		default:
			osType = "Ubuntu_64"
		}
	}

	// Use provided ISO URL or determine based on OS type
	isoURL := req.ISOURL
	if isoURL == "" {
		isoURL = s.getISOURLForOSType(osType)
	}

	return osType, isoURL, nil
}

// getISOURLForOSType returns the ISO URL for a given OS type
func (s *ServiceImpl) getISOURLForOSType(osType string) string {
	switch osType {
	case "Ubuntu_64":
		// Use a working Ubuntu 24.04 LTS server ISO URL
		return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
	case "Ubuntu_ARM64":
		// Use Ubuntu 24.04 LTS server ARM64 ISO
		return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-arm64.iso"
	case "Debian_64":
		return "https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-12.5.0-amd64-netinst.iso"
	case "Debian_ARM64":
		return "https://cdimage.debian.org/debian-cd/current/arm64/iso-cd/debian-12.5.0-arm64-netinst.iso"
	case "Fedora_64":
		return "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Server/x86_64/iso/Fedora-Server-dvd-x86_64-40-1.14.iso"
	case "Fedora_ARM64":
		return "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Server/aarch64/iso/Fedora-Server-dvd-aarch64-40-1.14.iso"
	case "CentOS_64":
		return "https://mirror.stream.centos.org/9-stream/BaseOS/x86_64/iso/CentOS-Stream-9-latest-x86_64-boot.iso"
	case "CentOS_ARM64":
		return "https://mirror.stream.centos.org/9-stream/BaseOS/aarch64/iso/CentOS-Stream-9-latest-aarch64-boot.iso"
	default:
		// Default to Ubuntu 64-bit
		return "https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso"
	}
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
