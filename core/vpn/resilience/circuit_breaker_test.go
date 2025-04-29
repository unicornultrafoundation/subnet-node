package resilience

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	// Create a circuit breaker
	cb := NewCircuitBreaker("test", 3, 100*time.Millisecond, 2)

	// Initially, the circuit should be closed
	assert.Equal(t, StateClosed, cb.GetState())

	// Test successful operations
	err := cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())

	// Test failed operations
	testErr := errors.New("test error")
	for i := 0; i < 2; i++ {
		err = cb.Execute(func() error {
			return testErr
		})
		assert.Equal(t, testErr, err)
		assert.Equal(t, StateClosed, cb.GetState())
	}

	// The third failure should open the circuit
	err = cb.Execute(func() error {
		return testErr
	})
	assert.Equal(t, testErr, err)
	assert.Equal(t, StateOpen, cb.GetState())

	// When the circuit is open, operations should not be executed
	callCount := 0
	err = cb.Execute(func() error {
		callCount++
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit open for test")
	assert.Equal(t, 0, callCount)

	// Wait for the reset timeout
	time.Sleep(150 * time.Millisecond)

	// The circuit should now be half-open
	err = cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, cb.GetState())

	// Another success should close the circuit
	err = cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())

	// Reset the circuit breaker
	cb.Reset()
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreakerManager(t *testing.T) {
	// Create a circuit breaker manager
	manager := NewCircuitBreakerManager(3, 100*time.Millisecond, 2)

	// Test getting a circuit breaker
	cb := manager.GetBreaker("test")
	assert.NotNil(t, cb)
	assert.Equal(t, StateClosed, cb.GetState())

	// Test executing an operation
	err := manager.Execute("test", func() error {
		return nil
	})
	assert.NoError(t, err)

	// Test executing a failing operation
	testErr := errors.New("test error")
	for i := 0; i < 3; i++ {
		err = manager.Execute("test", func() error {
			return testErr
		})
		assert.Equal(t, testErr, err)
	}

	// The circuit should now be open
	states := manager.GetBreakerStates()
	assert.Equal(t, StateOpen, states["test"])

	// Test resetting a circuit breaker
	manager.Reset("test")
	states = manager.GetBreakerStates()
	assert.Equal(t, StateClosed, states["test"])

	// Test metrics
	metrics := manager.GetMetrics()
	assert.Greater(t, metrics["circuit_open_count"], int64(0))
	assert.Greater(t, metrics["request_allow_count"], int64(0))
}

func TestCircuitBreakerConcurrency(t *testing.T) {
	// Create a circuit breaker manager
	manager := NewCircuitBreakerManager(5, 100*time.Millisecond, 2)

	// Test concurrent access
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Each goroutine uses a different destination
			dest := "test" + string(rune('A'+i))

			// Execute a successful operation
			err := manager.Execute(dest, func() error {
				return nil
			})
			assert.NoError(t, err)

			// Execute a failing operation
			err = manager.Execute(dest, func() error {
				return errors.New("test error")
			})
			assert.Error(t, err)
		}(i)
	}

	wg.Wait()

	// Verify that we have the expected number of breakers
	assert.Equal(t, numGoroutines, len(manager.GetAllBreakers()))
}

func TestCircuitBreakerHalfOpenFailure(t *testing.T) {
	// Create a circuit breaker
	cb := NewCircuitBreaker("test", 3, 100*time.Millisecond, 2)

	// Open the circuit
	testErr := errors.New("test error")
	for i := 0; i < 3; i++ {
		err := cb.Execute(func() error {
			return testErr
		})
		assert.Equal(t, testErr, err)
	}
	assert.Equal(t, StateOpen, cb.GetState())

	// Wait for the reset timeout
	time.Sleep(150 * time.Millisecond)

	// The circuit should now be half-open
	// A failure in half-open state should open the circuit again
	err := cb.Execute(func() error {
		return testErr
	})
	assert.Equal(t, testErr, err)
	assert.Equal(t, StateOpen, cb.GetState())
}
