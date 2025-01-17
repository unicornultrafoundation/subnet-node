package networkutil

import (
	"fmt"
	"net"
)

// DetectInternetInterfaces detects all network interfaces connected to the internet.
func DetectInternetInterfaces() ([]string, error) {
	// Get a list of all network interfaces on the system
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	var internetInterfaces []string
	// Iterate over each network interface
	for _, iface := range interfaces {
		// Get a list of addresses associated with the interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Iterate over each address associated with the interface
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}

			// Check if the IP address is a global unicast address
			if ipNet.IP.IsGlobalUnicast() {
				internetInterfaces = append(internetInterfaces, iface.Name)
			}
		}
	}

	// Return an error if no internet-connected interfaces are found
	if len(internetInterfaces) == 0 {
		return nil, fmt.Errorf("no internet-connected interfaces found")
	}

	return internetInterfaces, nil
}
