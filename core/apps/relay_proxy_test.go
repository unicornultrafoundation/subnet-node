package apps

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/stretchr/testify/assert"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

func TestConnectThroughRelay(t *testing.T) {
	// Create two "unreachable" libp2p hosts that want to communicate
	unreachable1, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.EnableRelay(),
	)
	if err != nil {
		t.Fatalf("Failed to create unreachable1: %v", err)
	}
	defer unreachable1.Close()

	unreachable2, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.EnableRelay(),
	)
	if err != nil {
		t.Fatalf("Failed to create unreachable2: %v", err)
	}
	defer unreachable2.Close()

	// Create services for both nodes
	unreachable1Service := &Service{
		PeerHost: unreachable1,
	}

	unreachable2Service := &Service{
		PeerHost: unreachable2,
	}

	// Register protocol handlers
	testProtocol := protocol.ID("/test/1.0.0")

	// Set up a handler on unreachable2 to verify successful connection
	receivedMsg := make(chan string, 1)
	unreachable2.SetStreamHandler(testProtocol, func(s network.Stream) {
		defer s.Close()

		// Read the message
		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil {
			t.Errorf("Error reading from stream: %v", err)
			return
		}

		// Send the message to the channel
		receivedMsg <- string(buf[:n])
	})

	// First attempt to connect directly - this should fail
	t.Log("Attempting direct connection (should fail)")
	unreachable2info := peer.AddrInfo{
		ID:    unreachable2.ID(),
		Addrs: unreachable2.Addrs(),
	}

	err = unreachable1.Connect(context.Background(), unreachable2info)
	assert.Error(t, err, "Direct connection should have failed due to no listen addresses")

	// Create a relay node
	t.Log("Creating relay node")
	relayNode, err := libp2p.New()
	if err != nil {
		t.Fatalf("Failed to create relay node: %v", err)
	}
	defer relayNode.Close()

	// Enable the relay service on the relay node
	_, err = relay.New(relayNode)
	if err != nil {
		t.Fatalf("Failed to instantiate the relay service: %v", err)
	}

	// Get relay node info
	relayInfo := peer.AddrInfo{
		ID:    relayNode.ID(),
		Addrs: relayNode.Addrs(),
	}

	// Connect both unreachable nodes to the relay using our ConnectToRelayNode function
	t.Log("Connecting unreachable1 to relay")
	err = unreachable1Service.ConnectToRelayNode(context.Background(), relayInfo)
	assert.NoError(t, err, "Failed to connect unreachable1 to relay")

	t.Log("Connecting unreachable2 to relay")
	err = unreachable2Service.ConnectToRelayNode(context.Background(), relayInfo)
	assert.NoError(t, err, "Failed to connect unreachable2 to relay")

	// Connect unreachable1 to unreachable2 through the relay
	t.Log("Connecting unreachable1 to unreachable2 through relay")
	err = unreachable1Service.ConnectToPeerViaRelay(context.Background(), relayInfo, unreachable2.ID())
	assert.NoError(t, err, "Failed to connect to peer via relay")

	// Open a stream from unreachable1 to unreachable2 through the relay
	t.Log("Opening stream from unreachable1 to unreachable2 through relay")
	stream, err := unreachable1Service.OpenStreamToPeerViaRelay(context.Background(), relayInfo, unreachable2.ID(), testProtocol)
	assert.NoError(t, err, "Failed to open stream to peer via relay")
	defer stream.Close()

	// Send a test message
	testMessage := "Hello through relay!"
	_, err = stream.Write([]byte(testMessage))
	assert.NoError(t, err, "Failed to write to stream")

	// Wait for the message to be received
	select {
	case received := <-receivedMsg:
		assert.Equal(t, testMessage, received, "Received message doesn't match sent message")
		t.Log("Successfully received message through relay")
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for message")
	}
}

func TestRelayProxyHandler(t *testing.T) {
	// Create two "unreachable" libp2p hosts that want to communicate
	unreachable1, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.EnableRelay(),
	)
	if err != nil {
		t.Fatalf("Failed to create unreachable1: %v", err)
	}
	defer unreachable1.Close()

	unreachable2, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.EnableRelay(),
	)
	if err != nil {
		t.Fatalf("Failed to create unreachable2: %v", err)
	}
	defer unreachable2.Close()

	// Create services for both nodes
	unreachable1Service := &Service{
		PeerHost: unreachable1,
	}

	// Set up a custom protocol on unreachable2
	testService := "testservice"
	customProtocol := protocol.ID(string(atypes.ProtocolProxyRelay) + "/" + testService)

	// Set up a handler on unreachable2 to verify successful connection
	receivedMsg := make(chan string, 1)
	unreachable2.SetStreamHandler(customProtocol, func(s network.Stream) {
		defer s.Close()

		// Read the message
		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil {
			t.Errorf("Error reading from stream: %v", err)
			return
		}

		// Send the message to the channel
		receivedMsg <- string(buf[:n])

		// Send a response
		s.Write([]byte("Response from unreachable2"))
	})

	// Create a relay node
	relayNode, err := libp2p.New()
	if err != nil {
		t.Fatalf("Failed to create relay node: %v", err)
	}
	defer relayNode.Close()

	// Set up the relay handler on the relay node
	relayService := &Service{
		PeerHost: relayNode,
	}

	// Register the relay stream handler
	relayNode.SetStreamHandler(protocol.ID(string(atypes.ProtocolProxyRelay)+"/"+unreachable2.ID().String()+"/"+testService), relayService.RelayStreamHandler)

	// Connect both unreachable nodes to the relay
	relayInfo := peer.AddrInfo{
		ID:    relayNode.ID(),
		Addrs: relayNode.Addrs(),
	}

	err = unreachable1.Connect(context.Background(), relayInfo)
	assert.NoError(t, err, "Failed to connect unreachable1 to relay")

	err = unreachable2.Connect(context.Background(), relayInfo)
	assert.NoError(t, err, "Failed to connect unreachable2 to relay")

	// Use the ConnectThroughRelayProxy function to connect through the relay
	stream, err := unreachable1Service.ConnectThroughRelayProxy(context.Background(), relayNode.ID(), unreachable2.ID(), testService)
	assert.NoError(t, err, "Failed to connect through relay proxy")
	defer stream.Close()

	// Send a test message
	testMessage := "Hello through relay proxy!"
	_, err = stream.Write([]byte(testMessage))
	assert.NoError(t, err, "Failed to write to stream")

	// Wait for the message to be received
	select {
	case received := <-receivedMsg:
		assert.Equal(t, testMessage, received, "Received message doesn't match sent message")
		t.Log("Successfully received message through relay proxy")
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for message")
	}

	// Read the response
	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	assert.NoError(t, err, "Failed to read response")

	response := string(buf[:n])
	assert.Equal(t, "Response from unreachable2", response, "Unexpected response")
	t.Log("Successfully received response from unreachable2")
}
