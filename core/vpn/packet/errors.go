package packet

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Package-level error constants

// Service-related errors
var (
	// ErrNoStreamService is returned when no stream service is available
	ErrNoStreamService = errors.New("no stream service available")
	// ErrStreamCreationFailed is returned when a stream cannot be created
	ErrStreamCreationFailed = errors.New("stream creation failed")
	// ErrCircuitOpen is returned when a circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker open")
)

// Worker-related errors
var (
	// ErrWorkerQueueFull is returned when a worker's queue is full
	ErrWorkerQueueFull = errors.New("worker queue full")
	// ErrWorkerStopped is returned when a worker has been stopped
	ErrWorkerStopped = errors.New("worker stopped")
)

// Peer-related errors
var (
	// ErrPeerNotFound is returned when a peer ID cannot be found for an IP
	ErrPeerNotFound = errors.New("peer not found")
)

// NetworkError represents an error that occurred during network operations.
// It provides detailed context about the error, including whether it's temporary,
// a connection reset, or a stream closed error, as well as information about the
// worker, peer, and destination IP involved.
type NetworkError struct {
	// Err is the underlying error that occurred
	Err error
	// Temporary indicates whether the error is temporary and can be retried
	Temporary bool
	// ConnectionReset indicates whether the error is a connection reset
	ConnectionReset bool
	// StreamClosed indicates whether the error is a stream closed error
	StreamClosed bool
	// WorkerID is the ID of the worker associated with the error
	WorkerID string
	// PeerID is the ID of the peer associated with the error
	PeerID string
	// DestIP is the destination IP address associated with the error
	DestIP string
}

// Error implements the error interface
func (e *NetworkError) Error() string {
	errType := "network error"
	if e.Temporary {
		errType = "temporary network error"
	} else if e.ConnectionReset {
		errType = "connection reset"
	} else if e.StreamClosed {
		errType = "stream closed"
	}

	return fmt.Sprintf("%s for worker %s to peer %s (IP: %s): %v",
		errType, e.WorkerID, e.PeerID, e.DestIP, e.Err)
}

// Unwrap returns the underlying error
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// IsTemporary returns whether the error is temporary
func (e *NetworkError) IsTemporary() bool {
	return e.Temporary
}

// IsConnectionReset returns whether the error is a connection reset
func (e *NetworkError) IsConnectionReset() bool {
	return e.ConnectionReset
}

// IsStreamClosed returns whether the error is a stream closed error
func (e *NetworkError) IsStreamClosed() bool {
	return e.StreamClosed
}

// NewNetworkError creates a new NetworkError
func NewNetworkError(err error, workerID, peerID, destIP string) *NetworkError {
	// Determine the error type
	isConnectionReset := isConnectionReset(err)
	isStreamClosed := isStreamClosed(err)
	isTemporary := isTemporaryError(err)

	return &NetworkError{
		Err:             err,
		Temporary:       isTemporary,
		ConnectionReset: isConnectionReset,
		StreamClosed:    isStreamClosed,
		WorkerID:        workerID,
		PeerID:          peerID,
		DestIP:          destIP,
	}
}

// ServiceError represents an error that occurred during service operations.
// It provides context about which service and operation failed, as well as
// information about the worker, peer, and destination IP involved.
type ServiceError struct {
	// Err is the underlying error that occurred
	Err error
	// Service is the name of the service where the error occurred
	Service string
	// Operation is the name of the operation that failed
	Operation string
	// WorkerID is the ID of the worker associated with the error
	WorkerID string
	// PeerID is the ID of the peer associated with the error
	PeerID string
	// DestIP is the destination IP address associated with the error
	DestIP string
}

// Error implements the error interface
func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s service %s failed for worker %s to peer %s (IP: %s): %v",
		e.Service, e.Operation, e.WorkerID, e.PeerID, e.DestIP, e.Err)
}

// Unwrap returns the underlying error
func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewServiceError creates a new ServiceError
func NewServiceError(err error, service, operation, workerID, peerID, destIP string) *ServiceError {
	return &ServiceError{
		Err:       err,
		Service:   service,
		Operation: operation,
		WorkerID:  workerID,
		PeerID:    peerID,
		DestIP:    destIP,
	}
}

// IsNetworkError checks if an error is a NetworkError
func IsNetworkError(err error) bool {
	var netErr *NetworkError
	return errors.As(err, &netErr)
}

// IsServiceError checks if an error is a ServiceError
func IsServiceError(err error) bool {
	var svcErr *ServiceError
	return errors.As(err, &svcErr)
}

// IsTemporaryError checks if an error is temporary
func IsTemporaryError(err error) bool {
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return netErr.IsTemporary()
	}
	return isTemporaryError(err)
}

// IsConnectionResetError checks if an error is a connection reset
func IsConnectionResetError(err error) bool {
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return netErr.IsConnectionReset()
	}
	return isConnectionReset(err)
}

// IsStreamClosedError checks if an error is a stream closed error
func IsStreamClosedError(err error) bool {
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return netErr.IsStreamClosed()
	}
	return isStreamClosed(err)
}

// isConnectionReset checks if the error is a connection reset
func isConnectionReset(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "connection reset") ||
		strings.Contains(err.Error(), "broken pipe")
}

// isStreamClosed checks if the error is due to a closed stream
func isStreamClosed(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "stream closed") ||
		strings.Contains(err.Error(), "stream reset")
}

// isTemporaryError checks if the error is temporary
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}
	// Check for timeout errors, which are typically temporary
	if netErr, ok := err.(net.Error); ok {
		// Note: Temporary() is deprecated, but we still check it for backward compatibility
		return netErr.Timeout()
	}
	// Check for other common temporary error messages (case insensitive)
	errLower := strings.ToLower(err.Error())
	return strings.Contains(errLower, "temporary") ||
		strings.Contains(errLower, "timeout") ||
		strings.Contains(errLower, "timed out") ||
		strings.Contains(errLower, "would block")
}
