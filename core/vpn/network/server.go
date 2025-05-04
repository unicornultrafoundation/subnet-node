package network

import (
	"io"
	"strconv"

	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/overlay"
)

// ServerConfig contains configuration for the server
type ServerConfig struct {
	// MTU for packets
	MTU int
	// Unallowed ports
	UnallowedPorts map[string]bool
}

// ServerService handles server-side VPN operations
type ServerService struct {
	// Configuration for the server
	config *ServerConfig
}

// NewServerService creates a new server service
func NewServerService(config *ServerConfig) *ServerService {
	return &ServerService{
		config: config,
	}
}

// HandleStream handles an incoming P2P stream
func (s *ServerService) HandleStream(stream api.VPNStream, device overlay.Device) {
	go func() {
		defer stream.Close()

		buf := make([]byte, s.config.MTU)
		for {
			n, err := stream.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Debugf("P2P stream EOF")
				} else {
					log.Errorf("error reading P2P stream: %v", err)
				}
				return
			}

			packetInfo, err := types.ExtractPacketInfo(buf[:n])
			if err != nil {
				log.Debugf("error extracting packet info: %v", err)
				continue
			}

			// Check if the destination port is allowed
			if packetInfo.DstPort != nil {
				dstPort := strconv.Itoa(*packetInfo.DstPort)
				// Reject all requests that are to unallowed ports
				unallowedPort, exist := s.config.UnallowedPorts[dstPort]
				if exist && unallowedPort {
					continue
				}

				// Reject all requests that are to ports outside the allowed range
				if *packetInfo.DstPort < 30000 || *packetInfo.DstPort > 65535 {
					continue
				}
			}

			// Write the packet to the TUN interface
			_, err = device.Write(buf[:n])
			if err != nil {
				log.Errorf("error writing to TUN interface: %v", err)
				continue
			}
		}
	}()
}
