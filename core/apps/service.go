package apps

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/unicornultrafoundation/subnet-node/config"
)

type Service struct {
	ethClient                 *ethclient.Client
	subnetAppRegistryContract common.Address
	containerdClient          *containerd.Client
}

// Initializes the Service with Ethereum and containerd clients.
func New(cfg *config.C) (*Service, error) {
	// Connect to Ethereum RPC
	client, err := ethclient.Dial(cfg.GetString("apps.rpc", ""))
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum node: %v", err)
		return nil, err
	}

	// Connect to containerd daemon
	containerdClient, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		fmt.Printf("Error connecting to containerd: %v\n", err)
		return nil, err
	}

	return &Service{
		ethClient:                 client,
		containerdClient:          containerdClient,
		subnetAppRegistryContract: common.HexToAddress(cfg.GetString("apps.subnet_app_registry_contract", "")),
	}, nil
}

// Retrieves a list of apps from the Ethereum contract and checks their container status.
func (s *Service) GetApps(ctx context.Context, start big.Int, end big.Int) ([]App, error) {
	// Prepare data for the listApps function call
	input, err := subnetABI.Pack("listApps", start, end)
	if err != nil {
		log.Fatalf("Failed to pack data: %v", err)
	}

	// Create a contract call message
	msg := ethereum.CallMsg{
		To:   &s.subnetAppRegistryContract,
		Data: input,
	}

	// Execute the contract call
	result, err := s.ethClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Fatalf("Failed to call contract: %v", err)
	}

	// Decode the result into a list of apps
	var apps []App
	err = subnetABI.UnpackIntoInterface(&apps, "listApps", result)
	if err != nil {
		log.Fatalf("Failed to unpack result: %v", err)
	}

	// Check the container status for each app
	for i, app := range apps {
		status, err := s.GetContainerStatus(ctx, app.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get container status for app %s: %w", app.Name, err)
		}
		apps[i].Status = status
	}

	return apps, nil
}

// Retrieves details of a specific app from the Ethereum contract and checks its container status.
func (s *Service) GetApp(ctx context.Context, appId big.Int) (App, error) {
	// Prepare data for the getApp function call
	input, err := subnetABI.Pack("getApp", appId)
	if err != nil {
		log.Fatalf("Failed to pack data: %v", err)
		return App{}, err
	}

	// Create a contract call message
	msg := ethereum.CallMsg{
		To:   &s.subnetAppRegistryContract,
		Data: input,
	}

	// Execute the contract call
	result, err := s.ethClient.CallContract(ctx, msg, nil)
	if err != nil {
		log.Fatalf("Failed to call contract: %v", err)
		return App{}, err
	}

	// Decode the result into an app
	var app App
	err = subnetABI.UnpackIntoInterface(&app, "getApp", result)
	if err != nil {
		log.Fatalf("Failed to unpack result: %v", err)
		return App{}, err
	}

	// Check the container status for the app
	status, err := s.GetContainerStatus(ctx, app.ID)
	if err != nil {
		return App{}, fmt.Errorf("failed to get container status: %w", err)
	}
	app.Status = status

	return app, nil
}

// Starts a container for the specified app using containerd.
func (s *Service) RunApp(ctx context.Context, appId big.Int) (App, error) {
	// Set the namespace for the container
	ctx = namespaces.WithNamespace(ctx, "subnet-apps")

	// Retrieve app details from the Ethereum contract
	app, err := s.GetApp(ctx, appId)
	if err != nil {
		return App{}, fmt.Errorf("failed to fetch app details: %w", err)
	}

	// Pull the image for the app
	image, err := s.containerdClient.Pull(ctx, app.PeerId, containerd.WithPullUnpack)
	if err != nil {
		return App{}, fmt.Errorf("failed to pull image: %w", err)
	}

	// Create a new container for the app
	container, err := s.containerdClient.NewContainer(
		ctx,
		app.ID.String(),
		containerd.WithImage(image),
		containerd.WithNewSnapshot(app.Symbol+"-snapshot", image),
		containerd.WithNewSpec(
			oci.WithMemoryLimit(app.MinMemory), // Minimum RAM
			oci.WithCPUShares(app.MinCpu),      // Minimum CPU
		),
	)
	if err != nil {
		return App{}, fmt.Errorf("failed to create container: %w", err)
	}

	// Start the container
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return App{}, fmt.Errorf("failed to create task: %w", err)
	}
	err = task.Start(ctx)
	if err != nil {
		return App{}, fmt.Errorf("failed to start task: %w", err)
	}

	log.Printf("App %s started successfully: Container ID: %s", app.Name, container.ID())
	return app, nil
}

// Stops and removes the container for the specified app.
func (s *Service) RemoveApp(ctx context.Context, appId big.Int) (App, error) {
	// Set the namespace for the container
	ctx = namespaces.WithNamespace(ctx, "subnet-apps")

	// Retrieve app details from the Ethereum contract
	app, err := s.GetApp(ctx, appId)
	if err != nil {
		return App{}, fmt.Errorf("failed to fetch app details: %w", err)
	}

	// Load the container for the app
	container, err := s.containerdClient.LoadContainer(ctx, app.ID.String())
	if err != nil {
		return App{}, fmt.Errorf("failed to load container: %w", err)
	}

	// Stop the container task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return App{}, fmt.Errorf("failed to get task: %w", err)
	}
	_, err = task.Delete(ctx, containerd.WithProcessKill)
	if err != nil {
		return App{}, fmt.Errorf("failed to delete task: %w", err)
	}

	// Remove the container and its snapshot
	err = container.Delete(ctx, containerd.WithSnapshotCleanup)
	if err != nil {
		return App{}, fmt.Errorf("failed to delete container: %w", err)
	}

	log.Printf("App %s removed successfully: Container ID: %s", app.Symbol, app.ID.String())
	return app, nil
}

// Retrieves the status of a container associated with a specific app.
func (s *Service) GetContainerStatus(ctx context.Context, appId big.Int) (containerd.ProcessStatus, error) {
	// Set the namespace for the container
	ctx = namespaces.WithNamespace(ctx, "subnet-apps")

	// Load the container for the app
	container, err := s.containerdClient.LoadContainer(ctx, appId.String())
	if err != nil {
		return containerd.Unknown, nil // Container does not exist
	}

	// Retrieve the task for the container
	task, err := container.Task(ctx, nil)
	if err != nil {
		return containerd.Stopped, nil // Task does not exist
	}

	// Retrieve the status of the task
	status, err := task.Status(ctx)
	if err != nil {
		return containerd.Stopped, fmt.Errorf("failed to get task status: %w", err)
	}

	return status.Status, nil
}
