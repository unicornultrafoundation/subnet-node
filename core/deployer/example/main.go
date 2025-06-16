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
	"github.com/unicornultrafoundation/subnet-node/core/deployer"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
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

	// Create store directory for payment manager
	storeDir := "./store"
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		logger.Fatal("Failed to create store directory", zap.Error(err))
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
	config := &deployer.ServiceConfig{
		KubeConfig:      kubeconfig,
		MaxRetries:      3,
		ProviderAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		IPFSURL:         "mock://", // Use mock IPFS client
		EthEndpoint:     "",        // Empty string will use mock Ethereum client
		ContractAddress: "",        // Empty string will use mock contract
		StoreDir:        storeDir,  // Set store directory for payment manager
		ClusterResources: struct {
			CPU     int64
			Memory  string
			Storage string
			GPU     int64
		}{
			CPU:     2,     // 2 CPU cores
			Memory:  "2Gi", // 2 GB memory
			Storage: "0Gi", // No storage requirement
			GPU:     0,     // No GPU by default
		},
		Pricing: struct {
			MemPriceMin      int64
			MemPriceMax      int64
			BidPriceStrategy string
			BidCPUScale      float64
			BidStorageScale  float64
			ProcessLimit     int
			ProcessTimeout   int
		}{
			MemPriceMin:      1000,  // Minimum memory price
			MemPriceMax:      10000, // Maximum memory price
			BidPriceStrategy: "fixed",
			BidCPUScale:      1.0, // CPU price scale
			BidStorageScale:  1.0, // Storage price scale
			ProcessLimit:     100, // Process limit
			ProcessTimeout:   30,  // Process timeout in seconds
		},
	}

	logger.Debug("Created service configuration",
		zap.String("storeDir", storeDir),
		zap.String("providerAddress", config.ProviderAddress.Hex()))

	// Create service instance
	service, err := deployer.NewService(config, nil)
	if err != nil {
		logger.Fatal("Failed to create service", zap.Error(err))
	}

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
	providerAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")

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

	// Add a mock bid for the provider address to avoid nil dereference
	_ = service.GetBidTracker().AddBid(ctx, "default", providerAddress, big.NewInt(1), 24*time.Hour)
	// Set the SDLHash for the mock bid
	if bid, _ := service.GetBidTracker().GetBid(ctx, "default", providerAddress); bid != nil {
		bid.SDLHash = sdlHash
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
	maxRetries = 6 // Increase max retries to 30 seconds (5s * 6)
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
	maxServiceRetries := 60 // Increase max retries to 5 minutes
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

		allPodsReady := true
		for _, pod := range pods {
			logger.Info("Pod status",
				zap.String("name", pod.Name),
				zap.String("status", pod.Status),
				zap.Bool("ready", pod.Ready))
			if !pod.Ready {
				allPodsReady = false
			}
		}

		if !allPodsReady {
			serviceRetryCount++
			if serviceRetryCount >= maxServiceRetries {
				logger.Fatal("Timeout waiting for all pods to be ready")
			}
			logger.Warn("Not all pods are ready, retrying...", zap.Int("retry", serviceRetryCount))
			continue
		}

		// Try to access the service
		if serviceURL == "" {
			// Get service URL for the echo service
			serviceName := "default-group-0-service-0"
			// Always try the global service name first
			url, err := service.GetServiceURL(ctx, "default", serviceName+"-global")
			if err != nil {
				// Fallback to the regular service name
				url, err = service.GetServiceURL(ctx, "default", serviceName)
				if err != nil {
					logger.Error("Failed to get service URL",
						zap.String("service", serviceName),
						zap.Error(err))
					serviceRetryCount++
					if serviceRetryCount >= maxServiceRetries {
						logger.Fatal("Failed to get service URL after retries")
					}
					logger.Warn("Service is not accessible, retrying...",
						zap.Int("retry", serviceRetryCount))
					continue
				}
			}

			logger.Info("Service is accessible",
				zap.String("service", serviceName),
				zap.String("url", url))

			// Add a small delay before trying to access the service
			time.Sleep(2 * time.Second)

			logger.Info("Attempting to access service",
				zap.String("service", serviceName),
				zap.String("url", url))

			// Create a custom HTTP client with timeout
			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			resp, err := client.Get(url)
			if err != nil {
				logger.Error("Failed to access service",
					zap.String("service", serviceName),
					zap.Error(err))
				serviceRetryCount++
				if serviceRetryCount >= maxServiceRetries {
					logger.Fatal("Failed to access service after retries")
				}
				logger.Warn("Service is not responding, retrying...",
					zap.Int("retry", serviceRetryCount))
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				logger.Error("Service returned non-200 status code",
					zap.String("service", serviceName),
					zap.Int("status", resp.StatusCode))
				serviceRetryCount++
				if serviceRetryCount >= maxServiceRetries {
					logger.Fatal("Service returned non-200 status code after retries")
				}
				logger.Warn("Service returned non-200 status code, retrying...",
					zap.Int("retry", serviceRetryCount))
				continue
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Failed to read response body",
					zap.String("service", serviceName),
					zap.Error(err))
				serviceRetryCount++
				if serviceRetryCount >= maxServiceRetries {
					logger.Fatal("Failed to read response body after retries")
				}
				logger.Warn("Failed to read response body, retrying...",
					zap.Int("retry", serviceRetryCount))
				continue
			}

			logger.Info("Service response",
				zap.String("service", serviceName),
				zap.String("body", string(body)))

			logger.Info("Service is running successfully!")
			logger.Info("Pausing for 5 seconds to allow inspection of resources...")
			time.Sleep(5 * time.Second)
			return
		}
	}
}
