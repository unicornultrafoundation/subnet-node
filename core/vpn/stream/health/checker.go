package health

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/metrics"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/pool"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/stream/types"
)

var log = logrus.WithField("service", "vpn-stream-health")

// HealthChecker checks the health of streams in the pool
type HealthChecker struct {
	// Stream pool manager
	poolManager *pool.StreamPoolManager
	// Health check interval
	healthCheckInterval time.Duration
	// Health check timeout
	healthCheckTimeout time.Duration
	// Maximum consecutive failures
	maxConsecutiveFailures int
	// Context for the checker
	ctx context.Context
	// Cancel function for the checker context
	cancel context.CancelFunc
	// Mutex to protect access to the checker
	mu sync.Mutex
	// Whether the checker is running
	running bool
	// Channel to signal checker shutdown
	stopChan chan struct{}
	// Metrics for the health checker
	metrics *metrics.HealthMetrics
	// Map of peer ID to consecutive failures
	consecutiveFailures map[string]int
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(
	poolManager *pool.StreamPoolManager,
	healthCheckInterval time.Duration,
	healthCheckTimeout time.Duration,
	maxConsecutiveFailures int,
) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())

	return &HealthChecker{
		poolManager:            poolManager,
		healthCheckInterval:    healthCheckInterval,
		healthCheckTimeout:     healthCheckTimeout,
		maxConsecutiveFailures: maxConsecutiveFailures,
		ctx:                    ctx,
		cancel:                 cancel,
		running:                false,
		stopChan:               make(chan struct{}),
		metrics:                metrics.NewHealthMetrics(),
		consecutiveFailures:    make(map[string]int),
	}
}

// Start starts the health checker
func (c *HealthChecker) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return
	}

	c.running = true

	// Start the health check goroutine
	go c.runHealthCheck()

	log.Infof("Health checker started with interval %s", c.healthCheckInterval)
}

// Stop stops the health checker
func (c *HealthChecker) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	c.running = false

	// Cancel the context
	c.cancel()

	// Signal the health check goroutine to stop
	close(c.stopChan)

	log.Info("Health checker stopped")
}

// runHealthCheck runs the health check loop
func (c *HealthChecker) runHealthCheck() {
	ticker := time.NewTicker(c.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check the health of streams
			c.checkHealth()
		case <-c.ctx.Done():
			return
		case <-c.stopChan:
			return
		}
	}
}

// checkHealth checks the health of streams
func (c *HealthChecker) checkHealth() {
	// Update metrics
	c.metrics.IncrementChecksPerformed()

	// Get a list of peers from the pool
	peers := c.poolManager.GetAllPeers()

	// Check each peer's streams
	for _, peerID := range peers {
		// Create a context with timeout for this health check
		ctx, cancel := context.WithTimeout(c.ctx, c.healthCheckTimeout)

		// Get a stream from the pool
		stream, err := c.poolManager.GetStream(ctx, peerID)
		if err != nil {
			log.Debugf("Failed to get stream for health check: %v", err)
			// Handle the unhealthy peer
			c.handleUnhealthyStream(peerID)
			cancel()
			continue
		}

		// Check the health of the stream
		healthy := c.checkStreamHealth(peerID, stream)

		// Release the stream back to the pool
		c.poolManager.ReleaseStream(peerID, stream, healthy)

		// Clean up the context
		cancel()
	}
}

// GetMetrics returns the current metrics
func (c *HealthChecker) GetMetrics() map[string]int64 {
	return map[string]int64{
		"checks_performed":  c.metrics.ChecksPerformed,
		"healthy_streams":   c.metrics.HealthyStreams,
		"unhealthy_streams": c.metrics.UnhealthyStreams,
	}
}

// StartHealthChecker starts the health checker (for HealthCheckerService interface)
func (c *HealthChecker) StartHealthChecker() {
	c.Start()
}

// StopHealthChecker stops the health checker (for HealthCheckerService interface)
func (c *HealthChecker) StopHealthChecker() {
	c.Stop()
}

// Ensure HealthChecker implements the required interfaces
var (
	_ types.HealthCheckerService   = (*HealthChecker)(nil)
	_ types.StreamLifecycleService = (*HealthChecker)(nil)
)

// checkStreamHealth checks the health of a stream
func (c *HealthChecker) checkStreamHealth(peerID peer.ID, s types.VPNStream) bool {
	peerIDStr := peerID.String()

	// Create a cancel function to clean up resources
	_, cancel := context.WithTimeout(c.ctx, c.healthCheckTimeout)
	defer cancel()

	// Set a deadline for the stream
	deadline := time.Now().Add(c.healthCheckTimeout)
	err := s.SetDeadline(deadline)
	if err != nil {
		log.Errorf("Failed to set deadline for stream: %v", err)
		return false
	}

	// Implement actual health check logic
	// Create a health check packet (a simple ping)
	pingPacket := []byte("PING")

	// Send the ping packet
	_, err = s.Write(pingPacket)
	if err != nil {
		log.Debugf("Failed to write health check packet: %v", err)
		return false
	}

	// Read the response
	responseBuffer := make([]byte, 4) // Expecting "PONG"
	_, err = s.Read(responseBuffer)
	if err != nil {
		log.Debugf("Failed to read health check response: %v", err)
		return false
	}

	// Check if the response is valid
	response := string(responseBuffer)
	if response != "PONG" {
		log.Debugf("Invalid health check response: %s", response)
		return false
	}

	// Reset the deadline
	err = s.SetDeadline(time.Time{})
	if err != nil {
		log.Errorf("Failed to reset deadline for stream: %v", err)
		return false
	}

	// Update consecutive failures
	c.mu.Lock()
	defer c.mu.Unlock()

	// Reset consecutive failures
	c.consecutiveFailures[peerIDStr] = 0

	// Update metrics
	c.metrics.IncrementHealthyStreams()

	return true
}

// handleUnhealthyStream handles an unhealthy stream
func (c *HealthChecker) handleUnhealthyStream(peerID peer.ID) {
	peerIDStr := peerID.String()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Increment consecutive failures
	c.consecutiveFailures[peerIDStr]++

	// Update metrics
	c.metrics.IncrementUnhealthyStreams()

	// Check if we've reached the maximum consecutive failures
	if c.consecutiveFailures[peerIDStr] >= c.maxConsecutiveFailures {
		log.Warnf("Peer %s has exceeded maximum consecutive failures (%d)", peerIDStr, c.maxConsecutiveFailures)

		// Reset consecutive failures
		c.consecutiveFailures[peerIDStr] = 0

		// Implement logic to handle a peer with too many consecutive failures
		// Close all streams to the peer to force new connections
		go c.closeAllStreamsForPeer(peerID)
	}
}

// closeAllStreamsForPeer closes all streams for a peer
func (c *HealthChecker) closeAllStreamsForPeer(peerID peer.ID) {
	// Get the number of streams for this peer
	streamCount := c.poolManager.GetStreamCount(peerID)

	// If there are no streams, there's nothing to do
	if streamCount == 0 {
		return
	}

	log.Infof("Closing all %d streams for peer %s due to health check failures", streamCount, peerID.String())

	// Create a context with timeout for getting streams
	ctx, cancel := context.WithTimeout(c.ctx, c.healthCheckTimeout)
	defer cancel()

	// Get and release all streams as unhealthy
	for i := 0; i < streamCount; i++ {
		stream, err := c.poolManager.GetStream(ctx, peerID)
		if err != nil {
			log.Debugf("Failed to get stream for closing: %v", err)
			continue
		}

		// Release the stream as unhealthy, which will close it
		c.poolManager.ReleaseStream(peerID, stream, false)
	}

	// Update metrics
	c.metrics.IncrementPeerResets()
}
