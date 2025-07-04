package metrics

import (
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// MetricsService implements the types.Metrics interface with thread-safe counters and gauges
type MetricsService struct {
	// Bid metrics
	bidsSubmitted   int64
	bidsAccepted    int64
	bidsRejected    int64
	bidLatencyTotal float64
	bidLatencyCount int64

	// Order metrics
	ordersTracked      int64
	ordersCompleted    int64
	orderDurationTotal float64
	orderDurationCount int64

	// Resource metrics
	resourceUtilization map[string]*ResourceUsageStats
	resourceAllocation  map[string]*ResourceUsageStats

	// Financial metrics
	revenueTotal big.Int
	profitTotal  big.Int
	costTotal    big.Int

	// Internal state
	mu sync.RWMutex
}

// ResourceUsageStats tracks statistics for resource usage
type ResourceUsageStats struct {
	CPU     *ResourceStats
	GPU     *ResourceStats
	Memory  *ResourceStats
	Disk    *ResourceStats
	Network *ResourceStats
}

// ResourceStats tracks individual resource statistics
type ResourceStats struct {
	Total   *big.Int
	Count   int64
	Average *big.Int
	Min     *big.Int
	Max     *big.Int
}

// NewMetrics creates a new MetricsService instance
func NewMetrics() *MetricsService {
	return &MetricsService{
		resourceUtilization: make(map[string]*ResourceUsageStats),
		resourceAllocation:  make(map[string]*ResourceUsageStats),
	}
}

// Bid metrics

// IncrementBidsSubmitted increments the count of submitted bids
func (m *MetricsService) IncrementBidsSubmitted() {
	atomic.AddInt64(&m.bidsSubmitted, 1)
}

// IncrementBidsAccepted increments the count of accepted bids
func (m *MetricsService) IncrementBidsAccepted() {
	atomic.AddInt64(&m.bidsAccepted, 1)
}

// IncrementBidsRejected increments the count of rejected bids
func (m *MetricsService) IncrementBidsRejected() {
	atomic.AddInt64(&m.bidsRejected, 1)
}

// RecordBidLatency records the latency of a bid operation
func (m *MetricsService) RecordBidLatency(duration float64) {
	atomic.AddInt64(&m.bidLatencyCount, 1)

	// Use atomic operations for float64 (simplified approach)
	m.mu.Lock()
	m.bidLatencyTotal += duration
	m.mu.Unlock()
}

// Order metrics

// IncrementOrdersTracked increments the count of tracked orders
func (m *MetricsService) IncrementOrdersTracked() {
	atomic.AddInt64(&m.ordersTracked, 1)
}

// IncrementOrdersCompleted increments the count of completed orders
func (m *MetricsService) IncrementOrdersCompleted() {
	atomic.AddInt64(&m.ordersCompleted, 1)
}

// RecordOrderDuration records the duration of an order
func (m *MetricsService) RecordOrderDuration(duration float64) {
	atomic.AddInt64(&m.orderDurationCount, 1)

	m.mu.Lock()
	m.orderDurationTotal += duration
	m.mu.Unlock()
}

// Resource metrics

// RecordResourceUtilization records resource utilization metrics
func (m *MetricsService) RecordResourceUtilization(usage *types.ResourceUsage) {
	if usage == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := "utilization"
	stats := m.getOrCreateResourceStats(m.resourceUtilization, key)
	m.updateResourceStats(stats, usage)
}

// RecordResourceAllocation records resource allocation metrics
func (m *MetricsService) RecordResourceAllocation(usage *types.ResourceUsage) {
	if usage == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := "allocation"
	stats := m.getOrCreateResourceStats(m.resourceAllocation, key)
	m.updateResourceStats(stats, usage)
}

// Financial metrics

// RecordRevenue records revenue amount
func (m *MetricsService) RecordRevenue(amount *big.Int) {
	if amount == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.revenueTotal.Add(&m.revenueTotal, amount)
}

// RecordProfit records profit amount
func (m *MetricsService) RecordProfit(amount *big.Int) {
	if amount == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.profitTotal.Add(&m.profitTotal, amount)
}

// RecordCost records cost amount
func (m *MetricsService) RecordCost(amount *big.Int) {
	if amount == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.costTotal.Add(&m.costTotal, amount)
}

// Helper methods

// getOrCreateResourceStats gets or creates resource usage statistics
func (m *MetricsService) getOrCreateResourceStats(statsMap map[string]*ResourceUsageStats, key string) *ResourceUsageStats {
	if stats, exists := statsMap[key]; exists {
		return stats
	}

	stats := &ResourceUsageStats{
		CPU:     m.newResourceStats(),
		GPU:     m.newResourceStats(),
		Memory:  m.newResourceStats(),
		Disk:    m.newResourceStats(),
		Network: m.newResourceStats(),
	}

	statsMap[key] = stats
	return stats
}

// newResourceStats creates a new ResourceStats instance
func (m *MetricsService) newResourceStats() *ResourceStats {
	return &ResourceStats{
		Total:   big.NewInt(0),
		Average: big.NewInt(0),
		Min:     nil,
		Max:     nil,
	}
}

// updateResourceStats updates resource statistics with new usage data
func (m *MetricsService) updateResourceStats(stats *ResourceUsageStats, usage *types.ResourceUsage) {
	if usage.CPUUsed != nil {
		m.updateIndividualResourceStats(stats.CPU, usage.CPUUsed)
	}
	if usage.GPUUsed != nil {
		m.updateIndividualResourceStats(stats.GPU, usage.GPUUsed)
	}
	if usage.MemoryUsed != nil {
		m.updateIndividualResourceStats(stats.Memory, usage.MemoryUsed)
	}
	if usage.DiskUsed != nil {
		m.updateIndividualResourceStats(stats.Disk, usage.DiskUsed)
	}

}

// updateIndividualResourceStats updates individual resource statistics
func (m *MetricsService) updateIndividualResourceStats(resourceStats *ResourceStats, value *big.Int) {
	if value == nil {
		return
	}

	// Update total
	resourceStats.Total.Add(resourceStats.Total, value)
	resourceStats.Count++

	// Update min/max
	if resourceStats.Min == nil || value.Cmp(resourceStats.Min) < 0 {
		resourceStats.Min = new(big.Int).Set(value)
	}
	if resourceStats.Max == nil || value.Cmp(resourceStats.Max) > 0 {
		resourceStats.Max = new(big.Int).Set(value)
	}

	// Update average
	if resourceStats.Count > 0 {
		resourceStats.Average.Div(resourceStats.Total, big.NewInt(resourceStats.Count))
	}
}

// GetStats returns a snapshot of all metrics
func (m *MetricsService) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bidLatencyAvg := 0.0
	if m.bidLatencyCount > 0 {
		bidLatencyAvg = m.bidLatencyTotal / float64(m.bidLatencyCount)
	}

	orderDurationAvg := 0.0
	if m.orderDurationCount > 0 {
		orderDurationAvg = m.orderDurationTotal / float64(m.orderDurationCount)
	}

	return map[string]interface{}{
		"bids": map[string]interface{}{
			"submitted": m.bidsSubmitted,
			"accepted":  m.bidsAccepted,
			"rejected":  m.bidsRejected,
			"latency": map[string]interface{}{
				"total":   m.bidLatencyTotal,
				"count":   m.bidLatencyCount,
				"average": bidLatencyAvg,
			},
		},
		"orders": map[string]interface{}{
			"tracked":   m.ordersTracked,
			"completed": m.ordersCompleted,
			"duration": map[string]interface{}{
				"total":   m.orderDurationTotal,
				"count":   m.orderDurationCount,
				"average": orderDurationAvg,
			},
		},
		"resources": map[string]interface{}{
			"utilization": m.serializeResourceStats(m.resourceUtilization),
			"allocation":  m.serializeResourceStats(m.resourceAllocation),
		},
		"financial": map[string]interface{}{
			"revenue": m.revenueTotal.String(),
			"profit":  m.profitTotal.String(),
			"cost":    m.costTotal.String(),
		},
	}
}

// serializeResourceStats serializes resource statistics for JSON output
func (m *MetricsService) serializeResourceStats(statsMap map[string]*ResourceUsageStats) map[string]interface{} {
	result := make(map[string]interface{})

	for key, stats := range statsMap {
		result[key] = map[string]interface{}{
			"cpu":     m.serializeIndividualResourceStats(stats.CPU),
			"gpu":     m.serializeIndividualResourceStats(stats.GPU),
			"memory":  m.serializeIndividualResourceStats(stats.Memory),
			"disk":    m.serializeIndividualResourceStats(stats.Disk),
			"network": m.serializeIndividualResourceStats(stats.Network),
		}
	}

	return result
}

// serializeIndividualResourceStats serializes individual resource statistics
func (m *MetricsService) serializeIndividualResourceStats(stats *ResourceStats) map[string]interface{} {
	result := map[string]interface{}{
		"total":   stats.Total.String(),
		"count":   stats.Count,
		"average": stats.Average.String(),
	}

	if stats.Min != nil {
		result["min"] = stats.Min.String()
	}
	if stats.Max != nil {
		result["max"] = stats.Max.String()
	}

	return result
}

// Reset resets all metrics to zero
func (m *MetricsService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreInt64(&m.bidsSubmitted, 0)
	atomic.StoreInt64(&m.bidsAccepted, 0)
	atomic.StoreInt64(&m.bidsRejected, 0)
	atomic.StoreInt64(&m.ordersTracked, 0)
	atomic.StoreInt64(&m.ordersCompleted, 0)
	atomic.StoreInt64(&m.bidLatencyCount, 0)
	atomic.StoreInt64(&m.orderDurationCount, 0)

	m.bidLatencyTotal = 0
	m.orderDurationTotal = 0
	m.revenueTotal.SetInt64(0)
	m.profitTotal.SetInt64(0)
	m.costTotal.SetInt64(0)

	// Reset resource stats
	m.resourceUtilization = make(map[string]*ResourceUsageStats)
	m.resourceAllocation = make(map[string]*ResourceUsageStats)
}
