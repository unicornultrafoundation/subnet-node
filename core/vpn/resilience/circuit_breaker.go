package resilience

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int32

const (
	// StateClosed indicates the circuit is closed and requests are allowed
	StateClosed CircuitState = iota
	// StateOpen indicates the circuit is open and requests are not allowed
	StateOpen
	// StateHalfOpen indicates the circuit is half-open and a single request is allowed
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	// State of the circuit breaker (closed, open, half-open)
	state int32
	// Failure threshold before opening the circuit
	failureThreshold int
	// Current failure count
	failureCount int32
	// Reset timeout
	resetTimeout time.Duration
	// Last failure time
	lastFailureTime atomic.Value
	// Mutex for state changes
	mu sync.Mutex
	// Destination identifier (e.g., peer ID)
	destination string
	// Success threshold for half-open state
	successThreshold int
	// Current success count in half-open state
	successCount int32
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(destination string, failureThreshold int, resetTimeout time.Duration, successThreshold int) *CircuitBreaker {
	cb := &CircuitBreaker{
		state:            int32(StateClosed),
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		destination:      destination,
		successThreshold: successThreshold,
	}
	cb.lastFailureTime.Store(time.Time{})
	return cb
}

// Execute executes the given function if the circuit is closed or half-open
func (cb *CircuitBreaker) Execute(operation func() error) error {
	if !cb.AllowRequest() {
		return fmt.Errorf("circuit open for %s", cb.destination)
	}

	err := operation()

	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// AllowRequest checks if a request should be allowed
func (cb *CircuitBreaker) AllowRequest() bool {
	state := CircuitState(atomic.LoadInt32(&cb.state))

	if state == StateClosed {
		return true
	}

	if state == StateOpen {
		lastFailure := cb.lastFailureTime.Load().(time.Time)
		if time.Since(lastFailure) > cb.resetTimeout {
			// Try half-open state
			cb.mu.Lock()
			if CircuitState(cb.state) == StateOpen {
				atomic.StoreInt32(&cb.state, int32(StateHalfOpen))
				atomic.StoreInt32(&cb.successCount, 0)
			}
			cb.mu.Unlock()
			return true
		}
		return false
	}

	// Half-open state allows limited requests
	return true
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	state := CircuitState(atomic.LoadInt32(&cb.state))

	if state == StateHalfOpen {
		newSuccessCount := atomic.AddInt32(&cb.successCount, 1)
		if newSuccessCount >= int32(cb.successThreshold) {
			// Reset to closed state
			cb.mu.Lock()
			if CircuitState(cb.state) == StateHalfOpen {
				atomic.StoreInt32(&cb.state, int32(StateClosed))
				atomic.StoreInt32(&cb.failureCount, 0)
			}
			cb.mu.Unlock()
		}
	} else if state == StateClosed {
		// Reset failure count on success in closed state
		atomic.StoreInt32(&cb.failureCount, 0)
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	state := CircuitState(atomic.LoadInt32(&cb.state))

	if state == StateClosed {
		newCount := atomic.AddInt32(&cb.failureCount, 1)
		if newCount >= int32(cb.failureThreshold) {
			// Open the circuit
			cb.mu.Lock()
			if CircuitState(cb.state) == StateClosed && cb.failureCount >= int32(cb.failureThreshold) {
				atomic.StoreInt32(&cb.state, int32(StateOpen))
				cb.lastFailureTime.Store(time.Now())
			}
			cb.mu.Unlock()
		}
	} else if state == StateHalfOpen {
		// Any failure in half-open state opens the circuit again
		cb.mu.Lock()
		if CircuitState(cb.state) == StateHalfOpen {
			atomic.StoreInt32(&cb.state, int32(StateOpen))
			cb.lastFailureTime.Store(time.Now())
		}
		cb.mu.Unlock()
	} else {
		// Update last failure time for open state
		cb.lastFailureTime.Store(time.Now())
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	return CircuitState(atomic.LoadInt32(&cb.state))
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	atomic.StoreInt32(&cb.state, int32(StateClosed))
	atomic.StoreInt32(&cb.failureCount, 0)
	atomic.StoreInt32(&cb.successCount, 0)
}

// CircuitBreakerManager manages circuit breakers for different destinations
type CircuitBreakerManager struct {
	// Map of destination to circuit breaker
	breakers map[string]*CircuitBreaker
	// Mutex to protect access
	mu sync.RWMutex
	// Default settings
	failureThreshold int
	resetTimeout     time.Duration
	successThreshold int
	// Metrics
	metrics struct {
		circuitOpenCount   int64
		circuitCloseCount  int64
		circuitResetCount  int64
		requestBlockCount  int64
		requestAllowCount  int64
	}
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(failureThreshold int, resetTimeout time.Duration, successThreshold int) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers:         make(map[string]*CircuitBreaker),
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		successThreshold: successThreshold,
	}
}

// GetBreaker gets or creates a circuit breaker for a destination
func (m *CircuitBreakerManager) GetBreaker(destination string) *CircuitBreaker {
	// Fast path: check if breaker exists
	m.mu.RLock()
	breaker, exists := m.breakers[destination]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	// Slow path: create new breaker
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check again in case another goroutine created the breaker
	breaker, exists = m.breakers[destination]
	if exists {
		return breaker
	}

	// Create new breaker
	breaker = NewCircuitBreaker(destination, m.failureThreshold, m.resetTimeout, m.successThreshold)
	m.breakers[destination] = breaker

	return breaker
}

// Execute executes an operation with circuit breaker protection
func (m *CircuitBreakerManager) Execute(destination string, operation func() error) error {
	breaker := m.GetBreaker(destination)
	
	if !breaker.AllowRequest() {
		atomic.AddInt64(&m.metrics.requestBlockCount, 1)
		return fmt.Errorf("circuit open for %s", destination)
	}
	
	atomic.AddInt64(&m.metrics.requestAllowCount, 1)
	
	err := operation()
	
	if err != nil {
		breaker.RecordFailure()
		// Check if this failure opened the circuit
		if breaker.GetState() == StateOpen {
			atomic.AddInt64(&m.metrics.circuitOpenCount, 1)
		}
		return err
	}
	
	breaker.RecordSuccess()
	// Check if this success closed the circuit
	if breaker.GetState() == StateClosed {
		atomic.AddInt64(&m.metrics.circuitCloseCount, 1)
	}
	
	return nil
}

// Reset resets a circuit breaker for a destination
func (m *CircuitBreakerManager) Reset(destination string) {
	m.mu.RLock()
	breaker, exists := m.breakers[destination]
	m.mu.RUnlock()

	if exists {
		breaker.Reset()
		atomic.AddInt64(&m.metrics.circuitResetCount, 1)
	}
}

// GetMetrics returns the current metrics
func (m *CircuitBreakerManager) GetMetrics() map[string]int64 {
	return map[string]int64{
		"circuit_open_count":   atomic.LoadInt64(&m.metrics.circuitOpenCount),
		"circuit_close_count":  atomic.LoadInt64(&m.metrics.circuitCloseCount),
		"circuit_reset_count":  atomic.LoadInt64(&m.metrics.circuitResetCount),
		"request_block_count":  atomic.LoadInt64(&m.metrics.requestBlockCount),
		"request_allow_count":  atomic.LoadInt64(&m.metrics.requestAllowCount),
		"active_breakers":      int64(len(m.breakers)),
	}
}

// GetAllBreakers returns all circuit breakers
func (m *CircuitBreakerManager) GetAllBreakers() map[string]*CircuitBreaker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[string]*CircuitBreaker, len(m.breakers))
	for k, v := range m.breakers {
		result[k] = v
	}

	return result
}

// GetBreakerStates returns the states of all circuit breakers
func (m *CircuitBreakerManager) GetBreakerStates() map[string]CircuitState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]CircuitState, len(m.breakers))
	for k, v := range m.breakers {
		result[k] = v.GetState()
	}

	return result
}
