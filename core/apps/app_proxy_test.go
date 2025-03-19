package apps

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	dtypes "github.com/docker/docker/api/types"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/config"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/core/docker"
	"github.com/unicornultrafoundation/subnet-node/core/proxy"
)

// **Setup P2P Nodes with Real Service**
func setupRealP2P(t *testing.T) (*proxy.Service, *Service) {
	networkSetting := &dtypes.NetworkSettings{}
	networkSetting.IPAddress = "127.0.0.1"
	dockerClientMock := &docker.MockDockerClient{
		InspectResponse: dtypes.ContainerJSON{
			NetworkSettings: networkSetting,
		},
		InspectError: nil,
	}

	senderPeer, err := libp2p.New()
	assert.NoError(t, err)

	receiverPeer, err := libp2p.New()
	assert.NoError(t, err)

	senderCfg := &config.C{
		Settings: map[interface{}]interface{}{
			"proxy": map[interface{}]interface{}{
				"enable":  true,
				"peer_id": receiverPeer.ID().String(),
				"app_id":  1,
				"ports":   []string{"5001:9999/tcp", "5002:9998/udp"},
			},
		}}
	sender := proxy.New(senderPeer, senderPeer.ID(), senderCfg)

	receiver := &Service{
		PeerHost:     receiverPeer,
		dockerClient: dockerClientMock,
	}

	// Register real receiver handler
	receiver.PeerHost.SetStreamHandler(atypes.ProtocolProxyReverse, receiver.OnReverseRequestReceive)

	// **Manually Connect Peers**
	err = senderPeer.Connect(context.Background(), peer.AddrInfo{
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
	sender, _ := setupRealP2P(t)

	err := sender.Start(context.Background())
	assert.NoError(t, err)

	// Start a TCP echo server
	listener, err := net.Listen("tcp", "127.0.0.1:9999")
	assert.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		defer conn.Close()
		io.Copy(conn, conn) // Echo server
	}()

	// Connect to sender proxy at 127.0.0.1:5001
	conn, err := net.Dial("tcp", "127.0.0.1:5001")
	assert.NoError(t, err)
	defer conn.Close()

	// Send metadata
	meta := "tcp:1:9999\n"
	_, err = conn.Write([]byte(meta))
	assert.NoError(t, err)

	// Send test data
	testData := "Hello, Real P2P TCP!"
	_, err = conn.Write([]byte(testData))
	assert.NoError(t, err)

	// Read response
	buf := make([]byte, len(testData))
	_, err = conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, testData, string(buf))
}

// **Test Real UDP Forwarding**
func TestRealUDPForwarding(t *testing.T) {
	sender, _ := setupRealP2P(t)
	err := sender.Start(context.Background())
	assert.NoError(t, err)

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

	// Send UDP metadata and request
	udpClient, err := net.Dial("udp", "127.0.0.1:5002")
	assert.NoError(t, err)
	defer udpClient.Close()

	meta := "udp:1:9998\n"
	_, err = udpClient.Write([]byte(meta))
	assert.NoError(t, err)

	testData := "Hello, Real P2P UDP!"
	_, err = udpClient.Write([]byte(testData))
	assert.NoError(t, err)

	buf := make([]byte, len(testData))
	n, err := udpClient.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, testData, string(buf[:n]))
}
