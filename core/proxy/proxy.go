package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// Client session struct to track UDP forwarding sessions
type udpSession struct {
	stream   network.Stream
	lastUsed time.Time
}

var udpSessions sync.Map // Maps client address to active UDP session

// forwardTraffic listens on localPort and forwards traffic to target appPort
func (s *Service) forwardTraffic(m PortMapping) {
	address := fmt.Sprintf("%s:%d", m.HostIP, m.LocalPort)

	if strings.HasPrefix(m.Protocol, "tcp") {
		// TCP Forwarding
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Printf("Failed to listen on %s (tcp): %v\n", address, err)
			return
		}
		defer listener.Close()

		log.Printf("Forwarding TCP %s -> peer %s (AppId: %s, AppPort: %d)\n", address, s.RemotePeerId, s.AppId, m.AppPort)

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Debugln("TCP Accept error:", err)
				continue
			}
			go s.handleTCPConnection(conn, m)
		}
	} else if strings.HasPrefix(m.Protocol, "udp") {
		// UDP Forwarding
		udpConn, err := net.ListenPacket("udp", address)
		if err != nil {
			log.Printf("Failed to listen on %s (udp): %v\n", address, err)
			return
		}
		defer udpConn.Close()

		log.Printf("Forwarding UDP %s -> peer %s (AppId: %s, AppPort: %d)\n", address, s.RemotePeerId, s.AppId, m.AppPort)

		s.handleUDPConnection(udpConn, m)
	} else {
		log.Debugln("Unsupported protocol:", m.Protocol)
	}
}

// handleConnection forwards TCP connections via P2P
func (s *Service) handleTCPConnection(conn net.Conn, m PortMapping) {
	defer conn.Close()

	log.Println("TCP: Accepted connection, setting up P2P stream...")

	stream, err := s.PeerHost.NewStream(context.Background(), s.RemotePeerId, atypes.ProtocolProxyReverse)
	if err != nil {
		log.Println("TCP: Failed to create P2P stream:", err)
		return
	}
	defer stream.Close()

	log.Println("TCP: P2P stream established, sending metadata...")

	meta := fmt.Sprintf("tcp:%s:%d\n", s.AppId, m.AppPort)
	if _, err := stream.Write([]byte(meta)); err != nil {
		log.Println("TCP: Failed to send metadata:", err)
		return
	}

	log.Println("TCP: Metadata sent, starting data forwarding...")

	// Forward bidirectional TCP traffic
	go io.Copy(stream, conn) // Client → P2P
	io.Copy(conn, stream)    // P2P → Client
}

// handleUDPConnection forwards UDP traffic via P2P
func (s *Service) handleUDPConnection(conn net.PacketConn, m PortMapping) {
	buf := make([]byte, 4096)

	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Println("UDP Read error:", err)
			continue
		}

		clientAddr := addr.String()
		session, found := udpSessions.Load(clientAddr)

		var stream network.Stream
		if found {
			s := session.(*udpSession)
			stream = s.stream
			s.lastUsed = time.Now()
		} else {
			// Open new P2P stream for this client
			stream, err = s.PeerHost.NewStream(context.Background(), s.RemotePeerId, atypes.ProtocolProxyReverse)
			if err != nil {
				log.Printf("Failed to create UDP stream for %s: %v", addr.String(), err)
				continue
			}

			udpSessions.Store(clientAddr, &udpSession{stream: stream, lastUsed: time.Now()})
			go s.receiveUDPPacket(stream, conn, addr)
		}

		// Send metadata and UDP data
		meta := fmt.Sprintf("udp:%s:%d\n", s.AppId, m.AppPort)
		stream.Write([]byte(meta))
		stream.Write(buf[:n])

		log.Printf("Forwarded UDP packet from %s to peer (%d bytes)", addr.String(), n)
	}
}

// receiveUDPPacket reads responses from the P2P stream and sends them back to the correct UDP client
func (s *Service) receiveUDPPacket(stream network.Stream, conn net.PacketConn, addr net.Addr) {
	defer stream.Close()
	buf := make([]byte, 4096)

	for {
		n, err := stream.Read(buf)
		if err != nil {
			log.Printf("[UDP] Stream read error: %v (addr: %s)", err, addr.String())
			return
		}

		log.Printf("[UDP] Received UDP response from peer: %d bytes (addr: %s)", n, addr.String())

		// Send response back to original sender
		_, err = conn.WriteTo(buf[:n], addr)
		if err != nil {
			log.Printf("[UDP] Failed to send UDP response to %s: %v", addr.String(), err)
			return
		}

		log.Printf("[UDP] Forwarded UDP response to client %s (%d bytes)", addr.String(), n)
	}
}
