package vpn

import (
	"errors"
	"fmt"
)

// Error codes
var (
	ErrVirtualIPNotSet  = errors.New("virtual IP is not set")
	ErrRoutesNotSet     = errors.New("routes are not set")
	ErrInvalidVirtualIP = errors.New("virtual IP is invalid")
	ErrInvalidRoutes    = errors.New("routes are invalid")
	ErrInvalidMTU       = errors.New("invalid MTU (must be between 576 and 9000)")
	ErrPeerNotFound     = errors.New("peer not found")
	ErrStreamClosed     = errors.New("stream closed")
)

// VPNError represents a VPN-specific error with a code and message
type VPNError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// ErrorCode defines the type of error
type ErrorCode int

const (
	ErrCodeNetworkFailure ErrorCode = iota
	ErrCodeConfigurationInvalid
	ErrCodePeerNotFound
	ErrCodeStreamCreationFailed
	ErrCodePacketProcessingFailed
)

// Error returns the error message
func (e *VPNError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *VPNError) Unwrap() error {
	return e.Err
}

// Helper functions for creating errors
func NewNetworkError(msg string, err error) *VPNError {
	return &VPNError{
		Code:    ErrCodeNetworkFailure,
		Message: msg,
		Err:     err,
	}
}

func NewConfigError(msg string, err error) *VPNError {
	return &VPNError{
		Code:    ErrCodeConfigurationInvalid,
		Message: msg,
		Err:     err,
	}
}

func NewPeerError(msg string, err error) *VPNError {
	return &VPNError{
		Code:    ErrCodePeerNotFound,
		Message: msg,
		Err:     err,
	}
}

func NewStreamError(msg string, err error) *VPNError {
	return &VPNError{
		Code:    ErrCodeStreamCreationFailed,
		Message: msg,
		Err:     err,
	}
}

func NewPacketError(msg string, err error) *VPNError {
	return &VPNError{
		Code:    ErrCodePacketProcessingFailed,
		Message: msg,
		Err:     err,
	}
}
