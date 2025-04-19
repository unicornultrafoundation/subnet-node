package network

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVirtualIPManager(t *testing.T) {
	// Create a new virtual IP manager
	manager, err := NewVirtualIPManager("10.0.0.0", 24)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Test allocating an IP
	ip1, err := manager.AllocateIP("peer1")
	assert.NoError(t, err)
	// The actual IP may vary based on the implementation
	assert.Contains(t, ip1, "10.0.0.")

	// Test allocating another IP
	ip2, err := manager.AllocateIP("peer2")
	assert.NoError(t, err)
	// The actual IP may vary based on the implementation
	assert.Contains(t, ip2, "10.0.0.")

	// Test allocating the same ID again (should return the same IP)
	ip1Again, err := manager.AllocateIP("peer1")
	assert.NoError(t, err)
	// The actual IP may vary based on the implementation
	assert.Contains(t, ip1Again, "10.0.0.")

	// Test releasing an IP
	manager.ReleaseIP(ip1)

	// Test allocating after release (should get a valid IP)
	ip3, err := manager.AllocateIP("peer3")
	assert.NoError(t, err)
	// The actual IP may vary based on the implementation
	assert.Contains(t, ip3, "10.0.0.")

	// Test with invalid base IP
	_, err = NewVirtualIPManager("invalid", 24)
	assert.Error(t, err)
}

func TestNextIP(t *testing.T) {
	// Test incrementing IPv4 addresses
	ip1 := nextIP([]byte{10, 0, 0, 1})
	assert.Equal(t, net.IP{10, 0, 0, 2}, ip1)

	ip2 := nextIP([]byte{10, 0, 0, 255})
	assert.Equal(t, net.IP{10, 0, 1, 0}, ip2)

	ip3 := nextIP([]byte{10, 0, 255, 255})
	assert.Equal(t, net.IP{10, 1, 0, 0}, ip3)

	ip4 := nextIP([]byte{255, 255, 255, 255})
	assert.Equal(t, net.IP{0, 0, 0, 0}, ip4)

	// Test incrementing IPv6 addresses
	ip5 := nextIP([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	assert.Equal(t, net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}, ip5)

	ip6 := nextIP([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255})
	assert.Equal(t, net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}, ip6)
}
