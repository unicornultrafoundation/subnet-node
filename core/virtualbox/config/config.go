package config

import (
	"fmt"
	"time"

	"os/exec"

	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/utils"
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
	VBoxHeadless bool

	// VM configuration
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
	TerraformTimeout  time.Duration
	TerraformParallel int

	// Path detector for dynamic path resolution
	pathDetector *utils.PathDetector
}

// NewServiceConfigFromConfig creates a service config from the main config
func NewServiceConfigFromConfig(cfg *config.C) (*ServiceConfig, error) {
	serviceConfig := DefaultServiceConfig()

	// VirtualBox configuration
	serviceConfig.VBoxHeadless = cfg.GetBool("virtualbox.headless", true)

	// VM configuration
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
	serviceConfig.TerraformTimeout = cfg.GetDuration("virtualbox.terraform_timeout", 300*time.Second)
	serviceConfig.TerraformParallel = cfg.GetInt("virtualbox.terraform_parallel", 4)

	if err := serviceConfig.Validate(); err != nil {
		return nil, err
	}

	return serviceConfig, nil
}

// DefaultServiceConfig returns a default service configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		VBoxHeadless:       true,
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
		TerraformTimeout:   300 * time.Second,
		TerraformParallel:  4,
		pathDetector:       utils.NewPathDetector(),
	}
}

// Validate validates the service configuration
func (c *ServiceConfig) Validate() error {
	// Validate that VBoxManage is available
	if !c.IsVBoxManageAvailable() {
		return fmt.Errorf("VBoxManage is not available. Please install VirtualBox")
	}

	return nil
}

// GetVBoxManagePath returns the detected VBoxManage path
func (c *ServiceConfig) GetVBoxManagePath() string {
	return c.pathDetector.DetectVBoxManagePath()
}

// GetDefaultVMPath returns the detected default VM path
func (c *ServiceConfig) GetDefaultVMPath() string {
	return c.pathDetector.DetectDefaultVMPath()
}

// GetTerraformPath returns the detected terraform path
func (c *ServiceConfig) GetTerraformPath() string {
	return c.pathDetector.DetectTerraformPath()
}

// GetTerraformWorkDir returns the detected terraform working directory
func (c *ServiceConfig) GetTerraformWorkDir() string {
	return c.pathDetector.DetectTerraformWorkDir()
}

// GetBaseImagePath returns the detected base image path
func (c *ServiceConfig) GetBaseImagePath() string {
	return c.pathDetector.DetectBaseImagePath()
}

// GetSnapshotPath returns the detected snapshot path
func (c *ServiceConfig) GetSnapshotPath() string {
	return c.pathDetector.DetectSnapshotPath()
}

// IsVBoxManageAvailable checks if VBoxManage is available
func (c *ServiceConfig) IsVBoxManageAvailable() bool {
	path := c.GetVBoxManagePath()
	if path == "VBoxManage" {
		// If we fallback to just "VBoxManage", test if it's available
		cmd := exec.Command("VBoxManage", "--version")
		return cmd.Run() == nil
	}
	return true
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
