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
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/hardware_detector"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/ssh_connection"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
	"github.com/unicornultrafoundation/subnet-node/internal/api/ws"
)

// WebSocket message codes for VM SSH
const (
	VMSSHCodeStdin   = 0
	VMSSHCodeStdout  = 1
	VMSSHCodeStderr  = 2
	VMSSHCodeResult  = 3
	VMSSHCodeFailure = 4
	VMSSHCodeResize  = 5
)

// VMSSHRequest represents the request body for SSH connection
type VMSSHRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// VMSSHResponse represents the response from SSH connection
type VMSSHResponse struct {
	ExitCode int    `json:"exit_code"`
	Message  string `json:"message,omitempty"`
}

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

func convertToSystemInfoResult(info *vbtypes.VMSystemInfo) *vmSystemInfoResult {
	if info == nil {
		return nil
	}
	return &vmSystemInfoResult{
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
	vboxService *virtualbox.VirtualboxService
	cfg         ConfigProvider
	ordersCache *OrdersWithCache
}

// NewVirtualBoxAPI creates a new instance of VirtualBoxAPI.
func NewVirtualBoxAPI(vboxService *virtualbox.VirtualboxService, cfg ConfigProvider, bidMarket BidMarketContract) *VirtualBoxAPI {

	return &VirtualBoxAPI{vboxService: vboxService, cfg: cfg, ordersCache: NewOrdersWithCache(bidMarket)}
}

func (api *VirtualBoxAPI) GetVMs(ctx context.Context) ([]vmResult, error) {
	vms, _, err := api.vboxService.GetVMs(ctx)
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

	order, err := api.ordersCache.GetOrder(r.Context(), req.OrderId)
	if err != nil {
		api.sendErrorResponse(w, "Order not found", http.StatusNotFound)
		return
	}

	// check current hardware information, if ARM architecture return 'Ubuntu_ARM64'
	// if x86 architecture return 'Ubuntu_x86_64'
	hardware, err := hardware_detector.DetectHardware()
	if err != nil {
		api.sendErrorResponse(w, "Failed to get hardware information", http.StatusInternalServerError)
		return
	}

	// Use the determined Ubuntu OS type from hardware info
	ubuntuOSType := determineUbuntuOSType(hardware.Architecture)

	vbReq := vbtypes.VMCreateRequest{
		Name:       "vm-" + order.ID.String(),
		CPUCores:   int(order.CpuCores.Int64()),
		MemoryMB:   int(order.MemoryMB.Int64()),
		DiskSizeGB: int(order.DiskGB.Int64()),
		OSType:     ubuntuOSType,
		Username:   req.Username,
		Password:   req.Password,
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

// getVMsHandler handles GET requests for VMs - gets all VMs or a specific VM by ID
func (api *VirtualBoxAPI) getVMsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if vmID is provided in the URL path
	vmID := chi.URLParam(r, "vmID")

	if vmID != "" {
		// Get specific VM
		vm, err := api.GetVM(r.Context(), vmID)
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
	vms, err := api.GetVMs(r.Context())
	if err != nil {
		api.sendErrorResponse(w, fmt.Sprintf("Failed to get VMs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vms)
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

// Router returns the chi router with all VirtualBox routes
func (api *VirtualBoxAPI) Router() *chi.Mux {

	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	authMiddleware := NewAuthMiddleware(api.cfg, api.ordersCache)

	// WebSocket route for SSH connection
	r.Get("/{vmID}/ssh", api.vmSSHWebSocketHandler)

	// Virtualbox API Router
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware())

		// Virtualbox API routes
		r.Post("/", api.createVMHandler)
		r.Get("/", api.getVMsHandler)
		r.Get("/{vmID}", api.getVMsHandler)
		r.Delete("/{vmID}", api.deleteVMHandler)

		// startVm, stopVm, pauseVm, resumeVm, resetVm
		r.Post("/{vmID}/start", api.startVMHandler)
		r.Post("/{vmID}/stop", api.stopVMHandler)
		r.Post("/{vmID}/pause", api.pauseVMHandler)
		r.Post("/{vmID}/resume", api.resumeVMHandler)
		r.Post("/{vmID}/reset", api.resetVMHandler)

		// Job progress
		r.Get("/jobs", api.getJobHandler)
		r.Get("/jobs/{jobID}", api.getJobHandler)
	})

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

// determineUbuntuOSType determines the appropriate Ubuntu OS type for the current architecture
func determineUbuntuOSType(architecture string) string {
	switch architecture {
	case "arm64", "aarch64":
		return "Ubuntu_ARM64"
	case "amd64", "x86_64":
		return "Ubuntu_64"
	case "arm":
		return "Ubuntu"
	case "386", "i386":
		return "Ubuntu"
	default:
		return "Ubuntu_64" // Default fallback
	}
}
