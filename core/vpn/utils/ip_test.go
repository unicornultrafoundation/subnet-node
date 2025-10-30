package utils

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
			name:      "Valid IP in 100.66.0.0/15 range",
			virtualIP: "100.71.255.254",
			expected:  167772161,
		},
		{
			name:      "Valid IP in 100.66.0.0/15 range 2",
			virtualIP: "100.66.0.254",
			expected:  167772414, // 167772161 + 253
		},
		{
			name:      "Valid IP in 100.66.0.0/15 range 3",
			virtualIP: "100.66.1.1",
			expected:  167772415, // 167772161 + 254
		},
		{
			name:      "Valid IP in 100.66.0.0/15 range 4",
			virtualIP: "100.66.1.2",
			expected:  167772416, // 167772161 + 254 + 1
		},
		{
			name:      "Valid IP in 100.66.0.0/15 range 5",
			virtualIP: "100.66.1.254",
			expected:  167772668, // 167772161 + 254 + 253
		},
		{
			name:      "Valid IP in 100.66.0.0/15 range 6",
			virtualIP: "100.66.2.1",
			expected:  167772669, // 167772161 + 254 + 254
		},
		{
			name:      "Valid IP in 100.66.0.0/15 range 7",
			virtualIP: "100.66.2.2",
			expected:  167772670, // 167772161 + 254 + 254 + 1
		},
		{
			name:      "Valid IP in 100.66.0.0/15 range 8",
			virtualIP: "100.66.10.1",
			expected:  167774701, // 167772161 + 254*10
		},
		{
			name:      "Valid IP in 100.66.0.0/15 range 9",
			virtualIP: "100.66.255.1",
			expected:  167772161 + 254*255,
		},
		{
			name:      "Valid IP in 100.67.0.0/15 range",
			virtualIP: "100.67.0.1",
			expected:  167772161 + 254*256, // 256 blocks = 65,024 offset
		},
		{
			name:      "Valid IP in 100.67.0.0/15 range 2",
			virtualIP: "100.67.0.254",
			expected:  167772161 + 254*256 + 253,
		},
		{
			name:      "Valid IP in 100.67.0.0/15 range 3",
			virtualIP: "100.67.1.1",
			expected:  167772161 + 254*257,
		},
		{
			name:      "Valid IP in 100.67.0.0/15 range 4",
			virtualIP: "100.67.255.1",
			expected:  167772161 + 254*511,
		},
		{
			name:      "Valid IP in 100.68.0.0/14 range",
			virtualIP: "100.68.0.1",
			expected:  167772161 + 254*512, // blockIndex = 512
		},
		{
			name:      "Valid IP in 100.68.0.0/14 range 2",
			virtualIP: "100.68.0.254",
			expected:  167772161 + 254*512 + 253,
		},
		{
			name:      "Valid IP in 100.68.0.0/14 range 3",
			virtualIP: "100.68.255.1",
			expected:  167772161 + 254*767, // blockIndex = 767
		},
		{
			name:      "Valid IP in 100.69.0.0/14 range",
			virtualIP: "100.69.0.1",
			expected:  167772161 + 254*768, // blockIndex = 768
		},
		{
			name:      "Valid IP in 100.70.0.0/14 range",
			virtualIP: "100.70.0.1",
			expected:  167772161 + 254*1024, // blockIndex = 1024
		},
		{
			name:      "Valid IP in 100.71.0.0/14 range",
			virtualIP: "100.71.0.1",
			expected:  167772161 + 254*1280, // blockIndex = 1280
		},
		{
			name:      "Valid IP in 100.71.0.0/14 range - last valid",
			virtualIP: "100.71.255.1",
			expected:  167772161 + 254*1535, // blockIndex = 1535
		},
		{
			name:      "IP not in 100.66.0.0/15 or 100.68.0.0/14 range",
			virtualIP: "192.168.1.1",
			expected:  0,
		},
		{
			name:      "IP not in 100.66.0.0/15 or 100.68.0.0/14 range - 10.x.x.x",
			virtualIP: "10.0.0.1",
			expected:  0,
		},
		{
			name:      "IP not in 100.66.0.0/15 or 100.68.0.0/14 range - different 2nd octet",
			virtualIP: "100.72.0.1",
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
			name:     "Valid IP in 100.66.0.0/15 range",
			expected: "100.66.0.1",
			tokenID:  167772161,
		},
		{
			name:     "Valid IP in 100.66.0.0/15 range 2",
			expected: "100.66.0.254",
			tokenID:  167772414, // 167772161 + 253
		},
		{
			name:     "Valid IP in 100.66.0.0/15 range 3",
			expected: "100.66.1.1",
			tokenID:  167772415, // 167772161 + 254
		},
		{
			name:     "Valid IP in 100.66.0.0/15 range 4",
			expected: "100.66.1.2",
			tokenID:  167772416, // 167772161 + 254 + 1
		},
		{
			name:     "Valid IP in 100.66.0.0/15 range 5",
			expected: "100.66.1.254",
			tokenID:  167772668, // 167772161 + 254 + 253
		},
		{
			name:     "Valid IP in 100.66.0.0/15 range 6",
			expected: "100.66.2.1",
			tokenID:  167772669, // 167772161 + 254 + 254
		},
		{
			name:     "Valid IP in 100.66.0.0/15 range 7",
			expected: "100.66.2.2",
			tokenID:  167772670, // 167772161 + 254 + 254 + 1
		},
		{
			name:     "Valid IP in 100.66.0.0/15 range 8",
			expected: "100.66.10.1",
			tokenID:  167774701, // 167772161 + 254*10
		},
		{
			name:     "Valid IP in 100.66.0.0/15 range 9",
			expected: "100.66.255.1",
			tokenID:  167772161 + 254*255,
		},
		{
			name:     "Valid IP in 100.67.0.0/15 range",
			expected: "100.67.0.1",
			tokenID:  167772161 + 254*256, // blockIndex = 256
		},
		{
			name:     "Valid IP in 100.67.0.0/15 range 2",
			expected: "100.67.0.254",
			tokenID:  167772161 + 254*256 + 253,
		},
		{
			name:     "Valid IP in 100.67.0.0/15 range 3",
			expected: "100.67.1.1",
			tokenID:  167772161 + 254*257,
		},
		{
			name:     "Valid IP in 100.67.0.0/15 range 4",
			expected: "100.67.255.1",
			tokenID:  167772161 + 254*511, // blockIndex = 511
		},
		{
			name:     "Valid IP in 100.68.0.0/14 range",
			expected: "100.68.0.1",
			tokenID:  167772161 + 254*512, // blockIndex = 512
		},
		{
			name:     "Valid IP in 100.68.0.0/14 range 2",
			expected: "100.68.0.254",
			tokenID:  167772161 + 254*512 + 253,
		},
		{
			name:     "Valid IP in 100.68.0.0/14 range 3",
			expected: "100.68.255.1",
			tokenID:  167772161 + 254*767, // blockIndex = 767
		},
		{
			name:     "Valid IP in 100.69.0.0/14 range",
			expected: "100.69.0.1",
			tokenID:  167772161 + 254*768, // blockIndex = 768
		},
		{
			name:     "Valid IP in 100.70.0.0/14 range",
			expected: "100.70.0.1",
			tokenID:  167772161 + 254*1024, // blockIndex = 1024
		},
		{
			name:     "Valid IP in 100.71.0.0/14 range",
			expected: "100.71.0.1",
			tokenID:  167772161 + 254*1280, // blockIndex = 1280
		},
		{
			name:     "Valid IP in 100.71.0.0/14 range - last valid",
			expected: "100.71.255.1",
			tokenID:  167772161 + 254*1535, // blockIndex = 1535
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

// TestGetIPType tests the GetIPType function
func TestGetIPType(t *testing.T) {
	tests := []struct {
		name      string
		virtualIP string
		expected  IPType
	}{
		{
			name:      "Static IP - 100.66.0.1",
			virtualIP: "100.66.0.1",
			expected:  STATIC_IP_TYPE,
		},
		{
			name:      "Static IP - 100.66.255.1",
			virtualIP: "100.66.255.1",
			expected:  STATIC_IP_TYPE,
		},
		{
			name:      "Static IP - 100.67.0.1",
			virtualIP: "100.67.0.1",
			expected:  STATIC_IP_TYPE,
		},
		{
			name:      "Static IP - 100.67.255.1",
			virtualIP: "100.67.255.1",
			expected:  STATIC_IP_TYPE,
		},
		{
			name:      "Dynamic IP - 100.68.0.1",
			virtualIP: "100.68.0.1",
			expected:  DYNAMIC_IP_TYPE,
		},
		{
			name:      "Dynamic IP - 100.68.255.1",
			virtualIP: "100.68.255.1",
			expected:  DYNAMIC_IP_TYPE,
		},
		{
			name:      "Dynamic IP - 100.69.0.1",
			virtualIP: "100.69.0.1",
			expected:  DYNAMIC_IP_TYPE,
		},
		{
			name:      "Dynamic IP - 100.70.0.1",
			virtualIP: "100.70.0.1",
			expected:  DYNAMIC_IP_TYPE,
		},
		{
			name:      "Dynamic IP - 100.71.0.1",
			virtualIP: "100.71.0.1",
			expected:  DYNAMIC_IP_TYPE,
		},
		{
			name:      "Dynamic IP - 100.71.255.1",
			virtualIP: "100.71.255.1",
			expected:  DYNAMIC_IP_TYPE,
		},
		{
			name:      "Unspecified IP - 192.168.1.1",
			virtualIP: "192.168.1.1",
			expected:  UNSPECIFIED_IP_TYPE,
		},
		{
			name:      "Unspecified IP - 10.0.0.1",
			virtualIP: "10.0.0.1",
			expected:  UNSPECIFIED_IP_TYPE,
		},
		{
			name:      "Unspecified IP - 100.65.0.1",
			virtualIP: "100.65.0.1",
			expected:  UNSPECIFIED_IP_TYPE,
		},
		{
			name:      "Unspecified IP - 100.72.0.1",
			virtualIP: "100.72.0.1",
			expected:  UNSPECIFIED_IP_TYPE,
		},
		{
			name:      "Unspecified IP - Invalid string",
			virtualIP: "not an IP",
			expected:  UNSPECIFIED_IP_TYPE,
		},
		{
			name:      "Unspecified IP - Empty string",
			virtualIP: "",
			expected:  UNSPECIFIED_IP_TYPE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetIPType(tt.virtualIP)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetTokenIPType tests the GetTokenIPType function
func TestGetTokenIPType(t *testing.T) {
	tests := []struct {
		name     string
		tokenID  uint32
		expected IPType
	}{
		{
			name:     "Static IP - Token for 100.66.0.1",
			tokenID:  167772161, // 100.66.0.1
			expected: STATIC_IP_TYPE,
		},
		{
			name:     "Static IP - Token for 100.66.255.1",
			tokenID:  167772161 + 254*255, // 100.66.255.1
			expected: STATIC_IP_TYPE,
		},
		{
			name:     "Static IP - Token for 100.67.0.1",
			tokenID:  167772161 + 254*256, // 100.67.0.1
			expected: STATIC_IP_TYPE,
		},
		{
			name:     "Static IP - Token for 100.67.255.1",
			tokenID:  167772161 + 254*511, // 100.67.255.1
			expected: STATIC_IP_TYPE,
		},
		{
			name:     "Dynamic IP - Token for 100.68.0.1",
			tokenID:  167772161 + 254*512, // 100.68.0.1
			expected: DYNAMIC_IP_TYPE,
		},
		{
			name:     "Dynamic IP - Token for 100.68.255.1",
			tokenID:  167772161 + 254*767, // 100.68.255.1
			expected: DYNAMIC_IP_TYPE,
		},
		{
			name:     "Dynamic IP - Token for 100.69.0.1",
			tokenID:  167772161 + 254*768, // 100.69.0.1
			expected: DYNAMIC_IP_TYPE,
		},
		{
			name:     "Dynamic IP - Token for 100.70.0.1",
			tokenID:  167772161 + 254*1024, // 100.70.0.1
			expected: DYNAMIC_IP_TYPE,
		},
		{
			name:     "Dynamic IP - Token for 100.71.0.1",
			tokenID:  167772161 + 254*1280, // 100.71.0.1
			expected: DYNAMIC_IP_TYPE,
		},
		{
			name:     "Dynamic IP - Token for 100.71.255.1",
			tokenID:  167772161 + 254*1535, // 100.71.255.1
			expected: DYNAMIC_IP_TYPE,
		},
		{
			name:     "Unspecified IP - Token ID 0",
			tokenID:  0,
			expected: UNSPECIFIED_IP_TYPE,
		},
		{
			name:     "Unspecified IP - Token ID below base",
			tokenID:  167772160, // base address
			expected: UNSPECIFIED_IP_TYPE,
		},
		{
			name:     "Unspecified IP - Token ID too high",
			tokenID:  167772161 + 254*1536, // blockIndex 1536, outside range
			expected: UNSPECIFIED_IP_TYPE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTokenIPType(tt.tokenID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
