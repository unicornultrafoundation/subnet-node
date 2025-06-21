package bidengine

import (
	"math/big"
	"sync"
	"time"
)

// MetricsService implements Metrics interface
type MetricsService struct {
	mu sync.RWMutex

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
	resourceUtilization map[string]*ResourceUsage
	resourceAllocation  map[string]*ResourceUsage

	// Financial metrics
	revenueTotal *big.Int
	profitTotal  *big.Int
	costTotal    *big.Int
}

// NewMetrics creates a new MetricsService instance
func NewMetrics() *MetricsService {
	return &MetricsService{
		resourceUtilization: make(map[string]*ResourceUsage),
		resourceAllocation:  make(map[string]*ResourceUsage),
		revenueTotal:        big.NewInt(0),
		profitTotal:         big.NewInt(0),
		costTotal:           big.NewInt(0),
	}
}

// IncrementBidsSubmitted increments the submitted bids counter
func (m *MetricsService) IncrementBidsSubmitted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bidsSubmitted++
}

// IncrementBidsAccepted increments the accepted bids counter
func (m *MetricsService) IncrementBidsAccepted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bidsAccepted++
}

// IncrementBidsRejected increments the rejected bids counter
func (m *MetricsService) IncrementBidsRejected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bidsRejected++
}

// RecordBidLatency records bid latency
func (m *MetricsService) RecordBidLatency(duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bidLatencyCount++
	m.bidLatencyTotal += duration
}

// IncrementOrdersTracked increments the tracked orders counter
func (m *MetricsService) IncrementOrdersTracked() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ordersTracked++
}

// IncrementOrdersCompleted increments the completed orders counter
func (m *MetricsService) IncrementOrdersCompleted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ordersCompleted++
}

// RecordOrderDuration records order duration
func (m *MetricsService) RecordOrderDuration(duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orderDurationCount++
	m.orderDurationTotal += duration
}

// RecordResourceUtilization records resource utilization
func (m *MetricsService) RecordResourceUtilization(usage *ResourceUsage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store utilization by timestamp (simplified)
	timestamp := time.Now().Format("2006-01-02T15:04:05")
	m.resourceUtilization[timestamp] = usage
}

// RecordResourceAllocation records resource allocation
func (m *MetricsService) RecordResourceAllocation(usage *ResourceUsage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store allocation by timestamp (simplified)
	timestamp := time.Now().Format("2006-01-02T15:04:05")
	m.resourceAllocation[timestamp] = usage
}

// RecordRevenue records revenue
func (m *MetricsService) RecordRevenue(amount *big.Int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.revenueTotal.Add(m.revenueTotal, amount)
}

// RecordProfit records profit
func (m *MetricsService) RecordProfit(amount *big.Int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profitTotal.Add(m.profitTotal, amount)
}

// RecordCost records cost
func (m *MetricsService) RecordCost(amount *big.Int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.costTotal.Add(m.costTotal, amount)
}

// GetStats returns current metrics statistics
func (m *MetricsService) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgBidLatency := 0.0
	if m.bidLatencyCount > 0 {
		avgBidLatency = m.bidLatencyTotal / float64(m.bidLatencyCount)
	}

	avgOrderDuration := 0.0
	if m.orderDurationCount > 0 {
		avgOrderDuration = m.orderDurationTotal / float64(m.orderDurationCount)
	}

	return map[string]interface{}{
		"bids": map[string]interface{}{
			"submitted":  m.bidsSubmitted,
			"accepted":   m.bidsAccepted,
			"rejected":   m.bidsRejected,
			"avgLatency": avgBidLatency,
		},
		"orders": map[string]interface{}{
			"tracked":     m.ordersTracked,
			"completed":   m.ordersCompleted,
			"avgDuration": avgOrderDuration,
		},
		"resources": map[string]interface{}{
			"utilizationCount": len(m.resourceUtilization),
			"allocationCount":  len(m.resourceAllocation),
		},
		"financial": map[string]interface{}{
			"revenue": m.revenueTotal.String(),
			"profit":  m.profitTotal.String(),
			"cost":    m.costTotal.String(),
		},
		"lastUpdated": time.Now(),
	}
}

// Reset resets all metrics
func (m *MetricsService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bidsSubmitted = 0
	m.bidsAccepted = 0
	m.bidsRejected = 0
	m.ordersTracked = 0
	m.ordersCompleted = 0

	m.bidLatencyTotal = 0
	m.bidLatencyCount = 0
	m.orderDurationTotal = 0
	m.orderDurationCount = 0

	m.resourceUtilization = make(map[string]*ResourceUsage)
	m.resourceAllocation = make(map[string]*ResourceUsage)

	m.revenueTotal = big.NewInt(0)
	m.profitTotal = big.NewInt(0)
	m.costTotal = big.NewInt(0)
}
