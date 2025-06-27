package deployment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	authchain "github.com/unicornultrafoundation/subnet-node/crypto/authchain"
)

// Context key types to avoid string literal collisions
type contextKey string

const (
	userAddressKey contextKey = "userAddress"
	orderIDKey     contextKey = "orderID"
	providerIDKey  contextKey = "providerId"
	machineIDKey   contextKey = "machineId"
)

type Deployer interface {
	RequestDeployment(ctx context.Context, req *types.DeploymentRequest) (*types.DeploymentResponse, error)
	GetDeployment(ctx context.Context, orderID string) (*types.DeploymentResponse, error)
	CleanupDeployment(ctx context.Context, orderID string) error
	GetDeployments(ctx context.Context, requester string) ([]*types.DeploymentResponse, error)
	GetDeploymentLogs(ctx context.Context, orderID string) ([]*types.ServiceLog, error)
	GetServiceStatus(ctx context.Context, orderID, serviceName string) (*types.ServiceStatus, error)
}

// HTTPServer provides HTTP endpoints for deployment management
type HTTPServer struct {
	cfg      *config.C
	router   *gin.Engine
	logger   *logrus.Logger
	deployer Deployer
}

// NewHTTPServer creates a new HTTP server for deployment API
func NewHTTPServer(logger *logrus.Logger, port int, cfg *config.C, deployer Deployer) *HTTPServer {
	gin.SetMode(gin.ReleaseMode)
	server := &HTTPServer{
		router:   gin.New(),
		logger:   logger,
		cfg:      cfg,
		deployer: deployer,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all the HTTP routes with middleware
func (s *HTTPServer) setupRoutes() {
	// Health check endpoint (no auth required)
	s.router.GET("/health", s.healthHandler)

	// API routes with authorization middleware
	api := s.router.Group("/api/v1")
	api.Use(s.authMiddleware)
	api.Use(s.loggingMiddleware)

	// Deployment management routes
	api.POST("/deployments", s.createDeploymentHandler)
	api.GET("/deployments/:orderID", s.getDeploymentHandler)
	api.DELETE("/deployments/:orderID", s.cleanupDeploymentHandler)

	// Service and pod management routes
	api.GET("/deployments/:orderID/services/:serviceName/status", s.getServiceStatusHandler)
	api.GET("/deployments/:orderID/logs", s.getDeploymentLogsHandler)
	api.POST("/deployments/:orderID/exec", s.execHandler)
}

// Start starts the HTTP server
func (s *HTTPServer) Start(ctx context.Context) error {
	port := s.cfg.GetInt("deployment.http.port", 7650)

	s.logger.WithField("port", port).Info("Starting deployment HTTP server")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("HTTP server error")
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}

// authMiddleware checks authorization for API requests using AuthChain
func (s *HTTPServer) authMiddleware(c *gin.Context) {
	// Extract and validate authchain from X-AuthChain header
	authChainHeader := c.GetHeader("X-AuthChain")
	if authChainHeader == "" {
		s.sendErrorResponse(c.Writer, "X-AuthChain header required", http.StatusUnauthorized)
		c.Abort()
		return
	}

	// Decode base64 authchain
	authChain, err := authchain.DeserializeAuthChain(authChainHeader)
	if err != nil {
		s.sendErrorResponse(c.Writer, "Invalid authchain format: "+err.Error(), http.StatusUnauthorized)
		c.Abort()
		return
	}

	// Extract order ID from URL path
	orderID := c.Param("orderID")

	// For endpoints that don't have orderID in path, try to get from query params
	if orderID == "" {
		orderID = c.Query("orderID")
	}

	// For POST /api/v1/deployments, extract orderID from request body
	if orderID == "" && c.Request.Method == "POST" && strings.HasSuffix(c.Request.URL.Path, "/deployments") {
		// Read body to extract orderID
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil {
			var reqBody struct {
				OrderID string `json:"order_id"`
			}
			if err := json.Unmarshal(bodyBytes, &reqBody); err == nil {
				orderID = reqBody.OrderID
			}
			// Reset body for handlers to read again
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	// Extract providerId and machineId from config
	providerId := s.cfg.GetString("deployment.provider.id", "")
	machineId := s.cfg.GetString("deployment.machine.id", "")

	// Create entityID according to schema: "subnet_deployment:{providerId}:{machineId}:{order_id}"
	entityID := fmt.Sprintf("subnet_deployment:%s:%s:%s", providerId, machineId, orderID)

	// Validate authchain for the entityID with default 60s expiry
	// For business logic, we validate against the entityID format "subnet_deployment:provider:machine:order"
	// instead of the ephemeral address
	result, err := authchain.ValidateAuthChainWithDefaultExpiry(authChain, entityID)
	if err != nil || result == nil || !result.OK {
		// If validation fails, try to extract entityID from authchain and compare
		if extractedEntityID, extractErr := authchain.GetEntityID(authChain); extractErr == nil {
			if extractedEntityID == entityID {
				// EntityID matches, consider it valid
				result = &authchain.ValidationResult{OK: true}
			} else {
				errMsg := fmt.Sprintf("EntityID mismatch. Expected: %s, Got: %s", entityID, extractedEntityID)
				s.sendErrorResponse(c.Writer, errMsg, http.StatusUnauthorized)
				c.Abort()
				return
			}
		} else {
			errMsg := "Invalid authchain"
			if err != nil {
				errMsg = err.Error()
			} else if result != nil && result.Message != "" {
				errMsg = result.Message
			}
			s.sendErrorResponse(c.Writer, errMsg, http.StatusUnauthorized)
			c.Abort()
			return
		}
	}

	// Extract user address from authchain
	userAddress := ""
	if len(authChain) > 0 {
		// For SIGNER type, the payload contains the address
		if authChain[0].Type == authchain.AuthLinkTypeSIGNER {
			userAddress = authChain[0].Payload
		}
	}

	// Add user info to request context
	c.Set(string(userAddressKey), userAddress)
	c.Set(string(orderIDKey), orderID)
	c.Set(string(providerIDKey), providerId)
	c.Set(string(machineIDKey), machineId)
}

// loggingMiddleware logs all API requests
func (s *HTTPServer) loggingMiddleware(c *gin.Context) {
	start := time.Now()

	// Process request
	c.Next()

	duration := time.Since(start)

	s.logger.WithFields(logrus.Fields{
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"status":     c.Writer.Status(),
		"duration":   duration,
		"userAgent":  c.Request.UserAgent(),
		"remoteAddr": c.Request.RemoteAddr,
	}).Info("API request")
}

// Helper methods

func (s *HTTPServer) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (s *HTTPServer) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"message": http.StatusText(statusCode),
	}

	s.sendJSONResponse(w, response, statusCode)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
