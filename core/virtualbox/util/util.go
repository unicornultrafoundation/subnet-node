package util

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func NormalizeUnit(unit string) string {
	switch strings.ToLower(unit) {
	case "%":
		return "%"
	case "kb", "k":
		return "KB"
	case "mb", "m":
		return "MB"
	case "gb", "g":
		return "GB"
	case "b/s":
		return "B/s"
	case "kb/s", "k/s":
		return "KB/s"
	case "mb/s", "m/s":
		return "MB/s"
	case "gb/s", "g/s":
		return "GB/s"
	default:
		return unit
	}
}

// find available TCP port in range 20000-30000
func FindAvailableTCPPort(start, end int) (int, error) {
	for port := start; port <= end; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close() // Close the listener if successful
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found in range %d-%d", start, end)
}

// metricParseData is a temporary struct for parsing VBoxManage output
type metricParseData struct {
	Timestamp string
	Metric    string
	Value     float64
	Unit      string
}

func ParseVBoxManageOutput(line string) (*metricParseData, error) {
	// Skip header lines and empty lines
	line = strings.TrimSpace(line)
	if line == "" || strings.Contains(line, "----") {
		return nil, nil
	}

	// Skip header lines that don't start with a timestamp
	// VBoxManage metrics output starts with timestamp like "04:19:01.615"
	if !regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3}`).MatchString(line) {
		// This is not a metric data line, skip it
		return nil, nil
	}

	// Parse the format: "04:19:01.615 vm-1       CPU/Load/User        5.11%" or "04:19:01.615 vm-1       RAM/Usage/Used       118816 kB"
	// Using regex to handle variable spacing and different units including kB, MB, B/s
	re := regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}\.\d{3})\s+(\S+)\s+([^\s]+(?:\s+[^\s]+)*)\s+([\d.]+)\s*([%]|[kK]?[bB]|[mM]?[bB]|[gG]?[bB]|[bB]/s|[kK][bB]/s|[mM][bB]/s|[gG][bB]/s)?$`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 6 {
		return nil, fmt.Errorf("could not parse line: %s", line)
	}

	timestamp := matches[1]
	metric := strings.TrimSpace(matches[3])
	valueStr := matches[4]
	unit := matches[5]

	// Convert value to float64
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse value '%s': %v", valueStr, err)
	}

	// Convert timestamp to full datetime
	now := time.Now()
	timeStr := fmt.Sprintf("%04d-%02d-%02d %s", now.Year(), now.Month(), now.Day(), timestamp)
	parsedTime, err := time.Parse("2006-01-02 15:04:05.000", timeStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse timestamp '%s': %v", timeStr, err)
	}

	// Normalize units for better consistency
	normalizedUnit := NormalizeUnit(unit)

	return &metricParseData{
		Timestamp: parsedTime.Format(time.RFC3339),
		Metric:    metric,
		Value:     value,
		Unit:      normalizedUnit,
	}, nil
}
