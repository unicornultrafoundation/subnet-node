package proxy

import (
	"fmt"
	"strconv"
	"strings"
)

// List of valid network protocols based on net/dial.go
var validProtocols = map[string]bool{
	"tcp": true, "tcp4": true, "tcp6": true,
	"udp": true, "udp4": true, "udp6": true,
	"unix": true, "unixpacket": true, "unixgram": true,
}

func ParsePortMappings(portMappings []string) ([]PortMapping, error) {
	var mappings []PortMapping

	for _, mapping := range portMappings {
		protocol := "tcp" // Default protocol

		// Extract protocol if provided
		if strings.Contains(mapping, "/") {
			parts := strings.Split(mapping, "/")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid format: %s", mapping)
			}
			mapping, protocol = parts[0], parts[1]

			// Validate protocol
			if !validProtocols[protocol] {
				return nil, fmt.Errorf("invalid protocol: %s", protocol)
			}
		}

		hostIP := ""
		mappingParts := strings.Split(mapping, ":")
		portParts := mappingParts

		// Handle IPv6 addresses like "[::]:8080:8000"
		if strings.HasPrefix(mapping, "[") {
			ipEnd := strings.Index(mapping, "]")
			if ipEnd == -1 || ipEnd+1 >= len(mapping) || mapping[ipEnd+1] != ':' {
				return nil, fmt.Errorf("invalid IPv6 format: %s", mapping)
			}

			hostIP = mapping[:ipEnd+1] // Extract `[IPv6]`
			portParts = strings.Split(mapping[ipEnd+2:], ":")
		} else if len(mappingParts) == 3 {
			hostIP = mappingParts[0]
			portParts = mappingParts[1:]
		}

		if len(portParts) < 1 || len(portParts) > 2 {
			return nil, fmt.Errorf("invalid port mapping: %s", mapping)
		}

		// Parse LocalPort (can be a range, e.g., "3000-3001")
		localPorts, err := parsePortRange(portParts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid local port: %s", portParts[0])
		}

		// If only LocalPort is provided, AppPort is set to 0
		if len(portParts) == 1 {
			for _, lp := range localPorts {
				mappings = append(mappings, PortMapping{HostIP: hostIP, LocalPort: lp, AppPort: 0, Protocol: protocol})
			}
			continue
		}

		// Parse AppPort
		appPorts, err := parsePortRange(portParts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid application port: %s", portParts[1])
		}

		if len(localPorts) > 1 && len(appPorts) > 1 && len(localPorts) != len(appPorts) {
			return nil, fmt.Errorf("mismatched port range: %s", mapping)
		}

		// Map LocalPort to AppPort
		for i := range localPorts {
			appPort := appPorts[0]
			if len(appPorts) == len(localPorts) {
				appPort = appPorts[i]
			}
			mappings = append(mappings, PortMapping{HostIP: hostIP, LocalPort: localPorts[i], AppPort: appPort, Protocol: protocol})
		}
	}

	return mappings, nil
}

// parsePortRange parses a single port or a range (e.g., "3000-3001")
func parsePortRange(portStr string) ([]int, error) {
	if strings.Contains(portStr, "-") {
		parts := strings.Split(portStr, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid port range: %s", portStr)
		}

		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || start > end || start < 1 || end > 65535 {
			return nil, fmt.Errorf("invalid port range: %s", portStr)
		}

		var ports []int
		for i := start; i <= end; i++ {
			ports = append(ports, i)
		}
		return ports, nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port > 65535 {
		return nil, fmt.Errorf("invalid port number: %s", portStr)
	}

	if port < 1024 {
		return nil, fmt.Errorf("port should not less than 1024, some system/kernel services will use that. You want to use port %s", portStr)
	}

	return []int{port}, nil
}
