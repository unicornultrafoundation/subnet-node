package types

import (
	"context"
)

// Service defines the interface for VirtualBox operations
type Service interface {
	// VM Management
	CreateVM(ctx context.Context, req VMCreateRequest) (*VM, error)
	CreateTemplateVM(ctx context.Context, req VMCreateRequest) (*VM, error)
	GetVM(ctx context.Context, vmID string) (*VM, error)
	GetVMs(ctx context.Context) ([]*VM, int, error)
	UpdateVM(ctx context.Context, vmID string, req VMUpdateRequest) (*VM, error)
	DeleteVM(ctx context.Context, vmID string) error
	SyncVMs(ctx context.Context) error

	// VM Control
	StartVM(ctx context.Context, vmID string) (*VM, error)
	StopVM(ctx context.Context, vmID string) (*VM, error)
	PauseVM(ctx context.Context, vmID string) (*VM, error)
	ResumeVM(ctx context.Context, vmID string) (*VM, error)
	ResetVM(ctx context.Context, vmID string) (*VM, error)

	GetSystemInfo(ctx context.Context) (*VMSystemInfo, error)

	// ISO Management
	DownloadISO(ctx context.Context, isoURL string) (*ISOInfo, error)
	ListOSTypes(ctx context.Context) ([]string, error)

	// SSH Token Management
	GenerateSSHToken(ctx context.Context, vmID string, username string, password string) (*SSHTokenResponse, error)
	ValidateAndConsumeSSHToken(token string) (*SSHAccessToken, error)

	// Service Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// StorageManager defines the interface for file storage operations
type StorageManager interface {
	// ISO Management
	DownloadFile(ctx context.Context, url, destPath string) error
	GetFileInfo(filePath string) (*ISOInfo, error)
	ListFiles(dirPath string) ([]string, error)
	DeleteFile(filePath string) error
	FileExists(filePath string) bool
	GetFileSize(filePath string) (int64, error)
	CalculateChecksum(filePath string) (string, error)

	// ISO OS Type Management
	DetermineOSTypeAndISO(ctx context.Context, req VMCreateRequest) (string, string, error)
	GetSupportedOSTypes() []string
}

type VBoxManageExecutor interface {
	CreateVM(vmName string, osType string) error
	ConfigureVMHardware(vmName string, cpuCount int, memoryMB int) error
	ConfigureNetwork(vmName string, networkType string) error
	GenerateCloudInitFiles(vmName, hostname, username, password string) (metaDataPath, userDataPath, cloudInitDir string, err error)
	GenerateCloneVMCloudInitFiles(vmName, hostname, username, password string) (metaDataPath, userDataPath, cloudInitDir string, err error)
	GenerateCloudInitISO(cloudInitDir string) (string, error)
	AttachCloudInitISO(vmName, isoPath string) error
	SetupStorage(vmName string, req VMCreateRequest, isoPath string, cloudInitISO string) error
	StartVM(vmName string, headless bool) error
	StopVM(vmName string) error
	PauseVM(vmName string) error
	ResumeVM(vmName string) error
	ResetVM(vmName string) error
	DeleteVM(vmName string) error
	GetVMStatus(vmName string) (string, error)
	ListVMs() ([]string, error)
	CheckVBoxManageVersion() (string, error)
}
