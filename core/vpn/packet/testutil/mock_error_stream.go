package testutil

import (
	"errors"
	"io"
	"sync"
	"time"
)

// TestErrorStream implements the VPNStream interface for testing error scenarios
type TestErrorStream struct {
	mu           sync.Mutex
	closed       bool
	ErrorOnWrite bool
	ErrorType    string
	writeCount   int
}

// Read implements the io.Reader interface
func (m *TestErrorStream) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

// Write implements the io.Writer interface
func (m *TestErrorStream) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, errors.New("stream closed")
	}

	m.writeCount++

	if m.ErrorOnWrite {
		switch m.ErrorType {
		case "connection_reset":
			return 0, errors.New("connection reset by peer")
		case "stream_closed":
			return 0, errors.New("stream closed")
		case "temporary":
			return 0, &temporaryError{msg: "temporary error"}
		default:
			return 0, errors.New("write error")
		}
	}

	return len(p), nil
}

// Close implements the io.Closer interface
func (m *TestErrorStream) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("stream already closed")
	}

	m.closed = true
	return nil
}

// Reset implements the network.Stream interface
func (m *TestErrorStream) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

// SetDeadline implements the network.Stream interface
func (m *TestErrorStream) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline implements the network.Stream interface
func (m *TestErrorStream) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline implements the network.Stream interface
func (m *TestErrorStream) SetWriteDeadline(t time.Time) error {
	return nil
}

// temporaryError implements the net.Error interface for testing temporary errors
type temporaryError struct {
	msg string
}

func (e *temporaryError) Error() string {
	return e.msg
}

func (e *temporaryError) Timeout() bool {
	return true
}

func (e *temporaryError) Temporary() bool {
	return true
}

// TestStream implements the VPNStream interface for testing
type TestStream struct {
	mu         sync.Mutex
	closed     bool
	writeCount int
}

// Read implements the io.Reader interface
func (m *TestStream) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

// Write implements the io.Writer interface
func (m *TestStream) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, errors.New("stream closed")
	}

	m.writeCount++
	return len(p), nil
}

// Close implements the io.Closer interface
func (m *TestStream) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("stream already closed")
	}

	m.closed = true
	return nil
}

// Reset implements the network.Stream interface
func (m *TestStream) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

// SetDeadline implements the network.Stream interface
func (m *TestStream) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline implements the network.Stream interface
func (m *TestStream) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline implements the network.Stream interface
func (m *TestStream) SetWriteDeadline(t time.Time) error {
	return nil
}
