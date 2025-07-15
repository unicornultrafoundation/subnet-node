package virtualbox

import (
	"context"
	"testing"

	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

func TestNewStorageManager(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	if storageMgr == nil {
		t.Error("Expected storage manager to be non-nil")
	}
}

func TestStorageManagerImpl_DetermineOSTypeAndISO(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test with explicit OS type
	req := vbtypes.VMCreateRequest{
		Name:       "test-vm",
		OSType:     "Ubuntu_ARM64",
		CPUCores:   2,
		MemoryMB:   2048,
		DiskSizeGB: 20,
	}

	ctx := context.Background()
	osType, isoURL, err := storageMgr.DetermineOSTypeAndISO(ctx, req)
	if err != nil {
		t.Fatalf("Failed to determine OS type and ISO: %v", err)
	}

	if osType != "Ubuntu_ARM64" {
		t.Errorf("Expected OS type Ubuntu_ARM64, got %s", osType)
	}

	if isoURL == "" {
		t.Error("Expected ISO URL to be non-empty")
	}

	// Test without explicit OS type (should use hardware detection)
	req.OSType = ""
	osType2, isoURL2, err := storageMgr.DetermineOSTypeAndISO(ctx, req)
	if err != nil {
		t.Fatalf("Failed to determine OS type and ISO without explicit type: %v", err)
	}

	if osType2 == "" {
		t.Error("Expected OS type to be determined automatically")
	}

	if isoURL2 == "" {
		t.Error("Expected ISO URL to be non-empty")
	}
}

func TestStorageManagerImpl_GetSupportedOSTypes(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	osTypes := storageMgr.GetSupportedOSTypes()
	if len(osTypes) == 0 {
		t.Error("Expected at least one supported OS type")
	}

	// Check for common OS types
	expectedTypes := []string{"Ubuntu_64", "Ubuntu_ARM64", "Ubuntu"}
	found := 0
	for _, expected := range expectedTypes {
		for _, actual := range osTypes {
			if actual == expected {
				found++
				break
			}
		}
	}

	if found == 0 {
		t.Error("Expected to find at least one common OS type")
	}
}

func TestStorageManagerImpl_getOSTypeForArchitecture(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test ARM64 architecture
	osType := storageMgr.getOSTypeForArchitecture("arm64")
	if osType != "Ubuntu_ARM64" {
		t.Errorf("Expected Ubuntu_ARM64 for arm64, got %s", osType)
	}

	// Test AMD64 architecture
	osType = storageMgr.getOSTypeForArchitecture("amd64")
	if osType != "Ubuntu_64" {
		t.Errorf("Expected Ubuntu_64 for amd64, got %s", osType)
	}

	// Test unknown architecture
	osType = storageMgr.getOSTypeForArchitecture("unknown")
	if osType != "Ubuntu_64" {
		t.Errorf("Expected Ubuntu_64 for unknown architecture, got %s", osType)
	}
}

func TestStorageManagerImpl_isValidOSType(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test valid OS types
	validTypes := []string{"Ubuntu_64", "Ubuntu_ARM64", "Ubuntu", "Debian_64"}
	for _, osType := range validTypes {
		if !storageMgr.isValidOSType(osType) {
			t.Errorf("Expected %s to be valid", osType)
		}
	}

	// Test invalid OS types
	invalidTypes := []string{"", "InvalidOS", "Windows_Invalid"}
	for _, osType := range invalidTypes {
		if storageMgr.isValidOSType(osType) {
			t.Errorf("Expected %s to be invalid", osType)
		}
	}
}

func TestStorageManagerImpl_isValidISOURL(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test valid ISO URLs
	validURLs := []string{
		"https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso",
		"http://example.com/test.iso",
	}
	for _, url := range validURLs {
		if !storageMgr.isValidISOURL(url) {
			t.Errorf("Expected %s to be valid", url)
		}
	}

	// Test invalid ISO URLs
	invalidURLs := []string{
		"",
		"not-a-url",
		"https://example.com/file.txt",
		"ftp://example.com/file.iso",
	}
	for _, url := range invalidURLs {
		if storageMgr.isValidISOURL(url) {
			t.Errorf("Expected %s to be invalid", url)
		}
	}
}

func TestStorageManagerImpl_getISOURLForOSType(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test Ubuntu_64
	url := storageMgr.getISOURLForOSType("Ubuntu_64")
	if url == "" {
		t.Error("Expected non-empty URL for Ubuntu_64")
	}
	if !storageMgr.isValidISOURL(url) {
		t.Errorf("Expected valid URL for Ubuntu_64, got %s", url)
	}

	// Test Ubuntu_ARM64
	url = storageMgr.getISOURLForOSType("Ubuntu_ARM64")
	if url == "" {
		t.Error("Expected non-empty URL for Ubuntu_ARM64")
	}
	if !storageMgr.isValidISOURL(url) {
		t.Errorf("Expected valid URL for Ubuntu_ARM64, got %s", url)
	}

	// Test unknown OS type
	url = storageMgr.getISOURLForOSType("UnknownOS")
	if url != "" {
		t.Errorf("Expected empty URL for unknown OS type, got %s", url)
	}
}

func TestStorageManagerImpl_validateArchitectureCompatibility(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test compatible combinations
	compatibleTests := []struct {
		arch       string
		osType     string
		shouldPass bool
	}{
		{"arm64", "Ubuntu_ARM64", true},
		{"amd64", "Ubuntu_64", true},
		{"arm64", "Ubuntu_64", false},    // ARM64 host with x64 OS
		{"amd64", "Ubuntu_ARM64", false}, // x64 host with ARM64 OS
	}

	for _, test := range compatibleTests {
		err := storageMgr.validateArchitectureCompatibility(test.osType)
		if test.shouldPass && err != nil {
			t.Errorf("Expected compatibility check to pass for %s on %s, but got error: %v", test.osType, test.arch, err)
		}
		if !test.shouldPass && err == nil {
			t.Errorf("Expected compatibility check to fail for %s on %s, but it passed", test.osType, test.arch)
		}
	}
}

func TestStorageManagerImpl_GetISODir(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	isoDir := storageMgr.GetISODir()
	if isoDir == "" {
		t.Error("Expected non-empty ISO directory")
	}
}

func TestStorageManagerImpl_GetISOFileName(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com/test.iso", "test.iso"},
		{"https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso", "ubuntu-24.04.2-live-server-amd64.iso"},
		{"http://example.com/path/to/file.iso", "file.iso"},
		{"", "unknown.iso"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := storageMgr.GetISOFileName(tt.url)
			if result != tt.expected {
				t.Errorf("GetISOFileName(%s) = %s, want %s", tt.url, result, tt.expected)
			}
		})
	}
}

func TestStorageManagerImpl_GetISOPath(t *testing.T) {
	storageMgr, err := NewStorageManager()
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	isoURL := "https://example.com/test.iso"
	isoPath := storageMgr.GetISOPath(isoURL)

	if isoPath == "" {
		t.Error("Expected non-empty ISO path")
	}

	// Should contain the ISO directory and filename
	if !storageMgr.FileExists(storageMgr.GetISODir()) {
		t.Error("Expected ISO directory to exist")
	}
}
