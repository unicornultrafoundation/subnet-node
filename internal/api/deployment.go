package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"bufio"
	"encoding/binary"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	bidenginetypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	authchain "github.com/unicornultrafoundation/subnet-node/crypto/authchain"
	"github.com/unicornultrafoundation/subnet-node/internal/api/ws"
	"k8s.io/client-go/tools/remotecommand"
)

// Context key types to avoid string literal collisions
type contextKey string

const (
	userAddressKey contextKey = "userAddress"
	orderIDKey     contextKey = "orderID"
	entityIDKey    contextKey = "entityID"
)

// WebSocket message codes for deployment shell
const (
	DeploymentShellCodeStdin   = 0
	DeploymentShellCodeStdout  = 1
	DeploymentShellCodeStderr  = 2
	DeploymentShellCodeResult  = 3
	DeploymentShellCodeFailure = 4
	DeploymentShellCodeResize  = 5
)

// DeploymentHandler provides HTTP endpoints for deployment management
type DeploymentHandler struct {
	deployer    DeployerService
	logger      *logrus.Entry
	cfg         ConfigProvider
	ordersCache *OrdersWithCache
}

type BidMarketContract interface {
	GetOrder(ctx context.Context, orderID *big.Int) (*bidenginetypes.Order, error)
}

// ConfigProvider interface for configuration access
type ConfigProvider interface {
	GetString(key string, defaultValue string) string
	GetBool(key string, defaultValue bool) bool
}

type OrdersWithCache struct {
	orders    map[string]*cachedOrder
	bidMarket BidMarketContract
	mu        sync.RWMutex
}

type cachedOrder struct {
	order     *bidenginetypes.Order
	timestamp time.Time
}

func NewOrdersWithCache(bidMarket BidMarketContract) *OrdersWithCache {
	cache := &OrdersWithCache{
		orders:    make(map[string]*cachedOrder),
		bidMarket: bidMarket,
	}

	// Start background cleanup goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Minute) // Clean up every 10 minutes
		defer ticker.Stop()

		for range ticker.C {
			cache.CleanupExpired()
		}
	}()

	return cache
}

func (o *OrdersWithCache) GetOrder(ctx context.Context, orderID string) (*bidenginetypes.Order, error) {
	// Check cache first
	o.mu.RLock()
	if cached, ok := o.orders[orderID]; ok {
		// Check if cache is still valid (1 hour)
		if time.Since(cached.timestamp) < time.Hour {
			o.mu.RUnlock()
			return cached.order, nil
		}
		// Cache expired, remove it
		o.mu.RUnlock()
		o.mu.Lock()
		delete(o.orders, orderID)
		o.mu.Unlock()
	} else {
		o.mu.RUnlock()
	}

	// Parse orderID to int64
	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid orderID format: %w", err)
	}

	// Fetch from bid market
	order, err := o.bidMarket.GetOrder(ctx, big.NewInt(orderIDInt))
	if err != nil {
		return nil, err
	}

	// Cache the result with timestamp
	o.mu.Lock()
	o.orders[orderID] = &cachedOrder{
		order:     order,
		timestamp: time.Now(),
	}
	o.mu.Unlock()

	return order, nil
}

func (o *OrdersWithCache) ClearCache() {
	o.mu.Lock()
	o.orders = make(map[string]*cachedOrder)
	o.mu.Unlock()
}

// CleanupExpired removes expired entries from cache
func (o *OrdersWithCache) CleanupExpired() {
	o.mu.Lock()
	defer o.mu.Unlock()

	now := time.Now()
	for orderID, cached := range o.orders {
		if now.Sub(cached.timestamp) >= time.Hour {
			delete(o.orders, orderID)
		}
	}
}

// GetCacheStats returns cache statistics
func (o *OrdersWithCache) GetCacheStats() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	now := time.Now()
	totalEntries := len(o.orders)
	expiredEntries := 0

	for _, cached := range o.orders {
		if now.Sub(cached.timestamp) >= time.Hour {
			expiredEntries++
		}
	}

	return map[string]interface{}{
		"total_entries":   totalEntries,
		"expired_entries": expiredEntries,
		"valid_entries":   totalEntries - expiredEntries,
	}
}

// DeployerService interface for deployment operations
type DeployerService interface {
	RequestDeployment(ctx context.Context, req *types.DeploymentRequest) (*types.DeploymentResponse, error)
	GetDeployment(ctx context.Context, orderID string) (*types.DeploymentResponse, error)
	CleanupDeployment(ctx context.Context, orderID string) error
	GetDeployments(ctx context.Context, requester string) ([]*types.DeploymentResponse, error)
	GetDeploymentLogs(ctx context.Context, orderID string) ([]*types.ServiceLog, error)
	GetServiceStatus(ctx context.Context, orderID, serviceName string) (*types.ServiceStatus, error)
	Exec(ctx context.Context, orderID, podName, serviceName string, cmd []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, tty bool, tsq remotecommand.TerminalSizeQueue) (types.ExecResult, error)
}

// NewDeploymentHandler creates a new deployment handler
func NewDeploymentHandler(deployer DeployerService, cfg ConfigProvider, bidMarket BidMarketContract) *DeploymentHandler {
	return &DeploymentHandler{
		deployer:    deployer,
		logger:      logrus.WithField("component", "deployment-handler"),
		cfg:         cfg,
		ordersCache: NewOrdersWithCache(bidMarket),
	}
}

// Router returns the chi router with all deployment routes
func (h *DeploymentHandler) Router() *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint (no auth required)
	r.Get("/health", h.healthHandler)

	// Create auth middleware
	authMiddleware := NewAuthMiddleware(h.cfg, h.ordersCache)

	// Group for authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware())

		// WebSocket routes for exec and logs
		r.Get("/{orderID}/ws/exec", h.execWebSocketHandler)
		r.Get("/{orderID}/ws/logs", h.logsWebSocketHandler)

		// Deployment management routes
		r.Post("/", h.createDeploymentHandler)
		r.Get("/{orderID}", h.getDeploymentHandler)
		r.Delete("/{orderID}", h.cleanupDeploymentHandler)

		// Service and pod management routes
		r.Get("/{orderID}/services/{serviceName}/status", h.getServiceStatusHandler)
		r.Get("/{orderID}/logs", h.getDeploymentLogsHandler)
	})

	return r
}

// AuthMiddleware handles authentication and authorization for deployment requests
type AuthMiddleware struct {
	logger      *logrus.Entry
	cfg         ConfigProvider
	ordersCache *OrdersWithCache
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(cfg ConfigProvider, ordersCache *OrdersWithCache) *AuthMiddleware {
	return &AuthMiddleware{
		logger:      logrus.WithField("component", "auth-middleware"),
		cfg:         cfg,
		ordersCache: ordersCache,
	}
}

// Middleware returns the HTTP middleware function
func (a *AuthMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract authchain from header
			authChainHeader := r.Header.Get("X-AuthChain")
			if authChainHeader == "" {
				a.sendErrorResponse(w, "X-AuthChain header required", http.StatusUnauthorized)
				return
			}

			// Deserialize authchain
			authChain, err := authchain.DeserializeAuthChain(authChainHeader)
			if err != nil {
				a.sendErrorResponse(w, "Invalid authchain format: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Validate authchain format
			if err := authchain.ValidateAuthChainFormat(authChain); err != nil {
				a.sendErrorResponse(w, "Invalid authchain format: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Extract order ID from URL path
			orderID := chi.URLParam(r, "orderID")

			// For endpoints that don't have orderID in path, try to get from query params
			if orderID == "" {
				orderID = r.URL.Query().Get("orderID")
			}

			// For POST /deployments, extract orderID from request body
			if r.Method == "POST" {
				// Read body to extract orderID
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					var reqBody struct {
						OrderID string `json:"order_id"`
					}
					if err := json.Unmarshal(bodyBytes, &reqBody); err == nil {
						orderID = reqBody.OrderID
					}
					// Reset body for handlers to read again
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			if orderID == "" {
				a.sendErrorResponse(w, "Unauthorized: orderID is required", http.StatusUnauthorized)
				return
			}

			// Get order from cache if orderID exists
			var order *bidenginetypes.Order
			if orderID != "" {
				if cachedOrder, err := a.ordersCache.GetOrder(r.Context(), orderID); err == nil {
					order = cachedOrder
				} else {
					a.sendErrorResponse(w, "Unauthorized: orderID is required", http.StatusUnauthorized)
					return
				}
			}

			// Extract provider and machine IDs from config
			providerID := a.cfg.GetString("provider.id", "")
			machineID := a.cfg.GetString("provider.machine_id", "")

			// Extract user address from authchain
			userAddress := ""
			if len(authChain) > 0 {
				// For SIGNER type, the payload contains the address
				if authChain[0].Type == authchain.AuthLinkTypeSIGNER {
					userAddress = authChain[0].Payload
				}
			}

			// For endpoints with orderID, validate order ownership and machine assignment

			if order.AcceptedProviderId.String() != providerID || order.AcceptedMachineId.String() != machineID {
				a.sendErrorResponse(w, "Unauthorized: order does not belong to authenticated user", http.StatusForbidden)
				return
			}

			if order.Owner.Hex() != userAddress {
				a.sendErrorResponse(w, "Unauthorized: order does not belong to authenticated user", http.StatusForbidden)
				return
			}

			// Create entityID according to schema: "subnet_deployment:{providerId}:{machineId}:{order_id}"
			entityID := fmt.Sprintf("subnet_deployment:%s:%s:%s", providerID, machineID, orderID)
			// Validate authchain for the entityID with default 60s expiry
			result, err := authchain.ValidateAuthChainWithDefaultExpiry(authChain, entityID)
			if err != nil || result == nil || !result.OK {
				a.sendErrorResponse(w, "Unauthorized: invalid authchain", http.StatusUnauthorized)
				return
			}

			// Add user info to request context
			ctx := context.WithValue(r.Context(), userAddressKey, userAddress)
			ctx = context.WithValue(ctx, orderIDKey, orderID)
			ctx = context.WithValue(ctx, entityIDKey, entityID)

			// Log the request
			a.logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"userAddress": userAddress,
				"orderID":     orderID,
				"entityID":    entityID,
				"providerID":  providerID,
				"machineID":   machineID,
			}).Info("Authenticated deployment request")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// sendErrorResponse sends a standardized error response
func (a *AuthMiddleware) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"message": http.StatusText(statusCode),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// healthHandler handles health check requests
func (h *DeploymentHandler) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "deployment-api",
		"cache":     h.ordersCache.GetCacheStats(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions to extract values from context
func (h *DeploymentHandler) getUserAddress(ctx context.Context) string {
	if userAddr, ok := ctx.Value(userAddressKey).(string); ok {
		return userAddr
	}
	return ""
}

type DeploymentRequest struct {
	OrderID  string       `json:"order_id"`
	Manifest manifest.SDL `json:"manifest"`
}

// createDeploymentHandler handles deployment creation requests
func (h *DeploymentHandler) createDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	var req DeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	userAddress := h.getUserAddress(ctx)

	response, err := h.deployer.RequestDeployment(ctx, &types.DeploymentRequest{
		OrderID:   req.OrderID,
		Manifest:  req.Manifest,
		Requester: userAddress,
		TTL:       120,
	})
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// getDeploymentHandler handles deployment retrieval requests
func (h *DeploymentHandler) getDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	if orderID == "" {
		h.sendErrorResponse(w, "orderID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	response, err := h.deployer.GetDeployment(ctx, orderID)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// cleanupDeploymentHandler handles deployment cleanup requests
func (h *DeploymentHandler) cleanupDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	if orderID == "" {
		h.sendErrorResponse(w, "orderID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := h.deployer.CleanupDeployment(ctx, orderID)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Deployment cleaned up successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getServiceStatusHandler handles service status requests
func (h *DeploymentHandler) getServiceStatusHandler(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	serviceName := chi.URLParam(r, "serviceName")

	if orderID == "" || serviceName == "" {
		h.sendErrorResponse(w, "orderID and serviceName are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	response, err := h.deployer.GetServiceStatus(ctx, orderID, serviceName)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getDeploymentLogsHandler handles deployment logs requests
func (h *DeploymentHandler) getDeploymentLogsHandler(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	if orderID == "" {
		h.sendErrorResponse(w, "orderID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	response, err := h.deployer.GetDeploymentLogs(ctx, orderID)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// execWebSocketHandler handles WebSocket connections for pod execution
func (h *DeploymentHandler) execWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	if orderID == "" {
		h.sendErrorResponse(w, "orderID is required", http.StatusBadRequest)
		return
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

	podName := vars.Get("pod")
	if len(podName) == 0 {
		h.sendErrorResponse(w, "missing parameter pod", http.StatusBadRequest)
		return
	}

	h.handleExecWebSocket(w, r, orderID, podName, serviceName, cmd, isTty, connectStdin)
}

// logsWebSocketHandler handles WebSocket connections for log streaming
func (h *DeploymentHandler) logsWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	if orderID == "" {
		h.sendErrorResponse(w, "orderID is required", http.StatusBadRequest)
		return
	}

	h.handleLogsWebSocket(w, r, orderID)
}

// handleExecWebSocket handles the WebSocket connection for pod execution
func (h *DeploymentHandler) handleExecWebSocket(w http.ResponseWriter, r *http.Request, orderID string, podName string, serviceName string, cmd []string, isTty bool, connectStdin bool) {
	logger := h.logger.WithField("orderID", orderID)
	logger.Debug("WebSocket exec connection requested")

	conn, err := ws.SetupWebSocket(w, r, logger)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var stdinPipeOut *io.PipeWriter
	var stdinPipeIn *io.PipeReader
	wg := &sync.WaitGroup{}

	var tsq remotecommand.TerminalSizeQueue
	var terminalSizeUpdate chan remotecommand.TerminalSize
	if isTty {
		terminalSizeUpdate = make(chan remotecommand.TerminalSize, 1)
		tsq = channelToTerminalSizeQueue(terminalSizeUpdate)
	}

	if connectStdin {
		stdinPipeIn, stdinPipeOut = io.Pipe()

		wg.Add(1)
		go h.deploymentShellWebsocketHandler(logger, wg, conn, stdinPipeOut, terminalSizeUpdate)
	}

	responseData := deploymentShellResponse{}
	l := &sync.Mutex{}

	resultWriter := ws.NewWsWriterWrapper(conn, DeploymentShellCodeResult, l)

	encodeData := true

	status, err := h.deployer.GetServiceStatus(ctx, orderID, serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if status.ReadyReplicas == 0 {
		err = errors.New("no active replicas for service")
		responseData.Message = err.Error()
	}

	if err == nil {
		stdout := ws.NewWsWriterWrapper(conn, DeploymentShellCodeStdout, l)
		stderr := ws.NewWsWriterWrapper(conn, DeploymentShellCodeStderr, l)

		subctx, subcancel := context.WithCancel(r.Context())
		wg.Add(1)
		go h.deploymentShellPingHandler(subctx, wg, conn)

		var stdinForExec io.Reader
		if connectStdin {
			stdinForExec = stdinPipeIn
		}
		result, err := h.deployer.Exec(subctx, orderID, podName, serviceName, cmd, stdinForExec, stdout, stderr, isTty, tsq)
		subcancel()

		if result != nil {
			responseData.ExitCode = result.ExitCode()

			logger.Info("deployment shell completed", "exitcode", result.ExitCode())
		} else {
			responseData.Message = err.Error()
			resultWriter = ws.NewWsWriterWrapper(conn, DeploymentShellCodeFailure, l)
			// Don't return errors like this to the client, they could contain information
			// that should not be let out
			encodeData = false

			logger.Error("deployment exec failed", "err", err)
		}
	}

	if encodeData {
		encoder := json.NewEncoder(resultWriter)
		err = encoder.Encode(responseData)
	} else {
		// Just send an empty message so the remote knows things are over
		_, err = resultWriter.Write([]byte{})
	}

	_ = conn.Close()

	if err != nil {
		logger.Error("failed writing response to client after exec", "err", err)
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

// handleLogsWebSocket handles the WebSocket connection for log streaming
func (h *DeploymentHandler) handleLogsWebSocket(w http.ResponseWriter, r *http.Request, orderID string) {
	logger := h.logger.WithField("orderID", orderID)
	logger.Debug("WebSocket logs connection requested")

	conn, err := ws.SetupWebSocket(w, r, logger)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Monitor client disconnection by reading messages
	go func() {
		defer cancel()

		for {
			// Try to read a message - this will fail when client disconnects
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.WithError(err).Debug("WebSocket connection closed unexpectedly")
				} else {
					logger.Debug("WebSocket client disconnected")
				}
				return
			}
			// If we receive a message, reset the read deadline
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		}
	}()

	// Get log stream
	logs, err := h.deployer.GetDeploymentLogs(ctx, orderID)
	if err != nil {
		logger.WithError(err).Error("Failed to get deployment logs")
		if err := ws.SendErrorResponse(conn, fmt.Sprintf("failed to get logs: %s", err.Error())); err != nil {
			logger.WithError(err).Error("Failed to send error response")
		}
		return
	}
	// Defer cleanup of all log streams
	defer func() {
		for _, lg := range logs {
			if lg.Stream != nil {
				_ = lg.Stream.Close()
			}
		}
	}()

	logger.Debug("Starting log streaming")

	// Stream logs to client with improved resource management
	var scanners sync.WaitGroup

	// Use buffered channel to prevent blocking
	logch := make(chan types.ServiceLogMessage, len(logs)*10) // Buffer for each scanner
	donech := make(chan struct{})

	scanners.Add(len(logs))

	for _, lg := range logs {
		go func(name string, scan *bufio.Scanner) {
			defer scanners.Done()

			for scan.Scan() && ctx.Err() == nil {
				select {
				case logch <- types.ServiceLogMessage{
					Name:    name,
					Message: scan.Text(),
				}:
					// Successfully sent log
				case <-ctx.Done():
					// Context cancelled, stop scanning
					return
				}
			}
		}(lg.Name, lg.Scanner)
	}

	// Close donech when all scanners finish
	go func() {
		scanners.Wait()
		close(donech)
	}()

	pingTicker := time.NewTicker(ws.PingPeriod)
	defer pingTicker.Stop()

	// Main event loop
	for {
		select {
		case line := <-logch:
			if err = conn.WriteJSON(line); err != nil {
				logger.WithError(err).Error("Failed to write log to WebSocket")
				return
			}
		case <-pingTicker.C:
			if err = ws.SendPing(conn); err != nil {
				logger.WithError(err).Error("Failed to send ping")
				return
			}
			if err = conn.SetReadDeadline(time.Now().Add(ws.PongWait)); err != nil {
				logger.WithError(err).Error("Failed to set read deadline")
				return
			}
		case <-donech:
			return
		case <-ctx.Done():
			logger.Debug("Context cancelled, stopping log stream")
			return
		}
	}
}

// deploymentShellWebsocketHandler handles WebSocket messages for shell input
func (h *DeploymentHandler) deploymentShellWebsocketHandler(log *logrus.Entry, wg *sync.WaitGroup, shellWs *websocket.Conn, stdinPipeOut io.Writer, terminalSizeUpdate chan<- remotecommand.TerminalSize) {
	defer wg.Done()
	for {
		shellWs.SetPongHandler(func(string) error {
			return shellWs.SetReadDeadline(time.Now().Add(ws.PingWait))
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
		case DeploymentShellCodeStdin:
			if stdinPipeOut != nil {
				_, err = stdinPipeOut.Write(msg)
				if err != nil {
					return
				}
			}
		case DeploymentShellCodeResize:
			if terminalSizeUpdate != nil {
				if len(msg) == 8 {
					width := binary.BigEndian.Uint32(msg[0:4])
					height := binary.BigEndian.Uint32(msg[4:8])
					terminalSizeUpdate <- remotecommand.TerminalSize{
						Width:  uint16(width),
						Height: uint16(height),
					}
				}
			}
		}
	}
}

// deploymentShellPingHandler sends periodic pings to keep the WebSocket connection alive
func (h *DeploymentHandler) deploymentShellPingHandler(ctx context.Context, wg *sync.WaitGroup, conn *websocket.Conn) {
	defer wg.Done()
	ticker := time.NewTicker(ws.PingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := ws.SendPing(conn); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// channelToTerminalSizeQueue converts a channel to a TerminalSizeQueue
type channelToTerminalSizeQueue <-chan remotecommand.TerminalSize

// deploymentShellResponse represents the response from a shell execution
type deploymentShellResponse struct {
	ExitCode int    `json:"exit_code"`
	Message  string `json:"message,omitempty"`
}

// Next returns the next terminal size from the channel
func (sq channelToTerminalSizeQueue) Next() *remotecommand.TerminalSize {
	select {
	case size := <-sq:
		return &size
	default:
		return nil
	}
}

// sendErrorResponse sends a standardized error response
func (h *DeploymentHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"message": http.StatusText(statusCode),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
