# VPN Resilience Package

This package implements resilience patterns for the VPN system, including circuit breakers and retries.

## Components

### CircuitBreaker

The `CircuitBreaker` implements the circuit breaker pattern, including:

- Closed state: requests are allowed
- Open state: requests are blocked
- Half-open state: a limited number of requests are allowed to test the service

### CircuitBreakerManager

The `CircuitBreakerManager` manages circuit breakers for different destinations, including:

- Creating and retrieving circuit breakers
- Executing operations with circuit breaker protection
- Resetting circuit breakers
- Collecting metrics

### RetryManager

The `RetryManager` handles retry operations with exponential backoff, including:

- Retrying operations with increasing delays
- Limiting the number of retry attempts
- Respecting context cancellation
- Collecting metrics

## Usage

```go
// Create a circuit breaker manager
circuitBreakerMgr := resilience.NewCircuitBreakerManager(
    failureThreshold,
    resetTimeout,
    successThreshold,
)

// Execute an operation with circuit breaker protection
err := circuitBreakerMgr.Execute("destination", func() error {
    // Operation to execute
    return nil
})

// Reset a circuit breaker
circuitBreakerMgr.Reset("destination")

// Create a retry manager
retryManager := resilience.NewRetryManager(
    maxAttempts,
    initialInterval,
    maxInterval,
)

// Retry an operation
err := retryManager.RetryOperation(ctx, func() error {
    // Operation to retry
    return nil
})
```
