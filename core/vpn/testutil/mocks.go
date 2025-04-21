package testutil

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

// MockStreamConfig contains configuration for mock streams
type MockStreamConfig struct {
	Latency     time.Duration
	Jitter      time.Duration
	PacketLoss  float64
	FailureRate float64
}

// MockStream is a configurable mock implementation of types.VPNStream
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

// MockServiceConfig contains configuration for mock services
type MockServiceConfig struct {
	Latency     time.Duration
	Jitter      time.Duration
	FailureRate float64
	PacketLoss  float64
}

// MockStreamService is a configurable mock implementation of types.Service
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

// CreateNewVPNStream implements the types.Service interface
func (s *MockStreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
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
	return args.Get(0).(types.VPNStream), args.Error(1)
}

// GetStats returns the service statistics
func (s *MockStreamService) GetStats() map[string]int64 {
	return map[string]int64{
		"streams_created": atomic.LoadInt64(&s.Stats.StreamsCreated),
		"stream_errors":   atomic.LoadInt64(&s.Stats.StreamErrors),
	}
}

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

// MockPoolService is a configurable mock implementation of types.PoolService
type MockPoolService struct {
	mock.Mock
	Config MockServiceConfig
	Stats  struct {
		StreamsAcquired   int64
		StreamsReleased   int64
		AcquisitionErrors int64
	}
}

// NewMockPoolService creates a new mock pool service with the given configuration
func NewMockPoolService(config *MockServiceConfig) *MockPoolService {
	if config == nil {
		config = &MockServiceConfig{}
	}
	return &MockPoolService{
		Config: *config,
	}
}

// GetStream implements the types.PoolService interface
func (s *MockPoolService) GetStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
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
		atomic.AddInt64(&s.Stats.AcquisitionErrors, 1)
		return nil, fmt.Errorf("simulated failure")
	}

	atomic.AddInt64(&s.Stats.StreamsAcquired, 1)
	args := s.Called(ctx, peerID)
	return args.Get(0).(types.VPNStream), args.Error(1)
}

// ReleaseStream implements the types.PoolService interface
func (s *MockPoolService) ReleaseStream(peerID peer.ID, stream types.VPNStream, healthy bool) {
	atomic.AddInt64(&s.Stats.StreamsReleased, 1)
	s.Called(peerID, stream, healthy)
}

// GetStats returns the service statistics
func (s *MockPoolService) GetStats() map[string]int64 {
	return map[string]int64{
		"streams_acquired":   atomic.LoadInt64(&s.Stats.StreamsAcquired),
		"streams_released":   atomic.LoadInt64(&s.Stats.StreamsReleased),
		"acquisition_errors": atomic.LoadInt64(&s.Stats.AcquisitionErrors),
	}
}

// MockPoolManager is a mock implementation of the pool manager
type MockPoolManager struct {
	mock.Mock
	MinStreamsPerPeer int
}

func (m *MockPoolManager) GetStream(ctx context.Context, peerID peer.ID) (types.VPNStream, error) {
	args := m.Called(ctx, peerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(types.VPNStream), args.Error(1)
}

func (m *MockPoolManager) ReleaseStream(peerID peer.ID, s types.VPNStream, healthy bool) {
	m.Called(peerID, s, healthy)
}

func (m *MockPoolManager) GetStreamPoolMetrics() map[string]int64 {
	args := m.Called()
	return args.Get(0).(map[string]int64)
}

func (m *MockPoolManager) GetHealthMetrics() map[string]map[string]int64 {
	args := m.Called()
	return args.Get(0).(map[string]map[string]int64)
}

func (m *MockPoolManager) GetStreamCount(peerID peer.ID) int {
	args := m.Called(peerID)
	return args.Int(0)
}

func (m *MockPoolManager) GetActiveStreamCount(peerID peer.ID) int {
	args := m.Called(peerID)
	return args.Int(0)
}

func (m *MockPoolManager) GetMinStreamsPerPeer() int {
	return m.MinStreamsPerPeer
}

func (m *MockPoolManager) GetAllPeers() []peer.ID {
	args := m.Called()
	return args.Get(0).([]peer.ID)
}

func (m *MockPoolManager) Start() {
	m.Called()
}

func (m *MockPoolManager) Stop() {
	m.Called()
}
