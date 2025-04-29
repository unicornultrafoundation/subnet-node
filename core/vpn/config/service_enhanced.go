package config

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
)

var configLog = logrus.WithField("service", "vpn-config")

// EnhancedConfigService extends the ConfigService with enhanced validation and standardization
type EnhancedConfigService struct {
	// Configuration
	vpnConfig *VPNConfig
	// Mutex to protect access to the configuration
	mu sync.RWMutex
	// Validation result
	validationResult ValidationResult
	// Whether to standardize configuration values
	standardize bool
}

// NewEnhancedConfigService creates a new enhanced configuration service
func NewEnhancedConfigService(cfg *config.C, standardize bool) *EnhancedConfigService {
	// Create the base configuration
	vpnConfig := New(cfg)
	
	// Standardize if requested
	if standardize {
		vpnConfig = StandardizeConfig(vpnConfig)
	}
	
	// Validate the configuration
	validationResult := vpnConfig.ValidateEnhanced()
	
	// Log validation results
	if !validationResult.Valid {
		configLog.WithError(validationResult.Error).Error("Invalid VPN configuration")
	}
	
	for _, warning := range validationResult.Warnings {
		configLog.Warn(warning)
	}
	
	return &EnhancedConfigService{
		vpnConfig:        vpnConfig,
		validationResult: validationResult,
		standardize:      standardize,
	}
}

// GetValidationResult returns the validation result
func (c *EnhancedConfigService) GetValidationResult() ValidationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.validationResult
}

// IsValid returns whether the configuration is valid
func (c *EnhancedConfigService) IsValid() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.validationResult.Valid
}

// GetWarnings returns any warnings about the configuration
func (c *EnhancedConfigService) GetWarnings() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.validationResult.Warnings
}

// UpdateConfig updates the configuration with enhanced validation
func (c *EnhancedConfigService) UpdateConfig(cfg *config.C) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Create a new configuration
	newConfig := New(cfg)
	
	// Standardize if requested
	if c.standardize {
		newConfig = StandardizeConfig(newConfig)
	}
	
	// Validate the new configuration
	validationResult := newConfig.ValidateEnhanced()
	
	// If the configuration is invalid, return the error
	if !validationResult.Valid {
		return validationResult.Error
	}
	
	// Log any warnings
	for _, warning := range validationResult.Warnings {
		configLog.Warn(warning)
	}
	
	// Update the configuration and validation result
	c.vpnConfig = newConfig
	c.validationResult = validationResult
	
	return nil
}

// UpdateConfigWithCallback updates the configuration with enhanced validation
// and calls the callback function with the old and new configurations
func (c *EnhancedConfigService) UpdateConfigWithCallback(
	cfg *config.C,
	callback func(*VPNConfig, *VPNConfig),
) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Create a new configuration
	newConfig := New(cfg)
	
	// Standardize if requested
	if c.standardize {
		newConfig = StandardizeConfig(newConfig)
	}
	
	// Validate the new configuration
	validationResult := newConfig.ValidateEnhanced()
	
	// If the configuration is invalid, return the error
	if !validationResult.Valid {
		return validationResult.Error
	}
	
	// Log any warnings
	for _, warning := range validationResult.Warnings {
		configLog.Warn(warning)
	}
	
	// Store the old configuration
	oldConfig := c.vpnConfig
	
	// Update the configuration and validation result
	c.vpnConfig = newConfig
	c.validationResult = validationResult
	
	// Call the callback function
	if callback != nil {
		callback(oldConfig, newConfig)
	}
	
	return nil
}

// GetVPNConfig returns the full VPN config
func (c *EnhancedConfigService) GetVPNConfig() *VPNConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig
}

// All the getter methods from the original ConfigService

// GetEnable returns whether the VPN is enabled
func (c *EnhancedConfigService) GetEnable() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.Enable
}

// GetMTU returns the MTU
func (c *EnhancedConfigService) GetMTU() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.MTU
}

// GetVirtualIP returns the virtual IP
func (c *EnhancedConfigService) GetVirtualIP() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.VirtualIP
}

// GetSubnet returns the subnet
func (c *EnhancedConfigService) GetSubnet() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return string(c.vpnConfig.Subnet)
}

// GetRoutes returns the routes
func (c *EnhancedConfigService) GetRoutes() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.Routes
}

// GetProtocol returns the protocol
func (c *EnhancedConfigService) GetProtocol() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.Protocol
}

// GetUnallowedPorts returns the unallowed ports
func (c *EnhancedConfigService) GetUnallowedPorts() map[string]bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.UnallowedPorts
}

// GetMaxStreamsPerPeer returns the maximum number of streams per peer
func (c *EnhancedConfigService) GetMaxStreamsPerPeer() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.MaxStreamsPerPeer
}

// GetStreamIdleTimeout returns the stream idle timeout
func (c *EnhancedConfigService) GetStreamIdleTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.StreamIdleTimeout
}

// GetCleanupInterval returns the cleanup interval
func (c *EnhancedConfigService) GetCleanupInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.CleanupInterval
}

// GetPacketBufferSize returns the packet buffer size
func (c *EnhancedConfigService) GetPacketBufferSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.PacketBufferSize
}

// GetUsageCountWeight returns the usage count weight
func (c *EnhancedConfigService) GetUsageCountWeight() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.UsageCountWeight
}

// GetBufferUtilWeight returns the buffer utilization weight
func (c *EnhancedConfigService) GetBufferUtilWeight() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.BufferUtilWeight
}

// GetBufferUtilThreshold returns the buffer utilization threshold
func (c *EnhancedConfigService) GetBufferUtilThreshold() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.BufferUtilThreshold
}

// GetUsageCountThreshold returns the usage count threshold
func (c *EnhancedConfigService) GetUsageCountThreshold() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.UsageCountThreshold
}

// GetCircuitBreakerFailureThreshold returns the circuit breaker failure threshold
func (c *EnhancedConfigService) GetCircuitBreakerFailureThreshold() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.CircuitBreakerFailureThreshold
}

// GetCircuitBreakerResetTimeout returns the circuit breaker reset timeout
func (c *EnhancedConfigService) GetCircuitBreakerResetTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.CircuitBreakerResetTimeout
}

// GetCircuitBreakerSuccessThreshold returns the circuit breaker success threshold
func (c *EnhancedConfigService) GetCircuitBreakerSuccessThreshold() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.CircuitBreakerSuccessThreshold
}

// GetRetryMaxAttempts returns the maximum number of retry attempts
func (c *EnhancedConfigService) GetRetryMaxAttempts() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.RetryMaxAttempts
}

// GetRetryInitialInterval returns the initial retry interval
func (c *EnhancedConfigService) GetRetryInitialInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.RetryInitialInterval
}

// GetRetryMaxInterval returns the maximum retry interval
func (c *EnhancedConfigService) GetRetryMaxInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.RetryMaxInterval
}

// GetDHTSyncTimeout returns the DHT sync timeout
func (c *EnhancedConfigService) GetDHTSyncTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.DHTSyncTimeout
}

// GetTUNSetupTimeout returns the TUN setup timeout
func (c *EnhancedConfigService) GetTUNSetupTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.TUNSetupTimeout
}

// GetPeerConnectionCheckInterval returns the peer connection check interval
func (c *EnhancedConfigService) GetPeerConnectionCheckInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.PeerConnectionCheckInterval
}

// GetShutdownGracePeriod returns the shutdown grace period
func (c *EnhancedConfigService) GetShutdownGracePeriod() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.vpnConfig.ShutdownGracePeriod
}
