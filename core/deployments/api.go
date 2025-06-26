package deployments

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	bidenginetypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// API represents the deployment API server
type API struct {
	service   ServiceManager
	logger    *logrus.Logger
	bidMarket bidenginetypes.BidMarketContract

	// Rate limiting
	rateLimiter *RateLimiter

	// Provider configuration
	configProviderID *big.Int
}

// RateLimiter implements simple rate limiting per address
type RateLimiter struct {
	requests map[string][]time.Time
	mux      sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed for the given address
func (rl *RateLimiter) Allow(address string) bool {
	rl.mux.Lock()
	defer rl.mux.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get existing requests for this address
	requests, exists := rl.requests[address]
	if !exists {
		requests = []time.Time{}
	}

	// Remove old requests outside the window
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if we're under the limit
	if len(validRequests) >= rl.limit {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[address] = validRequests

	return true
}

// AuthorizationRequest represents the authorization data in request headers
type AuthorizationRequest struct {
	Signature string `json:"signature"`
	Address   string `json:"address"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// WebSocketError represents a WebSocket error message
type WebSocketError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// UpdateImageRequest represents a request to update deployment image
type UpdateImageRequest struct {
	Image string `json:"image"`
}

// RegisterRoutes registers the API routes
func (api *API) RegisterRoutes(router *mux.Router) {
	// Deployment management
	router.HandleFunc("/api/v1/deployments", api.createDeployment).Methods("POST")
	router.HandleFunc("/api/v1/deployments", api.listDeployments).Methods("GET")
	router.HandleFunc("/api/v1/deployments/{id}", api.getDeployment).Methods("GET")
	router.HandleFunc("/api/v1/deployments/{id}", api.updateDeployment).Methods("PUT")
	router.HandleFunc("/api/v1/deployments/{id}", api.deleteDeployment).Methods("DELETE")
	router.HandleFunc("/api/v1/deployments/{id}/start", api.startDeployment).Methods("POST")
	router.HandleFunc("/api/v1/deployments/{id}/stop", api.stopDeployment).Methods("POST")
	router.HandleFunc("/api/v1/deployments/{id}/restart", api.restartDeployment).Methods("POST")

	// Monitoring and inspection
	router.HandleFunc("/api/v1/deployments/{id}/inspect", api.inspectDeployment).Methods("GET")
	router.HandleFunc("/api/v1/deployments/{id}/services/{service}/inspect", api.inspectService).Methods("GET")
	router.HandleFunc("/api/v1/deployments/{id}/metrics", api.getDeploymentMetrics).Methods("GET")
	router.HandleFunc("/api/v1/deployments/{id}/services/{service}/metrics", api.getServiceMetrics).Methods("GET")

	// Logs
	router.HandleFunc("/api/v1/deployments/{id}/logs", api.getDeploymentLogs).Methods("GET")
	router.HandleFunc("/api/v1/deployments/{id}/services/{service}/logs", api.getServiceLogs).Methods("GET")
	router.HandleFunc("/api/v1/deployments/{id}/logs/stream", api.streamDeploymentLogs).Methods("GET")
	router.HandleFunc("/api/v1/deployments/{id}/services/{service}/logs/stream", api.streamServiceLogs).Methods("GET")

	// Console execution
	router.HandleFunc("/api/v1/deployments/{id}/services/{service}/exec", api.execConsole).Methods("POST")
	router.HandleFunc("/api/v1/deployments/{id}/services/{service}/exec/ws", api.execConsoleWebSocket).Methods("GET")

	// Events
	router.HandleFunc("/api/v1/deployments/{id}/events", api.getDeploymentEvents).Methods("GET")
	router.HandleFunc("/api/v1/deployments/{id}/events/stream", api.streamDeploymentEvents).Methods("GET")
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    int    `json:"code"`
}

// Helper functions

func (api *API) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (api *API) writeError(w http.ResponseWriter, status int, message string) {
	api.writeJSON(w, status, ErrorResponse{
		Success: false,
		Error:   message,
		Code:    status,
	})
}

func (api *API) getVars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// validateDeploymentAuthorization validates that the request is authorized by the deployment owner
func (api *API) validateDeploymentAuthorization(w http.ResponseWriter, r *http.Request, deploymentID string) (common.Address, bool) {
	// Get authorization data from headers
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		api.writeError(w, http.StatusUnauthorized, "Missing authorization header")
		return common.Address{}, false
	}

	// Parse authorization data
	var authReq AuthorizationRequest
	if err := json.Unmarshal([]byte(authHeader), &authReq); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid authorization format")
		return common.Address{}, false
	}

	// Validate timestamp (prevent replay attacks)
	now := time.Now().Unix()
	if abs(now-authReq.Timestamp) >= 300 { // 5 minutes tolerance (inclusive)
		api.writeError(w, http.StatusUnauthorized, "Authorization timestamp expired")
		return common.Address{}, false
	}

	// Parse the signer address
	signerAddr := common.HexToAddress(authReq.Address)
	if signerAddr == (common.Address{}) {
		api.writeError(w, http.StatusBadRequest, "Invalid signer address")
		return common.Address{}, false
	}

	// Verify signature first (local operation, fast)
	if !api.verifySignature(authReq.Message, authReq.Signature, signerAddr) {
		api.writeError(w, http.StatusUnauthorized, "Invalid signature")
		return common.Address{}, false
	}

	// Get deployment to check ownership
	deployment, err := api.service.GetDeployment(r.Context(), deploymentID)
	if err != nil {
		api.writeError(w, http.StatusNotFound, "Deployment not found")
		return common.Address{}, false
	}

	// Check if the signer is the deployment owner
	deploymentOwner := common.HexToAddress(deployment.Owner)
	if deploymentOwner == (common.Address{}) {
		api.writeError(w, http.StatusInternalServerError, "Deployment owner not set")
		return common.Address{}, false
	}

	if deploymentOwner != signerAddr {
		api.writeError(w, http.StatusForbidden, "Not authorized: not the deployment owner")
		return common.Address{}, false
	}

	return signerAddr, true
}

// verifySignature verifies the signature of a message
func (api *API) verifySignature(message, signature string, expectedAddress common.Address) bool {
	// Decode signature
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		api.logger.WithError(err).Error("Failed to decode signature")
		return false
	}

	// Ensure signature is 65 bytes (32 + 32 + 1)
	if len(sigBytes) != 65 {
		api.logger.Error("Invalid signature length")
		return false
	}

	// Ensure recovery ID is 0 or 1 for Go-Ethereum
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	// Prefix message for Ethereum personal_sign
	prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	messageHash := crypto.Keccak256Hash([]byte(prefixed))

	// Recover public key from signature
	pubKey, err := crypto.SigToPub(messageHash.Bytes(), sigBytes)
	if err != nil {
		api.logger.WithError(err).Error("Failed to recover public key from signature")
		return false
	}

	// Get address from public key
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	// Compare with expected address
	return recoveredAddr == expectedAddress
}

// abs returns the absolute value of an int64
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// NewAPI creates a new deployment API server
func NewAPI(service ServiceManager, bidMarket bidenginetypes.BidMarketContract, logger *logrus.Logger, configProviderID *big.Int) *API {
	return &API{
		service:          service,
		bidMarket:        bidMarket,
		logger:           logger,
		rateLimiter:      NewRateLimiter(5, time.Minute), // 5 requests per minute per address
		configProviderID: configProviderID,
	}
}

// Deployment management handlers

func (api *API) createDeployment(w http.ResponseWriter, r *http.Request) {
	var deployment Deployment
	if err := json.NewDecoder(r.Body).Decode(&deployment); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get authorization data from headers
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		api.writeError(w, http.StatusUnauthorized, "Missing authorization header")
		return
	}

	// Parse authorization data
	var authReq AuthorizationRequest
	if err := json.Unmarshal([]byte(authHeader), &authReq); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid authorization format")
		return
	}

	// Validate timestamp
	now := time.Now().Unix()
	if abs(now-authReq.Timestamp) >= 300 {
		api.writeError(w, http.StatusUnauthorized, "Authorization timestamp expired")
		return
	}

	// Parse the signer address
	signerAddr := common.HexToAddress(authReq.Address)
	if signerAddr == (common.Address{}) {
		api.writeError(w, http.StatusBadRequest, "Invalid signer address")
		return
	}

	// Check rate limiting
	if !api.rateLimiter.Allow(signerAddr.Hex()) {
		api.writeError(w, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
		return
	}

	// Verify signature
	if !api.verifySignature(authReq.Message, authReq.Signature, signerAddr) {
		api.writeError(w, http.StatusUnauthorized, "Invalid signature")
		return
	}

	// For new deployments, validate ownership from blockchain
	orderID, ok := new(big.Int).SetString(deployment.ID, 10)
	if !ok {
		api.writeError(w, http.StatusBadRequest, "Invalid deployment ID format")
		return
	}

	// Get order from bid market to validate ownership and existence
	order, err := api.bidMarket.GetOrder(r.Context(), orderID)
	if err != nil {
		api.logger.WithError(err).Error("Failed to get order from bid market")
		api.writeError(w, http.StatusNotFound, "Order not found or invalid")
		return
	}

	// Check if the signer is the order owner
	if order.Owner != signerAddr {
		api.writeError(w, http.StatusForbidden, "Not authorized: not the order owner")
		return
	}

	// Validate order state to prevent spam
	if !api.isOrderValidForDeployment(order) {
		api.writeError(w, http.StatusBadRequest, "Order is not in a valid state for deployment")
		return
	}

	// Check if deployment already exists to prevent duplicate creation
	existingDeployment, err := api.service.GetDeployment(r.Context(), deployment.ID)
	if err == nil && existingDeployment != nil {
		api.writeError(w, http.StatusConflict, "Deployment already exists")
		return
	}

	// Set the deployment owner
	deployment.Owner = signerAddr.Hex()

	if err := api.service.CreateDeployment(r.Context(), &deployment); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusCreated, Response{Success: true, Data: deployment})
}

// isOrderValidForDeployment checks if an order is in a valid state for deployment creation
func (api *API) isOrderValidForDeployment(order *bidenginetypes.Order) bool {
	// Check if order is open (ready for deployment)
	if order.Status != bidenginetypes.OrderStatusOpen {
		return false
	}

	// Check if order has not expired
	if order.ExpiredAt != nil && time.Now().Unix() > order.ExpiredAt.Int64() {
		return false
	}

	// Check if order has valid resource requirements
	if order.CpuCores == nil || order.CpuCores.Sign() <= 0 {
		return false
	}

	// Check if order has valid pricing
	if order.MinBidPrice == nil || order.MinBidPrice.Sign() <= 0 {
		return false
	}

	// Check if order has valid duration
	if order.Duration == nil || order.Duration.Sign() <= 0 {
		return false
	}

	return true
}

func (api *API) listDeployments(w http.ResponseWriter, r *http.Request) {
	deployments, err := api.service.ListDeployments(r.Context())
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: deployments})
}

func (api *API) getDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	deployment, err := api.service.GetDeployment(r.Context(), deploymentID)
	if err != nil {
		api.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: deployment})
}

func (api *API) startDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	if err := api.service.StartDeployment(r.Context(), deploymentID); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true})
}

func (api *API) stopDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	if err := api.service.StopDeployment(r.Context(), deploymentID); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true})
}

func (api *API) deleteDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	if err := api.service.DeleteDeployment(r.Context(), deploymentID); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true})
}

// Monitoring and inspection handlers

func (api *API) inspectDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	inspection, err := api.service.InspectDeployment(r.Context(), deploymentID)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: inspection})
}

func (api *API) inspectService(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]
	serviceName := vars["service"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	inspection, err := api.service.InspectService(r.Context(), deploymentID, serviceName)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: inspection})
}

func (api *API) getDeploymentMetrics(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "1h"
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid duration format")
		return
	}

	metrics, err := api.service.GetDeploymentMetrics(r.Context(), deploymentID, duration)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: metrics})
}

func (api *API) getServiceMetrics(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]
	serviceName := vars["service"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "1h"
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid duration format")
		return
	}

	metrics, err := api.service.GetServiceMetrics(r.Context(), deploymentID, serviceName, duration)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: metrics})
}

// Logs handlers

func (api *API) getDeploymentLogs(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	tailStr := r.URL.Query().Get("tail")
	tail := 100 // default
	if tailStr != "" {
		if t, err := strconv.Atoi(tailStr); err == nil {
			tail = t
		}
	}

	logs, err := api.service.GetDeploymentLogs(r.Context(), deploymentID, "", tail)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer logs.Close()

	w.Header().Set("Content-Type", "text/plain")
	io.Copy(w, logs)
}

func (api *API) getServiceLogs(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]
	serviceName := vars["service"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	tailStr := r.URL.Query().Get("tail")
	tail := 100 // default
	if tailStr != "" {
		if t, err := strconv.Atoi(tailStr); err == nil {
			tail = t
		}
	}

	logs, err := api.service.GetDeploymentLogs(r.Context(), deploymentID, serviceName, tail)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer logs.Close()

	w.Header().Set("Content-Type", "text/plain")
	io.Copy(w, logs)
}

func (api *API) streamDeploymentLogs(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	follow := r.URL.Query().Get("follow") == "true"

	logChan, err := api.service.StreamDeploymentLogs(r.Context(), deploymentID, "", follow)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		api.writeError(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	for logEntry := range logChan {
		data, _ := json.Marshal(logEntry)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
}

func (api *API) streamServiceLogs(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]
	serviceName := vars["service"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	follow := r.URL.Query().Get("follow") == "true"

	logChan, err := api.service.StreamDeploymentLogs(r.Context(), deploymentID, serviceName, follow)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		api.writeError(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	for logEntry := range logChan {
		data, _ := json.Marshal(logEntry)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
}

// Console execution handlers

type ExecRequest struct {
	Command []string `json:"command"`
	TTY     bool     `json:"tty"`
}

func (api *API) execConsole(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]
	serviceName := vars["service"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	var req ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session, err := api.service.ExecConsole(r.Context(), deploymentID, serviceName, req.Command, req.TTY)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer session.Close()

	result, err := session.Execute(r.Context(), req.Command)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: result})
}

func (api *API) execConsoleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]
	serviceName := vars["service"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for now - in production, you should implement proper CORS
			return true
		},
		EnableCompression: true,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		api.logger.WithError(err).Error("Failed to upgrade connection to WebSocket")
		return
	}
	defer conn.Close()

	// Create exec session
	session, err := api.service.ExecConsole(r.Context(), deploymentID, serviceName, []string{"/bin/bash"}, true)
	if err != nil {
		api.logger.WithError(err).Error("Failed to create exec session")
		conn.WriteJSON(WebSocketError{
			Type:    "error",
			Message: "Failed to create exec session: " + err.Error(),
		})
		return
	}
	defer session.Close()

	// Create channels for communication
	stdinChan := make(chan string, 100)
	stdoutChan := make(chan string, 100)
	stderrChan := make(chan string, 100)
	errorChan := make(chan error, 100)
	doneChan := make(chan bool)

	// Start reading from WebSocket
	go func() {
		defer func() {
			doneChan <- true
		}()

		for {
			var msg WebSocketMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					api.logger.WithError(err).Error("WebSocket read error")
				}
				return
			}

			switch msg.Type {
			case "stdin":
				if msg.Data != nil {
					if data, ok := msg.Data.(string); ok {
						stdinChan <- data
					}
				}
			case "resize":
				// Handle terminal resize if needed
				api.logger.Debug("Terminal resize request received")
			case "close":
				return
			default:
				api.logger.WithField("type", msg.Type).Warn("Unknown WebSocket message type")
			}
		}
	}()

	// Start interactive execution
	go func() {
		defer func() {
			doneChan <- true
		}()

		// Create pipes for stdin, stdout, stderr
		stdinReader, stdinWriter := io.Pipe()
		stdoutReader, stdoutWriter := io.Pipe()
		stderrReader, stderrWriter := io.Pipe()

		// Start reading from stdin channel and writing to pipe
		go func() {
			defer stdinWriter.Close()
			for data := range stdinChan {
				stdinWriter.Write([]byte(data))
			}
		}()

		// Start reading from stdout pipe and sending to WebSocket
		go func() {
			defer stdoutReader.Close()
			buffer := make([]byte, 1024)
			for {
				n, err := stdoutReader.Read(buffer)
				if err != nil {
					if err != io.EOF {
						errorChan <- err
					}
					return
				}
				if n > 0 {
					stdoutChan <- string(buffer[:n])
				}
			}
		}()

		// Start reading from stderr pipe and sending to WebSocket
		go func() {
			defer stderrReader.Close()
			buffer := make([]byte, 1024)
			for {
				n, err := stderrReader.Read(buffer)
				if err != nil {
					if err != io.EOF {
						errorChan <- err
					}
					return
				}
				if n > 0 {
					stderrChan <- string(buffer[:n])
				}
			}
		}()

		// Execute the command interactively
		err := session.ExecuteInteractive(r.Context(), []string{"/bin/bash"}, stdinReader, stdoutWriter, stderrWriter)
		if err != nil {
			errorChan <- err
		}
	}()

	// Main loop for sending data to WebSocket
	for {
		select {
		case data := <-stdoutChan:
			err := conn.WriteJSON(WebSocketMessage{
				Type: "stdout",
				Data: data,
			})
			if err != nil {
				api.logger.WithError(err).Error("Failed to send stdout to WebSocket")
				return
			}
		case data := <-stderrChan:
			err := conn.WriteJSON(WebSocketMessage{
				Type: "stderr",
				Data: data,
			})
			if err != nil {
				api.logger.WithError(err).Error("Failed to send stderr to WebSocket")
				return
			}
		case err := <-errorChan:
			api.logger.WithError(err).Error("Exec session error")
			conn.WriteJSON(WebSocketError{
				Type:    "error",
				Message: "Exec session error: " + err.Error(),
			})
			return
		case <-doneChan:
			// Connection closed or session ended
			return
		case <-r.Context().Done():
			// Request context cancelled
			return
		}
	}
}

// Events handlers

func (api *API) getDeploymentEvents(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	// TODO: Add deployment events logic
	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: []DeploymentEvent{}})
}

func (api *API) streamDeploymentEvents(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	// TODO: Add deployment events streaming logic
	api.writeError(w, http.StatusNotImplemented, "Event streaming not implemented")
}

func (api *API) updateDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	var deployment Deployment
	if err := json.NewDecoder(r.Body).Decode(&deployment); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// TODO: Add deployment update logic
	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: deployment})
}

func (api *API) restartDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	// Validate authorization
	if _, authorized := api.validateDeploymentAuthorization(w, r, deploymentID); !authorized {
		return
	}

	// TODO: Add deployment restart logic
	api.writeJSON(w, http.StatusOK, Response{Success: true})
}
