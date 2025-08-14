package types

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
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Status     VMStatus `json:"status"`
	CPUCores   int      `json:"cpu_cores"`
	MemoryMB   int      `json:"memory_mb"`
	DiskSizeGB int      `json:"disk_size_gb"`
	IPAddress  string   `json:"ip_address,omitempty"`
	SSHPort    int      `json:"ssh_port,omitempty"`
	VMFolder   string   `json:"vm_folder"`
}

// VMCreateRequest represents a request to create a new VM
type VMCreateRequest struct {
	Name       string `json:"name"`
	CPUCores   int    `json:"cpu_cores"`
	MemoryMB   int    `json:"memory_mb"`
	DiskSizeGB int    `json:"disk_size_gb"`
	OSType     string `json:"os_type,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
}

// VMUpdateRequest represents a request to update an existing VM
type VMUpdateRequest struct {
	Name       string `json:"name,omitempty"`
	CPUCores   int    `json:"cpu_cores,omitempty"`
	MemoryMB   int    `json:"memory_mb,omitempty"`
	DiskSizeGB int    `json:"disk_size_gb,omitempty"`
}

type GenerateSSHTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

// SSHAccessToken represents a one-time access token for SSH connections
type SSHAccessToken struct {
	Token     string    `json:"token"`
	VMID      string    `json:"vm_id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

// SSHTokenRequest represents a request to generate an SSH access token
type SSHTokenRequest struct {
	VMID     string `json:"vm_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// SSHTokenResponse represents the response for SSH token generation
type SSHTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// VMEventType represents the type of VM event
type VMEventType string

const (
	VMEventCreateVM VMEventType = "create_vm"
)

// VMRequest represents a VM operation request sent through the channel
type VMRequest struct {
	Type      VMEventType            `json:"type"`
	VMID      string                 `json:"vm_id,omitempty"`
	VMName    string                 `json:"vm_name"`
	VMStatus  VMStatus               `json:"vm_status"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type VMEvent struct {
	EventType VMEventType            `json:"event_type"`
	VMID      string                 `json:"vm_id,omitempty"`
	VMName    string                 `json:"vm_name"`
	VMStatus  VMStatus               `json:"vm_status"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// JobStatus represents the status of a background job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// Job represents a background job
type Job struct {
	ID          string                 `json:"id"`
	EventType   VMEventType            `json:"event_type"`
	Status      JobStatus              `json:"status"`
	Request     map[string]interface{} `json:"request"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	VMID        string                 `json:"vm_id,omitempty"`
	VMName      string                 `json:"vm_name,omitempty"`
}

// JobCreateResponse represents the response when creating a job
type JobCreateResponse struct {
	JobID string `json:"job_id"`
}

type CreateVMFromImageRequest struct {
	OrderId  string `json:"order_id"`
	OS       string `json:"os"`
	Version  string `json:"version"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type VMCreateFromImageRequest struct {
	Name       string `json:"name"`
	CPUCores   int    `json:"cpu_cores"`
	OS         string `json:"os"`
	Version    string `json:"version"`
	MemoryMB   int    `json:"memory_mb"`
	DiskSizeGB int    `json:"disk_size_gb"`
	OSType     string `json:"os_type,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
}
