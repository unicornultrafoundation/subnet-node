package virtualbox

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

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
	vboxClient VBoxClient
	storageMgr StorageManager
	vmDir      string
	stopChan   chan struct{}
	config     *ServiceConfig
}

// NewService creates a new VirtualBox service
func NewService(config *ServiceConfig) (*ServiceImpl, error) {
	// Create VirtualBox client
	vboxClient, err := NewVBoxClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create VirtualBox client: %w", err)
	}

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
		vboxClient: vboxClient,
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

	// Verify VirtualBox is available
	version, err := s.vboxClient.GetVBoxVersion()
	if err != nil {
		return fmt.Errorf("VirtualBox not available: %w", err)
	}

	serviceLog.Infof("VirtualBox version: %s", version)
	return nil
}

// Stop stops the VirtualBox service
func (s *ServiceImpl) Stop(ctx context.Context) error {
	serviceLog.Info("Stopping VirtualBox service")
	close(s.stopChan)
	return nil
}

// CreateVM creates a new VM with the specified configuration
func (s *ServiceImpl) CreateVM(ctx context.Context, req vbtypes.VMCreateRequest) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	serviceLog.Infof("Creating VM: %s", req.Name)

	// Generate unique VM ID
	vmID := generateVMID(req.Name)

	// Set default ISO URL if not provided
	if req.ISOURL == "" {
		req.ISOURL = "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-arm64.iso"
	}

	// Download ISO if needed
	isoPath, err := s.ensureISO(ctx, req.ISOURL)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure ISO: %w", err)
	}

	// Create VM directory
	vmFolder := filepath.Join(s.vmDir, req.Name)
	if err := os.MkdirAll(vmFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create VM directory: %w", err)
	}

	// Create VM
	if err := s.vboxClient.CreateVM(req.Name, "Ubuntu_ARM64"); err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	// Configure VM hardware
	if err := s.configureVMHardware(req.Name, req); err != nil {
		// Clean up on failure
		s.vboxClient.DeleteVM(req.Name)
		return nil, fmt.Errorf("failed to configure VM hardware: %w", err)
	}

	// Create and attach storage
	if err := s.setupVMStorage(req.Name, req, isoPath); err != nil {
		// Clean up on failure
		s.vboxClient.DeleteVM(req.Name)
		return nil, fmt.Errorf("failed to setup VM storage: %w", err)
	}

	// Create VM object
	vm := &vbtypes.VM{
		ID:         vmID,
		Name:       req.Name,
		Status:     vbtypes.Stopped,
		CPUCores:   req.CPUCores,
		MemoryMB:   req.MemoryMB,
		DiskSizeGB: req.DiskSizeGB,
		ISOURL:     req.ISOURL,
		ISOPath:    isoPath,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		VBoxPath:   s.vboxClient.(*VBoxClientImpl).vboxManagePath,
		VMFolder:   vmFolder,
	}

	serviceLog.Infof("Successfully created VM: %s", req.Name)
	return vm, nil
}

// GetVM gets a VM by ID
func (s *ServiceImpl) GetVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// For now, we'll use the VM name as ID
	// In a real implementation, you'd have a mapping from ID to name
	vmName := vmID

	// Get VM info from VirtualBox
	info, err := s.vboxClient.GetVMInfo(vmName)
	if err != nil {
		return nil, fmt.Errorf("VM not found: %w", err)
	}

	// Get VM status
	status, err := s.vboxClient.GetVMStatus(vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM status: %w", err)
	}

	// Parse VM info to create VM object
	vm := &vbtypes.VM{
		ID:       vmID,
		Name:     vmName,
		Status:   vbtypes.VMStatus(status),
		VBoxPath: s.vboxClient.(*VBoxClientImpl).vboxManagePath,
		VMFolder: filepath.Join(s.vmDir, vmName),
	}

	// Parse additional info if available
	if memory, ok := info["Memory size"]; ok {
		if memMB, err := strconv.Atoi(strings.ReplaceAll(memory, "MB", "")); err == nil {
			vm.MemoryMB = memMB
		}
	}

	return vm, nil
}

// GetVMs gets a list of VMs with pagination and filtering
func (s *ServiceImpl) GetVMs(ctx context.Context, start, end *big.Int, filter vbtypes.VMFilter) ([]*vbtypes.VM, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all VM names
	vmNames, err := s.vboxClient.ListVMs()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list VMs: %w", err)
	}

	var vms []*vbtypes.VM
	for _, name := range vmNames {
		vm, err := s.GetVM(ctx, name)
		if err != nil {
			serviceLog.Warnf("Failed to get VM %s: %v", name, err)
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

	vmNames, err := s.vboxClient.ListVMs()
	if err != nil {
		return big.NewInt(0), fmt.Errorf("failed to list VMs: %w", err)
	}

	if filter == nil {
		return big.NewInt(int64(len(vmNames))), nil
	}

	count := 0
	for _, name := range vmNames {
		vm, err := s.GetVM(ctx, name)
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

// UpdateVM updates an existing VM
func (s *ServiceImpl) UpdateVM(ctx context.Context, vmID string, req vbtypes.VMUpdateRequest) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	// Get current VM
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	// Update VM parameters
	params := make(map[string]string)
	if req.Name != "" && req.Name != vm.Name {
		params["name"] = req.Name
		vm.Name = req.Name
	}
	if req.CPUCores > 0 && req.CPUCores != vm.CPUCores {
		params["cpus"] = strconv.Itoa(req.CPUCores)
		vm.CPUCores = req.CPUCores
	}
	if req.MemoryMB > 0 && req.MemoryMB != vm.MemoryMB {
		params["memory"] = strconv.Itoa(req.MemoryMB)
		vm.MemoryMB = req.MemoryMB
	}

	// Apply changes if any
	if len(params) > 0 {
		if err := s.vboxClient.ModifyVM(vmName, params); err != nil {
			return nil, fmt.Errorf("failed to modify VM: %w", err)
		}
		vm.UpdatedAt = time.Now()
	}

	return vm, nil
}

// DeleteVM deletes a VM
func (s *ServiceImpl) DeleteVM(ctx context.Context, vmID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Deleting VM: %s", vmName)

	// Stop VM if running
	status, err := s.vboxClient.GetVMStatus(vmName)
	if err == nil && status == "running" {
		if err := s.vboxClient.StopVM(vmName); err != nil {
			serviceLog.Warnf("Failed to stop VM before deletion: %v", err)
		}
	}

	// Delete VM
	if err := s.vboxClient.DeleteVM(vmName); err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	serviceLog.Infof("Successfully deleted VM: %s", vmName)
	return nil
}

// StartVM starts a VM
func (s *ServiceImpl) StartVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Starting VM: %s", vmName)

	if err := s.vboxClient.StartVM(vmName, false); err != nil {
		return nil, fmt.Errorf("failed to start VM: %w", err)
	}

	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	serviceLog.Infof("Successfully started VM: %s", vmName)
	return vm, nil
}

// StopVM stops a VM
func (s *ServiceImpl) StopVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Stopping VM: %s", vmName)

	if err := s.vboxClient.StopVM(vmName); err != nil {
		return nil, fmt.Errorf("failed to stop VM: %w", err)
	}

	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	serviceLog.Infof("Successfully stopped VM: %s", vmName)
	return vm, nil
}

// PauseVM pauses a VM
func (s *ServiceImpl) PauseVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Pausing VM: %s", vmName)

	if err := s.vboxClient.PauseVM(vmName); err != nil {
		return nil, fmt.Errorf("failed to pause VM: %w", err)
	}

	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	serviceLog.Infof("Successfully paused VM: %s", vmName)
	return vm, nil
}

// ResumeVM resumes a VM
func (s *ServiceImpl) ResumeVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Resuming VM: %s", vmName)

	if err := s.vboxClient.ResumeVM(vmName); err != nil {
		return nil, fmt.Errorf("failed to resume VM: %w", err)
	}

	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	serviceLog.Infof("Successfully resumed VM: %s", vmName)
	return vm, nil
}

// ResetVM resets a VM
func (s *ServiceImpl) ResetVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vmName := vmID // Using name as ID for now

	serviceLog.Infof("Resetting VM: %s", vmName)

	if err := s.vboxClient.ResetVM(vmName); err != nil {
		return nil, fmt.Errorf("failed to reset VM: %w", err)
	}

	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}

	serviceLog.Infof("Successfully reset VM: %s", vmName)
	return vm, nil
}

// GetVMUsage gets resource usage for a specific VM
func (s *ServiceImpl) GetVMUsage(ctx context.Context, vmID string) (*vbtypes.VMUsage, error) {
	// This would require additional implementation to get actual VM metrics
	// For now, return a placeholder
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
	// This would aggregate usage from all VMs
	// For now, return a placeholder
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
	version, err := s.vboxClient.GetVBoxVersion()
	if err != nil {
		return nil, err
	}

	hostInfo, err := s.vboxClient.GetHostInfo()
	if err != nil {
		return nil, err
	}

	return &vbtypes.VMSystemInfo{
		VBoxVersion:     version,
		HostOS:          hostInfo["os"],
		HostArch:        hostInfo["arch"],
		AvailableCPUs:   0, // Would need to parse from hostInfo
		AvailableRAMMB:  0, // Would need to parse from hostInfo
		AvailableDiskGB: 0, // Would need to calculate
	}, nil
}

// DownloadISO downloads an ISO file
func (s *ServiceImpl) DownloadISO(ctx context.Context, isoURL string) (*vbtypes.ISOInfo, error) {
	isoPath := s.storageMgr.(*StorageManagerImpl).GetISOPath(isoURL)

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
	isoDir := s.storageMgr.(*StorageManagerImpl).GetISODir()
	files, err := s.storageMgr.ListFiles(isoDir)
	if err != nil {
		return nil, err
	}

	var isos []*vbtypes.ISOInfo
	for _, file := range files {
		isoInfo, err := s.storageMgr.GetFileInfo(file)
		if err != nil {
			serviceLog.Warnf("Failed to get ISO info for %s: %v", file, err)
			continue
		}
		isos = append(isos, isoInfo)
	}

	return isos, nil
}

// DeleteISO deletes an ISO file
func (s *ServiceImpl) DeleteISO(ctx context.Context, isoPath string) error {
	return s.storageMgr.DeleteFile(isoPath)
}

// Helper methods

func (s *ServiceImpl) ensureISO(ctx context.Context, isoURL string) (string, error) {
	isoPath := s.storageMgr.(*StorageManagerImpl).GetISOPath(isoURL)

	// Check if already exists
	if s.storageMgr.FileExists(isoPath) {
		serviceLog.Infof("ISO already exists: %s", isoPath)
		return isoPath, nil
	}

	// Download the ISO
	serviceLog.Infof("Downloading ISO: %s", isoURL)
	if err := s.storageMgr.DownloadFile(ctx, isoURL, isoPath); err != nil {
		return "", err
	}

	return isoPath, nil
}

func (s *ServiceImpl) configureVMHardware(vmName string, req vbtypes.VMCreateRequest) error {
	params := map[string]string{
		"memory":             strconv.Itoa(req.MemoryMB),
		"vram":               "16",
		"cpus":               strconv.Itoa(req.CPUCores),
		"chipset":            "armv8virtual",
		"firmware":           "efi",
		"graphicscontroller": "vmsvga",
		"ioapic":             "off",
		"boot1":              "dvd",
		"boot2":              "disk",
		"boot3":              "none",
		"boot4":              "none",
		"mouse":              "usbtablet",
		"keyboard":           "usb",
		"usbohci":            "off",
		"usbehci":            "off",
		"usbxhci":            "on",
		"audio-controller":   "hda",
		"audio-out":          "on",
		"audio-in":           "off",
		"nic1":               "nat",
		"cableconnected1":    "on",
	}

	return s.vboxClient.ModifyVM(vmName, params)
}

func (s *ServiceImpl) setupVMStorage(vmName string, req vbtypes.VMCreateRequest, isoPath string) error {
	vmFolder := filepath.Join(s.vmDir, vmName)
	diskPath := filepath.Join(vmFolder, vmName+".vdi")

	// Create hard disk
	if err := s.vboxClient.CreateHD(diskPath, req.DiskSizeGB*1024); err != nil {
		return fmt.Errorf("failed to create hard disk: %w", err)
	}

	// Add storage controller
	if err := s.vboxClient.StorageController(vmName, "VirtioSCSI"); err != nil {
		return fmt.Errorf("failed to add storage controller: %w", err)
	}

	// Attach hard disk
	if err := s.vboxClient.StorageAttach(vmName, "VirtioSCSI", 0, 0, "hdd", diskPath); err != nil {
		return fmt.Errorf("failed to attach hard disk: %w", err)
	}

	// Attach ISO
	if err := s.vboxClient.StorageAttach(vmName, "VirtioSCSI", 1, 0, "dvddrive", isoPath); err != nil {
		return fmt.Errorf("failed to attach ISO: %w", err)
	}

	return nil
}

func generateVMID(name string) string {
	// Simple ID generation - in production, you might want a more sophisticated approach
	return fmt.Sprintf("vm-%s-%d", strings.ToLower(name), time.Now().Unix())
}
