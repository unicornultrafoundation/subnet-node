package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConvertVirtualIPToNumber tests the ConvertVirtualIPToNumber function
func TestConvertVirtualIPToNumber(t *testing.T) {
	tests := []struct {
		name      string
		virtualIP string
		expected  uint32
	}{
		{
			name:      "Valid IP in 10.0.0.0/8 range",
			virtualIP: "10.0.0.1",
			expected:  167772161,
		},
		{
			name:      "IP not in 10.0.0.0/8 range",
			virtualIP: "192.168.1.1",
			expected:  0,
		},
		{
			name:      "Invalid IP",
			virtualIP: "not an IP",
			expected:  0,
		},
		{
			name:      "IPv6 address",
			virtualIP: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected:  0,
		},
		{
			name:      "Empty string",
			virtualIP: "",
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertVirtualIPToNumber(tt.virtualIP)
			assert.Equal(t, tt.expected, result)
		})
	}
}
