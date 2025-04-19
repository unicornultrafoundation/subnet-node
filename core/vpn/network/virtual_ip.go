package network

import (
	"fmt"
	"net"
	"sync"
)

// VirtualIPManager manages virtual IP addresses
type VirtualIPManager struct {
	// Base IP for the VPN network
	baseIP net.IP
	// Subnet mask length
	subnet int
	// Map of allocated IPs
	allocatedIPs map[string]bool
	// Mutex for thread safety
	mu sync.Mutex
}

// NewVirtualIPManager creates a new virtual IP manager
func NewVirtualIPManager(baseIP string, subnet int) (*VirtualIPManager, error) {
	ip := net.ParseIP(baseIP)
	if ip == nil {
		return nil, fmt.Errorf("invalid base IP: %s", baseIP)
	}

	return &VirtualIPManager{
		baseIP:       ip,
		subnet:       subnet,
		allocatedIPs: make(map[string]bool),
	}, nil
}

// AllocateIP allocates a new virtual IP
func (m *VirtualIPManager) AllocateIP(id string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if the ID already has an IP
	for ip, allocated := range m.allocatedIPs {
		if allocated && ip == id {
			return ip, nil
		}
	}

	// Find an available IP
	ip, err := m.findAvailableIP()
	if err != nil {
		return "", err
	}

	// Allocate the IP
	m.allocatedIPs[ip] = true

	return ip, nil
}

// ReleaseIP releases a virtual IP
func (m *VirtualIPManager) ReleaseIP(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.allocatedIPs, ip)
}

// findAvailableIP finds an available IP in the subnet
func (m *VirtualIPManager) findAvailableIP() (string, error) {
	// Get the network
	_, network, err := net.ParseCIDR(fmt.Sprintf("%s/%d", m.baseIP.String(), m.subnet))
	if err != nil {
		return "", err
	}

	// Find an available IP
	for ip := nextIP(network.IP); network.Contains(ip); ip = nextIP(ip) {
		ipStr := ip.String()
		if !m.allocatedIPs[ipStr] {
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("no available IPs in the subnet")
}

// nextIP returns the next IP in the subnet
func nextIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)

	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] > 0 {
			break
		}
	}

	return next
}
