package discovery

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/testutil"
)

// TestGetPeerID tests the GetPeerID method with caching behavior
func TestGetPeerID(t *testing.T) {
	// Create mocks
	mockAccountService := new(testutil.MockAccountService)
	mockIPRegistry := new(testutil.MockIPRegistry)
	mockHostService := new(testutil.MockHostService)
	mockPeerstoreService := new(testutil.MockPeerstoreService)
	mockDHTService := new(testutil.MockDHTService)

	// Setup expectations
	mockHostService.On("ID").Return(peer.ID("peer1"))
	mockHostService.On("Peerstore").Return(mockPeerstoreService)
	mockAccountService.On("IPRegistry").Return(mockIPRegistry)

	// Setup GetPeer to return different values based on the tokenID
	mockIPRegistry.On("GetPeer", mock.Anything, mock.MatchedBy(func(tokenID *big.Int) bool {
		return tokenID != nil && tokenID.Int64() == int64(ConvertVirtualIPToNumber("10.0.0.2"))
	})).Return("peer1", nil)

	mockIPRegistry.On("GetPeer", mock.Anything, mock.MatchedBy(func(tokenID *big.Int) bool {
		return tokenID != nil && tokenID.Int64() == int64(ConvertVirtualIPToNumber("10.0.0.3"))
	})).Return("", fmt.Errorf("registry error"))

	// Create the peer discovery service
	peerDiscovery := &PeerDiscovery{
		host:           mockHostService,
		dht:            mockDHTService,
		peerIDCache:    cache.New(1*time.Minute, 30*time.Second),
		virtualIP:      "10.0.0.1",
		accountService: mockAccountService,
		mu:             sync.RWMutex{},
	}

	// Test getting a peer ID that's not in the cache
	ctx := context.Background()
	peerID, err := peerDiscovery.GetPeerID(ctx, "10.0.0.2")

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, "peer1", peerID)

	// Test getting a peer ID that's now in the cache
	peerID, err = peerDiscovery.GetPeerID(ctx, "10.0.0.2")

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, "peer1", peerID)
	// The mock should not be called again since we're using the cache
	mockIPRegistry.AssertNumberOfCalls(t, "GetPeer", 1)

	// Test error case
	peerID, err = peerDiscovery.GetPeerID(ctx, "10.0.0.3")

	// Verify results
	assert.Error(t, err)
	assert.Equal(t, "", peerID)
	assert.Contains(t, err.Error(), "failed to get peer ID from registry")

	// Verify only the registry expectations
	mockAccountService.AssertExpectations(t)
	mockIPRegistry.AssertExpectations(t)
	// Note: We don't verify mockHostService expectations as they're not used in GetPeerID
}
