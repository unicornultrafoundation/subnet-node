package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/kvm"
)

// KVMAPI provides HTTP API endpoints for KVM service
type KVMAPI struct {
	kvmService kvm.KVMService
	logger     *logrus.Logger
}

// NewKVMAPI creates a new KVM API instance
func NewKVMAPI(kvmService kvm.KVMService) *KVMAPI {
	return &KVMAPI{
		kvmService: kvmService,
		logger:     logrus.WithField("api", "kvm").Logger,
	}
}

// RegisterRoutes registers KVM API routes
func (api *KVMAPI) RegisterRoutes(router *mux.Router) {
	kvmRouter := router.PathPrefix("/kvm").Subrouter()

	// VM management endpoints
	kvmRouter.HandleFunc("/vms", api.CreateVM).Methods("POST")
	kvmRouter.HandleFunc("/vms", api.ListVMs).Methods("GET")
	kvmRouter.HandleFunc("/vms/{id}", api.GetVM).Methods("GET")
	kvmRouter.HandleFunc("/vms/{id}", api.DeleteVM).Methods("DELETE")
	kvmRouter.HandleFunc("/vms/{id}/start", api.StartVM).Methods("POST")
	kvmRouter.HandleFunc("/vms/{id}/stop", api.StopVM).Methods("POST")
	kvmRouter.HandleFunc("/vms/{id}/restart", api.RestartVM).Methods("POST")

	// Health check
	kvmRouter.HandleFunc("/health", api.HealthCheck).Methods("GET")
}

// CreateVM handles POST /kvm/vms
func (api *KVMAPI) CreateVM(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req kvm.CreateVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Create VM
	vmDetails, err := api.kvmService.CreateVM(ctx, &req)
	if err != nil {
		if kvmErr, ok := err.(*kvm.KVMError); ok {
			switch kvmErr.Code {
			case kvm.ErrCodeValidationError:
				api.writeError(w, http.StatusBadRequest, kvmErr.Message, err)
			case kvm.ErrCodeResourceInsufficient:
				api.writeError(w, http.StatusConflict, kvmErr.Message, err)
			case kvm.ErrCodeServiceUnavailable:
				api.writeError(w, http.StatusServiceUnavailable, kvmErr.Message, err)
			default:
				api.writeError(w, http.StatusInternalServerError, kvmErr.Message, err)
			}
		} else {
			api.writeError(w, http.StatusInternalServerError, "Failed to create VM", err)
		}
		return
	}

	api.writeJSON(w, http.StatusCreated, vmDetails)
}

// ListVMs handles GET /kvm/vms
func (api *KVMAPI) ListVMs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters for filters
	filters := make(map[string]string)
	query := r.URL.Query()

	for key, values := range query {
		if len(values) > 0 {
			filters[key] = values[0]
		}
	}

	// List VMs
	vms, err := api.kvmService.ListVMs(ctx, filters)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "Failed to list VMs", err)
		return
	}

	// Handle empty result
	if vms == nil {
		vms = []*kvm.VMDetails{}
	}

	api.writeJSON(w, http.StatusOK, map[string]interface{}{
		"vms":   vms,
		"count": len(vms),
	})
}

// GetVM handles GET /kvm/vms/{id}
func (api *KVMAPI) GetVM(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	vmID := vars["id"]

	if vmID == "" {
		api.writeError(w, http.StatusBadRequest, "VM ID is required", nil)
		return
	}

	// Get VM
	vmDetails, err := api.kvmService.GetVM(ctx, vmID)
	if err != nil {
		if kvmErr, ok := err.(*kvm.KVMError); ok && kvmErr.Code == kvm.ErrCodeVMNotFound {
			api.writeError(w, http.StatusNotFound, kvmErr.Message, err)
		} else {
			api.writeError(w, http.StatusInternalServerError, "Failed to get VM", err)
		}
		return
	}

	api.writeJSON(w, http.StatusOK, vmDetails)
}

// DeleteVM handles DELETE /kvm/vms/{id}
func (api *KVMAPI) DeleteVM(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	vmID := vars["id"]

	if vmID == "" {
		api.writeError(w, http.StatusBadRequest, "VM ID is required", nil)
		return
	}

	// Delete VM
	err := api.kvmService.DeleteVM(ctx, vmID)
	if err != nil {
		if kvmErr, ok := err.(*kvm.KVMError); ok && kvmErr.Code == kvm.ErrCodeVMNotFound {
			api.writeError(w, http.StatusNotFound, kvmErr.Message, err)
		} else {
			api.writeError(w, http.StatusInternalServerError, "Failed to delete VM", err)
		}
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]string{
		"message": "VM deleted successfully",
		"vm_id":   vmID,
	})
}

// StartVM handles POST /kvm/vms/{id}/start
func (api *KVMAPI) StartVM(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	vmID := vars["id"]

	if vmID == "" {
		api.writeError(w, http.StatusBadRequest, "VM ID is required", nil)
		return
	}

	// Start VM
	err := api.kvmService.StartVM(ctx, vmID)
	if err != nil {
		if kvmErr, ok := err.(*kvm.KVMError); ok && kvmErr.Code == kvm.ErrCodeVMNotFound {
			api.writeError(w, http.StatusNotFound, kvmErr.Message, err)
		} else {
			api.writeError(w, http.StatusInternalServerError, "Failed to start VM", err)
		}
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]string{
		"message": "VM started successfully",
		"vm_id":   vmID,
	})
}

// StopVM handles POST /kvm/vms/{id}/stop
func (api *KVMAPI) StopVM(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	vmID := vars["id"]

	if vmID == "" {
		api.writeError(w, http.StatusBadRequest, "VM ID is required", nil)
		return
	}

	// Stop VM
	err := api.kvmService.StopVM(ctx, vmID)
	if err != nil {
		if kvmErr, ok := err.(*kvm.KVMError); ok && kvmErr.Code == kvm.ErrCodeVMNotFound {
			api.writeError(w, http.StatusNotFound, kvmErr.Message, err)
		} else {
			api.writeError(w, http.StatusInternalServerError, "Failed to stop VM", err)
		}
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]string{
		"message": "VM stopped successfully",
		"vm_id":   vmID,
	})
}

// RestartVM handles POST /kvm/vms/{id}/restart
func (api *KVMAPI) RestartVM(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	vmID := vars["id"]

	if vmID == "" {
		api.writeError(w, http.StatusBadRequest, "VM ID is required", nil)
		return
	}

	// Restart VM
	err := api.kvmService.RestartVM(ctx, vmID)
	if err != nil {
		if kvmErr, ok := err.(*kvm.KVMError); ok && kvmErr.Code == kvm.ErrCodeVMNotFound {
			api.writeError(w, http.StatusNotFound, kvmErr.Message, err)
		} else {
			api.writeError(w, http.StatusInternalServerError, "Failed to restart VM", err)
		}
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]string{
		"message": "VM restarted successfully",
		"vm_id":   vmID,
	})
}

// HealthCheck handles GET /kvm/health
func (api *KVMAPI) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check KVM service health
	err := api.kvmService.IsHealthy(ctx)
	if err != nil {
		api.writeError(w, http.StatusServiceUnavailable, "KVM service is unhealthy", err)
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "kvm",
	})
}

// Helper methods

func (api *KVMAPI) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		api.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (api *KVMAPI) writeError(w http.ResponseWriter, status int, message string, err error) {
	api.logger.WithFields(logrus.Fields{
		"status":  status,
		"message": message,
		"error":   err,
	}).Error("API error")

	errorResponse := map[string]interface{}{
		"error":  message,
		"status": status,
	}

	if err != nil {
		errorResponse["details"] = err.Error()
	}

	api.writeJSON(w, status, errorResponse)
}

// Additional utility endpoints

// GetVMTemplates handles GET /kvm/templates
func (api *KVMAPI) GetVMTemplates(w http.ResponseWriter, r *http.Request) {
	// This would require access to the storage manager
	// For now, return a placeholder response
	templates := []map[string]interface{}{
		{
			"name":         "ubuntu-20.04",
			"os":           "Ubuntu",
			"version":      "20.04",
			"architecture": "x86_64",
			"description":  "Ubuntu 20.04 LTS Server",
		},
		{
			"name":         "centos-8",
			"os":           "CentOS",
			"version":      "8",
			"architecture": "x86_64",
			"description":  "CentOS 8 Server",
		},
	}

	api.writeJSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// GetSystemResources handles GET /kvm/resources
func (api *KVMAPI) GetSystemResources(w http.ResponseWriter, r *http.Request) {
	// This would require access to the resource checker
	// For now, return a placeholder response
	resources := map[string]interface{}{
		"cpu": map[string]interface{}{
			"total":     8,
			"available": 6,
			"used":      2,
		},
		"memory": map[string]interface{}{
			"total":     "16GB",
			"available": "12GB",
			"used":      "4GB",
		},
		"storage": map[string]interface{}{
			"total":     "1TB",
			"available": "800GB",
			"used":      "200GB",
		},
		"vms": map[string]interface{}{
			"total":   2,
			"running": 1,
			"stopped": 1,
			"max":     10,
		},
	}

	api.writeJSON(w, http.StatusOK, resources)
}
