package config

import (
	"errors"
	"fmt"
	"time"
)

// Additional error codes for enhanced validation
var (
	// Consistency errors
	ErrInconsistentBufferSettings  = errors.New("inconsistent buffer settings: buffer utilization threshold should be less than 100")
	ErrInconsistentTimeoutSettings = errors.New("inconsistent timeout settings: cleanup interval should be less than stream idle timeout")
	ErrInconsistentRetrySettings   = errors.New("inconsistent retry settings: max interval should be greater than initial interval")
	ErrInconsistentStreamSettings  = errors.New("inconsistent stream settings: max streams per peer should be greater than 0")

	// Performance warnings
	WarnHighBufferSize  = "high packet buffer size may consume excessive memory"
	WarnLowBufferSize   = "low packet buffer size may cause packet drops under load"
	WarnHighStreamCount = "high max streams per peer may consume excessive resources"
	WarnLowStreamCount  = "low max streams per peer may limit throughput"
)

// ValidationResult contains the result of a configuration validation
type ValidationResult struct {
	// Whether the configuration is valid
	Valid bool
	// Error if the configuration is invalid
	Error error
	// Warnings about the configuration
	Warnings []string
}

// ValidateEnhanced performs enhanced validation of the configuration
func (c *VPNConfig) ValidateEnhanced() ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Warnings: []string{},
	}

	// First, run the basic validation
	if err := c.Validate(); err != nil {
		result.Valid = false
		result.Error = err
		return result
	}

	// Check for consistency issues
	if err := c.validateConsistency(); err != nil {
		result.Valid = false
		result.Error = err
		return result
	}

	// Check for performance issues (these are warnings, not errors)
	c.checkPerformance(&result)

	return result
}

// validateConsistency checks for consistency issues in the configuration
func (c *VPNConfig) validateConsistency() error {
	// Check buffer utilization threshold
	if c.BufferUtilThreshold >= 100 {
		return fmt.Errorf("%w: buffer utilization threshold is %d",
			ErrInconsistentBufferSettings, c.BufferUtilThreshold)
	}

	// Check timeout settings
	if c.CleanupInterval >= c.StreamIdleTimeout {
		return fmt.Errorf("%w: cleanup interval (%s) should be less than stream idle timeout (%s)",
			ErrInconsistentTimeoutSettings, c.CleanupInterval, c.StreamIdleTimeout)
	}

	// Check retry settings
	if c.RetryMaxInterval <= c.RetryInitialInterval {
		return fmt.Errorf("%w: max interval (%s) should be greater than initial interval (%s)",
			ErrInconsistentRetrySettings, c.RetryMaxInterval, c.RetryInitialInterval)
	}

	// Check stream settings
	if c.MaxStreamsPerPeer <= 0 {
		return fmt.Errorf("%w: max streams per peer is %d",
			ErrInconsistentStreamSettings, c.MaxStreamsPerPeer)
	}

	return nil
}

// checkPerformance checks for potential performance issues in the configuration
func (c *VPNConfig) checkPerformance(result *ValidationResult) {
	// Check packet buffer size
	if c.PacketBufferSize > 10000 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("%s: %d", WarnHighBufferSize, c.PacketBufferSize))
	} else if c.PacketBufferSize < 100 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("%s: %d", WarnLowBufferSize, c.PacketBufferSize))
	}

	// Check max streams per peer
	if c.MaxStreamsPerPeer > 20 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("%s: %d", WarnHighStreamCount, c.MaxStreamsPerPeer))
	} else if c.MaxStreamsPerPeer < 3 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("%s: %d", WarnLowStreamCount, c.MaxStreamsPerPeer))
	}

	// Check stream idle timeout
	if c.StreamIdleTimeout < 60*time.Second {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("short stream idle timeout (%s) may cause frequent stream recreation",
				c.StreamIdleTimeout))
	}

	// Check buffer utilization and usage count thresholds
	if c.BufferUtilThreshold < 30 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("low buffer utilization threshold (%d%%) may cause premature stream creation",
				c.BufferUtilThreshold))
	}

	if c.UsageCountThreshold < 5 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("low usage count threshold (%d) may cause premature stream creation",
				c.UsageCountThreshold))
	}
}

// StandardizeConfig ensures that all configuration values are within reasonable ranges
// and sets default values for missing or invalid settings
func StandardizeConfig(cfg *VPNConfig) *VPNConfig {
	// Create a copy of the configuration
	standardized := *cfg

	// Standardize basic settings
	if standardized.MTU < 576 {
		standardized.MTU = 1400 // Default MTU
	} else if standardized.MTU > 9000 {
		standardized.MTU = 9000 // Max reasonable MTU
	}

	// Standardize stream pool settings
	if standardized.MaxStreamsPerPeer <= 0 {
		standardized.MaxStreamsPerPeer = 5 // Default max streams per peer
	} else if standardized.MaxStreamsPerPeer > 50 {
		standardized.MaxStreamsPerPeer = 50 // Max reasonable streams per peer
	}

	if standardized.StreamIdleTimeout <= 0 {
		standardized.StreamIdleTimeout = 5 * time.Minute // Default stream idle timeout
	}

	if standardized.CleanupInterval <= 0 {
		standardized.CleanupInterval = 1 * time.Minute // Default cleanup interval
	}

	if standardized.PacketBufferSize <= 0 {
		standardized.PacketBufferSize = 1000 // Default packet buffer size
	}

	// Standardize load balancing settings
	if standardized.BufferUtilThreshold <= 0 || standardized.BufferUtilThreshold >= 100 {
		standardized.BufferUtilThreshold = 30 // Default buffer utilization threshold
	}

	if standardized.UsageCountThreshold <= 0 {
		standardized.UsageCountThreshold = 10 // Default usage count threshold
	}

	// Ensure weights sum to 1.0
	totalWeight := standardized.UsageCountWeight + standardized.BufferUtilWeight
	if totalWeight <= 0 {
		// If both weights are zero or negative, set default weights
		standardized.UsageCountWeight = 0.5
		standardized.BufferUtilWeight = 0.5
	} else {
		// Normalize weights to sum to 1.0
		standardized.UsageCountWeight = standardized.UsageCountWeight / totalWeight
		standardized.BufferUtilWeight = standardized.BufferUtilWeight / totalWeight
	}

	// Standardize circuit breaker settings
	if standardized.CircuitBreakerFailureThreshold <= 0 {
		standardized.CircuitBreakerFailureThreshold = 5 // Default failure threshold
	}

	if standardized.CircuitBreakerResetTimeout <= 0 {
		standardized.CircuitBreakerResetTimeout = 30 * time.Second // Default reset timeout
	}

	if standardized.CircuitBreakerSuccessThreshold <= 0 {
		standardized.CircuitBreakerSuccessThreshold = 2 // Default success threshold
	}

	// Standardize retry settings
	if standardized.RetryMaxAttempts <= 0 {
		standardized.RetryMaxAttempts = 3 // Default max attempts
	}

	if standardized.RetryInitialInterval <= 0 {
		standardized.RetryInitialInterval = 100 * time.Millisecond // Default initial interval
	}

	if standardized.RetryMaxInterval <= 0 {
		standardized.RetryMaxInterval = 1 * time.Second // Default max interval
	}

	return &standardized
}
