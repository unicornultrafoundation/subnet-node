package deployments

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	bidenginetypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
	"github.com/unicornultrafoundation/subnet-node/config"
)

// ExampleServer demonstrates how to use the deployment API server
func ExampleServer() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create config
	cfg := config.NewC(logger)

	// Load configuration from file
	err := cfg.Load("./config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create deployment service
	service := NewService(cfg, logger)

	// Create mock bid market contract
	bidMarket := &mockBidMarket{}

	// Create API server
	server := NewServer(cfg, service, bidMarket, logger)

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.WithField("signal", sig).Info("Received shutdown signal")
		cancel()
	}()

	// Start the server
	logger.Info("Starting deployment API server")
	if err := server.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Server error")
	}
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
`

	err := cfg.LoadString(configYAML)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create deployment service
	service := NewService(cfg, logger)

	// Create mock bid market contract
	bidMarket := &mockBidMarket{}

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
	bidMarket := &mockBidMarket{}

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

// mockBidMarket is a mock implementation of the bid market contract
type mockBidMarket struct{}

func (m *mockBidMarket) GetOrder(ctx context.Context, orderID *big.Int) (*bidenginetypes.Order, error) {
	return &bidenginetypes.Order{
		ID:    orderID,
		Owner: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}, nil
}

func (m *mockBidMarket) GetOrderCount(ctx context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *mockBidMarket) OrderCount(ctx context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *mockBidMarket) Orders(ctx context.Context, orderID *big.Int) (*bidenginetypes.Order, error) {
	return m.GetOrder(ctx, orderID)
}

func (m *mockBidMarket) GetBids(ctx context.Context, orderID *big.Int) ([]bidenginetypes.Bid, error) {
	return []bidenginetypes.Bid{}, nil
}

func (m *mockBidMarket) IsBiddingOpen(ctx context.Context, orderID *big.Int) (bool, error) {
	return true, nil
}

func (m *mockBidMarket) GetRemainingBidTime(ctx context.Context, orderID *big.Int) (*big.Int, error) {
	return big.NewInt(3600), nil
}

func (m *mockBidMarket) SubmitBid(ctx context.Context, orderID *big.Int, pricePerSecond *big.Int, providerID *big.Int, machineID *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *mockBidMarket) GetBidIndexFromTransaction(ctx context.Context, tx *ethtypes.Transaction, orderID *big.Int) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *mockBidMarket) CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *mockBidMarket) AcceptBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *mockBidMarket) CancelOrder(ctx context.Context, orderID *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *mockBidMarket) CloseOrder(ctx context.Context, orderID *big.Int, reason string) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *mockBidMarket) ExtendOrder(ctx context.Context, orderID *big.Int, amount *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *mockBidMarket) GetUsedResource(ctx context.Context, providerID *big.Int, machineID *big.Int) (*bidenginetypes.ResourceUsage, error) {
	return &bidenginetypes.ResourceUsage{}, nil
}

func (m *mockBidMarket) ReleaseOrderResource(ctx context.Context, orderID *big.Int) (*ethtypes.Transaction, error) {
	return &ethtypes.Transaction{}, nil
}

func (m *mockBidMarket) WatchOrderCreated(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *mockBidMarket) WatchOrderClosed(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *mockBidMarket) WatchOrderExpired(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *mockBidMarket) WatchBidSubmitted(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *mockBidMarket) WatchBidAccepted(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *mockBidMarket) WatchBidCancelled(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	return nil
}

func (m *mockBidMarket) OrderBids(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*bidenginetypes.Bid, error) {
	return &bidenginetypes.Bid{}, nil
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
