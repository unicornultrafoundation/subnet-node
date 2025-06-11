package ipfs

import (
	"fmt"
	"sync"
	"time"
)

// MockClient is a mock implementation of the IPFS client
type MockClient struct {
	content map[string][]byte
	mu      sync.RWMutex
}

// NewMockClient creates a new mock IPFS client
func NewMockClient() *MockClient {
	return &MockClient{
		content: make(map[string][]byte),
		mu:      sync.RWMutex{},
	}
}

// Add adds content to IPFS
func (c *MockClient) Add(content []byte) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := fmt.Sprintf("mock-hash-%d", time.Now().UnixNano())
	c.content[hash] = content
	return hash, nil
}

// Get gets content from IPFS
func (c *MockClient) Get(hash string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if content, ok := c.content[hash]; ok {
		return content, nil
	}
	return nil, fmt.Errorf("content not found for hash: %s", hash)
}
