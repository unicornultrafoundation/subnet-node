package config

import (
	"errors"
	"net"
	"strconv"
)

// Error codes
var (
	// Basic settings errors
	ErrVirtualIPNotSet  = errors.New("virtual IP is not set")
	ErrRoutesNotSet     = errors.New("routes are not set")
	ErrInvalidVirtualIP = errors.New("virtual IP is invalid")
	ErrInvalidRoutes    = errors.New("routes are invalid")
	ErrInvalidMTU       = errors.New("invalid MTU (must be between 576 and 9000)")

	// Worker settings errors
	ErrInvalidWorkerIdleTimeout     = errors.New("invalid worker idle timeout (must be greater than 0)")
	ErrInvalidWorkerBufferSize      = errors.New("invalid worker buffer size (must be greater than 0)")
	ErrInvalidMaxWorkers            = errors.New("invalid max workers (must be greater than 0)")
	ErrInvalidWorkerCleanupInterval = errors.New("invalid worker cleanup interval (must be greater than 0)")

	// Stream pool settings errors
	ErrInvalidStreamIdleTimeout = errors.New("invalid stream idle timeout (must be greater than 0)")
	ErrInvalidCleanupInterval   = errors.New("invalid cleanup interval (must be greater than 0)")

	// Circuit breaker settings errors
	ErrInvalidCircuitBreakerFailureThreshold = errors.New("invalid circuit breaker failure threshold (must be greater than 0)")
	ErrInvalidCircuitBreakerResetTimeout     = errors.New("invalid circuit breaker reset timeout (must be greater than 0)")
	ErrInvalidCircuitBreakerSuccessThreshold = errors.New("invalid circuit breaker success threshold (must be greater than 0)")

	// Stream health settings errors
	ErrInvalidHealthCheckInterval    = errors.New("invalid health check interval (must be greater than 0)")
	ErrInvalidHealthCheckTimeout     = errors.New("invalid health check timeout (must be greater than 0)")
	ErrInvalidMaxConsecutiveFailures = errors.New("invalid max consecutive failures (must be greater than 0)")
	ErrInvalidWarmInterval           = errors.New("invalid warm interval (must be greater than 0)")

	// Retry settings errors
	ErrInvalidRetryMaxAttempts     = errors.New("invalid retry max attempts (must be greater than 0)")
	ErrInvalidRetryInitialInterval = errors.New("invalid retry initial interval (must be greater than 0)")
	ErrInvalidRetryMaxInterval     = errors.New("invalid retry max interval (must be greater than 0)")
)

// Validate checks if the configuration is valid
func (c *VPNConfig) Validate() error {
	// If VPN is disabled, no need to validate further
	if !c.Enable {
		return nil
	}

	// Validate basic settings
	if err := c.validateBasicSettings(); err != nil {
		return err
	}

	// Validate worker settings
	if err := c.validateWorkerSettings(); err != nil {
		return err
	}

	// Validate stream pool settings
	if err := c.validateStreamPoolSettings(); err != nil {
		return err
	}

	// Validate circuit breaker settings
	if err := c.validateCircuitBreakerSettings(); err != nil {
		return err
	}

	// Validate stream health settings
	if err := c.validateStreamHealthSettings(); err != nil {
		return err
	}

	// Validate retry settings
	if err := c.validateRetrySettings(); err != nil {
		return err
	}

	return nil
}

// validateBasicSettings validates the basic VPN settings
func (c *VPNConfig) validateBasicSettings() error {
	// Validate virtual IP
	if c.VirtualIP == "" {
		return ErrVirtualIPNotSet
	} else if _, _, err := net.ParseCIDR(c.VirtualIP + "/" + strconv.Itoa(c.Subnet)); err != nil {
		return ErrInvalidVirtualIP
	}

	// Validate subnet
	if c.Subnet < 0 || c.Subnet > 32 {
		return ErrInvalidVirtualIP
	}

	// Validate routes
	if len(c.Routes) == 0 {
		return ErrRoutesNotSet
	} else {
		for _, route := range c.Routes {
			if _, _, err := net.ParseCIDR(route); err != nil {
				return ErrInvalidRoutes
			}
		}
	}

	// Validate MTU
	if c.MTU < 576 || c.MTU > 9000 {
		return ErrInvalidMTU
	}

	return nil
}

// validateWorkerSettings validates the worker settings
func (c *VPNConfig) validateWorkerSettings() error {
	// Validate worker idle timeout
	if c.WorkerIdleTimeout <= 0 {
		return ErrInvalidWorkerIdleTimeout
	}

	// Validate worker buffer size
	if c.WorkerBufferSize <= 0 {
		return ErrInvalidWorkerBufferSize
	}

	// Validate max workers
	if c.MaxWorkers <= 0 {
		return ErrInvalidMaxWorkers
	}

	// Validate worker cleanup interval
	if c.WorkerCleanupInterval <= 0 {
		return ErrInvalidWorkerCleanupInterval
	}

	return nil
}

// validateStreamPoolSettings validates the stream pool settings
func (c *VPNConfig) validateStreamPoolSettings() error {

	// Validate stream idle timeout
	if c.StreamIdleTimeout <= 0 {
		return ErrInvalidStreamIdleTimeout
	}

	// Validate cleanup interval
	if c.CleanupInterval <= 0 {
		return ErrInvalidCleanupInterval
	}

	return nil
}

// validateCircuitBreakerSettings validates the circuit breaker settings
func (c *VPNConfig) validateCircuitBreakerSettings() error {
	// Validate circuit breaker failure threshold
	if c.CircuitBreakerFailureThreshold <= 0 {
		return ErrInvalidCircuitBreakerFailureThreshold
	}

	// Validate circuit breaker reset timeout
	if c.CircuitBreakerResetTimeout <= 0 {
		return ErrInvalidCircuitBreakerResetTimeout
	}

	// Validate circuit breaker success threshold
	if c.CircuitBreakerSuccessThreshold <= 0 {
		return ErrInvalidCircuitBreakerSuccessThreshold
	}

	return nil
}

// validateStreamHealthSettings validates the stream health settings
func (c *VPNConfig) validateStreamHealthSettings() error {
	// Validate health check interval
	if c.HealthCheckInterval <= 0 {
		return ErrInvalidHealthCheckInterval
	}

	// Validate health check timeout
	if c.HealthCheckTimeout <= 0 {
		return ErrInvalidHealthCheckTimeout
	}

	// Validate max consecutive failures
	if c.MaxConsecutiveFailures <= 0 {
		return ErrInvalidMaxConsecutiveFailures
	}

	// Validate warm interval
	if c.WarmInterval <= 0 {
		return ErrInvalidWarmInterval
	}

	return nil
}

// validateRetrySettings validates the retry settings
func (c *VPNConfig) validateRetrySettings() error {
	// Validate retry max attempts
	if c.RetryMaxAttempts <= 0 {
		return ErrInvalidRetryMaxAttempts
	}

	// Validate retry initial interval
	if c.RetryInitialInterval <= 0 {
		return ErrInvalidRetryInitialInterval
	}

	// Validate retry max interval
	if c.RetryMaxInterval <= 0 {
		return ErrInvalidRetryMaxInterval
	}

	return nil
}
