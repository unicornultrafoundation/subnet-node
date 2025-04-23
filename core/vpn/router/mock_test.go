package router

import (
	"context"
	"net"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

// We're using the MockMode from mock_packet.go

// MockStreamCreator implements the StreamCreator function type for testing
type MockStreamCreator struct {
	mock.Mock
}

// CreateNewVPNStream implements the StreamCreator function type for testing
func (m *MockStreamCreator) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (api.VPNStream, error) {
	args := m.Called(ctx, peerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(api.VPNStream), args.Error(1)
}

// MockVPNStream implements api.VPNStream for testing
type MockVPNStream struct {
	mock.Mock
}

func (m *MockVPNStream) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockVPNStream) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockVPNStream) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockVPNStream) Reset() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockVPNStream) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockVPNStream) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockVPNStream) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

// MockPeerDiscoveryService implements api.PeerDiscoveryService for testing
type MockPeerDiscoveryService struct {
	mock.Mock
}

func (m *MockPeerDiscoveryService) GetPeerID(ctx context.Context, ip string) (string, error) {
	args := m.Called(ctx, ip)
	return args.String(0), args.Error(1)
}

func (m *MockPeerDiscoveryService) GetPeerIDByRegistry(ctx context.Context, ip string) (string, error) {
	args := m.Called(ctx, ip)
	return args.String(0), args.Error(1)
}

func (m *MockPeerDiscoveryService) GetVirtualIP(ctx context.Context, peerID string) (string, error) {
	args := m.Called(ctx, peerID)
	return args.String(0), args.Error(1)
}

func (m *MockPeerDiscoveryService) StoreMappingInDHT(ctx context.Context, peerID string) error {
	args := m.Called(ctx, peerID)
	return args.Error(0)
}

func (m *MockPeerDiscoveryService) VerifyVirtualIPHasRegistered(ctx context.Context, virtualIP string) error {
	args := m.Called(ctx, virtualIP)
	return args.Error(0)
}

func (m *MockPeerDiscoveryService) SyncPeerIDToDHT(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockConfig implements config.C for testing
type MockConfig struct {
	values map[string]interface{}
}

func NewMockConfig(values map[string]interface{}) *MockConfig {
	return &MockConfig{
		values: values,
	}
}

func (c *MockConfig) GetBool(key string, defaultValue bool) bool {
	if val, ok := c.values[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

func (c *MockConfig) GetInt(key string, defaultValue int) int {
	if val, ok := c.values[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return defaultValue
}

func (c *MockConfig) GetString(key string, defaultValue string) string {
	if val, ok := c.values[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func (c *MockConfig) GetDuration(key string, defaultValue time.Duration) time.Duration {
	if val, ok := c.values[key]; ok {
		if strVal, ok := val.(string); ok {
			duration, err := time.ParseDuration(strVal)
			if err == nil {
				return duration
			}
		}
	}
	return defaultValue
}

// MockPacketInfo creates a mock packet info for testing
func MockPacketInfo() *struct {
	SrcIP    net.IP
	DstIP    net.IP
	SrcPort  *int
	DstPort  *int
	Protocol uint8
} {
	srcPort := 12345
	dstPort := 80
	return &struct {
		SrcIP    net.IP
		DstIP    net.IP
		SrcPort  *int
		DstPort  *int
		Protocol uint8
	}{
		SrcIP:    net.ParseIP("10.0.0.1"),
		DstIP:    net.ParseIP("192.168.1.1"),
		SrcPort:  &srcPort,
		DstPort:  &dstPort,
		Protocol: 6, // TCP
	}
}
