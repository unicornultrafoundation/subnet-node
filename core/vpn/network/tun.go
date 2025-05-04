package network

import (
	"fmt"
	"net/netip"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/overlay"
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
	// TUN interface using overlay.Device
	device overlay.Device
	// Logger
	logger *logrus.Logger
}

// NewTUNService creates a new TUN service
func NewTUNService(config *TUNConfig) *TUNService {
	// Create a new logger
	logger := logrus.New()

	return &TUNService{
		config: config,
		logger: logger,
	}
}

// createOverlayConfig creates a configuration for the overlay device
func (s *TUNService) createOverlayConfig() (*config.C, []netip.Prefix, error) {
	// Parse the virtual IP as a prefix
	prefix, err := netip.ParsePrefix(fmt.Sprintf("%s/%s", s.config.VirtualIP, s.config.Subnet))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse virtual IP: %w", err)
	}

	// Create a slice of network prefixes
	vpnNetworks := []netip.Prefix{prefix}

	// Create a new config for overlay
	cfg := config.NewC(s.logger)

	// Set TUN configuration
	cfg.Settings = map[interface{}]interface{}{
		"tun": map[interface{}]interface{}{
			"mtu": s.config.MTU,
		},
	}

	// Add routes if provided
	if len(s.config.Routes) > 0 {
		// Create a properly formatted routes array for the overlay package
		routesArray := make([]interface{}, 0, len(s.config.Routes))
		for _, route := range s.config.Routes {
			routeMap := map[string]interface{}{
				"mtu":   s.config.MTU, // Use the same MTU as the TUN interface
				"route": route,
			}
			routesArray = append(routesArray, routeMap)
		}

		// Set the routes in the config
		tunMap := cfg.Settings["tun"].(map[interface{}]interface{})
		tunMap["routes"] = routesArray
	}

	return cfg, vpnNetworks, nil
}

// SetupTUN initializes a TUN device using overlay package
func (s *TUNService) SetupTUN() (overlay.Device, error) {
	// Create overlay configuration
	cfg, vpnNetworks, err := s.createOverlayConfig()
	if err != nil {
		return nil, err
	}

	// Create a new device from config
	device, err := overlay.NewDeviceFromConfig(cfg, s.logger, vpnNetworks, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %w", err)
	}

	// Store the device
	s.device = device

	// Activate the device
	if err := device.Activate(); err != nil {
		device.Close()
		return nil, fmt.Errorf("failed to activate TUN device: %w", err)
	}

	log.Infof("TUN interface created with name %s", device.Name())

	return device, nil
}

// Close closes the TUN interface
func (s *TUNService) Close() error {
	if s.device != nil {
		return s.device.Close()
	}
	return nil
}
