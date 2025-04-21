package network

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

var log = logrus.WithField("service", "vpn-network")

// TUNConfig contains configuration for the TUN interface
type TUNConfig struct {
	// MTU for the TUN interface
	MTU int
	// Virtual IP for this node
	VirtualIP string
	// Subnet mask length (as a string)
	Subnet string
	// Routes to add
	Routes []string
}

// TUNService handles TUN interface operations
type TUNService struct {
	// Configuration for the TUN interface
	config *TUNConfig
	// TUN interface
	iface *water.Interface
}

// NewTUNService creates a new TUN service
func NewTUNService(config *TUNConfig) *TUNService {
	return &TUNService{
		config: config,
	}
}

// SetupTUN initializes a TUN device
func (s *TUNService) SetupTUN() (*water.Interface, error) {
	config := water.Config{DeviceType: water.TUN}
	iface, err := water.New(config)
	if err != nil {
		return nil, err
	}

	s.iface = iface

	link, err := netlink.LinkByName(iface.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to find interface: %w", err)
	}

	// Set MTU
	if err := netlink.LinkSetMTU(link, s.config.MTU); err != nil {
		return nil, fmt.Errorf("failed to set MTU: %w", err)
	}

	// Assign IP address
	addr, err := netlink.ParseAddr(fmt.Sprintf("%s/%s", s.config.VirtualIP, s.config.Subnet))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IP: %w", err)
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return nil, fmt.Errorf("failed to assign IP: %w", err)
	}

	// Bring the interface up
	if err := netlink.LinkSetUp(link); err != nil {
		return nil, fmt.Errorf("failed to bring up interface: %w", err)
	}

	// Add routes
	for _, r := range s.config.Routes {
		_, dst, err := net.ParseCIDR(r)
		if err != nil {
			log.Errorf("Invalid route: %s", r)
			continue
		}

		route := &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       dst,
		}

		if err := netlink.RouteReplace(route); err != nil {
			log.Errorf("Failed to add route %s: %v", dst, err)
		}
	}

	log.Infof("TUN interface created with name %s, address %s", iface.Name(), addr.String())

	return iface, nil
}

// Close closes the TUN interface
func (s *TUNService) Close() error {
	if s.iface != nil {
		return s.iface.Close()
	}
	return nil
}
