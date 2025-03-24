package proxy

import (
	"reflect"
	"testing"
)

func TestParsePortMappings(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		expectError bool
		expected    []PortMapping
	}{
		{
			name:  "Single port mapping",
			input: []string{"3000:8000"},
			expected: []PortMapping{
				{HostIP: "", LocalPort: 3000, AppPort: 8000, Protocol: "tcp"},
			},
		},
		{
			name:  "Host IP Binding",
			input: []string{"192.168.1.100:3000:8000"},
			expected: []PortMapping{
				{HostIP: "192.168.1.100", LocalPort: 3000, AppPort: 8000, Protocol: "tcp"},
			},
		},
		{
			name:  "IPv6 Host IP Binding",
			input: []string{"[::]:8080:8080/tcp6"},
			expected: []PortMapping{
				{HostIP: "[::]", LocalPort: 8080, AppPort: 8080, Protocol: "tcp6"},
			},
		},
		{
			name:  "Port range to port range",
			input: []string{"3000-3001:8000-8001/tcp"},
			expected: []PortMapping{
				{HostIP: "", LocalPort: 3000, AppPort: 8000, Protocol: "tcp"},
				{HostIP: "", LocalPort: 3001, AppPort: 8001, Protocol: "tcp"},
			},
		},
		{
			name:  "Port range to single port",
			input: []string{"4000-4001:9000/udp4"},
			expected: []PortMapping{
				{HostIP: "", LocalPort: 4000, AppPort: 9000, Protocol: "udp4"},
				{HostIP: "", LocalPort: 4001, AppPort: 9000, Protocol: "udp4"},
			},
		},
		{
			name:  "Only host port (random container port)",
			input: []string{"3000"},
			expected: []PortMapping{
				{HostIP: "", LocalPort: 3000, AppPort: 0, Protocol: "tcp"},
			},
		},
		{
			name:        "Invalid port number",
			input:       []string{"999999:999999/tcp"},
			expectError: true,
		},
		{
			name:        "Invalid format",
			input:       []string{"invalid_port"},
			expectError: true,
		},
		{
			name:        "Invalid protocol",
			input:       []string{"3000:8000/invalid"},
			expectError: true,
		},
		{
			name:        "Mismatched port range",
			input:       []string{"3000-3002:8000-8001"},
			expectError: true,
		},
		{
			name:        "Port should more than 1024",
			input:       []string{"80:9999/tcp"},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mappings, err := ParsePortMappings(test.input)

			if test.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if !reflect.DeepEqual(mappings, test.expected) {
					t.Errorf("expected %v, got %v", test.expected, mappings)
				}
			}
		})
	}
}
