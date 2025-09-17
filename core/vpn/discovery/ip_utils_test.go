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
			name:      "Valid IP in 10.0.0.0/8 range 2",
			virtualIP: "10.0.0.254",
			expected:  167772414, // 167772161 + 253
		},
		{
			name:      "Valid IP in 10.0.0.0/8 range 3",
			virtualIP: "10.0.1.1",
			expected:  167772415, // 167772161 + 254
		},
		{
			name:      "Valid IP in 10.0.0.0/8 range 4",
			virtualIP: "10.0.1.2",
			expected:  167772416, // 167772161 + 254 + 1
		},
		{
			name:      "Valid IP in 10.0.0.0/8 range 5",
			virtualIP: "10.0.1.254",
			expected:  167772668, // 167772161 + 254 + 253
		},
		{
			name:      "Valid IP in 10.0.0.0/8 range 6",
			virtualIP: "10.0.2.1",
			expected:  167772669, // 167772161 + 254 + 254
		},
		{
			name:      "Valid IP in 10.0.0.0/8 range 7",
			virtualIP: "10.0.2.2",
			expected:  167772670, // 167772161 + 254 + 254 + 1
		},
		{
			name:      "Valid IP in 10.0.0.0/8 range 8",
			virtualIP: "10.0.10.1",
			expected:  167774701, // 167772161 + 254*10
		},
		{
			name:      "Valid IP in 10.0.0.0/8 range 9",
			virtualIP: "10.0.255.1",
			expected:  167772161 + 254*255,
		},
		{
			name:      "Valid IP in 10.0.0.0/8 range 10",
			virtualIP: "10.1.0.1",
			expected:  167772161 + 254*255 + 254,
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

// TestConvertNumberToVirtualIP tests the ConvertNumberToVirtualIP function
func TestConvertNumberToVirtualIP(t *testing.T) {
	tests := []struct {
		name     string
		tokenID  uint32
		expected string
	}{
		{
			name:     "Valid IP in 10.0.0.0/8 range",
			expected: "10.0.0.1",
			tokenID:  167772161,
		},
		{
			name:     "Valid IP in 10.0.0.0/8 range 2",
			expected: "10.0.0.254",
			tokenID:  167772414, // 167772161 + 253
		},
		{
			name:     "Valid IP in 10.0.0.0/8 range 3",
			expected: "10.0.1.1",
			tokenID:  167772415, // 167772161 + 254
		},
		{
			name:     "Valid IP in 10.0.0.0/8 range 4",
			expected: "10.0.1.2",
			tokenID:  167772416, // 167772161 + 254 + 1
		},
		{
			name:     "Valid IP in 10.0.0.0/8 range 5",
			expected: "10.0.1.254",
			tokenID:  167772668, // 167772161 + 254 + 253
		},
		{
			name:     "Valid IP in 10.0.0.0/8 range 6",
			expected: "10.0.2.1",
			tokenID:  167772669, // 167772161 + 254 + 254
		},
		{
			name:     "Valid IP in 10.0.0.0/8 range 7",
			expected: "10.0.2.2",
			tokenID:  167772670, // 167772161 + 254 + 254 + 1
		},
		{
			name:     "Valid IP in 10.0.0.0/8 range 8",
			expected: "10.0.10.1",
			tokenID:  167774701, // 167772161 + 254*10
		},
		{
			name:     "Valid IP in 10.0.0.0/8 range 9",
			expected: "10.0.255.1",
			tokenID:  167772161 + 254*255,
		},
		{
			name:     "Valid IP in 10.0.0.0/8 range 10",
			expected: "10.1.0.1",
			tokenID:  167772161 + 254*255 + 254,
		},
		{
			name:     "Invalid token ID",
			tokenID:  0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertNumberToVirtualIP(tt.tokenID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
