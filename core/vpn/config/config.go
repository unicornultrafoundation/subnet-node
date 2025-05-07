package config

import (
	"strconv"
	"time"

	"github.com/unicornultrafoundation/subnet-node/config"
)

// parseFloat64 parses a string to a float64 or returns the default value if parsing fails
func parseFloat64(s string, defaultValue float64) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultValue
	}
	return v
}

// VPNConfig encapsulates all configuration settings for the VPN service
type VPNConfig struct {
	// Basic settings
	Enable    bool
	MTU       int
	VirtualIP string
	Subnet    int
	Routes    []string
	Protocol  string
	Routines  int

	// Security settings
	UnallowedPorts map[string]bool

	// Stream settings
	StreamIdleTimeout     time.Duration
	StreamCleanupInterval time.Duration

	// Circuit breaker settings
	CircuitBreakerFailureThreshold int
	CircuitBreakerResetTimeout     time.Duration
	CircuitBreakerSuccessThreshold int

	// Retry settings
	RetryMaxAttempts     int
	RetryInitialInterval time.Duration
	RetryMaxInterval     time.Duration

	// Timeout settings
	DHTSyncTimeout              time.Duration
	TUNSetupTimeout             time.Duration
	PeerConnectionCheckInterval time.Duration
	ShutdownGracePeriod         time.Duration
}

// New creates a new VPNConfig with values from the provided config
func New(cfg *config.C) *VPNConfig {
	// Get unallowed ports
	unallowedPortList := cfg.GetStringSlice("vpn.unallowed_ports", []string{})
	unallowedPorts := make(map[string]bool, len(unallowedPortList))
	for _, port := range unallowedPortList {
		unallowedPorts[port] = true
	}

	return &VPNConfig{
		// Basic settings
		Enable:    cfg.GetBool("vpn.enable", false),
		MTU:       cfg.GetInt("vpn.mtu", 1400),
		VirtualIP: cfg.GetString("vpn.virtual_ip", ""),
		Subnet:    cfg.GetInt("vpn.subnet", 8),
		Routes:    cfg.GetStringSlice("vpn.routes", []string{"10.0.0.0/8"}),
		Protocol:  cfg.GetString("vpn.protocol", "/vpn/1.0.0"),
		Routines:  cfg.GetInt("vpn.routines", 1), // Default to 1 routine

		// Security settings
		UnallowedPorts: unallowedPorts,

		// Stream pool settings
		StreamIdleTimeout:     time.Duration(cfg.GetInt("vpn.stream_idle_timeout", 10)) * time.Second,
		StreamCleanupInterval: time.Duration(cfg.GetInt("vpn.cleanup_interval", 20)) * time.Second,

		// Circuit breaker settings
		CircuitBreakerFailureThreshold: cfg.GetInt("vpn.circuit_breaker_failure_threshold", 5),                           // 5 failures default
		CircuitBreakerResetTimeout:     time.Duration(cfg.GetInt("vpn.circuit_breaker_reset_timeout", 60)) * time.Second, // 1 minute default
		CircuitBreakerSuccessThreshold: cfg.GetInt("vpn.circuit_breaker_success_threshold", 2),                           // 2 successes default

		// Retry settings
		RetryMaxAttempts:     cfg.GetInt("vpn.retry_max_attempts", 5),
		RetryInitialInterval: time.Duration(cfg.GetInt("vpn.retry_initial_interval", 1)) * time.Second,
		RetryMaxInterval:     time.Duration(cfg.GetInt("vpn.retry_max_interval", 30)) * time.Second,

		// Timeout settings
		DHTSyncTimeout:              time.Duration(cfg.GetInt("vpn.dht_sync_timeout", 60)) * time.Second,                     // 60 seconds default
		TUNSetupTimeout:             time.Duration(cfg.GetInt("vpn.tun_setup_timeout", 30)) * time.Second,                    // 30 seconds default
		PeerConnectionCheckInterval: time.Duration(cfg.GetInt("vpn.peer_connection_check_interval", 100)) * time.Millisecond, // 100 milliseconds default
		ShutdownGracePeriod:         time.Duration(cfg.GetInt("vpn.shutdown_grace_period", 50)) * time.Millisecond,           // 50 milliseconds default
	}
}

// GetStreamIdleTimeout returns the stream idle timeout
func (c *VPNConfig) GetStreamIdleTimeout() time.Duration {
	return c.StreamIdleTimeout
}

// GetCleanupInterval returns the cleanup interval
func (c *VPNConfig) GetStreamCleanupInterval() time.Duration {
	return c.StreamCleanupInterval
}

// GetCircuitBreakerFailureThreshold returns the circuit breaker failure threshold
func (c *VPNConfig) GetCircuitBreakerFailureThreshold() int {
	return c.CircuitBreakerFailureThreshold
}

// GetCircuitBreakerResetTimeout returns the circuit breaker reset timeout
func (c *VPNConfig) GetCircuitBreakerResetTimeout() time.Duration {
	return c.CircuitBreakerResetTimeout
}

// GetCircuitBreakerSuccessThreshold returns the circuit breaker success threshold
func (c *VPNConfig) GetCircuitBreakerSuccessThreshold() int {
	return c.CircuitBreakerSuccessThreshold
}

// GetRoutines returns the number of reader routines
func (c *VPNConfig) GetRoutines() int {
	return c.Routines
}
