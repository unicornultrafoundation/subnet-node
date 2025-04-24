package resilience

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
)

// Logger for the resilience package
var log = logrus.WithField("service", "vpn-resilience")

// ResilienceConfig contains configuration for resilience patterns
type ResilienceConfig struct {
	// Circuit breaker configuration
	CircuitBreakerFailureThreshold int
	CircuitBreakerResetTimeout     time.Duration
	CircuitBreakerSuccessThreshold int

	// Retry configuration
	RetryMaxAttempts     int
	RetryInitialInterval time.Duration
	RetryMaxInterval     time.Duration
}

// DefaultResilienceConfig returns a default configuration for resilience patterns
func DefaultResilienceConfig() *ResilienceConfig {
	return &ResilienceConfig{
		// Circuit breaker defaults
		CircuitBreakerFailureThreshold: 5,
		CircuitBreakerResetTimeout:     30 * time.Second,
		CircuitBreakerSuccessThreshold: 2,

		// Retry defaults
		RetryMaxAttempts:     3,
		RetryInitialInterval: 100 * time.Millisecond,
		RetryMaxInterval:     2 * time.Second,
	}
}

// ResilienceService provides centralized access to resilience patterns
type ResilienceService struct {
	circuitBreakerMgr *CircuitBreakerManager
	retryManager      *RetryManager
	config            *ResilienceConfig
}

// NewResilienceService creates a new resilience service with the provided configuration
func NewResilienceService(config *ResilienceConfig) *ResilienceService {
	if config == nil {
		config = DefaultResilienceConfig()
	}

	return &ResilienceService{
		circuitBreakerMgr: NewCircuitBreakerManager(
			config.CircuitBreakerFailureThreshold,
			config.CircuitBreakerResetTimeout,
			config.CircuitBreakerSuccessThreshold,
		),
		retryManager: NewRetryManager(
			config.RetryMaxAttempts,
			config.RetryInitialInterval,
			config.RetryMaxInterval,
		),
		config: config,
	}
}

// GetCircuitBreakerManager returns the circuit breaker manager
func (s *ResilienceService) GetCircuitBreakerManager() *CircuitBreakerManager {
	return s.circuitBreakerMgr
}

// GetRetryManager returns the retry manager
func (s *ResilienceService) GetRetryManager() *RetryManager {
	return s.retryManager
}

// ExecuteWithResilience executes an operation with circuit breaker and retry protection
func (s *ResilienceService) ExecuteWithResilience(
	ctx context.Context,
	breakerId string,
	operation func() error,
) (error, int) {
	logger := log.WithFields(logrus.Fields{
		"breaker_id": breakerId,
		"operation":  "resilience_pattern",
	})

	// Check if the circuit is already open
	if s.circuitBreakerMgr.GetBreaker(breakerId).GetState() == StateOpen {
		logger.Warn("Circuit breaker is open, fast-failing operation")
		return fmt.Errorf("circuit breaker open for %s, too many failures", breakerId), 0
	}

	logger.Debug("Executing operation with circuit breaker and retry protection")

	// Track retry attempts
	var retryAttempts int

	// Execute with circuit breaker, which will internally track failures
	err := s.circuitBreakerMgr.Execute(breakerId, func() error {
		// Use retry operation within the circuit breaker
		retryErr, attempts := s.retryManager.RetryOperationWithAttempts(ctx, operation)
		retryAttempts = attempts
		return retryErr
	})

	if err != nil {
		logger.WithError(err).WithField("retry_attempts", retryAttempts).Warn("Operation failed after retries and circuit breaker protection")
	} else {
		logger.WithField("retry_attempts", retryAttempts).Debug("Operation completed successfully")
	}

	return err, retryAttempts
}

// ExecuteWithMetrics executes an operation with metrics collection
func (s *ResilienceService) ExecuteWithMetrics(
	ctx context.Context,
	operation string,
	component string,
	target string,
	fn func() error,
) (error, int) {
	startTime := time.Now()
	breakerId := fmt.Sprintf("%s-%s-%s", component, operation, target)

	err, attempts := s.ExecuteWithResilience(ctx, breakerId, fn)

	duration := time.Since(startTime)

	// Log the operation
	logger := log.WithFields(logrus.Fields{
		"operation":  operation,
		"component":  component,
		"target":     target,
		"duration":   duration,
		"attempts":   attempts,
		"successful": err == nil,
		"breaker_id": breakerId,
	})

	if err != nil {
		logger.WithError(err).Warn("Operation failed")
	} else if attempts > 1 {
		logger.Debug("Operation succeeded after retries")
	} else {
		logger.Debug("Operation succeeded")
	}

	return err, attempts
}

// FormatPeerBreakerId formats a breaker ID for a peer operation
func (s *ResilienceService) FormatPeerBreakerId(peerID peer.ID, operation string) string {
	return fmt.Sprintf("peer-%s-%s", peerID.String(), operation)
}

// FormatWorkerBreakerId formats a breaker ID for a worker operation
func (s *ResilienceService) FormatWorkerBreakerId(syncKey string, operation string) string {
	return fmt.Sprintf("worker-%s-%s", syncKey, operation)
}
