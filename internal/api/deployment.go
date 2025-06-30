package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

// DeploymentHandler provides HTTP endpoints for deployment management
type DeploymentHandler struct {
	deployer DeployerService
	logger   *logrus.Entry
}

// DeployerService interface for deployment operations
type DeployerService interface {
	RequestDeployment(ctx context.Context, req *types.DeploymentRequest) (*types.DeploymentResponse, error)
	GetDeployment(ctx context.Context, orderID string) (*types.DeploymentResponse, error)
	CleanupDeployment(ctx context.Context, orderID string) error
	GetDeployments(ctx context.Context, requester string) ([]*types.DeploymentResponse, error)
	GetDeploymentLogs(ctx context.Context, orderID string) ([]*types.ServiceLog, error)
	GetServiceStatus(ctx context.Context, orderID, serviceName string) (*types.ServiceStatus, error)
}

// NewDeploymentHandler creates a new deployment handler
func NewDeploymentHandler(deployer DeployerService) *DeploymentHandler {
	return &DeploymentHandler{
		deployer: deployer,
		logger:   logrus.WithField("component", "deployment-handler"),
	}
}

// Router returns the chi router with all deployment routes
func (h *DeploymentHandler) Router() *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint
	r.Get("/health", h.healthHandler)

	// Deployment management routes
	r.Post("/deployments", h.createDeploymentHandler)
	r.Get("/deployments", h.listDeploymentsHandler)
	r.Get("/deployments/{orderID}", h.getDeploymentHandler)
	r.Delete("/deployments/{orderID}", h.cleanupDeploymentHandler)

	// Service and pod management routes
	r.Get("/deployments/{orderID}/services/{serviceName}/status", h.getServiceStatusHandler)
	r.Get("/deployments/{orderID}/logs", h.getDeploymentLogsHandler)

	return r
}

// healthHandler handles health check requests
func (h *DeploymentHandler) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "deployment-api",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// createDeploymentHandler handles deployment creation requests
func (h *DeploymentHandler) createDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	var req types.DeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	response, err := h.deployer.RequestDeployment(ctx, &req)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// listDeploymentsHandler handles deployment list requests
func (h *DeploymentHandler) listDeploymentsHandler(w http.ResponseWriter, r *http.Request) {
	requester := r.URL.Query().Get("requester")
	if requester == "" {
		h.sendErrorResponse(w, "requester parameter is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	deployments, err := h.deployer.GetDeployments(ctx, requester)
	if err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(deployments)
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
