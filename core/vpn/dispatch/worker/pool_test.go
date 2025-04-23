package worker

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/testutil"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

func TestWorkerPool_GetOrCreateWorker(t *testing.T) {
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

	// Create worker pool config
	config := &WorkerPoolConfig{
		WorkerIdleTimeout:     300,
		WorkerCleanupInterval: 1 * time.Minute,
		WorkerBufferSize:      100,
	}

	// Create worker pool
	workerPool := NewWorkerPool(testPeerID, mockStreamManager, config, nil)

	// Start the worker pool
	workerPool.Start()
	defer workerPool.Stop()

	// Test getting a worker
	connKey := types.ConnectionKey("12345:192.168.1.1:80")
	worker, err := workerPool.GetOrCreateWorker(context.Background(), connKey, "192.168.1.1")

	// Assert no error
	assert.NoError(t, err)
	assert.NotNil(t, worker)
	assert.Equal(t, connKey, worker.ConnectionKey)
	assert.Equal(t, "192.168.1.1", worker.DestIP)
	assert.Equal(t, testPeerID, worker.PeerID)

	// Test getting the same worker again
	worker2, err := workerPool.GetOrCreateWorker(context.Background(), connKey, "192.168.1.1")

	// Assert no error and same worker
	assert.NoError(t, err)
	assert.Equal(t, worker, worker2)

	// Verify metrics
	metrics := workerPool.GetMetrics()
	assert.Equal(t, int64(1), metrics["workers_created"])
	assert.Equal(t, int64(1), metrics["active_workers"])
}

func TestWorkerPool_DispatchPacket(t *testing.T) {
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

	// Create worker pool config
	config := &WorkerPoolConfig{
		WorkerIdleTimeout:     300,
		WorkerCleanupInterval: 1 * time.Minute,
		WorkerBufferSize:      100,
	}

	// Create worker pool
	workerPool := NewWorkerPool(testPeerID, mockStreamManager, config, nil)

	// Start the worker pool
	workerPool.Start()
	defer workerPool.Stop()

	// Create test packet
	packet := &types.QueuedPacket{
		Ctx:    context.Background(),
		DestIP: "192.168.1.1",
		Data:   []byte{1, 2, 3, 4, 5},
		DoneCh: make(chan error, 1),
	}

	// Test dispatching a packet
	connKey := types.ConnectionKey("12345:192.168.1.1:80")
	err := workerPool.DispatchPacket(context.Background(), connKey, "192.168.1.1", packet)

	// Assert no error
	assert.NoError(t, err)

	// Wait for packet to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify metrics
	metrics := workerPool.GetMetrics()
	assert.Equal(t, int64(1), metrics["packets_handled"])
}

func TestWorkerPool_CleanupInactiveWorkers(t *testing.T) {
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

	// Create worker pool config with short timeout for testing
	config := &WorkerPoolConfig{
		WorkerIdleTimeout:     1, // 1 second
		WorkerCleanupInterval: 100 * time.Millisecond,
		WorkerBufferSize:      100,
	}

	// Create worker pool
	workerPool := NewWorkerPool(testPeerID, mockStreamManager, config, nil)

	// Start the worker pool
	workerPool.Start()
	defer workerPool.Stop()

	// Create a worker
	connKey := types.ConnectionKey("12345:192.168.1.1:80")
	worker, err := workerPool.GetOrCreateWorker(context.Background(), connKey, "192.168.1.1")

	// Assert no error
	assert.NoError(t, err)
	assert.NotNil(t, worker)

	// Verify worker count
	assert.Equal(t, 1, workerPool.GetWorkerCount())

	// Wait for cleanup to run
	time.Sleep(1500 * time.Millisecond)

	// Verify worker was removed
	assert.Equal(t, 0, workerPool.GetWorkerCount())

	// Verify metrics
	metrics := workerPool.GetMetrics()
	assert.Equal(t, int64(1), metrics["workers_created"])
	assert.Equal(t, int64(1), metrics["workers_removed"])
}
