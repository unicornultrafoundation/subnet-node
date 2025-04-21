package packet

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/packet/testutil"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

func TestWorkerCreationAndMethods(t *testing.T) {
	// Create a test peer ID
	peerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")

	// Create a mock stream
	mockStream := &testutil.TestStream{}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())

	// Create a mock pool service
	mockPoolService := testutil.NewMockPoolService(mockStream)

	// Create resilience managers for testing
	circuitBreakerMgr := resilience.NewCircuitBreakerManager(3, 10*time.Second, 2)
	retryManager := resilience.NewRetryManager(3, 100*time.Millisecond, 2*time.Second)

	// Create a worker
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockPoolService,
		ctx,
		cancel,
		100,
		circuitBreakerMgr,
		retryManager,
	)

	assert.NotNil(t, worker)
	assert.Equal(t, "192.168.1.1:80", worker.SyncKey)
	assert.Equal(t, "192.168.1.1", worker.DestIP)
	assert.Equal(t, peerID, worker.PeerID)
	assert.Equal(t, mockPoolService, worker.PoolService)
	// Cannot directly compare context or cancel functions
	assert.NotNil(t, worker.Ctx)
	assert.NotNil(t, worker.Cancel)
	assert.True(t, worker.Running)
	assert.Equal(t, int64(0), worker.GetPacketCount())
	assert.Equal(t, int64(0), worker.GetErrorCount())
}

func TestWorkerStartStop(t *testing.T) {
	// Create a test peer ID
	peerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")

	// Create a mock stream
	mockStream := &testutil.TestStream{}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())

	// Create a mock pool service
	mockPoolService := testutil.NewMockPoolService(mockStream)

	// Create resilience managers for testing
	circuitBreakerMgr := resilience.NewCircuitBreakerManager(3, 10*time.Second, 2)
	retryManager := resilience.NewRetryManager(3, 100*time.Millisecond, 2*time.Second)

	// Create a worker
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockPoolService,
		ctx,
		cancel,
		100,
		circuitBreakerMgr,
		retryManager,
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

	// We don't verify if the stream was closed
	// as the pool service handles stream lifecycle
}

func TestWorkerProcessPacket(t *testing.T) {
	// Create a test peer ID
	peerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")

	// Create a mock stream
	mockStream := &testutil.TestStream{}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a mock pool service
	mockPoolService := testutil.NewMockPoolService(mockStream)

	// Create resilience managers for testing
	circuitBreakerMgr := resilience.NewCircuitBreakerManager(3, 10*time.Second, 2)
	retryManager := resilience.NewRetryManager(3, 100*time.Millisecond, 2*time.Second)

	// Create a worker
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockPoolService,
		ctx,
		cancel,
		100,
		circuitBreakerMgr,
		retryManager,
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

func TestWorkerIsIdle(t *testing.T) {
	// Create a test peer ID
	peerID, _ := peer.Decode("12D3KooWJbJFaZ1dwpuu9tQY3ELCKCpWMJc1dSRVbcMwhiLiFq7q")

	// Create a mock stream
	mockStream := &testutil.TestStream{}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a mock pool service
	mockPoolService := testutil.NewMockPoolService(mockStream)

	// Create resilience managers for testing
	circuitBreakerMgr := resilience.NewCircuitBreakerManager(3, 10*time.Second, 2)
	retryManager := resilience.NewRetryManager(3, 100*time.Millisecond, 2*time.Second)

	// Create a worker
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockPoolService,
		ctx,
		cancel,
		100,
		circuitBreakerMgr,
		retryManager,
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
	mockStream := &testutil.TestStream{}

	// Create a worker context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a mock pool service
	mockPoolService := testutil.NewMockPoolService(mockStream)

	// Create resilience managers for testing
	circuitBreakerMgr := resilience.NewCircuitBreakerManager(3, 10*time.Second, 2)
	retryManager := resilience.NewRetryManager(3, 100*time.Millisecond, 2*time.Second)

	// Create a worker with a small buffer
	worker := NewWorker(
		"192.168.1.1:80",
		"192.168.1.1",
		peerID,
		mockPoolService,
		ctx,
		cancel,
		2, // Small buffer size for testing
		circuitBreakerMgr,
		retryManager,
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
