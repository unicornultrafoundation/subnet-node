package dispatcher

// GetMetrics returns the current metrics for the dispatcher
func (d *Dispatcher) GetMetrics() map[string]uint64 {
	// Use the sampled metrics instead of calculating on the fly
	sampledMetrics := d.GetSampledMetrics()

	// Convert the sampled metrics to a map
	metrics := make(map[string]uint64)

	// General metrics
	metrics["active_streams_total"] = sampledMetrics.ActiveStreamsTotal
	metrics["packets_dispatched_total"] = sampledMetrics.PacketsDispatchedTotal
	metrics["bytes_dispatched_total"] = sampledMetrics.BytesDispatchedTotal
	metrics["streams_created_total"] = sampledMetrics.StreamsCreatedTotal
	metrics["streams_closed_total"] = sampledMetrics.StreamsClosedTotal
	metrics["stream_errors_total"] = sampledMetrics.StreamErrorsTotal
	metrics["peer_discovery_errors_total"] = sampledMetrics.PeerDiscoveryErrorsTotal
	metrics["packet_parse_errors_total"] = sampledMetrics.PacketParseErrorsTotal
	metrics["avg_bytes_per_packet"] = sampledMetrics.AvgBytesPerPacket

	// Stream age metrics
	metrics["oldest_stream_age_seconds"] = sampledMetrics.OldestStreamAgeSeconds
	metrics["newest_stream_age_seconds"] = sampledMetrics.NewestStreamAgeSeconds
	metrics["avg_stream_age_seconds"] = sampledMetrics.AvgStreamAgeSeconds

	// Peer metrics
	metrics["unique_peers"] = sampledMetrics.UniquePeers

	// Add top peers
	for _, peer := range sampledMetrics.TopPeers {
		metrics["top_peer_"+peer.ShortPeerID+"_packets"] = peer.Packets
		metrics["top_peer_"+peer.ShortPeerID+"_bytes"] = peer.Bytes
		metrics["top_peer_"+peer.ShortPeerID+"_active_streams"] = peer.ActiveStreams
		metrics["top_peer_"+peer.ShortPeerID+"_errors"] = peer.Errors
	}

	return metrics
}
