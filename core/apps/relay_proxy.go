package apps

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	ma "github.com/multiformats/go-multiaddr"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// ConnectToRelayNode connects this node to a relay node using the provided peer address info
func (s *Service) ConnectToRelayNode(ctx context.Context, relayAddrInfo peer.AddrInfo) error {
	log.Infof("Connecting to relay node %s", relayAddrInfo.ID.String())

	// Connect to the relay node
	if err := s.PeerHost.Connect(ctx, relayAddrInfo); err != nil {
		return fmt.Errorf("failed to connect to relay node: %w", err)
	}

	// Try to reserve a slot with the relay
	reservation, err := client.Reserve(ctx, s.PeerHost, relayAddrInfo)
	if err != nil {
		return fmt.Errorf("failed to reserve slot with relay: %w", err)
	}

	log.Infof("Successfully connected to relay node %s with reservation TTL: %v",
		relayAddrInfo.ID.String(), time.Until(reservation.Expiration))

	return nil
}

// ConnectToRelayNodeByMultiaddr connects this node to a relay node using a multiaddress string
func (s *Service) ConnectToRelayNodeByMultiaddr(ctx context.Context, relayMultiaddr string) error {
	// Parse the multiaddress
	addr, err := ma.NewMultiaddr(relayMultiaddr)
	if err != nil {
		return fmt.Errorf("invalid relay multiaddress: %w", err)
	}

	// Extract the peer ID and addresses
	addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("failed to extract peer info from multiaddress: %w", err)
	}

	// Connect to the relay node
	return s.ConnectToRelayNode(ctx, *addrInfo)
}

// // ConnectThroughRelayProxy connects to a target peer through a relay proxy
func (s *Service) ConnectThroughRelayProxy(ctx context.Context, relayPeerID peer.ID, targetPeerID peer.ID, targetService string) (network.Stream, error) {
	// Format protocol ID with target peer embedded in it
	protocolID := protocol.ID(fmt.Sprintf("%s/%s/%s", atypes.ProtocolProxyRelay, targetPeerID.String(), targetService))

	// Create context that allows limited connections for relay
	ctxWithRelay := network.WithAllowLimitedConn(ctx, "relay")

	// Connect to relay node
	stream, err := s.PeerHost.NewStream(ctxWithRelay, relayPeerID, protocolID)
	if err != nil {
		return nil, fmt.Errorf("error creating stream to relay: %w", err)
	}

	return stream, nil
}

// ConnectToPeerViaRelay connects to a target peer through a relay using the circuit relay protocol
func (s *Service) ConnectToPeerViaRelay(ctx context.Context, relayAddrInfo peer.AddrInfo, targetPeerID peer.ID) error {
	// First ensure we're connected to the relay
	if err := s.ConnectToRelayNode(ctx, relayAddrInfo); err != nil {
		return err
	}

	// Create a relay address for the target peer
	relayMultiAddr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s",
		relayAddrInfo.ID.String(), targetPeerID.String()))
	if err != nil {
		return fmt.Errorf("failed to create relay address: %w", err)
	}

	// Create address info for the target peer via relay
	targetAddrInfo := peer.AddrInfo{
		ID:    targetPeerID,
		Addrs: []ma.Multiaddr{relayMultiAddr},
	}

	// Connect to the target peer through the relay
	if err := s.PeerHost.Connect(ctx, targetAddrInfo); err != nil {
		return fmt.Errorf("failed to connect to target peer via relay: %w", err)
	}

	log.Infof("Successfully connected to peer %s via relay %s",
		targetPeerID.String(), relayAddrInfo.ID.String())

	return nil
}

// OpenStreamToPeerViaRelay opens a stream to a target peer through a relay using the specified protocol
func (s *Service) OpenStreamToPeerViaRelay(ctx context.Context, relayAddrInfo peer.AddrInfo,
	targetPeerID peer.ID, proto protocol.ID) (network.Stream, error) {

	// First ensure we're connected to the target peer via relay
	if err := s.ConnectToPeerViaRelay(ctx, relayAddrInfo, targetPeerID); err != nil {
		return nil, err
	}

	// Open a stream to the target peer
	ctxWithRelay := network.WithAllowLimitedConn(ctx, "relay")
	stream, err := s.PeerHost.NewStream(ctxWithRelay, targetPeerID, proto)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream to target peer: %w", err)
	}

	return stream, nil
}

// SetupRelayClient configures this node to use circuit relay for connections
func (s *Service) SetupRelayClient() error {
	if !s.cfg.GetBool("provider.relay_client.enabled", false) {
		return nil
	}

	log.Info("Setting up circuit relay client")

	// Get static relay addresses from config
	relayAddrs := s.cfg.GetStringSlice("provider.relay_client.static_relays", []string{})
	if len(relayAddrs) == 0 {
		log.Warn("No static relay addresses configured for relay client")
		return nil
	}

	// Connect to each relay
	ctx := context.Background()
	for _, addrStr := range relayAddrs {
		if err := s.ConnectToRelayNodeByMultiaddr(ctx, addrStr); err != nil {
			log.Warnf("Failed to connect to relay %s: %v", addrStr, err)
			continue
		}
	}

	return nil
}

// CreateRelayService configures this node to act as a circuit relay
func (s *Service) CreateRelayService() error {
	if !s.cfg.GetBool("provider.relay.enabled", false) {
		return nil
	}

	log.Info("Setting up circuit relay service")

	// This is just a placeholder - the actual relay service is configured
	// at a lower level in the libp2p stack in core/node/libp2p/relay.go
	// This function can be used to perform additional setup if needed

	return nil
}

// RelayStreamHandler handles streams for the relay protocol
func (s *Service) RelayStreamHandler(stream network.Stream) {
	defer stream.Close()

	// Extract target peer ID and protocol from the stream protocol ID
	protoStr := string(stream.Protocol())
	targetInfo, err := parseRelayProtocol(protoStr)
	if err != nil {
		writeErrorToStream(stream, fmt.Sprintf("Invalid protocol format: %v", err))
		return
	}

	// Create a context with relay allowed
	ctx := network.WithAllowLimitedConn(context.Background(), "relay")

	// Open a stream to the target peer
	targetStream, err := s.PeerHost.NewStream(ctx, targetInfo.targetPeerID, targetInfo.targetProto)
	if err != nil {
		writeErrorToStream(stream, fmt.Sprintf("Failed to connect to target peer: %v", err))
		return
	}
	defer targetStream.Close()

	// Create bidirectional pipe between streams
	errChan := make(chan error, 2)

	// Forward from source to target
	go func() {
		_, err := io.Copy(targetStream, stream)
		targetStream.Close()
		errChan <- err
	}()

	// Forward from target to source
	go func() {
		_, err := io.Copy(stream, targetStream)
		stream.Close()
		errChan <- err
	}()

	// Wait for either copy to finish
	<-errChan
}

// Helper struct for relay protocol parsing
type relayProtocolInfo struct {
	targetPeerID peer.ID
	targetProto  protocol.ID
}

// Parse relay protocol string to extract target peer ID and protocol
func parseRelayProtocol(protoStr string) (*relayProtocolInfo, error) {
	parts := strings.Split(protoStr, "/")
	if len(parts) < 6 {
		return nil, fmt.Errorf("insufficient protocol parts")
	}

	// Extract target peer ID
	targetPeerID, err := peer.Decode(parts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid peer ID: %w", err)
	}

	// Extract target service/protocol
	targetService := parts[5]
	targetProto := protocol.ID(fmt.Sprintf("%s/%s", atypes.ProtocolProxyRelay, targetService))

	return &relayProtocolInfo{
		targetPeerID: targetPeerID,
		targetProto:  targetProto,
	}, nil
}
