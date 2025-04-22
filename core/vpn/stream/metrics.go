package stream

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
)

// StreamMetrics tracks metrics for the stream service
type StreamMetrics struct {
	// Mutex to protect access to metrics
	mu sync.RWMutex

	// Stream metrics by peer and index
	streamReuse              map[string]map[int]int64
	streamRegistration       map[string]map[int]int64
	streamAcquisitionFailure map[string]map[int]int64
	streamRelease            map[string]map[int]int64
	streamThroughput         map[string]map[int]int64

	// Global metrics
	totalStreamReuse              int64
	totalStreamRegistration       int64
	totalStreamAcquisitionFailure int64
	totalStreamRelease            int64
	totalStreamThroughput         int64
}

// NewStreamMetrics creates a new stream metrics
func NewStreamMetrics() *StreamMetrics {
	return &StreamMetrics{
		streamReuse:              make(map[string]map[int]int64),
		streamRegistration:       make(map[string]map[int]int64),
		streamAcquisitionFailure: make(map[string]map[int]int64),
		streamRelease:            make(map[string]map[int]int64),
		streamThroughput:         make(map[string]map[int]int64),
	}
}

// IncrementStreamReuse increments the stream reuse counter
func (m *StreamMetrics) IncrementStreamReuse(peerID peer.ID, index int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	peerIDStr := peerID.String()
	if _, exists := m.streamReuse[peerIDStr]; !exists {
		m.streamReuse[peerIDStr] = make(map[int]int64)
	}
	m.streamReuse[peerIDStr][index]++
	m.totalStreamReuse++
}

// IncrementStreamRegistration increments the stream registration counter
func (m *StreamMetrics) IncrementStreamRegistration(peerID peer.ID, index int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	peerIDStr := peerID.String()
	if _, exists := m.streamRegistration[peerIDStr]; !exists {
		m.streamRegistration[peerIDStr] = make(map[int]int64)
	}
	m.streamRegistration[peerIDStr][index]++
	m.totalStreamRegistration++
}

// IncrementStreamAcquisitionFailure increments the stream acquisition failure counter
func (m *StreamMetrics) IncrementStreamAcquisitionFailure(peerID peer.ID, index int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	peerIDStr := peerID.String()
	if _, exists := m.streamAcquisitionFailure[peerIDStr]; !exists {
		m.streamAcquisitionFailure[peerIDStr] = make(map[int]int64)
	}
	m.streamAcquisitionFailure[peerIDStr][index]++
	m.totalStreamAcquisitionFailure++
}

// IncrementStreamRelease increments the stream release counter
func (m *StreamMetrics) IncrementStreamRelease(peerID peer.ID, index int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	peerIDStr := peerID.String()
	if _, exists := m.streamRelease[peerIDStr]; !exists {
		m.streamRelease[peerIDStr] = make(map[int]int64)
	}
	m.streamRelease[peerIDStr][index]++
	m.totalStreamRelease++
}

// IncrementStreamThroughput increments the stream throughput counter
func (m *StreamMetrics) IncrementStreamThroughput(peerID peer.ID, index int, bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	peerIDStr := peerID.String()
	if _, exists := m.streamThroughput[peerIDStr]; !exists {
		m.streamThroughput[peerIDStr] = make(map[int]int64)
	}
	m.streamThroughput[peerIDStr][index] += bytes
	m.totalStreamThroughput += bytes
}

// GetStreamReuseMetrics returns the stream reuse metrics
func (m *StreamMetrics) GetStreamReuseMetrics() map[string]map[int]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]map[int]int64)
	for peerID, metrics := range m.streamReuse {
		result[peerID] = make(map[int]int64)
		for index, count := range metrics {
			result[peerID][index] = count
		}
	}
	return result
}

// GetStreamRegistrationMetrics returns the stream registration metrics
func (m *StreamMetrics) GetStreamRegistrationMetrics() map[string]map[int]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]map[int]int64)
	for peerID, metrics := range m.streamRegistration {
		result[peerID] = make(map[int]int64)
		for index, count := range metrics {
			result[peerID][index] = count
		}
	}
	return result
}

// GetStreamAcquisitionFailureMetrics returns the stream acquisition failure metrics
func (m *StreamMetrics) GetStreamAcquisitionFailureMetrics() map[string]map[int]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]map[int]int64)
	for peerID, metrics := range m.streamAcquisitionFailure {
		result[peerID] = make(map[int]int64)
		for index, count := range metrics {
			result[peerID][index] = count
		}
	}
	return result
}

// GetStreamReleaseMetrics returns the stream release metrics
func (m *StreamMetrics) GetStreamReleaseMetrics() map[string]map[int]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]map[int]int64)
	for peerID, metrics := range m.streamRelease {
		result[peerID] = make(map[int]int64)
		for index, count := range metrics {
			result[peerID][index] = count
		}
	}
	return result
}

// GetStreamThroughputMetrics returns the stream throughput metrics
func (m *StreamMetrics) GetStreamThroughputMetrics() map[string]map[int]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]map[int]int64)
	for peerID, metrics := range m.streamThroughput {
		result[peerID] = make(map[int]int64)
		for index, count := range metrics {
			result[peerID][index] = count
		}
	}
	return result
}

// GetTotalStreamReuseMetrics returns the total stream reuse metrics
func (m *StreamMetrics) GetTotalStreamReuseMetrics() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalStreamReuse
}

// GetTotalStreamRegistrationMetrics returns the total stream registration metrics
func (m *StreamMetrics) GetTotalStreamRegistrationMetrics() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalStreamRegistration
}

// GetTotalStreamAcquisitionFailureMetrics returns the total stream acquisition failure metrics
func (m *StreamMetrics) GetTotalStreamAcquisitionFailureMetrics() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalStreamAcquisitionFailure
}

// GetTotalStreamReleaseMetrics returns the total stream release metrics
func (m *StreamMetrics) GetTotalStreamReleaseMetrics() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalStreamRelease
}

// GetTotalStreamThroughputMetrics returns the total stream throughput metrics
func (m *StreamMetrics) GetTotalStreamThroughputMetrics() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalStreamThroughput
}

// GetAllMetrics returns all metrics
func (m *StreamMetrics) GetAllMetrics() map[string]interface{} {
	return map[string]interface{}{
		"stream_reuse":                     m.GetStreamReuseMetrics(),
		"stream_registration":              m.GetStreamRegistrationMetrics(),
		"stream_acquisition_failure":       m.GetStreamAcquisitionFailureMetrics(),
		"stream_release":                   m.GetStreamReleaseMetrics(),
		"stream_throughput":                m.GetStreamThroughputMetrics(),
		"total_stream_reuse":               m.GetTotalStreamReuseMetrics(),
		"total_stream_registration":        m.GetTotalStreamRegistrationMetrics(),
		"total_stream_acquisition_failure": m.GetTotalStreamAcquisitionFailureMetrics(),
		"total_stream_release":             m.GetTotalStreamReleaseMetrics(),
		"total_stream_throughput":          m.GetTotalStreamThroughputMetrics(),
	}
}
