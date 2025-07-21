package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

func main() {
	// Create a datastore (required for the service)
	ds := sync.MutexWrap(datastore.NewMapDatastore())

	// Create the VirtualBox service
	vboxService, err := virtualbox.NewService(ds)
	if err != nil {
		log.Fatalf("Failed to create VirtualBox service: %v", err)
	}

	// Start the service
	ctx := context.Background()
	if err := vboxService.Start(ctx); err != nil {
		log.Fatalf("Failed to start VirtualBox service: %v", err)
	}
	defer vboxService.Stop(ctx)

	// Step 1: Create a template VM (if it doesn't exist)
	templateName := "template_sample"
	fmt.Printf("Step 1: Creating template VM '%s'...\n", templateName)

	// Check if template already exists
	vms, _, err := vboxService.GetVMs(ctx)
	if err != nil {
		log.Fatalf("Failed to get VMs: %v", err)
	}

	templateExists := false
	for _, vm := range vms {
		if vm.Name == templateName {
			templateExists = true
			fmt.Printf("Template VM '%s' already exists, skipping creation.\n", templateName)
			break
		}
	}

	if !templateExists {
		// Create template VM request
		templateReq := vbtypes.VMCreateRequest{
			Name:       templateName,
			CPUCores:   2,
			MemoryMB:   2048,           // 2GB RAM
			DiskSizeGB: 20,             // 20GB disk
			OSType:     "Ubuntu_ARM64", // or "Ubuntu_ARM64" for ARM
			Username:   "ubuntu",
			Password:   "ubuntu",
		}

		// Create the template VM using the service
		templateVM, err := vboxService.CreateVM(ctx, templateReq)
		if err != nil {
			log.Fatalf("Failed to create template VM: %v", err)
		}

		fmt.Printf("Template VM created successfully!\n")
		fmt.Printf("Template VM ID: %s\n", templateVM.ID)
		fmt.Printf("Template VM Name: %s\n", templateVM.Name)
		fmt.Printf("Status: %s\n", templateVM.Status)
		fmt.Printf("CPU Cores: %d\n", templateVM.CPUCores)
		fmt.Printf("Memory: %d MB\n", templateVM.MemoryMB)
		fmt.Printf("Disk Size: %d GB\n", templateVM.DiskSizeGB)
		fmt.Printf("Template VM Folder: %s\n", templateVM.VMFolder)
		fmt.Printf("Template Username: ubuntu\n")
		fmt.Printf("Template Password: ubuntu\n")
	}

	// Step 2: Create and start a new VM by cloning from template
	fmt.Printf("\nStep 2: Creating and starting new VM by cloning from template...\n")

	cloneReq := vbtypes.VMCreateRequest{
		Name:       "my-cloned-vm",
		CPUCores:   2,
		MemoryMB:   4096, // 4GB RAM (different from template)
		DiskSizeGB: 30,   // 30GB disk (different from template)
		OSType:     "Ubuntu_ARM64",
		Username:   "admin",         // Different username
		Password:   "mypassword123", // Different password
	}

	// Create and start the VM using the new clone functionality
	clonedVM, err := vboxService.CreateAndStartVM(ctx, cloneReq)
	if err != nil {
		log.Fatalf("Failed to create and start cloned VM: %v", err)
	}

	fmt.Printf("Cloned VM created and started successfully!\n")
	fmt.Printf("Cloned VM ID: %s\n", clonedVM.ID)
	fmt.Printf("Cloned VM Name: %s\n", clonedVM.Name)
	fmt.Printf("Status: %s\n", clonedVM.Status)
	fmt.Printf("CPU Cores: %d\n", clonedVM.CPUCores)
	fmt.Printf("Memory: %d MB\n", clonedVM.MemoryMB)
	fmt.Printf("Disk Size: %d GB\n", clonedVM.DiskSizeGB)
	fmt.Printf("Cloned VM Folder: %s\n", clonedVM.VMFolder)
	fmt.Printf("Cloned VM Username: %s\n", cloneReq.Username)
	fmt.Printf("Cloned VM Password: %s\n", cloneReq.Password)

	if clonedVM.SSHPort > 0 {
		fmt.Printf("SSH Port: %d\n", clonedVM.SSHPort)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("- Template VM '%s' created with ubuntu/ubuntu credentials\n", templateName)
	fmt.Printf("- New VM '%s' cloned from template and started with custom credentials\n", clonedVM.Name)
	fmt.Printf("- Clone VM uses new cloud-init templates for clone VMs\n")
	fmt.Printf("- Clone VM is ready for use!\n")
}
