package kvm

import (
	"time"
)

// VMStatus represents the current state of a virtual machine
type VMStatus string

const (
	VMStatusStopped  VMStatus = "stopped"
	VMStatusRunning  VMStatus = "running"
	VMStatusStarting VMStatus = "starting"
	VMStatusStopping VMStatus = "stopping"
	VMStatusError    VMStatus = "error"
)

// VM represents a virtual machine instance
type VM struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Status    VMStatus          `json:"status"`
	CPUCores  int               `json:"cpu_cores"`
	MemoryMB  int               `json:"memory_mb"`
	DiskGB    int               `json:"disk_gb"`
	IPAddress string            `json:"ip_address,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// CreateVMRequest represents a request to create a new VM
type CreateVMRequest struct {
	Name     string            `json:"name" validate:"required"`
	CPUCores int               `json:"cpu_cores" validate:"required,min=1,max=8"`
	MemoryMB int               `json:"memory_mb" validate:"required,min=512,max=8192"`
	DiskGB   int               `json:"disk_gb" validate:"required,min=10,max=100"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// VMStats represents VM resource usage statistics
type VMStats struct {
	VMID        string    `json:"vm_id"`
	CPUUsage    float64   `json:"cpu_usage"`     // percentage
	MemoryUsage float64   `json:"memory_usage"`  // percentage
	DiskUsage   float64   `json:"disk_usage"`    // percentage
	NetworkRxMB float64   `json:"network_rx_mb"` // received MB
	NetworkTxMB float64   `json:"network_tx_mb"` // transmitted MB
	CollectedAt time.Time `json:"collected_at"`
}

// SystemResources represents available system resources
type SystemResources struct {
	TotalCPUCores     int `json:"total_cpu_cores"`
	AvailableCPUCores int `json:"available_cpu_cores"`
	TotalMemoryMB     int `json:"total_memory_mb"`
	AvailableMemoryMB int `json:"available_memory_mb"`
	TotalDiskGB       int `json:"total_disk_gb"`
	AvailableDiskGB   int `json:"available_disk_gb"`
	RunningVMs        int `json:"running_vms"`
	MaxVMs            int `json:"max_vms"`
}
