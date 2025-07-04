package types

import (
	"time"
)

// VMState represents the state of a VirtualBox VM
type VMState string

const (
	VMStatePowerOff VMState = "poweroff"
	VMStateRunning  VMState = "running"
	VMStatePaused   VMState = "paused"
	VMStateSaved    VMState = "saved"
	VMStateAborted  VMState = "aborted"
)

// VMRequest represents a request to create a VM
type VMRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Requester   string            `json:"requester"`
	TTL         time.Duration     `json:"ttl"`
	CreatedAt   time.Time         `json:"created_at"`
	Config      *VMConfig         `json:"config"`
	Metadata    map[string]string `json:"metadata"`
}

// VMConfig represents the configuration for a VM
type VMConfig struct {
	// Basic VM settings
	MemoryMB     int    `json:"memory_mb"`
	CPUs         int    `json:"cpus"`
	DiskSizeGB   int    `json:"disk_size_gb"`
	OSType       string `json:"os_type"`
	BaseImage    string `json:"base_image"`
	SnapshotName string `json:"snapshot_name"`

	// Network configuration
	NetworkType string `json:"network_type"`
	BridgeName  string `json:"bridge_name"`
	MACAddress  string `json:"mac_address"`

	// Storage configuration
	StorageController string `json:"storage_controller"`
	StorageType       string `json:"storage_type"`

	// Advanced settings
	EnableAudio        bool `json:"enable_audio"`
	EnableUSB          bool `json:"enable_usb"`
	EnableVRDE         bool `json:"enable_vrde"`
	VRDEPort           int  `json:"vrde_port"`
	EnablePAE          bool `json:"enable_pae"`
	EnableNestedPaging bool `json:"enable_nested_paging"`

	// Custom settings
	CustomSettings map[string]string `json:"custom_settings"`
}

// VMResponse represents the response after VM creation
type VMResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	State     VMState           `json:"state"`
	UUID      string            `json:"uuid"`
	Requester string            `json:"requester"`
	CreatedAt time.Time         `json:"created_at"`
	ExpiresAt time.Time         `json:"expires_at"`
	Config    *VMConfig         `json:"config"`
	Status    *VMStatus         `json:"status"`
	Metadata  map[string]string `json:"metadata"`
}

// VMStatus represents the current status of a VM
type VMStatus struct {
	State       VMState       `json:"state"`
	LastStarted *time.Time    `json:"last_started"`
	Uptime      time.Duration `json:"uptime"`
	CPUUsage    float64       `json:"cpu_usage"`
	MemoryUsage float64       `json:"memory_usage"`
	DiskUsage   float64       `json:"disk_usage"`
	NetworkIn   uint64        `json:"network_in"`
	NetworkOut  uint64        `json:"network_out"`
}

// VMInfo represents detailed information about a VM
type VMInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	UUID        string            `json:"uuid"`
	State       VMState           `json:"state"`
	OSType      string            `json:"os_type"`
	MemoryMB    int               `json:"memory_mb"`
	CPUs        int               `json:"cpus"`
	CreatedAt   time.Time         `json:"created_at"`
	LastStarted *time.Time        `json:"last_started"`
	Uptime      time.Duration     `json:"uptime"`
	Config      *VMConfig         `json:"config"`
	Status      *VMStatus         `json:"status"`
	Metadata    map[string]string `json:"metadata"`
}

// NetworkInterface represents a network interface configuration
type NetworkInterface struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	BridgeName string `json:"bridge_name"`
	MACAddress string `json:"mac_address"`
	Connected  bool   `json:"connected"`
}

// StorageDevice represents a storage device configuration
type StorageDevice struct {
	Name       string `json:"name"`
	Controller string `json:"controller"`
	Type       string `json:"type"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	Attached   bool   `json:"attached"`
}

// TerraformConfig represents Terraform configuration for VM provisioning
type TerraformConfig struct {
	VMName        string            `json:"vm_name"`
	OSType        string            `json:"os_type"`
	MemoryMB      int               `json:"memory_mb"`
	CPUs          int               `json:"cpus"`
	DiskSizeGB    int               `json:"disk_size_gb"`
	NetworkType   string            `json:"network_type"`
	BridgeName    string            `json:"bridge_name"`
	BaseImagePath string            `json:"base_image_path"`
	CustomVars    map[string]string `json:"custom_vars"`
}

// TerraformOutput represents the output from Terraform execution
type TerraformOutput struct {
	VMUUID        string `json:"vm_uuid"`
	VMName        string `json:"vm_name"`
	State         string `json:"state"`
	IPAddress     string `json:"ip_address"`
	SSHPort       int    `json:"ssh_port"`
	VRDEPort      int    `json:"vrde_port"`
	DiskPath      string `json:"disk_path"`
	NetworkConfig string `json:"network_config"`
}

// VMOperation represents a VM operation request
type VMOperation struct {
	ID        string                 `json:"id"`
	Operation string                 `json:"operation"`
	VMID      string                 `json:"vm_id"`
	Params    map[string]interface{} `json:"params"`
	CreatedAt time.Time              `json:"created_at"`
}

// VMOperationResult represents the result of a VM operation
type VMOperationResult struct {
	ID          string                 `json:"id"`
	Operation   string                 `json:"operation"`
	VMID        string                 `json:"vm_id"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error"`
	Result      map[string]interface{} `json:"result"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at"`
}
