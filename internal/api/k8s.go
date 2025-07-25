package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	apclient "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/provider/client"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/apitypes"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl"
	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
	provider "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/provider/v1"
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
		// Deployment management routes
		r.Post("/api/v1/deployments", h.createDeploymentHandler)
		r.Get("/api/v1/status", h.getStatusHandler)
		r.Get("/api/v1/leases", h.getAllLeaseStatusHandler)
		r.Delete("/api/v1/leases/{owner}/{dseq}", h.deleteDeploymentHandler)
		r.Get("/api/v1/leases/{owner}/{dseq}", h.getLeaseStatusHandler)
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
