package packet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

// MockErrorStream is a mock stream that returns errors on write
type MockErrorStream struct {
	MockStream
	errorOnWrite bool
	writeCount   int
	errorType    string // "connection_reset", "stream_closed", "temporary", "other"
}

func (m *MockErrorStream) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.writeCount++

	if m.errorOnWrite {
		switch m.errorType {
		case "connection_reset":
			return 0, errors.New("connection reset by peer")
		case "stream_closed":
			return 0, errors.New("stream closed")
		case "temporary":
			// Create a temporary error
			return 0, &tempError{}
		default:
			return 0, errors.New("write error")
		}
	}

	return len(p), nil
}

// tempError implements the net.Error interface with Temporary() returning true
type tempError struct{}

func (e *tempError) Error() string   { return "temporary error" }
func (e *tempError) Timeout() bool   { return false }
func (e *tempError) Temporary() bool { return true }

func TestWorkerCreationAndMethods(t *testing.T) {
	// Create a test peer ID
	peerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")

	// Create a mock stream
	mockStream := &MockStream{}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())

	// Create a worker
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockStream,
		ctx,
		cancel,
		100,
	)

	assert.NotNil(t, worker)
	assert.Equal(t, "192.168.1.1:80", worker.SyncKey)
	assert.Equal(t, "192.168.1.1", worker.DestIP)
	assert.Equal(t, peerID, worker.PeerID)
	assert.Equal(t, mockStream, worker.Stream)
	// Cannot directly compare context or cancel functions
	assert.NotNil(t, worker.Ctx)
	assert.NotNil(t, worker.Cancel)
	assert.True(t, worker.Running)
	assert.Equal(t, int64(0), worker.PacketCount)
	assert.Equal(t, int64(0), worker.ErrorCount)
}

func TestWorkerStartStop(t *testing.T) {
	// Create a test peer ID
	peerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")

	// Create a mock stream
	mockStream := &MockStream{}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())

	// Create a worker
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockStream,
		ctx,
		cancel,
		100,
	)

	// Start the worker
	worker.Start()

	// Verify the worker is running
	assert.True(t, worker.Running)

	// Stop the worker
	worker.Stop()

	// Wait for the worker to stop
	time.Sleep(100 * time.Millisecond)

	// Verify the worker has stopped
	assert.False(t, worker.Running)

	// Verify the stream was closed
	assert.True(t, mockStream.closed)
}

func TestWorkerProcessPacket(t *testing.T) {
	// Create a test peer ID
	peerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")

	// Create a mock stream
	mockStream := &MockStream{}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a worker
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockStream,
		ctx,
		cancel,
		100,
	)

	// Start the worker
	worker.Start()

	// Create a test packet
	packet := &QueuedPacket{
		Ctx:    context.Background(),
		DestIP: "192.168.1.1",
		Data:   []byte("test packet"),
		DoneCh: make(chan error, 1),
	}

	// Enqueue the packet
	success := worker.EnqueuePacket(packet)
	assert.True(t, success)

	// Wait for the packet to be processed
	err := <-packet.DoneCh
	assert.NoError(t, err)

	// Verify the packet count was incremented
	assert.Equal(t, int64(1), worker.GetPacketCount())
	assert.Equal(t, int64(0), worker.GetErrorCount())
}

func TestWorkerProcessPacketError(t *testing.T) {
	// Create a test peer ID
	peerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")

	// Create a mock stream that returns errors
	mockStream := &MockErrorStream{
		errorOnWrite: true,
		errorType:    "other",
	}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a worker
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockStream,
		ctx,
		cancel,
		100,
	)

	// Start the worker
	worker.Start()

	// Create a test packet
	packet := &QueuedPacket{
		Ctx:    context.Background(),
		DestIP: "192.168.1.1",
		Data:   []byte("test packet"),
		DoneCh: make(chan error, 1),
	}

	// Enqueue the packet
	success := worker.EnqueuePacket(packet)
	assert.True(t, success)

	// Wait for the packet to be processed
	err := <-packet.DoneCh
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error writing to P2P stream")

	// Verify the error count was incremented
	assert.Equal(t, int64(1), worker.GetPacketCount())
	assert.Equal(t, int64(1), worker.GetErrorCount())
}

func TestWorkerIsIdle(t *testing.T) {
	// Create a test peer ID
	peerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")

	// Create a mock stream
	mockStream := &MockStream{}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a worker
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockStream,
		ctx,
		cancel,
		100,
	)

	// Set the last activity time to 2 seconds ago
	worker.LastActivity = time.Now().Add(-2 * time.Second)

	// Check if the worker is idle with a 1 second timeout
	assert.True(t, worker.IsIdle(1*time.Second))

	// Check if the worker is idle with a 3 second timeout
	assert.False(t, worker.IsIdle(3*time.Second))

	// Update the last activity time
	worker.UpdateLastActivity()

	// Check if the worker is idle with a 1 second timeout
	assert.False(t, worker.IsIdle(1*time.Second))
}

func TestWorkerEnqueuePacket(t *testing.T) {
	// Create a test peer ID
	peerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")

	// Create a mock stream
	mockStream := &MockStream{}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a worker with a small buffer
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockStream,
		ctx,
		cancel,
		2, // Small buffer size for testing
	)

	// Create test packets
	packet1 := &QueuedPacket{
		Ctx:    context.Background(),
		DestIP: "192.168.1.1",
		Data:   []byte("test packet 1"),
	}

	packet2 := &QueuedPacket{
		Ctx:    context.Background(),
		DestIP: "192.168.1.1",
		Data:   []byte("test packet 2"),
	}

	packet3 := &QueuedPacket{
		Ctx:    context.Background(),
		DestIP: "192.168.1.1",
		Data:   []byte("test packet 3"),
	}

	// Enqueue the first packet
	success := worker.EnqueuePacket(packet1)
	assert.True(t, success)

	// Enqueue the second packet
	success = worker.EnqueuePacket(packet2)
	assert.True(t, success)

	// Enqueue the third packet (should fail because the buffer is full)
	success = worker.EnqueuePacket(packet3)
	assert.False(t, success)
}
