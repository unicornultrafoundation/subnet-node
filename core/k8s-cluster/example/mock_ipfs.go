package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// MockClient implements the IPFSClient interface with in-memory storage
type MockClient struct {
	mu    sync.RWMutex
	store map[string][]byte
}

// NewMockClient creates a new mock IPFS client
func NewMockClient() *MockClient {
	client := &MockClient{
		store: make(map[string][]byte),
	}

	// Read SDL content from echo-service.yaml
	sdlPath := "echo-service.yaml" // Use relative path since file is in the same directory
	sdlContent, err := os.ReadFile(sdlPath)
	if err != nil {
		// If file not found, use a default SDL content
		sdlContent = []byte(`version: "2.0"
services:
  echo:
    image: hashicorp/http-echo
    command: ["/http-echo"]
    args: ["-listen", ":5678", "-text", "Hello from http-echo!"]
    expose:
      - port: 5678
        as: 80
        proto: TCP
        to:
          - global: true
    resources:
      cpu:
        units:
          value: 1
          unit: m
      memory:
        size:
          value: 256
          unit: Mi
      storage:
        size:
          value: 512
          unit: Mi
    count: 1
    params:
      health:
        http:
          path: /
          port: 5678
          interval: "30s"
          timeout: "5s"
          retries: 3

profiles:
  compute:
    default:
      resources:
        cpu:
          request: 1
          limit: 2
        memory:
          request: 128
          limit: 256
        storage:
          - size: 512
  placement:
    default:
      attributes:
        region: us-west
      pricing:
        cpu:
          denom: uakt
          amount: 1000
        memory:
          denom: uakt
          amount: 500
        storage:
          denom: uakt
          amount: 200

deployment:
  default:
    profile: default
    count: 1

endpoints:
  echo:
    kind: web
`)
	}

	// Add the SDL content to IPFS with a known hash
	client.store["test-hash"] = sdlContent
	return client
}

// Get retrieves content from mock IPFS storage by hash
func (c *MockClient) Get(hash string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// For testing, always return the test SDL content
	if hash == "test-hash" {
		return c.store["test-hash"], nil
	}

	data, exists := c.store[hash]
	if !exists {
		return nil, fmt.Errorf("content not found for hash: %s", hash)
	}

	return data, nil
}

// Add adds content to mock IPFS storage and returns a mock hash
func (c *MockClient) Add(data []byte) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// For testing, always return the test hash
	return "test-hash", nil
}

// Verify that MockClient implements IPFSClient interface
var _ types.IPFSClient = (*MockClient)(nil)
