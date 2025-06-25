# Metrics Package

The metrics package provides a comprehensive metrics collection system for the bidengine, implementing the `types.Metrics` interface with thread-safe counters, gauges, and histograms.

## Features

- **Thread-safe operations**: All metrics operations are thread-safe using atomic operations and mutexes
- **Bid metrics**: Track submitted, accepted, and rejected bids with latency measurements
- **Order metrics**: Monitor tracked and completed orders with duration tracking
- **Resource metrics**: Record resource utilization and allocation statistics
- **Financial metrics**: Track revenue, profit, and cost amounts
- **Statistical analysis**: Calculate totals, averages, minimums, and maximums
- **JSON serialization**: Export metrics in a structured format for monitoring

## Usage

### Basic Usage

```go
import "github.com/unicornultrafoundation/subnet-node/bidengine/metrics"

// Create a new metrics instance
metricsService := metrics.NewMetrics()

// Record bid metrics
metricsService.IncrementBidsSubmitted()
metricsService.IncrementBidsAccepted()
metricsService.RecordBidLatency(1.5)

// Record order metrics
metricsService.IncrementOrdersTracked()
metricsService.IncrementOrdersCompleted()
metricsService.RecordOrderDuration(10.5)

// Record resource metrics
usage := &types.ResourceUsage{
    CPUUsed:     big.NewInt(4),
    GPUUsed:     big.NewInt(2),
    MemoryUsed:  big.NewInt(8192),
    DiskUsed:    big.NewInt(100),
    NetworkUsed: big.NewInt(1000),
}
metricsService.RecordResourceUtilization(usage)
metricsService.RecordResourceAllocation(usage)

// Record financial metrics
metricsService.RecordRevenue(big.NewInt(1000))
metricsService.RecordProfit(big.NewInt(500))
metricsService.RecordCost(big.NewInt(300))

// Get metrics snapshot
stats := metricsService.GetStats()
```

### Metrics Structure

The `GetStats()` method returns a structured map with the following hierarchy:

```json
{
  "bids": {
    "submitted": 10,
    "accepted": 8,
    "rejected": 2,
    "latency": {
      "total": 15.5,
      "count": 10,
      "average": 1.55
    }
  },
  "orders": {
    "tracked": 25,
    "completed": 20,
    "duration": {
      "total": 500.0,
      "count": 20,
      "average": 25.0
    }
  },
  "resources": {
    "utilization": {
      "utilization": {
        "cpu": {
          "total": "100",
          "count": 10,
          "average": "10",
          "min": "1",
          "max": "20"
        },
        "gpu": { ... },
        "memory": { ... },
        "disk": { ... },
        "network": { ... }
      }
    },
    "allocation": { ... }
  },
  "financial": {
    "revenue": "10000",
    "profit": "5000",
    "cost": "5000"
  }
}
```

## API Reference

### Constructor

- `NewMetrics() *MetricsService`: Creates a new metrics service instance

### Bid Metrics

- `IncrementBidsSubmitted()`: Increments the count of submitted bids
- `IncrementBidsAccepted()`: Increments the count of accepted bids
- `IncrementBidsRejected()`: Increments the count of rejected bids
- `RecordBidLatency(duration float64)`: Records the latency of a bid operation

### Order Metrics

- `IncrementOrdersTracked()`: Increments the count of tracked orders
- `IncrementOrdersCompleted()`: Increments the count of completed orders
- `RecordOrderDuration(duration float64)`: Records the duration of an order

### Resource Metrics

- `RecordResourceUtilization(usage *types.ResourceUsage)`: Records resource utilization metrics
- `RecordResourceAllocation(usage *types.ResourceUsage)`: Records resource allocation metrics

### Financial Metrics

- `RecordRevenue(amount *big.Int)`: Records revenue amount
- `RecordProfit(amount *big.Int)`: Records profit amount
- `RecordCost(amount *big.Int)`: Records cost amount

### Utility Methods

- `GetStats() map[string]interface{}`: Returns a snapshot of all metrics
- `Reset()`: Resets all metrics to zero

## Thread Safety

All metrics operations are thread-safe:

- **Atomic operations**: Used for simple counters (bids, orders)
- **Mutex protection**: Used for complex operations (resource stats, financial totals)
- **Concurrent access**: Multiple goroutines can safely access the same metrics instance

## Performance Considerations

- **Memory efficient**: Uses `big.Int` for large numbers and financial amounts
- **Lock-free counters**: Simple counters use atomic operations for maximum performance
- **Minimal locking**: Complex operations use fine-grained locking to minimize contention
- **Efficient serialization**: Stats are computed on-demand and cached appropriately

## Integration with BidEngine

The metrics package is designed to integrate seamlessly with the bidengine components:

```go
// In factory.go
metrics := metrics.NewMetrics()

// Pass to components
resourceManager := managerpkg.NewResourceManager(config, provider, logger, metrics, datastore)
bidManager := managerpkg.NewManager(config, bidMarket, logger, metrics, datastore, resourceManager, pricingEngine)
orderMonitor := NewMonitor(config, bidMarket, logger, metrics, datastore, bidManager)
```

## Testing

The package includes comprehensive tests covering:

- Basic functionality
- Thread safety
- Edge cases (nil values)
- Statistical calculations
- Concurrent access patterns

Run tests with:

```bash
go test ./bidengine/metrics
```

## Monitoring Integration

The metrics can be easily integrated with monitoring systems:

```go
// Export metrics for Prometheus
func exportPrometheusMetrics(metrics *metrics.MetricsService) {
    stats := metrics.GetStats()
    
    // Convert to Prometheus format
    bidsSubmitted.Set(float64(stats["bids"].(map[string]interface{})["submitted"].(int64)))
    bidsAccepted.Set(float64(stats["bids"].(map[string]interface{})["accepted"].(int64)))
    // ... etc
}

// Export metrics for JSON API
func getMetricsJSON(metrics *metrics.MetricsService) []byte {
    stats := metrics.GetStats()
    jsonData, _ := json.Marshal(stats)
    return jsonData
}
``` 