package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

// VirtualBox default configuration constants
const (
	DefaultVMMemoryMB           = 2048
	DefaultVMCPUs               = 2
	DefaultVMDiskSizeGB         = 20
	DefaultVMOSType             = "Linux_64"
	DefaultVMNetworkType        = "bridged"
	DefaultVMStorageController  = "SATA"
	DefaultVMStorageType        = "vdi"
	DefaultVMEnableAudio        = false
	DefaultVMEnableUSB          = false
	DefaultVMEnableVRDE         = false
	DefaultVMVRDEPort           = 3389
	DefaultVMEnablePAE          = false
	DefaultVMEnableNestedPaging = true
)

// ServiceConfig represents VirtualBox service configuration
type ServiceConfig struct {
	// VirtualBox configuration
	VBoxManagePath string
	VBoxHeadless   bool

	// VM configuration
	DefaultVMPath     string
	DefaultMemoryMB   int
	DefaultCPUs       int
	DefaultDiskSizeGB int

	// Network configuration
	DefaultNetworkType string
	DefaultBridgeName  string

	// Monitor interval
	MonitorInterval time.Duration

	// VM lifecycle configuration
	VMStartTimeout    time.Duration
	VMStopTimeout     time.Duration
	VMDeleteTimeout   time.Duration
	VMShutdownTimeout time.Duration

	// Terraform configuration
	TerraformPath     string
	TerraformWorkDir  string
	TerraformTimeout  time.Duration
	TerraformParallel int

	// Storage configuration
	BaseImagePath string
	SnapshotPath  string
}

// NewServiceConfigFromConfig creates a service config from the main config
func NewServiceConfigFromConfig(cfg *config.C) (*ServiceConfig, error) {
	serviceConfig := DefaultServiceConfig()

	// Get user's home directory for default paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "" // Fallback to empty string if home directory cannot be determined
	}

	// VirtualBox configuration
	serviceConfig.VBoxManagePath = cfg.GetString("virtualbox.vboxmanage_path", "VBoxManage")
	serviceConfig.VBoxHeadless = cfg.GetBool("virtualbox.headless", true)

	// VM configuration
	defaultVMPath := ""
	if homeDir != "" {
		defaultVMPath = filepath.Join(homeDir, "VirtualBox VMs")
	}
	serviceConfig.DefaultVMPath = cfg.GetString("virtualbox.default_vm_path", defaultVMPath)
	serviceConfig.DefaultMemoryMB = cfg.GetInt("virtualbox.default_memory_mb", DefaultVMMemoryMB)
	serviceConfig.DefaultCPUs = cfg.GetInt("virtualbox.default_cpus", DefaultVMCPUs)
	serviceConfig.DefaultDiskSizeGB = cfg.GetInt("virtualbox.default_disk_size_gb", DefaultVMDiskSizeGB)

	// Network configuration
	serviceConfig.DefaultNetworkType = cfg.GetString("virtualbox.default_network_type", DefaultVMNetworkType)
	serviceConfig.DefaultBridgeName = cfg.GetString("virtualbox.default_bridge_name", "")

	// Monitor interval
	serviceConfig.MonitorInterval = cfg.GetDuration("virtualbox.monitor_interval", 30*time.Second)

	// VM lifecycle configuration
	serviceConfig.VMStartTimeout = cfg.GetDuration("virtualbox.vm_start_timeout", 60*time.Second)
	serviceConfig.VMStopTimeout = cfg.GetDuration("virtualbox.vm_stop_timeout", 30*time.Second)
	serviceConfig.VMDeleteTimeout = cfg.GetDuration("virtualbox.vm_delete_timeout", 60*time.Second)
	serviceConfig.VMShutdownTimeout = cfg.GetDuration("virtualbox.vm_shutdown_timeout", 30*time.Second)

	// Terraform configuration
	serviceConfig.TerraformPath = cfg.GetString("virtualbox.terraform_path", "terraform")
	defaultTerraformWorkDir := ""
	if homeDir != "" {
		defaultTerraformWorkDir = filepath.Join(homeDir, ".subnet-node", "terraform")
	}
	serviceConfig.TerraformWorkDir = cfg.GetString("virtualbox.terraform_work_dir", defaultTerraformWorkDir)
	serviceConfig.TerraformTimeout = cfg.GetDuration("virtualbox.terraform_timeout", 300*time.Second)
	serviceConfig.TerraformParallel = cfg.GetInt("virtualbox.terraform_parallel", 4)

	// Storage configuration
	defaultBaseImagePath := ""
	if homeDir != "" {
		defaultBaseImagePath = filepath.Join(homeDir, ".subnet-node", "images")
	}
	serviceConfig.BaseImagePath = cfg.GetString("virtualbox.base_image_path", defaultBaseImagePath)
	defaultSnapshotPath := ""
	if homeDir != "" {
		defaultSnapshotPath = filepath.Join(homeDir, ".subnet-node", "snapshots")
	}
	serviceConfig.SnapshotPath = cfg.GetString("virtualbox.snapshot_path", defaultSnapshotPath)

	if err := serviceConfig.Validate(); err != nil {
		return nil, err
	}

	return serviceConfig, nil
}

// DefaultServiceConfig returns a default service configuration
func DefaultServiceConfig() *ServiceConfig {
	// Get user's home directory for default paths
	homeDir, err := os.UserHomeDir()
	defaultVMPath := ""
	defaultTerraformWorkDir := ""
	defaultBaseImagePath := ""
	defaultSnapshotPath := ""
	if err == nil && homeDir != "" {
		defaultVMPath = filepath.Join(homeDir, "VirtualBox VMs")
		defaultTerraformWorkDir = filepath.Join(homeDir, ".subnet-node", "terraform")
		defaultBaseImagePath = filepath.Join(homeDir, ".subnet-node", "images")
		defaultSnapshotPath = filepath.Join(homeDir, ".subnet-node", "snapshots")
	}

	return &ServiceConfig{
		VBoxManagePath:     "VBoxManage",
		VBoxHeadless:       true,
		DefaultVMPath:      defaultVMPath,
		DefaultMemoryMB:    DefaultVMMemoryMB,
		DefaultCPUs:        DefaultVMCPUs,
		DefaultDiskSizeGB:  DefaultVMDiskSizeGB,
		DefaultNetworkType: DefaultVMNetworkType,
		DefaultBridgeName:  "",
		MonitorInterval:    30 * time.Second,
		VMStartTimeout:     60 * time.Second,
		VMStopTimeout:      30 * time.Second,
		VMDeleteTimeout:    60 * time.Second,
		VMShutdownTimeout:  30 * time.Second,
		TerraformPath:      "terraform",
		TerraformWorkDir:   defaultTerraformWorkDir,
		TerraformTimeout:   300 * time.Second,
		TerraformParallel:  4,
		BaseImagePath:      defaultBaseImagePath,
		SnapshotPath:       defaultSnapshotPath,
	}
}

// Validate validates the service configuration
func (c *ServiceConfig) Validate() error {
	// Ensure required directories exist
	dirs := []string{
		c.DefaultVMPath,
		c.TerraformWorkDir,
		c.BaseImagePath,
		c.SnapshotPath,
	}

	for _, dir := range dirs {
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
	}

	return nil
}

// ApplyDefaults applies default configuration values to a VM request
func ApplyDefaults(cfg *config.C, request *types.VMRequest) {
	// Set default values for the request itself
	if request.ID == "" {
		request.ID = generateVMID()
	}
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now()
	}
	if request.Config == nil {
		request.Config = &types.VMConfig{}
	}

	// Get default configuration from config file
	defaultConfig := cfg.GetVirtualBoxDefaultConfig()

	// Apply default config values
	if request.Config.MemoryMB == 0 {
		if memoryMB, ok := defaultConfig["memory_mb"].(int); ok {
			request.Config.MemoryMB = memoryMB
		} else {
			request.Config.MemoryMB = config.DefaultVMMemoryMB
		}
	}

	if request.Config.CPUs == 0 {
		if cpus, ok := defaultConfig["cpus"].(int); ok {
			request.Config.CPUs = cpus
		} else {
			request.Config.CPUs = config.DefaultVMCPUs
		}
	}

	if request.Config.DiskSizeGB == 0 {
		if diskSizeGB, ok := defaultConfig["disk_size_gb"].(int); ok {
			request.Config.DiskSizeGB = diskSizeGB
		} else {
			request.Config.DiskSizeGB = config.DefaultVMDiskSizeGB
		}
	}

	if request.Config.OSType == "" {
		if osType, ok := defaultConfig["os_type"].(string); ok {
			request.Config.OSType = osType
		} else {
			request.Config.OSType = config.DefaultVMOSType
		}
	}

	if request.Config.NetworkType == "" {
		if networkType, ok := defaultConfig["network_type"].(string); ok {
			request.Config.NetworkType = networkType
		} else {
			request.Config.NetworkType = config.DefaultVMNetworkType
		}
	}

	if request.Config.StorageController == "" {
		if storageController, ok := defaultConfig["storage_controller"].(string); ok {
			request.Config.StorageController = storageController
		} else {
			request.Config.StorageController = config.DefaultVMStorageController
		}
	}

	if request.Config.StorageType == "" {
		if storageType, ok := defaultConfig["storage_type"].(string); ok {
			request.Config.StorageType = storageType
		} else {
			request.Config.StorageType = config.DefaultVMStorageType
		}
	}

	// Apply boolean settings
	if !request.Config.EnableAudio {
		if enableAudio, ok := defaultConfig["enable_audio"].(bool); ok {
			request.Config.EnableAudio = enableAudio
		} else {
			request.Config.EnableAudio = config.DefaultVMEnableAudio
		}
	}

	if !request.Config.EnableUSB {
		if enableUSB, ok := defaultConfig["enable_usb"].(bool); ok {
			request.Config.EnableUSB = enableUSB
		} else {
			request.Config.EnableUSB = config.DefaultVMEnableUSB
		}
	}

	if !request.Config.EnableVRDE {
		if enableVRDE, ok := defaultConfig["enable_vrde"].(bool); ok {
			request.Config.EnableVRDE = enableVRDE
		} else {
			request.Config.EnableVRDE = config.DefaultVMEnableVRDE
		}
	}

	if request.Config.VRDEPort == 0 {
		if vrdePort, ok := defaultConfig["vrde_port"].(int); ok {
			request.Config.VRDEPort = vrdePort
		} else {
			request.Config.VRDEPort = config.DefaultVMVRDEPort
		}
	}

	if !request.Config.EnablePAE {
		if enablePAE, ok := defaultConfig["enable_pae"].(bool); ok {
			request.Config.EnablePAE = enablePAE
		} else {
			request.Config.EnablePAE = config.DefaultVMEnablePAE
		}
	}

	if !request.Config.EnableNestedPaging {
		if enableNestedPaging, ok := defaultConfig["enable_nested_paging"].(bool); ok {
			request.Config.EnableNestedPaging = enableNestedPaging
		} else {
			request.Config.EnableNestedPaging = config.DefaultVMEnableNestedPaging
		}
	}

	// Apply custom settings
	if request.Config.CustomSettings == nil {
		request.Config.CustomSettings = make(map[string]string)
	}
	if customSettings, ok := defaultConfig["custom_settings"].(map[string]interface{}); ok {
		for k, v := range customSettings {
			if strVal, ok := v.(string); ok {
				request.Config.CustomSettings[k] = strVal
			}
		}
	}
}

// generateVMID generates a unique VM ID
func generateVMID() string {
	return "vm-" + time.Now().Format("20060102150405")
}
