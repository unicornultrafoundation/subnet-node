package k8scluster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/bid"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/crd"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/deployment"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/events"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/ipfs"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/marketplace"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/payment"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/session"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// Service represents the marketplace service
type Service struct {
	config         *ServiceConfig
	ethClient      *ethclient.Client
	accountService *account.AccountService
	deploymentMgr  types.DeploymentManagerInterface
	eventBus       *events.DefaultEventBus[types.MarketplaceEvent]
	paymentMgr     payment.PaymentManagerInterface
	logger         *zap.Logger
	bidTracker     *bid.BidTracker
	bidManager     *marketplace.BidManager
	client         *kubernetes.Clientset
	session        *session.Session
	sdlParser      *manifest.Parser
	ipfsClient     types.IPFSClient
	stopCh         chan struct{}
	contract       types.ContractInterface
}

// NewService creates a new marketplace service
func NewService(config *ServiceConfig) (*Service, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create Kubernetes client
	client, err := kubernetes.NewForConfig(config.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create event bus
	eventBus := events.NewEventBus[types.MarketplaceEvent]()
	if err := eventBus.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start event bus: %w", err)
	}

	// Create IPFS client
	var ipfsClient types.IPFSClient
	if config.IPFSURL == "mock://" {
		ipfsClient = ipfs.NewMockClient()
	} else {
		ipfsClient = ipfs.NewClient(config.IPFSURL)
	}

	// Create bid tracker
	bidTracker := bid.NewBidTracker(logger)

	// Create session
	session := session.New()
	session.SetProviderAddress(config.ProviderAddress)

	// Create service
	service := &Service{
		config:     config,
		logger:     logger,
		client:     client,
		eventBus:   eventBus.(*events.DefaultEventBus[types.MarketplaceEvent]),
		bidTracker: bidTracker,
		session:    session,
		ipfsClient: ipfsClient,
		stopCh:     make(chan struct{}),
	}

	// Create deployment manager
	deploymentMgr, err := deployment.NewDeploymentManager(
		logger,
		client,
		&types.DeploymentManagerConfig{
			MaxRetries: config.MaxRetries,
		},
		eventBus.(*events.DefaultEventBus[types.MarketplaceEvent]),
		crd.NewClient(dynamicClient),
		session,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment manager: %w", err)
	}
	service.deploymentMgr = deploymentMgr

	// Create Ethereum client and contract
	var ethClient *ethclient.Client
	var contract types.ContractInterface

	if config.EthEndpoint == "" {
		// Use mock contract for local/demo runs
		contract = marketplace.NewMockMarketplaceContract(logger)
	} else {
		ethClient, err = ethclient.Dial(config.EthEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to create Ethereum client: %w", err)
		}
		service.ethClient = ethClient

		contractAddr := common.HexToAddress(config.ContractAddress)
		contract, err = marketplace.NewMarketplaceContract(
			ethClient,
			contractAddr,
			bidTracker,
			nil, // Payment manager will be set later
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create marketplace contract: %w", err)
		}
	}
	service.contract = contract

	// Create payment manager
	paymentConfig := &payment.Config{
		StoreDir:      config.StoreDir,
		CheckInterval: time.Minute,
	}
	defaultEventBus, ok := eventBus.(*events.DefaultEventBus[types.MarketplaceEvent])
	if !ok {
		return nil, fmt.Errorf("failed to convert event bus to default event bus")
	}
	paymentManager, err := payment.NewPaymentManager(contract, defaultEventBus, paymentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment manager: %w", err)
	}
	service.paymentMgr = paymentManager

	// Create bid manager
	priceConfig := &payment.PricingConfig{
		MemPriceMin:      1000000000000000,  // 0.001 ETH
		MemPriceMax:      10000000000000000, // 0.01 ETH
		BidPriceStrategy: "dynamic",
		BidCPUScale:      1.5,
		BidStorageScale:  1.2,
		ProcessLimit:     10,
		ProcessTimeout:   30,
	}
	service.bidManager = marketplace.NewBidManager(
		deploymentMgr,
		config.ProviderAddress,
		contract,
		ethClient,
		logger,
		priceConfig,
		ipfsClient,
	)

	// Create SDL parser
	service.sdlParser = manifest.NewParser()

	// Register event handlers
	if err := service.Subscribe(); err != nil {
		return nil, fmt.Errorf("failed to subscribe to events: %w", err)
	}

	return service, nil
}

// Start starts the marketplace service
func (s *Service) Start(ctx context.Context) error {
	// Setup cluster resources if needed
	if err := s.setupClusterResources(ctx); err != nil {
		return fmt.Errorf("failed to setup cluster resources: %w", err)
	}

	// Verify Kubernetes permissions
	if err := s.verifyKubernetesPermissions(ctx); err != nil {
		return fmt.Errorf("failed to verify Kubernetes permissions: %w", err)
	}

	// Start deployment manager
	if err := s.deploymentMgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start deployment manager: %w", err)
	}

	// Start payment manager
	if err := s.paymentMgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start payment manager: %w", err)
	}

	// Start marketplace service
	go s.run(ctx)

	return nil
}

// Stop stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping service...")

	// Signal stop
	close(s.stopCh)

	// Stop payment manager
	if s.paymentMgr != nil {
		s.paymentMgr.Stop()
	}

	// Stop event bus
	if s.eventBus != nil {
		if err := s.eventBus.Stop(context.Background()); err != nil {
			s.logger.Error("Failed to stop event bus", zap.Error(err))
		}
	}

	// Stop deployment manager last
	if s.deploymentMgr != nil {
		s.deploymentMgr.Stop()
	}

	s.logger.Info("Service stopped")
}

// run runs the service
func (s *Service) run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// Create metrics directory if it doesn't exist
	metricsDir := filepath.Join("deployments", "default", "metrics")
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		s.logger.Error("failed to create metrics directory", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("service stopped due to context cancellation")
			s.Stop()
			return
		case <-s.stopCh:
			s.logger.Info("service stopped due to stop signal")
			return
		case <-ticker.C:
			// Check health
			if err := s.checkHealth(); err != nil {
				s.logger.Error("health check failed", zap.Error(err))
				// Consider stopping the service if health check fails repeatedly
				continue
			}
		}
	}
}

// checkHealth checks the health of all components
func (s *Service) checkHealth() error {
	// Check deployment manager health
	if !s.deploymentMgr.IsHealthy() {
		return fmt.Errorf("deployment manager is unhealthy")
	}

	// Check payment manager health
	if !s.paymentMgr.IsHealthy() {
		return fmt.Errorf("payment manager is unhealthy")
	}

	// Check Ethereum client connection
	if s.ethClient != nil {
		if _, err := s.ethClient.BlockNumber(context.Background()); err != nil {
			return fmt.Errorf("ethereum client is unhealthy: %w", err)
		}
	}

	// Check IPFS client connection
	if s.ipfsClient != nil {
		// Try to get a test hash to verify connection
		_, err := s.ipfsClient.Get("QmTest")
		if err == nil {
			return fmt.Errorf("IPFS client is unhealthy: %w", err)
		}
	}

	return nil
}

// GetEventBus returns the event bus
func (s *Service) GetEventBus() *events.DefaultEventBus[types.MarketplaceEvent] {
	return s.eventBus
}

// GetIPFSClient returns the IPFS client
func (s *Service) GetIPFSClient() types.IPFSClient {
	return s.ipfsClient
}

// SetIPFSClient sets the IPFS client
func (s *Service) SetIPFSClient(ipfsClient types.IPFSClient) {
	s.ipfsClient = ipfsClient
}

// GetBidManager returns the bid manager
func (s *Service) GetBidManager() *marketplace.BidManager {
	return s.bidManager
}

// GetContract returns the marketplace contract
func (s *Service) GetContract() types.ContractInterface {
	return s.contract
}

// GetClient returns the Ethereum client
func (s *Service) GetClient() *ethclient.Client {
	return s.ethClient
}

// GetBidTracker returns the bid tracker
func (s *Service) GetBidTracker() *bid.BidTracker {
	return s.bidTracker
}
