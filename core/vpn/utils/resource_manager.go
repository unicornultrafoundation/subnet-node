package utils

import (
	"io"
	"slices"
	"sync"
)

// ResourceManager is a utility for tracking and cleaning up resources.
// It ensures that all registered resources are properly closed when the manager is closed.
type ResourceManager struct {
	resources []io.Closer
	mu        sync.RWMutex
}

// NewResourceManager creates a new resource manager.
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources: make([]io.Closer, 0),
	}
}

// Register adds a resource to be managed.
// The resources will be closed in reverse order when the manager is closed.
func (m *ResourceManager) Register(resource io.Closer) {
	if resource == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.resources = append(m.resources, resource)
}

// Unregister removes a resource from the manager.
// This is useful when a resource is closed manually and should no longer be managed.
func (m *ResourceManager) Unregister(resource io.Closer) {
	if resource == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for i, r := range m.resources {
		if r == resource {
			// Remove the resource from the slice
			m.resources = slices.Delete(m.resources, i, i+1)
			return
		}
	}
}

// GetResourceCount returns the number of registered resources.
func (m *ResourceManager) GetResourceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.resources)
}

// Close closes all registered resources in reverse order.
// This ensures that dependent resources are closed in the correct order.
// If a resource fails to close, the error is logged but the process continues.
func (m *ResourceManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error

	// Close resources in reverse order (LIFO)
	for i := len(m.resources) - 1; i >= 0; i-- {
		resource := m.resources[i]
		if err := resource.Close(); err != nil {
			// Log the error but continue closing other resources
			lastErr = err
		}
	}

	// Clear the resources slice
	m.resources = make([]io.Closer, 0)

	return lastErr
}
