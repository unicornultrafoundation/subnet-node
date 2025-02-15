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
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/apps/stats"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
)

var log = logrus.WithField("service", "apps")

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
	PeerHost         p2phost.Host  `optional:"true"` // the network host (server+client)
	stopChan         chan struct{} // Channel to stop background tasks
	accountService   *account.AccountService
	statService      *stats.Stats
	Datastore        datastore.Datastore // Datastore for storing resource usage
	verifier         *Verifier           // Verifier for resource usage
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
		verifier:       NewVerifier(),
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
		s.statService = stats.NewStats(s.containerdClient)

		s.RestartStoppedContainers(ctx)
		s.upgradeAppVersion(ctx)

		// Start app sub-services
		go s.startUpgradeAppVersion(ctx)
		go s.statService.Start()
		go s.startReportLoop(ctx)
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

	// Close sub-services
	s.statService.Stop()

	log.Info("Subnet Apps Service stopped successfully.")
	return nil
}

func (s *Service) GetAppCount() (*big.Int, error) {
	return s.accountService.AppStore().AppCount(nil)
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

// Retrieves the status of a container associated with a specific app.
func (s *Service) GetContainerStatus(ctx context.Context, appId *big.Int) (ProcessStatus, error) {
	// Set the namespace for the container
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

	containerId := getContainerIdFromAppId(appId)

	// Load the container for the app
	container, err := s.containerdClient.LoadContainer(ctx, containerId)
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

// Extract containerId from appId
func getContainerIdFromAppId(appId *big.Int) string {
	app := App{ID: appId}

	return app.ContainerId()
}
