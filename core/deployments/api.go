package deployments

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// API represents the deployment API server
type API struct {
	service ServiceManager
	logger  *logrus.Logger
}

// NewAPI creates a new deployment API server
func NewAPI(service ServiceManager, logger *logrus.Logger) *API {
	return &API{
		service: service,
		logger:  logger,
	}
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

	// Scaling and updates
	router.HandleFunc("/api/v1/deployments/{id}/services/{service}/scale", api.scaleDeployment).Methods("POST")
	router.HandleFunc("/api/v1/deployments/{id}/services/{service}/image", api.updateDeploymentImage).Methods("PUT")

	// Tenant management
	router.HandleFunc("/api/v1/tenants", api.createTenant).Methods("POST")
	router.HandleFunc("/api/v1/tenants", api.listTenants).Methods("GET")
	router.HandleFunc("/api/v1/tenants/{id}", api.getTenant).Methods("GET")
	router.HandleFunc("/api/v1/tenants/{id}", api.updateTenant).Methods("PUT")
	router.HandleFunc("/api/v1/tenants/{id}", api.deleteTenant).Methods("DELETE")
	router.HandleFunc("/api/v1/tenants/{id}/resources", api.getTenantResourceUsage).Methods("GET")

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

// Deployment management handlers

func (api *API) createDeployment(w http.ResponseWriter, r *http.Request) {
	var deployment Deployment
	if err := json.NewDecoder(r.Body).Decode(&deployment); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := api.service.CreateDeployment(r.Context(), &deployment); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusCreated, Response{Success: true, Data: deployment})
}

func (api *API) listDeployments(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	deployments, err := api.service.ListDeployments(r.Context(), tenantID)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: deployments})
}

func (api *API) getDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

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

	if err := api.service.StartDeployment(r.Context(), deploymentID); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true})
}

func (api *API) stopDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

	if err := api.service.StopDeployment(r.Context(), deploymentID); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true})
}

func (api *API) deleteDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]

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
	// WebSocket implementation for interactive console
	// This would require WebSocket library like gorilla/websocket
	api.writeError(w, http.StatusNotImplemented, "WebSocket exec not implemented yet")
}

// Scaling and updates handlers

type ScaleRequest struct {
	Replicas int `json:"replicas"`
}

func (api *API) scaleDeployment(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]
	serviceName := vars["service"]

	var req ScaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := api.service.ScaleDeployment(r.Context(), deploymentID, serviceName, req.Replicas); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true})
}

type UpdateImageRequest struct {
	Image string `json:"image"`
}

func (api *API) updateDeploymentImage(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	deploymentID := vars["id"]
	serviceName := vars["service"]

	var req UpdateImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := api.service.UpdateDeploymentImage(r.Context(), deploymentID, serviceName, req.Image); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true})
}

// Tenant management handlers

func (api *API) createTenant(w http.ResponseWriter, r *http.Request) {
	var tenant Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		api.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := api.service.CreateTenant(r.Context(), &tenant); err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusCreated, Response{Success: true, Data: tenant})
}

func (api *API) listTenants(w http.ResponseWriter, r *http.Request) {
	// This would need to be implemented in the service
	api.writeError(w, http.StatusNotImplemented, "List tenants not implemented yet")
}

func (api *API) getTenant(w http.ResponseWriter, r *http.Request) {
	// This would need to be implemented in the service
	api.writeError(w, http.StatusNotImplemented, "Get tenant not implemented yet")
}

func (api *API) updateTenant(w http.ResponseWriter, r *http.Request) {
	// This would need to be implemented in the service
	api.writeError(w, http.StatusNotImplemented, "Update tenant not implemented yet")
}

func (api *API) deleteTenant(w http.ResponseWriter, r *http.Request) {
	// This would need to be implemented in the service
	api.writeError(w, http.StatusNotImplemented, "Delete tenant not implemented yet")
}

func (api *API) getTenantResourceUsage(w http.ResponseWriter, r *http.Request) {
	vars := api.getVars(r)
	tenantID := vars["id"]

	usage, err := api.service.GetTenantResourceUsage(r.Context(), tenantID)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.writeJSON(w, http.StatusOK, Response{Success: true, Data: usage})
}

// Events handlers

func (api *API) getDeploymentEvents(w http.ResponseWriter, r *http.Request) {
	// This would need to be implemented in the service
	api.writeError(w, http.StatusNotImplemented, "Get deployment events not implemented yet")
}

func (api *API) streamDeploymentEvents(w http.ResponseWriter, r *http.Request) {
	// This would need to be implemented in the service
	api.writeError(w, http.StatusNotImplemented, "Stream deployment events not implemented yet")
}

// Placeholder handlers for unimplemented methods

func (api *API) updateDeployment(w http.ResponseWriter, r *http.Request) {
	api.writeError(w, http.StatusNotImplemented, "Update deployment not implemented yet")
}

func (api *API) restartDeployment(w http.ResponseWriter, r *http.Request) {
	api.writeError(w, http.StatusNotImplemented, "Restart deployment not implemented yet")
}
