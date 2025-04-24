package types

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrEmptyPacket            = errors.New("empty packet")
	ErrInvalidPacketLength    = errors.New("invalid packet length")
	ErrUnsupportedIPVersion   = errors.New("unsupported IP version")
	ErrNoStreamService        = errors.New("no stream service available")
	ErrStreamCreationFailed   = errors.New("failed to create stream")
	ErrStreamWriteFailed      = errors.New("failed to write to stream")
	ErrWorkerNotFound         = errors.New("worker not found")
	ErrWorkerQueueFull        = errors.New("worker queue full")
	ErrWorkerStopped          = errors.New("worker stopped")
	ErrWorkerChannelFull      = errors.New("worker channel full")
	ErrWorkerPoolStopped      = errors.New("worker pool stopped")
	ErrMaxWorkersReached      = errors.New("maximum number of workers reached")
	ErrDispatcherStopped      = errors.New("dispatcher stopped")
	ErrNoPeerMapping          = errors.New("no peer mapping found")
	ErrInvalidPeerID          = errors.New("invalid peer ID")
	ErrNoStreamAvailable      = errors.New("no stream available")
	ErrStreamPoolExhausted    = errors.New("stream pool exhausted")
	ErrStreamChannelClosed    = errors.New("stream channel closed")
	ErrStreamChannelFull      = errors.New("stream channel full")
	ErrContextCancelled       = errors.New("context cancelled")
	ErrNoHealthyStreams       = errors.New("no healthy streams available")
	ErrStreamAssignmentFailed = errors.New("failed to assign stream")
	// Worker manager errors
	ErrWorkerManagerBusy    = errors.New("worker manager is busy")
	ErrWorkerManagerTimeout = errors.New("worker manager operation timed out")
	// Retry and resilience errors
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

// NetworkError represents a network-related error with context information
type NetworkError struct {
	// Underlying error
	Err error
	// Operation that failed
	Op string
	// Destination IP
	DestIP string
	// Peer ID
	PeerID string
}

// Error implements the error interface
func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error during %s to %s (peer %s): %v", e.Op, e.DestIP, e.PeerID, e.Err)
}

// Unwrap returns the underlying error
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// ServiceError represents a service-related error with context information
type ServiceError struct {
	// Underlying error
	Err error
	// Service that failed
	Service string
	// Operation that failed
	Op string
	// Connection key
	ConnKey string
	// Peer ID
	PeerID string
	// Destination IP
	DestIP string
}

// Error implements the error interface
func (e *ServiceError) Error() string {
	return fmt.Sprintf("service error in %s.%s for %s (peer %s, dest %s): %v",
		e.Service, e.Op, e.ConnKey, e.PeerID, e.DestIP, e.Err)
}

// Unwrap returns the underlying error
func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewNetworkError creates a new NetworkError
func NewNetworkError(err error, op, destIP, peerID string) *NetworkError {
	return &NetworkError{
		Err:    err,
		Op:     op,
		DestIP: destIP,
		PeerID: peerID,
	}
}

// NewServiceError creates a new ServiceError
func NewServiceError(err error, service, op, connKey, peerID, destIP string) *ServiceError {
	return &ServiceError{
		Err:     err,
		Service: service,
		Op:      op,
		ConnKey: connKey,
		PeerID:  peerID,
		DestIP:  destIP,
	}
}
