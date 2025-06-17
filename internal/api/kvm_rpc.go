package api

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/kvm"
)

// KVMRPCAPI provides JSON-RPC API for KVM operations
type KVMRPCAPI struct {
	kvmService *kvm.Service
}

// NewKVMRPCAPI creates a new KVM RPC API handler
func NewKVMRPCAPI(kvmService *kvm.Service) *KVMRPCAPI {
	return &KVMRPCAPI{
		kvmService: kvmService,
	}
}

// CreateVM creates a new virtual machine
func (api *KVMRPCAPI) CreateVM(ctx context.Context, req *kvm.CreateVMRequest) (*kvm.VM, error) {
	if !api.kvmService.IsEnabled() {
		return nil, &APIError{Code: ServiceUnavailable, Message: "KVM service is disabled"}
	}
	return api.kvmService.CreateVM(ctx, req)
}

// GetVM retrieves a VM by ID
func (api *KVMRPCAPI) GetVM(ctx context.Context, vmID string) (*kvm.VM, error) {
	if !api.kvmService.IsEnabled() {
		return nil, &APIError{Code: ServiceUnavailable, Message: "KVM service is disabled"}
	}
	return api.kvmService.GetVM(ctx, vmID)
}

// ListVMs returns all VMs
func (api *KVMRPCAPI) ListVMs(ctx context.Context) ([]*kvm.VM, error) {
	if !api.kvmService.IsEnabled() {
		return nil, &APIError{Code: ServiceUnavailable, Message: "KVM service is disabled"}
	}
	return api.kvmService.ListVMs(ctx)
}

// StartVM starts a virtual machine
func (api *KVMRPCAPI) StartVM(ctx context.Context, vmID string) error {
	if !api.kvmService.IsEnabled() {
		return &APIError{Code: ServiceUnavailable, Message: "KVM service is disabled"}
	}
	return api.kvmService.StartVM(ctx, vmID)
}

// StopVM stops a virtual machine
func (api *KVMRPCAPI) StopVM(ctx context.Context, vmID string) error {
	if !api.kvmService.IsEnabled() {
		return &APIError{Code: ServiceUnavailable, Message: "KVM service is disabled"}
	}
	return api.kvmService.StopVM(ctx, vmID)
}

// DeleteVM deletes a virtual machine
func (api *KVMRPCAPI) DeleteVM(ctx context.Context, vmID string) error {
	if !api.kvmService.IsEnabled() {
		return &APIError{Code: ServiceUnavailable, Message: "KVM service is disabled"}
	}
	return api.kvmService.DeleteVM(ctx, vmID)
}

// GetSystemResources returns system resource information
func (api *KVMRPCAPI) GetSystemResources(ctx context.Context) (*kvm.SystemResources, error) {
	if !api.kvmService.IsEnabled() {
		return nil, &APIError{Code: ServiceUnavailable, Message: "KVM service is disabled"}
	}
	return api.kvmService.GetSystemResources(ctx)
}

// GetVMStats returns VM statistics
func (api *KVMRPCAPI) GetVMStats(ctx context.Context, vmID string) (*kvm.VMStats, error) {
	if !api.kvmService.IsEnabled() {
		return nil, &APIError{Code: ServiceUnavailable, Message: "KVM service is disabled"}
	}
	return api.kvmService.GetVMStats(ctx, vmID)
}

// Status returns KVM service status
func (api *KVMRPCAPI) Status(ctx context.Context) (map[string]interface{}, error) {
	status := map[string]interface{}{
		"enabled": api.kvmService.IsEnabled(),
		"service": "kvm",
	}

	if api.kvmService.IsEnabled() {
		vms, err := api.kvmService.ListVMs(ctx)
		if err == nil {
			status["total_vms"] = len(vms)
			running := 0
			for _, vm := range vms {
				if vm.Status == kvm.VMStatusRunning {
					running++
				}
			}
			status["running_vms"] = running
		}
	}

	return status, nil
}

// APIError represents an API error
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

// ErrorCode represents error codes
type ErrorCode int

const (
	ServiceUnavailable ErrorCode = 503
)
