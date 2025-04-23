package pool

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/dispatch/types"
)

// StreamChannelV2 represents a channel dedicated to a specific stream
type StreamChannelV2 struct {
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
}

// GetBufferUtilization returns the current buffer utilization as a percentage (0-100)
func (s *StreamChannelV2) GetBufferUtilization() int {
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
func (s *StreamChannelV2) IsHealthy() bool {
	return atomic.LoadInt32(&s.healthy) == 1
}

// SetHealthy sets the stream's health status
func (s *StreamChannelV2) SetHealthy(healthy bool) {
	var value int32 = 0
	if healthy {
		value = 1
	}
	atomic.StoreInt32(&s.healthy, value)
}

// GetLastActivity returns the timestamp of the last activity
func (s *StreamChannelV2) GetLastActivity() time.Time {
	return time.Unix(0, atomic.LoadInt64(&s.lastActivity))
}

// UpdateLastActivity updates the last activity timestamp
func (s *StreamChannelV2) UpdateLastActivity() {
	atomic.StoreInt64(&s.lastActivity, time.Now().UnixNano())
}

// Close closes the stream channel
func (s *StreamChannelV2) Close() {
	s.cancel()
	close(s.PacketChan)
	s.Stream.Close()
}

// ProcessPackets starts processing packets from the channel
func (s *StreamChannelV2) ProcessPackets(peerID string) {
	logger := streamLog.WithFields(logrus.Fields{
		"peer_id": peerID,
	})

	defer func() {
		// Mark the stream as unhealthy
		s.SetHealthy(false)
		logger.Debug("Stream processor stopped")
	}()

	for {
		select {
		case <-s.ctx.Done():
			logger.Debug("Stream context cancelled")
			return
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
					}
				}

				// Mark the stream as unhealthy
				s.SetHealthy(false)

				// Increment error count
				atomic.AddInt64(&s.Metrics.ErrorCount, 1)

				return
			}

			// Signal success on the done channel if provided
			if packet.DoneCh != nil {
				select {
				case packet.DoneCh <- nil:
					// Successfully sent the result
				default:
					// Channel might be closed or full, don't panic
				}
			}

			// Update metrics
			atomic.AddInt64(&s.Metrics.PacketCount, 1)
			atomic.AddInt64(&s.Metrics.BytesSent, int64(len(packet.Data)))
		}
	}
}

// writePacketToStream writes a packet to the stream
func (s *StreamChannelV2) writePacketToStream(packet *types.QueuedPacket) error {
	// Write the packet to the stream
	_, err := s.Stream.Write(packet.Data)
	if err != nil {
		return types.NewNetworkError(err, "write", packet.DestIP, "")
	}

	return nil
}
