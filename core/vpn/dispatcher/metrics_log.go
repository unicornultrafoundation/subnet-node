package dispatcher

import (
	"github.com/sirupsen/logrus"
)

// logMetricsPeriodically periodically logs the dispatcher metrics
func (d *Dispatcher) logMetricsPeriodically() {
	for {
		select {
		case <-d.stopChan:
			return
		case <-d.metricsLogTicker.C:
			metrics := d.GetMetrics()

			// Log the most important metrics
			log.WithFields(logrus.Fields{
				"active_streams":       metrics["active_streams_total"],
				"packets_dispatched":   metrics["packets_dispatched_total"],
				"bytes_dispatched":     metrics["bytes_dispatched_total"],
				"streams_created":      metrics["streams_created_total"],
				"streams_closed":       metrics["streams_closed_total"],
				"stream_errors":        metrics["stream_errors_total"],
				"unique_peers":         metrics["unique_peers"],
				"avg_bytes_per_packet": metrics["avg_bytes_per_packet"],
			}).Info("VPN dispatcher metrics")

			// Log per-peer metrics if there are any peers
			if metrics["unique_peers"] > 0 {
				// Find top peers
				var topPeers []string
				for k := range metrics {
					if len(k) > 9 && k[:9] == "top_peer_" && k[len(k)-8:] == "_packets" {
						peerID := k[9 : len(k)-8]
						topPeers = append(topPeers, peerID)
					}
				}

				// Log metrics for each top peer
				for _, peerID := range topPeers {
					packets, packetsOK := metrics["top_peer_"+peerID+"_packets"]
					bytes, bytesOK := metrics["top_peer_"+peerID+"_bytes"]
					streams, streamsOK := metrics["top_peer_"+peerID+"_active_streams"]
					errors, errorsOK := metrics["top_peer_"+peerID+"_errors"]

					fields := logrus.Fields{
						"peer_id": peerID,
					}

					if packetsOK {
						fields["packets"] = packets
					}
					if bytesOK {
						fields["bytes"] = bytes
					}
					if streamsOK {
						fields["active_streams"] = streams
					}
					if errorsOK {
						fields["errors"] = errors
					}

					log.WithFields(fields).Info("VPN peer metrics")
				}
			}
		}
	}
}
