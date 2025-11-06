package virtualbox

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sirupsen/logrus"
	job_manager "github.com/unicornultrafoundation/subnet-node/core/virtualbox/job-manager"
	vbox_service "github.com/unicornultrafoundation/subnet-node/core/virtualbox/service"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/ssh_connection"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/storage"
	vbtypes "github.com/unicornultrafoundation/subnet-node/core/virtualbox/types"
	"github.com/unicornultrafoundation/subnet-node/core/virtualbox/util"
)

var serviceLog = logrus.WithField("service", "virtualbox")

type IVirtualboxService interface {
	Start(ctx context.Context) error
	CreateVM(ctx context.Context, req vbtypes.VMCreateFromImageRequest) (*vbtypes.JobCreateResponse, error)
	GetVM(ctx context.Context, vmID string) (*vbtypes.VM, error)
	GetVMs(ctx context.Context) ([]*vbtypes.VM, error)
	GetVMByOrderId(ctx context.Context, orderId string) (*vbtypes.VM, error)
	GetVMIdByOrderId(orderId string) (string, bool)
	DeleteVM(ctx context.Context, vmID string) error
	StartVM(ctx context.Context, vmID string) (*vbtypes.VM, error)
	UpdateVM(ctx context.Context, vmID string, req vbtypes.VMUpdateRequest) (*vbtypes.VM, error)
	StopVM(ctx context.Context, vmID string) error
	PauseVM(ctx context.Context, vmID string) error
	ResumeVM(ctx context.Context, vmID string) error
	ResetVM(ctx context.Context, vmID string) error
	CloneVM(ctx context.Context, baseVmName string, newVMName string, register bool) error

	GetVMPortForwarding(ctx context.Context, vmID string) ([]vbox_service.PFRule, error)

	TakeSnapshotVM(ctx context.Context, vmID string, snapshotName string) error
	RestoreSnapshot(ctx context.Context, vmID string, snapshotName string) error
	DeleteSnapshot(ctx context.Context, vmID string, snapshotName string) error
	ListSnapshots(ctx context.Context, vmID string) ([]string, error)

	GetJobProgress(ctx context.Context, jobID string) (*vbtypes.Job, error)
	ListJobs(ctx context.Context) ([]*vbtypes.Job, error)
	GenerateSSHToken(ctx context.Context, vmID string, username string, password string) (*vbtypes.SSHTokenResponse, error)
	ValidateAndConsumeSSHToken(token string) (*vbtypes.SSHAccessToken, error)
	AddNATPF(ctx context.Context, vmID string, adapterNumber int, portName string, proto string, guestPort uint16) error
	DeleteNATPF(ctx context.Context, vmID string, adapterNumber int, portName string) error
	SetNIC(ctx context.Context, vmID string, n int, network string, hardware string, hostInterface string, macAddr string) error
	GetSSHServer() *ssh_connection.SSHServer
	CollectMetrics(ctx context.Context, vmId string, conn *websocket.Conn, period int) error
}

type virtualboxService struct {
	vBoxService    *vbox_service.VBoxService
	jobManager     *job_manager.JobManager
	storageManager *storage.ImageStorage
	ds             datastore.Datastore

	sshServer *ssh_connection.SSHServer
	wsHandler *ssh_connection.WebSocketHandler

	orderToVMMap map[string]string
	orderMapMu   sync.RWMutex

	syncTicker *time.Ticker
	stopChan   chan struct{}
}

func NewService(ds datastore.Datastore, cfg VirtualBoxConfig) IVirtualboxService {
	storageManager := storage.NewImageStorage(cfg.BaseFolder, cfg.ImagesDir)
	vboxService := vbox_service.NewVBoxService()

	orderToVMMap := make(map[string]string)
	jobManager := job_manager.NewJobManager(storageManager, vboxService, &orderToVMMap, ds)

	sshServer := ssh_connection.NewSSHServer()
	wsHandler := ssh_connection.NewWebSocketHandler(sshServer)

	return &virtualboxService{
		vBoxService:    vboxService,
		jobManager:     jobManager,
		storageManager: storageManager,
		ds:             ds,
		orderToVMMap:   orderToVMMap,
		orderMapMu:     sync.RWMutex{},
		stopChan:       make(chan struct{}),
		sshServer:      sshServer,
		wsHandler:      wsHandler,
	}
}

func (s *virtualboxService) Start(ctx context.Context) error {
	if err := s.loadAllOrderMappingsFromDatastore(ctx); err != nil {
		serviceLog.Warnf("Failed to load order mappings from datastore: %v", err)
	}

	if err := s.syncSSHServerWithRunningVMs(ctx); err != nil {
		serviceLog.Warnf("Failed to sync SSHServer with running VMs: %v", err)
	}

	return nil
}

// CreateVM will last long time to create a VM,
// Add the request into the job manager and return the job id
// The job manager will handle the request and create the VM in the background
func (s *virtualboxService) CreateVM(ctx context.Context, req vbtypes.VMCreateFromImageRequest) (*vbtypes.JobCreateResponse, error) {

	// check if the VM already exists
	vm, _ := s.GetVMByOrderId(ctx, req.OrderId)

	if vm != nil {
		return nil, fmt.Errorf("VM already exists")
	}

	job, err := s.jobManager.CreateJob(ctx, vbtypes.VMEventCreateVM, req, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return &vbtypes.JobCreateResponse{
		JobID: job.ID,
	}, nil
}

func (s *virtualboxService) GetVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {

	vm, err := s.vBoxService.GetVM(vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	return vm, nil
}

func (s *virtualboxService) GetVMs(ctx context.Context) ([]*vbtypes.VM, error) {
	vms, err := s.vBoxService.ListVMs()
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs: %w", err)
	}
	return vms, nil
}

func (s *virtualboxService) loadAllOrderMappingsFromDatastore(ctx context.Context) error {
	q := query.Query{Prefix: "virtualbox/order_mapping/"}
	results, err := s.ds.Query(ctx, q)
	if err != nil {
		return err
	}
	defer results.Close()

	for result := range results.Next() {
		if result.Error != nil {
			continue
		}
		var mapping map[string]string
		if err := json.Unmarshal(result.Value, &mapping); err != nil {
			continue
		}
		orderId := mapping["orderId"]
		vmId := mapping["vmId"]
		if orderId != "" && vmId != "" {
			s.orderToVMMap[orderId] = vmId
		}
	}

	serviceLog.Infof("Loaded %d order mappings from datastore", len(s.orderToVMMap))
	return nil
}

func (s *virtualboxService) GetVMByOrderId(ctx context.Context, orderId string) (*vbtypes.VM, error) {

	vmId, exists := s.orderToVMMap[orderId]
	if !exists {
		return nil, fmt.Errorf("VM not found for this orderId")
	}
	return s.GetVM(ctx, vmId)
}

func (s *virtualboxService) GetVMIdByOrderId(orderId string) (string, bool) {
	s.orderMapMu.RLock()
	defer s.orderMapMu.RUnlock()
	vmId, exists := s.orderToVMMap[orderId]
	return vmId, exists
}

func (s *virtualboxService) DeleteVM(ctx context.Context, vmID string) error {
	err := s.vBoxService.DeleteVM(vmID)
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}
	return nil
}

func (s *virtualboxService) StartVM(ctx context.Context, vmID string) (*vbtypes.VM, error) {

	err := s.vBoxService.StartVM(vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to start VM: %w", err)
	}
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}
	return vm, nil
}

func (s *virtualboxService) UpdateVM(ctx context.Context, vmID string, req vbtypes.VMUpdateRequest) (*vbtypes.VM, error) {
	err := s.vBoxService.UpdateVM(vmID, req.CPUCores, req.MemoryMB, req.DiskSizeGB)
	if err != nil {
		return nil, fmt.Errorf("failed to update VM: %w", err)
	}
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}
	return vm, nil
}

func (s *virtualboxService) StopVM(ctx context.Context, vmID string) error {
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}

	if vm.Status != vbtypes.Running {
		return fmt.Errorf("VM is not running, current status: %s", vm.Status)
	}

	return s.vBoxService.StopVM(vmID)
}

func (s *virtualboxService) PauseVM(ctx context.Context, vmID string) error {
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}

	if vm.Status != vbtypes.Running {
		return fmt.Errorf("VM is not running, current status: %s", vm.Status)
	}

	return s.vBoxService.PauseVM(vmID)
}

func (s *virtualboxService) ResumeVM(ctx context.Context, vmID string) error {

	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}

	if vm.Status != vbtypes.Paused {
		return fmt.Errorf("VM is not paused, current status: %s", vm.Status)
	}

	return s.vBoxService.ResumeVM(vmID)
}

func (s *virtualboxService) ResetVM(ctx context.Context, vmID string) error {

	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}

	if vm.Status != vbtypes.Running {
		return fmt.Errorf("VM is not running, current status: %s", vm.Status)
	}

	return s.vBoxService.ResetVM(vmID)
}

func (s *virtualboxService) GetJobProgress(ctx context.Context, jobID string) (*vbtypes.Job, error) {
	return s.jobManager.GetJobProgress(ctx, jobID)
}

func (s *virtualboxService) ListJobs(ctx context.Context) ([]*vbtypes.Job, error) {
	return s.jobManager.ListJobs(ctx)
}

func (s *virtualboxService) AddNATPF(ctx context.Context, vmID string, adapterNumber int, portName string, proto string, guestPort uint16) error {

	var protoValue vbox_service.PFProto
	if proto != "" {
		protoValue = vbox_service.PFProto(proto)
	} else {
		protoValue = vbox_service.PFTCP
	}
	// find available TCP port in range 20000-30000
	availablePort, err := util.FindAvailableTCPPort(20000, 30000)
	if err != nil {
		return fmt.Errorf("failed to get available port: %w", err)
	}

	rule := vbox_service.PFRule{
		PortName:  portName,
		Proto:     protoValue,
		HostIP:    nil,
		GuestIP:   nil,
		HostPort:  uint16(availablePort),
		GuestPort: guestPort,
	}
	return s.vBoxService.AddNATPF(adapterNumber, vmID, rule)
}

func (s *virtualboxService) DeleteNATPF(ctx context.Context, vmID string, adapterNumber int, portName string) error {
	return s.vBoxService.DelNATPF(adapterNumber, vmID, portName)
}

func (s *virtualboxService) SetNIC(ctx context.Context, vmID string, n int, network string, hardware string, hostInterface string, macAddr string) error {
	return s.vBoxService.SetNIC(vmID, n, vbox_service.NIC{
		Network:       vbox_service.NICNetwork(network),
		Hardware:      vbox_service.NICHardware(hardware),
		HostInterface: hostInterface,
		MacAddr:       macAddr,
	})
}

func (s *virtualboxService) CloneVM(ctx context.Context, baseVmName string, newVMName string, register bool) error {
	return s.vBoxService.CloneVM(baseVmName, newVMName, register)
}

func (s *virtualboxService) TakeSnapshotVM(ctx context.Context, vmID string, snapshotName string) error {
	return s.vBoxService.TakeSnapshotVM(vmID, snapshotName)
}

func (s *virtualboxService) RestoreSnapshot(ctx context.Context, vmID string, snapshotName string) error {
	return s.vBoxService.RestoreSnapshot(vmID, snapshotName)
}

func (s *virtualboxService) DeleteSnapshot(ctx context.Context, vmID string, snapshotName string) error {
	return s.vBoxService.DeleteSnapshot(vmID, snapshotName)
}

func (s *virtualboxService) ListSnapshots(ctx context.Context, vmID string) ([]string, error) {
	return s.vBoxService.ListSnapshots(vmID)
}

func (s *virtualboxService) GetVMPortForwarding(ctx context.Context, vmID string) ([]vbox_service.PFRule, error) {
	return s.vBoxService.GetVMPortForwarding(vmID)
}

func (s *virtualboxService) syncSSHServerWithRunningVMs(ctx context.Context) error {
	serviceLog.Info("Syncing SSHServer with running VMs...")

	vms, err := s.GetVMs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get VMs for SSHServer sync: %w", err)
	}

	for _, vm := range vms {
		if vm.Status == vbtypes.Running && vm.SSHPort > 0 {
			s.sshServer.AddVMConfig(vm.ID, "127.0.0.1", strconv.Itoa(vm.SSHPort))
			serviceLog.Infof("Synced running VM %s with SSHServer: localhost:%d", vm.ID, vm.SSHPort)
		}
	}
	return nil

}

// GenerateSSHToken generates a one-time access token for SSH connections
func (s *virtualboxService) GenerateSSHToken(ctx context.Context, vmID string, username string, password string) (*vbtypes.SSHTokenResponse, error) {
	// Validate VM exists and is running
	vm, err := s.GetVM(ctx, vmID)
	if err != nil {
		return nil, fmt.Errorf("VM not found: %w", err)
	}

	if vm.Status != vbtypes.Running {
		return nil, fmt.Errorf("VM is not running")
	}

	_, exists := s.sshServer.GetVMConfig(vmID)
	if !exists {
		if vm.SSHPort == 0 {
			return nil, fmt.Errorf("VM not found in SSH configuration")
		}
		s.sshServer.AddVMConfig(vmID, "127.0.0.1", strconv.Itoa(vm.SSHPort))
	}

	// Use the WebSocket handler to generate the token
	tokenResponse, err := s.wsHandler.GenerateSSHToken(vmID, username, password)
	if err != nil {
		return nil, err
	}

	return &vbtypes.SSHTokenResponse{
		Token:     tokenResponse.Token,
		ExpiresAt: tokenResponse.ExpiresAt,
	}, nil
}

// ValidateAndConsumeSSHToken validates a token and returns the credentials if valid
func (s *virtualboxService) ValidateAndConsumeSSHToken(token string) (*vbtypes.SSHAccessToken, error) {
	accessToken, err := s.wsHandler.ValidateAndConsumeSSHToken(token)
	if err != nil {
		return nil, err
	}

	return &vbtypes.SSHAccessToken{
		Token:     accessToken.Token,
		VMID:      accessToken.VMID,
		Username:  accessToken.Username,
		Password:  accessToken.Password,
		CreatedAt: accessToken.CreatedAt,
		ExpiresAt: accessToken.ExpiresAt,
		Used:      accessToken.Used,
	}, nil
}

// GetSSHServer returns the SSH server instance
func (s *virtualboxService) GetSSHServer() *ssh_connection.SSHServer {
	return s.sshServer
}

func (s *virtualboxService) CollectMetrics(ctx context.Context, vmId string, conn *websocket.Conn, period int) error {
	// Add panic recovery to the entire function
	defer func() {
		if r := recover(); r != nil {
			serviceLog.Errorf("Panic in CollectMetrics for VM %s: %v", vmId, r)
		}
	}()

	serviceLog.Infof("Starting metrics collection for VM: %s with period: %d seconds", vmId, period)

	// Check available metrics using VBoxManageExecutor
	checkOutput, err := s.vBoxService.ListAvailableMetrics(vmId)
	if err != nil {
		serviceLog.Errorf("Error checking available metrics: %v", err)
	} else {
		serviceLog.Debugf("Available metrics:\n%s", checkOutput)
	}

	// Define all metrics to collect - based on what we know works
	allMetrics := []string{
		"CPU/Load/User",
		"CPU/Load/Kernel",
		"RAM/Usage/Used",
		"Disk/Usage/Used",
		"Net/Rate/Rx",
		"Net/Rate/Tx",
	}

	// Start VBoxManage metrics collect command using VBoxManageExecutor
	cmd, err := s.vBoxService.CollectMetricsCmd(ctx, vmId, allMetrics, period)
	if err != nil {
		return fmt.Errorf("error starting metrics collection: %w", err)
	}

	// Create pipes to capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting VBoxManage: %w", err)
	}

	// Ensure the process is cleaned up when the function returns
	defer func() {
		if cmd.Process != nil {
			serviceLog.Debugf("Cleaning up VBoxManage process for VM: %s", vmId)

			// Try graceful termination first
			if err := cmd.Process.Signal(os.Interrupt); err != nil {
				serviceLog.Debugf("Process already terminated or error sending interrupt: %v", err)
			} else {
				// Give it a moment to terminate gracefully
				time.Sleep(1 * time.Second)
			}

			// Force kill if still running
			if err := cmd.Process.Kill(); err != nil {
				serviceLog.Debugf("Process already terminated or error killing: %v", err)
			}

			// Wait a bit more to ensure process is fully terminated
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Create a context to signal when WebSocket disconnects
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up WebSocket close handler to detect disconnections
	originalCloseHandler := conn.CloseHandler()
	conn.SetCloseHandler(func(code int, text string) error {
		serviceLog.Debugf("WebSocket close handler triggered for VM: %s (code: %d, text: %s)", vmId, code, text)
		cancel() // Cancel context instead of sending to channel
		if originalCloseHandler != nil {
			return originalCloseHandler(code, text)
		}
		return nil
	})

	// Set up pong handler to detect disconnections
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	})

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Safe WebSocket write function with panic recovery
	safeWriteMessage := func(messageType int, data []byte) error {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in WebSocket write for VM %s: %v", vmId, r)
			}
		}()

		// Set a write deadline to prevent hanging
		if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return err
		}

		err := conn.WriteMessage(messageType, data)

		// Reset write deadline
		conn.SetWriteDeadline(time.Time{})

		return err
	}

	// Monitor WebSocket connection in a goroutine with better error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in VM status monitoring goroutine for VM %s: %v", vmId, r)
				cancel() // Cancel context on panic
			}
		}()

		ticker := time.NewTicker(1 * time.Second) // Check every 1 second for faster detection
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				serviceLog.Debugf("Context cancelled, stopping WebSocket monitoring for VM: %s", vmId)
				return
			case <-ticker.C:
				// Check if VM is still running
				vmState, err := s.vBoxService.GetVMStatus(vmId)
				if err != nil {
					serviceLog.Warnf("Failed to get VM status for %s: %v", vmId, err)
					// If we can't get VM status, assume it's stopped and terminate
					serviceLog.Infof("VM %s status check failed, terminating metrics collection", vmId)
					cancel() // Cancel context on error
					return
				} else if vmState != "running" {
					serviceLog.Infof("VM %s is no longer running (status: %s), terminating metrics collection", vmId, vmState)

					// Send error message to WebSocket client before disconnecting
					errorData := map[string]string{
						"type":    "error",
						"message": fmt.Sprintf("VM stopped during metrics collection. Status: %s", vmState),
					}
					jsonData, _ := json.Marshal(errorData)
					if err := safeWriteMessage(websocket.TextMessage, jsonData); err != nil {
						serviceLog.Debugf("Failed to send VM stopped error to WebSocket: %v", err)
					} else {
						// Small delay to ensure message is delivered
						time.Sleep(100 * time.Millisecond)
					}

					cancel() // Cancel context on VM stopped
					return
				}

				// Set a shorter write deadline to detect disconnection faster
				if err := conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
					serviceLog.Debugf("WebSocket disconnected, terminating VBoxManage process for VM: %s", vmId)
					cancel() // Cancel context on WebSocket disconnect
					return
				}

				// Send ping to check connection
				if err := safeWriteMessage(websocket.PingMessage, nil); err != nil {
					serviceLog.Debugf("WebSocket disconnected, terminating VBoxManage process for VM: %s", vmId)
					cancel() // Cancel context on WebSocket disconnect
					return
				}

				// Reset write deadline
				if err := conn.SetWriteDeadline(time.Time{}); err != nil {
					serviceLog.Debugf("WebSocket disconnected, terminating VBoxManage process for VM: %s", vmId)
					cancel() // Cancel context on WebSocket disconnect
					return
				}
			}
		}
	}()

	// Start a background goroutine to actively read from WebSocket to detect disconnections immediately
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in WebSocket read goroutine for VM %s: %v", vmId, r)
				cancel() // Cancel context on panic
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Set read deadline to prevent hanging
			if err := conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
				serviceLog.Debugf("Failed to set read deadline for VM %s: %v", vmId, err)
				cancel() // Cancel context on error
				return
			}

			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					serviceLog.Debugf("WebSocket read error (client disconnected): %v", err)
				} else {
					serviceLog.Debugf("WebSocket closed normally: %v", err)
				}
				cancel() // Cancel context on WebSocket disconnect
				return
			}
		}
	}()

	// Read stdout in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in stdout reading goroutine for VM %s: %v", vmId, r)
				cancel() // Cancel context on panic
			}
		}()

		scanner := bufio.NewScanner(stdout)
		var currentSnapshot *vbtypes.VMMetricsSnapshot
		var currentTimestamp string

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			serviceLog.Debugf("VBoxManage output: %s", line)

			// Parse the line into structured data
			metricData, err := util.ParseVBoxManageOutput(line)
			if err != nil {
				serviceLog.Warnf("Error parsing line: %v", err)
				continue
			}

			// Skip if it's a header or empty line
			if metricData == nil {
				serviceLog.Debugf("Skipping line (header or empty): %s", line)
				continue
			}

			serviceLog.Debugf("Successfully parsed metric: %s = %f %s", metricData.Metric, metricData.Value, metricData.Unit)

			// Check if this is a new timestamp (new snapshot)
			if currentTimestamp != metricData.Timestamp {
				// Send previous snapshot if it exists
				if currentSnapshot != nil {
					jsonData, err := json.Marshal(currentSnapshot)
					if err != nil {
						serviceLog.Errorf("Error marshaling snapshot JSON: %v", err)
					} else {
						// Send complete snapshot to WebSocket client
						if err := safeWriteMessage(websocket.TextMessage, jsonData); err != nil {
							serviceLog.Debugf("WebSocket write error (client likely disconnected): %v", err)
							cancel() // Cancel context on error
							return
						}
						serviceLog.Debugf("Sent complete snapshot for timestamp %s with %d metrics", currentSnapshot.Timestamp, len(currentSnapshot.Metrics))
					}
				}

				// Start new snapshot
				currentTimestamp = metricData.Timestamp
				currentSnapshot = &vbtypes.VMMetricsSnapshot{
					Timestamp: metricData.Timestamp,
					VMId:      vmId,
					Metrics:   make(map[string]vbtypes.VMMetricValue),
				}
			}

			// Add metric to current snapshot (only value and unit, no redundant timestamp/metric name)
			currentSnapshot.Metrics[metricData.Metric] = vbtypes.VMMetricValue{
				Value: metricData.Value,
				Unit:  metricData.Unit,
			}
		}

		// Send the last snapshot if it exists
		if currentSnapshot != nil {
			jsonData, err := json.Marshal(currentSnapshot)
			if err != nil {
				serviceLog.Errorf("Error marshaling final snapshot JSON: %v", err)
			} else {
				if err := safeWriteMessage(websocket.TextMessage, jsonData); err != nil {
					serviceLog.Debugf("WebSocket write error on final snapshot (client likely disconnected): %v", err)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			serviceLog.Errorf("Error reading stdout: %v", err)
		}
	}()

	// Read stderr in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in stderr reading goroutine for VM %s: %v", vmId, r)
				cancel() // Cancel context on panic
			}
		}()

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()

			// Send error messages to WebSocket client as well
			errorData := map[string]string{
				"type":    "error",
				"message": line,
			}
			jsonData, _ := json.Marshal(errorData)
			if err := safeWriteMessage(websocket.TextMessage, jsonData); err != nil {
				serviceLog.Debugf("WebSocket write error on stderr (client likely disconnected): %v", err)
				cancel() // Cancel context on error
				return
			}
		}
		if err := scanner.Err(); err != nil {
			serviceLog.Errorf("Error reading stderr: %v", err)
		}
	}()

	// Wait for either completion or WebSocket disconnect
	select {
	case <-ctx.Done():
		serviceLog.Infof("WebSocket disconnected, terminating VBoxManage metrics collection for VM: %s", vmId)
	case <-time.After(24 * time.Hour): // Safety timeout
		serviceLog.Infof("Safety timeout reached, terminating VBoxManage metrics collection for VM: %s", vmId)
	}

	// Wait for the process to finish with timeout
	processDone := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serviceLog.Errorf("Panic in process wait goroutine for VM %s: %v", vmId, r)
			}
		}()
		processDone <- cmd.Wait()
	}()

	select {
	case err := <-processDone:
		if err != nil {
			serviceLog.Debugf("VBoxManage process finished with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		serviceLog.Debugf("Timeout waiting for VBoxManage process to finish for VM: %s", vmId)
	}

	serviceLog.Infof("Metrics collection stopped for VM: %s", vmId)
	return nil
}
