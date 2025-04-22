package router

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// RouterMetrics tracks performance metrics for scaling decisions
type RouterMetrics struct {
	// Per-peer metrics
	streamThroughput    map[peer.ID]map[int]float64 // Packets/sec per stream
	peerThroughput      map[peer.ID]float64         // Total packets/sec per peer
	throughputHistory   map[peer.ID][]float64       // Recent throughput samples
	lastPacketCounts    map[peer.ID]map[int]int64   // Last packet counts for throughput calculation
	currentPacketCounts map[peer.ID]map[int]int64   // Current packet counts
	lastUpdateTime      time.Time                   // Last metrics update time
	mu                  sync.RWMutex
}

// NewRouterMetrics creates a new router metrics instance
func NewRouterMetrics() *RouterMetrics {
	return &RouterMetrics{
		streamThroughput:    make(map[peer.ID]map[int]float64),
		peerThroughput:      make(map[peer.ID]float64),
		throughputHistory:   make(map[peer.ID][]float64),
		lastPacketCounts:    make(map[peer.ID]map[int]int64),
		currentPacketCounts: make(map[peer.ID]map[int]int64),
		lastUpdateTime:      time.Now(),
	}
}

// recordPacket records a packet for metrics
func (m *RouterMetrics) recordPacket(peerID peer.ID, streamIndex int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Initialize maps if needed
	if m.currentPacketCounts[peerID] == nil {
		m.currentPacketCounts[peerID] = make(map[int]int64)
	}

	// Increment packet count
	m.currentPacketCounts[peerID][streamIndex]++
}

// updateThroughput calculates throughput based on packet counts
func (m *RouterMetrics) updateThroughput() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(m.lastUpdateTime).Seconds()

	if elapsed < 0.1 {
		// Too soon to update
		return
	}

	// Initialize maps if needed
	if m.streamThroughput == nil {
		m.streamThroughput = make(map[peer.ID]map[int]float64)
	}
	if m.peerThroughput == nil {
		m.peerThroughput = make(map[peer.ID]float64)
	}

	// Calculate throughput for each peer and stream
	for peerID, streamCounts := range m.currentPacketCounts {
		// Initialize peer maps if needed
		if m.streamThroughput[peerID] == nil {
			m.streamThroughput[peerID] = make(map[int]float64)
		}
		if m.lastPacketCounts[peerID] == nil {
			m.lastPacketCounts[peerID] = make(map[int]int64)
		}

		// Calculate total throughput for this peer
		var totalPackets int64

		for streamIndex, count := range streamCounts {
			// Get previous count
			lastCount := m.lastPacketCounts[peerID][streamIndex]

			// Calculate packets per second for this stream
			packetsDelta := count - lastCount
			throughput := float64(packetsDelta) / elapsed

			// Store throughput
			m.streamThroughput[peerID][streamIndex] = throughput

			// Add to total
			totalPackets += packetsDelta

			// Update last count
			m.lastPacketCounts[peerID][streamIndex] = count
		}

		// Calculate total throughput for peer
		peerThroughput := float64(totalPackets) / elapsed
		m.peerThroughput[peerID] = peerThroughput

		// Store in history (keep last 5 samples)
		if m.throughputHistory[peerID] == nil {
			m.throughputHistory[peerID] = make([]float64, 0, 5)
		}
		m.throughputHistory[peerID] = append(m.throughputHistory[peerID], peerThroughput)
		if len(m.throughputHistory[peerID]) > 5 {
			m.throughputHistory[peerID] = m.throughputHistory[peerID][1:]
		}
	}

	// Update last update time
	m.lastUpdateTime = now
}
