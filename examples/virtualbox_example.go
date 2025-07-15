package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/virtualbox"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

func main() {
	fmt.Println("VirtualBox Service Example with Hardware Detection")
	fmt.Println("==================================================")

	// Demonstrate hardware detection
	fmt.Println("\n1. Hardware Detection Demo:")
	detector := virtualbox.NewHardwareDetector()
	hardware, err := detector.DetectHardware()
	if err != nil {
		log.Fatalf("Failed to detect hardware: %v", err)
	}

	fmt.Printf("   Architecture: %s\n", hardware.Architecture)
	fmt.Printf("   CPU Count: %d\n", hardware.CPUCount)
	fmt.Printf("   Total Memory: %d MB\n", hardware.TotalMemoryMB)
	fmt.Printf("   GPU Type: %s (Cores: %d)\n", hardware.GPUType, hardware.GPUCount)
	fmt.Printf("   Platform: %s\n", hardware.OSType)

	// Get VirtualBox settings based on hardware
	settings := detector.GetVirtualBoxSettings(hardware)
	fmt.Printf("   Recommended OS Type: %s\n", settings.OSType)
	fmt.Printf("   Recommended Chipset: %s\n", settings.Chipset)
	fmt.Printf("   Recommended VRAM: %d MB\n", settings.VRAMMB)

	// Create VirtualBox service with hardware-aware configuration
	fmt.Println("\n2. Creating VirtualBox service with hardware detection...")
	config := &virtualbox.ServiceConfig{
		Enable:             true,
		DefaultMemoryMB:    2048,
		DefaultCPUs:        2,
		DefaultDiskSizeGB:  20,
		DefaultNetworkType: "nat",
		DefaultOSType:      settings.OSType, // Use detected OS type
		VMStartTimeout:     60 * time.Second,
		VMStopTimeout:      30 * time.Second,
		VMDeleteTimeout:    60 * time.Second,
		MonitorInterval:    30 * time.Second,
		Headless:           true,
		Network: virtualbox.NetworkConfig{
			DefaultBridgeName: "en0",
			EnableNAT:         true,
			EnableBridged:     true,
			EnableHostOnly:    false,
			EnableInternal:    false,
		},
		Storage: virtualbox.StorageConfig{
			DefaultController: "SATA",
			DefaultType:       "vdi",
			EnableTrim:        true,
			EnableCompression: false,
		},
		Advanced: virtualbox.AdvancedConfig{
			EnableAudio:        false,
			EnableUSB:          false,
			EnableVRDE:         true,
			VRDEPort:           3389,
			EnablePAE:          false,
			EnableNestedPaging: true,
			EnableHWVirt:       true,
		},
	}

	service, err := virtualbox.NewService(config)
	if err != nil {
		log.Fatalf("Failed to create VirtualBox service: %v", err)
	}

	// Start the service
	ctx := context.Background()
	fmt.Println("Starting VirtualBox service...")
	err = service.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}
	defer func() {
		fmt.Println("Stopping VirtualBox service...")
		service.Stop(ctx)
	}()

	// Get system information
	fmt.Println("\nGetting system information...")
	info, err := service.GetSystemInfo(ctx)
	if err != nil {
		log.Printf("Failed to get system info: %v", err)
	} else {
		fmt.Printf("Host OS: %s\n", info.HostOS)
		fmt.Printf("Host Architecture: %s\n", info.HostArch)
	}

	// List existing VMs
	fmt.Println("\nListing existing VMs...")
	vms, _, err := service.GetVMs(ctx, big.NewInt(0), big.NewInt(10), vbtypes.VMFilter{})
	if err != nil {
		log.Printf("Failed to list VMs: %v", err)
	} else {
		fmt.Printf("Found %d VMs:\n", len(vms))
		for _, vm := range vms {
			fmt.Printf("  - %s (Status: %s)\n", vm.Name, vm.Status)
		}
	}

	// Create a new VM
	vmName := fmt.Sprintf("example-vm-%d", time.Now().Unix())
	fmt.Printf("\nCreating VM: %s\n", vmName)

	req := vbtypes.VMCreateRequest{
		Name:       vmName,
		CPUCores:   1,
		MemoryMB:   2048,
		DiskSizeGB: 10,
		// Note: ISO URL is handled internally by the service based on OS type
	}

	vm, err := service.CreateVM(ctx, req)
	if err != nil {
		log.Fatalf("Failed to create VM: %v", err)
	}

	fmt.Printf("Successfully created VM:\n")
	fmt.Printf("  ID: %s\n", vm.ID)
	fmt.Printf("  Name: %s\n", vm.Name)
	fmt.Printf("  Status: %s\n", vm.Status)
	fmt.Printf("  CPU Cores: %d\n", vm.CPUCores)
	fmt.Printf("  Memory: %d MB\n", vm.MemoryMB)
	fmt.Printf("  Disk Size: %d GB\n", vm.DiskSizeGB)

	// List ISOs
	fmt.Println("\nListing available ISOs...")
	isos, err := service.ListISOs(ctx)
	if err != nil {
		log.Printf("Failed to list ISOs: %v", err)
	} else {
		fmt.Printf("Found %d ISOs:\n", len(isos))
		for _, iso := range isos {
			fmt.Printf("  - %s (%d bytes)\n", iso.Path, iso.Size)
		}
	}

	// Get VM usage (placeholder)
	fmt.Println("\nGetting VM usage...")
	usage, err := service.GetVMUsage(ctx, vm.ID)
	if err != nil {
		log.Printf("Failed to get VM usage: %v", err)
	} else {
		fmt.Printf("VM Usage:\n")
		fmt.Printf("  CPU: %.2f%%\n", usage.CPUPerc)
		fmt.Printf("  Memory: %d MB\n", usage.MemoryMB)
		fmt.Printf("  Disk: %.2f GB\n", usage.DiskUsageGB)
	}

	// Clean up - delete the VM
	fmt.Printf("\nCleaning up - deleting VM: %s\n", vmName)
	err = service.DeleteVM(ctx, vm.ID)
	if err != nil {
		log.Printf("Failed to delete VM: %v", err)
	} else {
		fmt.Println("VM deleted successfully")
	}

	fmt.Println("\nExample completed successfully!")
}
