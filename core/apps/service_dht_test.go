package apps

import (
	"context"
	"fmt"
	"math/big"
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

	// Create a test host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)

	// Create a test config
	l := test.NewLogger()
	cfg := config.NewC(l)
	cfg.Settings = map[interface{}]interface{}{
		"routing": map[interface{}]interface{}{
			"type": "dht",
		},
		"experimental": map[interface{}]interface{}{
			"optimistic_provide":                true,
			"optimistic_provide_jobs_pool_size": 10,
			"loopback_addresses_on_lan_dhl":     false,
		},
	}

	// Create a test datastore
	datastore := dsync.MutexWrap(ds.NewMapDatastore())

	// Create real P2P instance
	p2pInstance := p2p.New(h.ID(), h, h.Peerstore())

	// Setup mocks
	mockAcc := &account.AccountService{}
	mockDocker := &docker.Service{}

	// Create the service with real P2P host
	service := New(h, h.ID(), cfg, p2pInstance, datastore, mockAcc, mockDocker)
	require.NotNil(t, service)

	return service, h, ctx
}

func TestDHTBasicOperations(t *testing.T) {
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

func TestDHTResilience(t *testing.T) {
	service1, _, ctx := setupTestService(t)
	service2, host2, _ := setupTestService(t)
	service3, host3, _ := setupTestService(t)

	// Start all services
	for _, s := range []*Service{service1, service2, service3} {
		err := s.Start(ctx)
		require.NoError(t, err)
		defer s.Stop(ctx)
	}

	// Connect the hosts in a chain: 1 <-> 2 <-> 3
	connectHosts(t, ctx, service1.PeerHost, host2)
	connectHosts(t, ctx, host2, host3)

	testApp := &atypes.App{
		ID:      big.NewInt(2),
		Name:    "resilience-test-app",
		PeerIds: []string{service1.PeerHost.ID().String()},
	}

	t.Run("DHT Propagation Through Network", func(t *testing.T) {
		// Store data from service1
		err := service1.StorePeerIDsInDHT(ctx, testApp)
		require.NoError(t, err)

		// Wait for DHT propagation
		time.Sleep(time.Second)

		// Verify service3 can find service1 as provider
		providers, err := service3.FindAppProviders(ctx, testApp)
		require.NoError(t, err)

		found := false
		for _, p := range providers {
			if p.ID == service1.PeerHost.ID() {
				found = true
				break
			}
		}
		assert.True(t, found, "Service1 should be discoverable through DHT")
	})

	t.Run("DHT Recovery After Node Failure", func(t *testing.T) {
		// Store data in DHT
		err := service1.StorePeerIDsInDHT(ctx, testApp)
		require.NoError(t, err)

		time.Sleep(time.Second)

		// Stop service2 (middle node)
		service2.Stop(ctx)

		// Service3 should still be able to find service1
		providers, err := service3.FindAppProviders(ctx, testApp)
		require.NoError(t, err)

		found := false
		for _, p := range providers {
			if p.ID == service1.PeerHost.ID() {
				found = true
				break
			}
		}
		assert.True(t, found, "Service1 should still be discoverable after service2 failure")
	})
}

func TestDHTConcurrency(t *testing.T) {
	service1, _, ctx := setupTestService(t)
	service2, host2, _ := setupTestService(t)

	// Start services
	err := service1.Start(ctx)
	require.NoError(t, err)
	defer service1.Stop(ctx)

	err = service2.Start(ctx)
	require.NoError(t, err)
	defer service2.Stop(ctx)

	// Connect the hosts
	connectHosts(t, ctx, service1.PeerHost, host2)

	t.Run("Concurrent DHT Operations", func(t *testing.T) {
		numOperations := 10
		done := make(chan bool)

		for i := 0; i < numOperations; i++ {
			go func(appID int64) {
				testApp := &atypes.App{
					ID:      big.NewInt(appID),
					Name:    fmt.Sprintf("concurrent-test-app-%d", appID),
					PeerIds: []string{service1.PeerHost.ID().String()},
				}

				err := service1.StorePeerIDsInDHT(ctx, testApp)
				assert.NoError(t, err)

				done <- true
			}(int64(i))
		}

		// Wait for all operations to complete
		for i := 0; i < numOperations; i++ {
			<-done
		}

		// Wait for DHT propagation
		time.Sleep(2 * time.Second)

		// Verify all apps can be discovered
		for i := 0; i < numOperations; i++ {
			testApp := &atypes.App{
				ID:   big.NewInt(int64(i)),
				Name: fmt.Sprintf("concurrent-test-app-%d", i),
			}

			providers, err := service2.FindAppProviders(ctx, testApp)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(providers), 1)
		}
	})
}

// Setup function specifically for benchmarks
func setupBenchService(b *testing.B) (*Service, host.Host, context.Context) {
	ctx := context.Background()

	// Create a test host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(b, err)

	// Create a test config
	l := test.NewLogger()
	cfg := config.NewC(l)
	cfg.Settings = map[interface{}]interface{}{
		"routing": map[interface{}]interface{}{
			"type": "dht",
		},
		"experimental": map[interface{}]interface{}{
			"optimistic_provide":                true,
			"optimistic_provide_jobs_pool_size": 10,
			"loopback_addresses_on_lan_dhl":     false,
		},
	}

	// Create a test datastore
	datastore := dsync.MutexWrap(ds.NewMapDatastore())

	// Create real P2P instance
	p2pInstance := p2p.New(h.ID(), h, h.Peerstore())

	// Setup mocks
	mockAcc := &account.AccountService{}
	mockDocker := &docker.Service{}

	// Create the service with real P2P host
	service := New(h, h.ID(), cfg, p2pInstance, datastore, mockAcc, mockDocker)
	require.NotNil(b, service)

	return service, h, ctx
}

func BenchmarkDHTOperations(b *testing.B) {
	service1, _, ctx := setupBenchService(b)
	service2, host2, _ := setupBenchService(b)

	err := service1.Start(ctx)
	require.NoError(b, err)
	defer service1.Stop(ctx)

	err = service2.Start(ctx)
	require.NoError(b, err)
	defer service2.Stop(ctx)

	connectHosts(b, ctx, service1.PeerHost, host2)

	b.Run("Store and Find Providers", func(b *testing.B) {
		testApp := &atypes.App{
			ID:      big.NewInt(1),
			Name:    "bench-test-app",
			PeerIds: []string{service1.PeerHost.ID().String()},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := service1.StorePeerIDsInDHT(ctx, testApp)
			require.NoError(b, err)

			_, err = service2.FindAppProviders(ctx, testApp)
			require.NoError(b, err)
		}
	})
}

// Helper function to connect two hosts
func connectHosts(tb testing.TB, ctx context.Context, h1, h2 host.Host) {
	err := h1.Connect(ctx, peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	})
	require.NoError(tb, err)
}
