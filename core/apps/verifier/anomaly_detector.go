package verifier

import (
	"math"

	pvtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/app/verifier"
)

// AnomalyDetector contains data and logic for detecting anomalies
type AnomalyDetector struct {
	Logs      []*pvtypes.UsageReport
	Threshold float64
	MeanStd   map[string]struct{ mean, std float64 }
}

// Calculate mean and standard deviation
func calculateMeanStd(data []float64) (mean, std float64) {
	n := float64(len(data))
	if n == 0 {
		return 0, 0
	}

	var sum float64
	for _, v := range data {
		sum += v
	}
	mean = sum / n

	var variance float64
	for _, v := range data {
		variance += math.Pow(v-mean, 2)
	}
	std = math.Sqrt(variance / n)
	return mean, std
}

// Method to compute mean and standard deviation
func (a *AnomalyDetector) computeMeanStd() {
	peerStats := map[string][]float64{
		"CPU":           {},
		"GPU":           {},
		"Memory":        {},
		"UploadBytes":   {},
		"DownloadBytes": {},
		"Storage":       {},
	}

	// Collect data
	for _, log := range a.Logs {
		peerStats["CPU"] = append(peerStats["CPU"], float64(log.Cpu))
		peerStats["GPU"] = append(peerStats["GPU"], float64(log.Cpu))
		peerStats["Memory"] = append(peerStats["Memory"], float64(log.Memory))
		peerStats["UploadBytes"] = append(peerStats["UploadBytes"], float64(log.UploadBytes))
		peerStats["DownloadBytes"] = append(peerStats["DownloadBytes"], float64(log.DownloadBytes))
		peerStats["Storage"] = append(peerStats["Storage"], float64(log.Storage))
	}

	// Compute mean and standard deviation for each resource type
	a.MeanStd = make(map[string]struct{ mean, std float64 })
	for metric, values := range peerStats {
		mean, std := calculateMeanStd(values)
		a.MeanStd[metric] = struct{ mean, std float64 }{mean, std}
	}
}

// Method to detect abnormal peers
func (a *AnomalyDetector) detect() map[string]int {
	a.computeMeanStd() // Compute mean and standard deviation
	peerScores := make(map[string]int)
	for _, log := range a.Logs {
		score := 0
		for metric, stats := range a.MeanStd {
			var value float64
			switch metric {
			case "CPU":
				value = float64(log.Cpu)
			case "GPU":
				value = float64(log.Gpu)
			case "Memory":
				value = float64(log.Memory)
			case "DownloadBytes":
				value = float64(log.DownloadBytes)
			case "UploadBytes":
				value = float64(log.UploadBytes)
			case "Storage":
				value = float64(log.Storage)
			}

			// Check if it exceeds mean ± threshold * std
			if math.Abs(value-stats.mean) > a.Threshold*stats.std {
				score-- // Abnormal
			} else {
				score++ // Normal
			}
		}
		peerScores[log.PeerId] += score
	}

	return peerScores
}
