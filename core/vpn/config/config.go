package config

import (
	"time"

	"github.com/unicornultrafoundation/subnet-node/config"
)

// VPNConfig encapsulates all configuration settings for the VPN service
type VPNConfig struct {
	// Basic settings
	Enable    bool
	MTU       int
	VirtualIP string
	Subnet    int
	Routes    []string

	// Security settings
	UnallowedPorts map[string]bool

	// Worker settings
	WorkerIdleTimeout     int
	WorkerBufferSize      int
	MaxWorkers            int
	WorkerCleanupInterval time.Duration

	// Stream pool settings
	MaxStreamsPerPeer int
	MinStreamsPerPeer int
	StreamIdleTimeout time.Duration
	CleanupInterval   time.Duration

	// Circuit breaker settings
	CircuitBreakerFailureThreshold int
	CircuitBreakerResetTimeout     time.Duration
	CircuitBreakerSuccessThreshold int

	// Stream health settings
	HealthCheckInterval    time.Duration
	HealthCheckTimeout     time.Duration
	MaxConsecutiveFailures int
	WarmInterval           time.Duration

	// Stream multiplexing settings
	MultiplexingEnabled      bool
	MaxStreamsPerMultiplexer int
	MinStreamsPerMultiplexer int
	AutoScalingInterval      time.Duration

	// Retry settings
	RetryMaxAttempts     int
	RetryInitialInterval time.Duration
	RetryMaxInterval     time.Duration
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

		// Security settings
		UnallowedPorts: unallowedPorts,

		// Worker settings
		WorkerIdleTimeout:     cfg.GetInt("vpn.worker_idle_timeout", 300),                                 // 5 minutes default
		WorkerBufferSize:      cfg.GetInt("vpn.worker_buffer_size", 100),                                  // Default buffer size
		MaxWorkers:            cfg.GetInt("vpn.max_workers", 1000),                                        // Maximum number of workers
		WorkerCleanupInterval: time.Duration(cfg.GetInt("vpn.worker_cleanup_interval", 60)) * time.Second, // 1 minute default

		// Stream pool settings
		MaxStreamsPerPeer: cfg.GetInt("vpn.max_streams_per_peer", 10),                              // 10 streams per peer default
		MinStreamsPerPeer: cfg.GetInt("vpn.min_streams_per_peer", 3),                               // 3 streams per peer default
		StreamIdleTimeout: time.Duration(cfg.GetInt("vpn.stream_idle_timeout", 300)) * time.Second, // 5 minutes default
		CleanupInterval:   time.Duration(cfg.GetInt("vpn.cleanup_interval", 60)) * time.Second,     // 1 minute default

		// Circuit breaker settings
		CircuitBreakerFailureThreshold: cfg.GetInt("vpn.circuit_breaker_failure_threshold", 5),                           // 5 failures default
		CircuitBreakerResetTimeout:     time.Duration(cfg.GetInt("vpn.circuit_breaker_reset_timeout", 60)) * time.Second, // 1 minute default
		CircuitBreakerSuccessThreshold: cfg.GetInt("vpn.circuit_breaker_success_threshold", 2),                           // 2 successes default

		// Stream health settings
		HealthCheckInterval:    time.Duration(cfg.GetInt("vpn.health_check_interval", 30)) * time.Second, // 30 seconds default
		HealthCheckTimeout:     time.Duration(cfg.GetInt("vpn.health_check_timeout", 5)) * time.Second,   // 5 seconds default
		MaxConsecutiveFailures: cfg.GetInt("vpn.max_consecutive_failures", 3),                            // 3 failures default
		WarmInterval:           time.Duration(cfg.GetInt("vpn.warm_interval", 60)) * time.Second,         // 1 minute default

		// Stream multiplexing settings
		MultiplexingEnabled:      cfg.GetBool("vpn.multiplexing_enabled", true),                            // Enabled by default
		MaxStreamsPerMultiplexer: cfg.GetInt("vpn.max_streams_per_multiplexer", 5),                         // 5 streams per multiplexer default
		MinStreamsPerMultiplexer: cfg.GetInt("vpn.min_streams_per_multiplexer", 2),                         // 2 streams per multiplexer default
		AutoScalingInterval:      time.Duration(cfg.GetInt("vpn.auto_scaling_interval", 30)) * time.Second, // 30 seconds default

		// Retry settings
		RetryMaxAttempts:     cfg.GetInt("vpn.retry_max_attempts", 5),
		RetryInitialInterval: time.Duration(cfg.GetInt("vpn.retry_initial_interval", 1)) * time.Second,
		RetryMaxInterval:     time.Duration(cfg.GetInt("vpn.retry_max_interval", 30)) * time.Second,
	}
}

// GetWorkerBufferSize returns the worker buffer size
func (c *VPNConfig) GetWorkerBufferSize() int {
	return c.WorkerBufferSize
}

// GetWorkerIdleTimeout returns the worker idle timeout in seconds
func (c *VPNConfig) GetWorkerIdleTimeout() int {
	return c.WorkerIdleTimeout
}

// GetWorkerCleanupInterval returns the interval for worker cleanup
func (c *VPNConfig) GetWorkerCleanupInterval() time.Duration {
	return c.WorkerCleanupInterval
}

// GetMaxStreamsPerPeer returns the maximum number of streams per peer
func (c *VPNConfig) GetMaxStreamsPerPeer() int {
	return c.MaxStreamsPerPeer
}

// GetMinStreamsPerPeer returns the minimum number of streams per peer
func (c *VPNConfig) GetMinStreamsPerPeer() int {
	return c.MinStreamsPerPeer
}

// GetStreamIdleTimeout returns the stream idle timeout
func (c *VPNConfig) GetStreamIdleTimeout() time.Duration {
	return c.StreamIdleTimeout
}

// GetCleanupInterval returns the cleanup interval
func (c *VPNConfig) GetCleanupInterval() time.Duration {
	return c.CleanupInterval
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

// GetHealthCheckInterval returns the health check interval
func (c *VPNConfig) GetHealthCheckInterval() time.Duration {
	return c.HealthCheckInterval
}

// GetHealthCheckTimeout returns the health check timeout
func (c *VPNConfig) GetHealthCheckTimeout() time.Duration {
	return c.HealthCheckTimeout
}

// GetMaxConsecutiveFailures returns the maximum consecutive failures
func (c *VPNConfig) GetMaxConsecutiveFailures() int {
	return c.MaxConsecutiveFailures
}

// GetWarmInterval returns the warm interval
func (c *VPNConfig) GetWarmInterval() time.Duration {
	return c.WarmInterval
}

// GetMaxStreamsPerMultiplexer returns the maximum number of streams per multiplexer
func (c *VPNConfig) GetMaxStreamsPerMultiplexer() int {
	return c.MaxStreamsPerMultiplexer
}

// GetMinStreamsPerMultiplexer returns the minimum number of streams per multiplexer
func (c *VPNConfig) GetMinStreamsPerMultiplexer() int {
	return c.MinStreamsPerMultiplexer
}

// GetAutoScalingInterval returns the auto-scaling interval
func (c *VPNConfig) GetAutoScalingInterval() time.Duration {
	return c.AutoScalingInterval
}

// GetMultiplexingEnabled returns whether multiplexing is enabled
func (c *VPNConfig) GetMultiplexingEnabled() bool {
	return c.MultiplexingEnabled
}
