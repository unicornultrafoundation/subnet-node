package virtualbox

import (
	"context"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
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

	// Test Network configuration
	if serviceConfig.Network.DefaultBridgeName != "en1" {
		t.Errorf("Expected DefaultBridgeName to be 'en1', got %s", serviceConfig.Network.DefaultBridgeName)
	}
	if serviceConfig.Network.EnableNAT {
		t.Error("Expected EnableNAT to be false")
	}
	if !serviceConfig.Network.EnableBridged {
		t.Error("Expected EnableBridged to be true")
	}
	if !serviceConfig.Network.EnableHostOnly {
		t.Error("Expected EnableHostOnly to be true")
	}
	if serviceConfig.Network.EnableInternal {
		t.Error("Expected EnableInternal to be false")
	}

	// Test Storage configuration
	if serviceConfig.Storage.DefaultController != "virtioSCSI" {
		t.Errorf("Expected DefaultController to be 'virtioSCSI', got %s", serviceConfig.Storage.DefaultController)
	}
	if serviceConfig.Storage.DefaultType != "vmdk" {
		t.Errorf("Expected DefaultType to be 'vmdk', got %s", serviceConfig.Storage.DefaultType)
	}
	if serviceConfig.Storage.EnableTrim {
		t.Error("Expected EnableTrim to be false")
	}
	if !serviceConfig.Storage.EnableCompression {
		t.Error("Expected EnableCompression to be true")
	}

	// Test Advanced configuration
	if !serviceConfig.Advanced.EnableAudio {
		t.Error("Expected EnableAudio to be true")
	}
	if !serviceConfig.Advanced.EnableUSB {
		t.Error("Expected EnableUSB to be true")
	}
	if serviceConfig.Advanced.EnableVRDE {
		t.Error("Expected EnableVRDE to be false")
	}
	if serviceConfig.Advanced.VRDEPort != 3390 {
		t.Errorf("Expected VRDEPort to be 3390, got %d", serviceConfig.Advanced.VRDEPort)
	}
	if !serviceConfig.Advanced.EnablePAE {
		t.Error("Expected EnablePAE to be true")
	}
	if serviceConfig.Advanced.EnableNestedPaging {
		t.Error("Expected EnableNestedPaging to be false")
	}
	if serviceConfig.Advanced.EnableHWVirt {
		t.Error("Expected EnableHWVirt to be false")
	}
}

// TestGenerateCloudInitISO tests the generateCloudInitISO function specifically
func TestGenerateCloudInitISO(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "virtualbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after test

	// Create a minimal service configuration
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

	// Create a storage manager
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Create a service instance with the temp directory
	service := &ServiceImpl{
		storageMgr: storageMgr,
		vmDir:      tempDir,
		stopChan:   make(chan struct{}),
		config:     config,
		vboxExec:   NewVBoxManageExecutor(tempDir),
	}

	// Test the generateCloudInitISO function
	vmName := "test-vm"

	// Create the VM directory structure first (this is normally done in CreateVM)
	vmDir := filepath.Join(tempDir, vmName)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatalf("Failed to create VM directory: %v", err)
	}

	cloudInitISO, err := service.generateCloudInitISO(vmName)

	if err != nil {
		t.Fatalf("generateCloudInitISO failed: %v", err)
	}

	// Verify the ISO file was created
	if cloudInitISO == "" {
		t.Error("Expected cloud-init ISO path to be non-empty")
	}

	// Check if the ISO file actually exists
	if _, err := os.Stat(cloudInitISO); os.IsNotExist(err) {
		t.Errorf("Cloud-init ISO file does not exist: %s", cloudInitISO)
	}

	// Verify the expected directory structure was created
	expectedCloudInitDir := filepath.Join(tempDir, vmName, "cloud-init")
	if _, err := os.Stat(expectedCloudInitDir); os.IsNotExist(err) {
		t.Errorf("Cloud-init directory does not exist: %s", expectedCloudInitDir)
	}

	// Check for meta-data file
	metaDataPath := filepath.Join(expectedCloudInitDir, "meta-data")
	if _, err := os.Stat(metaDataPath); os.IsNotExist(err) {
		t.Errorf("Meta-data file does not exist: %s", metaDataPath)
	}

	// Check for user-data file
	userDataPath := filepath.Join(expectedCloudInitDir, "user-data")
	if _, err := os.Stat(userDataPath); os.IsNotExist(err) {
		t.Errorf("User-data file does not exist: %s", userDataPath)
	}

	// Check for the ISO file
	expectedISOPath := filepath.Join(expectedCloudInitDir, "cloud-init.iso")
	if _, err := os.Stat(expectedISOPath); os.IsNotExist(err) {
		t.Errorf("Cloud-init ISO file does not exist: %s", expectedISOPath)
	}

	t.Logf("Successfully generated cloud-init ISO: %s", cloudInitISO)
	t.Logf("Cloud-init directory: %s", expectedCloudInitDir)
	t.Logf("Meta-data file: %s", metaDataPath)
	t.Logf("User-data file: %s", userDataPath)
	t.Logf("ISO file: %s", expectedISOPath)
}

// TestGenerateCloudInitISOReal tests the generateCloudInitISO function with real VirtualBox directory
func TestGenerateCloudInitISOReal(t *testing.T) {
	// Skip this test if you don't want to use the real VirtualBox directory
	// Uncomment the next line to run this test
	// t.Skip("Skipping test with real VirtualBox directory")

	// Create a minimal service configuration
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

	// Create a storage manager
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Use the real VirtualBox directory (same as in NewService)
	realVMDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	realVMDir = filepath.Join(realVMDir, "VirtualBox VMs")

	// Create a service instance with the real VirtualBox directory
	service := &ServiceImpl{
		storageMgr: storageMgr,
		vmDir:      realVMDir,
		stopChan:   make(chan struct{}),
		config:     config,
		vboxExec:   NewVBoxManageExecutor(realVMDir),
	}

	// Test the generateCloudInitISO function
	vmName := "test-vm-real"

	// Create the VM directory structure first
	vmDir := filepath.Join(realVMDir, vmName)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatalf("Failed to create VM directory: %v", err)
	}

	// Clean up after test
	defer func() {
		os.RemoveAll(vmDir)
		t.Logf("Cleaned up test VM directory: %s", vmDir)
	}()

	cloudInitISO, err := service.generateCloudInitISO(vmName)

	if err != nil {
		t.Fatalf("generateCloudInitISO failed: %v", err)
	}

	// Verify the ISO file was created
	if cloudInitISO == "" {
		t.Error("Expected cloud-init ISO path to be non-empty")
	}

	// Check if the ISO file actually exists
	if _, err := os.Stat(cloudInitISO); os.IsNotExist(err) {
		t.Errorf("Cloud-init ISO file does not exist: %s", cloudInitISO)
	}

	// Verify the expected directory structure was created
	expectedCloudInitDir := filepath.Join(realVMDir, vmName, "cloud-init")
	if _, err := os.Stat(expectedCloudInitDir); os.IsNotExist(err) {
		t.Errorf("Cloud-init directory does not exist: %s", expectedCloudInitDir)
	}

	t.Logf("Successfully generated cloud-init ISO in real VirtualBox directory: %s", cloudInitISO)
	t.Logf("Cloud-init directory: %s", expectedCloudInitDir)
	t.Logf("Real VirtualBox directory: %s", realVMDir)
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

// TestCloudInitPerVM tests that each VM gets its own separate cloud-init data
func TestCloudInitPerVM(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "virtualbox-cloudinit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after test

	// Create a minimal service configuration
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

	// Create a storage manager
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Create a service instance with the temp directory
	service := &ServiceImpl{
		storageMgr: storageMgr,
		vmDir:      tempDir,
		stopChan:   make(chan struct{}),
		config:     config,
		vboxExec:   NewVBoxManageExecutor(tempDir),
	}

	// Test creating cloud-init for multiple VMs
	vmNames := []string{"test-vm-1", "test-vm-2", "test-vm-3"}
	cloudInitPaths := make(map[string]string)

	for _, vmName := range vmNames {
		// Create the VM directory structure
		vmDir := filepath.Join(tempDir, vmName)
		if err := os.MkdirAll(vmDir, 0755); err != nil {
			t.Fatalf("Failed to create VM directory for %s: %v", vmName, err)
		}

		// Generate cloud-init ISO for this VM
		cloudInitISO, err := service.generateCloudInitISO(vmName)
		if err != nil {
			t.Fatalf("Failed to generate cloud-init ISO for %s: %v", vmName, err)
		}

		cloudInitPaths[vmName] = cloudInitISO

		// Verify the cloud-init directory structure
		expectedCloudInitDir := filepath.Join(tempDir, vmName, "cloud-init")
		if _, err := os.Stat(expectedCloudInitDir); os.IsNotExist(err) {
			t.Errorf("Cloud-init directory does not exist for %s: %s", vmName, expectedCloudInitDir)
		}

		// Check for required files
		requiredFiles := []string{"meta-data", "user-data", "cloud-init.iso"}
		for _, file := range requiredFiles {
			filePath := filepath.Join(expectedCloudInitDir, file)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Required file %s does not exist for %s: %s", file, vmName, filePath)
			}
		}

		// Verify that the user-data file contains VM-specific information
		userDataPath := filepath.Join(expectedCloudInitDir, "user-data")
		userData, err := ioutil.ReadFile(userDataPath)
		if err != nil {
			t.Errorf("Failed to read user-data for %s: %v", vmName, err)
		} else {
			userDataStr := string(userData)
			if !strings.Contains(userDataStr, vmName) {
				t.Errorf("User-data for %s does not contain VM name: %s", vmName, vmName)
			}
			if !strings.Contains(userDataStr, "hostname:") {
				t.Errorf("User-data for %s does not contain hostname configuration", vmName)
			}
		}

		// Verify that the meta-data file contains VM-specific information
		metaDataPath := filepath.Join(expectedCloudInitDir, "meta-data")
		metaData, err := ioutil.ReadFile(metaDataPath)
		if err != nil {
			t.Errorf("Failed to read meta-data for %s: %v", vmName, err)
		} else {
			metaDataStr := string(metaData)
			if !strings.Contains(metaDataStr, vmName) {
				t.Errorf("Meta-data for %s does not contain VM name: %s", vmName, vmName)
			}
		}

		t.Logf("Successfully generated cloud-init for VM %s: %s", vmName, cloudInitISO)
	}

	// Verify that each VM has a different cloud-init ISO path
	isoPaths := make(map[string]bool)
	for vmName, isoPath := range cloudInitPaths {
		if isoPaths[isoPath] {
			t.Errorf("Duplicate cloud-init ISO path found for VM %s: %s", vmName, isoPath)
		}
		isoPaths[isoPath] = true
	}

	// Verify that each VM has its own separate directory
	for _, vmName := range vmNames {
		vmDir := filepath.Join(tempDir, vmName)
		otherVMDirs := make([]string, 0)
		for _, otherVMName := range vmNames {
			if otherVMName != vmName {
				otherVMDirs = append(otherVMDirs, filepath.Join(tempDir, otherVMName))
			}
		}

		// Check that this VM's directory is separate from others
		for _, otherVMDir := range otherVMDirs {
			if vmDir == otherVMDir {
				t.Errorf("VM directories are not separate: %s and %s", vmDir, otherVMDir)
			}
		}
	}

	t.Logf("Successfully verified that each VM has its own separate cloud-init data")
}

// TestGenerateCloudInitFilesSample demonstrates the GenerateCloudInitFiles function
func TestGenerateCloudInitFilesSample(t *testing.T) {
	// Use the real VirtualBox VMs directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	vmDir := filepath.Join(homeDir, "VirtualBox VMs")

	// Create a minimal service configuration
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

	// Create a storage manager
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Create a service instance with the real VM directory
	service := &ServiceImpl{
		storageMgr: storageMgr,
		vmDir:      vmDir,
		stopChan:   make(chan struct{}),
		config:     config,
		vboxExec:   NewVBoxManageExecutor(vmDir),
	}

	// Test data for different VMs
	testVMs := []struct {
		name     string
		hostname string
		username string
		password string
	}{
		{
			name:     "test-vm-1",
			hostname: "test-vm-1",
			username: "ubuntu",
			password: "vm-test-vm-1",
		},
		{
			name:     "test-vm-2",
			hostname: "test-vm-2",
			username: "ubuntu",
			password: "vm-test-vm-2",
		},
		{
			name:     "production-server",
			hostname: "prod-server",
			username: "ubuntu",
			password: "vm-production-server",
		},
	}

	var createdVMDirs []string

	for _, vm := range testVMs {
		t.Logf("\n=== Testing VM: %s ===", vm.name)

		// Create the VM directory structure
		vmDirPath := filepath.Join(vmDir, vm.name)
		if err := os.MkdirAll(vmDirPath, 0755); err != nil {
			t.Fatalf("Failed to create VM directory for %s: %v", vm.name, err)
		}
		createdVMDirs = append(createdVMDirs, vmDirPath)

		// Call GenerateCloudInitFiles function
		metaDataPath, userDataPath, cloudInitDir, err := service.vboxExec.GenerateCloudInitFiles(
			vm.name, vm.hostname, vm.username, vm.password)
		if err != nil {
			t.Fatalf("GenerateCloudInitFiles failed for %s: %v", vm.name, err)
		}

		t.Logf("Generated cloud-init directory: %s", cloudInitDir)
		t.Logf("Meta-data file: %s", metaDataPath)
		t.Logf("User-data file: %s", userDataPath)

		// Read and display meta-data content
		metaData, err := ioutil.ReadFile(metaDataPath)
		if err != nil {
			t.Errorf("Failed to read meta-data for %s: %v", vm.name, err)
		} else {
			t.Logf("Meta-data content for %s:\n%s", vm.name, string(metaData))
		}

		// Read and display user-data content
		userData, err := ioutil.ReadFile(userDataPath)
		if err != nil {
			t.Errorf("Failed to read user-data for %s: %v", vm.name, err)
		} else {
			t.Logf("User-data content for %s:\n%s", vm.name, string(userData))
		}

		// Verify the files contain VM-specific information
		metaDataStr := string(metaData)
		userDataStr := string(userData)

		// Check meta-data
		if !strings.Contains(metaDataStr, vm.name) {
			t.Errorf("Meta-data for %s does not contain VM name: %s", vm.name, vm.name)
		}
		if !strings.Contains(metaDataStr, vm.hostname) {
			t.Errorf("Meta-data for %s does not contain hostname: %s", vm.name, vm.hostname)
		}

		// Check user-data
		if !strings.Contains(userDataStr, vm.hostname) {
			t.Errorf("User-data for %s does not contain hostname: %s", vm.name, vm.hostname)
		}
		if !strings.Contains(userDataStr, vm.username) {
			t.Errorf("User-data for %s does not contain username: %s", vm.name, vm.username)
		}
		if !strings.Contains(userDataStr, vm.password) {
			t.Errorf("User-data for %s does not contain password: %s", vm.name, vm.password)
		}

		// Verify directory structure
		expectedCloudInitDir := filepath.Join(vmDir, vm.name, "cloud-init")
		if cloudInitDir != expectedCloudInitDir {
			t.Errorf("Expected cloud-init directory %s, got %s", expectedCloudInitDir, cloudInitDir)
		}

		// Check that files exist
		if _, err := os.Stat(metaDataPath); os.IsNotExist(err) {
			t.Errorf("Meta-data file does not exist: %s", metaDataPath)
		}
		if _, err := os.Stat(userDataPath); os.IsNotExist(err) {
			t.Errorf("User-data file does not exist: %s", userDataPath)
		}

		t.Logf("✓ Successfully generated cloud-init files for VM: %s", vm.name)
	}

	// Cleanup: remove created test VM directories
	for _, dir := range createdVMDirs {
		os.RemoveAll(dir)
	}

	t.Logf("\n=== Summary ===")
	t.Logf("Generated cloud-init files for %d VMs in: %s", len(testVMs), vmDir)
	t.Logf("Each VM has its own separate cloud-init directory and files")
}

// TestSimpleCloudInit generates cloud-init files for a single VM with specific parameters
func TestSimpleCloudInit(t *testing.T) {
	// Use the real VirtualBox VMs directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	vmDir := filepath.Join(homeDir, "VirtualBox VMs")

	// Create a minimal service configuration
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

	// Create a storage manager
	storageMgr, err := NewStorageManagerWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Create a service instance with the real VM directory
	service := &ServiceImpl{
		storageMgr: storageMgr,
		vmDir:      vmDir,
		stopChan:   make(chan struct{}),
		config:     config,
		vboxExec:   NewVBoxManageExecutor(vmDir),
	}

	// Test parameters
	vmName := "name_1"
	hostname := "name_1"
	username := "ubuntu"
	password := "1234"

	t.Logf("=== Testing Simple Cloud-Init Generation ===")
	t.Logf("VM Name: %s", vmName)
	t.Logf("Hostname: %s", hostname)
	t.Logf("Username: %s", username)
	t.Logf("Password: %s", password)
	t.Logf("VM Directory: %s", vmDir)

	// Create the VM directory structure
	vmDirPath := filepath.Join(vmDir, vmName)
	if err := os.MkdirAll(vmDirPath, 0755); err != nil {
		t.Fatalf("Failed to create VM directory: %v", err)
	}

	// Call GenerateCloudInitFiles function
	metaDataPath, userDataPath, cloudInitDir, err := service.vboxExec.GenerateCloudInitFiles(
		vmName, hostname, username, password)
	if err != nil {
		t.Fatalf("GenerateCloudInitFiles failed: %v", err)
	}

	t.Logf("\n=== Generated Files ===")
	t.Logf("Cloud-init directory: %s", cloudInitDir)
	t.Logf("Meta-data file: %s", metaDataPath)
	t.Logf("User-data file: %s", userDataPath)

	// Read and display meta-data content
	metaData, err := ioutil.ReadFile(metaDataPath)
	if err != nil {
		t.Fatalf("Failed to read meta-data: %v", err)
	}

	t.Logf("\n=== Meta-data Content ===")
	t.Logf("%s", string(metaData))

	// Read and display user-data content
	userData, err := ioutil.ReadFile(userDataPath)
	if err != nil {
		t.Fatalf("Failed to read user-data: %v", err)
	}

	t.Logf("\n=== User-data Content ===")
	t.Logf("%s", string(userData))

	// Verify the files contain the expected information
	metaDataStr := string(metaData)
	userDataStr := string(userData)

	// Check meta-data
	if !strings.Contains(metaDataStr, vmName) {
		t.Errorf("Meta-data does not contain VM name: %s", vmName)
	}
	if !strings.Contains(metaDataStr, hostname) {
		t.Errorf("Meta-data does not contain hostname: %s", hostname)
	}

	// Check user-data
	if !strings.Contains(userDataStr, hostname) {
		t.Errorf("User-data does not contain hostname: %s", hostname)
	}
	if !strings.Contains(userDataStr, username) {
		t.Errorf("User-data does not contain username: %s", username)
	}
	if !strings.Contains(userDataStr, password) {
		t.Errorf("User-data does not contain password: %s", password)
	}

	// Verify directory structure
	expectedCloudInitDir := filepath.Join(vmDir, vmName, "cloud-init")
	if cloudInitDir != expectedCloudInitDir {
		t.Errorf("Expected cloud-init directory %s, got %s", expectedCloudInitDir, cloudInitDir)
	}

	// Check that files exist
	if _, err := os.Stat(metaDataPath); os.IsNotExist(err) {
		t.Errorf("Meta-data file does not exist: %s", metaDataPath)
	}
	if _, err := os.Stat(userDataPath); os.IsNotExist(err) {
		t.Errorf("User-data file does not exist: %s", userDataPath)
	}

	t.Logf("\n=== Success ===")
	t.Logf("✓ Successfully generated cloud-init files for VM: %s", vmName)
	t.Logf("✓ Files are located in: %s", cloudInitDir)
	t.Logf("✓ Template variables have been replaced with actual values")

	// Note: Files are left in the VM directory for inspection
	// To clean up, uncomment the next line:
	// os.RemoveAll(vmDirPath)
}
