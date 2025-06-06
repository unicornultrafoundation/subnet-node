package k8scluster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/bid"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/crd"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/deployment"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/events"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/ipfs"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/payment"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/session"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// ServiceConfig represents service configuration
type ServiceConfig struct {
	ProviderAddress common.Address
	DeploymentDir   string
	MaxRetries      int
	KubeConfig      *rest.Config
	EthEndpoint     string
	IPFSURL         string
}

// Service represents the marketplace service
type Service struct {
	config        *ServiceConfig
	deploymentMgr types.DeploymentManagerInterface
	eventBus      *events.DefaultEventBus[types.MarketplaceEvent]
	paymentMgr    payment.PaymentManagerInterface
	logger        *zap.Logger
	bidTracker    *bid.BidTracker
	client        *kubernetes.Clientset
	session       *session.Session
	sdlParser     *manifest.Parser
	ipfsClient    types.IPFSClient
	stopCh        chan struct{}
}

// NewService creates a new marketplace service
func NewService(config *ServiceConfig) (*Service, error) {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create Kubernetes client
	k8sClient, err := kubernetes.NewForConfig(config.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create event bus
	eventBus := events.NewEventBus[types.MarketplaceEvent]()
	if err := eventBus.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start event bus: %w", err)
	}

	// Create IPFS client
	ipfsClient := ipfs.NewClient(config.IPFSURL)

	// Create dynamic client for CRD
	dynamicClient, err := dynamic.NewForConfig(config.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create CRD client
	crdClient := crd.NewClient(dynamicClient)

	// Create bid tracker
	bidTracker := bid.NewBidTracker()

	// Create service instance
	s := &Service{
		config:     config,
		logger:     logger,
		stopCh:     make(chan struct{}),
		eventBus:   eventBus.(*events.DefaultEventBus[types.MarketplaceEvent]),
		client:     k8sClient,
		bidTracker: bidTracker,
		ipfsClient: ipfsClient,
	}

	// Create session
	session := session.New()
	session.SetProviderAddress(config.ProviderAddress)
	s.session = session

	// Create deployment manager
	deploymentMgr := deployment.NewDeploymentManager(
		logger,
		k8sClient,
		&types.DeploymentManagerConfig{
			DeploymentDir: config.DeploymentDir,
			MaxRetries:    config.MaxRetries,
		},
		eventBus.(*events.DefaultEventBus[types.MarketplaceEvent]),
		crdClient,
		session,
	)
	s.deploymentMgr = deploymentMgr

	// Initialize payment manager based on configuration
	if config.EthEndpoint == "" {
		// Use mock payment manager for local/demo runs
		mockEventBus := events.NewEventBus[interface{}]()
		if err := mockEventBus.Start(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to start mock event bus: %w", err)
		}
		s.paymentMgr = payment.NewMockPaymentManager(mockEventBus.(*events.DefaultEventBus[interface{}]))
	} else {
		// Use real payment manager with Ethereum
		paymentMgr, err := payment.NewPaymentManager(&payment.Config{
			ProviderAddr:  config.ProviderAddress,
			StoreDir:      config.DeploymentDir,
			CheckInterval: time.Minute,
		}, eventBus.(*events.DefaultEventBus[types.MarketplaceEvent]))
		if err != nil {
			return nil, fmt.Errorf("failed to create payment manager: %w", err)
		}
		s.paymentMgr = paymentMgr
	}

	// Register event handlers for all major event types
	if err := s.Subscribe(); err != nil {
		return nil, fmt.Errorf("failed to subscribe to events: %w", err)
	}

	return s, nil
}

// Start starts the marketplace service
func (s *Service) Start(ctx context.Context) error {
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
				s.logger.Error("failed to check health", zap.Error(err))
			}
		}
	}
}

// checkHealth checks the health of the marketplace service
func (s *Service) checkHealth() error {
	// Check deployment manager health
	if !s.deploymentMgr.IsHealthy() {
		return fmt.Errorf("deployment manager is not healthy")
	}

	// Check payment manager health
	if !s.paymentMgr.IsHealthy() {
		return fmt.Errorf("payment manager is not healthy")
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
