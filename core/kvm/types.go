package kvm

import (
	"math/big"
	"net"
	"time"
)

// VMStatus represents the current state of a VM
type VMStatus string

const (
	VMStatusStopped  VMStatus = "stopped"
	VMStatusRunning  VMStatus = "running"
	VMStatusPaused   VMStatus = "paused"
	VMStatusShutdown VMStatus = "shutdown"
	VMStatusCrashed  VMStatus = "crashed"
	VMStatusPending  VMStatus = "pending"
	VMStatusCreating VMStatus = "creating"
	VMStatusDeleting VMStatus = "deleting"
	VMStatusError    VMStatus = "error"
	VMStatusUnknown  VMStatus = "unknown"
)

// NetworkType defines the type of network configuration
type NetworkType string

const (
	NetworkTypeBridge   NetworkType = "bridge"
	NetworkTypeNAT      NetworkType = "nat"
	NetworkTypeIsolated NetworkType = "isolated"
)

// CreateVMRequest represents a request to create a new VM
type CreateVMRequest struct {
	Name      string            `json:"name" validate:"required"`
	Template  string            `json:"template" validate:"required"`
	CPU       int               `json:"cpu" validate:"min=1,max=16"`
	Memory    int               `json:"memory" validate:"min=512"` // MB
	Disk      int               `json:"disk" validate:"min=1"`     // GB
	Network   string            `json:"network"`
	CloudInit *CloudInitConfig  `json:"cloud_init,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// VMDetails represents detailed information about a VM
type VMDetails struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Status      VMStatus          `json:"status"`
	IPAddress   string            `json:"ip_address"`
	MACAddress  string            `json:"mac_address"`
	Resources   *ResourceInfo     `json:"resources"`
	ConsoleURL  string            `json:"console_url,omitempty"`
	LoginInfo   *LoginInfo        `json:"login_info,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Template    string            `json:"template"`
	NetworkType NetworkType       `json:"network_type"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// VM represents a virtual machine instance
type VM struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Status     VMStatus          `json:"status"`
	DomainName string            `json:"domain_name"`
	Resources  *ResourceInfo     `json:"resources"`
	Network    *NetworkInfo      `json:"network"`
	Storage    *StorageInfo      `json:"storage"`
	CreatedAt  time.Time         `json:"created_at"`
	Metadata   map[string]string `json:"metadata"`
}

// VMSpec defines the specification for creating a VM
type VMSpec struct {
	Name        string            `json:"name"`
	Template    string            `json:"template"`
	CPU         int               `json:"cpu"`
	Memory      int               `json:"memory"` // MB
	Disk        int               `json:"disk"`   // GB
	NetworkType NetworkType       `json:"network_type"`
	CloudInit   *CloudInitConfig  `json:"cloud_init,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// VMMetadata stores VM metadata in the registry
type VMMetadata struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Status      VMStatus            `json:"status"`
	IPAddress   string              `json:"ip_address"`
	MACAddress  string              `json:"mac_address"`
	Template    string              `json:"template"`
	Resources   *ResourceAllocation `json:"resources"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Metadata    map[string]string   `json:"metadata"`
	NetworkType NetworkType         `json:"network_type"`
	DomainName  string              `json:"domain_name"`
}

// ResourceInfo represents resource allocation for a VM
type ResourceInfo struct {
	CPU    int `json:"cpu"`
	Memory int `json:"memory"` // MB
	Disk   int `json:"disk"`   // GB
}

// ResourceAllocation represents allocated resources
type ResourceAllocation struct {
	CPU    *big.Int `json:"cpu"`
	Memory *big.Int `json:"memory"` // bytes
	Disk   *big.Int `json:"disk"`   // bytes
}

// ResourceRequirement defines resource requirements for validation
type ResourceRequirement struct {
	CPU    *big.Int `json:"cpu"`
	Memory *big.Int `json:"memory"` // bytes
	Disk   *big.Int `json:"disk"`   // bytes
}

// SystemCapabilities represents available system resources
type SystemCapabilities struct {
	TotalCPU         int      `json:"total_cpu"`
	TotalMemory      *big.Int `json:"total_memory"`  // bytes
	TotalStorage     *big.Int `json:"total_storage"` // bytes
	AvailableCPU     int      `json:"available_cpu"`
	AvailableMemory  *big.Int `json:"available_memory"`  // bytes
	AvailableStorage *big.Int `json:"available_storage"` // bytes
	MaxVMs           int      `json:"max_vms"`
	SupportedArch    []string `json:"supported_arch"`
}

// ResourceUsage represents current resource usage
type ResourceUsage struct {
	UsedCPU     int      `json:"used_cpu"`
	UsedMemory  *big.Int `json:"used_memory"`  // bytes
	UsedStorage *big.Int `json:"used_storage"` // bytes
	TotalVMs    int      `json:"total_vms"`
	RunningVMs  int      `json:"running_vms"`
}

// NetworkInfo represents network configuration
type NetworkInfo struct {
	Name        string      `json:"name"`
	Type        NetworkType `json:"type"`
	IPAddress   net.IP      `json:"ip_address"`
	MACAddress  string      `json:"mac_address"`
	BridgeName  string      `json:"bridge_name,omitempty"`
	NetworkCIDR string      `json:"network_cidr,omitempty"`
}

// NetworkConfig represents network configuration for a VM
type NetworkConfig struct {
	Type       NetworkType `json:"type"`
	IPAddress  net.IP      `json:"ip_address"`
	MACAddress string      `json:"mac_address"`
	Gateway    net.IP      `json:"gateway,omitempty"`
	DNS        []net.IP    `json:"dns,omitempty"`
	BridgeName string      `json:"bridge_name,omitempty"`
}

// StorageInfo represents storage information
type StorageInfo struct {
	DiskPath    string `json:"disk_path"`
	Size        uint64 `json:"size"`   // bytes
	Format      string `json:"format"` // qcow2, raw, etc.
	BackingFile string `json:"backing_file,omitempty"`
}

// DiskInfo represents disk information
type DiskInfo struct {
	Path          string `json:"path"`
	Size          uint64 `json:"size"`           // bytes
	AllocatedSize uint64 `json:"allocated_size"` // bytes
	Format        string `json:"format"`
	BackingFile   string `json:"backing_file,omitempty"`
}

// VolumeInfo represents storage volume information
type VolumeInfo struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Capacity   uint64 `json:"capacity"`   // bytes
	Allocation uint64 `json:"allocation"` // bytes
	Format     string `json:"format"`
}

// ImageTemplate represents a VM template
type ImageTemplate struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	OS           string            `json:"os"`
	Version      string            `json:"version"`
	Architecture string            `json:"architecture"`
	Size         uint64            `json:"size"` // bytes
	Format       string            `json:"format"`
	Description  string            `json:"description"`
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
}

// CloudInitConfig represents cloud-init configuration
type CloudInitConfig struct {
	Hostname      string            `yaml:"hostname" json:"hostname"`
	Users         []User            `yaml:"users" json:"users"`
	SSHKeys       []string          `yaml:"ssh_authorized_keys" json:"ssh_authorized_keys"`
	Packages      []string          `yaml:"packages" json:"packages"`
	RunCommands   []string          `yaml:"runcmd" json:"runcmd,omitempty"`
	WriteFiles    []WriteFile       `yaml:"write_files" json:"write_files,omitempty"`
	NetworkConfig *NetworkConfig    `yaml:"network" json:"network,omitempty"`
	Metadata      map[string]string `yaml:"metadata" json:"metadata,omitempty"`
}

// User represents a user in cloud-init configuration
type User struct {
	Name              string   `yaml:"name" json:"name"`
	Gecos             string   `yaml:"gecos,omitempty" json:"gecos,omitempty"`
	Shell             string   `yaml:"shell,omitempty" json:"shell,omitempty"`
	Groups            []string `yaml:"groups,omitempty" json:"groups,omitempty"`
	Sudo              string   `yaml:"sudo,omitempty" json:"sudo,omitempty"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty" json:"ssh_authorized_keys,omitempty"`
	LockPasswd        bool     `yaml:"lock_passwd,omitempty" json:"lock_passwd,omitempty"`
}

// WriteFile represents a file to be written during cloud-init
type WriteFile struct {
	Path        string `yaml:"path" json:"path"`
	Content     string `yaml:"content" json:"content"`
	Owner       string `yaml:"owner,omitempty" json:"owner,omitempty"`
	Permissions string `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Encoding    string `yaml:"encoding,omitempty" json:"encoding,omitempty"`
}

// LoginInfo represents VM login information
type LoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	SSHKey   string `json:"ssh_key,omitempty"`
	SSHPort  int    `json:"ssh_port,omitempty"`
}

// KVMConfig represents KVM service configuration
type KVMConfig struct {
	Enabled          bool                      `yaml:"enabled" mapstructure:"enabled"`
	LibvirtURI       string                    `yaml:"libvirt_uri" mapstructure:"libvirt_uri"`
	StoragePath      string                    `yaml:"storage_path" mapstructure:"storage_path"`
	TemplatePath     string                    `yaml:"template_path" mapstructure:"template_path"`
	Networks         map[string]*NetworkConfig `yaml:"networks" mapstructure:"networks"`
	DefaultResources *DefaultResources         `yaml:"default_resources" mapstructure:"default_resources"`
	MaxVMs           int                       `yaml:"max_vms" mapstructure:"max_vms"`
	ReservedCPU      float64                   `yaml:"reserved_cpu" mapstructure:"reserved_cpu"`
	ReservedMemory   float64                   `yaml:"reserved_memory" mapstructure:"reserved_memory"`
	DefaultNetwork   string                    `yaml:"default_network" mapstructure:"default_network"`
	StoragePool      string                    `yaml:"storage_pool" mapstructure:"storage_pool"`
}

// DefaultResources represents default resource allocation
type DefaultResources struct {
	CPU    int `yaml:"cpu" mapstructure:"cpu"`
	Memory int `yaml:"memory" mapstructure:"memory"` // MB
	Disk   int `yaml:"disk" mapstructure:"disk"`     // GB
}

// VMCreateOptions represents options for VM creation
type VMCreateOptions struct {
	AutoStart   bool   `json:"auto_start"`
	ConsoleType string `json:"console_type,omitempty"`
	VNCPassword string `json:"vnc_password,omitempty"`
	BootOrder   string `json:"boot_order,omitempty"`
}

// Error types for KVM operations
type KVMError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *KVMError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// Common error codes
const (
	ErrCodeResourceInsufficient = "RESOURCE_INSUFFICIENT"
	ErrCodeVMNotFound           = "VM_NOT_FOUND"
	ErrCodeVMAlreadyExists      = "VM_ALREADY_EXISTS"
	ErrCodeTemplateNotFound     = "TEMPLATE_NOT_FOUND"
	ErrCodeNetworkError         = "NETWORK_ERROR"
	ErrCodeStorageError         = "STORAGE_ERROR"
	ErrCodeLibvirtError         = "LIBVIRT_ERROR"
	ErrCodeValidationError      = "VALIDATION_ERROR"
	ErrCodeServiceUnavailable   = "SERVICE_UNAVAILABLE"
)
