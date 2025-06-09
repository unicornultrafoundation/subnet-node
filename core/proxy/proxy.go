package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"

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

				// Allow IP filtering
				remoteAddr, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
				if !isIPAllowed(remoteAddr, m.AllowIPs) {
					log.Warnf("Connection from %s is not allowed (AppId: %s, AppPort: %d)", remoteAddr, appId, m.AppPort)
					conn.Close()
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

	// Get the remote address
	remoteAddr, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	// Check if the source IP is allowed
	if !isIPAllowed(remoteAddr, m.AllowIPs) {
		log.Warnf("Connection from %s is not allowed (AppId: %s, AppPort: %d). Proceeding to next step.", remoteAddr, appId, m.AppPort)
		return // Proceed to the next connection
	}

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

// isIPAllowed checks if the given IP is in the allowed list (supports CIDR and single IP)
func isIPAllowed(ip string, allowList []string) bool {
	if len(allowList) == 0 {
		return true // If no allow list, allow all
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	for _, cidr := range allowList {
		if strings.Contains(cidr, "/") {
			_, ipnet, err := net.ParseCIDR(cidr)
			if err == nil && ipnet.Contains(parsedIP) {
				return true
			}
		} else {
			if cidr == ip {
				return true
			}
		}
	}
	return false
}
