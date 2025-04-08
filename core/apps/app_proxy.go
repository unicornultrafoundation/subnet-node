package apps

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"net"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// Register a handler with a prefix matcher
func (s *Service) RegisterReverseProxyHandler() {
	s.PeerHost.SetStreamHandlerMatch(atypes.ProtocolProxyReverse, matchPrefix, s.OnReverseRequestReceive)
}

// Custom function to match protocol prefixes
func matchPrefix(proto protocol.ID) bool {
	parts := strings.Split(string(proto), "/")
	return len(parts) >= 4 && strings.HasPrefix(string(proto), string(atypes.ProtocolProxyReverse)+"/")
}

// Reverse request to specific port inside an app in Docker container
func (s *Service) OnReverseRequestReceive(stream network.Stream) {
	defer stream.Close()

	// Extract metadata from protocol name
	protocolParts := strings.Split(string(stream.Protocol()), "/")
	if len(protocolParts) < 4 { // At least 4 parts for /{p2pPrefixProtocol}/{requestProtocol}/{appId}/{appPort}
		writeErrorToStream(stream, "Invalid protocol format")
		return
	}

	protocol := protocolParts[len(protocolParts)-3]
	appIdStr := protocolParts[len(protocolParts)-2]
	appPort := protocolParts[len(protocolParts)-1]

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
			writeErrorToStream(stream, "TCP forward error (stream to target):"+err.Error())
		}
		targetConn.Close()
		stream.Close()
	}()

	_, err = io.Copy(stream, targetConn)
	if err != nil {
		writeErrorToStream(stream, "TCP forward error (target to stream):"+err.Error())
	}
}

func (s *Service) handleReverseUDP(stream network.Stream, _ string) {
	// TODO: handleReverseUDP
	writeErrorToStream(stream, "Unsupported udp protocol yet")
}

func writeErrorToStream(stream network.Stream, errorMessage string) {
	log.Println("Error:", errorMessage)
	stream.Write([]byte("ERROR: " + errorMessage + "\n"))
}
