package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryManagerWithMaxAttempts(t *testing.T) {
	// Create a retry manager with 3 max attempts
	retryManager := NewRetryManager(3, 10*time.Millisecond, 100*time.Millisecond)

	// Create a function that fails the first 2 times and succeeds on the 3rd attempt
	callCount := 0
	err := retryManager.RetryOperation(context.Background(), func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
}

func TestRetryManagerWithMaxAttemptsExceeded(t *testing.T) {
	// Create a retry manager with 3 max attempts
	retryManager := NewRetryManager(3, 10*time.Millisecond, 100*time.Millisecond)

	// Create a function that always fails
	callCount := 0
	err := retryManager.RetryOperation(context.Background(), func() error {
		callCount++
		return errors.New("persistent error")
	})

	assert.Error(t, err)
	assert.Equal(t, "persistent error", err.Error())
	// The function should be called exactly 3 times (the max attempts)
	assert.Equal(t, 3, callCount)
}

func TestRetryManagerWithContextCancellation(t *testing.T) {
	// Create a retry manager with 3 max attempts
	retryManager := NewRetryManager(3, 10*time.Millisecond, 100*time.Millisecond)

	// Create a context that will be canceled
	ctx, cancel := context.WithCancel(context.Background())

	// Create a function that fails and triggers context cancellation
	callCount := 0
	go func() {
		time.Sleep(15 * time.Millisecond)
		cancel()
	}()

	err := retryManager.RetryOperation(ctx, func() error {
		callCount++
		time.Sleep(20 * time.Millisecond)
		return errors.New("temporary error")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.LessOrEqual(t, callCount, 2) // Should be 1 or 2 depending on timing
}

func TestRetryManagerWithBackoff(t *testing.T) {
	// Create a retry manager with 5 max attempts and exponential backoff
	retryManager := NewRetryManager(5, 10*time.Millisecond, 100*time.Millisecond)

	// Create a function that fails the first 3 times and succeeds on the 4th attempt
	callCount := 0
	startTime := time.Now()
	err := retryManager.RetryOperation(context.Background(), func() error {
		callCount++
		if callCount < 4 {
			return errors.New("temporary error")
		}
		return nil
	})
	duration := time.Since(startTime)

	assert.NoError(t, err)
	assert.Equal(t, 4, callCount)
	// The total time should be at least the sum of the first 3 backoff intervals
	// Initial: 10ms, 2nd: ~20ms, 3rd: ~40ms = ~70ms minimum
	// Use a lower value to account for timing variations in CI environments
	assert.GreaterOrEqual(t, duration, 30*time.Millisecond)
}
