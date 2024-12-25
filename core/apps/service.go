package apps

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	wstats "github.com/Microsoft/hcsshim/cmd/containerd-shim-runhcs-v1/stats"
	v1 "github.com/containerd/cgroups/v3/cgroup1/stats"
	v2 "github.com/containerd/cgroups/v3/cgroup2/stats"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/errdefs"
	"github.com/containerd/typeurl/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipfs/go-datastore"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/p2p"
)

var log = logrus.New().WithField("service", "apps")

const NAMESPACE = "subnet-apps"
const PROTOCOL_ID = protocol.ID("subnet-apps")
const RESOURCE_USAGE_KEY = "resource-usage"

type Service struct {
	peerId            peer.ID
	IsProvider        bool
	cfg               *config.C
	privKey           *ecdsa.PrivateKey
	ethClient         *ethclient.Client
	containerdClient  *containerd.Client
	P2P               *p2p.P2P
	PeerHost          p2phost.Host `optional:"true"` // the network host (server+client)
	subnetID          big.Int
	stopChan          chan struct{} // Channel to stop background tasks
	subnetAppRegistry *SubnetAppRegistry
	subnetRegistry    *SubnetRegistry
	accountService    *account.AccountService
	Datastore         datastore.Datastore // Datastore for storing resource usage
}

// Initializes the Service with Ethereum and containerd clients.
func New(cfg *config.C, P2P *p2p.P2P, dataStore datastore.Datastore) *Service {
	return &Service{
		P2P:        P2P,
		cfg:        cfg,
		Datastore:  dataStore,
		IsProvider: cfg.GetBool("provider.enable", false),
		stopChan:   make(chan struct{}),
	}
}

func (s *Service) SubnetRegistry() *SubnetRegistry {
	return s.subnetRegistry
}

func (s *Service) Start(ctx context.Context) error {
	// Log that the service is starting
	// log.Info("Starting Subnet Apps Service...")
	var err error

	hexKey := s.cfg.GetString("wallet.private_key", "")
	if hexKey != "" {
		s.privKey, err = PrivateKeyFromHex(hexKey)
		if err != nil {
			return err
		}
	}

	peerId := s.cfg.GetString("identity.peer_id", "")
	if peerId != "" {
		s.peerId = peer.ID(peerId)
	}

	// Register the P2P protocol for signing
	err = s.RegisterSignProtocol()
	if err != nil {
		return fmt.Errorf("failed to register signing protocol: %w", err)
	}

	// Connect to Ethereum RPC
	s.ethClient, err = ethclient.Dial(s.cfg.GetString("apps.rpc", config.DefaultRPC))
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum node: %v", err)
	}

	s.subnetAppRegistry, err = NewSubnetAppRegistry(
		common.HexToAddress(s.cfg.GetString("apps.subnet_app_registry_contract", config.DefaultSubnetAppRegistryContract)),
		s.ethClient,
	)

	if err != nil {
		return err
	}

	s.subnetRegistry, err = NewSubnetRegistry(
		common.HexToAddress(s.cfg.GetString("apps.subnet_registry_contract", config.DefaultSubnetRegistryCOntract)),
		s.ethClient,
	)

	if err != nil {
		return err
	}

	if s.IsProvider {
		subnetID, err := s.GetSubnetIDFromPeerID(ctx)

		if err != nil {
			return fmt.Errorf("error connecting to containerd: %v", err)
		}

		var defaultSubnetID = big.NewInt(0)
		if defaultSubnetID.Cmp(subnetID) == 0 {
			return fmt.Errorf("PeerID is not registered for any SubnetID. Please register first")
		}

		s.subnetID = *subnetID

		// Connect to containerd daemon
		s.containerdClient, err = containerd.New("/run/containerd/containerd.sock")
		if err != nil {
			return fmt.Errorf("error connecting to containerd: %v", err)
		}

		//go s.StartRewardClaimer(ctx)

		// Update latest resource usage into datastore
		s.updateAllRunningContainersUsage(ctx)
		go s.startMonitoringUsage(ctx)
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

func (s *Service) startMonitoringUsage(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			return
		case <-ticker.C: // On every tick, update all latest running containers's resource usages
			s.updateAllRunningContainersUsage(ctx)
		}
	}
}

func (s *Service) GetAppCount() (*big.Int, error) {
	return s.subnetAppRegistry.AppCount(nil)
}

// Retrieves a list of apps from the Ethereum contract and checks their container status.
func (s *Service) GetApps(ctx context.Context, start *big.Int, end *big.Int) ([]*App, error) {
	subnetApps, err := s.subnetAppRegistry.ListApps(nil, start, end)
	if err != nil {
		return nil, err
	}

	apps := make([]*App, len(subnetApps))
	for i, subnetApp := range subnetApps {
		appId := start.Add(start, big.NewInt(int64(i)))
		appStatus, err := s.GetContainerStatus(ctx, appId)
		if err != nil {
			return nil, err
		}

		apps[i] = convertToApp(subnetApp, appId, appStatus)
	}
	return apps, nil
}

func (s *Service) GetApp(ctx context.Context, appId *big.Int) (*App, error) {
	subnetApp, err := s.subnetAppRegistry.Apps(nil, appId)
	if err != nil {
		return nil, err
	}

	appStatus, err := s.GetContainerStatus(ctx, appId)
	if err != nil {
		return nil, err
	}

	return convertToApp(subnetApp, appId, appStatus), nil
}

// Starts a container for the specified app using containerd.
func (s *Service) RunApp(ctx context.Context, appId *big.Int) (*App, error) {
	// Set the namespace for the container
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

	// Retrieve app details from the Ethereum contract
	app, err := s.GetApp(ctx, appId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app details: %w", err)
	}

	if app.Status != NotFound {
		return app, nil
	}

	imageName := app.Metadata.ContainerConfig.Image
	if !strings.Contains(imageName, "/") {
		imageName = "docker.io/library/" + imageName
	}

	// Pull the image for the app
	image, err := s.containerdClient.Pull(ctx, imageName, containerd.WithPullUnpack)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Create a new container for the app
	container, err := s.containerdClient.NewContainer(
		ctx,
		app.ContainerId(),
		containerd.WithNewSnapshot(app.ContainerId()+"-snapshot", image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
		// containerd.WithNewSpec(
		// 	oci.WithMemoryLimit(app.MinMemory.Uint64()), // Minimum RAM
		// 	oci.WithCPUShares(app.MinCpu.Uint64()),      // Minimum CPU
		// ),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start the container
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	_, err = task.Wait(ctx)
	if err != nil {
		return nil, err
	}

	err = task.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start task: %w", err)
	}

	log.Printf("App %s started successfully: Container ID: %s", app.Name, container.ID())

	app.Status, err = s.GetContainerStatus(ctx, appId)
	if err != nil {
		return nil, err
	}

	return app, nil
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
	signHandler := func(data network.Stream) (string, error) {
		// Replace with your actual signing logic
		return "dummy-signature", nil
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

// func (s *Service) RequestSignature(ctx context.Context, peerID peer.ID, protoID protocol.ID, usage *ResourceUsage) (string, error) {
// 	// Open a stream to the remote peer
// 	stream, err := s.PeerHost.NewStream(ctx, peerID, protoID)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to open stream to peer %s: %w", peerID, err)
// 	}
// 	defer stream.Close()

// 	// Send the resource usage data
// 	if err := json.NewEncoder(stream).Encode(usage); err != nil {
// 		return "", fmt.Errorf("failed to send resource usage data: %w", err)
// 	}

// 	// Receive the signature response
// 	var response SignatureResponse
// 	if err := json.NewDecoder(stream).Decode(&response); err != nil {
// 		return "", fmt.Errorf("failed to decode signature response: %w", err)
// 	}

// 	return response.Signature, nil
// }

// func (s *Service) ClaimReward(ctx context.Context, appID big.Int, usage *ResourceUsage, signature string) (common.Hash, error) {
// 	// Pack the claimReward call
// 	data, err := subnetABI.Pack("claimReward",
// 		s.subnetID,
// 		appID,
// 		big.NewInt(int64(usage.UsedCpu)),
// 		big.NewInt(int64(usage.UsedGpu)),
// 		big.NewInt(int64(usage.UsedMemory)),
// 		big.NewInt(int64(usage.UsedStorage)),
// 		big.NewInt(int64(usage.UsedUploadBytes)),
// 		big.NewInt(int64(usage.UsedDownloadBytes)),
// 		big.NewInt(int64(usage.Duration)),
// 		[]byte(signature), // Signature from the app owner's peer
// 	)
// 	if err != nil {
// 		return common.Hash{}, fmt.Errorf("failed to pack transaction data: %w", err)
// 	}

// 	// Get the nonce for the account
// 	fromAddress := crypto.PubkeyToAddress(s.privKey.PublicKey)
// 	nonce, err := s.ethClient.PendingNonceAt(ctx, fromAddress)
// 	if err != nil {
// 		return common.Hash{}, fmt.Errorf("failed to get account nonce: %w", err)
// 	}

// 	// Get the suggested gas price
// 	gasPrice, err := s.ethClient.SuggestGasPrice(ctx)
// 	if err != nil {
// 		return common.Hash{}, fmt.Errorf("failed to get suggested gas price: %w", err)
// 	}

// 	// Estimate the gas limit
// 	msg := ethereum.CallMsg{
// 		From: fromAddress,
// 		To:   &s.subnetAppRegistryContract,
// 		Data: data,
// 	}
// 	gasLimit, err := s.ethClient.EstimateGas(ctx, msg)
// 	if err != nil {
// 		return common.Hash{}, fmt.Errorf("failed to estimate gas: %w", err)
// 	}

// 	toAddress := &s.subnetAppRegistryContract // Get the pointer to the contract address

// 	// Send the transaction
// 	tx := types.NewTx(&types.LegacyTx{
// 		Nonce:    nonce,
// 		GasPrice: gasPrice,
// 		Gas:      gasLimit,
// 		To:       toAddress,
// 		Value:    big.NewInt(0),
// 		Data:     data,
// 	})

// 	// Sign the transaction
// 	chainID, err := s.ethClient.NetworkID(ctx)
// 	if err != nil {
// 		return common.Hash{}, fmt.Errorf("failed to get chain ID: %w", err)
// 	}
// 	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privKey)
// 	if err != nil {
// 		return common.Hash{}, fmt.Errorf("failed to sign transaction: %w", err)
// 	}

// 	err = s.ethClient.SendTransaction(ctx, signedTx)
// 	if err != nil {
// 		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
// 	}

// 	log.Infof("Reward claimed successfully. Transaction hash: %s", signedTx.Hash())
// 	return signedTx.Hash(), nil
// }

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

func (s *Service) GetUsage(ctx context.Context, appId *big.Int) (*ResourceUsage, error) {
	return s.getUsageFromExternal(ctx, appId)
}

func (s *Service) getUsage(ctx context.Context, appId *big.Int) (*ResourceUsage, error) {
	// Load the container
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)
	app := App{ID: appId}
	container, err := s.containerdClient.LoadContainer(ctx, app.ContainerId())
	if err != nil {
		return nil, fmt.Errorf("failed to load container: %w", err)
	}

	// Get the task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get task for container: %w", err)
	}

	// Get metrics
	metric, err := task.Metrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	var data interface{}
	switch {
	case typeurl.Is(metric.Data, (*v1.Metrics)(nil)):
		data = &v1.Metrics{}
	case typeurl.Is(metric.Data, (*v2.Metrics)(nil)):
		data = &v2.Metrics{}
	case typeurl.Is(metric.Data, (*wstats.Statistics)(nil)):
		data = &wstats.Statistics{}
	default:
		return nil, errors.New("cannot convert metric data to cgroups.Metrics or windows.Statistics")
	}
	if err := typeurl.UnmarshalTo(metric.Data, data); err != nil {
		return nil, err
	}

	// Lấy thông tin Network I/O theo PID
	proc, err := process.NewProcess(int32(task.Pid()))
	if err != nil {
		return nil, err
	}

	var totalRxBytes uint64 = 0
	var totalTxBytes uint64 = 0
	netIO, err := proc.NetIOCounters(false)

	if err != nil {
		return nil, err
	}

	for _, stat := range netIO {
		totalRxBytes += stat.BytesRecv
		totalTxBytes += stat.BytesSent
	}

	// Get used storage
	var storageBytes int64 = 0
	info, err := container.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container info: %v", err)
	}
	snapshotKey := info.SnapshotKey
	snapshotService := s.containerdClient.SnapshotService(info.Snapshotter)
	snapshotUsage, err := snapshotService.Usage(ctx, snapshotKey)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, fmt.Errorf("snapshot %s not found: %v", snapshotKey, err)
		}
		return nil, fmt.Errorf("failed to get snapshot usage: %v", err)
	}
	storageBytes = snapshotUsage.Size

	createTime, err := proc.CreateTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get process create time: %w", err)
	}

	now := time.Now().UnixMilli()

	duration := now - createTime
	durationSeconds := duration / 1000

	// Extract resource usage details from the metrics
	usage := &ResourceUsage{
		AppId:             appId,
		UsedUploadBytes:   big.NewInt(int64(totalRxBytes)), // Set if applicable
		UsedDownloadBytes: big.NewInt(int64(totalTxBytes)), // Set if applicable
		Duration:          big.NewInt(durationSeconds),     // Calculate or set duration if available
		UsedStorage:       big.NewInt(storageBytes),
	}
	switch metrics := data.(type) {
	case *v1.Metrics:
		usage.UsedCpu = big.NewInt(int64(metrics.CPU.Usage.Total))
		usage.UsedMemory = big.NewInt(int64(metrics.Memory.Usage.Usage))
	case *v2.Metrics:
		usage.UsedCpu = big.NewInt(int64(metrics.CPU.UsageUsec))
		usage.UsedMemory = big.NewInt(int64(metrics.Memory.Usage))
	case *wstats.Statistics:
		usage.UsedCpu = big.NewInt(int64(metrics.GetWindows().Processor.TotalRuntimeNS))
		usage.UsedMemory = big.NewInt(int64(metrics.GetWindows().Memory.MemoryUsageCommitBytes))
	default:
		return nil, errors.New("unsupported metrics type")
	}

	// Optionally, calculate or set the duration if metrics provide a timestamp
	// For example:
	// usage.Duration = calculateDurationBasedOnMetrics(metrics)

	return usage, nil
}

func (s *Service) getUsageFromExternal(ctx context.Context, appId *big.Int) (*ResourceUsage, error) {
	// Get usage from datastore
	resourceUsageKey := datastore.NewKey(fmt.Sprintf("%s-%s", RESOURCE_USAGE_KEY, appId.String()))
	updatedJsonData, err := s.Datastore.Get(ctx, resourceUsageKey)
	if err == nil { // If the record exists in the datastore
		var updatedData ResourceUsage
		if err := json.Unmarshal(updatedJsonData, &updatedData); err == nil {
			return &updatedData, nil
		}
	}

	// No record found from datastore
	// Get usage from smart contract
	usage, err := s.subnetAppRegistry.GetAppNode(nil, appId, &s.subnetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource usage from datastore & smart contract for appId %s: %v", appId.String(), err)
	}

	return convertToResourceUsage(usage, appId, &s.subnetID), nil
}

func (s *Service) updateUsage(ctx context.Context, appId *big.Int, usage *ResourceUsage) error {
	resourceUsageData, err := json.Marshal(usage)
	if err != nil {
		return fmt.Errorf("failed to marshal resource usage for appId %s: %v", appId.String(), err)
	}
	resourceUsageKey := datastore.NewKey(fmt.Sprintf("%s-%s", RESOURCE_USAGE_KEY, appId.String()))
	if err := s.Datastore.Put(ctx, resourceUsageKey, resourceUsageData); err != nil {
		return fmt.Errorf("failed to update resource usage for appId %s: %v", appId.String(), err)
	}
	return nil
}

func (s *Service) updateAllRunningContainersUsage(ctx context.Context) error {
	// Set the namespace for the containers
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

	// Fetch all running containers
	containers, err := s.containerdClient.Containers(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch running containers: %w", err)
	}

	// Iterate over each container and update its latest resource usage
	for _, container := range containers {
		// Extract appId from container ID
		containerID := container.ID()
		appIDStr := strings.TrimPrefix(containerID, "subnet-app-")
		appID, ok := new(big.Int).SetString(appIDStr, 10)
		if !ok {
			return fmt.Errorf("invalid container ID: %s", containerID)
		}

		usage, err := s.getUsage(ctx, appID)

		if err != nil {
			return fmt.Errorf("failed to get latest resource usage for appId %s: %v", appID.String(), err)
		}

		err = s.updateUsage(ctx, appID, usage)

		if err != nil {
			return fmt.Errorf("failed to update resource usage for container %s: %w", containerID, err)
		}
	}

	return nil
}

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

// func (s *Service) StartRewardClaimer(ctx context.Context) {
// 	ticker := time.NewTicker(1 * time.Hour) // Đặt thời gian định kỳ là 1 giờ
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			log.Infof("Starting reward claim for all running containers...")
// 			s.ClaimRewardsForAllRunningContainers(ctx)
// 		case <-s.stopChan:
// 			log.Infof("Stopping reward claimer for all containers")
// 			return
// 		case <-ctx.Done():
// 			log.Infof("Context canceled, stopping reward claimer")
// 			return
// 		}
// 	}
// }

// func (s *Service) ClaimRewardsForAllRunningContainers(ctx context.Context) {
// 	// Fetch all running containers
// 	containers, err := s.containerdClient.Containers(namespaces.WithNamespace(ctx, NAMESPACE))
// 	if err != nil {
// 		log.Errorf("Failed to fetch running containers: %v", err)
// 		return
// 	}

// 	for _, container := range containers {
// 		// Get container ID (assuming appID is same as container ID)
// 		appID, ok := new(big.Int).SetString(container.ID(), 10)
// 		if !ok {
// 			log.Errorf("Invalid container ID: %s", container.ID())
// 			continue
// 		}

// 		// Fetch resource usage
// 		usage, err := s.GetAppResourceUsage(ctx, *appID)
// 		if err != nil {
// 			log.Errorf("Failed to get resource usage for container %s: %v", container.ID(), err)
// 			continue
// 		}

// 		// Request signature
// 		signature, err := s.RequestSignature(ctx, s.PeerHost.ID(), PROTOCOL_ID, &usage)
// 		if err != nil {
// 			log.Errorf("Failed to get signature for container %s: %v", container.ID(), err)
// 			continue
// 		}

// 		// Claim reward
// 		txHash, err := s.ClaimReward(ctx, *appID, &usage, signature)
// 		if err != nil {
// 			log.Errorf("Failed to claim reward for container %s: %v", container.ID(), err)
// 		} else {
// 			log.Infof("Reward claimed successfully for container %s, transaction hash: %s", container.ID(), txHash.Hex())
// 		}
// 	}
// }

func (s *Service) GetSubnetIDFromPeerID(ctx context.Context) (*big.Int, error) {
	return s.subnetRegistry.PeerToSubnet(nil, string(s.peerId))
}
