package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"

	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// forwardTraffic listens on localPort and forwards traffic to target appPort
func (s *Service) forwardTraffic(m PortMapping) {
	address := fmt.Sprintf("%s:%d", m.HostIP, m.LocalPort)

	if strings.HasPrefix(m.Protocol, "tcp") {
		// TCP Forwarding
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Errorf("Failed to listen on %s (tcp): %v\n", address, err)
			return
		}
		defer listener.Close()

		log.Printf("Forwarding TCP %s -> peer %s (AppId: %s, AppPort: %d)\n", address, s.RemotePeerId, s.AppId, m.AppPort)

		for {
			select {
			case <-s.stopChan:
				log.Infof("Stopping forwarding TCP trafic %s -> peer %s (AppId: %s, AppPort: %d)", address, s.RemotePeerId, s.AppId, m.AppPort)
				listener.Close()
				return

			default:
				conn, err := listener.Accept()
				if err != nil {
					log.Debugln("TCP Accept error:", err)
					continue
				}
				go s.handleTCPConnection(conn, m)
			}
		}
	} else {
		log.Debugln("Unsupported protocol:", m.Protocol)
	}
}

// handleConnection forwards TCP connections via P2P
func (s *Service) handleTCPConnection(conn net.Conn, m PortMapping) {
	defer conn.Close()

	stream, err := s.PeerHost.NewStream(context.Background(), s.RemotePeerId, atypes.ProtocolProxyReverse)
	if err != nil {
		log.Debugf("TCP: Failed to create P2P stream to peer %s (AppId: %s, AppPort: %d): %v", s.RemotePeerId.String(), s.AppId, m.AppPort, err)
		return
	}
	defer stream.Close()

	meta := fmt.Sprintf("tcp:%s:%d\n", s.AppId, m.AppPort)
	if _, err := stream.Write([]byte(meta)); err != nil {
		log.Debugf("TCP: Failed to send metadata to peer %s (AppId: %s, AppPort: %d): %v", s.RemotePeerId.String(), s.AppId, m.AppPort, err)
		return
	}

	// Forward bidirectional TCP traffic
	go io.Copy(stream, conn) // Client → P2P
	io.Copy(conn, stream)    // P2P → Client
}
