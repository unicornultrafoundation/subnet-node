package network

// VPNMetricsInterface defines the interface for VPN metrics
type VPNMetricsInterface interface {
	// IncrementPacketsReceived increments the packets received counter
	IncrementPacketsReceived(bytes int)
	// IncrementPacketsSent increments the packets sent counter
	IncrementPacketsSent(bytes int)
	// IncrementPacketsDropped increments the packets dropped counter
	IncrementPacketsDropped()
	// IncrementStreamErrors increments the stream errors counter
	IncrementStreamErrors()
	// IncrementCircuitOpenDrops increments the circuit open drops counter
	IncrementCircuitOpenDrops()
	// GetMetrics returns the current metrics as a map
	GetMetrics() map[string]int64
}
