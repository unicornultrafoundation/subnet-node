package apps

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/gogo/protobuf/proto"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/errdefs"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipfs/go-datastore"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
)

var log = logrus.New().WithField("service", "apps")

const NAMESPACE = "subnet-apps"
const PROTOCOL_ID = protocol.ID("subnet-apps")
const RESOURCE_USAGE_KEY = "resource-usage-v2"

type Service struct {
	peerId           peer.ID
	IsProvider       bool
	cfg              *config.C
	ethClient        *ethclient.Client
	containerdClient *containerd.Client
	P2P              *p2p.P2P
	PeerHost         p2phost.Host `optional:"true"` // the network host (server+client)
	subnetID         big.Int
	stopChan         chan struct{} // Channel to stop background tasks
	accountService   *account.AccountService
	Datastore        datastore.Datastore // Datastore for storing resource usage
}

// Initializes the Service with Ethereum and containerd clients.
func New(peerHost p2phost.Host, peerId peer.ID, cfg *config.C, P2P *p2p.P2P, dataStore datastore.Datastore, accountService *account.AccountService) *Service {
	return &Service{
		peerId:         peerId,
		PeerHost:       peerHost,
		P2P:            P2P,
		cfg:            cfg,
		Datastore:      dataStore,
		IsProvider:     cfg.GetBool("provider.enable", false),
		stopChan:       make(chan struct{}),
		accountService: accountService,
		ethClient:      accountService.GetClient(),
	}
}

func (s *Service) Start(ctx context.Context) error {
	// Register the P2P protocol for signing
	if err := s.RegisterSignProtocol(); err != nil {
		return fmt.Errorf("failed to register signing protocol: %w", err)
	}

	if s.IsProvider {
		// Connect to containerd daemon
		var err error
		s.containerdClient, err = containerd.New("/run/containerd/containerd.sock")
		if err != nil {
			return fmt.Errorf("error connecting to containerd: %v", err)
		}

		s.RestartStoppedContainers(ctx)

		// Update latest resource usage into datastore
		if err := s.updateAllRunningContainersUsage(ctx); err != nil {
			return err
		}

		// Start app sub-services
		go s.startMonitoringUsage(ctx)
		go s.startRewardClaimer(ctx)
	}

	log.Info("Subnet Apps Service started successfully.")
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	log.Info("Stopping Subnet Apps Service...")

	// Close stopChan to stop all background tasks
	close(s.stopChan)

	// Close the containerd client
	if s.containerdClient != nil {
		err := s.containerdClient.Close()
		if err != nil {
			return fmt.Errorf("failed to close containerd client: %w", err)
		}
	}

	log.Info("Subnet Apps Service stopped successfully.")
	return nil
}

func (s *Service) GetAppCount() (*big.Int, error) {
	return s.accountService.AppStore().AppCount(nil)
}

func (s *Service) GetApp(ctx context.Context, appId *big.Int) (*App, error) {
	subnetApp, err := s.accountService.AppStore().Apps(nil, appId)
	if err != nil {
		return nil, err
	}

	appStatus, err := s.GetContainerStatus(ctx, appId)
	if err != nil {
		return nil, err
	}

	app := convertToApp(subnetApp, appId, appStatus)

	appConfig, err := s.GetContainerConfigProto(ctx, appId)
	if err == nil {
		app.Metadata.ContainerConfig.Env = appConfig.ContainerConfig.Env
	}

	// Retrieve metadata from datastore if available
	metadata, err := s.GetContainerConfigProto(ctx, appId)
	if err == nil {
		app.Metadata.ContainerConfig = metadata.ContainerConfig
	}

	if appStatus == Running {
		ip, err := s.GetContainerIP(ctx, appId)
		if err != nil {
			return nil, err
		}
		app.IP = ip
	}

	return app, nil
}

func (s *Service) GetContainerConfigProto(ctx context.Context, appId *big.Int) (*AppMetadata, error) {
	configKey := datastore.NewKey(fmt.Sprintf("container-config-%s", appId.String()))
	configData, err := s.Datastore.Get(ctx, configKey)
	if err != nil {
		return nil, err
	}

	var protoConfig pbapp.AppMetadata
	if err := proto.Unmarshal(configData, &protoConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal container config: %w", err)
	}

	return ProtoToAppMetadata(&protoConfig), nil
}

// Stops and removes the container for the specified app.
func (s *Service) RemoveApp(ctx context.Context, appId *big.Int) (*App, error) {
	// Set the namespace for the container
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

	// Retrieve app details from the Ethereum contract
	app, err := s.GetApp(ctx, appId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app details: %w", err)
	}

	// Load the container for the app
	container, err := s.containerdClient.LoadContainer(ctx, app.ContainerId())
	if err != nil {
		return nil, fmt.Errorf("failed to load container: %w", err)
	}

	// Stop the container task
	task, err := container.Task(ctx, nil)
	taskNotFound := false
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}
		taskNotFound = true
	}

	if !taskNotFound {
		_, err = task.Delete(ctx, containerd.WithProcessKill)
		if err != nil {
			return nil, fmt.Errorf("failed to delete task: %w", err)
		}
	}

	// Remove the container and its snapshot
	err = container.Delete(ctx, containerd.WithSnapshotCleanup)
	if err != nil {
		return nil, fmt.Errorf("failed to delete container: %w", err)
	}

	log.Printf("App %s removed successfully: Container ID: %s", app.Symbol, app.ContainerId())
	app.Status = NotFound
	return app, nil
}

// Retrieves the status of a container associated with a specific app.
func (s *Service) GetContainerStatus(ctx context.Context, appId *big.Int) (ProcessStatus, error) {
	// Set the namespace for the container
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

	app := App{ID: appId}

	// Load the container for the app
	container, err := s.containerdClient.LoadContainer(ctx, app.ContainerId())
	if err != nil {
		if errdefs.IsNotFound(err) {
			return NotFound, nil
		}
		return Unknown, err
	}

	// Retrieve the task for the container
	task, err := container.Task(ctx, nil)
	if err != nil {
		return Stopped, nil // Task does not exist
	}

	// Retrieve the status of the task
	status, err := task.Status(ctx)
	if err != nil {
		return Stopped, fmt.Errorf("failed to get task status: %w", err)
	}

	switch status.Status {
	case containerd.Created:
		return Created, nil
	case containerd.Paused:
		return Paused, nil
	case containerd.Running:
		return Running, nil
	case containerd.Stopped:
		return Stopped, nil
	case containerd.Pausing:
		return Pausing, nil
	default:
		return Unknown, nil
	}
}

func (s *Service) RegisterSignProtocol() error {
	// Define the signing logic
	signHandler := func(stream network.Stream) ([]byte, error) {
		signRequest, err := ReceiveSignRequest(stream)

		if err != nil {
			return []byte{}, fmt.Errorf("failed to receive sign request: %w", err)
		}

		var signature []byte
		switch data := signRequest.Data.(type) {
		case *pbapp.SignatureRequest_Usage:
			usage := convertUsageFromProto(*data.Usage)
			signature, err = s.SignResourceUsage(usage)

			if err != nil {
				return []byte{}, fmt.Errorf("failed to sign resource usage: %w", err)
			}
		default:
			return []byte{}, fmt.Errorf("unsupport this signature request type: %w", err)
		}

		return signature, nil
	}

	// Create the listener for the signing protocol
	listener := p2p.NewSignProtocolListener(PROTOCOL_ID, signHandler)

	// Register the listener in the ListenersP2P
	err := s.P2P.ListenersP2P.Register(listener)
	if err != nil {
		return err
	}

	log.Debugf("Registered signing protocol: %s", PROTOCOL_ID)
	return nil
}

func (s *Service) RestartStoppedContainers(ctx context.Context) error {
	// Set the namespace for the containers
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

	// Fetch all running containers
	containers, err := s.containerdClient.Containers(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch running containers: %w", err)
	}

	for _, container := range containers {
		// Get container ID (assuming appID is same as container ID)
		containerId := container.ID()
		appId, err := getAppIdFromContainerId(containerId)

		if err != nil {
			log.Errorf("failed to get appId from containerId %s: %v", containerId, err)
			continue
		}

		status, err := s.GetContainerStatus(ctx, appId)

		if err != nil {
			log.Errorf("failed to get status from containerId %s: %v", containerId, err)
			continue
		}

		if status == Stopped {
			err := s.RestartContainer(ctx, appId)

			if err != nil {
				log.Errorf("failed to restart containerId %s: %v", containerId, err)
				continue
			}
		}
	}

	return nil
}

func (s *Service) RestartContainer(ctx context.Context, appId *big.Int) error {
	_, err := s.RemoveApp(ctx, appId)

	if err != nil {
		return fmt.Errorf("failed to remove appId %v: %v", appId, err)
	}

	_, err = s.RunApp(ctx, appId)
	if err != nil {
		return fmt.Errorf("failed to run appId %v: %v", appId, err)
	}

	return nil
}

// // Calculate container duration
// func (s *Service) CalculateContainerDuration(ctx context.Context, appId big.Int) (int64, error) {
// 	// Load the container
// 	container, err := s.containerdClient.LoadContainer(ctx, appId.String())
// 	if err != nil {
// 		return 0, err // Return 0 if container not found
// 	}

// 	// Get the task associated with the container
// 	info, err := container.Info(ctx, nil)
// 	if err != nil {
// 		log.Errorf("Failed to get container task: %v", err)
// 		return 0, err // Return 0 if task not found
// 	}

// 	// Calculate duration (current time - start time)
// 	startTime := info.CreatedAt
// 	if startTime.IsZero() {
// 		return 0, nil // If no start time, duration is 0
// 	}
// 	duration := time.Since(startTime).Seconds()
// 	return int64(duration), nil // Return duration in seconds
// }

// // GetAppNode retrieves the details of an AppNode by its ID or similar identifier.
// func (s *Service) GetAppNode(ctx context.Context, appId big.Int, subnetId big.Int) (ResourceUsage, error) {
// 	// Prepare input for the contract call to retrieve AppNode details
// 	input, err := subnetABI.Pack("getAppNode", appId, subnetId)
// 	if err != nil {
// 		log.Errorf("Failed to pack input for getAppNode: %v", err)
// 		return ResourceUsage{}, err
// 	}

// 	// Create the contract call message
// 	msg := ethereum.CallMsg{
// 		To:   &s.subnetAppRegistryContract,
// 		Data: input,
// 	}

// 	// Execute the contract call
// 	result, err := s.ethClient.CallContract(ctx, msg, nil)
// 	if err != nil {
// 		log.Errorf("Failed to call contract for AppNode details: %v", err)
// 		return ResourceUsage{}, err
// 	}

// 	// Unpack the result into the AppNode structure
// 	var usage ResourceUsage
// 	err = subnetABI.UnpackIntoInterface(&usage, "getAppNode", result)
// 	if err != nil {
// 		log.Errorf("Failed to unpack AppNode details: %v", err)
// 		return ResourceUsage{}, err
// 	}

// 	// Return the retrieved AppNode details
// 	return usage, nil
// }

// Retrieves the IP address of a running container.
func (s *Service) GetContainerIP(ctx context.Context, appId *big.Int) (string, error) {
	// Use the netns package to enter the network namespace and get the IP address
	// This is a placeholder for the actual implementation
	ip := "127.0.0.1" // Replace with actual logic to retrieve the IP address

	return ip, nil
}

func (s *Service) SaveContainerConfigProto(ctx context.Context, appId *big.Int, config ContainerConfig) error {
	protoConfig := AppMetadataToProto(&AppMetadata{ContainerConfig: config})
	configData, err := proto.Marshal(protoConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal container config: %w", err)
	}

	configKey := datastore.NewKey(fmt.Sprintf("container-config-%s", appId.String()))
	if err := s.Datastore.Put(ctx, configKey, configData); err != nil {
		return fmt.Errorf("failed to save container config: %w", err)
	}

	return nil
}

// Extract appId from container ID
func getAppIdFromContainerId(containerId string) (*big.Int, error) {
	appIDStr := strings.TrimPrefix(containerId, "subnet-app-")
	appID, ok := new(big.Int).SetString(appIDStr, 10)

	if !ok {
		return nil, fmt.Errorf("invalid container ID: %s", containerId)
	}

	return appID, nil
}
