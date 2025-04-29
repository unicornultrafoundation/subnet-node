package utils

import (
	"errors"
	"testing"
)

// mockCloser is a mock implementation of io.Closer for testing
type mockCloser struct {
	closed    bool
	err       error
	id        int
	closeFunc func() error
}

func (m *mockCloser) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	m.closed = true
	return m.err
}

func TestResourceManager(t *testing.T) {
	// Test basic functionality and error handling
	t.Run("Basic functionality and error handling", func(t *testing.T) {
		manager := NewResourceManager()

		// Create resources with one that returns an error
		resource1 := &mockCloser{id: 1}
		resource2 := &mockCloser{id: 2, err: errors.New("close error")}

		// Register and close
		manager.Register(resource1)
		manager.Register(resource2)
		err := manager.Close()

		// Verify all resources were closed
		if !resource1.closed || !resource2.closed {
			t.Error("Not all resources were closed")
		}

		// Verify error was returned
		if err == nil || err.Error() != "close error" {
			t.Errorf("Expected 'close error', got %v", err)
		}
	})

	// Test unregister functionality
	t.Run("Unregister functionality", func(t *testing.T) {
		manager := NewResourceManager()

		// Create and register resources
		resource1 := &mockCloser{id: 1}
		resource2 := &mockCloser{id: 2}
		manager.Register(resource1)
		manager.Register(resource2)

		// Unregister one resource
		manager.Unregister(resource1)
		manager.Close()

		// Verify only the registered resource was closed
		if resource1.closed {
			t.Error("Unregistered resource was closed")
		}
		if !resource2.closed {
			t.Error("Registered resource was not closed")
		}
	})

	// Test close order
	t.Run("Close order", func(t *testing.T) {
		manager := NewResourceManager()
		closeOrder := make([]int, 0, 3)

		// Create resources that record close order
		resources := make([]*mockCloser, 3)
		for i := 0; i < 3; i++ {
			id := i + 1
			resources[i] = &mockCloser{id: id}
			resources[i].closeFunc = func(id int) func() error {
				return func() error {
					closeOrder = append(closeOrder, id)
					return nil
				}
			}(id)
			manager.Register(resources[i])
		}

		// Close and verify order
		manager.Close()
		expectedOrder := []int{3, 2, 1} // LIFO order
		for i, expected := range expectedOrder {
			if i >= len(closeOrder) || closeOrder[i] != expected {
				t.Errorf("Expected close order %v, got %v", expectedOrder, closeOrder)
				break
			}
		}
	})
}
