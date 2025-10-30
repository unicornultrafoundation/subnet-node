package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
)

func (api *VirtualBoxAPI) createVMFromImage(w http.ResponseWriter, r *http.Request) {

	var req vbtypes.CreateVMFromImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := api.ordersCache.GetOrder(r.Context(), req.OrderId)
	if err != nil {
		api.sendErrorResponse(w, "Order not found", http.StatusNotFound)
		return
	}

	vbReq := vbtypes.VMCreateFromImageRequest{
		Name:       "vm-" + order.ID.String(),
		CPUCores:   int(order.CpuCores.Int64()),
		MemoryMB:   int(order.MemoryMB.Int64()),
		DiskSizeGB: int(order.DiskGB.Int64()),
		OSType:     req.OS,
		Version:    req.Version,
		Username:   req.Username,
		Password:   req.Password,
		OrderId:    req.OrderId,
	}

	// Create the VM using the determined OS type
	jobResponse, err := api.vboxService.CreateVM(r.Context(), vbReq)
	if err != nil {
		api.sendErrorResponse(w, "Failed to create VM", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobResponse)
}

// getVMHandler handles GET requests for VMs - gets all VMs or a specific VM by orderId
func (api *VirtualBoxAPI) getVMHandler(w http.ResponseWriter, r *http.Request) {
	// Check if orderId is provided in the URL path
	orderId := chi.URLParam(r, "orderId")

	if orderId != "" {
		// Get specific VM by orderId
		vmId, exists := api.vboxService.GetVMIdByOrderId(orderId)
		if !exists {
			api.sendErrorResponse(w, "VM not found for this orderId", http.StatusNotFound)
			return
		}

		vm, err := api.vboxService.GetVM(r.Context(), vmId)
		if err != nil {
			api.sendErrorResponse(w, fmt.Sprintf("Failed to get VM: %v", err), http.StatusInternalServerError)
			return
		}

		if vm == nil {
			api.sendErrorResponse(w, "VM not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vm)
		return
	}

	// Get all VMs
	vms, _, err := api.vboxService.GetVMs(r.Context())
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to get VMs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vms)
}

// deleteVMHandler handles DELETE requests for a specific VM by orderId
func (api *VirtualBoxAPI) deleteVMHandler(w http.ResponseWriter, r *http.Request) {
	orderId := chi.URLParam(r, "orderId")
	if orderId == "" {
		api.sendErrorResponse(w, "orderId is required", http.StatusBadRequest)
		return
	}

	vmId, exists := api.vboxService.GetVMIdByOrderId(orderId)
	if !exists {
		api.sendErrorResponse(w, "VM not found for this orderId", http.StatusNotFound)
		return
	}

	err := api.vboxService.DeleteVM(r.Context(), vmId)
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to delete VM: %v", err), http.StatusInternalServerError)
		return
	}

	api.vboxService.RemoveOrderVMMapping(orderId)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "VM deleted successfully"})
}
