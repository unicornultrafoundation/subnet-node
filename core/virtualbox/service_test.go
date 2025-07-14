package virtualbox

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/unicornultrafoundation/subnet-node/config"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

func TestLoadServiceConfig(t *testing.T) {
	// Create a test config
	cfg := config.NewC(nil)

	// Set some test values with proper nesting
	cfg.Settings = map[string]any{
		"virtualbox": map[string]any{
			"enable":               true,
			"default_memory_mb":    4096,
			"default_cpus":         4,
			"default_disk_size_gb": 40,
			"default_network_type": "bridged",
			"default_os_type":      "Ubuntu_64",
			"vm_start_timeout":     "120s",
			"vm_stop_timeout":      "60s",
			"vm_delete_timeout":    "120s",
			"monitor_interval":     "60s",
			"headless":             false,
			"network": map[string]any{
				"default_bridge_name": "en1",
				"enable_nat":          false,
				"enable_bridged":      true,
				"enable_hostonly":     true,
				"enable_internal":     false,
			},
			"storage": map[string]any{
				"default_controller": "virtioSCSI",
				"default_type":       "vmdk",
				"enable_trim":        false,
				"enable_compression": true,
			},
			"advanced": map[string]any{
				"enable_audio":         true,
				"enable_usb":           true,
				"enable_vrde":          false,
				"vrde_port":            3390,
				"enable_pae":           true,
				"enable_nested_paging": false,
				"enable_hw_virt":       false,
			},
		},
	}

	// Load the configuration
	serviceConfig := LoadServiceConfig(cfg)

	// Verify the configuration was loaded correctly
	if !serviceConfig.Enable {
		t.Error("Expected Enable to be true")
	}
	if serviceConfig.DefaultMemoryMB != 4096 {
		t.Errorf("Expected DefaultMemoryMB to be 4096, got %d", serviceConfig.DefaultMemoryMB)
	}
	if serviceConfig.DefaultCPUs != 4 {
		t.Errorf("Expected DefaultCPUs to be 4, got %d", serviceConfig.DefaultCPUs)
	}
	if serviceConfig.DefaultDiskSizeGB != 40 {
		t.Errorf("Expected DefaultDiskSizeGB to be 40, got %d", serviceConfig.DefaultDiskSizeGB)
	}
	if serviceConfig.DefaultNetworkType != "bridged" {
		t.Errorf("Expected DefaultNetworkType to be 'bridged', got %s", serviceConfig.DefaultNetworkType)
	}
	if serviceConfig.DefaultOSType != "Ubuntu_64" {
		t.Errorf("Expected DefaultOSType to be 'Ubuntu_64', got %s", serviceConfig.DefaultOSType)
	}
	if serviceConfig.VMStartTimeout != 120*time.Second {
		t.Errorf("Expected VMStartTimeout to be 120s, got %v", serviceConfig.VMStartTimeout)
	}
	if serviceConfig.VMStopTimeout != 60*time.Second {
		t.Errorf("Expected VMStopTimeout to be 60s, got %v", serviceConfig.VMStopTimeout)
	}
	if serviceConfig.VMDeleteTimeout != 120*time.Second {
		t.Errorf("Expected VMDeleteTimeout to be 120s, got %v", serviceConfig.VMDeleteTimeout)
	}
	if serviceConfig.MonitorInterval != 60*time.Second {
		t.Errorf("Expected MonitorInterval to be 60s, got %v", serviceConfig.MonitorInterval)
	}
	if serviceConfig.Headless {
		t.Error("Expected Headless to be false")
	}
	if serviceConfig.Network.DefaultBridgeName != "en1" {
		t.Errorf("Expected Network.DefaultBridgeName to be 'en1', got %s", serviceConfig.Network.DefaultBridgeName)
	}
	if serviceConfig.Network.EnableNAT {
		t.Error("Expected Network.EnableNAT to be false")
	}
	if !serviceConfig.Network.EnableBridged {
		t.Error("Expected Network.EnableBridged to be true")
	}
	if !serviceConfig.Network.EnableHostOnly {
		t.Error("Expected Network.EnableHostOnly to be true")
	}
	if serviceConfig.Network.EnableInternal {
		t.Error("Expected Network.EnableInternal to be false")
	}
	if serviceConfig.Storage.DefaultController != "virtioSCSI" {
		t.Errorf("Expected Storage.DefaultController to be 'virtioSCSI', got %s", serviceConfig.Storage.DefaultController)
	}
	if serviceConfig.Storage.DefaultType != "vmdk" {
		t.Errorf("Expected Storage.DefaultType to be 'vmdk', got %s", serviceConfig.Storage.DefaultType)
	}
	if serviceConfig.Storage.EnableTrim {
		t.Error("Expected Storage.EnableTrim to be false")
	}
	if !serviceConfig.Storage.EnableCompression {
		t.Error("Expected Storage.EnableCompression to be true")
	}
	if !serviceConfig.Advanced.EnableAudio {
		t.Error("Expected Advanced.EnableAudio to be true")
	}
	if !serviceConfig.Advanced.EnableUSB {
		t.Error("Expected Advanced.EnableUSB to be true")
	}
	if serviceConfig.Advanced.EnableVRDE {
		t.Error("Expected Advanced.EnableVRDE to be false")
	}
	if serviceConfig.Advanced.VRDEPort != 3390 {
		t.Errorf("Expected Advanced.VRDEPort to be 3390, got %d", serviceConfig.Advanced.VRDEPort)
	}
	if !serviceConfig.Advanced.EnablePAE {
		t.Error("Expected Advanced.EnablePAE to be true")
	}
	if serviceConfig.Advanced.EnableNestedPaging {
		t.Error("Expected Advanced.EnableNestedPaging to be false")
	}
	if serviceConfig.Advanced.EnableHWVirt {
		t.Error("Expected Advanced.EnableHWVirt to be false")
	}
}

func TestNewService(t *testing.T) {
	// This test will only work if VirtualBox is installed
	config := &ServiceConfig{
		Enable:             true,
		DefaultMemoryMB:    2048,
		DefaultCPUs:        2,
		DefaultDiskSizeGB:  20,
		DefaultNetworkType: "nat",
		DefaultOSType:      "Ubuntu_arm64",
		VMStartTimeout:     60 * time.Second,
		VMStopTimeout:      30 * time.Second,
		VMDeleteTimeout:    60 * time.Second,
		MonitorInterval:    30 * time.Second,
		Headless:           true,
	}

	service, err := NewService(config)
	if err != nil {
		t.Skipf("VirtualBox service creation failed (VirtualBox may not be installed): %v", err)
		return
	}

	// Test service start
	ctx := context.Background()
	err = service.Start(ctx)
	if err != nil {
		t.Errorf("Failed to start service: %v", err)
	}

	// Test system info
	info, err := service.GetSystemInfo(ctx)
	if err != nil {
		t.Errorf("Failed to get system info: %v", err)
	} else {
		t.Logf("Host OS: %s", info.HostOS)
		t.Logf("Host Arch: %s", info.HostArch)
	}

	// Test service stop
	err = service.Stop(ctx)
	if err != nil {
		t.Errorf("Failed to stop service: %v", err)
	}
}

func TestCreateVM(t *testing.T) {
	// This test will only work if VirtualBox is installed
	config := &ServiceConfig{
		Enable:             true,
		DefaultMemoryMB:    2048,
		DefaultCPUs:        2,
		DefaultDiskSizeGB:  20,
		DefaultNetworkType: "nat",
		DefaultOSType:      "Ubuntu_arm64",
		VMStartTimeout:     60 * time.Second,
		VMStopTimeout:      30 * time.Second,
		VMDeleteTimeout:    60 * time.Second,
		MonitorInterval:    30 * time.Second,
		Headless:           true,
	}

	service, err := NewService(config)
	if err != nil {
		t.Skipf("VirtualBox service creation failed (VirtualBox may not be installed): %v", err)
		return
	}

	ctx := context.Background()
	err = service.Start(ctx)
	if err != nil {
		t.Skipf("Failed to start service: %v", err)
	}
	defer service.Stop(ctx)

	// Test VM creation
	req := vbtypes.VMCreateRequest{
		Name:       "test-vm",
		CPUCores:   1,
		MemoryMB:   1024,
		DiskSizeGB: 10,
		ISOURL:     "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-arm64.iso",
	}

	vm, err := service.CreateVM(ctx, req)
	if err != nil {
		t.Errorf("Failed to create VM: %v", err)
		return
	}

	t.Logf("Created VM: %s", vm.Name)
	t.Logf("VM ID: %s", vm.ID)
	t.Logf("VM Status: %s", vm.Status)

	// Clean up - delete the VM
	err = service.DeleteVM(ctx, vm.ID)
	if err != nil {
		t.Errorf("Failed to delete VM: %v", err)
	}
}

func TestListVMs(t *testing.T) {
	// This test will only work if VirtualBox is installed
	config := &ServiceConfig{
		Enable:             true,
		DefaultMemoryMB:    2048,
		DefaultCPUs:        2,
		DefaultDiskSizeGB:  20,
		DefaultNetworkType: "nat",
		DefaultOSType:      "Ubuntu_arm64",
		VMStartTimeout:     60 * time.Second,
		VMStopTimeout:      30 * time.Second,
		VMDeleteTimeout:    60 * time.Second,
		MonitorInterval:    30 * time.Second,
		Headless:           true,
	}

	service, err := NewService(config)
	if err != nil {
		t.Skipf("VirtualBox service creation failed (VirtualBox may not be installed): %v", err)
		return
	}

	ctx := context.Background()
	err = service.Start(ctx)
	if err != nil {
		t.Skipf("Failed to start service: %v", err)
	}
	defer service.Stop(ctx)

	// Test listing VMs
	vms, _, err := service.GetVMs(ctx, big.NewInt(0), big.NewInt(10), vbtypes.VMFilter{})
	if err != nil {
		t.Errorf("Failed to list VMs: %v", err)
		return
	}

	t.Logf("Found %d VMs", len(vms))
	for _, vm := range vms {
		t.Logf("VM: %s (Status: %s)", vm.Name, vm.Status)
	}
}

func TestDownloadISO(t *testing.T) {
	// This test will only work if VirtualBox is installed
	config := &ServiceConfig{
		Enable:             true,
		DefaultMemoryMB:    2048,
		DefaultCPUs:        2,
		DefaultDiskSizeGB:  20,
		DefaultNetworkType: "nat",
		DefaultOSType:      "Ubuntu_arm64",
		VMStartTimeout:     60 * time.Second,
		VMStopTimeout:      30 * time.Second,
		VMDeleteTimeout:    60 * time.Second,
		MonitorInterval:    30 * time.Second,
		Headless:           true,
	}

	service, err := NewService(config)
	if err != nil {
		t.Skipf("VirtualBox service creation failed (VirtualBox may not be installed): %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err = service.Start(ctx)
	if err != nil {
		t.Skipf("Failed to start service: %v", err)
	}
	defer service.Stop(ctx)

	// Test ISO download
	isoURL := "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04.2-live-server-arm64.iso"
	iso, err := service.DownloadISO(ctx, isoURL)
	if err != nil {
		t.Errorf("Failed to download ISO: %v", err)
		return
	}

	t.Logf("Downloaded ISO: %s", iso.Path)
	t.Logf("ISO Size: %d bytes", iso.Size)
	t.Logf("ISO Checksum: %s", iso.Checksum)

	// Test listing ISOs
	isos, err := service.ListISOs(ctx)
	if err != nil {
		t.Errorf("Failed to list ISOs: %v", err)
		return
	}

	t.Logf("Found %d ISOs", len(isos))
	for _, iso := range isos {
		t.Logf("ISO: %s", iso.Path)
	}
}

func TestServiceImpl_validateResources(t *testing.T) {
	config := &ServiceConfig{
		DefaultOSType: "Ubuntu_ARM64",
	}
	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tests := []struct {
		name    string
		req     vbtypes.VMCreateRequest
		wantErr bool
	}{
		{
			name: "valid resources",
			req: vbtypes.VMCreateRequest{
				Name:       "test-vm",
				CPUCores:   1,
				MemoryMB:   1024,
				DiskSizeGB: 10,
			},
			wantErr: false,
		},
		{
			name: "excessive CPU cores",
			req: vbtypes.VMCreateRequest{
				Name:       "test-vm",
				CPUCores:   1000, // Unrealistic number
				MemoryMB:   1024,
				DiskSizeGB: 10,
			},
			wantErr: true,
		},
		{
			name: "excessive memory",
			req: vbtypes.VMCreateRequest{
				Name:       "test-vm",
				CPUCores:   1,
				MemoryMB:   1000000, // 1TB - unrealistic
				DiskSizeGB: 10,
			},
			wantErr: true,
		},
		{
			name: "excessive disk space",
			req: vbtypes.VMCreateRequest{
				Name:       "test-vm",
				CPUCores:   1,
				MemoryMB:   1024,
				DiskSizeGB: 1000000, // 1PB - unrealistic
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateResources(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceImpl_GetSystemInfo(t *testing.T) {
	config := &ServiceConfig{
		DefaultOSType: "Ubuntu_ARM64",
	}
	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	info, err := service.GetSystemInfo(context.Background())
	if err != nil {
		t.Fatalf("GetSystemInfo() failed: %v", err)
	}

	// Basic validation
	if info.HostOS == "" {
		t.Error("Expected non-empty HostOS")
	}
	if info.HostArch == "" {
		t.Error("Expected non-empty HostArch")
	}
	if info.AvailableCPUs <= 0 {
		t.Error("Expected positive AvailableCPUs")
	}

	// Note: AvailableRAMMB and AvailableDiskGB might be 0 if resource detection fails
	// This is acceptable as the method has fallback behavior
}
