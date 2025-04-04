package vpn_test

import (
	"testing"

	"github.com/unicornultrafoundation/subnet-node/core/vpn"
)

func TestGetAppIDFromVirtualIP(t *testing.T) {
	testCases := []struct {
		virtualIP string
		expected  int64
		shouldErr bool
	}{
		{"10.0.0.2", 2, false},
		{"10.1.1.1", 64771, false},
		{"10.1.1.254", 65024, false},
		{"10.1.2.1", 65025, false},
		{"10.1.2.6", 65030, false},
		{"10.254.254.254", 16451834, false},
		// Edge Cases
		{"10.0.0.0", 0, true},       // Invalid reserved IP
		{"10.255.255.255", 0, true}, // Invalid broadcast address
		{"invalid_ip", 0, true},     // Invalid format
	}

	for _, tc := range testCases {
		appID, err := vpn.GetAppIDFromVirtualIP(tc.virtualIP)
		if (err != nil) != tc.shouldErr {
			t.Errorf("Unexpected error for IP %s: %v", tc.virtualIP, err)
		}
		if appID != tc.expected {
			t.Errorf("For IP %s, expected AppID %d but got %d", tc.virtualIP, tc.expected, appID)
		}
	}
}

func TestGetVirtualIPFromAppID(t *testing.T) {
	testCases := []struct {
		appID     int
		expected  string
		shouldErr bool
	}{
		{1, "11.0.0.1", false},
		{254, "11.0.0.254", false},
		{255, "11.0.1.1", false},
		{260, "11.0.1.6", false},
		// Edge Cases
		{0, "", true}, // Out of range
	}

	for _, tc := range testCases {
		ip, err := vpn.GetVirtualIPFromAppID(int64(tc.appID))
		if (err != nil) != tc.shouldErr {
			t.Errorf("Unexpected error for AppID %d: %v", tc.appID, err)
		}
		if ip != tc.expected {
			t.Errorf("For AppID %d, expected IP %s but got %s", tc.appID, tc.expected, ip)
		}
	}
}

func TestBidirectionalMapping(t *testing.T) {
	var appID int64
	for appID = 1; appID <= 1000; appID++ { // Test a range of values
		ip, err := vpn.GetVirtualIPFromAppID(appID)
		if err != nil {
			t.Errorf("Failed to get IP for AppID %d: %v", appID, err)
			continue
		}

		reverseAppID, err := vpn.GetAppIDFromVirtualIP(ip)
		if err != nil {
			t.Errorf("Failed to get AppID for IP %s: %v", ip, err)
			continue
		}

		if reverseAppID != appID {
			t.Errorf("Mismatch: AppID %d mapped to IP %s but reversed to AppID %d", appID, ip, reverseAppID)
		}
	}
}
