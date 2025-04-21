package network

import (
	"io"
	"strconv"

	"github.com/songgao/water"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
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
	// Metrics for monitoring
	metrics VPNMetricsInterface
}

// NewServerService creates a new server service
func NewServerService(config *ServerConfig, metrics VPNMetricsInterface) *ServerService {
	return &ServerService{
		config:  config,
		metrics: metrics,
	}
}

// HandleStream handles an incoming P2P stream
func (s *ServerService) HandleStream(stream types.VPNStream, iface *water.Interface) {
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

			packetInfo, err := packet.ExtractIPAndPorts(buf[:n])
			if err != nil {
				log.Debugf("error extracting packet info: %v", err)
				s.metrics.IncrementPacketsDropped()
				continue
			}

			// Check if the destination port is allowed
			if packetInfo.DstPort != nil {
				dstPort := strconv.Itoa(*packetInfo.DstPort)
				// Reject all requests that are to unallowed ports
				unallowedPort, exist := s.config.UnallowedPorts[dstPort]
				if exist && unallowedPort {
					s.metrics.IncrementPacketsDropped()
					continue
				}

				// Reject all requests that are to ports outside the allowed range
				if *packetInfo.DstPort < 30000 || *packetInfo.DstPort > 65535 {
					s.metrics.IncrementPacketsDropped()
					continue
				}
			}

			// Write the packet to the TUN interface
			_, err = iface.Write(buf[:n])
			if err != nil {
				log.Errorf("error writing to TUN interface: %v", err)
				s.metrics.IncrementPacketsDropped()
				continue
			}

			// Update metrics
			s.metrics.IncrementPacketsSent(n)
		}
	}()
}
