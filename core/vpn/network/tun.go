package network

import (
	"fmt"
	"io"
	"net/netip"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/overlay"
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
	// Number of reader routines
	Routines int
}

// TUNService handles TUN interface operations
type TUNService struct {
	// Configuration for the TUN interface
	config *TUNConfig
	// TUN interface using overlay.Device
	device overlay.Device
	// Reader queues for the TUN device
	readers []io.ReadWriteCloser
	// Number of reader routines
	routines int
	// Cidr for the TUN interface
	cidr netip.Prefix
	// Logger
	logger *logrus.Entry
}

// NewTUNService creates a new TUN service
func NewTUNService(config *TUNConfig) *TUNService {
	// Create a new logger
	logger := logrus.WithField("service", "vpn-tun")

	// Default to 1 routine if not specified in config
	routines := config.Routines
	if routines <= 0 {
		routines = 1
	}

	return &TUNService{
		config:   config,
		readers:  make([]io.ReadWriteCloser, 0, routines),
		routines: routines,
		logger:   logger,
	}
}

// SetupTUN initializes a TUN device using overlay package
func (s *TUNService) SetupTUN() error {
	// Create overlay configuration
	cfg, cidr, err := s.createOverlayConfig()
	if err != nil {
		return err
	}
	s.cidr = cidr

	// Create a new device from config
	device, err := overlay.NewDeviceFromConfig(cfg, s.logger.Logger, []netip.Prefix{cidr}, s.routines)
	if err != nil {
		return fmt.Errorf("failed to create TUN device: %w", err)
	}

	// Store the device
	s.device = device

	// Activate the device
	if err := device.Activate(); err != nil {
		device.Close()
		return fmt.Errorf("failed to activate TUN device: %w", err)
	}

	// Prepare reader queues
	s.readers = make([]io.ReadWriteCloser, s.routines)

	// First reader is the device itself
	var reader io.ReadWriteCloser = device
	for i := 0; i < s.routines; i++ {
		if i > 0 {
			// Create additional readers for multi-queue support
			reader, err = device.NewMultiQueueReader()
			if err != nil {
				s.logger.WithError(err).Error("Failed to create multi-queue reader")
				device.Close()
				return fmt.Errorf("failed to create multi-queue reader: %w", err)
			}
		}
		s.readers[i] = reader
	}

	log.Infof("TUN interface created with name %s", device.Name())

	return nil
}

// createOverlayConfig creates a configuration for the overlay device
func (s *TUNService) createOverlayConfig() (*config.C, netip.Prefix, error) {
	// Parse the virtual IP as a prefix
	cidr, err := netip.ParsePrefix(fmt.Sprintf("%s/%s", s.config.VirtualIP, s.config.Subnet))
	if err != nil {
		return nil, netip.Prefix{}, fmt.Errorf("failed to parse virtual IP: %w", err)
	}

	// Create a new config for overlay
	cfg := config.NewC(s.logger.Logger)

	// Set TUN configuration
	cfg.Settings = map[string]any{
		"tun": map[string]any{
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
		tunMap := cfg.Settings["tun"].(map[string]any)
		tunMap["routes"] = routesArray
	}

	return cfg, cidr, nil
}

// Close closes the TUN interface and all readers
func (s *TUNService) Close() error {
	// Close all readers except the first one (which is the device itself)
	for i := 1; i < len(s.readers); i++ {
		if s.readers[i] != nil {
			if err := s.readers[i].Close(); err != nil {
				s.logger.WithError(err).Errorf("Failed to close reader %d", i)
			}
		}
	}

	// Clear the readers slice
	s.readers = nil

	// Close the device (which will also close the first reader)
	if s.device != nil {
		return s.device.Close()
	}
	return nil
}

// GetReaders returns the readers for the TUN device
func (s *TUNService) GetReaders() []io.ReadWriteCloser {
	return s.readers
}

// GetRoutines returns the number of reader routines
func (s *TUNService) GetRoutines() int {
	return s.routines
}

// GetCIDR returns the CIDR for the TUN interface
func (s *TUNService) GetCIDR() netip.Prefix {
	return s.cidr
}
