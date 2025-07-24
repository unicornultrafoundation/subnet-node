package virtualbox

import (
	"context"

	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

// Service defines the interface for VirtualBox operations
type Service interface {
	// VM Management
	CreateVM(ctx context.Context, req vbtypes.VMCreateRequest) (*vbtypes.VM, error)
	CreateAndStartVM(ctx context.Context, req vbtypes.VMCreateRequest) (*vbtypes.VM, error)
	GetVM(ctx context.Context, vmID string) (*vbtypes.VM, error)
	GetVMs(ctx context.Context) ([]*vbtypes.VM, int, error)
	UpdateVM(ctx context.Context, vmID string, req vbtypes.VMUpdateRequest) (*vbtypes.VM, error)
	DeleteVM(ctx context.Context, vmID string) error
	SyncVMs(ctx context.Context) error

	// VM Control
	StartVM(ctx context.Context, vmID string) (*vbtypes.VM, error)
	StopVM(ctx context.Context, vmID string) (*vbtypes.VM, error)
	PauseVM(ctx context.Context, vmID string) (*vbtypes.VM, error)
	ResumeVM(ctx context.Context, vmID string) (*vbtypes.VM, error)
	ResetVM(ctx context.Context, vmID string) (*vbtypes.VM, error)

	GetSystemInfo(ctx context.Context) (*vbtypes.VMSystemInfo, error)

	// ISO Management
	DownloadISO(ctx context.Context, isoURL string) (*vbtypes.ISOInfo, error)
	ListOSTypes(ctx context.Context) ([]string, error)

	// Service Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// StorageManager defines the interface for file storage operations
type StorageManager interface {
	// ISO Management
	DownloadFile(ctx context.Context, url, destPath string) error
	GetFileInfo(filePath string) (*vbtypes.ISOInfo, error)
	ListFiles(dirPath string) ([]string, error)
	DeleteFile(filePath string) error
	FileExists(filePath string) bool
	GetFileSize(filePath string) (int64, error)
	CalculateChecksum(filePath string) (string, error)

	// ISO OS Type Management
	DetermineOSTypeAndISO(ctx context.Context, req vbtypes.VMCreateRequest) (string, string, error)
	GetSupportedOSTypes() []string
}
