package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/testutil"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/resilience"
)

func TestWorker_EnqueuePacket(t *testing.T) {
	// Create mock stream manager
	mockStreamManager := new(testutil.MockStreamManager)
	mockStreamManager.On("SendPacket", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockStreamManager.On("Start").Return()
	mockStreamManager.On("Stop").Return()
	mockStreamManager.On("GetMetrics").Return(map[string]int64{"test": 1})
	mockStreamManager.On("GetConnectionCount").Return(1)
	mockStreamManager.On("Close").Return(nil)

	// Create test peer ID
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create worker context
	ctx, cancel := context.WithCancel(context.Background())

	// Create worker
	worker := NewWorker(
		types.ConnectionKey("12345:192.168.1.1:80"),
		"192.168.1.1",
		testPeerID,
		mockStreamManager,
		ctx,
		cancel,
		10, // Small buffer for testing
		nil,
	)

	// Start the worker
	worker.Start()
	defer worker.Stop()

	// Create test packet
	packet := &types.QueuedPacket{
		Ctx:    context.Background(),
		DestIP: "192.168.1.1",
		Data:   []byte{1, 2, 3, 4, 5},
	}

	// Test enqueueing a packet
	success := worker.EnqueuePacket(packet)
	assert.True(t, success)

	// Fill the buffer
	for i := 0; i < 10; i++ {
		worker.EnqueuePacket(packet)
	}

	// Test enqueueing to a full buffer
	success = worker.EnqueuePacket(packet)
	assert.False(t, success)
}

func TestWorker_ProcessPacket(t *testing.T) {
	// Create mock stream manager
	mockStreamManager := new(testutil.MockStreamManager)
	mockStreamManager.On("SendPacket", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockStreamManager.On("Start").Return()
	mockStreamManager.On("Stop").Return()
	mockStreamManager.On("GetMetrics").Return(map[string]int64{"test": 1})
	mockStreamManager.On("GetConnectionCount").Return(1)
	mockStreamManager.On("Close").Return(nil)

	// Create test peer ID
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create worker context
	ctx, cancel := context.WithCancel(context.Background())

	// Create worker
	worker := NewWorker(
		types.ConnectionKey("12345:192.168.1.1:80"),
		"192.168.1.1",
		testPeerID,
		mockStreamManager,
		ctx,
		cancel,
		100,
		nil,
	)

	// Start the worker
	worker.Start()
	defer worker.Stop()

	// Create test packet with callback
	doneCh := make(chan error, 1)
	packet := &types.QueuedPacket{
		Ctx:    context.Background(),
		DestIP: "192.168.1.1",
		Data:   []byte{1, 2, 3, 4, 5},
		DoneCh: doneCh,
	}

	// Enqueue the packet
	success := worker.EnqueuePacket(packet)
	assert.True(t, success)

	// Wait for packet to be processed
	callbackErr := <-doneCh

	// Assert no error from callback
	assert.NoError(t, callbackErr)

	// Verify metrics
	metrics := worker.GetMetrics()
	assert.Equal(t, int64(1), metrics.PacketCount)
	assert.Equal(t, int64(5), metrics.BytesSent)
}

func TestWorker_ProcessPacket_Error(t *testing.T) {
	// Create mock stream manager
	mockStreamManager := new(testutil.MockStreamManager)
	mockStreamManager.On("SendPacket", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("send error"))
	mockStreamManager.On("Start").Return()
	mockStreamManager.On("Stop").Return()
	mockStreamManager.On("GetMetrics").Return(map[string]int64{"test": 1})
	mockStreamManager.On("GetConnectionCount").Return(1)
	mockStreamManager.On("Close").Return(nil)

	// Create test peer ID
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create worker context
	ctx, cancel := context.WithCancel(context.Background())

	// Create resilience service with no retries for testing
	resilienceConfig := &resilience.ResilienceConfig{
		RetryMaxAttempts: 1,
	}
	resilienceService := resilience.NewResilienceService(resilienceConfig)

	// Create worker
	worker := NewWorker(
		types.ConnectionKey("12345:192.168.1.1:80"),
		"192.168.1.1",
		testPeerID,
		mockStreamManager,
		ctx,
		cancel,
		100,
		resilienceService,
	)

	// Start the worker
	worker.Start()
	defer worker.Stop()

	// Create test packet with callback
	doneCh := make(chan error, 1)
	packet := &types.QueuedPacket{
		Ctx:    context.Background(),
		DestIP: "192.168.1.1",
		Data:   []byte{1, 2, 3, 4, 5},
		DoneCh: doneCh,
	}

	// Enqueue the packet
	success := worker.EnqueuePacket(packet)
	assert.True(t, success)

	// Wait for packet to be processed
	callbackErr := <-doneCh

	// Assert error from callback
	assert.Error(t, callbackErr)
	assert.Contains(t, callbackErr.Error(), "send error")

	// Verify metrics
	metrics := worker.GetMetrics()
	assert.Equal(t, int64(0), metrics.PacketCount)
	assert.Equal(t, int64(1), metrics.ErrorCount)
}

func TestWorker_Stop(t *testing.T) {
	// Create mock stream manager
	mockStreamManager := new(testutil.MockStreamManager)
	mockStreamManager.On("SendPacket", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockStreamManager.On("Start").Return()
	mockStreamManager.On("Stop").Return()
	mockStreamManager.On("GetMetrics").Return(map[string]int64{"test": 1})
	mockStreamManager.On("GetConnectionCount").Return(1)
	mockStreamManager.On("Close").Return(nil)

	// Create test peer ID
	testPeerID, _ := peer.Decode("QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")

	// Create worker context
	ctx, cancel := context.WithCancel(context.Background())

	// Create worker
	worker := NewWorker(
		types.ConnectionKey("12345:192.168.1.1:80"),
		"192.168.1.1",
		testPeerID,
		mockStreamManager,
		ctx,
		cancel,
		100,
		nil,
	)

	// Start the worker
	worker.Start()

	// Verify worker is running
	assert.True(t, worker.IsRunning())

	// Stop the worker
	worker.Stop()

	// Wait for worker to stop
	time.Sleep(100 * time.Millisecond)

	// Verify worker is stopped
	assert.False(t, worker.IsRunning())
}
