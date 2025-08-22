package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
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

// startVMHandler handles POST requests to start a VM
func (api *VirtualBoxAPI) startVMHandler(w http.ResponseWriter, r *http.Request) {
	orderId := chi.URLParam(r, "orderId")
	if orderId == "" {
		api.sendErrorResponse(w, "orderId is required", http.StatusBadRequest)
		return
	}

	// Get vmId from orderId
	vmId, exists := api.vboxService.GetVMIdByOrderId(orderId)
	if !exists {
		api.sendErrorResponse(w, "VM not found for this orderId", http.StatusNotFound)
		return
	}

	vm, err := api.vboxService.StartVM(r.Context(), vmId)
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
	orderId := chi.URLParam(r, "orderId")
	if orderId == "" {
		api.sendErrorResponse(w, "orderId is required", http.StatusBadRequest)
		return
	}

	// Get vmId from orderId
	vmId, exists := api.vboxService.GetVMIdByOrderId(orderId)
	if !exists {
		api.sendErrorResponse(w, "VM not found for this orderId", http.StatusNotFound)
		return
	}

	vm, err := api.vboxService.StopVM(r.Context(), vmId)
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
	orderId := chi.URLParam(r, "orderId")
	if orderId == "" {
		api.sendErrorResponse(w, "orderId is required", http.StatusBadRequest)
		return
	}

	// Get vmId from orderId
	vmId, exists := api.vboxService.GetVMIdByOrderId(orderId)
	if !exists {
		api.sendErrorResponse(w, "VM not found for this orderId", http.StatusNotFound)
		return
	}

	vm, err := api.vboxService.PauseVM(r.Context(), vmId)
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
	orderId := chi.URLParam(r, "orderId")
	if orderId == "" {
		api.sendErrorResponse(w, "orderId is required", http.StatusBadRequest)
		return
	}

	// Get vmId from orderId
	vmId, exists := api.vboxService.GetVMIdByOrderId(orderId)
	if !exists {
		api.sendErrorResponse(w, "VM not found for this orderId", http.StatusNotFound)
		return
	}

	vm, err := api.vboxService.ResumeVM(r.Context(), vmId)
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
	orderId := chi.URLParam(r, "orderId")
	if orderId == "" {
		api.sendErrorResponse(w, "orderId is required", http.StatusBadRequest)
		return
	}

	// Get vmId from orderId
	vmId, exists := api.vboxService.GetVMIdByOrderId(orderId)
	if !exists {
		api.sendErrorResponse(w, "VM not found for this orderId", http.StatusNotFound)
		return
	}

	vm, err := api.vboxService.ResetVM(r.Context(), vmId)
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
	orderId := chi.URLParam(r, "orderId")
	if orderId == "" {
		api.sendErrorResponse(w, "orderId is required", http.StatusBadRequest)
		return
	}

	// Get vmId from orderId
	vmId, exists := api.vboxService.GetVMIdByOrderId(orderId)
	if !exists {
		api.sendErrorResponse(w, "VM not found for this orderId", http.StatusNotFound)
		return
	}

	// username, password from body
	var req vbtypes.GenerateSSHTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := api.vboxService.GenerateSSHToken(r.Context(), vmId, req.Username, req.Password)
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to generate SSH token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}

func (api *VirtualBoxAPI) WebSocketRouter() *chi.Mux {
	r := chi.NewRouter()

	authMiddleware := NewAuthMiddleware(api.cfg, api.ordersCache)
	r.Use(authMiddleware.Middleware())

	r.Get("/{orderId}/ssh", api.vmSSHWebSocketHandler)
	r.Get("/{orderId}/metrics", api.getVMMetricsHandler)

	return r
}

// Router returns the chi router with all VirtualBox routes
func (api *VirtualBoxAPI) Router() *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	authMiddleware := NewAuthMiddleware(api.cfg, api.ordersCache)
	r.Use(authMiddleware.Middleware())

	// Virtualbox API routes
	r.Post("/", api.createVMFromImage)
	r.Get("/{orderId}", api.getVMHandler) // Get specific VM by orderId
	r.Delete("/{orderId}", api.deleteVMHandler)

	// startVm, stopVm, pauseVm, resumeVm, resetVm
	r.Post("/{orderId}/start", api.startVMHandler)
	r.Post("/{orderId}/stop", api.stopVMHandler)
	r.Post("/{orderId}/pause", api.pauseVMHandler)
	r.Post("/{orderId}/resume", api.resumeVMHandler)
	r.Post("/{orderId}/reset", api.resetVMHandler)

	// generateSSHToken
	r.Post("/{orderId}/ssh/token", api.generateSSHTokenHandler)

	// Job progress
	r.Get("/jobs", api.getJobHandler)
	r.Get("/jobs/{jobID}", api.getJobHandler)

	return r
}

// vmSSHWebSocketHandler handles WebSocket connections for VM SSH
func (api *VirtualBoxAPI) vmSSHWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	orderId := chi.URLParam(r, "orderId")

	if orderId == "" {
		api.sendErrorResponse(w, "orderId is required", http.StatusBadRequest)
		return
	}

	// Get vmId from orderId
	vmId, exists := api.vboxService.GetVMIdByOrderId(orderId)
	if !exists {
		api.sendErrorResponse(w, "VM not found for this orderId", http.StatusNotFound)
		return
	}

	// Create context BEFORE any WebSocket operations
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Get VM to verify it exists and is running
	vm, err := api.vboxService.GetVM(ctx, vmId)
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

	logger := logrus.WithField("orderId", orderId).WithField("vmId", vmId)

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
	if vmPayload.VMID != vmId {
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
	sshConn := ssh_connection.NewSSHConnection(vmId, conn, sshServer)

	// Verify VM exists in SSHServer configs
	_, exists = sshServer.GetVMConfig(vmId)

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

func (api *VirtualBoxAPI) getVMMetricsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse period from query parameter, default to 1 if not provided
	periodStr := r.URL.Query().Get("period")
	period := 1 // default value
	if periodStr != "" {
		if parsedPeriod, err := strconv.Atoi(periodStr); err == nil && parsedPeriod > 0 {
			// Validate period is within reasonable bounds (1-60 seconds)
			if parsedPeriod > 60 {
				api.sendErrorResponse(w, "Period must be between 1 and 60 seconds", http.StatusBadRequest)
				return
			}
			period = parsedPeriod
		} else {
			api.sendErrorResponse(w, "Invalid period parameter. Must be a positive integer", http.StatusBadRequest)
			return
		}
	}

	logger := logrus.WithField("orderId", orderId).WithField("vmId", vmId).WithField("period", period)

	conn, err := ws.SetupWebSocketForMetrics(w, r, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}

	defer conn.Close()

	// Check if VM is running after WebSocket connection is established
	vm, err := api.vboxService.GetVM(r.Context(), vmId)
	if err != nil {
		errorData := map[string]string{
			"type":    "error",
			"message": fmt.Sprintf("Failed to get VM: %v", err),
		}
		jsonData, _ := json.Marshal(errorData)
		conn.WriteMessage(websocket.TextMessage, jsonData)
		return
	}

	if vm.Status != vbtypes.Running {
		errorData := map[string]string{
			"type":    "error",
			"message": fmt.Sprintf("VM is not running. Status: %s", vm.Status),
		}
		jsonData, _ := json.Marshal(errorData)
		conn.WriteMessage(websocket.TextMessage, jsonData)
		return
	}

	err = api.vboxService.CollectMetrics(r.Context(), vmId, conn, period)
	if err != nil {
		logger.WithError(err).Error("Failed to collect metrics")
		return
	}
}

func (api *VirtualBoxAPI) createVMFromImage(w http.ResponseWriter, r *http.Request) {

	var req vbtypes.CreateVMFromImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Println("createVMFromImage", req)

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
