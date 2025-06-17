package config

import (
	"fmt"
	"time"

	"github.com/unicornultrafoundation/subnet-node/config"
)

// KVMConfig represents the KVM service configuration
type KVMConfig struct {
	Enabled    bool                `mapstructure:"enabled"`
	LibvirtURI string              `mapstructure:"libvirt_uri"`
	Limits     ResourceLimits      `mapstructure:"limits"`
	Storage    StorageConfig       `mapstructure:"storage"`
	Networks   map[string]Network  `mapstructure:"networks"`
	Templates  map[string]Template `mapstructure:"templates"`
	Security   SecurityConfig      `mapstructure:"security"`
	Monitoring MonitoringConfig    `mapstructure:"monitoring"`
}

// ResourceLimits defines resource constraints
type ResourceLimits struct {
	MaxCPUCores  int `mapstructure:"max_cpu_cores"`
	MaxMemoryMB  int `mapstructure:"max_memory_mb"`
	MaxDiskGB    int `mapstructure:"max_disk_gb"`
	MaxInstances int `mapstructure:"max_instances"`
}

// StorageConfig defines storage configuration
type StorageConfig struct {
	BasePath      string        `mapstructure:"base_path"`
	DefaultPool   string        `mapstructure:"default_pool"`
	ImageFormat   string        `mapstructure:"image_format"`
	CleanupPolicy CleanupPolicy `mapstructure:"cleanup_policy"`
}

// CleanupPolicy defines how to handle resource cleanup
type CleanupPolicy struct {
	AutoCleanup   bool          `mapstructure:"auto_cleanup"`
	RetentionTime time.Duration `mapstructure:"retention_time"`
}

// Network defines network configuration
type Network struct {
	Type        string   `mapstructure:"type"`
	CIDR        string   `mapstructure:"cidr"`
	Gateway     string   `mapstructure:"gateway"`
	Netmask     string   `mapstructure:"netmask"`
	DNS         []string `mapstructure:"dns"`
	Bridge      string   `mapstructure:"bridge"`
	VLAN        int      `mapstructure:"vlan"`
	Isolated    bool     `mapstructure:"isolated"`
	DHCPEnabled bool     `mapstructure:"dhcp_enabled"`
}

// Template defines VM template configuration
type Template struct {
	Name        string     `mapstructure:"name"`
	Description string     `mapstructure:"description"`
	ImagePath   string     `mapstructure:"image_path"`
	MinCPU      int        `mapstructure:"min_cpu"`
	MinMemoryMB int        `mapstructure:"min_memory_mb"`
	MinDiskGB   int        `mapstructure:"min_disk_gb"`
	CloudInit   *CloudInit `mapstructure:"cloud_init"`
	OSType      string     `mapstructure:"os_type"`
	Arch        string     `mapstructure:"arch"`
}

// CloudInit defines cloud-init configuration
type CloudInit struct {
	UserData string `mapstructure:"user_data"`
	MetaData string `mapstructure:"meta_data"`
}

// SecurityConfig defines security settings
type SecurityConfig struct {
	EnableSELinux   bool          `mapstructure:"enable_selinux"`
	AllowedUsers    []string      `mapstructure:"allowed_users"`
	IsolateNetworks bool          `mapstructure:"isolate_networks"`
	RequireAuth     bool          `mapstructure:"require_auth"`
	MaxSessionTime  time.Duration `mapstructure:"max_session_time"`
}

// MonitoringConfig defines monitoring settings
type MonitoringConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	MetricsInterval   time.Duration `mapstructure:"metrics_interval"`
	LogLevel          string        `mapstructure:"log_level"`
	EnableProfiling   bool          `mapstructure:"enable_profiling"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
}

// LoadConfig loads KVM configuration from config.C
func LoadConfig(cfg *config.C) (*KVMConfig, error) {
	kvmConfig := &KVMConfig{
		// Set defaults
		Enabled:    cfg.GetBool("kvm.enabled", false),
		LibvirtURI: cfg.GetString("kvm.libvirt_uri", "qemu:///system"),
		Limits: ResourceLimits{
			MaxCPUCores:  cfg.GetInt("kvm.limits.max_cpu_cores", 8),
			MaxMemoryMB:  cfg.GetInt("kvm.limits.max_memory_mb", 8192),
			MaxDiskGB:    cfg.GetInt("kvm.limits.max_disk_gb", 100),
			MaxInstances: cfg.GetInt("kvm.limits.max_instances", 10),
		},
		Storage: StorageConfig{
			BasePath:    cfg.GetString("kvm.storage.base_path", "/var/lib/subnet-node/kvm"),
			DefaultPool: cfg.GetString("kvm.storage.default_pool", "default"),
			ImageFormat: cfg.GetString("kvm.storage.image_format", "qcow2"),
			CleanupPolicy: CleanupPolicy{
				AutoCleanup:   cfg.GetBool("kvm.storage.cleanup_policy.auto_cleanup", true),
				RetentionTime: cfg.GetDuration("kvm.storage.cleanup_policy.retention_time", 24*time.Hour),
			},
		},
		Security: SecurityConfig{
			EnableSELinux:   cfg.GetBool("kvm.security.enable_selinux", false),
			AllowedUsers:    cfg.GetStringSlice("kvm.security.allowed_users", []string{}),
			IsolateNetworks: cfg.GetBool("kvm.security.isolate_networks", true),
			RequireAuth:     cfg.GetBool("kvm.security.require_auth", false),
			MaxSessionTime:  cfg.GetDuration("kvm.security.max_session_time", 2*time.Hour),
		},
		Monitoring: MonitoringConfig{
			Enabled:           cfg.GetBool("kvm.monitoring.enabled", true),
			MetricsInterval:   cfg.GetDuration("kvm.monitoring.metrics_interval", 30*time.Second),
			LogLevel:          cfg.GetString("kvm.monitoring.log_level", "info"),
			EnableProfiling:   cfg.GetBool("kvm.monitoring.enable_profiling", false),
			HealthCheckPeriod: cfg.GetDuration("kvm.monitoring.health_check_period", 60*time.Second),
		},
		Networks:  make(map[string]Network),
		Templates: make(map[string]Template),
	}

	// Load networks from configuration
	if cfg.IsSet("kvm.networks") {
		networksData := cfg.Get("kvm.networks")
		if networksMap, ok := networksData.(map[string]interface{}); ok {
			for name, networkData := range networksMap {
				if networkConfig, ok := networkData.(map[string]interface{}); ok {
					network := Network{
						Type:        getStringFromInterface(networkConfig["type"], "nat"),
						CIDR:        getStringFromInterface(networkConfig["cidr"], ""),
						Gateway:     getStringFromInterface(networkConfig["gateway"], ""),
						Netmask:     getStringFromInterface(networkConfig["netmask"], ""),
						Bridge:      getStringFromInterface(networkConfig["bridge"], ""),
						VLAN:        getIntFromInterface(networkConfig["vlan"], 0),
						Isolated:    getBoolFromInterface(networkConfig["isolated"], false),
						DHCPEnabled: getBoolFromInterface(networkConfig["dhcp_enabled"], true),
					}

					// Handle DNS array
					if dnsData, exists := networkConfig["dns"]; exists {
						if dnsSlice, ok := dnsData.([]interface{}); ok {
							for _, dns := range dnsSlice {
								if dnsStr, ok := dns.(string); ok {
									network.DNS = append(network.DNS, dnsStr)
								}
							}
						}
					}

					kvmConfig.Networks[name] = network
				}
			}
		}
	}

	// Load templates from configuration
	if cfg.IsSet("kvm.templates") {
		templatesData := cfg.Get("kvm.templates")
		if templatesMap, ok := templatesData.(map[string]interface{}); ok {
			for name, templateData := range templatesMap {
				if templateConfig, ok := templateData.(map[string]interface{}); ok {
					template := Template{
						Name:        name,
						Description: getStringFromInterface(templateConfig["description"], ""),
						ImagePath:   getStringFromInterface(templateConfig["image_path"], ""),
						MinCPU:      getIntFromInterface(templateConfig["min_cpu"], 1),
						MinMemoryMB: getIntFromInterface(templateConfig["min_memory_mb"], 1024),
						MinDiskGB:   getIntFromInterface(templateConfig["min_disk_gb"], 10),
						OSType:      getStringFromInterface(templateConfig["os_type"], "linux"),
						Arch:        getStringFromInterface(templateConfig["arch"], "x86_64"),
					}

					// Handle cloud-init configuration
					if cloudInitData, exists := templateConfig["cloud_init"]; exists {
						if cloudInitConfig, ok := cloudInitData.(map[string]interface{}); ok {
							template.CloudInit = &CloudInit{
								UserData: getStringFromInterface(cloudInitConfig["user_data"], ""),
								MetaData: getStringFromInterface(cloudInitConfig["meta_data"], ""),
							}
						}
					}

					kvmConfig.Templates[name] = template
				}
			}
		}
	}

	// Load default network if none configured
	if len(kvmConfig.Networks) == 0 {
		kvmConfig.Networks["default"] = Network{
			Type:        "nat",
			CIDR:        "192.168.122.0/24",
			Gateway:     "192.168.122.1",
			DNS:         []string{"8.8.8.8", "8.8.4.4"},
			DHCPEnabled: true,
		}
	}

	return kvmConfig, nil
}

// Helper functions to safely extract values from interface{}
func getStringFromInterface(val interface{}, defaultVal string) string {
	if val == nil {
		return defaultVal
	}
	if str, ok := val.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", val)
}

func getIntFromInterface(val interface{}, defaultVal int) int {
	if val == nil {
		return defaultVal
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := fmt.Sscanf(v, "%d", &defaultVal); err == nil && i == 1 {
			return defaultVal
		}
	}
	return defaultVal
}

func getBoolFromInterface(val interface{}, defaultVal bool) bool {
	if val == nil {
		return defaultVal
	}
	if b, ok := val.(bool); ok {
		return b
	}
	if str, ok := val.(string); ok {
		switch str {
		case "true", "yes", "y", "1":
			return true
		case "false", "no", "n", "0":
			return false
		}
	}
	return defaultVal
}

// GetNetworkConfig returns network configuration by name
func (c *KVMConfig) GetNetworkConfig(name string) (*Network, bool) {
	network, exists := c.Networks[name]
	return &network, exists
}

// GetTemplateConfig returns template configuration by name
func (c *KVMConfig) GetTemplateConfig(name string) (*Template, bool) {
	template, exists := c.Templates[name]
	return &template, exists
}

// ListNetworks returns all configured networks
func (c *KVMConfig) ListNetworks() map[string]Network {
	return c.Networks
}

// ListTemplates returns all configured templates
func (c *KVMConfig) ListTemplates() map[string]Template {
	return c.Templates
}

// IsEnabled returns whether KVM service is enabled
func (c *KVMConfig) IsEnabled() bool {
	return c.Enabled
}

// GetStoragePath returns the full path for a specific storage item
func (c *KVMConfig) GetStoragePath(subPath string) string {
	if subPath == "" {
		return c.Storage.BasePath
	}
	return fmt.Sprintf("%s/%s", c.Storage.BasePath, subPath)
}

// GetVMStoragePath returns the storage path for a specific VM
func (c *KVMConfig) GetVMStoragePath(vmID string) string {
	return c.GetStoragePath(fmt.Sprintf("vms/%s", vmID))
}

// GetTemplateStoragePath returns the storage path for templates
func (c *KVMConfig) GetTemplateStoragePath() string {
	return c.GetStoragePath("templates")
}

// GetString returns a string configuration value
func (c *KVMConfig) GetString(key string) string {
	switch key {
	case "libvirt_uri":
		return c.LibvirtURI
	case "storage.base_path":
		return c.Storage.BasePath
	case "storage.default_pool":
		return c.Storage.DefaultPool
	case "storage.image_format":
		return c.Storage.ImageFormat
	case "monitoring.log_level":
		return c.Monitoring.LogLevel
	default:
		return ""
	}
}

// GetInt returns an integer configuration value
func (c *KVMConfig) GetInt(key string) int {
	switch key {
	case "limits.max_cpu_cores":
		return c.Limits.MaxCPUCores
	case "limits.max_memory_mb":
		return c.Limits.MaxMemoryMB
	case "limits.max_disk_gb":
		return c.Limits.MaxDiskGB
	case "limits.max_instances":
		return c.Limits.MaxInstances
	default:
		return 0
	}
}

// GetBool returns a boolean configuration value
func (c *KVMConfig) GetBool(key string) bool {
	switch key {
	case "enabled":
		return c.Enabled
	case "storage.cleanup_policy.auto_cleanup":
		return c.Storage.CleanupPolicy.AutoCleanup
	case "security.enable_selinux":
		return c.Security.EnableSELinux
	case "security.isolate_networks":
		return c.Security.IsolateNetworks
	case "security.require_auth":
		return c.Security.RequireAuth
	case "monitoring.enabled":
		return c.Monitoring.Enabled
	case "monitoring.enable_profiling":
		return c.Monitoring.EnableProfiling
	default:
		return false
	}
}

// GetDuration returns a duration configuration value
func (c *KVMConfig) GetDuration(key string) time.Duration {
	switch key {
	case "storage.cleanup_policy.retention_time":
		return c.Storage.CleanupPolicy.RetentionTime
	case "security.max_session_time":
		return c.Security.MaxSessionTime
	case "monitoring.metrics_interval":
		return c.Monitoring.MetricsInterval
	case "monitoring.health_check_period":
		return c.Monitoring.HealthCheckPeriod
	default:
		return 0
	}
}
