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
	Routines  int

	// Security settings
	ConntrackCacheTimeout time.Duration // Timeout for the conntrack cache for each routine

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
	routines := cfg.GetInt("vpn.routines", 1) // Default to 1 routine

	// Get the conntrack cache
	conntrackCacheTimeout := cfg.GetDuration("firewall.conntrack.routine_cache_timeout", 0)
	if routines > 1 && !cfg.IsSet("firewall.conntrack.routine_cache_timeout") {
		// Use a different default if we are running with multiple routines
		conntrackCacheTimeout = 1 * time.Second
	}
	if conntrackCacheTimeout > 0 {
		log.WithField("duration", conntrackCacheTimeout).Info("Using routine-local conntrack cache")
	}

	return &VPNConfig{
		// Basic settings
		Enable:    cfg.GetBool("vpn.enable", false),
		MTU:       cfg.GetInt("vpn.mtu", 1400),
		VirtualIP: cfg.GetString("vpn.virtual_ip", ""),
		Subnet:    cfg.GetInt("vpn.subnet", 8),
		Routes:    cfg.GetStringSlice("vpn.routes", []string{"10.0.0.0/8"}),
		Protocol:  cfg.GetString("vpn.protocol", "/vpn/1.0.0"),
		Routines:  routines,

		// Security settings
		ConntrackCacheTimeout: conntrackCacheTimeout,

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
