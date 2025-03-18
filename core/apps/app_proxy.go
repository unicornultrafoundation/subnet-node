package apps

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/big"
	"net"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

// Reverse request to specific port inside an app in Docker container
func (s *Service) OnReverseRequestReceive(stream network.Stream) {
	defer stream.Close()
	reader := bufio.NewReader(stream)

	// Read metadata (protocol:AppId:AppPort)
	metaLine, err := reader.ReadString('\n')
	if err != nil {
		writeErrorToStream(stream, "Failed to read metadata: "+err.Error())
		return
	}

	metaParts := strings.Split(strings.TrimSpace(metaLine), ":")
	if len(metaParts) != 3 {
		writeErrorToStream(stream, "Invalid metadata format")
		return
	}

	protocol, appIdStr, appPort := metaParts[0], metaParts[1], metaParts[2]

	// Parse AppId
	appId := new(big.Int)
	appId, ok := appId.SetString(appIdStr, 16)
	if !ok {
		writeErrorToStream(stream, "Failed to parse AppId")
		return
	}

	// Get container IP
	container, err := s.ContainerInspect(context.Background(), appId)
	if err != nil {
		writeErrorToStream(stream, "Failed to inspect container: "+err.Error())
		return
	}

	containerIP := container.NetworkSettings.IPAddress
	targetAddr := fmt.Sprintf("%s:%s", containerIP, appPort)

	// Handle based on protocol
	switch protocol {
	case "tcp":
		s.handleReverseTCP(stream, targetAddr)
	case "udp":
		s.handleReverseUDP(stream, targetAddr)
	default:
		writeErrorToStream(stream, "Unsupported protocol: "+protocol)
	}
}

func (s *Service) handleReverseTCP(stream network.Stream, targetAddr string) {
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		writeErrorToStream(stream, "Failed to connect to target (TCP): "+err.Error())
		return
	}
	defer targetConn.Close()

	// Forward bidirectional traffic
	go func() {
		_, err := io.Copy(targetConn, stream)
		if err != nil {
			log.Println("TCP forward error (stream to target):", err)
		}
		targetConn.Close()
		stream.Close()
	}()

	_, err = io.Copy(stream, targetConn)
	if err != nil {
		log.Println("TCP forward error (target to stream):", err)
	}
}

func (s *Service) handleReverseUDP(stream network.Stream, targetAddr string) {
	targetConn, err := net.Dial("udp", targetAddr)
	if err != nil {
		writeErrorToStream(stream, "Failed to connect to target (UDP): "+err.Error())
		return
	}
	defer targetConn.Close()

	buf := make([]byte, 4096)

	for {
		n, err := stream.Read(buf)
		if err != nil {
			log.Println("UDP Stream Read error:", err)
			return
		}

		// Send to target container
		_, err = targetConn.Write(buf[:n])
		if err != nil {
			writeErrorToStream(stream, "Failed to send UDP packet: "+err.Error())
			return
		}

		// Read response
		n, err = targetConn.Read(buf)
		if err != nil {
			writeErrorToStream(stream, "UDP Read response error: "+err.Error())
			return
		}

		// Forward back to sender
		_, err = stream.Write(buf[:n])
		if err != nil {
			log.Println("Failed to forward UDP response:", err)
			return
		}
	}
}

func writeErrorToStream(stream network.Stream, errorMessage string) {
	log.Println("Error:", errorMessage)
	stream.Write([]byte("ERROR: " + errorMessage + "\n"))
}
