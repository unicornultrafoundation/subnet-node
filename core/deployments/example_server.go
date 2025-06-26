package deployments

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	bidenginetypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
	"github.com/unicornultrafoundation/subnet-node/config"
)

// SimpleMockBidMarketContract is a simple mock implementation for examples
type SimpleMockBidMarketContract struct{}

func (m *SimpleMockBidMarketContract) GetOrder(ctx context.Context, orderID *big.Int) (*bidenginetypes.Order, error) {
	return &bidenginetypes.Order{
		ID:                 orderID,
		Owner:              common.HexToAddress("0x123456789abcdef123456789abcdef123456789a"),
		Status:             bidenginetypes.OrderStatusOpen,
		AcceptedProviderId: big.NewInt(123),
		AcceptedMachineId:  big.NewInt(1),
		CpuCores:           big.NewInt(4),
		MinBidPrice:        big.NewInt(1000000000000000000),
		Duration:           big.NewInt(3600),
		ExpiredAt:          big.NewInt(time.Now().Add(24 * time.Hour).Unix()),
	}, nil
}

func (m *SimpleMockBidMarketContract) GetOrderCount(ctx context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *SimpleMockBidMarketContract) OrderCount(ctx context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *SimpleMockBidMarketContract) Orders(ctx context.Context, orderID *big.Int) (*bidenginetypes.Order, error) {
	return m.GetOrder(ctx, orderID)
}

func (m *SimpleMockBidMarketContract) GetBids(ctx context.Context, orderID *big.Int) ([]bidenginetypes.Bid, error) {
	return []bidenginetypes.Bid{}, nil
}

func (m *SimpleMockBidMarketContract) IsBiddingOpen(ctx context.Context, orderID *big.Int) (bool, error) {
	return true, nil
}

func (m *SimpleMockBidMarketContract) GetRemainingBidTime(ctx context.Context, orderID *big.Int) (*big.Int, error) {
	return big.NewInt(3600), nil
}

func (m *SimpleMockBidMarketContract) SubmitBid(ctx context.Context, orderID *big.Int, pricePerSecond *big.Int, providerID *big.Int, machineID *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *SimpleMockBidMarketContract) GetBidIndexFromTransaction(ctx context.Context, tx *ethtypes.Transaction, orderID *big.Int) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *SimpleMockBidMarketContract) CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *SimpleMockBidMarketContract) AcceptBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *SimpleMockBidMarketContract) CancelOrder(ctx context.Context, orderID *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *SimpleMockBidMarketContract) CloseOrder(ctx context.Context, orderID *big.Int, reason string) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *SimpleMockBidMarketContract) ExtendOrder(ctx context.Context, orderID *big.Int, amount *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *SimpleMockBidMarketContract) GetUsedResource(ctx context.Context, providerID *big.Int, machineID *big.Int) (*bidenginetypes.ResourceUsage, error) {
	return &bidenginetypes.ResourceUsage{}, nil
}

func (m *SimpleMockBidMarketContract) ReleaseOrderResource(ctx context.Context, orderID *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *SimpleMockBidMarketContract) WatchOrderCreated(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *SimpleMockBidMarketContract) WatchOrderClosed(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *SimpleMockBidMarketContract) WatchOrderExpired(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *SimpleMockBidMarketContract) WatchBidSubmitted(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *SimpleMockBidMarketContract) WatchBidAccepted(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *SimpleMockBidMarketContract) WatchBidCancelled(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *SimpleMockBidMarketContract) OrderBids(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*bidenginetypes.Bid, error) {
	return &bidenginetypes.Bid{}, nil
}

// SimpleMockServiceManager is a simple mock implementation for examples
type SimpleMockServiceManager struct{}

func (m *SimpleMockServiceManager) Start(ctx context.Context) error {
	return nil
}

func (m *SimpleMockServiceManager) Stop(ctx context.Context) error {
	return nil
}

func (m *SimpleMockServiceManager) CreateDeployment(ctx context.Context, deployment *Deployment) error {
	return nil
}

func (m *SimpleMockServiceManager) GetDeployment(ctx context.Context, id string) (*Deployment, error) {
	return nil, fmt.Errorf("deployment not found")
}

func (m *SimpleMockServiceManager) ListDeployments(ctx context.Context) ([]*Deployment, error) {
	return []*Deployment{}, nil
}

func (m *SimpleMockServiceManager) StartDeployment(ctx context.Context, id string) error {
	return nil
}

func (m *SimpleMockServiceManager) StopDeployment(ctx context.Context, id string) error {
	return nil
}

func (m *SimpleMockServiceManager) DeleteDeployment(ctx context.Context, id string) error {
	return nil
}

func (m *SimpleMockServiceManager) GetDeploymentLogs(ctx context.Context, deploymentID, serviceName string, tail int) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *SimpleMockServiceManager) StreamDeploymentLogs(ctx context.Context, deploymentID, serviceName string, follow bool) (<-chan LogEntry, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *SimpleMockServiceManager) ExecConsole(ctx context.Context, deploymentID, serviceName string, command []string, tty bool) (ExecSession, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *SimpleMockServiceManager) InspectDeployment(ctx context.Context, deploymentID string) (*DeploymentInspection, error) {
	return &DeploymentInspection{}, nil
}

func (m *SimpleMockServiceManager) InspectService(ctx context.Context, deploymentID, serviceName string) (*ServiceInspection, error) {
	return &ServiceInspection{}, nil
}

func (m *SimpleMockServiceManager) GetDeploymentMetrics(ctx context.Context, deploymentID string, duration time.Duration) (*DeploymentMetrics, error) {
	return &DeploymentMetrics{}, nil
}

func (m *SimpleMockServiceManager) GetServiceMetrics(ctx context.Context, deploymentID, serviceName string, duration time.Duration) (*ServiceMetrics, error) {
	return &ServiceMetrics{}, nil
}

func (m *SimpleMockServiceManager) UpdateDeploymentImage(ctx context.Context, deploymentID, serviceName, image string) error {
	return nil
}

// ExampleServer demonstrates how to use the deployment API server
func ExampleServer() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg := config.NewC(logger)

	// Get provider ID from config
	providerID := big.NewInt(int64(cfg.GetInt("deployment.provider_id", 123)))

	// Initialize mock bid market contract
	mockBidMarket := &SimpleMockBidMarketContract{}

	// Initialize service manager
	serviceManager := &SimpleMockServiceManager{}

	// Initialize API with provider validation
	api := NewAPI(serviceManager, mockBidMarket, logger, providerID)

	// Set up router
	router := mux.NewRouter()
	api.RegisterRoutes(router)

	// Add middleware
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	// Create server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting deployment API server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exited")
}

// ExampleServerWithCustomConfig demonstrates how to create a server with custom configuration
func ExampleServerWithCustomConfig() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create config
	cfg := config.NewC(logger)

	// Load configuration from string
	configYAML := `
deployment:
  api:
    server:
      host: "localhost"
      port: 9090
      read_timeout: "60s"
      write_timeout: "60s"
      idle_timeout: "300s"
      cors:
        enabled: true
        allowed_origins:
          - "http://localhost:3000"
      rate_limit:
        enabled: true
        requests_per_minute: 200
  provider_id: 123
  machine_ids:
    - "1"
    - "2"
    - "3"
`

	err := cfg.LoadString(configYAML)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create deployment service
	service := NewService(cfg, logger)

	// Create mock bid market contract
	bidMarket := &SimpleMockBidMarketContract{}

	// Create API server
	server := NewServer(cfg, service, bidMarket, logger)

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil {
			logger.WithError(err).Error("Server error")
		}
	}()

	// Wait for a moment to see the server start
	select {
	case <-ctx.Done():
		logger.Info("Server stopped")
	}
}

// ExampleServerWithEnvironmentVariables demonstrates how to use environment variables
func ExampleServerWithEnvironmentVariables() {
	// Set environment variables for configuration
	os.Setenv("DEPLOYMENT_API_SERVER_HOST", "0.0.0.0")
	os.Setenv("DEPLOYMENT_API_SERVER_PORT", "8080")
	os.Setenv("DEPLOYMENT_API_SERVER_CORS_ENABLED", "true")
	os.Setenv("DEPLOYMENT_API_SERVER_RATE_LIMIT_ENABLED", "true")

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create config
	cfg := config.NewC(logger)

	// Load configuration (will pick up environment variables)
	err := cfg.Load("./config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create deployment service
	service := NewService(cfg, logger)

	// Create mock bid market contract
	bidMarket := &SimpleMockBidMarketContract{}

	// Create API server
	server := NewServer(cfg, service, bidMarket, logger)

	// Get server configuration
	serverConfig := server.GetConfig()

	// Print some configuration values
	host := serverConfig.GetString("deployment.api.server.host", "unknown")
	port := serverConfig.GetInt("deployment.api.server.port", 0)
	corsEnabled := serverConfig.GetBool("deployment.api.server.cors.enabled", false)

	fmt.Printf("Server configured for %s:%d (CORS: %v)\n", host, port, corsEnabled)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil {
			logger.WithError(err).Error("Server error")
		}
	}()

	// Wait for shutdown
	<-ctx.Done()
}

// ExampleConfigurationStructure shows the expected configuration structure
func ExampleConfigurationStructure() {
	configExample := `
# Deployment API Server Configuration
deployment:
  api:
    server:
      # Basic server settings
      host: "0.0.0.0"
      port: 8080
      read_timeout: "30s"
      write_timeout: "30s"
      idle_timeout: "120s"
      
      # TLS configuration
      tls:
        enabled: false
        cert_file: "/path/to/cert.pem"
        key_file: "/path/to/key.pem"
      
      # CORS configuration
      cors:
        enabled: true
        allowed_origins:
          - "*"
          - "http://localhost:3000"
      
      # Rate limiting
      rate_limit:
        enabled: true
        requests_per_minute: 100

  # Provider and machine configuration
  provider_id: 123
  machine_ids:
    - "1"
    - "2"
    - "3"

  # Kubernetes configuration
  kubernetes:
    enabled: true
    kubeconfig: "/path/to/kubeconfig"
    namespace_prefix: "subnet-"
    resource_quotas:
      enabled: true
      default_cpu: "1000m"
      default_memory: "1Gi"
      default_storage: "10Gi"
  
  # Resource limits
  resource_limits:
    max_cpu_cores: 8
    max_memory_gb: 16
    max_disk_gb: 100
    max_containers: 10
    max_ports: 50
  
  # Network configuration
  network:
    default_driver: "bridge"
    subnet_pool:
      start: "172.16.0.0/12"
      end: "172.31.0.0/12"
    enable_ipv6: false

# Bid market configuration
bid_market:
  contract_address: "0x..."
  rpc_url: "http://localhost:8545"
  chain_id: 1
`

	fmt.Println("Example configuration structure:")
	fmt.Println(configExample)
}

// ExampleLoadConfigFromString demonstrates loading config from a string
func ExampleLoadConfigFromString() {
	logger := logrus.New()
	cfg := config.NewC(logger)

	// Load config from string
	configYAML := `
bidengine:
  provider_id: 123
  min_bid_percent: 70
  max_bid_percent: 95
contracts:
  provider: "0x123456789abcdef123456789abcdef123456789a"
  bid_market: "0xabcdef123456789abcdef123456789abcdef1234"
logging:
  level: "info"
`

	if err := cfg.LoadString(configYAML); err != nil {
		log.Fatal("Failed to load config from string:", err)
	}

	// Get provider ID
	providerID := big.NewInt(123) // From config above

	// Initialize API
	mockBidMarket := &SimpleMockBidMarketContract{}
	serviceManager := &SimpleMockServiceManager{}
	api := NewAPI(serviceManager, mockBidMarket, logger, providerID)

	fmt.Printf("API initialized with provider ID: %s\n", providerID.String())

	// Use api...
	_ = api
}

// ExampleLoadConfigFromEnvironment demonstrates loading config from environment variables
func ExampleLoadConfigFromEnvironment() {
	logger := logrus.New()

	// Set environment variables
	os.Setenv("BIDENGINE_PROVIDER_ID", "456")
	os.Setenv("BIDENGINE_MIN_BID_PERCENT", "75")
	os.Setenv("CONTRACTS_PROVIDER", "0x987654321fedcba987654321fedcba987654321f")

	// Load config from environment (you would need to implement this)
	// For now, we'll use the values directly
	providerID := big.NewInt(456)

	// Initialize API
	mockBidMarket := &SimpleMockBidMarketContract{}
	serviceManager := &SimpleMockServiceManager{}
	api := NewAPI(serviceManager, mockBidMarket, logger, providerID)

	fmt.Printf("API initialized with provider ID: %s\n", providerID.String())

	// Use api...
	_ = api
}

// Middleware functions
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.WithFields(logrus.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"duration":   time.Since(start),
				"user_agent": r.UserAgent(),
			}).Info("HTTP request")
		})
	}
}

func recoveryMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.WithField("error", err).Error("Panic recovered")
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
