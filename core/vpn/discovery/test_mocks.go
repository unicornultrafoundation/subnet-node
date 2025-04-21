package discovery

import (
	"context"
	"math/big"
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
