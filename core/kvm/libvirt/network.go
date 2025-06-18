//go:build libvirt
// +build libvirt

package libvirt

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

// NetworkManager handles network operations
type NetworkManager struct {
	client *Client
	logger *logrus.Entry
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(client *Client, logger *logrus.Entry) *NetworkManager {
	return &NetworkManager{
		client: client,
		logger: logger.WithField("component", "network-manager"),
	}
}

// NetworkXML represents libvirt network XML
type NetworkXML struct {
	XMLName xml.Name           `xml:"network"`
	Name    string             `xml:"name"`
	Bridge  NetworkBridgeXML   `xml:"bridge"`
	Forward *NetworkForwardXML `xml:"forward,omitempty"`
	IP      *NetworkIPXML      `xml:"ip,omitempty"`
}

type NetworkBridgeXML struct {
	Name string `xml:"name,attr"`
	STP  string `xml:"stp,attr"`
}

type NetworkForwardXML struct {
	Mode string `xml:"mode,attr"`
}

type NetworkIPXML struct {
	Address string          `xml:"address,attr"`
	Netmask string          `xml:"netmask,attr"`
	DHCP    *NetworkDHCPXML `xml:"dhcp,omitempty"`
}

type NetworkDHCPXML struct {
	Range NetworkRangeXML `xml:"range"`
}

type NetworkRangeXML struct {
	Start string `xml:"start,attr"`
	End   string `xml:"end,attr"`
}

// CreateNATNetwork creates a NAT network
func (nm *NetworkManager) CreateNATNetwork(name, bridgeName, subnet string) (*libvirt.Network, error) {
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %w", err)
	}

	// Calculate network parameters
	netmask := net.IP(ipNet.Mask).String()

	// DHCP range: give the first 100 IPs for DHCP
	ip4 := ip.To4()
	if ip4 == nil {
		return nil, fmt.Errorf("only IPv4 subnets are supported")
	}

	startIP := net.IPv4(ip4[0], ip4[1], ip4[2], ip4[3]+2) // Start from .2
	endIP := net.IPv4(ip4[0], ip4[1], ip4[2], ip4[3]+100) // End at .100

	networkXML := NetworkXML{
		Name: name,
		Bridge: NetworkBridgeXML{
			Name: bridgeName,
			STP:  "on",
		},
		Forward: &NetworkForwardXML{
			Mode: "nat",
		},
		IP: &NetworkIPXML{
			Address: ip.String(),
			Netmask: netmask,
			DHCP: &NetworkDHCPXML{
				Range: NetworkRangeXML{
					Start: startIP.String(),
					End:   endIP.String(),
				},
			},
		},
	}

	xmlData, err := xml.MarshalIndent(networkXML, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal network XML: %w", err)
	}

	xmlString := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(xmlData)

	network, err := nm.client.CreateNetwork(xmlString)
	if err != nil {
		return nil, err
	}

	nm.logger.WithFields(logrus.Fields{
		"network": name,
		"bridge":  bridgeName,
		"subnet":  subnet,
	}).Info("NAT network created")

	return network, nil
}

// CreateBridgeNetwork creates a bridge network
func (nm *NetworkManager) CreateBridgeNetwork(name, bridgeName string) (*libvirt.Network, error) {
	networkXML := NetworkXML{
		Name: name,
		Bridge: NetworkBridgeXML{
			Name: bridgeName,
			STP:  "on",
		},
		Forward: &NetworkForwardXML{
			Mode: "bridge",
		},
	}

	xmlData, err := xml.MarshalIndent(networkXML, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal network XML: %w", err)
	}

	xmlString := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(xmlData)

	network, err := nm.client.CreateNetwork(xmlString)
	if err != nil {
		return nil, err
	}

	nm.logger.WithFields(logrus.Fields{
		"network": name,
		"bridge":  bridgeName,
		"type":    "bridge",
	}).Info("Bridge network created")

	return network, nil
}

// GetOrCreateNetwork gets existing network or creates a default NAT network
func (nm *NetworkManager) GetOrCreateNetwork(name string) (*libvirt.Network, error) {
	// Try to get existing network
	network, err := nm.client.GetNetworkByName(name)
	if err == nil {
		// Network exists, make sure it's active
		active, err := network.IsActive()
		if err != nil {
			return nil, fmt.Errorf("failed to check network status: %w", err)
		}
		if !active {
			if err := network.Create(); err != nil {
				return nil, fmt.Errorf("failed to activate network: %w", err)
			}
		}
		return network, nil
	}

	// Network doesn't exist, create a default NAT network
	bridgeName := fmt.Sprintf("virbr-%s", name)
	subnet := nm.generateSubnet()

	return nm.CreateNATNetwork(name, bridgeName, subnet)
}

// DeleteNetwork removes a network
func (nm *NetworkManager) DeleteNetwork(name string) error {
	network, err := nm.client.GetNetworkByName(name)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}

	// Stop the network if it's running
	active, err := network.IsActive()
	if err != nil {
		return fmt.Errorf("failed to check network status: %w", err)
	}
	if active {
		if err := network.Destroy(); err != nil {
			return fmt.Errorf("failed to stop network: %w", err)
		}
	}

	// Undefine the network
	if err := network.Undefine(); err != nil {
		return fmt.Errorf("failed to undefine network: %w", err)
	}

	nm.logger.WithField("network", name).Info("Network deleted")
	return nil
}

// GenerateMAC generates a random MAC address
func (nm *NetworkManager) GenerateMAC() string {
	// Use a locally administered unicast MAC address
	// First byte: 52 (01010010) - bit 1 set (locally administered), bit 0 clear (unicast)
	mac := make([]byte, 6)
	mac[0] = 0x52

	// Fill remaining bytes with random data
	rand.Seed(time.Now().UnixNano())
	for i := 1; i < 6; i++ {
		mac[i] = byte(rand.Intn(256))
	}

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}

// generateSubnet generates a random private subnet
func (nm *NetworkManager) generateSubnet() string {
	// Use 192.168.x.0/24 subnets
	rand.Seed(time.Now().UnixNano())
	subnet := rand.Intn(254) + 1 // 1-254
	return fmt.Sprintf("192.168.%d.1/24", subnet)
}

// GetNetworkInfo returns information about a network
func (nm *NetworkManager) GetNetworkInfo(name string) (map[string]interface{}, error) {
	network, err := nm.client.GetNetworkByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %w", err)
	}

	active, err := network.IsActive()
	if err != nil {
		return nil, fmt.Errorf("failed to check network status: %w", err)
	}

	persistent, err := network.IsPersistent()
	if err != nil {
		return nil, fmt.Errorf("failed to check network persistence: %w", err)
	}

	xmlDesc, err := network.GetXMLDesc(0)
	if err != nil {
		return nil, fmt.Errorf("failed to get network XML: %w", err)
	}

	var networkXML NetworkXML
	if err := xml.Unmarshal([]byte(xmlDesc), &networkXML); err != nil {
		return nil, fmt.Errorf("failed to parse network XML: %w", err)
	}

	info := map[string]interface{}{
		"name":       networkXML.Name,
		"bridge":     networkXML.Bridge.Name,
		"active":     active,
		"persistent": persistent,
	}

	if networkXML.Forward != nil {
		info["forward_mode"] = networkXML.Forward.Mode
	}

	if networkXML.IP != nil {
		info["ip_address"] = networkXML.IP.Address
		info["netmask"] = networkXML.IP.Netmask

		if networkXML.IP.DHCP != nil {
			info["dhcp_start"] = networkXML.IP.DHCP.Range.Start
			info["dhcp_end"] = networkXML.IP.DHCP.Range.End
		}
	}

	return info, nil
}

// ListNetworks returns all networks
func (nm *NetworkManager) ListNetworks() ([]string, error) {
	conn, err := nm.client.GetConnection()
	if err != nil {
		return nil, err
	}

	networks, err := conn.ListAllNetworks(0)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	var networkNames []string
	for _, network := range networks {
		name, err := network.GetName()
		if err != nil {
			continue // Skip networks we can't get names for
		}
		networkNames = append(networkNames, name)
		network.Free()
	}

	return networkNames, nil
}

// EnsureDefaultNetwork ensures the default network exists
func (nm *NetworkManager) EnsureDefaultNetwork(networkName string) error {
	_, err := nm.GetOrCreateNetwork(networkName)
	if err != nil {
		return fmt.Errorf("failed to ensure default network: %w", err)
	}

	nm.logger.WithField("network", networkName).Info("Default network ensured")
	return nil
}

// GetBridgeIP returns the IP address of a bridge interface
func (nm *NetworkManager) GetBridgeIP(bridgeName string) (string, error) {
	iface, err := net.InterfaceByName(bridgeName)
	if err != nil {
		return "", fmt.Errorf("failed to get interface %s: %w", bridgeName, err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get interface addresses: %w", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no IPv4 address found for bridge %s", bridgeName)
}
