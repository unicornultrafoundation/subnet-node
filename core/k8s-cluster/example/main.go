package main

import (
	"context"
	"flag"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	k8scluster "github.com/unicornultrafoundation/subnet-node/core/k8s-cluster"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Parse command line flags
	sdlPath := flag.String("sdl", "echo-service.yaml", "Path to SDL file")
	kubeconfigPath := flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "Path to kubeconfig file")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Initialize logger with debug level if verbose is enabled
	var logger *zap.Logger
	var err error
	if *verbose {
		config := zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		logger, err = config.Build()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting deployment process with verbose logging",
		zap.String("sdl", *sdlPath),
		zap.String("kubeconfig", *kubeconfigPath),
		zap.Bool("verbose", *verbose))

	// Load Kubernetes configuration
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfigPath)
	if err != nil {
		logger.Fatal("Failed to build kubeconfig", zap.Error(err))
	}

	// Create deployment directory
	deploymentDir := "./deployments"
	if err := os.MkdirAll(deploymentDir, 0755); err != nil {
		logger.Fatal("Failed to create deployment directory", zap.Error(err))
	}

	// Parse SDL file
	parser := manifest.NewParser()
	sdl, err := parser.ParseFile(*sdlPath)
	if err != nil {
		logger.Fatal("Failed to parse SDL file", zap.Error(err))
	}

	// Validate SDL
	if err := sdl.Validate(); err != nil {
		logger.Fatal("Invalid SDL file", zap.Error(err))
	}

	// Create service configuration
	config := &k8scluster.ServiceConfig{
		KubeConfig:      kubeconfig,
		DeploymentDir:   deploymentDir,
		MaxRetries:      3,
		ProviderAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		IPFSURL:         "mock://", // Use mock IPFS client
	}

	logger.Debug("Created service configuration",
		zap.String("deploymentDir", deploymentDir),
		zap.String("providerAddress", config.ProviderAddress.Hex()))

	// Create service instance
	service, err := k8scluster.NewService(config)
	if err != nil {
		logger.Fatal("Failed to create service", zap.Error(err))
	}

	// Create mock IPFS client & set it to the service
	mockClient := NewMockClient()
	service.SetIPFSClient(mockClient)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Start service
	logger.Info("Starting service...")
	if err := service.Start(ctx); err != nil {
		logger.Fatal("Failed to start service", zap.Error(err))
	}
	defer func() {
		// Cleanup deployment on exit
		if err := service.StopDeployment(ctx, "default"); err != nil {
			// Log warning instead of error since this is expected during shutdown
			logger.Warn("Failed to cleanup deployment", zap.Error(err))
		}
		// Add a small delay to ensure resources are properly terminated
		time.Sleep(2 * time.Second)
		// Stop the service after deployments are stopped
		service.Stop()
	}()

	requesterAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	providerAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")

	// Create channels for events
	bidSubmittedCh := make(chan *types.BidSubmittedEvent, 1)
	deploymentCompletedCh := make(chan *types.DeploymentCompletedEvent, 1)
	deploymentFailedCh := make(chan error, 1)

	// Subscribe to events
	if err := service.GetEventBus().Subscribe(string(types.MarketplaceEventTypeBidSubmitted), func(ctx context.Context, event types.MarketplaceEvent) error {
		if e, ok := event.(*types.BidSubmittedEvent); ok {
			logger.Info("Bid submitted", zap.String("deploymentID", e.DeploymentID), zap.String("provider", e.Provider.Hex()))
			bidSubmittedCh <- e
		}
		return nil
	}); err != nil {
		logger.Fatal("Failed to subscribe to bid submitted events", zap.Error(err))
	}
	if err := service.GetEventBus().Subscribe(string(types.MarketplaceEventTypeDeploymentCompleted), func(ctx context.Context, event types.MarketplaceEvent) error {
		if e, ok := event.(*types.DeploymentCompletedEvent); ok {
			logger.Info("Deployment completed", zap.String("deploymentID", e.DeploymentID))
			deploymentCompletedCh <- e
		}
		return nil
	}); err != nil {
		logger.Fatal("Failed to subscribe to deployment completed events", zap.Error(err))
	}

	// Add SDL content to IPFS
	sdlContent, err := os.ReadFile(*sdlPath)
	if err != nil {
		logger.Fatal("Failed to read SDL file", zap.Error(err))
	}
	sdlHash, err := service.GetIPFSClient().Add(sdlContent)
	if err != nil {
		logger.Fatal("Failed to add SDL to IPFS", zap.Error(err))
	}

	// Simulate deployment request
	logger.Info("Simulating user submitting deployment request...")
	err = service.GetEventBus().Publish(ctx, string(types.MarketplaceEventTypeDeploymentRequested), &types.DeploymentRequestedEvent{
		BaseEvent: types.BaseEvent{
			Timestamp: time.Now(),
		},
		DeploymentID: "default",
		Requester:    requesterAddress,
		SDLHash:      sdlHash,
		MaxPrice:     big.NewInt(0),
	})

	if err != nil {
		logger.Fatal("Failed to publish deployment requested event", zap.Error(err))
	}

	// Wait for bid submitted
	logger.Info("Waiting for bid submitted...")
	select {
	case <-bidSubmittedCh:
		logger.Info("Bid submitted successfully")
	case <-time.After(60 * time.Second):
		logger.Fatal("Timeout waiting for bid submitted")
	}

	// Simulate service provider selected with retry logic
	maxRetries := 3
	for retry := 0; retry < maxRetries; retry++ {
		logger.Info("Simulating service provider selected...", zap.Int("retry", retry+1))
		err = service.GetEventBus().Publish(ctx, string(types.MarketplaceEventTypeProviderSelected), &types.ProviderSelectedEvent{
			BaseEvent: types.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: "default",
			Provider:     providerAddress,
			Amount:       big.NewInt(0),
		})

		if err == nil {
			break
		}

		logger.Warn("Failed to publish provider selected event, retrying...",
			zap.Error(err),
			zap.Int("retry", retry+1),
			zap.Int("maxRetries", maxRetries))

		if retry < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(retry+1))
		}
	}

	if err != nil {
		logger.Fatal("Failed to publish provider selected event after retries", zap.Error(err))
	}

	// Wait for deployment to be completed
	logger.Info("Waiting for deployment to be completed...")
	retryCount := 0
	maxRetries = 30 // Increase max retries
	for {
		select {
		case <-deploymentCompletedCh:
			logger.Info("Deployment completed successfully")
			goto deploymentCompleted
		case err := <-deploymentFailedCh:
			logger.Fatal("Deployment failed", zap.Error(err))
		case <-time.After(5 * time.Second):
			retryCount++
			if retryCount >= maxRetries {
				logger.Fatal("Timeout waiting for deployment to complete")
			}
			logger.Info("Waiting for deployment to complete...", zap.Int("retry", retryCount))
		}
	}

deploymentCompleted:
	// Monitor deployment status
	logger.Info("Monitoring deployment status...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Wait for service to be accessible
	maxServiceRetries := 30 // Increase max retries for service access
	serviceRetryCount := 0
	serviceURL := ""

	for range ticker.C {
		// Get deployment status
		deployment, err := service.GetDeployment(ctx, "default")
		if err != nil {
			logger.Error("Failed to get deployment status", zap.Error(err))
			continue
		}

		logger.Info("Deployment status",
			zap.String("id", deployment.ID),
			zap.String("status", string(deployment.Status)),
			zap.Time("createdAt", deployment.CreatedAt),
			zap.Time("updatedAt", deployment.UpdatedAt))

		// Get pod status
		pods, err := service.GetDeploymentPods(ctx, "default")
		if err != nil {
			logger.Error("Failed to get pod status", zap.Error(err))
			continue
		}

		for _, pod := range pods {
			logger.Info("Pod status",
				zap.String("name", pod.Name),
				zap.String("status", pod.Status),
				zap.Bool("ready", pod.Ready))
		}

		// Try to access the service
		if serviceURL == "" {
			// Get service URL
			serviceURL, err = service.GetServiceURL(ctx, "default", "default-service-0")
			if err != nil {
				// Try the global service name
				serviceURL, err = service.GetServiceURL(ctx, "default", "default-service-0-global")
				if err != nil {
					logger.Error("Failed to get service URL", zap.Error(err))
					continue
				}
			}
			logger.Info("Service is accessible", zap.String("url", serviceURL))
		}

		if serviceURL != "" {
			// Try to access the service
			resp, err := http.Get(serviceURL)
			if err != nil {
				serviceRetryCount++
				if serviceRetryCount >= maxServiceRetries {
					logger.Fatal("Failed to access service after retries", zap.Error(err))
				}
				logger.Warn("Failed to fetch service, retrying...",
					zap.Error(err),
					zap.Int("retry", serviceRetryCount))
				continue
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Failed to read response body", zap.Error(err))
				continue
			}

			logger.Info("Service response", zap.String("body", string(body)))
			logger.Info("Deployment is running successfully!")
			logger.Info("Pausing for 5 seconds to allow inspection of resources...")
			time.Sleep(5 * time.Second)
			return
		}
	}
}
