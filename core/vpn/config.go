package vpn

import (
	"net"
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
	WorkerIdleTimeout int
	WorkerBufferSize  int
	MaxWorkers        int

	// Retry settings
	RetryMaxAttempts     int
	RetryInitialInterval time.Duration
	RetryMaxInterval     time.Duration
}

// NewVPNConfig creates a new VPNConfig with values from the provided config
func NewVPNConfig(cfg *config.C) *VPNConfig {
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
		WorkerIdleTimeout: cfg.GetInt("vpn.worker_idle_timeout", 300), // 5 minutes default
		WorkerBufferSize:  cfg.GetInt("vpn.worker_buffer_size", 100),  // Default buffer size
		MaxWorkers:        cfg.GetInt("vpn.max_workers", 1000),        // Maximum number of workers

		// Retry settings
		RetryMaxAttempts:     cfg.GetInt("vpn.retry_max_attempts", 5),
		RetryInitialInterval: time.Duration(cfg.GetInt("vpn.retry_initial_interval", 1)) * time.Second,
		RetryMaxInterval:     time.Duration(cfg.GetInt("vpn.retry_max_interval", 30)) * time.Second,
	}
}

// Validate checks if the configuration is valid
func (c *VPNConfig) Validate() error {
	if !c.Enable {
		return nil
	}

	if c.VirtualIP == "" {
		return ErrVirtualIPNotSet
	} else if _, _, err := net.ParseCIDR(c.VirtualIP); err != nil {
		return ErrInvalidVirtualIP
	}

	if len(c.Routes) == 0 {
		return ErrRoutesNotSet
	} else {
		for _, route := range c.Routes {
			if _, _, err := net.ParseCIDR(route); err != nil {
				return ErrInvalidRoutes
			}
		}
	}

	if c.MTU < 576 || c.MTU > 9000 {
		return ErrInvalidMTU
	}

	return nil
}
