package pool

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

var streamLog = logrus.WithField("service", "vpn-stream")

// StreamChannel represents a channel dedicated to a specific stream
type StreamChannel struct {
	// The actual stream
	Stream api.VPNStream
	// Channel for sending packets to the stream
	PacketChan chan *types.QueuedPacket
	// Last activity time (Unix nano time)
	lastActivity int64
	// Whether the stream is healthy (0 = unhealthy, 1 = healthy)
	healthy int32
	// Metrics for the stream
	Metrics types.StreamMetrics
	// Context for the stream
	ctx context.Context
	// Cancel function for the stream context
	cancel context.CancelFunc
	// Last failed packet (to be retried when stream is recreated)
	lastFailedPacket *types.QueuedPacket
}

// GetBufferUtilization returns the current buffer utilization as a percentage (0-100)
func (s *StreamChannel) GetBufferUtilization() int {
	// Get the current length of the packet channel
	currentLen := len(s.PacketChan)

	// Get the capacity of the packet channel
	capacity := cap(s.PacketChan)

	// Calculate utilization percentage
	if capacity == 0 {
		return 0
	}

	return (currentLen * 100) / capacity
}

// IsHealthy returns whether the stream is healthy
func (s *StreamChannel) IsHealthy() bool {
	return atomic.LoadInt32(&s.healthy) == 1
}

// SetHealthy sets the stream's health status
func (s *StreamChannel) SetHealthy(healthy bool) {
	var value int32 = 0
	if healthy {
		value = 1
	}
	atomic.StoreInt32(&s.healthy, value)
}

// GetLastActivity returns the timestamp of the last activity
func (s *StreamChannel) GetLastActivity() time.Time {
	return time.Unix(0, atomic.LoadInt64(&s.lastActivity))
}

// UpdateLastActivity updates the last activity timestamp
func (s *StreamChannel) UpdateLastActivity() {
	atomic.StoreInt64(&s.lastActivity, time.Now().UnixNano())
}

// GetLastFailedPacket returns the last packet that failed to be sent
func (s *StreamChannel) GetLastFailedPacket() *types.QueuedPacket {
	return s.lastFailedPacket
}

// Close closes the stream channel
func (s *StreamChannel) Close() {
	s.cancel()
	close(s.PacketChan)
	s.Stream.Close()
}

// ProcessPackets starts processing packets from the channel
func (s *StreamChannel) ProcessPackets(peerID string) {
	logger := streamLog.WithFields(logrus.Fields{
		"peer_id": peerID,
	})

	// Add panic recovery to prevent stream processor crashes
	defer func() {
		if r := recover(); r != nil {
			logger.WithField("panic", r).Error("Recovered from panic in stream processor")
		}
		// Mark the stream as unhealthy
		s.SetHealthy(false)
		logger.Debug("Stream processor stopped")
	}()

	// Track stream statistics
	var stats struct {
		packetsProcessed int64
		packetsDropped   int64
		errors           int64
		lastStatsReport  time.Time
	}
	stats.lastStatsReport = time.Now()

	// Create a ticker for periodic stats reporting
	statsTicker := time.NewTicker(1 * time.Minute)
	defer statsTicker.Stop()

	// Set the stream as healthy initially
	s.SetHealthy(true)

	// Track consecutive errors
	consecutiveErrors := 0
	maxConsecutiveErrors := 3 // After this many consecutive errors, mark stream as unhealthy

	for {
		select {
		case <-s.ctx.Done():
			logger.Debug("Stream context cancelled")
			return

		case <-statsTicker.C:
			// Report stream statistics periodically
			if atomic.LoadInt64(&stats.packetsProcessed) > 0 || atomic.LoadInt64(&stats.errors) > 0 {
				logger.WithFields(logrus.Fields{
					"packets_processed": atomic.LoadInt64(&stats.packetsProcessed),
					"packets_dropped":   atomic.LoadInt64(&stats.packetsDropped),
					"errors":            atomic.LoadInt64(&stats.errors),
					"duration":          time.Since(stats.lastStatsReport).String(),
					"buffer_util":       s.GetBufferUtilization(),
					"healthy":           s.IsHealthy(),
				}).Info("Stream statistics")

				// Reset stats
				atomic.StoreInt64(&stats.packetsProcessed, 0)
				atomic.StoreInt64(&stats.packetsDropped, 0)
				atomic.StoreInt64(&stats.errors, 0)
				stats.lastStatsReport = time.Now()
			}

		case packet, ok := <-s.PacketChan:
			if !ok {
				logger.Debug("Stream packet channel closed")
				return
			}

			// Update last activity time
			s.UpdateLastActivity()

			// Process the packet
			err := s.writePacketToStream(packet)
			if err != nil {
				logger.WithError(err).Debug("Failed to write packet to stream")

				// Signal the error on the done channel if provided
				if packet.DoneCh != nil {
					// Use a non-blocking send to avoid panics if the channel is closed
					select {
					case packet.DoneCh <- err:
						// Successfully sent the error
					default:
						// Channel might be closed or full, don't panic
						logger.Debug("Could not send error to done channel, it might be closed or full")
					}
				}

				// Track consecutive errors
				consecutiveErrors++
				atomic.AddInt64(&stats.errors, 1)
				atomic.AddInt64(&stats.packetsDropped, 1)

				// Increment error count
				atomic.AddInt64(&s.Metrics.ErrorCount, 1)

				// If we've had too many consecutive errors, mark the stream as unhealthy and exit
				if consecutiveErrors >= maxConsecutiveErrors {
					logger.WithField("consecutive_errors", consecutiveErrors).Warn("Too many consecutive errors, marking stream as unhealthy")
					s.SetHealthy(false)
					// Note: We don't return here immediately to allow the failed packet to be stored
					// The stream will be replaced and the failed packet will be retried
					if s.lastFailedPacket != nil {
						logger.WithField("dest_ip", s.lastFailedPacket.DestIP).Info("Stream marked unhealthy with failed packet to retry")
					}
					return
				}
			} else {
				// Reset consecutive errors on success
				consecutiveErrors = 0

				// Signal success on the done channel if provided
				if packet.DoneCh != nil {
					select {
					case packet.DoneCh <- nil:
						// Successfully sent the result
					default:
						// Channel might be closed or full, don't panic
						logger.Debug("Could not send success to done channel, it might be closed or full")
					}
				}

				// Update metrics
				atomic.AddInt64(&s.Metrics.PacketCount, 1)
				atomic.AddInt64(&s.Metrics.BytesSent, int64(len(packet.Data)))
				atomic.AddInt64(&stats.packetsProcessed, 1)
			}
		}
	}
}

// writePacketToStream writes a packet to the stream
func (s *StreamChannel) writePacketToStream(packet *types.QueuedPacket) error {
	// Record start time for performance monitoring
	startTime := time.Now()

	// Write the packet to the stream
	_, err := s.Stream.Write(packet.Data)
	if err != nil {
		// Store the failed packet for retry when stream is recreated
		s.lastFailedPacket = packet

		// Check for common libp2p stream errors that indicate the stream is broken
		errStr := err.Error()
		if strings.Contains(errStr, "stream reset") ||
			strings.Contains(errStr, "protocol not supported") ||
			strings.Contains(errStr, "connection closed") ||
			strings.Contains(errStr, "stream closed") ||
			strings.Contains(errStr, "deadline exceeded") ||
			strings.Contains(errStr, "EOF") {
			// Mark the stream as unhealthy
			s.SetHealthy(false)
			streamLog.WithError(err).Debug("Stream marked as unhealthy due to libp2p error")
		}
		return types.NewNetworkError(err, "write", packet.DestIP, "")
	}

	// Clear any previously stored failed packet on success
	s.lastFailedPacket = nil

	// Calculate processing duration
	processingDuration := time.Since(startTime)

	// Log performance information if write took too long
	if processingDuration > 200*time.Millisecond {
		streamLog.WithFields(logrus.Fields{
			"stream_id":   fmt.Sprintf("%p", s),
			"dest_ip":     packet.DestIP,
			"duration_ms": processingDuration.Milliseconds(),
		}).Warn("Stream write took longer than expected")
	}

	return nil
}
