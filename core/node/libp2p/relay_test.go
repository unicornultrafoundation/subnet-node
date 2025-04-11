package libp2p

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/test"
)

func TestDockerACLFilter_AllowReserve(t *testing.T) {
	// Create a mock config with allow_all set to false
	l := test.NewLogger()
	cfg := config.NewC(l)
	cfg.Settings = map[interface{}]interface{}{
		"swarm": map[interface{}]interface{}{
			"relay_service": map[interface{}]interface{}{
				"allow_all": false,
			},
		},
	}

	// Create a DockerACLFilter with the mock config
	filter := NewDockerACLFilter(cfg)

	// Test cases
	testCases := []struct {
		name     string
		addrStr  string
		expected bool
	}{
		{
			name:     "Docker bridge network IP",
			addrStr:  "/ip4/172.17.0.5/tcp/1234",
			expected: true,
		},
		{
			name:     "Private network IP (10.x.x.x)",
			addrStr:  "/ip4/10.0.0.5/tcp/1234",
			expected: true,
		},
		{
			name:     "Private network IP (192.168.x.x)",
			addrStr:  "/ip4/192.168.1.5/tcp/1234",
			expected: true,
		},
		{
			name:     "Public IP",
			addrStr:  "/ip4/203.0.113.5/tcp/1234",
			expected: false,
		},
		{
			name:     "IPv6 loopback",
			addrStr:  "/ip6/::1/tcp/1234",
			expected: false,
		},
	}

	// Mock peer ID
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addr, err := ma.NewMultiaddr(tc.addrStr)
			assert.NoError(t, err)

			result := filter.AllowReserve(peerID, addr)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDockerACLFilter_AllowConnect(t *testing.T) {
	// Create a mock config with allow_all set to false
	l := test.NewLogger()
	cfg := config.NewC(l)
	cfg.Settings = map[interface{}]interface{}{
		"swarm": map[interface{}]interface{}{
			"relay_service": map[interface{}]interface{}{
				"allow_all": false,
			},
		},
	}

	// Create a DockerACLFilter with the mock config
	filter := NewDockerACLFilter(cfg)

	// Test cases
	testCases := []struct {
		name     string
		addrStr  string
		expected bool
	}{
		{
			name:     "Docker bridge network IP",
			addrStr:  "/ip4/172.17.0.5/tcp/1234",
			expected: true,
		},
		{
			name:     "Private network IP (10.x.x.x)",
			addrStr:  "/ip4/10.0.0.5/tcp/1234",
			expected: true,
		},
		{
			name:     "Private network IP (192.168.x.x)",
			addrStr:  "/ip4/192.168.1.5/tcp/1234",
			expected: true,
		},
		{
			name:     "Public IP",
			addrStr:  "/ip4/203.0.113.5/tcp/1234",
			expected: false,
		},
	}

	// Mock peer IDs
	srcPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	destPeerID, _ := peer.Decode("QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addr, err := ma.NewMultiaddr(tc.addrStr)
			assert.NoError(t, err)

			result := filter.AllowConnect(srcPeerID, addr, destPeerID)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDockerACLFilter_AllowAll(t *testing.T) {
	// Create a mock config with allow_all set to true
	l := test.NewLogger()
	cfg := config.NewC(l)
	cfg.Settings = map[interface{}]interface{}{
		"swarm": map[interface{}]interface{}{
			"relay_service": map[interface{}]interface{}{
				"allow_all": true,
			},
		},
	}

	// Create a DockerACLFilter with the mock config
	filter := NewDockerACLFilter(cfg)

	// Mock peer IDs
	peerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	destPeerID, _ := peer.Decode("QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ")

	// Test with a public IP (which would normally be rejected)
	addr, err := ma.NewMultiaddr("/ip4/203.0.113.5/tcp/1234")
	assert.NoError(t, err)

	// Both AllowReserve and AllowConnect should return true when allow_all is true
	assert.True(t, filter.AllowReserve(peerID, addr))
	assert.True(t, filter.AllowConnect(peerID, addr, destPeerID))
}
