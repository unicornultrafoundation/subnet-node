package discovery

import (
	"context"
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/testutil"
	subnet_vpn "github.com/unicornultrafoundation/subnet-node/proto/subnet/vpn"
)

// TestDHTStorage tests both GetVirtualIP and StoreMappingInDHT methods
func TestDHTStorage(t *testing.T) {
	// Create mocks
	mockPeerstoreService := new(testutil.MockPeerstoreService)
	mockHostService := new(testutil.MockHostService)
	mockDHTService := new(testutil.MockDHTService)
	mockPeerID := peer.ID("peer1")

	// Generate a test key pair
	privKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	assert.NoError(t, err)

	// Create a test record
	info := &subnet_vpn.VirtualIPInfo{
		Ip:        "10.1.2.3",
		Timestamp: 123456789,
	}
	record := &subnet_vpn.VirtualIPRecord{
		Info:      info,
		Signature: []byte("test-signature"),
	}
	recordData, _ := proto.Marshal(record)

	// Setup expectations for storing
	mockHostService.On("ID").Return(mockPeerID)
	mockHostService.On("Peerstore").Return(mockPeerstoreService)
	mockPeerstoreService.On("PrivKey", mockPeerID).Return(privKey)
	mockDHTService.On("PutValue", mock.Anything, "/vpn/mapping/peer1", mock.Anything).Return(nil)

	// Setup expectations for retrieving
	mockDHTService.On("GetValue", mock.Anything, "/vpn/mapping/peer1").Return(recordData, nil)

	// Create the peer discovery service
	peerDiscovery := &PeerDiscovery{
		host:      mockHostService,
		dht:       mockDHTService,
		virtualIP: "10.1.2.3",
	}

	// Test storing a mapping
	ctx := context.Background()
	err = peerDiscovery.StoreMappingInDHT(ctx, "peer1")

	// Verify results
	assert.NoError(t, err)

	// Test retrieving the mapping
	virtualIP, err := peerDiscovery.GetVirtualIP(ctx, "peer1")

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, "10.1.2.3", virtualIP)

	// Test error cases

	// Setup expectations for DHT error
	mockDHTService.On("GetValue", mock.Anything, "/vpn/mapping/peer2").Return(nil, fmt.Errorf("DHT error"))

	// Test getting a virtual IP with an error
	virtualIP, err = peerDiscovery.GetVirtualIP(ctx, "peer2")

	// Verify results
	assert.Error(t, err)
	assert.Equal(t, "", virtualIP)
	assert.Contains(t, err.Error(), "DHT error")

	// Test storing with no virtual IP
	peerDiscovery.virtualIP = ""
	err = peerDiscovery.StoreMappingInDHT(ctx, "peer1")

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "virtual IP is not set")

	// Verify all expectations were met
	mockHostService.AssertExpectations(t)
	mockPeerstoreService.AssertExpectations(t)
	mockDHTService.AssertExpectations(t)
}
