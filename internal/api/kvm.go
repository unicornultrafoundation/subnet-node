package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/kvm"
)

// KVMAPI provides HTTP API for KVM operations
type KVMAPI struct {
	kvmService *kvm.Service
	logger     *logrus.Entry
}

// NewKVMAPI creates a new KVM API handler
func NewKVMAPI(kvmService *kvm.Service, logger *logrus.Entry) *KVMAPI {
	return &KVMAPI{
		kvmService: kvmService,
		logger:     logger.WithField("api", "kvm"),
	}
}

// RegisterRoutes registers KVM API routes
func (api *KVMAPI) RegisterRoutes(router *mux.Router) {
	kvmRouter := router.PathPrefix("/kvm").Subrouter()

	// VM management endpoints
	kvmRouter.HandleFunc("/vms", api.listVMs).Methods("GET")
	kvmRouter.HandleFunc("/vms", api.createVM).Methods("POST")
	kvmRouter.HandleFunc("/vms/{vmID}", api.getVM).Methods("GET")
	kvmRouter.HandleFunc("/vms/{vmID}", api.deleteVM).Methods("DELETE")
	kvmRouter.HandleFunc("/vms/{vmID}/start", api.startVM).Methods("POST")
	kvmRouter.HandleFunc("/vms/{vmID}/stop", api.stopVM).Methods("POST")
	kvmRouter.HandleFunc("/vms/{vmID}/stats", api.getVMStats).Methods("GET")

	// System information endpoints
	kvmRouter.HandleFunc("/resources", api.getSystemResources).Methods("GET")
	kvmRouter.HandleFunc("/status", api.getStatus).Methods("GET")
}

// listVMs returns all VMs
func (api *KVMAPI) listVMs(w http.ResponseWriter, r *http.Request) {
	if !api.kvmService.IsEnabled() {
		api.writeError(w, http.StatusServiceUnavailable, "KVM service is disabled")
		return
	}

	ctx := r.Context()
	vms, err := api.kvmService.ListVMs(ctx)
	if err != nil {
		api.logger.WithError(err).Error("Failed to list VMs")
		api.writeError(w, http.StatusInternalServerError, "Failed to list VMs")
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]interface{}{
		"vms":   vms,
		"count": len(vms),
	})
}

// createVM creates a new VM
func (api *KVMAPI) createVM(w http.ResponseWriter, r *http.Request) {
	if !api.kvmService.IsEnabled() {
		api.writeError(w, http.StatusServiceUnavailable, "KVM service is disabled")
		return
	}

	var req kvm.CreateVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()
	vm, err := api.kvmService.CreateVM(ctx, &req)
	if err != nil {
		api.logger.WithError(err).Error("Failed to create VM")
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusCreated, vm)
}

// getVM returns a specific VM
func (api *KVMAPI) getVM(w http.ResponseWriter, r *http.Request) {
	if !api.kvmService.IsEnabled() {
		api.writeError(w, http.StatusServiceUnavailable, "KVM service is disabled")
		return
	}

	vars := mux.Vars(r)
	vmID := vars["vmID"]

	ctx := r.Context()
	vm, err := api.kvmService.GetVM(ctx, vmID)
	if err != nil {
		api.logger.WithError(err).WithField("vm_id", vmID).Error("Failed to get VM")
		api.writeError(w, http.StatusNotFound, "VM not found")
		return
	}

	api.writeJSON(w, http.StatusOK, vm)
}

// deleteVM deletes a VM
func (api *KVMAPI) deleteVM(w http.ResponseWriter, r *http.Request) {
	if !api.kvmService.IsEnabled() {
		api.writeError(w, http.StatusServiceUnavailable, "KVM service is disabled")
		return
	}

	vars := mux.Vars(r)
	vmID := vars["vmID"]

	ctx := r.Context()
	if err := api.kvmService.DeleteVM(ctx, vmID); err != nil {
		api.logger.WithError(err).WithField("vm_id", vmID).Error("Failed to delete VM")
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "VM deleted successfully",
		"vm_id":   vmID,
	})
}

// startVM starts a VM
func (api *KVMAPI) startVM(w http.ResponseWriter, r *http.Request) {
	if !api.kvmService.IsEnabled() {
		api.writeError(w, http.StatusServiceUnavailable, "KVM service is disabled")
		return
	}

	vars := mux.Vars(r)
	vmID := vars["vmID"]

	ctx := r.Context()
	if err := api.kvmService.StartVM(ctx, vmID); err != nil {
		api.logger.WithError(err).WithField("vm_id", vmID).Error("Failed to start VM")
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "VM start initiated",
		"vm_id":   vmID,
	})
}

// stopVM stops a VM
func (api *KVMAPI) stopVM(w http.ResponseWriter, r *http.Request) {
	if !api.kvmService.IsEnabled() {
		api.writeError(w, http.StatusServiceUnavailable, "KVM service is disabled")
		return
	}

	vars := mux.Vars(r)
	vmID := vars["vmID"]

	ctx := r.Context()
	if err := api.kvmService.StopVM(ctx, vmID); err != nil {
		api.logger.WithError(err).WithField("vm_id", vmID).Error("Failed to stop VM")
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "VM stop initiated",
		"vm_id":   vmID,
	})
}

// getVMStats returns VM statistics
func (api *KVMAPI) getVMStats(w http.ResponseWriter, r *http.Request) {
	if !api.kvmService.IsEnabled() {
		api.writeError(w, http.StatusServiceUnavailable, "KVM service is disabled")
		return
	}

	vars := mux.Vars(r)
	vmID := vars["vmID"]

	ctx := r.Context()
	stats, err := api.kvmService.GetVMStats(ctx, vmID)
	if err != nil {
		api.logger.WithError(err).WithField("vm_id", vmID).Error("Failed to get VM stats")
		api.writeError(w, http.StatusNotFound, "VM not found")
		return
	}

	api.writeJSON(w, http.StatusOK, stats)
}

// getSystemResources returns system resource information
func (api *KVMAPI) getSystemResources(w http.ResponseWriter, r *http.Request) {
	if !api.kvmService.IsEnabled() {
		api.writeError(w, http.StatusServiceUnavailable, "KVM service is disabled")
		return
	}

	ctx := r.Context()
	resources, err := api.kvmService.GetSystemResources(ctx)
	if err != nil {
		api.logger.WithError(err).Error("Failed to get system resources")
		api.writeError(w, http.StatusInternalServerError, "Failed to get system resources")
		return
	}

	api.writeJSON(w, http.StatusOK, resources)
}

// getStatus returns KVM service status
func (api *KVMAPI) getStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"enabled": api.kvmService.IsEnabled(),
		"service": "kvm",
	}

	if api.kvmService.IsEnabled() {
		ctx := r.Context()
		vms, err := api.kvmService.ListVMs(ctx)
		if err == nil {
			status["vm_count"] = len(vms)

			runningCount := 0
			for _, vm := range vms {
				if vm.Status == kvm.VMStatusRunning {
					runningCount++
				}
			}
			status["running_vms"] = runningCount
		}
	}

	api.writeJSON(w, http.StatusOK, status)
}

// Helper methods

func (api *KVMAPI) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		api.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (api *KVMAPI) writeError(w http.ResponseWriter, statusCode int, message string) {
	api.writeJSON(w, statusCode, map[string]interface{}{
		"error":  message,
		"status": statusCode,
	})
}
