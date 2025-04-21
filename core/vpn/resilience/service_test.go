package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

func TestResilienceServiceWithDefaultConfig(t *testing.T) {
	// Create a resilience service with default config
	service := NewResilienceService(nil)

	// Verify the default config was applied
	assert.NotNil(t, service.circuitBreakerMgr)
	assert.NotNil(t, service.retryManager)
	assert.NotNil(t, service.config)
}

func TestResilienceServiceWithCustomConfig(t *testing.T) {
	// Create a custom config
	config := &ResilienceConfig{
		CircuitBreakerFailureThreshold: 10,
		CircuitBreakerResetTimeout:     60 * time.Second,
		CircuitBreakerSuccessThreshold: 3,
		RetryMaxAttempts:               5,
		RetryInitialInterval:           200 * time.Millisecond,
		RetryMaxInterval:               5 * time.Second,
	}

	// Create a resilience service with custom config
	service := NewResilienceService(config)

	// Verify the custom config was applied
	assert.Equal(t, config, service.config)
}

func TestExecuteWithResilience(t *testing.T) {
	// Create a resilience service
	service := NewResilienceService(nil)

	// Test successful operation
	callCount := 0
	err, attempts := service.ExecuteWithResilience(context.Background(), "test", func() error {
		callCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
	assert.Equal(t, 1, callCount)

	// Test operation that fails once then succeeds
	callCount = 0
	err, attempts = service.ExecuteWithResilience(context.Background(), "test", func() error {
		callCount++
		if callCount == 1 {
			return errors.New("temporary error")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
	assert.Equal(t, 2, callCount)

	// Test operation that always fails
	callCount = 0
	err, attempts = service.ExecuteWithResilience(context.Background(), "test", func() error {
		callCount++
		return errors.New("persistent error")
	})

	assert.Error(t, err)
	assert.Equal(t, "persistent error", err.Error())
	// The attempts count includes the initial attempt plus retries
	assert.Equal(t, service.config.RetryMaxAttempts, callCount)
	// The attempts returned may include the final attempt that determined max was reached
	assert.GreaterOrEqual(t, attempts, service.config.RetryMaxAttempts)
}

func TestFormatBreakerIds(t *testing.T) {
	// Create a resilience service
	service := NewResilienceService(nil)

	// Test peer breaker ID formatting
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	peerId := service.FormatPeerBreakerId(peerID, "stream")
	assert.Equal(t, "peer-QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N-stream", peerId)

	// Test worker breaker ID formatting
	workerId := service.FormatWorkerBreakerId("192.168.1.1:80", "write")
	assert.Equal(t, "worker-192.168.1.1:80-write", workerId)
}

func TestGetMetrics(t *testing.T) {
	// Create a resilience service
	service := NewResilienceService(nil)

	// Execute some operations to generate metrics
	service.ExecuteWithResilience(context.Background(), "test", func() error {
		return nil
	})

	service.ExecuteWithResilience(context.Background(), "test-fail", func() error {
		return errors.New("error")
	})

	// Get metrics
	metrics := service.GetMetrics()

	// Verify metrics
	assert.NotEmpty(t, metrics)
	assert.Contains(t, metrics, "retry_retry_attempts")
	assert.Contains(t, metrics, "retry_retry_success")
	assert.Contains(t, metrics, "retry_retry_failure")
	assert.Contains(t, metrics, "cb_request_allow_count")
}
