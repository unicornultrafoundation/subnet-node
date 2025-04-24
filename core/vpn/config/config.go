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
	Protocol  string

	// Security settings
	UnallowedPorts map[string]bool

	// Worker settings
	WorkerIdleTimeout     int
	WorkerBufferSize      int
	MaxWorkersPerPeer     int
	WorkerCleanupInterval time.Duration

	// Stream pool settings
	MaxStreamsPerPeer int
	StreamIdleTimeout time.Duration
	CleanupInterval   time.Duration

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

		// Security settings
		UnallowedPorts: unallowedPorts,

		// Worker settings
		WorkerIdleTimeout:     cfg.GetInt("vpn.worker_idle_timeout", 5),                                  // 5 seconds default
		WorkerBufferSize:      cfg.GetInt("vpn.worker_buffer_size", 200),                                 // 200 packets buffer size
		MaxWorkersPerPeer:     cfg.GetInt("vpn.max_workers_per_peer", 20),                                // 20 workers per peer default
		WorkerCleanupInterval: time.Duration(cfg.GetInt("vpn.worker_cleanup_interval", 5)) * time.Second, // 5 seconds default

		// Stream pool settings
		MaxStreamsPerPeer: cfg.GetInt("vpn.max_streams_per_peer", 30),                            // 30 streams per peer default
		StreamIdleTimeout: time.Duration(cfg.GetInt("vpn.stream_idle_timeout", 5)) * time.Second, // 5 seconds default
		CleanupInterval:   time.Duration(cfg.GetInt("vpn.cleanup_interval", 5)) * time.Second,    // 5 seconds default

		// Circuit breaker settings
		CircuitBreakerFailureThreshold: cfg.GetInt("vpn.circuit_breaker_failure_threshold", 5),                           // 5 failures default
		CircuitBreakerResetTimeout:     time.Duration(cfg.GetInt("vpn.circuit_breaker_reset_timeout", 60)) * time.Second, // 1 minute default
		CircuitBreakerSuccessThreshold: cfg.GetInt("vpn.circuit_breaker_success_threshold", 2),                           // 2 successes default

		// Retry settingsies
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

// GetMaxWorkersPerPeer returns the maximum number of workers per peer
func (c *VPNConfig) GetMaxWorkersPerPeer() int {
	return c.MaxWorkersPerPeer
}
