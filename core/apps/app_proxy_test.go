package apps

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
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

// **Test: Multiple TCP Forwarding Cases**
func TestMultipleTCPForwarding(t *testing.T) {
	portMappings := []string{
		fmt.Sprintf("%d:%d/tcp", findAvailablePort(), findAvailablePort()),
		fmt.Sprintf("%d:%d/tcp", findAvailablePort(), findAvailablePort()),
		fmt.Sprintf("%d:%d/tcp", findAvailablePort(), findAvailablePort()),
	}

	sender, _ := setupRealP2P(t, portMappings)
	ctx := context.Background()
	if err := sender.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer sender.Stop(ctx)

	var wg sync.WaitGroup

	for _, mapping := range portMappings {
		parts := parsePortMapping(mapping)

		wg.Add(1)
		go func(proxyPort, targetPort int) {
			defer wg.Done()
			testTCPEcho(t, parts)
		}(parts.proxyPort, parts.targetPort)
	}

	wg.Wait()
}

func TestMultipleClients(t *testing.T) {
	// Start Proxy server
	portMapping := fmt.Sprintf("%d:%d/tcp", findAvailablePort(), findAvailablePort())
	sender, _ := setupRealP2P(t, []string{portMapping})
	ctx := context.Background()
	if err := sender.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer sender.Stop(ctx)

	parts := parsePortMapping(portMapping)
	numClients := 10
	var wg sync.WaitGroup

	// Start echo server
	listener, err := startTCPEchoServer(parts.targetPort)
	if err != nil {
		t.Fatalf("Failed to start echo server: %v", err)
	}
	defer listener.Close()

	// Wait until the server is ready
	waitForServers(t, parts)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			testData := fmt.Sprintf("Client %d: Hello!", clientID)
			conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", parts.proxyPort))
			if err != nil {
				t.Fatalf("Client %d failed to connect: %v", clientID, err)
			}
			defer conn.Close()

			// Send data
			time.Sleep(50 * time.Millisecond) // Temporary delay, we’ll replace this
			if _, err := conn.Write([]byte(testData)); err != nil {
				t.Fatalf("Client %d failed to send data: %v", clientID, err)
			}

			// Receive response
			response, err := waitForResponse(conn, len(testData), 5*time.Second)
			if err != nil {
				t.Fatalf("Client %d failed to receive response: %v", clientID, err)
			}

			// Validate response
			if string(response) != testData {
				t.Fatalf("Client %d: expected '%s', got '%s'", clientID, testData, string(response))
			}
		}(i)
	}

	wg.Wait()
}

func TestLargeDataTransfer(t *testing.T) {
	// Start Proxy server
	portMapping := fmt.Sprintf("%d:%d/tcp", findAvailablePort(), findAvailablePort())
	sender, _ := setupRealP2P(t, []string{portMapping})
	ctx := context.Background()
	if err := sender.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer sender.Stop(ctx)

	parts := parsePortMapping(portMapping)

	// Start Echo server
	listener, err := startTCPEchoServer(parts.targetPort)
	if err != nil {
		t.Fatalf("Failed to start echo server: %v", err)
	}
	defer listener.Close()

	// Wait until the server is ready
	waitForServers(t, parts)

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", parts.proxyPort))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Generate a large payload (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256) // Fill with repeating byte pattern
	}

	// Send large data
	time.Sleep(50 * time.Millisecond) // Temporary delay, we’ll replace this
	if _, err := conn.Write(largeData); err != nil {
		t.Fatal(err)
	}

	// Receive large response
	response, err := waitForResponse(conn, len(largeData), 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Validate response integrity
	if string(response) != string(largeData) {
		t.Fatalf("Data mismatch: expected %d bytes, got %d bytes", len(largeData), len(response))
	}
}

func TestProxyRestart(t *testing.T) {
	portMapping := fmt.Sprintf("%d:%d/tcp", findAvailablePort(), findAvailablePort())
	sender, _ := setupRealP2P(t, []string{portMapping})
	ctx := context.Background()
	if err := sender.Start(ctx); err != nil {
		t.Fatal(err)
	}

	parts := parsePortMapping(portMapping)

	// Start echo server
	listener, err := startTCPEchoServer(parts.targetPort)
	if err != nil {
		t.Fatalf("Failed to start echo server: %v", err)
	}
	defer listener.Close()

	// Wait until the server is ready
	waitForServers(t, parts)

	// Establish connection
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", parts.proxyPort))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Restart proxy service
	sender.Stop(ctx)
	time.Sleep(5 * time.Second) // Allow clean shutdown
	if err := sender.Start(ctx); err != nil {
		t.Fatal(err)
	}

	// Wait for proxy to be back online
	if !waitForPort("127.0.0.1", parts.proxyPort, 5*time.Second) {
		t.Fatal("Proxy did not recover after restart")
	}

	// Send message after restart
	testData := "Hello after restart!"
	time.Sleep(50 * time.Millisecond) // Temporary delay, we’ll replace this
	if _, err := conn.Write([]byte(testData)); err != nil {
		t.Fatal(err)
	}

	// Receive response
	response, err := waitForResponse(conn, len(testData), 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Validate response
	if string(response) != testData {
		t.Fatalf("Expected '%s', got '%s'", testData, string(response))
	}
}

func TestSlowReceiver(t *testing.T) {
	portMapping := fmt.Sprintf("%d:%d/tcp", findAvailablePort(), findAvailablePort())
	sender, _ := setupRealP2P(t, []string{portMapping})
	ctx := context.Background()
	if err := sender.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer sender.Stop(ctx)

	parts := parsePortMapping(portMapping)

	// Start a slow echo server
	listener, err := startTCPEchoServer(parts.targetPort)
	if err != nil {
		t.Fatalf("Failed to start echo server: %v", err)
	}
	defer listener.Close()

	// Wait until the server is ready
	waitForServers(t, parts)

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", parts.proxyPort))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	testData := "Slow test!"
	time.Sleep(50 * time.Millisecond) // Temporary delay, we’ll replace this
	if _, err := conn.Write([]byte(testData)); err != nil {
		t.Fatal(err)
	}

	// Simulate slow response
	time.Sleep(3 * time.Second)

	// Receive response
	response, err := waitForResponse(conn, len(testData), 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Validate response
	if string(response) != testData {
		t.Fatalf("Expected '%s', got '%s'", testData, string(response))
	}
}

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

// **Helper function to test TCP Echo**
func testTCPEcho(t *testing.T, parts portMapping) {
	proxyPort := parts.proxyPort
	targetPort := parts.targetPort

	// Start the echo server
	listener, err := startTCPEchoServer(targetPort)
	if err != nil {
		t.Fatalf("Failed to start echo server: %v", err)
	}
	defer listener.Close()

	// Wait until the server is ready
	waitForServers(t, parts)

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", proxyPort))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond) // Temporary delay, we’ll replace this
	testData := "Hello, P2P!"
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

func waitForServers(t *testing.T, parts portMapping) {
	if !waitForPort("127.0.0.1", parts.targetPort, 5*time.Second) {
		t.Fatalf("Echo server did not start on port %d", parts.targetPort)
	}
	if !waitForPort("127.0.0.1", parts.proxyPort, 5*time.Second) {
		t.Fatalf("Proxy server did not start on port %d", parts.proxyPort)
	}
}
