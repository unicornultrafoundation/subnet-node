package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

// This is a simple example of how to use the dispatch package.
// It creates a dispatcher and sends a packet to it.

func main() {
	// Set up logging
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Create a mock peer discovery service
	peerDiscovery := NewMockPeerDiscovery()

	// Create a mock stream service
	streamService := NewMockStreamService()

	// Create dispatcher config
	config := &dispatch.Config{
		MaxStreamsPerPeer:     10,
		StreamIdleTimeout:     5 * time.Minute,
		StreamCleanupInterval: 1 * time.Minute,
		WorkerIdleTimeout:     300, // seconds
		WorkerCleanupInterval: 1 * time.Minute,
		WorkerBufferSize:      100,
		PacketBufferSize:      100,
	}

	// Create dispatcher
	dispatcher := dispatch.NewDispatcher(peerDiscovery, streamService, config, nil)

	// Start the dispatcher
	dispatcher.Start()
	defer dispatcher.Stop()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a goroutine to send packets
	go func() {
		for i := 0; i < 10; i++ {
			// Create a packet
			packet := []byte{1, 2, 3, 4, 5}
			connKey := types.ConnectionKey(fmt.Sprintf("12345:192.168.1.%d:80", i%3+1))
			destIP := fmt.Sprintf("192.168.1.%d", i%3+1)

			// Dispatch the packet
			err := dispatcher.DispatchPacket(ctx, connKey, destIP, packet)
			if err != nil {
				logrus.WithError(err).Error("Failed to dispatch packet")
			}

			// Sleep to simulate packet arrival
			time.Sleep(100 * time.Millisecond)
		}

		// Print metrics
		logrus.Info("Metrics after sending packets:")
		for k, v := range dispatcher.GetMetrics() {
			logrus.Infof("%s: %d", k, v)
		}
	}()

	// Wait for signal or completion
	select {
	case <-sigChan:
		logrus.Info("Received signal, shutting down")
	case <-time.After(5 * time.Second):
		logrus.Info("Example completed")
	}
}

// MockPeerDiscovery is a mock implementation of the PeerDiscoveryService interface
type MockPeerDiscovery struct{}

func NewMockPeerDiscovery() *MockPeerDiscovery {
	return &MockPeerDiscovery{}
}

func (m *MockPeerDiscovery) GetPeerID(ctx context.Context, destIP string) (string, error) {
	// Return a mock peer ID based on the destination IP
	return "QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N", nil
}

func (m *MockPeerDiscovery) GetPeerIDByRegistry(ctx context.Context, destIP string) (string, error) {
	return m.GetPeerID(ctx, destIP)
}

func (m *MockPeerDiscovery) SyncPeerIDToDHT(ctx context.Context) error {
	return nil
}

func (m *MockPeerDiscovery) GetVirtualIP(ctx context.Context, peerID string) (string, error) {
	return "192.168.1.1", nil
}

func (m *MockPeerDiscovery) StoreMappingInDHT(ctx context.Context, peerID string) error {
	return nil
}

func (m *MockPeerDiscovery) VerifyVirtualIPHasRegistered(ctx context.Context, virtualIP string) error {
	return nil
}

// MockStreamService is a mock implementation of the StreamService interface
type MockStreamService struct{}

func NewMockStreamService() *MockStreamService {
	return &MockStreamService{}
}

func (m *MockStreamService) CreateNewVPNStream(ctx context.Context, peerID peer.ID) (api.VPNStream, error) {
	// Return a mock stream
	return NewMockStream(), nil
}

// MockStream is a mock implementation of the VPNStream interface
type MockStream struct{}

func NewMockStream() *MockStream {
	return &MockStream{}
}

func (m *MockStream) Read(p []byte) (n int, err error) {
	// Simulate reading data
	return len(p), nil
}

func (m *MockStream) Write(p []byte) (n int, err error) {
	// Simulate writing data
	logrus.Debugf("Mock stream writing %d bytes", len(p))
	return len(p), nil
}

func (m *MockStream) Close() error {
	// Simulate closing the stream
	logrus.Debug("Mock stream closed")
	return nil
}

func (m *MockStream) Reset() error {
	// Simulate resetting the stream
	logrus.Debug("Mock stream reset")
	return nil
}

func (m *MockStream) SetDeadline(t time.Time) error {
	// Simulate setting a deadline
	return nil
}

func (m *MockStream) SetReadDeadline(t time.Time) error {
	// Simulate setting a read deadline
	return nil
}

func (m *MockStream) SetWriteDeadline(t time.Time) error {
	// Simulate setting a write deadline
	return nil
}
