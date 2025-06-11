package apps

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"net"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/patrickmn/go-cache"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/discovery"
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

// extractSourceIP extracts the source IP from a network stream
func (s *Service) extractSourceIP(stream network.Stream) string {
	remoteAddr := stream.Conn().RemoteMultiaddr().String()
	// Try to extract IP from multiaddr format
	parts := strings.Split(remoteAddr, "/")
	for i, part := range parts {
		if part == "ip4" && i+1 < len(parts) {
			return parts[i+1]
		}
		if part == "ip6" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	// Fallback to peer ID if no IP found
	return stream.Conn().RemotePeer().String()
}

// checkClusterMembership verifies if source IP and destination IP are in the same cluster via smart contract
func (s *Service) checkClusterMembership(sourceIP, destIP string) bool {
	if s.accountService == nil {
		log.Debug("Account service not available, skipping cluster membership check")
		return true // Allow if no account service (backward compatibility)
	}

	// Create cache key by combining both IPs
	cacheKey := fmt.Sprintf("%s_%s", sourceIP, destIP)

	// Check cache first
	if cachedResult, found := s.clusterMembershipCache.Get(cacheKey); found {
		if result, ok := cachedResult.(bool); ok {
			log.Debugf("Cluster membership check cache hit: sourceIP=%s, destIP=%s, sameCluster=%v", sourceIP, destIP, result)
			return result
		}
	}

	// Get subnet cluster market contract from account service
	subnetClusterMarket := s.accountService.SubnetClusterMarket()
	if subnetClusterMarket == nil {
		log.Debug("Subnet cluster market contract not available, skipping cluster membership check")
		return true
	}

	// Convert IPs to big integers as expected by the smart contract
	sourceIPNum := discovery.ConvertVirtualIPToNumber(sourceIP)
	if sourceIPNum == 0 {
		log.Warnf("Failed to convert source IP %s to number", sourceIP)
		result := false
		s.clusterMembershipCache.Set(cacheKey, result, cache.DefaultExpiration)
		return result
	}

	destIPNum := discovery.ConvertVirtualIPToNumber(destIP)
	if destIPNum == 0 {
		log.Warnf("Failed to convert destination IP %s to number", destIP)
		result := false
		s.clusterMembershipCache.Set(cacheKey, result, cache.DefaultExpiration)
		return result
	}

	// Call AreNodesInAnySameCluster method
	callOpts := &bind.CallOpts{
		Context: context.Background(),
	}

	sourceIPBig := big.NewInt(int64(sourceIPNum))
	destIPBig := big.NewInt(int64(destIPNum))

	areInSameCluster, err := subnetClusterMarket.AreNodesInAnySameCluster(callOpts, sourceIPBig, destIPBig)
	if err != nil {
		log.Errorf("Failed to check cluster membership for IPs %s and %s: %v", sourceIP, destIP, err)
		result := false
		s.clusterMembershipCache.Set(cacheKey, result, cache.DefaultExpiration)
		return result
	}

	// Cache the successful result
	s.clusterMembershipCache.Set(cacheKey, areInSameCluster, cache.DefaultExpiration)

	log.Debugf("Cluster membership check: sourceIP=%s, destIP=%s, sameCluster=%v (cached)", sourceIP, destIP, areInSameCluster)
	return areInSameCluster
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

	// Get container information (this gives us both IP and actual container ID)
	container, err := s.ContainerInspect(context.Background(), appId)
	if err != nil {
		writeErrorToStream(stream, "Failed to inspect container: "+err.Error())
		return
	}

	containerIP := container.NetworkSettings.IPAddress
	if containerIP == "" {
		writeErrorToStream(stream, "Container IP not available")
		return
	}

	containerID := container.ID

	// Extract source IP from stream
	sourceIP := s.extractSourceIP(stream)

	// Check if IP is in allow list first using the actual container ID (OR logic implementation)
	if s.containerACLManager != nil && s.containerACLManager.IsIPAllowedByContainerID(sourceIP, container.ID) {
		log.Debugf("Connection from %s allowed by Container ACL, skipping cluster membership check (Container ID: %s)", sourceIP, containerID[:12]) // Show first 12 chars of container ID
	} else {
		// If not in allow list, perform cluster membership check
		if !s.checkClusterMembership(sourceIP, containerIP) {
			log.Warnf("Cluster membership check failed: source IP %s not in same cluster as destination IP %s (Container ID: %s)", sourceIP, containerIP, containerID[:12])
			writeErrorToStream(stream, "Access denied: not in same cluster")
			return
		}
		log.Debugf("Cluster membership check passed: source IP %s and destination IP %s are in same cluster (Container ID: %s)", sourceIP, containerIP, containerID[:12])
	}

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
