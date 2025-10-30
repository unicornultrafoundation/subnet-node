package vbox_service

import (
	"fmt"
	"testing"
)

func TestVBoxService_Start(t *testing.T) {
	cfg := MockVBoxConfig()
	mockDS := NewMockDatastore()
	service := NewVboxService(cfg, mockDS)

	vmList, err := service.ListVMs()
	if err != nil {
		t.Fatalf("Failed to list VMs: %v", err)
	}
	fmt.Println("VM List:", vmList[0].ID, vmList[0].Name, vmList[0].Status, vmList[0].CPUCores, vmList[0].MemoryMB, vmList[0].DiskSizeGB)

}
