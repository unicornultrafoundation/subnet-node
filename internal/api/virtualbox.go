package api

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

type vmResult struct {
	ID         string           `json:"id,omitempty"`
	Name       string           `json:"name,omitempty"`
	Status     vbtypes.VMStatus `json:"status,omitempty"`
	CPUCores   int              `json:"cpu_cores,omitempty"`
	MemoryMB   int              `json:"memory_mb,omitempty"`
	DiskSizeGB int              `json:"disk_size_gb,omitempty"`
	ISOURL     string           `json:"iso_url,omitempty"`
	ISOPath    string           `json:"iso_path,omitempty"`
	CreatedAt  string           `json:"created_at,omitempty"`
	UpdatedAt  string           `json:"updated_at,omitempty"`
	IPAddress  string           `json:"ip_address,omitempty"`
	SSHPort    int              `json:"ssh_port,omitempty"`
	VBoxPath   string           `json:"vbox_path,omitempty"`
	VMFolder   string           `json:"vm_folder,omitempty"`
}

type vmFilter struct {
	Status string `json:"status,omitempty"`
	Query  string `json:"query,omitempty"`
}

type vmUsageResult struct {
	VMID          string  `json:"vm_id"`
	CPUPerc       float64 `json:"cpu_perc"`
	MemoryMB      int     `json:"memory_mb"`
	DiskUsageGB   float64 `json:"disk_usage_gb"`
	NetworkRxMB   float64 `json:"network_rx_mb"`
	NetworkTxMB   float64 `json:"network_tx_mb"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	Timestamp     string  `json:"timestamp"`
}

type vmSystemInfoResult struct {
	VBoxVersion     string `json:"vbox_version"`
	HostOS          string `json:"host_os"`
	HostArch        string `json:"host_arch"`
	AvailableCPUs   int    `json:"available_cpus"`
	AvailableRAMMB  int    `json:"available_ram_mb"`
	AvailableDiskGB int    `json:"available_disk_gb"`
}

type isoInfoResult struct {
	URL          string `json:"url"`
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	Checksum     string `json:"checksum"`
	DownloadedAt string `json:"downloaded_at"`
}

func convertToVMResult(vm *vbtypes.VM) *vmResult {
	if vm == nil {
		return nil
	}
	return &vmResult{
		ID:         vm.ID,
		Name:       vm.Name,
		Status:     vm.Status,
		CPUCores:   vm.CPUCores,
		MemoryMB:   vm.MemoryMB,
		DiskSizeGB: vm.DiskSizeGB,
		ISOURL:     vm.ISOURL,
		ISOPath:    vm.ISOPath,
		CreatedAt:  vm.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  vm.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		IPAddress:  vm.IPAddress,
		SSHPort:    vm.SSHPort,
		VBoxPath:   vm.VBoxPath,
		VMFolder:   vm.VMFolder,
	}
}

func convertToVMUsageResult(usage *vbtypes.VMUsage) *vmUsageResult {
	if usage == nil {
		return nil
	}
	return &vmUsageResult{
		VMID:          usage.VMID,
		CPUPerc:       usage.CPUPerc,
		MemoryMB:      usage.MemoryMB,
		DiskUsageGB:   usage.DiskUsageGB,
		NetworkRxMB:   usage.NetworkRxMB,
		NetworkTxMB:   usage.NetworkTxMB,
		UptimeSeconds: usage.UptimeSeconds,
		Timestamp:     usage.Timestamp.Format("2006-01-02T15:04:05Z"),
	}
}

func convertToSystemInfoResult(info *vbtypes.VMSystemInfo) *vmSystemInfoResult {
	if info == nil {
		return nil
	}
	return &vmSystemInfoResult{
		VBoxVersion:     info.VBoxVersion,
		HostOS:          info.HostOS,
		HostArch:        info.HostArch,
		AvailableCPUs:   info.AvailableCPUs,
		AvailableRAMMB:  info.AvailableRAMMB,
		AvailableDiskGB: info.AvailableDiskGB,
	}
}

func convertToISOInfoResult(iso *vbtypes.ISOInfo) *isoInfoResult {
	if iso == nil {
		return nil
	}
	return &isoInfoResult{
		URL:          iso.URL,
		Path:         iso.Path,
		Size:         iso.Size,
		Checksum:     iso.Checksum,
		DownloadedAt: iso.DownloadedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type VirtualBoxAPI struct {
	vboxService virtualbox.Service
}

// NewVirtualBoxAPI creates a new instance of VirtualBoxAPI.
func NewVirtualBoxAPI(vboxService virtualbox.Service) *VirtualBoxAPI {
	return &VirtualBoxAPI{vboxService: vboxService}
}

func (api *VirtualBoxAPI) GetVMCount(ctx context.Context, vmFilter *vmFilter) (*hexutil.Big, error) {
	var filter *vbtypes.VMFilter
	if vmFilter != nil {
		filter = &vbtypes.VMFilter{
			Status: vmFilter.Status,
			Query:  vmFilter.Query,
		}
	}

	vmCount, err := api.vboxService.GetVMCount(ctx, filter)
	if err != nil {
		return nil, err
	}

	return (*hexutil.Big)(vmCount), nil
}

func (api *VirtualBoxAPI) GetVMs(ctx context.Context, start int64, end int64, vmFilter *vmFilter) ([]vmResult, error) {
	var filter vbtypes.VMFilter
	if vmFilter != nil {
		filter = vbtypes.VMFilter{
			Status: vmFilter.Status,
			Query:  vmFilter.Query,
		}
	}

	vms, _, err := api.vboxService.GetVMs(ctx, big.NewInt(start), big.NewInt(end), filter)
	if err != nil {
		return nil, err
	}

	result := make([]vmResult, len(vms))
	for i, vm := range vms {
		result[i] = *convertToVMResult(vm)
	}
	return result, nil
}

func (api *VirtualBoxAPI) GetVM(ctx context.Context, vmID string) (*vmResult, error) {
	vm, err := api.vboxService.GetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}
	return convertToVMResult(vm), nil
}

func (api *VirtualBoxAPI) CreateVM(ctx context.Context, name string, cpuCores int, memoryMB int, diskSizeGB int, osType string) (*vmResult, error) {
	vbReq := vbtypes.VMCreateRequest{
		Name:       name,
		CPUCores:   cpuCores,
		MemoryMB:   memoryMB,
		DiskSizeGB: diskSizeGB,
		OSType:     osType, // Pass the OS type to the service
	}

	vm, err := api.vboxService.CreateVM(ctx, vbReq)
	if err != nil {
		return nil, err
	}

	return convertToVMResult(vm), nil
}

func (api *VirtualBoxAPI) UpdateVM(ctx context.Context, vmID string, name string, cpuCores int, memoryMB int, diskSizeGB int) (*vmResult, error) {
	vbReq := vbtypes.VMUpdateRequest{
		Name:       name,
		CPUCores:   cpuCores,
		MemoryMB:   memoryMB,
		DiskSizeGB: diskSizeGB,
	}

	vm, err := api.vboxService.UpdateVM(ctx, vmID, vbReq)
	if err != nil {
		return nil, err
	}

	return convertToVMResult(vm), nil
}

func (api *VirtualBoxAPI) DeleteVM(ctx context.Context, vmID string) error {
	return api.vboxService.DeleteVM(ctx, vmID)
}

func (api *VirtualBoxAPI) StartVM(ctx context.Context, vmID string) (*vmResult, error) {
	vm, err := api.vboxService.StartVM(ctx, vmID)
	if err != nil {
		return nil, err
	}
	return convertToVMResult(vm), nil
}

func (api *VirtualBoxAPI) StopVM(ctx context.Context, vmID string) (*vmResult, error) {
	vm, err := api.vboxService.StopVM(ctx, vmID)
	if err != nil {
		return nil, err
	}
	return convertToVMResult(vm), nil
}

func (api *VirtualBoxAPI) PauseVM(ctx context.Context, vmID string) (*vmResult, error) {
	vm, err := api.vboxService.PauseVM(ctx, vmID)
	if err != nil {
		return nil, err
	}
	return convertToVMResult(vm), nil
}

func (api *VirtualBoxAPI) ResumeVM(ctx context.Context, vmID string) (*vmResult, error) {
	vm, err := api.vboxService.ResumeVM(ctx, vmID)
	if err != nil {
		return nil, err
	}
	return convertToVMResult(vm), nil
}

func (api *VirtualBoxAPI) ResetVM(ctx context.Context, vmID string) (*vmResult, error) {
	vm, err := api.vboxService.ResetVM(ctx, vmID)
	if err != nil {
		return nil, err
	}
	return convertToVMResult(vm), nil
}

func (api *VirtualBoxAPI) GetVMUsage(ctx context.Context, vmID string) (*vmUsageResult, error) {
	usage, err := api.vboxService.GetVMUsage(ctx, vmID)
	if err != nil {
		return nil, err
	}
	return convertToVMUsageResult(usage), nil
}

func (api *VirtualBoxAPI) GetAllVMUsage(ctx context.Context) (*vmUsageResult, error) {
	usage, err := api.vboxService.GetAllVMUsage(ctx)
	if err != nil {
		return nil, err
	}
	return convertToVMUsageResult(usage), nil
}

func (api *VirtualBoxAPI) GetSystemInfo(ctx context.Context) (*vmSystemInfoResult, error) {
	info, err := api.vboxService.GetSystemInfo(ctx)
	if err != nil {
		return nil, err
	}
	return convertToSystemInfoResult(info), nil
}

func (api *VirtualBoxAPI) DownloadISO(ctx context.Context, isoURL string) (*isoInfoResult, error) {
	iso, err := api.vboxService.DownloadISO(ctx, isoURL)
	if err != nil {
		return nil, err
	}
	return convertToISOInfoResult(iso), nil
}

func (api *VirtualBoxAPI) GetISOInfo(ctx context.Context, isoPath string) (*isoInfoResult, error) {
	iso, err := api.vboxService.GetISOInfo(ctx, isoPath)
	if err != nil {
		return nil, err
	}
	return convertToISOInfoResult(iso), nil
}

func (api *VirtualBoxAPI) ListISOs(ctx context.Context) ([]isoInfoResult, error) {
	isos, err := api.vboxService.ListISOs(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]isoInfoResult, len(isos))
	for i, iso := range isos {
		result[i] = *convertToISOInfoResult(iso)
	}
	return result, nil
}

func (api *VirtualBoxAPI) DeleteISO(ctx context.Context, isoPath string) error {
	return api.vboxService.DeleteISO(ctx, isoPath)
}

func (api *VirtualBoxAPI) ListOSTypes(ctx context.Context) ([]string, error) {
	return api.vboxService.ListOSTypes(ctx)
}
