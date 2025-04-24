package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createValidConfig() *VPNConfig {
	return &VPNConfig{
		// Basic settings
		Enable:    true,
		VirtualIP: "10.0.0.1",
		Subnet:    24,
		Routes:    []string{"10.0.0.0/24"},
		MTU:       1500,

		// Worker settings
		WorkerIdleTimeout:     300,
		WorkerBufferSize:      100,
		MaxWorkersPerPeer:     1000,
		WorkerCleanupInterval: 60 * time.Second,

		// Stream pool settings
		StreamIdleTimeout: 300 * time.Second,
		CleanupInterval:   60 * time.Second,

		// Circuit breaker settings
		CircuitBreakerFailureThreshold: 5,
		CircuitBreakerResetTimeout:     60 * time.Second,
		CircuitBreakerSuccessThreshold: 2,

		// Retry settings
		RetryMaxAttempts:     5,
		RetryInitialInterval: 1 * time.Second,
		RetryMaxInterval:     30 * time.Second,
	}
}

func TestValidate(t *testing.T) {
	// Test cases for the main Validate function
	tests := []struct {
		name    string
		modify  func(*VPNConfig)
		wantErr error
	}{
		{
			name:    "valid config",
			modify:  func(cfg *VPNConfig) {},
			wantErr: nil,
		},
		{
			name:    "disabled config",
			modify:  func(cfg *VPNConfig) { cfg.Enable = false },
			wantErr: nil,
		},
		// Basic settings errors
		{
			name:    "missing virtual IP",
			modify:  func(cfg *VPNConfig) { cfg.VirtualIP = "" },
			wantErr: ErrVirtualIPNotSet,
		},
		{
			name:    "invalid virtual IP",
			modify:  func(cfg *VPNConfig) { cfg.VirtualIP = "invalid" },
			wantErr: ErrInvalidVirtualIP,
		},
		{
			name:    "invalid subnet",
			modify:  func(cfg *VPNConfig) { cfg.Subnet = 33 },
			wantErr: ErrInvalidVirtualIP, // The virtual IP validation fails first
		},
		{
			name:    "missing routes",
			modify:  func(cfg *VPNConfig) { cfg.Routes = []string{} },
			wantErr: ErrRoutesNotSet,
		},
		{
			name:    "invalid routes",
			modify:  func(cfg *VPNConfig) { cfg.Routes = []string{"invalid"} },
			wantErr: ErrInvalidRoutes,
		},
		{
			name:    "MTU too small",
			modify:  func(cfg *VPNConfig) { cfg.MTU = 500 },
			wantErr: ErrInvalidMTU,
		},
		// Worker settings errors
		{
			name:    "invalid worker idle timeout",
			modify:  func(cfg *VPNConfig) { cfg.WorkerIdleTimeout = 0 },
			wantErr: ErrInvalidWorkerIdleTimeout,
		},
		// Stream pool settings errors
		// Circuit breaker settings errors
		{
			name:    "invalid circuit breaker failure threshold",
			modify:  func(cfg *VPNConfig) { cfg.CircuitBreakerFailureThreshold = 0 },
			wantErr: ErrInvalidCircuitBreakerFailureThreshold,
		},
		// Retry settings errors
		{
			name:    "invalid retry max attempts",
			modify:  func(cfg *VPNConfig) { cfg.RetryMaxAttempts = 0 },
			wantErr: ErrInvalidRetryMaxAttempts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh config for each test
			var cfg *VPNConfig
			if tt.name == "disabled config" {
				// Special case for disabled config
				cfg = &VPNConfig{}
			} else {
				cfg = createValidConfig()
			}
			// Apply the modification
			tt.modify(cfg)
			// Run the validation
			err := cfg.Validate()
			// Check the result
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAllSettings(t *testing.T) {
	// Define test cases for all validation functions
	testCases := map[string][]struct {
		name    string
		modify  func(*VPNConfig)
		wantErr error
	}{
		"validateBasicSettings": {
			{
				name:    "valid basic settings",
				modify:  func(cfg *VPNConfig) {},
				wantErr: nil,
			},
			{
				name:    "missing virtual IP",
				modify:  func(cfg *VPNConfig) { cfg.VirtualIP = "" },
				wantErr: ErrVirtualIPNotSet,
			},
			{
				name:    "invalid virtual IP",
				modify:  func(cfg *VPNConfig) { cfg.VirtualIP = "invalid" },
				wantErr: ErrInvalidVirtualIP,
			},
			{
				name:    "invalid subnet (too small)",
				modify:  func(cfg *VPNConfig) { cfg.Subnet = -1 },
				wantErr: ErrInvalidVirtualIP, // The virtual IP validation fails first
			},
			{
				name:    "invalid subnet (too large)",
				modify:  func(cfg *VPNConfig) { cfg.Subnet = 33 },
				wantErr: ErrInvalidVirtualIP, // The virtual IP validation fails first
			},
			{
				name:    "missing routes",
				modify:  func(cfg *VPNConfig) { cfg.Routes = []string{} },
				wantErr: ErrRoutesNotSet,
			},
			{
				name:    "invalid routes",
				modify:  func(cfg *VPNConfig) { cfg.Routes = []string{"invalid"} },
				wantErr: ErrInvalidRoutes,
			},
			{
				name:    "MTU too small",
				modify:  func(cfg *VPNConfig) { cfg.MTU = 500 },
				wantErr: ErrInvalidMTU,
			},
			{
				name:    "MTU too large",
				modify:  func(cfg *VPNConfig) { cfg.MTU = 10000 },
				wantErr: ErrInvalidMTU,
			},
		},
		"validateWorkerSettings": {
			{
				name:    "valid worker settings",
				modify:  func(cfg *VPNConfig) {},
				wantErr: nil,
			},
			{
				name:    "invalid worker idle timeout",
				modify:  func(cfg *VPNConfig) { cfg.WorkerIdleTimeout = 0 },
				wantErr: ErrInvalidWorkerIdleTimeout,
			},
			{
				name:    "invalid worker buffer size",
				modify:  func(cfg *VPNConfig) { cfg.WorkerBufferSize = 0 },
				wantErr: ErrInvalidWorkerBufferSize,
			},
			{
				name:    "invalid max workers per peer",
				modify:  func(cfg *VPNConfig) { cfg.MaxWorkersPerPeer = 0 },
				wantErr: ErrInvalidMaxWorkersPerPeer,
			},
			{
				name:    "invalid worker cleanup interval",
				modify:  func(cfg *VPNConfig) { cfg.WorkerCleanupInterval = 0 },
				wantErr: ErrInvalidWorkerCleanupInterval,
			},
		},
		"validateStreamPoolSettings": {
			{
				name:    "valid stream pool settings",
				modify:  func(cfg *VPNConfig) {},
				wantErr: nil,
			},

			{
				name:    "invalid stream idle timeout",
				modify:  func(cfg *VPNConfig) { cfg.StreamIdleTimeout = 0 },
				wantErr: ErrInvalidStreamIdleTimeout,
			},
			{
				name:    "invalid cleanup interval",
				modify:  func(cfg *VPNConfig) { cfg.CleanupInterval = 0 },
				wantErr: ErrInvalidCleanupInterval,
			},
		},
		"validateCircuitBreakerSettings": {
			{
				name:    "valid circuit breaker settings",
				modify:  func(cfg *VPNConfig) {},
				wantErr: nil,
			},
			{
				name:    "invalid circuit breaker failure threshold",
				modify:  func(cfg *VPNConfig) { cfg.CircuitBreakerFailureThreshold = 0 },
				wantErr: ErrInvalidCircuitBreakerFailureThreshold,
			},
			{
				name:    "invalid circuit breaker reset timeout",
				modify:  func(cfg *VPNConfig) { cfg.CircuitBreakerResetTimeout = 0 },
				wantErr: ErrInvalidCircuitBreakerResetTimeout,
			},
			{
				name:    "invalid circuit breaker success threshold",
				modify:  func(cfg *VPNConfig) { cfg.CircuitBreakerSuccessThreshold = 0 },
				wantErr: ErrInvalidCircuitBreakerSuccessThreshold,
			},
		},
		"validateRetrySettings": {
			{
				name:    "valid retry settings",
				modify:  func(cfg *VPNConfig) {},
				wantErr: nil,
			},
			{
				name:    "invalid retry max attempts",
				modify:  func(cfg *VPNConfig) { cfg.RetryMaxAttempts = 0 },
				wantErr: ErrInvalidRetryMaxAttempts,
			},
			{
				name:    "invalid retry initial interval",
				modify:  func(cfg *VPNConfig) { cfg.RetryInitialInterval = 0 },
				wantErr: ErrInvalidRetryInitialInterval,
			},
			{
				name:    "invalid retry max interval",
				modify:  func(cfg *VPNConfig) { cfg.RetryMaxInterval = 0 },
				wantErr: ErrInvalidRetryMaxInterval,
			},
		},
	}

	// Map of validation functions to test
	validationFuncs := map[string]func(*VPNConfig) error{
		"validateBasicSettings":          (*VPNConfig).validateBasicSettings,
		"validateWorkerSettings":         (*VPNConfig).validateWorkerSettings,
		"validateStreamPoolSettings":     (*VPNConfig).validateStreamPoolSettings,
		"validateCircuitBreakerSettings": (*VPNConfig).validateCircuitBreakerSettings,
		"validateRetrySettings":          (*VPNConfig).validateRetrySettings,
	}

	// Run tests for each validation function
	for funcName, tests := range testCases {
		t.Run(funcName, func(t *testing.T) {
			validateFunc := validationFuncs[funcName]
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					// Create a fresh config for each test
					cfg := createValidConfig()
					// Apply the modification
					tt.modify(cfg)
					// Run the validation function
					err := validateFunc(cfg)
					// Check the result
					if tt.wantErr != nil {
						assert.Equal(t, tt.wantErr, err)
					} else {
						assert.NoError(t, err)
					}
				})
			}
		})
	}
}
