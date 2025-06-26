package deployer

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	"github.com/unicornultrafoundation/subnet-node/internal/api/ws"
)

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

// HandleLogs handles WebSocket connections for log streaming
func (api *DeployerWSAPI) HandleLogs(w http.ResponseWriter, r *http.Request, orderID string) {
	logger := api.logger.WithField("orderID", orderID)
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
	logs, err := api.deployerService.GetDeploymentLogs(ctx, orderID)
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
