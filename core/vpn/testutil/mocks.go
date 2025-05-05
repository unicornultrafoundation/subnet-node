package testutil

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

//
// Stream Mocks
//

// MockStreamConfig contains configuration for mock streams
type MockStreamConfig struct {
	Latency     time.Duration
	Jitter      time.Duration
	PacketLoss  float64
	FailureRate float64
}

// MockStream is a configurable mock implementation of api.VPNStream
type MockStream struct {
	mock.Mock
	Config MockStreamConfig
	closed bool
	mu     sync.Mutex
}

// NewMockStream creates a new mock stream with the given configuration
func NewMockStream(config *MockStreamConfig) *MockStream {
	if config == nil {
		config = &MockStreamConfig{}
	}
	return &MockStream{
		Config: *config,
	}
}

// Read implements the io.Reader interface
func (s *MockStream) Read(p []byte) (n int, err error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return 0, fmt.Errorf("stream closed")
	}
	s.mu.Unlock()

	// Simulate latency
	if s.Config.Latency > 0 {
		time.Sleep(s.Config.Latency)
	}

	// Simulate jitter
	if s.Config.Jitter > 0 {
		jitter := time.Duration(rand.Int63n(int64(s.Config.Jitter)))
		time.Sleep(jitter)
	}

	// Simulate packet loss
	if s.Config.PacketLoss > 0 && rand.Float64() < s.Config.PacketLoss {
		return 0, fmt.Errorf("simulated packet loss")
	}

	// Simulate failure
	if s.Config.FailureRate > 0 && rand.Float64() < s.Config.FailureRate {
		return 0, fmt.Errorf("simulated failure")
	}

	args := s.Called(p)
	return args.Int(0), args.Error(1)
}

// Write implements the io.Writer interface
func (s *MockStream) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return 0, fmt.Errorf("stream closed")
	}
	s.mu.Unlock()

	// Simulate latency
	if s.Config.Latency > 0 {
		time.Sleep(s.Config.Latency)
	}

	// Simulate jitter
	if s.Config.Jitter > 0 {
		jitter := time.Duration(rand.Int63n(int64(s.Config.Jitter)))
		time.Sleep(jitter)
	}

	// Simulate packet loss
	if s.Config.PacketLoss > 0 && rand.Float64() < s.Config.PacketLoss {
		return 0, fmt.Errorf("simulated packet loss")
	}

	// Simulate failure
	if s.Config.FailureRate > 0 && rand.Float64() < s.Config.FailureRate {
		return 0, fmt.Errorf("simulated failure")
	}

	args := s.Called(p)
	return args.Int(0), args.Error(1)
}

// Close implements the io.Closer interface
func (s *MockStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	args := s.Called()
	return args.Error(0)
}

// Reset implements the network.Stream interface
func (s *MockStream) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	args := s.Called()
	return args.Error(0)
}

// SetDeadline implements the net.Conn interface
func (s *MockStream) SetDeadline(t time.Time) error {
	args := s.Called(t)
	return args.Error(0)
}

// SetReadDeadline implements the net.Conn interface
func (s *MockStream) SetReadDeadline(t time.Time) error {
	args := s.Called(t)
	return args.Error(0)
}

// SetWriteDeadline implements the net.Conn interface
func (s *MockStream) SetWriteDeadline(t time.Time) error {
	args := s.Called(t)
	return args.Error(0)
}

//
// Stream Service Mocks
//

// MockServiceConfig contains configuration for mock services
type MockServiceConfig struct {
	Latency     time.Duration
	Jitter      time.Duration
	FailureRate float64
	PacketLoss  float64
}

// MockStreamService is a configurable mock implementation of api.StreamService
type MockStreamService struct {
	mock.Mock
	Config MockServiceConfig
	Stats  struct {
		StreamsCreated int64
		StreamErrors   int64
	}
	mutex          sync.Mutex
	latency        time.Duration
	jitter         time.Duration
	failureRate    float64
	packetLossRate float64
}

// NewMockStreamService creates a new mock stream service with the given configuration
func NewMockStreamService(config *MockServiceConfig) *MockStreamService {
	if config == nil {
		config = &MockServiceConfig{}
	}
	return &MockStreamService{
		Config:         *config,
		latency:        config.Latency,
		jitter:         config.Jitter,
		failureRate:    config.FailureRate,
		packetLossRate: config.PacketLoss,
	}
}

// CreateNewVPNStream implements the api.StreamService interface
func (s *MockStreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (api.VPNStream, error) {
	// Simulate latency
	if s.Config.Latency > 0 {
		time.Sleep(s.Config.Latency)
	}

	// Simulate jitter
	if s.Config.Jitter > 0 {
		jitter := time.Duration(rand.Int63n(int64(s.Config.Jitter)))
		time.Sleep(jitter)
	}

	// Simulate failure
	if s.Config.FailureRate > 0 && rand.Float64() < s.Config.FailureRate {
		atomic.AddInt64(&s.Stats.StreamErrors, 1)
		return nil, fmt.Errorf("simulated failure")
	}

	atomic.AddInt64(&s.Stats.StreamsCreated, 1)
	args := s.Called(ctx, peerID)
	return args.Get(0).(api.VPNStream), args.Error(1)
}

// GetStats returns the service statistics
func (s *MockStreamService) GetStats() map[string]int64 {
	return map[string]int64{
		"streams_created": atomic.LoadInt64(&s.Stats.StreamsCreated),
		"stream_errors":   atomic.LoadInt64(&s.Stats.StreamErrors),
	}
}

//
// Discovery Service Mocks
//

// MockDiscoveryServiceConfig contains configuration for mock discovery services
type MockDiscoveryServiceConfig struct {
	Latency     time.Duration
	Jitter      time.Duration
	FailureRate float64
}

// MockDiscoveryService is a configurable mock implementation of api.PeerDiscoveryService
type MockDiscoveryService struct {
	Config    MockDiscoveryServiceConfig
	PeerIDMap map[string]string
	Stats     struct {
		Lookups       int64
		LookupErrors  int64
		Registrations int64
		RegErrors     int64
	}
	mu          sync.RWMutex
	mutex       sync.Mutex
	failureRate float64
}

// NewMockDiscoveryService creates a new mock discovery service with the given configuration
func NewMockDiscoveryService(config *MockDiscoveryServiceConfig) *MockDiscoveryService {
	if config == nil {
		config = &MockDiscoveryServiceConfig{}
	}
	return &MockDiscoveryService{
		Config:      *config,
		PeerIDMap:   make(map[string]string),
		failureRate: config.FailureRate,
	}
}

// GetPeerID implements the discovery service interface
func (s *MockDiscoveryService) GetPeerID(ctx context.Context, destIP string) (string, error) {
	// Simulate latency
	if s.Config.Latency > 0 {
		time.Sleep(s.Config.Latency)
	}

	// Simulate jitter
	if s.Config.Jitter > 0 {
		jitter := time.Duration(rand.Int63n(int64(s.Config.Jitter)))
		time.Sleep(jitter)
	}

	atomic.AddInt64(&s.Stats.Lookups, 1)

	// Simulate failure
	if s.Config.FailureRate > 0 && rand.Float64() < s.Config.FailureRate {
		atomic.AddInt64(&s.Stats.LookupErrors, 1)
		return "", fmt.Errorf("simulated failure")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	peerID, exists := s.PeerIDMap[destIP]
	if !exists {
		atomic.AddInt64(&s.Stats.LookupErrors, 1)
		return "", fmt.Errorf("no peer mapping found for IP %s", destIP)
	}

	return peerID, nil
}

// SyncPeerIDToDHT implements the discovery service interface
func (s *MockDiscoveryService) SyncPeerIDToDHT(ctx context.Context) error {
	// Simulate latency
	if s.Config.Latency > 0 {
		time.Sleep(s.Config.Latency)
	}

	// Simulate jitter
	if s.Config.Jitter > 0 {
		jitter := time.Duration(rand.Int63n(int64(s.Config.Jitter)))
		time.Sleep(jitter)
	}

	atomic.AddInt64(&s.Stats.Registrations, 1)

	// Simulate failure
	if s.Config.FailureRate > 0 && rand.Float64() < s.Config.FailureRate {
		atomic.AddInt64(&s.Stats.RegErrors, 1)
		return fmt.Errorf("simulated failure")
	}

	return nil
}

// GetStats returns the service statistics
func (s *MockDiscoveryService) GetStats() map[string]int64 {
	return map[string]int64{
		"lookups":       atomic.LoadInt64(&s.Stats.Lookups),
		"lookup_errors": atomic.LoadInt64(&s.Stats.LookupErrors),
		"registrations": atomic.LoadInt64(&s.Stats.Registrations),
		"reg_errors":    atomic.LoadInt64(&s.Stats.RegErrors),
	}
}

// GetVirtualIP implements the api.PeerDiscoveryService interface
func (s *MockDiscoveryService) GetVirtualIP(ctx context.Context, peerID string) (string, error) {
	// Simulate latency
	if s.Config.Latency > 0 {
		time.Sleep(s.Config.Latency)
	}

	// Simulate jitter
	if s.Config.Jitter > 0 {
		jitter := time.Duration(rand.Int63n(int64(s.Config.Jitter)))
		time.Sleep(jitter)
	}

	// Simulate failure
	if s.Config.FailureRate > 0 && rand.Float64() < s.Config.FailureRate {
		return "", fmt.Errorf("simulated failure")
	}

	// Return a fixed virtual IP for testing
	return "10.0.0.1", nil
}

// StoreMappingInDHT implements the api.PeerDiscoveryService interface
func (s *MockDiscoveryService) StoreMappingInDHT(ctx context.Context, peerID string) error {
	// Simulate latency
	if s.Config.Latency > 0 {
		time.Sleep(s.Config.Latency)
	}

	// Simulate jitter
	if s.Config.Jitter > 0 {
		jitter := time.Duration(rand.Int63n(int64(s.Config.Jitter)))
		time.Sleep(jitter)
	}

	// Simulate failure
	if s.Config.FailureRate > 0 && rand.Float64() < s.Config.FailureRate {
		return fmt.Errorf("simulated failure")
	}

	return nil
}

// VerifyVirtualIPHasRegistered implements the api.PeerDiscoveryService interface
func (s *MockDiscoveryService) VerifyVirtualIPHasRegistered(ctx context.Context, virtualIP string) error {
	// Simulate latency
	if s.Config.Latency > 0 {
		time.Sleep(s.Config.Latency)
	}

	// Simulate jitter
	if s.Config.Jitter > 0 {
		jitter := time.Duration(rand.Int63n(int64(s.Config.Jitter)))
		time.Sleep(jitter)
	}

	// Simulate failure
	if s.Config.FailureRate > 0 && rand.Float64() < s.Config.FailureRate {
		return fmt.Errorf("simulated failure")
	}

	return nil
}

// GetPeerIDByRegistry implements the api.PeerDiscoveryService interface
func (s *MockDiscoveryService) GetPeerIDByRegistry(ctx context.Context, destIP string) (string, error) {
	// For simplicity, just delegate to GetPeerID
	return s.GetPeerID(ctx, destIP)
}

//
// Account Service Mocks
//

// MockAccountService is a mock implementation of the AccountService interface
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) IPRegistry() api.IPRegistry {
	args := m.Called()
	return args.Get(0).(api.IPRegistry)
}

// MockIPRegistry is a mock implementation of the api.IPRegistry interface
type MockIPRegistry struct {
	mock.Mock
}

func (m *MockIPRegistry) GetPeer(opts interface{}, tokenID *big.Int) (string, error) {
	args := m.Called(opts, tokenID)
	return args.String(0), args.Error(1)
}

//
// Peerstore Mocks
//

// MockPeerstore is a mock implementation of the peerstore.Peerstore interface
type MockPeerstore struct {
	mock.Mock
}

func (m *MockPeerstore) AddAddr(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	m.Called(p, addr, ttl)
}

func (m *MockPeerstore) AddAddrs(p peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration) {
	m.Called(p, addrs, ttl)
}

func (m *MockPeerstore) SetAddrs(p peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration) {
	m.Called(p, addrs, ttl)
}

func (m *MockPeerstore) Addrs(p peer.ID) []multiaddr.Multiaddr {
	args := m.Called(p)
	return args.Get(0).([]multiaddr.Multiaddr)
}

func (m *MockPeerstore) AddrStream(ctx context.Context, p peer.ID) <-chan multiaddr.Multiaddr {
	args := m.Called(ctx, p)
	return args.Get(0).(<-chan multiaddr.Multiaddr)
}

func (m *MockPeerstore) ClearAddrs(p peer.ID) {
	m.Called(p)
}

func (m *MockPeerstore) PeersWithAddrs() peer.IDSlice {
	args := m.Called()
	return args.Get(0).(peer.IDSlice)
}

func (m *MockPeerstore) PubKey(p peer.ID) crypto.PubKey {
	args := m.Called(p)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(crypto.PubKey)
}

func (m *MockPeerstore) AddPubKey(p peer.ID, pk crypto.PubKey) error {
	args := m.Called(p, pk)
	return args.Error(0)
}

func (m *MockPeerstore) PrivKey(p peer.ID) crypto.PrivKey {
	args := m.Called(p)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(crypto.PrivKey)
}

func (m *MockPeerstore) AddPrivKey(p peer.ID, sk crypto.PrivKey) error {
	args := m.Called(p, sk)
	return args.Error(0)
}

func (m *MockPeerstore) GetProtocols(p peer.ID) ([]protocol.ID, error) {
	args := m.Called(p)
	return args.Get(0).([]protocol.ID), args.Error(1)
}

func (m *MockPeerstore) AddProtocols(p peer.ID, s ...protocol.ID) error {
	args := m.Called(p, s)
	return args.Error(0)
}

func (m *MockPeerstore) SetProtocols(p peer.ID, s ...protocol.ID) error {
	args := m.Called(p, s)
	return args.Error(0)
}

func (m *MockPeerstore) SupportsProtocols(p peer.ID, s ...protocol.ID) ([]protocol.ID, error) {
	args := m.Called(p, s)
	return args.Get(0).([]protocol.ID), args.Error(1)
}

func (m *MockPeerstore) FirstSupportedProtocol(p peer.ID, s ...protocol.ID) (protocol.ID, error) {
	args := m.Called(p, s)
	return protocol.ID(args.String(0)), args.Error(1)
}

func (m *MockPeerstore) Peers() peer.IDSlice {
	args := m.Called()
	return args.Get(0).(peer.IDSlice)
}

func (m *MockPeerstore) Get(p peer.ID, key string) (interface{}, error) {
	args := m.Called(p, key)
	return args.Get(0), args.Error(1)
}

func (m *MockPeerstore) Put(p peer.ID, key string, val interface{}) error {
	args := m.Called(p, key, val)
	return args.Error(0)
}

func (m *MockPeerstore) Close() error {
	args := m.Called()
	return args.Error(0)
}

//
// Host Mocks
//

// MockHost is a mock implementation of the p2phost.Host interface
type MockHost struct {
	mock.Mock
}

func (m *MockHost) ID() peer.ID {
	args := m.Called()
	return args.Get(0).(peer.ID)
}

func (m *MockHost) Peerstore() peerstore.Peerstore {
	args := m.Called()
	return args.Get(0).(peerstore.Peerstore)
}

func (m *MockHost) Addrs() []multiaddr.Multiaddr {
	args := m.Called()
	return args.Get(0).([]multiaddr.Multiaddr)
}

func (m *MockHost) Network() network.Network {
	args := m.Called()
	return args.Get(0).(network.Network)
}

func (m *MockHost) Mux() protocol.Switch {
	args := m.Called()
	return args.Get(0).(protocol.Switch)
}

func (m *MockHost) Connect(ctx context.Context, pi peer.AddrInfo) error {
	args := m.Called(ctx, pi)
	return args.Error(0)
}

func (m *MockHost) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	m.Called(pid, handler)
}

func (m *MockHost) SetStreamHandlerMatch(pid protocol.ID, match func(string) bool, handler network.StreamHandler) {
	m.Called(pid, match, handler)
}

func (m *MockHost) RemoveStreamHandler(pid protocol.ID) {
	m.Called(pid)
}

func (m *MockHost) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	args := m.Called(ctx, p, pids)
	return args.Get(0).(network.Stream), args.Error(1)
}

func (m *MockHost) Close() error {
	args := m.Called()
	return args.Error(0)
}

//
// DHT Service Mocks
//

// MockDHTService is a mock implementation of the api.DHTService interface
type MockDHTService struct {
	mock.Mock
}

func (m *MockDHTService) GetValue(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockDHTService) PutValue(ctx context.Context, key string, value []byte) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

//
// Host Service Mocks
//

// MockHostService is a mock implementation of the api.HostService interface
type MockHostService struct {
	mock.Mock
}

func (m *MockHostService) ID() peer.ID {
	args := m.Called()
	return args.Get(0).(peer.ID)
}

func (m *MockHostService) Peerstore() api.PeerstoreService {
	args := m.Called()
	return args.Get(0).(api.PeerstoreService)
}

// MockPeerstoreService is a mock implementation of the api.PeerstoreService interface
type MockPeerstoreService struct {
	mock.Mock
}

func (m *MockPeerstoreService) PrivKey(p peer.ID) crypto.PrivKey {
	args := m.Called(p)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(crypto.PrivKey)
}
