package pool

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

// MockStreamServiceBase is a mock implementation of the StreamService interface
type MockStreamServiceBase struct {
	mock.Mock
}

func (m *MockStreamServiceBase) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (api.VPNStream, error) {
	args := m.Called(ctx, peerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(api.VPNStream), args.Error(1)
}

// MockStreamBase is a mock implementation of the VPNStream interface
type MockStreamBase struct {
	mock.Mock
}

func (m *MockStreamBase) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockStreamBase) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockStreamBase) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStreamBase) Reset() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStreamBase) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockStreamBase) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockStreamBase) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

// MockStreamService is a mock implementation of the StreamService interface for tests
type MockStreamService struct {
	MockStreamServiceBase
}

// MockStream is a mock implementation of the VPNStream interface for tests
type MockStream struct {
	MockStreamBase
	closed bool
}

// Close overrides the base Close method to track closed state
func (m *MockStream) Close() error {
	args := m.Called()
	m.closed = true
	return args.Error(0)
}

// MockStreamServiceV2 is a mock implementation of the StreamService interface for V2 tests
type MockStreamServiceV2 struct {
	MockStreamServiceBase
}

// GetPacketBufferSize implements ConfigurableStreamService
func (m *MockStreamServiceV2) GetPacketBufferSize() int {
	args := m.Called()
	return args.Int(0)
}
