package apps

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	ds "github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/core/docker"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	"github.com/unicornultrafoundation/subnet-node/test"
)

func setupTestService(t *testing.T) (*Service, host.Host, context.Context) {
	ctx := context.Background()

	// Create two bootstrap hosts
	bootstrapHost1, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)

	bootstrapHost2, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)

	// Create a test host (the service host)
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)

	// Create a test config with bootstrap peers
	l := test.NewLogger()
	cfg := config.NewC(l)

	// Get bootstrap addresses
	bootstrapAddrs := []string{}

	// Add bootstrap host 1
	bootstrapAddrInfo1 := peer.AddrInfo{
		ID:    bootstrapHost1.ID(),
		Addrs: bootstrapHost1.Addrs(),
	}
	bootstrapMultiaddr1, err := peer.AddrInfoToP2pAddrs(&bootstrapAddrInfo1)
	require.NoError(t, err)

	for _, addr := range bootstrapMultiaddr1 {
		bootstrapAddrs = append(bootstrapAddrs, addr.String())
	}

	// Add bootstrap host 2
	bootstrapAddrInfo2 := peer.AddrInfo{
		ID:    bootstrapHost2.ID(),
		Addrs: bootstrapHost2.Addrs(),
	}
	bootstrapMultiaddr2, err := peer.AddrInfoToP2pAddrs(&bootstrapAddrInfo2)
	require.NoError(t, err)

	for _, addr := range bootstrapMultiaddr2 {
		bootstrapAddrs = append(bootstrapAddrs, addr.String())
	}

	// Build YAML config string
	configYaml := fmt.Sprintf(`
routing:
  type: dht
experimental:
  optimistic_provide: true
  optimistic_provide_jobs_pool_size: 10
  loopback_addresses_on_lan_dhl: false
bootstrap:
%s
`, formatBootstrapAddrsForYaml(bootstrapAddrs))

	// Load the config from the YAML string
	err = cfg.LoadString(configYaml)
	require.NoError(t, err)

	// Create a test datastore
	datastore := dsync.MutexWrap(ds.NewMapDatastore())

	// Create real P2P instance
	p2pInstance := p2p.New(h.ID(), h, h.Peerstore())

	// Setup mocks
	mockAcc := &account.AccountService{}
	mockDocker := &docker.Service{}

	// Create the service with real P2P host
	service := New(h, h.ID(), cfg, p2pInstance, datastore, mockAcc, mockDocker)
	service.cfg = cfg
	require.NotNil(t, service)

	// Ensure DHT is initialized
	err = service.dht.Bootstrap(ctx)
	require.NoError(t, err)

	// Verify bootstrap configuration
	bootstrapPeers, err := parseBootstrapPeers(cfg)

	require.NoError(t, err)
	require.Equal(t, 2, len(bootstrapPeers), "Should have 2 bootstrap peers configured")

	// For test purposes, we'll manually connect to bootstrap peers
	// This is normally done in Start(), but we want to ensure it's done here
	err = service.connectToDHTBootstrap(ctx)
	require.NoError(t, err)

	// Verify that the DHT has peers
	require.Eventually(t, func() bool {
		return len(service.dht.RoutingTable().ListPeers()) > 0
	}, 5*time.Second, 100*time.Millisecond, "DHT failed to connect to bootstrap peers")

	// Verify connections to bootstrap peers
	connectedBootstrapCount := 0
	for _, conn := range h.Network().Conns() {
		remotePeer := conn.RemotePeer()
		if remotePeer == bootstrapHost1.ID() || remotePeer == bootstrapHost2.ID() {
			connectedBootstrapCount++
		}
	}
	require.Greater(t, connectedBootstrapCount, 0, "Should be connected to at least one bootstrap peer")

	return service, h, ctx
}

func TestDHTThroughBootstrap(t *testing.T) {
	service1, host1, ctx := setupTestService(t)
	service2, host2, _ := setupTestService(t)

	// Start both services
	err := service1.Start(ctx)
	require.NoError(t, err)
	defer service1.Stop(ctx)

	err = service2.Start(ctx)
	require.NoError(t, err)
	defer service2.Stop(ctx)

	// Connect the hosts
	err = host1.Connect(ctx, peer.AddrInfo{
		ID:    host2.ID(),
		Addrs: host2.Addrs(),
	})
	require.NoError(t, err)

	// Create a test app
	testApp := &atypes.App{
		ID:      big.NewInt(1),
		Name:    "test-app",
		PeerIds: []string{host1.ID().String()},
	}

	t.Run("Verify DHT Bootstrap", func(t *testing.T) {
		// Check that both services have peers in their DHT
		require.Greater(t, len(service1.dht.RoutingTable().ListPeers()), 0,
			"Service1 DHT should have peers after bootstrap")
		require.Greater(t, len(service2.dht.RoutingTable().ListPeers()), 0,
			"Service2 DHT should have peers after bootstrap")

		require.NoError(t, err, "DHT bootstrap validation should succeed")
	})

	t.Run("Store and Find App Providers", func(t *testing.T) {
		// Store peer IDs in DHT
		err := service1.StorePeerIDsInDHT(ctx, testApp)
		require.NoError(t, err)

		// Wait for DHT propagation
		time.Sleep(time.Second)

		// Find providers using service2
		providers, err := service2.FindAppProviders(ctx, testApp)

		require.NoError(t, err)
		require.GreaterOrEqual(t, len(providers), 1)

		// Verify that service1 is among the providers
		found := false
		for _, p := range providers {
			if p.ID == host1.ID() {
				found = true
				break
			}
		}
		assert.True(t, found, "Service1 should be found as a provider")
	})
}

// Helper function to format bootstrap addresses for YAML
func formatBootstrapAddrsForYaml(addrs []string) string {
	var result strings.Builder
	for _, addr := range addrs {
		result.WriteString(fmt.Sprintf("  - %s\n", addr))
	}
	return result.String()
}

func setupTestServiceWithDirectConnection(t *testing.T) (*Service, *Service, context.Context) {
	ctx := context.Background()

	// Create two hosts
	host1, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)

	host2, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)

	// Create minimal config without bootstrap
	l := test.NewLogger()
	cfg := config.NewC(l)

	// Load minimal config
	err = cfg.LoadString(`
routing:
  type: dht
experimental:
  optimistic_provide: true
`)
	require.NoError(t, err)

	// Create datastores
	datastore1 := dsync.MutexWrap(ds.NewMapDatastore())
	datastore2 := dsync.MutexWrap(ds.NewMapDatastore())

	// Create P2P instances
	p2pInstance1 := p2p.New(host1.ID(), host1, host1.Peerstore())
	p2pInstance2 := p2p.New(host2.ID(), host2, host2.Peerstore())

	// Setup mocks
	mockAcc := &account.AccountService{}
	mockDocker := &docker.Service{}

	// Create services
	service1 := New(host1, host1.ID(), cfg, p2pInstance1, datastore1, mockAcc, mockDocker)
	service2 := New(host2, host2.ID(), cfg, p2pInstance2, datastore2, mockAcc, mockDocker)

	// Initialize DHT
	err = service1.dht.Bootstrap(ctx)
	require.NoError(t, err)

	err = service2.dht.Bootstrap(ctx)
	require.NoError(t, err)

	// Connect the hosts directly
	host2Info := peer.AddrInfo{
		ID:    host2.ID(),
		Addrs: host2.Addrs(),
	}

	err = host1.Connect(ctx, host2Info)
	require.NoError(t, err)

	// Add peers to each other's DHT routing tables
	service1.dht.RoutingTable().TryAddPeer(host2.ID(), true, true)
	service2.dht.RoutingTable().TryAddPeer(host1.ID(), true, true)

	// Verify connection
	require.Eventually(t, func() bool {
		return len(host1.Network().ConnsToPeer(host2.ID())) > 0
	}, 5*time.Second, 100*time.Millisecond, "Failed to connect hosts")

	// Verify DHT routing tables have the peers
	require.Eventually(t, func() bool {
		peers1 := service1.dht.RoutingTable().ListPeers()
		for _, p := range peers1 {
			if p == host2.ID() {
				return true
			}
		}
		return false
	}, 5*time.Second, 100*time.Millisecond, "Peer not found in service1's DHT routing table")

	require.Eventually(t, func() bool {
		peers2 := service2.dht.RoutingTable().ListPeers()
		for _, p := range peers2 {
			if p == host1.ID() {
				return true
			}
		}
		return false
	}, 5*time.Second, 100*time.Millisecond, "Peer not found in service2's DHT routing table")

	return service1, service2, ctx
}

func TestDHTDirectConnection(t *testing.T) {
	service1, service2, ctx := setupTestServiceWithDirectConnection(t)

	// Create a test app
	testApp := &atypes.App{
		ID:      big.NewInt(1),
		Name:    "direct-test-app",
		PeerIds: []string{service1.peerId.String()},
	}

	// Store peer IDs in DHT
	err := service1.StorePeerIDsInDHT(ctx, testApp)
	require.NoError(t, err)

	// Wait for DHT propagation
	time.Sleep(time.Second)

	// Find providers using service2
	providers, err := service2.FindAppProviders(ctx, testApp)
	require.NoError(t, err)

	// Verify that service1 is among the providers
	found := false
	for _, p := range providers {
		if p.ID == service1.peerId {
			found = true
			break
		}
	}
	assert.True(t, found, "Service1 should be found as a provider")
}
