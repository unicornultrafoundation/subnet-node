package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/ssh_connection"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
	"github.com/unicornultrafoundation/subnet-node/internal/api/ws"
)

type vmResult struct {
	ID         string           `json:"id,omitempty"`
	Name       string           `json:"name,omitempty"`
	Status     vbtypes.VMStatus `json:"status,omitempty"`
	CPUCores   int              `json:"cpu_cores,omitempty"`
	MemoryMB   int              `json:"memory_mb,omitempty"`
	DiskSizeGB int              `json:"disk_size_gb,omitempty"`
	IPAddress  string           `json:"ip_address,omitempty"`
	SSHPort    int              `json:"ssh_port,omitempty"`
	VMFolder   string           `json:"vm_folder,omitempty"`
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
		IPAddress:  vm.IPAddress,
		SSHPort:    vm.SSHPort,
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

type VirtualBoxAPI struct {
	vboxService *virtualbox.VirtualboxService
	cfg         ConfigProvider
	ordersCache *OrdersWithCache
}

// NewVirtualBoxAPI creates a new instance of VirtualBoxAPI.
func NewVirtualBoxAPI(vboxService *virtualbox.VirtualboxService, cfg ConfigProvider, bidMarket BidMarketContract) *VirtualBoxAPI {

	return &VirtualBoxAPI{vboxService: vboxService, cfg: cfg, ordersCache: NewOrdersWithCache(bidMarket)}
}

// corsMiddleware adds CORS headers to allow frontend testing
func (api *VirtualBoxAPI) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins for development/testing
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (api *VirtualBoxAPI) CreateTemplateVM(ctx context.Context, name string, cpuCores int, memoryMB int, diskSizeGB int, osType string, username string, password string) (*vbtypes.JobCreateResponse, error) {
	vbReq := vbtypes.VMCreateRequest{
		Name:       name,
		CPUCores:   cpuCores,
		MemoryMB:   memoryMB,
		DiskSizeGB: diskSizeGB,
		OSType:     osType, // Pass the OS type to the service
		Username:   username,
		Password:   password,
	}

	return api.vboxService.CreateTemplateVM(ctx, vbReq)
}

func (api *VirtualBoxAPI) createVMHandler(w http.ResponseWriter, r *http.Request) {

	var req vbtypes.CreateVMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := api.ordersCache.GetOrder(r.Context(), req.OrderId)
	if err != nil {
		api.sendErrorResponse(w, "Order not found", http.StatusNotFound)
		return
	}

	// // check current hardware information, if ARM architecture return 'Ubuntu_ARM64'
	// // if x86 architecture return 'Ubuntu_x86_64'
	// hardware, err := hardware_detector.DetectHardware()
	// if err != nil {
	// 	api.sendErrorResponse(w, "Failed to get hardware information", http.StatusInternalServerError)
	// 	return
	// }

	// vbReq := vbtypes.VMCreateRequest{
	// 	Name:       "vm-" + order.ID.String(),
	// 	CPUCores:   int(order.CpuCores.Int64()),
	// 	MemoryMB:   int(order.MemoryMB.Int64()),
	// 	DiskSizeGB: int(order.DiskGB.Int64()),
	// 	OSType:     ubuntuOSType,
	// 	Username:   req.Username,
	// 	Password:   req.Password,
	// }

	// // Create the VM using the determined OS type
	// jobResponse, err := api.vboxService.CreateVM(r.Context(), vbReq)
	// if err != nil {
	// 	api.sendErrorResponse(w, "Failed to create VM", http.StatusInternalServerError)
	// 	return
	// }

	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(jobResponse)

}

// getVMsHandler handles GET requests for VMs - gets all VMs or a specific VM by ID
func (api *VirtualBoxAPI) getVMsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if vmID is provided in the URL path
	vmID := chi.URLParam(r, "vmID")

	if vmID != "" {
		// Get specific VM
		vm, err := api.vboxService.GetVM(r.Context(), vmID)
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
	vms, total, err := api.vboxService.GetVMs(r.Context())

	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to get VMs: %v", err), http.StatusInternalServerError)
		return
	}

	result := make([]vmResult, total)
	for i, vm := range vms {
		result[i] = *convertToVMResult(vm)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// deleteVMHandler handles DELETE requests for a specific VM by ID
func (api *VirtualBoxAPI) deleteVMHandler(w http.ResponseWriter, r *http.Request) {
	vmID := chi.URLParam(r, "vmID")
	if vmID == "" {
		api.sendErrorResponse(w, "vmID is required", http.StatusBadRequest)
		return
	}

	err := api.vboxService.DeleteVM(r.Context(), vmID)
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to delete VM: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "VM deleted successfully"})
}

// startVMHandler handles POST requests to start a VM
func (api *VirtualBoxAPI) startVMHandler(w http.ResponseWriter, r *http.Request) {
	vmID := chi.URLParam(r, "vmID")
	if vmID == "" {
		api.sendErrorResponse(w, "vmID is required", http.StatusBadRequest)
		return
	}

	vm, err := api.vboxService.StartVM(r.Context(), vmID)
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to start VM: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(convertToVMResult(vm))
}

// stopVMHandler handles POST requests to stop a VM
func (api *VirtualBoxAPI) stopVMHandler(w http.ResponseWriter, r *http.Request) {
	vmID := chi.URLParam(r, "vmID")
	if vmID == "" {
		api.sendErrorResponse(w, "vmID is required", http.StatusBadRequest)
		return
	}

	vm, err := api.vboxService.StopVM(r.Context(), vmID)
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to stop VM: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(convertToVMResult(vm))
}

// pauseVMHandler handles POST requests to pause a VM
func (api *VirtualBoxAPI) pauseVMHandler(w http.ResponseWriter, r *http.Request) {
	vmID := chi.URLParam(r, "vmID")
	if vmID == "" {
		api.sendErrorResponse(w, "vmID is required", http.StatusBadRequest)
		return
	}

	vm, err := api.vboxService.PauseVM(r.Context(), vmID)
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to pause VM: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(convertToVMResult(vm))
}

// resumeVMHandler handles POST requests to resume a VM
func (api *VirtualBoxAPI) resumeVMHandler(w http.ResponseWriter, r *http.Request) {
	vmID := chi.URLParam(r, "vmID")
	if vmID == "" {
		api.sendErrorResponse(w, "vmID is required", http.StatusBadRequest)
		return
	}

	vm, err := api.vboxService.ResumeVM(r.Context(), vmID)
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to resume VM: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(convertToVMResult(vm))
}

// resetVMHandler handles POST requests to reset a VM
func (api *VirtualBoxAPI) resetVMHandler(w http.ResponseWriter, r *http.Request) {
	vmID := chi.URLParam(r, "vmID")
	if vmID == "" {
		api.sendErrorResponse(w, "vmID is required", http.StatusBadRequest)
		return
	}

	vm, err := api.vboxService.ResetVM(r.Context(), vmID)
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to reset VM: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(convertToVMResult(vm))
}

// getJobHandler handles GET requests to get the progress of a job
func (api *VirtualBoxAPI) getJobHandler(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	if jobID != "" {

		job, err := api.vboxService.GetJobProgress(r.Context(), jobID)
		if err != nil {
			api.sendErrorResponse(w, fmt.Sprintf("Failed to get job progress: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(job)
		return
	}

	// get all jobs
	jobs, err := api.vboxService.ListJobs(r.Context())
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to get jobs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobs)
}

// generateSSHTokenHandler handles POST requests to generate an SSH token for a VM
func (api *VirtualBoxAPI) generateSSHTokenHandler(w http.ResponseWriter, r *http.Request) {
	vmID := chi.URLParam(r, "vmID")
	if vmID == "" {
		api.sendErrorResponse(w, "vmID is required", http.StatusBadRequest)
		return
	}

	// username, password from body
	var req vbtypes.GenerateSSHTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := api.vboxService.GenerateSSHToken(r.Context(), vmID, req.Username, req.Password)
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to generate SSH token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}

// Router returns the chi router with all VirtualBox routes
func (api *VirtualBoxAPI) Router() *chi.Mux {

	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Add CORS middleware for frontend testing
	r.Use(api.corsMiddleware)

	authMiddleware := NewAuthMiddleware(api.cfg, api.ordersCache)

	// WebSocket route for SSH connection
	r.Get("/{vmID}/ssh", api.vmSSHWebSocketHandler)

	// Virtualbox API routes
	r.With(authMiddleware.Middleware()).Post("/", api.createVMHandler)
	r.Post("/image", api.createVMFromImage)
	r.Get("/", api.getVMsHandler)
	r.Get("/{vmID}", api.getVMsHandler)
	r.Delete("/{vmID}", api.deleteVMHandler)

	// startVm, stopVm, pauseVm, resumeVm, resetVm
	r.Post("/{vmID}/start", api.startVMHandler)
	r.Post("/{vmID}/stop", api.stopVMHandler)
	r.Post("/{vmID}/pause", api.pauseVMHandler)
	r.Post("/{vmID}/resume", api.resumeVMHandler)
	r.Post("/{vmID}/reset", api.resetVMHandler)

	// generateSSHToken
	r.Post("/{vmID}/ssh/token", api.generateSSHTokenHandler)

	// Job progress
	r.Get("/jobs", api.getJobHandler)
	r.Get("/jobs/{jobID}", api.getJobHandler)

	return r
}

// vmSSHWebSocketHandler handles WebSocket connections for VM SSH
func (api *VirtualBoxAPI) vmSSHWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	vmID := chi.URLParam(r, "vmID")

	if vmID == "" {
		api.sendErrorResponse(w, "vmID is required", http.StatusBadRequest)
		return
	}

	// Create context BEFORE any WebSocket operations
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Get VM to verify it exists and is running
	vm, err := api.vboxService.GetVM(ctx, vmID)
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("VM not found: %v", err), http.StatusNotFound)
		return
	}

	if vm.Status != vbtypes.Running {
		api.sendErrorResponse(w, "VM is not running", http.StatusBadRequest)
		return
	}

	if vm.SSHPort == 0 {
		api.sendErrorResponse(w, "SSH port not configured for VM", http.StatusBadRequest)
		return
	}

	logger := logrus.WithField("vmID", vmID)

	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		logger.Error("Missing SSH access token")
		api.sendErrorResponse(w, "SSH access token is required as query parameter", http.StatusBadRequest)
		return
	}

	// Validate and consume the token
	vmPayload, err := api.vboxService.ValidateAndConsumeSSHToken(token)
	if err != nil {
		logger.WithError(err).Error("Invalid or expired SSH access token")
		api.sendErrorResponse(w, fmt.Sprintf("Invalid or expired token: %v", err), http.StatusUnauthorized)
		return
	}

	// Verify the token is for the correct VM
	if vmPayload.VMID != vmID {
		logger.Error("Token is for a different VM")
		api.sendErrorResponse(w, "Token is not valid for this VM", http.StatusUnauthorized)
		return
	}

	username := vmPayload.Username
	password := vmPayload.Password

	conn, err := ws.SetupWebSocketForVirtualBox(w, r, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}

	defer conn.Close()

	// Send a simple message
	err = conn.WriteJSON(map[string]string{"message": "WebSocket connected successfully"})
	if err != nil {
		logger.WithError(err).Error("Failed to send message")
		return
	}

	// Keep connection alive for a bit
	time.Sleep(5 * time.Second)

	// Create SSH connection
	sshServer := api.vboxService.GetSSHServer()
	sshConn := ssh_connection.NewSSHConnection(vmID, conn, sshServer)

	// Verify VM exists in SSHServer configs
	_, exists := sshServer.GetVMConfig(vmID)

	if !exists {
		logger.Error("VM not found in SSHServer configs")
		api.sendErrorResponse(w, "VM SSH configuration not found", http.StatusInternalServerError)
		return
	}

	// Connect to SSH
	if err := sshConn.Connect(username, password); err != nil {
		logger.WithError(err).Error("Failed to establish SSH connection")
		if err := ws.SendErrorResponse(conn, fmt.Sprintf("SSH connection failed: %v", err)); err != nil {
			logger.WithError(err).Error("Failed to send error response")
		}
		return
	}

	logger.Info("SSH connection established successfully")

	//  Start SSH connection handling
	sshConn.Start()

	// Wait for context cancellation (client disconnect or error)
	<-ctx.Done()
	logger.Info("SSH WebSocket connection closed")
}

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
	}

	// Create the VM using the determined OS type
	jobResponse, err := api.vboxService.CreateVMFromImage(r.Context(), vbReq)
	if err != nil {
		api.sendErrorResponse(w, "Failed to create VM", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jobResponse)
}

// sendErrorResponse sends a standardized error response
func (api *VirtualBoxAPI) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"message": http.StatusText(statusCode),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
