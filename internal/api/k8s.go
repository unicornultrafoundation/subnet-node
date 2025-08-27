package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	k8s "github.com/unicornultrafoundation/subnet-node/core/k8s"
	kubeclienterrors "github.com/unicornultrafoundation/subnet-node/core/k8s/kube/errors"
	ctypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1"
	apclient "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/provider/client"
	wsutil "github.com/unicornultrafoundation/subnet-node/internal/api/ws"
	crd "github.com/unicornultrafoundation/subnet-node/pkg/k8s/apis/subnet.node/v1"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/apitypes"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
	provider "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/provider/v1"
	"k8s.io/client-go/tools/remotecommand"
)

type ContextKey int

const (
	LeaseContextKey ContextKey = iota + 1
	DeploymentContextKey
	LogFollowContextKey
	TailLinesContextKey
	ServiceContextKey
	OwnerContextKey
	ProviderContextKey
	ServicesContextKey
	ClaimsContextKey
)

const (
	// as per RFC https://www.iana.org/assignments/websocket/websocket.xhtml#close-code-number
	// errors from private use staring
	websocketInternalServerErrorCode = 4000
	websocketLeaseNotFound           = 4001
	manifestSubmitTimeout            = 120 * time.Second
)

const (
	LeaseShellCodeStdout         = 100
	LeaseShellCodeStderr         = 101
	LeaseShellCodeResult         = 102
	LeaseShellCodeFailure        = 103
	LeaseShellCodeStdin          = 104
	LeaseShellCodeTerminalResize = 105
)

// K8sHandler provides HTTP endpoints for deployment management
type K8sHandler struct {
	deployer        K8sService
	providerAddress common.Address
}

type K8sService interface {
	RequestDeployment(ctx context.Context, deployment dtypes.DeploymentID, sdlManifest sdl.SDL) error
	StatusV1(ctx context.Context) (*provider.ClusterStatus, error)
	GetAllLeaseStatus(ctx context.Context) ([]apitypes.DeploymentStatus, error)
	DeleteDeployment(lid mtypes.LeaseID) error
	GetLeaseStatus(ctx context.Context, leaseID mtypes.LeaseID) (apclient.LeaseStatus, error)
	ServiceStatus(context.Context, mtypes.LeaseID, string) (*apclient.ServiceStatus, error)
	GetManifestGroup(ctx context.Context, leaseID mtypes.LeaseID) (bool, crd.ManifestGroup, error)

	// WebSocket routes
	Exec(ctx context.Context,
		lID mtypes.LeaseID,
		service string,
		podIndex uint,
		cmd []string,
		stdin io.Reader,
		stdout io.Writer,
		stderr io.Writer,
		tty bool,
		tsq remotecommand.TerminalSizeQueue) (ctypes.ExecResult, error)
	LeaseLogs(context.Context, mtypes.LeaseID, string, bool, *int64) ([]*ctypes.ServiceLog, error)
}

func NewK8sHandler(deployer K8sService, account *account.AccountService) *K8sHandler {
	return &K8sHandler{
		deployer:        deployer,
		providerAddress: account.GetAddress(),
	}
}

// Router returns the chi router with all deployment routes
func (h *K8sHandler) Router() *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	//Group for authenticated routes
	r.Group(func(r chi.Router) {
		// WebSocket routes for exec and logs
		r.Get("/api/v1/deployments/{owner}/{dseq}/ws/exec", h.execWebSocketHandler)
		r.Get("/api/v1/deployments/{owner}/{dseq}/ws/logs", h.logsWebSocketHandler)

		// Deployment management routes
		r.Post("/api/v1/deployments", h.createDeploymentHandler)
		r.Get("/api/v1/status", h.getStatusHandler)
		r.Get("/api/v1/leases", h.getAllLeaseStatusHandler)
		r.Delete("/api/v1/leases/{owner}/{dseq}", h.deleteDeploymentHandler)
		r.Get("/api/v1/leases/{owner}/{dseq}", h.getLeaseStatusHandler)
		r.Get("/api/v1/leases/{owner}/{dseq}/manifest", h.getLeaseManifestHandler)
	})

	return r
}

type K8sDeploymentRequest struct {
	Owner    string          `json:"owner"`
	OrderID  string          `json:"order_id"`
	Manifest json.RawMessage `json:"manifest"`
}

// createDeploymentHandler handles deployment creation requests
func (h *K8sHandler) createDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	var req K8sDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	dseq, err := strconv.ParseUint(req.OrderID, 10, 64)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	owner := req.Owner
	if owner == "" || !common.IsHexAddress(owner) {
		h.sendErrorResponse(w, "Invalid owner address", http.StatusBadRequest)
		return
	}
	owner = strings.ToLower(common.HexToAddress(owner).Hex())

	deploymentID := dtypes.DeploymentID{
		Owner: owner,
		DSeq:  dseq,
	}
	sdlManifest, err := sdl.ReadJSON(req.Manifest)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.deployer.RequestDeployment(ctx, deploymentID, sdlManifest)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Deployment requested successfully"})
}

func (h *K8sHandler) getStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status, err := h.deployer.StatusV1(ctx)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *K8sHandler) getAllLeaseStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	deploymentStatuses, err := h.deployer.GetAllLeaseStatus(ctx)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deploymentStatuses)
}

func (h *K8sHandler) getLeaseStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	owner := chi.URLParam(r, "owner")
	if owner == "" || !common.IsHexAddress(owner) {
		h.sendErrorResponse(w, "Invalid owner address", http.StatusBadRequest)
		return
	}

	dseq, err := strconv.ParseUint(chi.URLParam(r, "dseq"), 10, 64)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	leaseID := mtypes.LeaseID{
		Owner:    strings.ToLower(owner),
		DSeq:     dseq,
		Provider: strings.ToLower(h.providerAddress.Hex()),
	}

	leaseStatus, err := h.deployer.GetLeaseStatus(ctx, leaseID)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaseStatus)
}

func (h *K8sHandler) getLeaseManifestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	owner := chi.URLParam(r, "owner")
	if owner == "" || !common.IsHexAddress(owner) {
		h.sendErrorResponse(w, "Invalid owner address", http.StatusBadRequest)
		return
	}

	dseq, err := strconv.ParseUint(chi.URLParam(r, "dseq"), 10, 64)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	leaseID := mtypes.LeaseID{
		Owner:    strings.ToLower(owner),
		DSeq:     dseq,
		Provider: strings.ToLower(h.providerAddress.Hex()),
	}

	found, grp, err := h.deployer.GetManifestGroup(ctx, leaseID)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !found {
		h.sendErrorResponse(w, "lease not found", http.StatusNotFound)
		return
	}

	mgrp, _, err := grp.FromCRD()
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mgrp)
}

func (h *K8sHandler) deleteDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	if owner == "" || !common.IsHexAddress(owner) {
		h.sendErrorResponse(w, "Invalid owner address", http.StatusBadRequest)
		return
	}
	owner = strings.ToLower(common.HexToAddress(owner).Hex())

	dseq, err := strconv.ParseUint(chi.URLParam(r, "dseq"), 10, 64)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	leaseID := mtypes.LeaseID{
		Owner:    owner,
		DSeq:     dseq,
		Provider: strings.ToLower(h.providerAddress.Hex()),
	}

	err = h.deployer.DeleteDeployment(leaseID)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Deployment deleted successfully"})
}

type leaseShellResponse struct {
	ExitCode int    `json:"exit_code"`
	Message  string `json:"message,omitempty"`
}

func (h *K8sHandler) execWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Get parameters
	owner := chi.URLParam(r, "owner")
	if owner == "" || !common.IsHexAddress(owner) {
		h.sendErrorResponse(w, "Invalid owner address", http.StatusBadRequest)
		return
	}

	dseq, err := strconv.ParseUint(chi.URLParam(r, "dseq"), 10, 64)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	leaseID := mtypes.LeaseID{
		Owner:    owner,
		DSeq:     dseq,
		Provider: strings.ToLower(h.providerAddress.Hex()),
	}

	// Get parameters from query string
	vars := r.URL.Query()
	var cmd []string
	for i := 0; true; i++ {
		v := vars.Get(fmt.Sprintf("cmd%d", i))
		if len(v) == 0 {
			break
		}
		cmd = append(cmd, v)
	}

	tty := vars.Get("tty")
	if len(tty) == 0 {
		h.sendErrorResponse(w, "missing parameter tty", http.StatusBadRequest)
		return
	}
	isTty := tty == "1"

	serviceName := vars.Get("service")
	if len(serviceName) == 0 {
		h.sendErrorResponse(w, "missing parameter service", http.StatusBadRequest)
		return
	}

	stdin := vars.Get("stdin")
	if len(stdin) == 0 {
		h.sendErrorResponse(w, "missing parameter stdin", http.StatusBadRequest)
		return
	}
	connectStdin := stdin == "1"

	podIndex, err := strconv.ParseUint(vars.Get("pod_index"), 10, 32)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Upgrade to websocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  0,
		WriteBufferSize: 0,
	}

	shellWs, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.sendErrorResponse(w, fmt.Sprintf("failed to upgrade to websocket: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Handle exec
	var stdinPipeOut *io.PipeWriter
	var stdinPipeIn *io.PipeReader
	wg := &sync.WaitGroup{}

	var tsq remotecommand.TerminalSizeQueue
	var terminalSizeUpdate chan remotecommand.TerminalSize
	if isTty {
		terminalSizeUpdate = make(chan remotecommand.TerminalSize, 1)
		tsq = channelToTerminalSizeQueue(terminalSizeUpdate)
	}

	localLog := logrus.New()
	localLog.SetLevel(logrus.InfoLevel)

	if connectStdin {
		stdinPipeIn, stdinPipeOut = io.Pipe()

		wg.Add(1)
		go leaseShellWebsocketHandler(localLog, wg, shellWs, stdinPipeOut, terminalSizeUpdate)
	}

	responseData := leaseShellResponse{}
	l := &sync.Mutex{}

	resultWriter := wsutil.NewWsWriterWrapper(shellWs, LeaseShellCodeResult, l)

	encodeData := true

	status, err := h.deployer.ServiceStatus(r.Context(), leaseID, serviceName)
	if err != nil {
		if k8s.ErrorIsOkToSendToClient(err) || errors.Is(err, kubeclienterrors.ErrNoServiceForLease) {
			responseData.Message = err.Error()
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	if err == nil && status.ReadyReplicas == 0 {
		err = errors.New("no active replicase for service")
		responseData.Message = err.Error()
	}

	if err == nil {
		stdout := wsutil.NewWsWriterWrapper(shellWs, LeaseShellCodeStdout, l)
		stderr := wsutil.NewWsWriterWrapper(shellWs, LeaseShellCodeStderr, l)

		subctx, subcancel := context.WithCancel(r.Context())
		wg.Add(1)
		go leaseShellPingHandler(subctx, wg, shellWs)

		var stdinForExec io.Reader
		if connectStdin {
			stdinForExec = stdinPipeIn
		}
		result, err := h.deployer.Exec(subctx, leaseID, serviceName, uint(podIndex), cmd, stdinForExec, stdout, stderr, isTty, tsq)
		subcancel()

		if result != nil {
			responseData.ExitCode = result.ExitCode()

			localLog.WithField("exitcode", result.ExitCode()).Info("lease shell completed")
		} else {
			if k8s.ErrorIsOkToSendToClient(err) {
				responseData.Message = err.Error()
			} else {
				resultWriter = wsutil.NewWsWriterWrapper(shellWs, LeaseShellCodeFailure, l)
				// Don't return errors like this to the client, they could contain information
				// that should not be let out
				encodeData = false

				localLog.WithError(err).Error("lease exec failed")
			}
		}
	}

	if encodeData {
		encoder := json.NewEncoder(resultWriter)
		err = encoder.Encode(responseData)
	} else {
		// Just send an empty message so the remote knows things are over
		_, err = resultWriter.Write([]byte{})
	}

	_ = shellWs.Close()

	if err != nil {
		localLog.WithError(err).Error("failed writing response to client after exec")
	}

	wg.Wait()

	if stdinPipeOut != nil {
		_ = stdinPipeOut.Close()
	}
	if stdinPipeIn != nil {
		_ = stdinPipeIn.Close()
	}

	if terminalSizeUpdate != nil {
		close(terminalSizeUpdate)
	}

}

func leaseShellPingHandler(ctx context.Context, wg *sync.WaitGroup, ws *websocket.Conn) {
	defer wg.Done()
	pingTicker := time.NewTicker(wsutil.PingPeriod)
	defer pingTicker.Stop()

	for {
		select {
		case <-pingTicker.C:
			const pingWriteWaitTime = 5 * time.Second
			if err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(pingWriteWaitTime)); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func leaseShellWebsocketHandler(log *logrus.Logger, wg *sync.WaitGroup, shellWs *websocket.Conn, stdinPipeOut io.Writer, terminalSizeUpdate chan<- remotecommand.TerminalSize) {
	defer wg.Done()
	for {
		shellWs.SetPongHandler(func(string) error {
			return shellWs.SetReadDeadline(time.Now().Add(wsutil.PingWait))
		})

		msgType, data, err := shellWs.ReadMessage()
		if err != nil {
			return
		}

		// Just ignore anything not a binary message or that is empty
		if msgType != websocket.BinaryMessage || len(data) == 0 {
			continue
		}

		msgID := data[0]
		msg := data[1:]
		switch msgID {
		case LeaseShellCodeStdin:
			_, err := stdinPipeOut.Write(msg)
			if err != nil {
				return
			}
		case LeaseShellCodeTerminalResize:
			var size remotecommand.TerminalSize
			r := bytes.NewReader(msg)
			// Unpack data, its just binary encoded data in big endian
			err = binary.Read(r, binary.BigEndian, &size.Width)
			if err != nil {
				return
			}
			err = binary.Read(r, binary.BigEndian, &size.Height)
			if err != nil {
				return
			}

			log.WithFields(logrus.Fields{
				"width":  size.Width,
				"height": size.Height,
			}).Debug("terminal resize received")
			if terminalSizeUpdate != nil {
				terminalSizeUpdate <- size
			}
		default:
			log.WithField("code", msgID).Error("unknown message ID on websocket")
			return
		}

	}
}

type wsStreamConfig struct {
	lid       mtypes.LeaseID
	services  string
	follow    bool
	tailLines *int64
	log       *logrus.Logger
	client    K8sService
}

func (h *K8sHandler) logsWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	if owner == "" || !common.IsHexAddress(owner) {
		h.sendErrorResponse(w, "Invalid owner address", http.StatusBadRequest)
		return
	}

	dseq, err := strconv.ParseUint(chi.URLParam(r, "dseq"), 10, 64)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := r.URL.Query()

	var tailLines *int64

	services := vars.Get("service")
	if strings.HasSuffix(services, ",") {
		err = errors.Errorf("parameter \"service\" must not contain trailing comma")
		return
	}

	follow := false

	if val := vars.Get("follow"); val != "" {
		follow, err = strconv.ParseBool(val)
		if err != nil {
			return
		}
	}

	vl := new(int64)
	if val := vars.Get("tail"); val != "" {
		*vl, err = strconv.ParseInt(val, 10, 32)
		if err != nil {
			return
		}

		if *vl < -1 {
			err = errors.Errorf("parameter \"tail\" contains invalid value")
			return
		}
	} else {
		*vl = -1
	}

	if *vl > -1 {
		tailLines = vl
	}

	leaseID := mtypes.LeaseID{
		Owner:    owner,
		DSeq:     dseq,
		Provider: strings.ToLower(h.providerAddress.Hex()),
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  0,
		WriteBufferSize: 0,
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// At this point the connection either has a response sent already
		// or it has been closed
		return
	}

	cfg := wsStreamConfig{
		lid:       leaseID,
		services:  services,
		follow:    follow,
		tailLines: tailLines,
		client:    h.deployer,
		log:       logrus.New().WithField("api", "stream-logs").WithField("lease_id", leaseID.String()).Logger,
	}

	wsLogWriter(r.Context(), ws, cfg)
}

func wsLogWriter(ctx context.Context, ws *websocket.Conn, cfg wsStreamConfig) {
	pingTicker := time.NewTicker(wsutil.PingPeriod)

	cctx, cancel := context.WithCancel(ctx)
	defer func() {
		pingTicker.Stop()
		cancel()
		_ = ws.Close()
	}()

	logs, err := cfg.client.LeaseLogs(cctx, cfg.lid, cfg.services, cfg.follow, cfg.tailLines)
	if err != nil {
		cfg.log.WithError(err).Error("couldn't fetch logs")
		err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocketInternalServerErrorCode, ""))
		if err != nil {
			cfg.log.WithError(err).Error("couldn't push control message through websocket")
		}
		return
	}

	if len(logs) == 0 {
		_ = ws.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocketInternalServerErrorCode, "no running pods"))
		return
	}

	if err = wsSetupPongHandler(ws, cancel); err != nil {
		return
	}

	var scanners sync.WaitGroup

	logch := make(chan apclient.ServiceLogMessage)

	scanners.Add(len(logs))

	for _, lg := range logs {
		go func(name string, scan *bufio.Scanner) {
			defer scanners.Done()

			for scan.Scan() && ctx.Err() == nil {
				logch <- apclient.ServiceLogMessage{
					Name:    name,
					Message: scan.Text(),
				}
			}
		}(lg.Name, lg.Scanner)
	}

	donech := make(chan struct{})

	go func() {
		scanners.Wait()
		close(donech)
	}()

done:
	for {
		select {
		case line := <-logch:
			if err = ws.WriteJSON(line); err != nil {
				break done
			}
		case <-pingTicker.C:
			if err = ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				break done
			}
			if err = ws.SetReadDeadline(time.Now().Add(wsutil.PingWait)); err != nil {
				break done
			}
		case <-donech:
			break done
		}
	}

	cancel()

	for i := range logs {
		_ = logs[i].Stream.Close()
	}

	// drain logs channel in separate goroutine to unblock seeders waiting for write space
	go func() {
	drain:
		for {
			select {
			case <-donech:
				break drain
			case <-logch:
			}
		}
	}()
}

func wsSetupPongHandler(ws *websocket.Conn, cancel func()) error {
	if err := ws.SetReadDeadline(time.Time{}); err != nil {
		return err
	}

	ws.SetPongHandler(func(string) error {
		return ws.SetReadDeadline(time.Now().Add(wsutil.PingWait))
	})

	go func() {
		var err error

		defer func() {
			if err != nil {
				cancel()
			}
		}()

		for {
			var mtype int
			if mtype, _, err = ws.ReadMessage(); err != nil {
				break
			}

			if mtype == websocket.CloseMessage {
				err = errors.Errorf("disconnect")
			}
		}
	}()

	return nil
}

// sendErrorResponse sends a standardized error response
func (h *K8sHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"message": http.StatusText(statusCode),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
