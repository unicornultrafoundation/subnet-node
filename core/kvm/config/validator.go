package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// Validator provides configuration validation
type Validator struct {
	config *KVMConfig
}

// NewValidator creates a new configuration validator
func NewValidator(config *KVMConfig) *Validator {
	return &Validator{config: config}
}

// Validate performs comprehensive validation of KVM configuration
func (v *Validator) Validate() error {
	if !v.config.Enabled {
		return nil // Skip validation if KVM is disabled
	}

	validators := []func() error{
		v.validateBasicConfig,
		v.validateResourceLimits,
		v.validateStorageConfig,
		v.validateNetworkConfigs,
		v.validateTemplateConfigs,
		v.validateSecurityConfig,
		v.validateMonitoringConfig,
	}

	for _, validator := range validators {
		if err := validator(); err != nil {
			return err
		}
	}

	return nil
}

// validateBasicConfig validates basic KVM configuration
func (v *Validator) validateBasicConfig() error {
	if v.config.LibvirtURI == "" {
		return fmt.Errorf("libvirt_uri cannot be empty")
	}

	// Validate libvirt URI format
	validPrefixes := []string{"qemu://", "qemu+tcp://", "qemu+ssh://", "qemu:///"}
	validURI := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(v.config.LibvirtURI, prefix) {
			validURI = true
			break
		}
	}

	if !validURI {
		return fmt.Errorf("invalid libvirt_uri format: %s", v.config.LibvirtURI)
	}

	return nil
}

// validateResourceLimits validates resource limit configuration
func (v *Validator) validateResourceLimits() error {
	limits := v.config.Limits

	if limits.MaxCPUCores <= 0 {
		return fmt.Errorf("max_cpu_cores must be greater than 0")
	}

	if limits.MaxMemoryMB <= 0 {
		return fmt.Errorf("max_memory_mb must be greater than 0")
	}

	if limits.MaxDiskGB <= 0 {
		return fmt.Errorf("max_disk_gb must be greater than 0")
	}

	if limits.MaxInstances <= 0 {
		return fmt.Errorf("max_instances must be greater than 0")
	}

	// Reasonable upper bounds
	if limits.MaxCPUCores > 256 {
		return fmt.Errorf("max_cpu_cores cannot exceed 256")
	}

	if limits.MaxMemoryMB > 1024*1024 { // 1TB
		return fmt.Errorf("max_memory_mb cannot exceed 1048576 (1TB)")
	}

	if limits.MaxDiskGB > 10*1024 { // 10TB
		return fmt.Errorf("max_disk_gb cannot exceed 10240 (10TB)")
	}

	if limits.MaxInstances > 1000 {
		return fmt.Errorf("max_instances cannot exceed 1000")
	}

	return nil
}

// validateStorageConfig validates storage configuration
func (v *Validator) validateStorageConfig() error {
	storage := v.config.Storage

	if storage.BasePath == "" {
		return fmt.Errorf("storage base_path cannot be empty")
	}

	// Validate base path exists or can be created
	if err := v.validateOrCreatePath(storage.BasePath); err != nil {
		return fmt.Errorf("invalid storage base_path: %w", err)
	}

	// Validate image format
	validFormats := []string{"qcow2", "raw", "vmdk", "vdi"}
	if !contains(validFormats, storage.ImageFormat) {
		return fmt.Errorf("unsupported image format: %s (supported: %s)",
			storage.ImageFormat, strings.Join(validFormats, ", "))
	}

	// Validate cleanup policy
	if storage.CleanupPolicy.RetentionTime < 0 {
		return fmt.Errorf("cleanup retention_time cannot be negative")
	}

	return nil
}

// validateNetworkConfigs validates all network configurations
func (v *Validator) validateNetworkConfigs() error {
	if len(v.config.Networks) == 0 {
		return fmt.Errorf("at least one network must be configured")
	}

	for name, network := range v.config.Networks {
		if err := v.validateNetworkConfig(name, network); err != nil {
			return fmt.Errorf("network '%s': %w", name, err)
		}
	}

	return nil
}

// validateNetworkConfig validates a single network configuration
func (v *Validator) validateNetworkConfig(name string, network Network) error {
	if name == "" {
		return fmt.Errorf("network name cannot be empty")
	}

	// Validate network type
	validTypes := []string{"nat", "bridge", "isolated", "host"}
	if !contains(validTypes, network.Type) {
		return fmt.Errorf("invalid network type: %s (supported: %s)",
			network.Type, strings.Join(validTypes, ", "))
	}

	// Validate CIDR if provided
	if network.CIDR != "" {
		_, _, err := net.ParseCIDR(network.CIDR)
		if err != nil {
			return fmt.Errorf("invalid CIDR: %w", err)
		}
	}

	// Validate gateway IP if provided
	if network.Gateway != "" {
		if net.ParseIP(network.Gateway) == nil {
			return fmt.Errorf("invalid gateway IP: %s", network.Gateway)
		}
	}

	// Validate DNS servers
	for i, dns := range network.DNS {
		if net.ParseIP(dns) == nil {
			return fmt.Errorf("invalid DNS server at index %d: %s", i, dns)
		}
	}

	// Validate VLAN ID
	if network.VLAN < 0 || network.VLAN > 4094 {
		return fmt.Errorf("invalid VLAN ID: %d (must be 0-4094)", network.VLAN)
	}

	return nil
}

// validateTemplateConfigs validates all template configurations
func (v *Validator) validateTemplateConfigs() error {
	for name, template := range v.config.Templates {
		if err := v.validateTemplateConfig(name, template); err != nil {
			return fmt.Errorf("template '%s': %w", name, err)
		}
	}

	return nil
}

// validateTemplateConfig validates a single template configuration
func (v *Validator) validateTemplateConfig(name string, template Template) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if template.ImagePath == "" {
		return fmt.Errorf("image_path cannot be empty")
	}

	// Validate image path exists
	if _, err := os.Stat(template.ImagePath); err != nil {
		return fmt.Errorf("image path does not exist: %s", template.ImagePath)
	}

	// Validate minimum resource requirements
	if template.MinCPU <= 0 {
		return fmt.Errorf("min_cpu must be greater than 0")
	}

	if template.MinMemoryMB <= 0 {
		return fmt.Errorf("min_memory_mb must be greater than 0")
	}

	if template.MinDiskGB <= 0 {
		return fmt.Errorf("min_disk_gb must be greater than 0")
	}

	// Validate against global limits
	if template.MinCPU > v.config.Limits.MaxCPUCores {
		return fmt.Errorf("min_cpu (%d) exceeds global max_cpu_cores (%d)",
			template.MinCPU, v.config.Limits.MaxCPUCores)
	}

	if template.MinMemoryMB > v.config.Limits.MaxMemoryMB {
		return fmt.Errorf("min_memory_mb (%d) exceeds global max_memory_mb (%d)",
			template.MinMemoryMB, v.config.Limits.MaxMemoryMB)
	}

	if template.MinDiskGB > v.config.Limits.MaxDiskGB {
		return fmt.Errorf("min_disk_gb (%d) exceeds global max_disk_gb (%d)",
			template.MinDiskGB, v.config.Limits.MaxDiskGB)
	}

	// Validate OS type and architecture
	if template.OSType != "" {
		validOSTypes := []string{"linux", "windows", "freebsd", "other"}
		if !contains(validOSTypes, template.OSType) {
			return fmt.Errorf("invalid os_type: %s (supported: %s)",
				template.OSType, strings.Join(validOSTypes, ", "))
		}
	}

	if template.Arch != "" {
		validArchs := []string{"x86_64", "i386", "aarch64", "arm"}
		if !contains(validArchs, template.Arch) {
			return fmt.Errorf("invalid arch: %s (supported: %s)",
				template.Arch, strings.Join(validArchs, ", "))
		}
	}

	return nil
}

// validateSecurityConfig validates security configuration
func (v *Validator) validateSecurityConfig() error {
	security := v.config.Security

	// Validate session time
	if security.MaxSessionTime < 0 {
		return fmt.Errorf("max_session_time cannot be negative")
	}

	// Validate allowed users (basic validation)
	for i, user := range security.AllowedUsers {
		if user == "" {
			return fmt.Errorf("allowed user at index %d cannot be empty", i)
		}
		if strings.Contains(user, "/") || strings.Contains(user, "\\") {
			return fmt.Errorf("invalid allowed user at index %d: %s", i, user)
		}
	}

	return nil
}

// validateMonitoringConfig validates monitoring configuration
func (v *Validator) validateMonitoringConfig() error {
	monitoring := v.config.Monitoring

	// Validate metrics interval
	if monitoring.MetricsInterval < 0 {
		return fmt.Errorf("metrics_interval cannot be negative")
	}

	// Validate health check period
	if monitoring.HealthCheckPeriod < 0 {
		return fmt.Errorf("health_check_period cannot be negative")
	}

	// Validate log level
	if monitoring.LogLevel != "" {
		validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
		if !contains(validLevels, monitoring.LogLevel) {
			return fmt.Errorf("invalid log level: %s (supported: %s)",
				monitoring.LogLevel, strings.Join(validLevels, ", "))
		}
	}

	return nil
}

// validateOrCreatePath validates a path exists or can be created
func (v *Validator) validateOrCreatePath(path string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err == nil {
		// Path exists, check if it's a directory
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", absPath)
		}
		return nil
	}

	// Path doesn't exist, try to create it
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
