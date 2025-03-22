package proxy

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// forwardTraffic listens on localPort and forwards traffic to target appPort
func (s *Service) forwardTraffic(remotePeerId peer.ID, appId string, m PortMapping) {
	address := fmt.Sprintf("%s:%d", m.HostIP, m.LocalPort)

	switch m.Protocol {
	case "tcp":
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Errorf("Failed to listen on %s (tcp): %v\n", address, err)
			return
		}
		defer listener.Close()

		log.Printf("Forwarding TCP %s -> peer %s (AppId: %s, AppPort: %d)\n", address, remotePeerId, appId, m.AppPort)

		for {
			select {
			case <-s.stopChan:
				log.Infof("Stopping forwarding TCP trafic %s -> peer %s (AppId: %s, AppPort: %d)", address, remotePeerId, appId, m.AppPort)
				listener.Close()
				return

			default:
				conn, err := listener.Accept()
				if err != nil {
					log.Debugln("TCP Accept error:", err)
					continue
				}
				go s.handleTCPConnection(conn, remotePeerId, appId, m)
			}
		}
	default:
		log.Debugln("Unsupported protocol:", m.Protocol)
	}

}

// handleConnection forwards TCP connections via P2P
func (s *Service) handleTCPConnection(conn net.Conn, remotePeerId peer.ID, appId string, m PortMapping) {
	defer conn.Close()

	// Encode metadata in the protocol name
	protocolWithMeta := protocol.ID(fmt.Sprintf("%s/tcp/%s/%d", atypes.ProtocolProxyReverse, appId, m.AppPort))
	stream, err := s.PeerHost.NewStream(context.Background(), remotePeerId, protocolWithMeta)
	if err != nil {
		log.Debugf("TCP: Failed to create P2P stream to peer %s (AppId: %s, AppPort: %d): %v", remotePeerId.String(), appId, m.AppPort, err)
		return
	}
	defer stream.Close()

	// Forward bidirectional TCP traffic
	go io.Copy(stream, conn) // Client → P2P
	io.Copy(conn, stream)    // P2P → Client
}
