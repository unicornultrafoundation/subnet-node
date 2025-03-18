package proxy

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

// **Setup P2P Nodes with Real Service**
func setupRealP2P(t *testing.T) (*Service, *apps.Service) {
	h1, err := libp2p.New()
	assert.NoError(t, err)

	h2, err := libp2p.New()
	assert.NoError(t, err)

	sender := &Service{
		PeerHost: h1,
	}
	receiver := &apps.Service{
		PeerHost: h2,
	}

	// Register real receiver handler
	receiver.PeerHost.SetStreamHandler(atypes.ProtocolProxyReverse, receiver.OnReverseRequestReceive)

	// **Manually Connect Peers**
	err = sender.PeerHost.Connect(context.Background(), peer.AddrInfo{
		ID:    receiver.PeerHost.ID(),
		Addrs: receiver.PeerHost.Addrs(),
	})
	assert.NoError(t, err)

	// Wait for connection establishment
	time.Sleep(500 * time.Millisecond)

	return sender, receiver
}

// **Test Real TCP Forwarding**
func TestRealTCPForwarding(t *testing.T) {
	sender, receiver := setupRealP2P(t)

	// Start a TCP echo server
	listener, err := net.Listen("tcp", "127.0.0.1:9999")
	assert.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		defer conn.Close()
		io.Copy(conn, conn) // Echo server
	}()

	// Open a real P2P stream
	stream, err := sender.PeerHost.NewStream(context.Background(), receiver.PeerHost.ID(), atypes.ProtocolProxyReverse)
	assert.NoError(t, err)
	defer stream.Close()

	// Send TCP metadata
	meta := "tcp:test-app:9999\n"
	_, err = stream.Write([]byte(meta))
	assert.NoError(t, err)

	// Send test data
	testData := "Hello, Real P2P TCP!"
	_, err = stream.Write([]byte(testData))
	assert.NoError(t, err)

	// Read response
	buf := make([]byte, len(testData))
	_, err = stream.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, testData, string(buf))
}

// **Test Real UDP Forwarding**
func TestRealUDPForwarding(t *testing.T) {
	sender, receiver := setupRealP2P(t)

	// Start a UDP echo server
	udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9998")
	udpConn, err := net.ListenUDP("udp", udpAddr)
	assert.NoError(t, err)
	defer udpConn.Close()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, addr, _ := udpConn.ReadFromUDP(buf)
			udpConn.WriteToUDP(buf[:n], addr) // Echo back
		}
	}()

	// Open a real P2P stream
	stream, err := sender.PeerHost.NewStream(context.Background(), receiver.PeerHost.ID(), atypes.ProtocolProxyReverse)
	assert.NoError(t, err)
	defer stream.Close()

	// Send UDP metadata
	meta := "udp:test-app:9998\n"
	_, err = stream.Write([]byte(meta))
	assert.NoError(t, err)

	// Send test data
	testData := "Hello, Real P2P UDP!"
	_, err = stream.Write([]byte(testData))
	assert.NoError(t, err)

	// Read response
	buf := make([]byte, len(testData))
	_, err = stream.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, testData, string(buf))
}
