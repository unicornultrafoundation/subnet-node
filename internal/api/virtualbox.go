package api

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/virtualbox"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

// VirtualBoxAPI represents the VirtualBox RPC API
type VirtualBoxAPI struct {
	service *virtualbox.Service
}

// NewVirtualBoxAPI creates a new VirtualBox API
func NewVirtualBoxAPI(service *virtualbox.Service) *VirtualBoxAPI {
	return &VirtualBoxAPI{
		service: service,
	}
}

// CreateVM creates a new VM
func (api *VirtualBoxAPI) CreateVM(ctx context.Context, request types.VMRequest) (*types.VMResponse, error) {

	return api.service.CreateVM(ctx, &request)
}

// CreateVMWithTerraform creates a new VM using Terraform
func (api *VirtualBoxAPI) CreateVMWithTerraform(ctx context.Context, request types.VMRequest) (*types.VMResponse, error) {
	return api.service.CreateVMWithTerraform(ctx, &request)
}

// GetVMs retrieves all VMs for a requester
func (api *VirtualBoxAPI) GetVMs(ctx context.Context, requester string) ([]*types.VMResponse, error) {
	return api.service.GetVMs(ctx, requester)
}

// GetVM retrieves a specific VM by ID
func (api *VirtualBoxAPI) GetVM(ctx context.Context, vmID string) (*types.VMResponse, error) {
	return api.service.GetVM(ctx, vmID)
}

// StartVM starts a VM
func (api *VirtualBoxAPI) StartVM(ctx context.Context, vmID string) error {
	return api.service.StartVM(ctx, vmID)
}

// StopVM stops a VM
func (api *VirtualBoxAPI) StopVM(ctx context.Context, vmID string) error {
	return api.service.StopVM(ctx, vmID)
}

// ShutdownVM shuts down a VM gracefully
func (api *VirtualBoxAPI) ShutdownVM(ctx context.Context, vmID string) error {
	return api.service.ShutdownVM(ctx, vmID)
}

// DeleteVM deletes a VM
func (api *VirtualBoxAPI) DeleteVM(ctx context.Context, vmID string) error {
	return api.service.DeleteVM(ctx, vmID)
}

// GetVMRequest retrieves a VM request by ID
func (api *VirtualBoxAPI) GetVMRequest(ctx context.Context, vmID string) (*types.VMRequest, error) {
	return api.service.GetVMRequest(ctx, vmID)
}
