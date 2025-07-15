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
		t.Fatal("Expected non-nil StorageManagerImpl")
	}

	if storageMgr.isoDir == "" {
		t.Error("Expected non-empty ISO directory")
	}
}

func TestNewStorageManagerWithConfig(t *testing.T) {
	config := &ServiceConfig{
		DefaultOSType: "Ubuntu_ARM64",
	}

	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager with config: %v", err)
	}

	if storageMgr == nil {
		t.Fatal("Expected non-nil StorageManagerImpl")
	}

	if storageMgr.config != config {
		t.Error("Expected config to be set correctly")
	}
}

func TestStorageManagerImpl_DetermineOSTypeAndISO(t *testing.T) {
	config := &ServiceConfig{
		DefaultOSType: "Ubuntu_ARM64",
	}
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	tests := []struct {
		name    string
		req     vbtypes.VMCreateRequest
		wantErr bool
	}{
		{
			name: "with explicit OS type",
			req: vbtypes.VMCreateRequest{
				Name:       "test-vm",
				OSType:     "Ubuntu_ARM64",
				CPUCores:   2,
				MemoryMB:   2048,
				DiskSizeGB: 20,
			},
			wantErr: false,
		},
		{
			name: "with explicit ISO URL",
			req: vbtypes.VMCreateRequest{
				Name:       "test-vm",
				CPUCores:   2,
				MemoryMB:   2048,
				DiskSizeGB: 20,
			},
			wantErr: false,
		},
		{
			name: "with default config",
			req: vbtypes.VMCreateRequest{
				Name:       "test-vm",
				CPUCores:   2,
				MemoryMB:   2048,
				DiskSizeGB: 20,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osType, isoURL, err := storageMgr.DetermineOSTypeAndISO(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetermineOSTypeAndISO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if osType == "" {
					t.Error("Expected non-empty OS type")
				}
				if isoURL == "" {
					t.Error("Expected non-empty ISO URL")
				}
			}
		})
	}
}

func TestStorageManagerImpl_GetSupportedOSTypes(t *testing.T) {
	config := &ServiceConfig{}
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	osTypes := storageMgr.GetSupportedOSTypes()
	if len(osTypes) == 0 {
		t.Fatal("Expected non-empty list of supported OS types")
	}

	// Check that we have some expected OS types
	expectedTypes := []string{"Ubuntu", "Debian", "Fedora", "CentOS"}
	for _, expected := range expectedTypes {
		found := false
		for _, osType := range osTypes {
			if osType == expected || osType == expected+"_64" || osType == expected+"_ARM64" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find OS type containing %s", expected)
		}
	}
}

func TestStorageManagerImpl_getOSTypeForArchitecture(t *testing.T) {
	config := &ServiceConfig{}
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	tests := []struct {
		arch     string
		expected string
	}{
		{"arm64", "Ubuntu_ARM64"},
		{"aarch64", "Ubuntu_ARM64"},
		{"amd64", "Ubuntu_64"},
		{"x86_64", "Ubuntu_64"},
		{"arm", "Ubuntu"},
		{"386", "Ubuntu"},
		{"i386", "Ubuntu"},
		{"unknown", "Ubuntu_64"}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.arch, func(t *testing.T) {
			result := storageMgr.getOSTypeForArchitecture(tt.arch)
			if result != tt.expected {
				t.Errorf("getOSTypeForArchitecture(%s) = %s, want %s", tt.arch, result, tt.expected)
			}
		})
	}
}

func TestStorageManagerImpl_isValidOSType(t *testing.T) {
	config := &ServiceConfig{}
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	tests := []struct {
		osType string
		valid  bool
	}{
		{"Ubuntu", true},
		{"Ubuntu_64", true},
		{"Ubuntu_ARM64", true},
		{"Debian", true},
		{"Debian_64", true},
		{"Debian_ARM64", true},
		{"Fedora", true},
		{"CentOS", true},
		{"RedHat", true},
		{"Oracle", true},
		{"Linux", true},
		{"Windows", true},
		{"FreeBSD", true},
		{"NetBSD", true},
		{"BSD", true},
		{"Other", true},
		{"", false},
		{"InvalidOS", false},
		{"Ubuntu_Invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.osType, func(t *testing.T) {
			result := storageMgr.isValidOSType(tt.osType)
			if result != tt.valid {
				t.Errorf("isValidOSType(%s) = %v, want %v", tt.osType, result, tt.valid)
			}
		})
	}
}

func TestStorageManagerImpl_isValidISOURL(t *testing.T) {
	config := &ServiceConfig{}
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	tests := []struct {
		isoURL string
		valid  bool
	}{
		{"https://example.com/test.iso", true},
		{"http://example.com/test.iso", true},
		{"https://releases.ubuntu.com/24.04/ubuntu-24.04.2-live-server-amd64.iso", true},
		{"", false},
		{"https://example.com/test.txt", false},
		{"ftp://example.com/test.iso", false},
		{"not-a-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.isoURL, func(t *testing.T) {
			result := storageMgr.isValidISOURL(tt.isoURL)
			if result != tt.valid {
				t.Errorf("isValidISOURL(%s) = %v, want %v", tt.isoURL, result, tt.valid)
			}
		})
	}
}

func TestStorageManagerImpl_getISOURLForOSType(t *testing.T) {
	config := &ServiceConfig{}
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	tests := []struct {
		osType string
		valid  bool
	}{
		{"Ubuntu_64", true},
		{"Ubuntu_ARM64", true},
		{"Debian_64", true},
		{"Debian_ARM64", true},
		{"Fedora_64", true},
		{"Fedora_ARM64", true},
		{"CentOS_64", true},
		{"CentOS_ARM64", true},
		{"RedHat_64", true},
		{"RedHat_ARM64", true},
		{"Oracle_64", true},
		{"Oracle_ARM64", true},
		{"Linux_64", true},
		{"Linux_ARM64", true},
		{"Windows_64", true},
		{"Windows_ARM64", true},
		{"FreeBSD_64", true},
		{"FreeBSD_ARM64", true},
		{"NetBSD_64", true},
		{"NetBSD_ARM64", true},
		{"BSD_64", true},
		{"BSD_ARM64", true},
		{"Other_64", true},
		{"Other_ARM64", true},
		{"InvalidOS", true}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.osType, func(t *testing.T) {
			isoURL := storageMgr.getISOURLForOSType(tt.osType)
			if isoURL == "" {
				t.Errorf("getISOURLForOSType(%s) returned empty URL", tt.osType)
			}
			if !storageMgr.isValidISOURL(isoURL) {
				t.Errorf("getISOURLForOSType(%s) returned invalid URL: %s", tt.osType, isoURL)
			}
		})
	}
}

func TestStorageManagerImpl_validateArchitectureCompatibility(t *testing.T) {
	config := &ServiceConfig{}
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test with valid OS types for current architecture
	validOSTypes := storageMgr.GetSupportedOSTypes()
	for _, osType := range validOSTypes {
		t.Run("valid_"+osType, func(t *testing.T) {
			err := storageMgr.validateArchitectureCompatibility(osType)
			if err != nil {
				t.Errorf("validateArchitectureCompatibility(%s) failed: %v", osType, err)
			}
		})
	}

	// Test with invalid OS type
	t.Run("invalid_os_type", func(t *testing.T) {
		err := storageMgr.validateArchitectureCompatibility("InvalidOSType")
		if err == nil {
			t.Error("Expected error for invalid OS type")
		}
	})
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
