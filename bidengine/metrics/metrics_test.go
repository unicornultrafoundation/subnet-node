package metrics

import (
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

func TestNewMetrics(t *testing.T) {
	metrics := NewMetrics()
	if metrics == nil {
		t.Fatal("NewMetrics() returned nil")
	}

	// Check initial values
	stats := metrics.GetStats()
	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}

	// Check bid metrics
	bids := stats["bids"].(map[string]interface{})
	if bids["submitted"] != int64(0) {
		t.Errorf("Expected bids submitted to be 0, got %v", bids["submitted"])
	}
}

func TestBidMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Test bid submission
	metrics.IncrementBidsSubmitted()
	metrics.IncrementBidsSubmitted()

	// Test bid acceptance
	metrics.IncrementBidsAccepted()

	// Test bid rejection
	metrics.IncrementBidsRejected()

	// Test bid latency
	metrics.RecordBidLatency(1.5)
	metrics.RecordBidLatency(2.5)

	stats := metrics.GetStats()
	bids := stats["bids"].(map[string]interface{})

	if bids["submitted"] != int64(2) {
		t.Errorf("Expected 2 bids submitted, got %v", bids["submitted"])
	}

	if bids["accepted"] != int64(1) {
		t.Errorf("Expected 1 bid accepted, got %v", bids["accepted"])
	}

	if bids["rejected"] != int64(1) {
		t.Errorf("Expected 1 bid rejected, got %v", bids["rejected"])
	}

	latency := bids["latency"].(map[string]interface{})
	if latency["count"] != int64(2) {
		t.Errorf("Expected 2 latency records, got %v", latency["count"])
	}

	if latency["total"] != 4.0 {
		t.Errorf("Expected total latency 4.0, got %v", latency["total"])
	}

	if latency["average"] != 2.0 {
		t.Errorf("Expected average latency 2.0, got %v", latency["average"])
	}
}

func TestOrderMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Test order tracking
	metrics.IncrementOrdersTracked()
	metrics.IncrementOrdersTracked()
	metrics.IncrementOrdersTracked()

	// Test order completion
	metrics.IncrementOrdersCompleted()

	// Test order duration
	metrics.RecordOrderDuration(10.5)
	metrics.RecordOrderDuration(15.5)

	stats := metrics.GetStats()
	orders := stats["orders"].(map[string]interface{})

	if orders["tracked"] != int64(3) {
		t.Errorf("Expected 3 orders tracked, got %v", orders["tracked"])
	}

	if orders["completed"] != int64(1) {
		t.Errorf("Expected 1 order completed, got %v", orders["completed"])
	}

	duration := orders["duration"].(map[string]interface{})
	if duration["count"] != int64(2) {
		t.Errorf("Expected 2 duration records, got %v", duration["count"])
	}

	if duration["total"] != 26.0 {
		t.Errorf("Expected total duration 26.0, got %v", duration["total"])
	}

	if duration["average"] != 13.0 {
		t.Errorf("Expected average duration 13.0, got %v", duration["average"])
	}
}

func TestResourceMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Test resource utilization
	usage1 := &types.ResourceUsage{
		CPUUsed:     big.NewInt(4),
		GPUUsed:     big.NewInt(2),
		MemoryUsed:  big.NewInt(8192),
		DiskUsed:    big.NewInt(100),
		NetworkUsed: big.NewInt(1000),
	}

	usage2 := &types.ResourceUsage{
		CPUUsed:     big.NewInt(8),
		GPUUsed:     big.NewInt(4),
		MemoryUsed:  big.NewInt(16384),
		DiskUsed:    big.NewInt(200),
		NetworkUsed: big.NewInt(2000),
	}

	metrics.RecordResourceUtilization(usage1)
	metrics.RecordResourceUtilization(usage2)

	// Test resource allocation
	metrics.RecordResourceAllocation(usage1)

	stats := metrics.GetStats()
	resources := stats["resources"].(map[string]interface{})

	utilization := resources["utilization"].(map[string]interface{})
	allocation := resources["allocation"].(map[string]interface{})

	// Check utilization stats
	utilCPU := utilization["utilization"].(map[string]interface{})["cpu"].(map[string]interface{})
	if utilCPU["total"] != "12" {
		t.Errorf("Expected CPU utilization total 12, got %v", utilCPU["total"])
	}

	if utilCPU["count"] != int64(2) {
		t.Errorf("Expected CPU utilization count 2, got %v", utilCPU["count"])
	}

	if utilCPU["average"] != "6" {
		t.Errorf("Expected CPU utilization average 6, got %v", utilCPU["average"])
	}

	// Check allocation stats
	allocCPU := allocation["allocation"].(map[string]interface{})["cpu"].(map[string]interface{})
	if allocCPU["total"] != "4" {
		t.Errorf("Expected CPU allocation total 4, got %v", allocCPU["total"])
	}
}

func TestFinancialMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Test revenue recording
	revenue1 := big.NewInt(1000)
	revenue2 := big.NewInt(2000)
	metrics.RecordRevenue(revenue1)
	metrics.RecordRevenue(revenue2)

	// Test profit recording
	profit := big.NewInt(500)
	metrics.RecordProfit(profit)

	// Test cost recording
	cost := big.NewInt(300)
	metrics.RecordCost(cost)

	stats := metrics.GetStats()
	financial := stats["financial"].(map[string]interface{})

	if financial["revenue"] != "3000" {
		t.Errorf("Expected revenue 3000, got %v", financial["revenue"])
	}

	if financial["profit"] != "500" {
		t.Errorf("Expected profit 500, got %v", financial["profit"])
	}

	if financial["cost"] != "300" {
		t.Errorf("Expected cost 300, got %v", financial["cost"])
	}
}

func TestNilHandling(t *testing.T) {
	metrics := NewMetrics()

	// Test with nil resource usage
	metrics.RecordResourceUtilization(nil)
	metrics.RecordResourceAllocation(nil)

	// Test with nil financial amounts
	metrics.RecordRevenue(nil)
	metrics.RecordProfit(nil)
	metrics.RecordCost(nil)

	// Should not panic and should not affect metrics
	stats := metrics.GetStats()
	if stats == nil {
		t.Fatal("GetStats() returned nil after nil operations")
	}
}

func TestReset(t *testing.T) {
	metrics := NewMetrics()

	// Add some metrics
	metrics.IncrementBidsSubmitted()
	metrics.IncrementBidsAccepted()
	metrics.RecordBidLatency(1.0)
	metrics.IncrementOrdersTracked()
	metrics.RecordOrderDuration(10.0)
	metrics.RecordRevenue(big.NewInt(1000))

	// Reset
	metrics.Reset()

	// Check that all metrics are zero
	stats := metrics.GetStats()
	bids := stats["bids"].(map[string]interface{})
	orders := stats["orders"].(map[string]interface{})
	financial := stats["financial"].(map[string]interface{})

	if bids["submitted"] != int64(0) {
		t.Errorf("Expected bids submitted to be 0 after reset, got %v", bids["submitted"])
	}

	if orders["tracked"] != int64(0) {
		t.Errorf("Expected orders tracked to be 0 after reset, got %v", orders["tracked"])
	}

	if financial["revenue"] != "0" {
		t.Errorf("Expected revenue to be 0 after reset, got %v", financial["revenue"])
	}
}

func TestConcurrentAccess(t *testing.T) {
	metrics := NewMetrics()
	done := make(chan bool, 10)

	// Start multiple goroutines to test concurrent access
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				metrics.IncrementBidsSubmitted()
				metrics.IncrementBidsAccepted()
				metrics.RecordBidLatency(float64(j))
				metrics.IncrementOrdersTracked()
				metrics.RecordOrderDuration(float64(j))
				metrics.RecordRevenue(big.NewInt(int64(j)))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check final values
	stats := metrics.GetStats()
	bids := stats["bids"].(map[string]interface{})

	expectedSubmitted := int64(1000) // 10 goroutines * 100 iterations
	if bids["submitted"] != expectedSubmitted {
		t.Errorf("Expected %d bids submitted, got %v", expectedSubmitted, bids["submitted"])
	}
}

func TestResourceStatsMinMax(t *testing.T) {
	metrics := NewMetrics()

	// Record resource usage with different values
	usage1 := &types.ResourceUsage{
		CPUUsed: big.NewInt(1),
		GPUUsed: big.NewInt(10),
	}
	usage2 := &types.ResourceUsage{
		CPUUsed: big.NewInt(5),
		GPUUsed: big.NewInt(5),
	}
	usage3 := &types.ResourceUsage{
		CPUUsed: big.NewInt(10),
		GPUUsed: big.NewInt(1),
	}

	metrics.RecordResourceUtilization(usage1)
	metrics.RecordResourceUtilization(usage2)
	metrics.RecordResourceUtilization(usage3)

	stats := metrics.GetStats()
	resources := stats["resources"].(map[string]interface{})
	utilization := resources["utilization"].(map[string]interface{})
	utilCPU := utilization["utilization"].(map[string]interface{})["cpu"].(map[string]interface{})

	// Check min/max values
	if utilCPU["min"] != "1" {
		t.Errorf("Expected CPU min 1, got %v", utilCPU["min"])
	}

	if utilCPU["max"] != "10" {
		t.Errorf("Expected CPU max 10, got %v", utilCPU["max"])
	}

	utilGPU := utilization["utilization"].(map[string]interface{})["gpu"].(map[string]interface{})
	if utilGPU["min"] != "1" {
		t.Errorf("Expected GPU min 1, got %v", utilGPU["min"])
	}

	if utilGPU["max"] != "10" {
		t.Errorf("Expected GPU max 10, got %v", utilGPU["max"])
	}
}
