package api

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/deployer"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

const (
	// Time allowed writing the file to the client.
	pingWait = 15 * time.Second

	// Time allowed reading the next pong message from the client.
	pongWait = 15 * time.Second

	// Send pings to a client with this period. Must be less than pongWait.
	pingPeriod = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Rename DeployerWS to API for go-ethereum compatibility
type DeployerWSAPI struct {
	deployerService *deployer.Service
	logger          *logrus.Logger
}

func NewDeployerWSAPI(deployerService *deployer.Service) *DeployerWSAPI {
	return &DeployerWSAPI{
		deployerService: deployerService,
		logger:          logrus.New(),
	}
}

// GetLogsHandler returns an http.HandlerFunc that extracts orderID using chi path parameters.
func (api *DeployerWSAPI) GetLogsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "orderID")
		if orderID == "" {
			http.Error(w, "missing orderID", http.StatusBadRequest)
			return
		}

		api.HandleLogs(w, r, orderID)
	}
}

// Add a helper to setup the websocket connection
func setupWebSocket(w http.ResponseWriter, r *http.Request, logger *logrus.Entry) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		// Return a proper error response instead of letting the handler continue
		http.Error(w, fmt.Sprintf("WebSocket upgrade failed: %v", err), http.StatusBadRequest)
		return nil, err
	}

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pingWait))
	})

	logger.Debug("WebSocket connection established")

	// Set up connection with proper read deadline
	conn.SetReadLimit(512)                                 // Limit message size
	conn.SetReadDeadline(time.Now().Add(30 * time.Second)) // Initial read deadline

	return conn, nil
}

// HandleLogs handles WebSocket connections for log streaming
func (api *DeployerWSAPI) HandleLogs(w http.ResponseWriter, r *http.Request, orderID string) {
	logger := api.logger.WithField("orderID", orderID)
	logger.Debug("WebSocket logs connection requested")

	conn, err := setupWebSocket(w, r, logger)
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
	logs, err := api.deployerService.GetDeploymentLogs(ctx, orderID)
	if err != nil {
		logger.WithError(err).Error("Failed to get deployment logs")
		errorMsg := fmt.Sprintf(`{"error":"failed to get logs: %s"}`, err.Error())
		conn.WriteMessage(websocket.TextMessage, []byte(errorMsg))
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

	pingTicker := time.NewTicker(pingPeriod)
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
			if err = conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				logger.WithError(err).Error("Failed to send ping")
				return
			}
			if err = conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
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
