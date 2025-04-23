package testutil

import (
	"context"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

// MockStreamService is a mock implementation of api.StreamService for testing
type MockStreamService struct {
	mu            sync.Mutex
	stream        api.VPNStream
	getStreamErr  error
}

// NewMockStreamService creates a new mock stream service
func NewMockStreamService(stream api.VPNStream) *MockStreamService {
	return &MockStreamService{
		stream: stream,
	}
}

// SetGetStreamError sets the error to return from CreateNewVPNStream
func (m *MockStreamService) SetGetStreamError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getStreamErr = err
}

// CreateNewVPNStream returns the mock stream
func (m *MockStreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (api.VPNStream, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getStreamErr != nil {
		return nil, m.getStreamErr
	}
	return m.stream, nil
}
