package virtualbox

import (
	"time"
)

// VMStatus represents the current status of a Virtual Machine
type VMStatus string

const (
	// Running indicates the VM is currently executing
	Running VMStatus = "running"
	// Stopped indicates that the VM has been stopped
	Stopped VMStatus = "stopped"
	// Paused indicates that the VM is currently paused
	Paused VMStatus = "paused"
	// Starting indicates that the VM is in the process of starting
	Starting VMStatus = "starting"
	// Stopping indicates that the VM is in the process of stopping
	Stopping VMStatus = "stopping"
	// Pausing indicates that the VM is in the process of pausing
	Pausing VMStatus = "pausing"
	// Unknown indicates that we could not determine the status
	Unknown VMStatus = "unknown"
	// NotFound indicates that the VM was not found
	NotFound VMStatus = "notfound"
)

// VM represents a Virtual Machine configuration and state
type VM struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Status     VMStatus  `json:"status"`
	CPUCores   int       `json:"cpu_cores"`
	MemoryMB   int       `json:"memory_mb"`
	DiskSizeGB int       `json:"disk_size_gb"`
	ISOURL     string    `json:"iso_url"`
	ISOPath    string    `json:"iso_path"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IPAddress  string    `json:"ip_address,omitempty"`
	SSHPort    int       `json:"ssh_port,omitempty"`
	VBoxPath   string    `json:"vbox_path"`
	VMFolder   string    `json:"vm_folder"`
}

// VMCreateRequest represents a request to create a new VM
type VMCreateRequest struct {
	Name       string `json:"name"`
	CPUCores   int    `json:"cpu_cores"`
	MemoryMB   int    `json:"memory_mb"`
	DiskSizeGB int    `json:"disk_size_gb"`
	OSType     string `json:"os_type,omitempty"`
}

// VMUpdateRequest represents a request to update an existing VM
type VMUpdateRequest struct {
	Name       string `json:"name,omitempty"`
	CPUCores   int    `json:"cpu_cores,omitempty"`
	MemoryMB   int    `json:"memory_mb,omitempty"`
	DiskSizeGB int    `json:"disk_size_gb,omitempty"`
}

// VMFilter represents filters for querying VMs
type VMFilter struct {
	Status string `json:"status,omitempty"`
	Query  string `json:"query,omitempty"`
}

// VMUsage represents resource usage of a VM
type VMUsage struct {
	VMID          string    `json:"vm_id"`
	CPUPerc       float64   `json:"cpu_perc"`
	MemoryMB      int       `json:"memory_mb"`
	DiskUsageGB   float64   `json:"disk_usage_gb"`
	NetworkRxMB   float64   `json:"network_rx_mb"`
	NetworkTxMB   float64   `json:"network_tx_mb"`
	UptimeSeconds int64     `json:"uptime_seconds"`
	Timestamp     time.Time `json:"timestamp"`
}

// VMSystemInfo represents system information about the VirtualBox installation
type VMSystemInfo struct {
	HostOS          string `json:"host_os"`
	HostArch        string `json:"host_arch"`
	AvailableCPUs   int    `json:"available_cpus"`
	AvailableRAMMB  int    `json:"available_ram_mb"`
	AvailableDiskGB int    `json:"available_disk_gb"`
}

// ISOInfo represents information about an ISO file
type ISOInfo struct {
	URL          string    `json:"url"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	Checksum     string    `json:"checksum"`
	DownloadedAt time.Time `json:"downloaded_at"`
}
