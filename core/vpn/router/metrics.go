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
	// Use a read lock to check if maps are initialized
	m.mu.RLock()
	_, peerExists := m.currentPacketCounts[peerID]
	m.mu.RUnlock()

	// Initialize maps if needed (with write lock)
	if !peerExists {
		m.mu.Lock()
		// Check again in case another goroutine initialized it
		if m.currentPacketCounts[peerID] == nil {
			m.currentPacketCounts[peerID] = make(map[int]int64)
		}
		m.mu.Unlock()
	}

	// Use a shorter lock scope for incrementing the counter
	m.mu.Lock()
	m.currentPacketCounts[peerID][streamIndex]++
	m.mu.Unlock()
}

// updateThroughput calculates throughput based on packet counts
func (m *RouterMetrics) updateThroughput() {
	// Take a snapshot of the current packet counts with minimal lock time
	now := time.Now()

	// First, check if it's too soon to update
	m.mu.RLock()
	elapsed := now.Sub(m.lastUpdateTime).Seconds()
	if elapsed < 0.1 {
		// Too soon to update
		m.mu.RUnlock()
		return
	}
	m.mu.RUnlock()

	// Create a snapshot of the current packet counts
	currentSnapshot := make(map[peer.ID]map[int]int64)
	lastSnapshot := make(map[peer.ID]map[int]int64)

	m.mu.Lock()
	// Initialize maps if needed
	if m.streamThroughput == nil {
		m.streamThroughput = make(map[peer.ID]map[int]float64)
	}
	if m.peerThroughput == nil {
		m.peerThroughput = make(map[peer.ID]float64)
	}

	// Copy the current and last packet counts
	for peerID, streamCounts := range m.currentPacketCounts {
		currentSnapshot[peerID] = make(map[int]int64)
		for streamIndex, count := range streamCounts {
			currentSnapshot[peerID][streamIndex] = count
		}

		// Initialize peer maps if needed
		if m.lastPacketCounts[peerID] == nil {
			m.lastPacketCounts[peerID] = make(map[int]int64)
		}
		if m.streamThroughput[peerID] == nil {
			m.streamThroughput[peerID] = make(map[int]float64)
		}

		lastSnapshot[peerID] = make(map[int]int64)
		for streamIndex, count := range m.lastPacketCounts[peerID] {
			lastSnapshot[peerID][streamIndex] = count
		}
	}

	// Store the current counts as the last counts for next time
	for peerID, streamCounts := range m.currentPacketCounts {
		if m.lastPacketCounts[peerID] == nil {
			m.lastPacketCounts[peerID] = make(map[int]int64)
		}
		for streamIndex, count := range streamCounts {
			m.lastPacketCounts[peerID][streamIndex] = count
		}
	}

	// Update last update time
	lastUpdateTime := m.lastUpdateTime
	m.lastUpdateTime = now
	m.mu.Unlock()

	// Calculate throughput for each peer and stream outside the lock
	newThroughputs := make(map[peer.ID]map[int]float64)
	newPeerThroughputs := make(map[peer.ID]float64)
	newHistory := make(map[peer.ID][]float64)

	// Use the actual elapsed time for calculations
	elapsed = now.Sub(lastUpdateTime).Seconds()

	for peerID, streamCounts := range currentSnapshot {
		newThroughputs[peerID] = make(map[int]float64)

		// Calculate total throughput for this peer
		var totalPackets int64

		for streamIndex, count := range streamCounts {
			// Get previous count
			lastCount := int64(0)
			if lastSnapshot[peerID] != nil {
				lastCount = lastSnapshot[peerID][streamIndex]
			}

			// Calculate packets per second for this stream
			packetsDelta := count - lastCount
			throughput := float64(packetsDelta) / elapsed

			// Store throughput
			newThroughputs[peerID][streamIndex] = throughput

			// Add to total
			totalPackets += packetsDelta
		}

		// Calculate total throughput for peer
		peerThroughput := float64(totalPackets) / elapsed
		newPeerThroughputs[peerID] = peerThroughput

		// Get existing history
		m.mu.RLock()
		history := m.throughputHistory[peerID]
		m.mu.RUnlock()

		// Create a copy of the history
		newHistory[peerID] = make([]float64, len(history), 5)
		copy(newHistory[peerID], history)

		// Add new sample
		newHistory[peerID] = append(newHistory[peerID], peerThroughput)
		if len(newHistory[peerID]) > 5 {
			newHistory[peerID] = newHistory[peerID][1:]
		}
	}

	// Update the metrics with the new values
	m.mu.Lock()
	for peerID, throughputs := range newThroughputs {
		for streamIndex, throughput := range throughputs {
			m.streamThroughput[peerID][streamIndex] = throughput
		}
		m.peerThroughput[peerID] = newPeerThroughputs[peerID]
		m.throughputHistory[peerID] = newHistory[peerID]
	}
	m.mu.Unlock()
}
