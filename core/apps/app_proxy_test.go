package apps

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	dtypes "github.com/docker/docker/api/types"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/config"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/core/docker"
	"github.com/unicornultrafoundation/subnet-node/core/proxy"
)

// **Find Available Port**
func findAvailablePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

// **Setup P2P Nodes with Multiple Ports**
func setupRealP2P(t *testing.T, portMappings []string) (*proxy.Service, *Service) {
	networkSetting := &dtypes.NetworkSettings{}
	networkSetting.IPAddress = "127.0.0.1"
	dockerClientMock := &docker.MockDockerClient{
		InspectResponse: dtypes.ContainerJSON{
			NetworkSettings: networkSetting,
		},
		InspectError: nil,
	}

	senderPeer, err := libp2p.New()
	if err != nil {
		t.Fatal(err)
	}

	receiverPeer, err := libp2p.New()
	if err != nil {
		t.Fatal(err)
	}

	senderCfg := &config.C{
		Settings: map[interface{}]interface{}{
			"proxy": map[interface{}]interface{}{
				"enable":  true,
				"peer_id": receiverPeer.ID().String(),
				"app_id":  1,
				"ports":   portMappings,
			},
		},
	}
	sender := proxy.New(senderPeer, senderPeer.ID(), senderCfg)

	receiver := &Service{
		PeerHost:     receiverPeer,
		dockerClient: dockerClientMock,
	}

	receiver.PeerHost.SetStreamHandler(atypes.ProtocolProxyReverse, receiver.OnReverseRequestReceive)

	err = senderPeer.Connect(context.Background(), peer.AddrInfo{
		ID:    receiver.PeerHost.ID(),
		Addrs: receiver.PeerHost.Addrs(),
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	return sender, receiver
}

// **Test: Multiple TCP Forwarding Cases**
func TestMultipleTCPForwarding(t *testing.T) {
	portMappings := []string{
		fmt.Sprintf("%d:%d/tcp", findAvailablePort(), findAvailablePort()),
		// fmt.Sprintf("%d:%d/tcp", findAvailablePort(), findAvailablePort()),
		// fmt.Sprintf("%d:%d/tcp", findAvailablePort(), findAvailablePort()),
	}

	sender, _ := setupRealP2P(t, portMappings)
	ctx := context.Background()
	if err := sender.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer sender.Stop(ctx)

	// var wg sync.WaitGroup

	for _, mapping := range portMappings {
		parts := parsePortMapping(mapping)

		// wg.Add(1)
		// go func(proxyPort, targetPort int) {
		// 	defer wg.Done()
		testTCPEcho(t, parts.proxyPort, parts.targetPort)
		// 	}(parts.proxyPort, parts.targetPort)
		// }
	}
	// wg.Wait()
}

// **Helper function to test TCP Echo**
func testTCPEcho(t *testing.T, proxyPort, targetPort int) {
	// Start the echo server
	listener, err := startTCPEchoServer(targetPort)
	if err != nil {
		t.Fatalf("Failed to start echo server: %v", err)
	}
	defer listener.Close()

	// Wait until the server is ready
	if !waitForPort("127.0.0.1", targetPort, 5*time.Second) {
		t.Fatalf("Echo server did not start on port %d", targetPort)
	}
	if !waitForPort("127.0.0.1", proxyPort, 5*time.Second) {
		t.Fatalf("Proxy server did not start on port %d", proxyPort)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", proxyPort))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	testData := "Hello, P2P!"
	// ✅ Check if connection is fully established before writing
	log.Println("Connected to proxy, waiting for readiness...")

	time.Sleep(1 * time.Second) // Temporary delay, we’ll replace this

	log.Println("Proxy should be ready, sending data...")
	if _, err := conn.Write([]byte(testData)); err != nil {
		t.Fatal(err)
	}

	// **Ensure we wait for a response before closing**
	response, err := waitForResponse(conn, len(testData), 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if string(response) != testData {
		t.Fatalf("Expected %s, got %s", testData, string(response))
	}
}

// Reads full response before closing connection
func waitForResponse(conn net.Conn, expectedLen int, timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)
	conn.SetReadDeadline(deadline) // Prevent indefinite blocking

	buf := make([]byte, expectedLen)
	totalRead := 0

	for totalRead < expectedLen {
		n, err := conn.Read(buf[totalRead:])
		if err != nil {
			return nil, err // Handle read errors
		}
		totalRead += n
	}
	return buf, nil
}

// StartEchoServer initializes an echo server on the given port.
func startTCPEchoServer(port int) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					continue
				}
				return // Exit if the listener is closed
			}
			go handleConnection(conn)
		}
	}()
	return listener, nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024) // Buffer to read data

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break // Client closed connection
			}
			return // Handle other errors
		}
		conn.Write(buf[:n]) // Echo back the received data
	}
}

// **Helper function to parse "proxyPort:targetPort/protocol"**
type portMapping struct {
	proxyPort  int
	targetPort int
	protocol   string
}

func parsePortMapping(mapping string) portMapping {
	var p portMapping
	fmt.Sscanf(mapping, "%d:%d/%s", &p.proxyPort, &p.targetPort, &p.protocol)
	return p
}

// **Helper function to wait for a port to be available**
func waitForPort(host string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}
