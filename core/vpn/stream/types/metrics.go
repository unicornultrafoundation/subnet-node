package types

// MultiplexerMetrics contains metrics for a multiplexer
type MultiplexerMetrics struct {
	// Packets sent through this multiplexer
	PacketsSent int64
	// Packets dropped by this multiplexer
	PacketsDropped int64
	// Bytes sent through this multiplexer
	BytesSent int64
	// Stream errors encountered
	StreamErrors int64
	// Streams created by this multiplexer
	StreamsCreated int64
	// Streams closed by this multiplexer
	StreamsClosed int64
	// Scale up operations performed
	ScaleUpOperations int64
	// Scale down operations performed
	ScaleDownOperations int64
	// Average latency in milliseconds
	AvgLatency int64
	// Number of latency measurements
	LatencyMeasurements int64
}

// GetMetrics returns the metrics as a map
func (m *MultiplexerMetrics) GetMetrics() map[string]int64 {
	return map[string]int64{
		"packets_sent":          m.PacketsSent,
		"packets_dropped":       m.PacketsDropped,
		"bytes_sent":            m.BytesSent,
		"stream_errors":         m.StreamErrors,
		"streams_created":       m.StreamsCreated,
		"streams_closed":        m.StreamsClosed,
		"scale_up_operations":   m.ScaleUpOperations,
		"scale_down_operations": m.ScaleDownOperations,
		"avg_latency":           m.AvgLatency,
		"latency_measurements":  m.LatencyMeasurements,
	}
}
