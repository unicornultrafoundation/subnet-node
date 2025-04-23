package testutil

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

// MockStreamManager is a mock implementation of the StreamManagerInterface
type MockStreamManager struct {
	mock.Mock
}

func (m *MockStreamManager) SendPacket(ctx context.Context, connKey types.ConnectionKey, peerID peer.ID, packet *types.QueuedPacket) error {
	args := m.Called(ctx, connKey, peerID, packet)
	return args.Error(0)
}

func (m *MockStreamManager) Start() {
	m.Called()
}

func (m *MockStreamManager) Stop() {
	m.Called()
}

func (m *MockStreamManager) GetMetrics() map[string]int64 {
	args := m.Called()
	return args.Get(0).(map[string]int64)
}

func (m *MockStreamManager) GetConnectionCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockStreamManager) Close() error {
	args := m.Called()
	return args.Error(0)
}
