package kvm

import (
	"context"
	"math/big"
	"net"
)

// KVMService defines the main KVM service interface
type KVMService interface {
	// VM Lifecycle
	CreateVM(ctx context.Context, req *CreateVMRequest) (*VMDetails, error)
	DeleteVM(ctx context.Context, vmID string) error
	ListVMs(ctx context.Context, filters map[string]string) ([]*VMDetails, error)
	GetVM(ctx context.Context, vmID string) (*VMDetails, error)
	StartVM(ctx context.Context, vmID string) error
	StopVM(ctx context.Context, vmID string) error
	RestartVM(ctx context.Context, vmID string) error

	// Service lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsHealthy(ctx context.Context) error
}

// VMManager handles VM lifecycle operations
type VMManager interface {
	Provision(ctx context.Context, spec *VMSpec) (*VM, error)
	Start(ctx context.Context, vmID string) error
	Stop(ctx context.Context, vmID string) error
	Restart(ctx context.Context, vmID string) error
	Destroy(ctx context.Context, vmID string) error
	GetStatus(ctx context.Context, vmID string) (VMStatus, error)
	ListDomains(ctx context.Context) ([]*VM, error)
}

// ResourceChecker validates system resources for VM operations
type ResourceChecker interface {
	CheckAvailableResources(req *ResourceRequirement) error
	GetSystemCapabilities() (*SystemCapabilities, error)
	ValidateVMResources(cpu, memory, disk *big.Int) error
	GetResourceUsage() (*ResourceUsage, error)
}

// StorageManager handles disk image and storage operations
type StorageManager interface {
	CreateDiskFromTemplate(templateName, vmID string, size uint64) (string, error)
	DeleteDisk(vmID string) error
	ListTemplates() ([]*ImageTemplate, error)
	CreateCloudInitISO(config *CloudInitConfig, vmID string) (string, error)
	GetDiskInfo(vmID string) (*DiskInfo, error)
}

// NetworkManager handles VM networking
type NetworkManager interface {
	SetupVMNetworking(vmID string, networkType NetworkType) (*NetworkConfig, error)
	AllocateIP(vmID string) (net.IP, error)
	ReleaseIP(vmID string) error
	GetNetworkInfo(vmID string) (*NetworkInfo, error)
	ListNetworks() ([]*NetworkInfo, error)
}

// VMRegistry handles VM metadata storage and retrieval
type VMRegistry interface {
	RegisterVM(vm *VMMetadata) error
	UpdateVM(vmID string, updates map[string]interface{}) error
	UnregisterVM(vmID string) error
	GetVM(vmID string) (*VMMetadata, error)
	ListVMs(filters map[string]string) ([]*VMMetadata, error)
	VMExists(vmID string) bool
}

// LibvirtClient wraps libvirt operations
type LibvirtClient interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
	CreateDomain(xml string) (Domain, error)
	GetDomain(name string) (Domain, error)
	ListDomains() ([]Domain, error)
	GetStoragePool(name string) (StoragePool, error)
	GetNetwork(name string) (Network, error)
}

// Domain represents a libvirt domain (VM)
type Domain interface {
	GetName() (string, error)
	GetState() (VMStatus, error)
	Start() error
	Stop() error
	Restart() error
	Destroy() error
	GetXML() (string, error)
}

// StoragePool represents a libvirt storage pool
type StoragePool interface {
	GetName() (string, error)
	CreateVolume(xml string) (StorageVolume, error)
	GetVolume(name string) (StorageVolume, error)
	ListVolumes() ([]StorageVolume, error)
}

// StorageVolume represents a libvirt storage volume
type StorageVolume interface {
	GetName() (string, error)
	GetPath() (string, error)
	Delete() error
	GetInfo() (*VolumeInfo, error)
}

// Network represents a libvirt network
type Network interface {
	GetName() (string, error)
	GetBridgeName() (string, error)
	IsActive() (bool, error)
}
