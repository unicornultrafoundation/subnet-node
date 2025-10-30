package types

import (
	"context"
	"os/exec"

	"github.com/gorilla/websocket"
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

	// OrderId-based VM Management
	StoreOrderVMMapping(orderId, vmId string)
	RemoveOrderVMMapping(orderId string)
	GetVMIdByOrderId(orderId string) (string, bool)

	// VM Control
	StartVM(ctx context.Context, vmID string) (*VM, error)
	StopVM(ctx context.Context, vmID string) (*VM, error)
	PauseVM(ctx context.Context, vmID string) (*VM, error)
	ResumeVM(ctx context.Context, vmID string) (*VM, error)
	ResetVM(ctx context.Context, vmID string) (*VM, error)

	GetSystemInfo(ctx context.Context) (*VMSystemInfo, error)

	// SSH Token Management
	GenerateSSHToken(ctx context.Context, vmID string, username string, password string) (*SSHTokenResponse, error)
	ValidateAndConsumeSSHToken(token string) (*SSHAccessToken, error)

	// Metrics Collection
	CollectMetrics(ctx context.Context, vmId string, conn *websocket.Conn, period int) error // Returns concise metrics data without redundant fields

	// Service Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type VBoxManageExecutor interface {
	CreateVM(vmName string, osType string) error
	ConfigureVMHardware(vmName string, cpuCount int, memoryMB int) error
	ConfigureNetwork(vmName string, networkType string) error
	SetupStorage(vmName string, req VMCreateRequest, isoPath string, cloudInitISO string) error
	StartVM(vmId string, headless bool) error
	StopVM(vmId string) error
	PauseVM(vmId string) error
	ResumeVM(vmId string) error
	ResetVM(vmId string) error
	DeleteVM(vmId string) error
	GetVMStatus(vmId string) (string, error)
	ListVMs() ([]string, error)
	CheckVBoxManageVersion() (string, error)

	// Metrics Collection
	ListAvailableMetrics(vmId string) (string, error)
	StartMetricsCollection(ctx context.Context, vmId string, metrics []string, period int) (*exec.Cmd, error)
}
