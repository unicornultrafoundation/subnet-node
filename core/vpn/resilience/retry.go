package resilience

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// RetryManager handles retry operations with exponential backoff
type RetryManager struct {
	// Maximum number of retry attempts
	maxAttempts int
	// Initial interval between retries
	initialInterval time.Duration
	// Maximum interval between retries
	maxInterval time.Duration
	// Metrics
	metrics struct {
		retryAttempts int64
		retrySuccess  int64
		retryFailure  int64
	}
}

// NewRetryManager creates a new retry manager
func NewRetryManager(maxAttempts int, initialInterval, maxInterval time.Duration) *RetryManager {
	return &RetryManager{
		maxAttempts:     maxAttempts,
		initialInterval: initialInterval,
		maxInterval:     maxInterval,
	}
}

// RetryOperation retries an operation with exponential backoff
func (r *RetryManager) RetryOperation(ctx context.Context, operation func() error) error {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = r.initialInterval
	bo.MaxInterval = r.maxInterval
	bo.MaxElapsedTime = 0 // No time limit, we'll use attempt count instead

	var err error
	var attempt int

	// Create a backoff operation that counts attempts
	backoffOperation := func() error {
		// Check if context is canceled
		if ctx.Err() != nil {
			return backoff.Permanent(ctx.Err())
		}

		attempt++
		if attempt > r.maxAttempts {
			return backoff.Permanent(err) // Stop retrying after max attempts
		}
		
		err = operation()
		if err != nil {
			return err // Will be retried
		}
		
		return nil // Success, stop retrying
	}

	// Run the retry operation
	retryErr := backoff.Retry(backoffOperation, backoff.WithContext(bo, ctx))
	
	if retryErr != nil {
		r.metrics.retryFailure++
		// If context was canceled, return that error
		if ctx.Err() != nil {
			return fmt.Errorf("context canceled: %w", ctx.Err())
		}
		return err // Return the original error from the operation
	} else {
		r.metrics.retrySuccess++
		return nil
	}
}

// GetMetrics returns the current metrics
func (r *RetryManager) GetMetrics() map[string]int64 {
	return map[string]int64{
		"retry_attempts": r.metrics.retryAttempts,
		"retry_success":  r.metrics.retrySuccess,
		"retry_failure":  r.metrics.retryFailure,
	}
}
