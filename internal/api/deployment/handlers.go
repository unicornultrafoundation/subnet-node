package deployment

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

// healthHandler handles health check requests
func (s *HTTPServer) healthHandler(c *gin.Context) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "deployment-api",
	}
	c.JSON(http.StatusOK, response)
}

// createDeploymentHandler handles deployment creation requests
func (s *HTTPServer) createDeploymentHandler(c *gin.Context) {
	var req types.DeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendErrorResponse(c.Writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	userAddress, _ := c.Get(string(userAddressKey))
	if userAddressStr, ok := userAddress.(string); ok {
		req.Requester = userAddressStr
	}

	ctx := c.Request.Context()
	response, err := s.deployer.RequestDeployment(ctx, &req)

	if err != nil {
		s.sendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// getDeploymentHandler handles deployment retrieval requests
func (s *HTTPServer) getDeploymentHandler(c *gin.Context) {
	orderID := c.Param("orderID")
	ctx := c.Request.Context()
	response, err := s.deployer.GetDeployment(ctx, orderID)

	if err != nil {
		s.sendErrorResponse(c.Writer, err.Error(), http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, response)
}

// cleanupDeploymentHandler handles deployment cleanup requests
func (s *HTTPServer) cleanupDeploymentHandler(c *gin.Context) {
	orderID := c.Param("orderID")
	ctx := c.Request.Context()
	err := s.deployer.CleanupDeployment(ctx, orderID)

	if err != nil {
		s.sendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deployment cleaned up successfully"})
}

// getServiceStatusHandler handles service status requests
func (s *HTTPServer) getServiceStatusHandler(c *gin.Context) {
	orderID := c.Param("orderID")
	serviceName := c.Param("serviceName")
	ctx := c.Request.Context()
	response, err := s.deployer.GetServiceStatus(ctx, orderID, serviceName)

	if err != nil {
		s.sendErrorResponse(c.Writer, err.Error(), http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, response)
}

// getDeploymentLogsHandler handles deployment logs requests
func (s *HTTPServer) getDeploymentLogsHandler(c *gin.Context) {
	orderID := c.Param("orderID")
	ctx := c.Request.Context()
	response, err := s.deployer.GetDeploymentLogs(ctx, orderID)

	if err != nil {
		s.sendErrorResponse(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

// execHandler handles command execution requests
func (s *HTTPServer) execHandler(c *gin.Context) {
	orderID := c.Param("orderID")

	var execReq struct {
		PodName     string   `json:"podName"`
		ServiceName string   `json:"serviceName"`
		Command     []string `json:"command"`
		TTY         bool     `json:"tty"`
	}

	if err := c.ShouldBindJSON(&execReq); err != nil {
		s.sendErrorResponse(c.Writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Exec endpoint requires WebSocket connection for streaming",
		"orderID": orderID,
	})
}
