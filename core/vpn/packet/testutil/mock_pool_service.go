package testutil

import (
	"context"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// MockPoolService is a mock implementation of types.PoolService for testing
type MockPoolService struct {
	mu            sync.Mutex
	stream        types.VPNStream
	getStreamErr  error
	releaseStream bool
}

// NewMockPoolService creates a new mock pool service
func NewMockPoolService(stream types.VPNStream) *MockPoolService {
	return &MockPoolService{
		stream: stream,
	}
}

// SetGetStreamError sets the error to return from GetStream
func (m *MockPoolService) SetGetStreamError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getStreamErr = err
}

// GetStream returns the mock stream
func (m *MockPoolService) GetStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getStreamErr != nil {
		return nil, m.getStreamErr
	}
	return m.stream, nil
}

// ReleaseStream records that the stream was released
func (m *MockPoolService) ReleaseStream(peerID peer.ID, stream types.VPNStream, healthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.releaseStream = true
}

// WasStreamReleased returns whether ReleaseStream was called
func (m *MockPoolService) WasStreamReleased() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.releaseStream
}
