package config

import (
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
)

var log = logrus.WithField("service", "vpn-config")

// ConfigService is the interface for the centralized configuration service
type ConfigService interface {
	// Basic settings
	GetEnable() bool
	GetMTU() int
	GetVirtualIP() string
	GetSubnet() string
	GetRoutes() []string
	GetProtocol() string
	GetRoutines() int

	// Security settings
	GetConntrackCacheTimeout() time.Duration

	// Stream pool settings
	GetStreamIdleTimeout() time.Duration
	GetStreamCleanupInterval() time.Duration

	// Circuit breaker settings
	GetCircuitBreakerFailureThreshold() int
	GetCircuitBreakerResetTimeout() time.Duration
	GetCircuitBreakerSuccessThreshold() int

	// Retry settings
	GetRetryMaxAttempts() int
	GetRetryInitialInterval() time.Duration
	GetRetryMaxInterval() time.Duration

	// Timeout settings
	GetDHTSyncTimeout() time.Duration
	GetTUNSetupTimeout() time.Duration
	GetPeerConnectionCheckInterval() time.Duration
	GetShutdownGracePeriod() time.Duration

	// Get the full VPN config
	GetVPNConfig() *VPNConfig

	// Update configuration
	UpdateConfig(cfg *config.C) error

	// Update configuration with callback
	UpdateConfigWithCallback(cfg *config.C, callback func(*VPNConfig, *VPNConfig)) error
}

// ConfigServiceImpl is the implementation of the ConfigService interface
type ConfigServiceImpl struct {
	// Configuration
	vpnConfig *VPNConfig
	// Mutex to protect access to the configuration
	mu sync.RWMutex
}

// NewConfigService creates a new configuration service
func NewConfigService(cfg *config.C) *ConfigServiceImpl {
	return &ConfigServiceImpl{
		vpnConfig: New(cfg),
	}
}

// Basic settings

// GetEnable returns whether the VPN is enabled
func (c *ConfigServiceImpl) GetEnable() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.Enable
}

// GetMTU returns the MTU
func (c *ConfigServiceImpl) GetMTU() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.MTU
}

// GetVirtualIP returns the virtual IP
func (c *ConfigServiceImpl) GetVirtualIP() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.VirtualIP
}

// GetSubnet returns the subnet as a string (CIDR notation)
func (c *ConfigServiceImpl) GetSubnet() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return strconv.Itoa(c.vpnConfig.Subnet)
}

// GetRoutes returns the routes
func (c *ConfigServiceImpl) GetRoutes() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.Routes
}

// GetProtocol returns the protocol
func (c *ConfigServiceImpl) GetProtocol() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.Protocol
}

// GetRoutines returns the number of reader routines
func (c *ConfigServiceImpl) GetRoutines() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.Routines
}

// Security settings

// GetConntrackCacheTimeout returns the conntrack cache timeout
func (c *ConfigServiceImpl) GetConntrackCacheTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.ConntrackCacheTimeout
}

// Stream settings

// GetStreamIdleTimeout returns the stream idle timeout
func (c *ConfigServiceImpl) GetStreamIdleTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.StreamIdleTimeout
}

// GetStreamCleanupInterval returns the stream cleanup interval
func (c *ConfigServiceImpl) GetStreamCleanupInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.StreamCleanupInterval // Reuse the same interval for now
}

// Circuit breaker settings

// GetCircuitBreakerFailureThreshold returns the circuit breaker failure threshold
func (c *ConfigServiceImpl) GetCircuitBreakerFailureThreshold() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.CircuitBreakerFailureThreshold
}

// GetCircuitBreakerResetTimeout returns the circuit breaker reset timeout
func (c *ConfigServiceImpl) GetCircuitBreakerResetTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.CircuitBreakerResetTimeout
}

// GetCircuitBreakerSuccessThreshold returns the circuit breaker success threshold
func (c *ConfigServiceImpl) GetCircuitBreakerSuccessThreshold() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.CircuitBreakerSuccessThreshold
}

// Retry settings

// GetRetryMaxAttempts returns the maximum number of retry attempts
func (c *ConfigServiceImpl) GetRetryMaxAttempts() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.RetryMaxAttempts
}

// GetRetryInitialInterval returns the initial retry interval
func (c *ConfigServiceImpl) GetRetryInitialInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.RetryInitialInterval
}

// GetRetryMaxInterval returns the maximum retry interval
func (c *ConfigServiceImpl) GetRetryMaxInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.RetryMaxInterval
}

// Timeout settings

// GetDHTSyncTimeout returns the DHT sync timeout
func (c *ConfigServiceImpl) GetDHTSyncTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.DHTSyncTimeout
}

// GetTUNSetupTimeout returns the TUN setup timeout
func (c *ConfigServiceImpl) GetTUNSetupTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.TUNSetupTimeout
}

// GetPeerConnectionCheckInterval returns the peer connection check interval
func (c *ConfigServiceImpl) GetPeerConnectionCheckInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.PeerConnectionCheckInterval
}

// GetShutdownGracePeriod returns the shutdown grace period
func (c *ConfigServiceImpl) GetShutdownGracePeriod() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.ShutdownGracePeriod
}

// GetVPNConfig returns the full VPN config
func (c *ConfigServiceImpl) GetVPNConfig() *VPNConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig
}

// UpdateConfig updates the configuration
func (c *ConfigServiceImpl) UpdateConfig(cfg *config.C) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a new configuration
	newConfig := New(cfg)

	// Validate the new configuration
	if err := newConfig.Validate(); err != nil {
		return err
	}

	// Update the configuration
	c.vpnConfig = newConfig

	return nil
}

// UpdateConfigWithCallback updates the configuration and calls the callback function
// with the old and new configurations if the update is successful
func (c *ConfigServiceImpl) UpdateConfigWithCallback(cfg *config.C, callback func(*VPNConfig, *VPNConfig)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a new configuration
	newConfig := New(cfg)

	// Validate the new configuration
	if err := newConfig.Validate(); err != nil {
		return err
	}

	// Store the old configuration
	oldConfig := c.vpnConfig

	// Update the configuration
	c.vpnConfig = newConfig

	// Call the callback function
	if callback != nil {
		callback(oldConfig, newConfig)
	}

	return nil
}
