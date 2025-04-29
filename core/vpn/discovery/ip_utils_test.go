package discovery

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/testutil"
)

// TestConvertVirtualIPToNumber tests the ConvertVirtualIPToNumber function
func TestConvertVirtualIPToNumber(t *testing.T) {
	tests := []struct {
		name      string
		virtualIP string
		expected  uint32
	}{
		{
			name:      "Valid IP in 10.0.0.0/8 range",
			virtualIP: "10.0.0.1",
			expected:  167772161,
		},
		{
			name:      "IP not in 10.0.0.0/8 range",
			virtualIP: "192.168.1.1",
			expected:  0,
		},
		{
			name:      "Invalid IP",
			virtualIP: "not an IP",
			expected:  0,
		},
		{
			name:      "IPv6 address",
			virtualIP: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected:  0,
		},
		{
			name:      "Empty string",
			virtualIP: "",
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertVirtualIPToNumber(tt.virtualIP)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetPeerIDByRegistry tests the GetPeerIDByRegistry method with both success and error cases
func TestGetPeerIDByRegistry(t *testing.T) {
	// Create mocks
	mockAccountService := new(testutil.MockAccountService)
	mockIPRegistry := new(testutil.MockIPRegistry)
	mockHostService := new(testutil.MockHostService)
	mockDHTService := new(testutil.MockDHTService)

	// Setup expectations for success case
	mockAccountService.On("IPRegistry").Return(mockIPRegistry)
	// The IP 10.1.2.3 converts to 167838211 in decimal (0x0A010203 in hex)
	mockIPRegistry.On("GetPeer", mock.Anything, big.NewInt(167838211)).Return("peer1", nil)

	// Create the peer discovery service
	peerDiscovery := &PeerDiscovery{
		host:           mockHostService,
		dht:            mockDHTService,
		peerIDCache:    nil, // Not needed for this test
		virtualIP:      "10.0.0.1",
		accountService: mockAccountService,
	}

	// Test getting a peer ID from the registry - success case
	ctx := context.Background()
	peerID, err := peerDiscovery.GetPeerIDByRegistry(ctx, "10.1.2.3")

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, "peer1", peerID)

	// Setup expectations for error case
	// The IP 10.2.3.4 converts to 167904004 in decimal (0x0A020304 in hex)
	mockIPRegistry.On("GetPeer", mock.Anything, big.NewInt(167904004)).Return("", fmt.Errorf("registry error"))

	// Test getting a peer ID from the registry with an error
	peerID, err = peerDiscovery.GetPeerIDByRegistry(ctx, "10.2.3.4")

	// Verify results
	assert.Error(t, err)
	assert.Equal(t, "", peerID)
	assert.Contains(t, err.Error(), "registry error")

	// Test with an invalid IP
	peerID, err = peerDiscovery.GetPeerIDByRegistry(ctx, "192.168.1.1")

	// Verify results
	assert.Error(t, err)
	assert.Equal(t, "", peerID)
	assert.Contains(t, err.Error(), "not within the range")
}
