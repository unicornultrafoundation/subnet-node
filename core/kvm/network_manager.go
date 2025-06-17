package kvm

import (
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

// NetworkManagerImpl implements NetworkManager interface
type NetworkManagerImpl struct {
	config        *KVMConfig
	libvirtClient LibvirtClient
	ipamManager   *IPAMManager
	logger        *logrus.Logger
}

// NewNetworkManagerImpl creates a new network manager instance
func NewNetworkManagerImpl(config *KVMConfig, libvirtClient LibvirtClient, logger *logrus.Logger) (NetworkManager, error) {
	// Initialize IPAM manager with default subnet
	_, subnet, err := net.ParseCIDR("192.168.100.0/24")
	if err != nil {
		return nil, fmt.Errorf("failed to parse default subnet: %w", err)
	}

	ipamManager := &IPAMManager{
		subnet:    *subnet,
		allocated: make(map[string]net.IP),
		mutex:     sync.RWMutex{},
	}

	return &NetworkManagerImpl{
		config:        config,
		libvirtClient: libvirtClient,
		ipamManager:   ipamManager,
		logger:        logger,
	}, nil
}

// SetupVMNetworking configures networking for a VM
func (nm *NetworkManagerImpl) SetupVMNetworking(vmID string, networkType NetworkType) (*NetworkConfig, error) {
	nm.logger.WithFields(logrus.Fields{
		"vm_id":        vmID,
		"network_type": networkType,
	}).Info("Setting up VM networking")

	// Allocate IP address
	ip, err := nm.ipamManager.AllocateIP(vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP: %w", err)
	}

	// Generate MAC address
	macAddress := nm.generateMACAddress(vmID)

	config := &NetworkConfig{
		Type:       networkType,
		IPAddress:  ip,
		MACAddress: macAddress,
		Gateway:    nm.getGatewayIP(),
		DNS:        nm.getDNSServers(),
	}

	switch networkType {
	case NetworkTypeBridge:
		config.BridgeName = "virbr0"
	case NetworkTypeNAT:
		config.BridgeName = "virbr0"
	case NetworkTypeIsolated:
		config.BridgeName = "virbr1"
	}

	return config, nil
}

// AllocateIP allocates an IP address for a VM
func (nm *NetworkManagerImpl) AllocateIP(vmID string) (net.IP, error) {
	return nm.ipamManager.AllocateIP(vmID)
}

// ReleaseIP releases an IP address from a VM
func (nm *NetworkManagerImpl) ReleaseIP(vmID string) error {
	return nm.ipamManager.ReleaseIP(vmID)
}

// GetNetworkInfo returns network information for a VM
func (nm *NetworkManagerImpl) GetNetworkInfo(vmID string) (*NetworkInfo, error) {
	ip, exists := nm.ipamManager.GetAllocatedIP(vmID)
	if !exists {
		return nil, fmt.Errorf("no IP allocated for VM %s", vmID)
	}

	return &NetworkInfo{
		Name:        "default",
		Type:        NetworkTypeNAT,
		IPAddress:   ip,
		MACAddress:  nm.generateMACAddress(vmID),
		BridgeName:  "virbr0",
		NetworkCIDR: nm.ipamManager.subnet.String(),
	}, nil
}

// ListNetworks lists available networks
func (nm *NetworkManagerImpl) ListNetworks() ([]*NetworkInfo, error) {
	networks := []*NetworkInfo{
		{
			Name:        "default",
			Type:        NetworkTypeNAT,
			BridgeName:  "virbr0",
			NetworkCIDR: "192.168.100.0/24",
		},
		{
			Name:        "bridge",
			Type:        NetworkTypeBridge,
			BridgeName:  "br0",
			NetworkCIDR: "192.168.1.0/24",
		},
	}

	return networks, nil
}

// Helper methods

func (nm *NetworkManagerImpl) generateMACAddress(vmID string) string {
	// TODO: Generate proper MAC address based on VM ID
	// This is a simple implementation for demonstration
	return fmt.Sprintf("52:54:00:12:34:%02x", len(vmID)%256)
}

func (nm *NetworkManagerImpl) getGatewayIP() net.IP {
	// TODO: Get actual gateway IP
	return net.ParseIP("192.168.100.1")
}

func (nm *NetworkManagerImpl) getDNSServers() []net.IP {
	// TODO: Get actual DNS servers
	return []net.IP{
		net.ParseIP("8.8.8.8"),
		net.ParseIP("8.8.4.4"),
	}
}

// IPAMManager handles IP address allocation
type IPAMManager struct {
	subnet    net.IPNet
	allocated map[string]net.IP
	nextIP    uint32
	mutex     sync.RWMutex
}

// AllocateIP allocates an IP address for a VM
func (ipam *IPAMManager) AllocateIP(vmID string) (net.IP, error) {
	ipam.mutex.Lock()
	defer ipam.mutex.Unlock()

	// Check if VM already has an IP allocated
	if ip, exists := ipam.allocated[vmID]; exists {
		return ip, nil
	}

	// Find next available IP
	ip, err := ipam.getNextAvailableIP()
	if err != nil {
		return nil, err
	}

	// Allocate IP to VM
	ipam.allocated[vmID] = ip
	return ip, nil
}

// ReleaseIP releases an IP address from a VM
func (ipam *IPAMManager) ReleaseIP(vmID string) error {
	ipam.mutex.Lock()
	defer ipam.mutex.Unlock()

	if _, exists := ipam.allocated[vmID]; !exists {
		return fmt.Errorf("no IP allocated for VM %s", vmID)
	}

	delete(ipam.allocated, vmID)
	return nil
}

// GetAllocatedIP returns the allocated IP for a VM
func (ipam *IPAMManager) GetAllocatedIP(vmID string) (net.IP, bool) {
	ipam.mutex.RLock()
	defer ipam.mutex.RUnlock()

	ip, exists := ipam.allocated[vmID]
	return ip, exists
}

// getNextAvailableIP finds the next available IP in the subnet
func (ipam *IPAMManager) getNextAvailableIP() (net.IP, error) {
	// Convert subnet to uint32 for easy iteration
	subnetIP := ipam.subnet.IP.To4()
	if subnetIP == nil {
		return nil, fmt.Errorf("invalid subnet IP")
	}

	subnetInt := uint32(subnetIP[0])<<24 + uint32(subnetIP[1])<<16 + uint32(subnetIP[2])<<8 + uint32(subnetIP[3])

	// Get subnet mask
	maskSize, _ := ipam.subnet.Mask.Size()
	hostBits := 32 - maskSize
	maxHosts := (1 << hostBits) - 2 // Exclude network and broadcast addresses

	// Start from IP after network address
	if ipam.nextIP == 0 {
		ipam.nextIP = subnetInt + 1
	}

	// Try to find an available IP
	for i := 0; i < maxHosts; i++ {
		testIP := make(net.IP, 4)
		testIP[0] = byte((ipam.nextIP >> 24) & 0xFF)
		testIP[1] = byte((ipam.nextIP >> 16) & 0xFF)
		testIP[2] = byte((ipam.nextIP >> 8) & 0xFF)
		testIP[3] = byte(ipam.nextIP & 0xFF)

		// Check if IP is already allocated
		allocated := false
		for _, allocatedIP := range ipam.allocated {
			if testIP.Equal(allocatedIP) {
				allocated = true
				break
			}
		}

		if !allocated {
			ipam.nextIP++
			if ipam.nextIP >= subnetInt+(1<<hostBits)-1 {
				ipam.nextIP = subnetInt + 1 // Wrap around
			}
			return testIP, nil
		}

		ipam.nextIP++
		if ipam.nextIP >= subnetInt+(1<<hostBits)-1 {
			ipam.nextIP = subnetInt + 1 // Wrap around
		}
	}

	return nil, fmt.Errorf("no available IP addresses in subnet")
}
