package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

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
				log.Println("TCP Accept error:", err)
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
		log.Println("Unsupported protocol:", m.Protocol)
	}
}

// handleConnection forwards TCP connections via P2P
func (s *Service) handleTCPConnection(conn net.Conn, m PortMapping) {
	defer conn.Close()

	stream, err := s.PeerHost.NewStream(context.Background(), s.RemotePeerId, atypes.ProtocolProxyReverse)
	if err != nil {
		log.Println("Failed to create TCP stream:", err)
		conn.Write([]byte("ERROR: Failed to create TCP stream\n"))
		return
	}
	defer stream.Close()

	// Send metadata first (protocol:appId:appPort)
	meta := fmt.Sprintf("tcp:%s:%d\n", s.AppId, m.AppPort)
	stream.Write([]byte(meta))

	// Forward bidirectional TCP traffic
	go io.Copy(stream, conn)
	io.Copy(conn, stream)
}

// handleConnection forwards UDP connections via P2P
func (s *Service) handleUDPConnection(conn net.PacketConn, m PortMapping) {
	buf := make([]byte, 4096)

	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Println("UDP Read error:", err)
			continue
		}

		// Open a P2P stream
		stream, err := s.PeerHost.NewStream(context.Background(), s.RemotePeerId, atypes.ProtocolProxyReverse)
		if err != nil {
			log.Println("Failed to create UDP stream:", err)
			continue
		}

		// Send metadata first
		meta := fmt.Sprintf("udp:%s:%d\n", s.AppId, m.AppPort)
		stream.Write([]byte(meta))

		// Send UDP data
		_, err = stream.Write(buf[:n])
		if err != nil {
			log.Println("Failed to forward UDP packet:", err)
			stream.Write([]byte("ERROR: Failed to forward UDP packet\n"))
		}

		// Handle response
		go s.receiveUDPPacket(stream, conn, addr)
	}
}

func (s *Service) receiveUDPPacket(stream network.Stream, conn net.PacketConn, addr net.Addr) {
	defer stream.Close()
	buf := make([]byte, 4096)

	for {
		n, err := stream.Read(buf)
		if err != nil {
			log.Println("UDP stream read error:", err)
			return
		}

		// Send response back to original sender
		_, err = conn.WriteTo(buf[:n], addr)
		if err != nil {
			log.Println("Failed to send UDP response:", err)
			return
		}
	}
}
