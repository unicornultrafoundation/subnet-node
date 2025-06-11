package apps

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ContainerACLConfig represents the structure of the container allow IPs configuration
type ContainerACLConfig struct {
	Global     GlobalACLConfig                    `json:"global"`
	Containers map[string]ContainerSpecificConfig `json:"containers"`
}

// GlobalACLConfig represents global allow IP settings
type GlobalACLConfig struct {
	AllowIPs []string `json:"allow_ips"`
	Enabled  bool     `json:"enabled"`
}

// ContainerSpecificConfig represents container-specific allow IP settings
type ContainerSpecificConfig struct {
	AllowIPs []string `json:"allow_ips"`
	Enabled  bool     `json:"enabled"`
}

// ContainerACLManager manages container access control lists
type ContainerACLManager struct {
	config         *ContainerACLConfig
	configPath     string
	mu             sync.RWMutex
	lastModified   time.Time
	autoReload     bool
	reloadInterval time.Duration
	stopReload     chan struct{}
}

// NewContainerACLManager creates a new container ACL manager
func NewContainerACLManager(configPath string) (*ContainerACLManager, error) {
	manager := &ContainerACLManager{
		configPath:     configPath,
		autoReload:     true,
		reloadInterval: 5 * time.Second, // Check for file changes every 5 seconds
		stopReload:     make(chan struct{}),
	}

	err := manager.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Start auto-reload watcher if enabled
	if manager.autoReload {
		go manager.watchConfigFile()
	}

	return manager, nil
}

// watchConfigFile monitors the config file for changes and auto-reloads
func (m *ContainerACLManager) watchConfigFile() {
	ticker := time.NewTicker(m.reloadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if m.checkFileModification() {
				if err := m.LoadConfig(); err != nil {
					log.Warnf("Failed to reload container ACL config: %v", err)
				} else {
					log.Info("Container ACL configuration reloaded due to file change")
				}
			}
		case <-m.stopReload:
			return
		}
	}
}

// checkFileModification checks if the config file has been modified
func (m *ContainerACLManager) checkFileModification() bool {
	info, err := os.Stat(m.configPath)
	if err != nil {
		return false
	}

	modTime := info.ModTime()
	if modTime.After(m.lastModified) {
		m.lastModified = modTime
		return true
	}
	return false
}

// StopAutoReload stops the automatic config file monitoring
func (m *ContainerACLManager) StopAutoReload() {
	close(m.stopReload)
}

// LoadConfig loads the container ACL configuration from the JSON file
func (m *ContainerACLManager) LoadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Create default configuration if file doesn't exist
		m.config = m.createDefaultConfig()
		return m.saveConfigToFile()
	}

	file, err := os.Open(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to open container ACL config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read container ACL config file: %w", err)
	}

	var config ContainerACLConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("failed to parse container ACL config: %w", err)
	}

	m.config = &config
	return nil
}

// SaveConfig saves the current configuration to the JSON file
func (m *ContainerACLManager) saveConfigToFile() error {
	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to create container ACL config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(m.config)
	if err != nil {
		return fmt.Errorf("failed to write container ACL config: %w", err)
	}

	return nil
}

// createDefaultConfig creates a default configuration
func (m *ContainerACLManager) createDefaultConfig() *ContainerACLConfig {
	return &ContainerACLConfig{
		Global: GlobalACLConfig{
			AllowIPs: []string{"192.168.0.0/16", "10.0.0.0/8", "172.16.0.0/12"},
			Enabled:  true,
		},
		Containers: make(map[string]ContainerSpecificConfig),
	}
}

// GetContainerAllowIPs returns the allow IPs for a specific container
func (m *ContainerACLManager) GetContainerAllowIPs(containerName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// If global ACL is disabled, return empty list (no checking)
	if !m.config.Global.Enabled {
		return []string{}
	}

	// Check if container has specific configuration
	if containerConfig, exists := m.config.Containers[containerName]; exists {
		// If container-specific ACL is enabled, use container settings
		if containerConfig.Enabled {
			return containerConfig.AllowIPs
		}
	}

	// Fall back to global settings
	return m.config.Global.AllowIPs
}

// GetContainerAllowIPsByID returns the allow IPs for a container by its Docker container ID
// This method first tries to find the container by ID, then falls back to the pattern-based lookup
func (m *ContainerACLManager) GetContainerAllowIPsByID(containerID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// If global ACL is disabled, return empty list (no checking)
	if !m.config.Global.Enabled {
		return []string{}
	}

	// First, try to find container config by exact container ID
	if containerConfig, exists := m.config.Containers[containerID]; exists {
		if containerConfig.Enabled {
			return containerConfig.AllowIPs
		}
	}

	// If not found by ID, try to find by container ID as key
	// Docker container IDs are long hashes, but we might use short forms
	for configName, containerConfig := range m.config.Containers {
		// Check if the config name starts with the container ID (for short container ID matching)
		if strings.HasPrefix(containerID, configName) || strings.HasPrefix(configName, containerID) {
			if containerConfig.Enabled {
				return containerConfig.AllowIPs
			}
		}
	}

	// Fall back to global settings
	return m.config.Global.AllowIPs
}

// IsIPAllowed checks if the given IP is allowed for the specified container
func (m *ContainerACLManager) IsIPAllowed(ip string, containerName string) bool {
	allowIPs := m.GetContainerAllowIPs(containerName)
	return isIPInAllowList(ip, allowIPs)
}

// IsIPAllowedByContainerID checks if the given IP is allowed for the specified container by container ID
func (m *ContainerACLManager) IsIPAllowedByContainerID(ip string, containerID string) bool {
	allowIPs := m.GetContainerAllowIPsByID(containerID)
	return isIPInAllowList(ip, allowIPs)
}

// AddContainerConfig adds or updates container-specific configuration
func (m *ContainerACLManager) AddContainerConfig(containerName string, config ContainerSpecificConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Containers == nil {
		m.config.Containers = make(map[string]ContainerSpecificConfig)
	}

	m.config.Containers[containerName] = config
	return m.saveConfigToFile()
}

// RemoveContainerConfig removes container-specific configuration
func (m *ContainerACLManager) RemoveContainerConfig(containerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.config.Containers, containerName)
	return m.saveConfigToFile()
}

// GetConfig returns a copy of the current configuration
func (m *ContainerACLManager) GetConfig() ContainerACLConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return *m.config
}

// isIPInAllowList checks if the given IP is in the allowed list (supports CIDR and single IP)
func isIPInAllowList(ip string, allowList []string) bool {
	if len(allowList) == 0 {
		return false // If no allow list, not allowed
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, cidr := range allowList {
		if strings.Contains(cidr, "/") {
			_, ipnet, err := net.ParseCIDR(cidr)
			if err == nil && ipnet.Contains(parsedIP) {
				return true
			}
		} else {
			if cidr == ip {
				return true
			}
		}
	}
	return false
}

// ReloadConfig reloads the configuration from the file
func (m *ContainerACLManager) ReloadConfig() error {
	return m.LoadConfig()
}
