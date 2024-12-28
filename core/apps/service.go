package apps

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"

	wstats "github.com/Microsoft/hcsshim/cmd/containerd-shim-runhcs-v1/stats"
	v1 "github.com/containerd/cgroups/v3/cgroup1/stats"
	v2 "github.com/containerd/cgroups/v3/cgroup2/stats"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/errdefs"
	"github.com/containerd/typeurl/v2"
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
	"github.com/unicornultrafoundation/subnet-node/core/contracts"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
	pusage "github.com/unicornultrafoundation/subnet-node/proto/subnet/usage"
)

var log = logrus.New().WithField("service", "apps")

const NAMESPACE = "subnet-apps"
const PROTOCOL_ID = protocol.ID("subnet-apps")
const RESOURCE_USAGE_KEY = "resource-usage-v2"

type Service struct {
	peerId            peer.ID
	IsProvider        bool
	cfg               *config.C
	ethClient         *ethclient.Client
	containerdClient  *containerd.Client
	P2P               *p2p.P2P
	PeerHost          p2phost.Host `optional:"true"` // the network host (server+client)
	subnetID          big.Int
	stopChan          chan struct{} // Channel to stop background tasks
	subnetAppRegistry *contracts.SubnetAppRegistry
	subnetRegistry    *contracts.SubnetRegistry
	accountService    *account.AccountService
	Datastore         datastore.Datastore // Datastore for storing resource usage
}

// Initializes the Service with Ethereum and containerd clients.
func New(peerId peer.ID, cfg *config.C, P2P *p2p.P2P, dataStore datastore.Datastore, accountService *account.AccountService) *Service {
	return &Service{
		peerId:            peerId,
		P2P:               P2P,
		cfg:               cfg,
		Datastore:         dataStore,
		IsProvider:        cfg.GetBool("provider.enable", false),
		stopChan:          make(chan struct{}),
		accountService:    accountService,
		ethClient:         accountService.GetClient(),
		subnetAppRegistry: accountService.SubnetAppRegistry(),
		subnetRegistry:    accountService.SubnetRegistry(),
	}
}

func (s *Service) Start(ctx context.Context) error {
	// Register the P2P protocol for signing
	if err := s.RegisterSignProtocol(); err != nil {
		return fmt.Errorf("failed to register signing protocol: %w", err)
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
		if err := s.updateAllRunningContainersUsage(ctx); err != nil {
			return err
		}
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
		appId := big.NewInt(0)
		appId.Add(start, big.NewInt(int64(i)))
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

	app := convertToApp(subnetApp, appId, appStatus)

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
// 	if (err != nil) {
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

func (s *Service) getUsage(ctx context.Context, appId *big.Int) (*ResourceUsage, *uint32, error) {
	// Load the container
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)
	app := App{ID: appId}
	container, err := s.containerdClient.LoadContainer(ctx, app.ContainerId())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load container: %w", err)
	}

	// Get the task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get task for container: %w", err)
	}

	// Get metrics
	metric, err := task.Metrics(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get metrics: %w", err)
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
		return nil, nil, errors.New("cannot convert metric data to cgroups.Metrics or windows.Statistics")
	}
	if err := typeurl.UnmarshalTo(metric.Data, data); err != nil {
		return nil, nil, err
	}

	// Lấy thông tin Network I/O theo PID
	pid := task.Pid()
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, nil, err
	}

	var totalRxBytes uint64 = 0
	var totalTxBytes uint64 = 0
	netIO, err := proc.NetIOCounters(false)

	if err != nil {
		return nil, nil, err
	}

	for _, stat := range netIO {
		totalRxBytes += stat.BytesRecv
		totalTxBytes += stat.BytesSent
	}

	// Get used storage
	var storageBytes int64 = 0
	info, err := container.Info(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get container info: %v", err)
	}
	snapshotKey := info.SnapshotKey
	snapshotService := s.containerdClient.SnapshotService(info.Snapshotter)
	snapshotUsage, err := snapshotService.Usage(ctx, snapshotKey)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, nil, fmt.Errorf("snapshot %s not found: %v", snapshotKey, err)
		}
		return nil, nil, fmt.Errorf("failed to get snapshot usage: %v", err)
	}
	storageBytes = snapshotUsage.Size

	createTime, err := proc.CreateTime()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get process create time: %w", err)
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
		return nil, nil, errors.New("unsupported metrics type")
	}

	// Optionally, calculate or set the duration if metrics provide a timestamp
	// For example:
	// usage.Duration = calculateDurationBasedOnMetrics(metrics)

	return usage, &pid, nil
}

func (s *Service) getUsageFromExternal(ctx context.Context, appId *big.Int) (*ResourceUsage, error) {
	// Get usage from datastore first
	// Get all pids had run the appId from usage metadata
	usageMetadata, err := s.getUsageMetadata(ctx, appId)
	if err == nil {
		// Some records exist in the datastore
		// Get all usages of the appId
		totalUsage := &ResourceUsage{
			UsedCpu:           big.NewInt(0),
			UsedGpu:           big.NewInt(0),
			UsedMemory:        big.NewInt(0),
			UsedStorage:       big.NewInt(0),
			UsedUploadBytes:   big.NewInt(0),
			UsedDownloadBytes: big.NewInt(0),
			Duration:          big.NewInt(0),
		}

		totalStorageRecord := big.NewInt(0)
		for index, pid := range usageMetadata.Pids {
			usage, err := s.getUsageFromStorage(ctx, appId, pid)
			if err != nil {
				return nil, fmt.Errorf("failed to get resource usage from datastore for appId %s - pid %d: %v", appId.String(), pid, err)
			}
			usage = fillDefaultResourceUsage(usage)

			// Aggregate the usage
			totalUsage.UsedCpu.Add(totalUsage.UsedCpu, usage.UsedCpu)
			totalUsage.UsedGpu.Add(totalUsage.UsedGpu, usage.UsedGpu)
			totalUsage.UsedUploadBytes.Add(totalUsage.UsedUploadBytes, usage.UsedUploadBytes)
			totalUsage.UsedDownloadBytes.Add(totalUsage.UsedDownloadBytes, usage.UsedDownloadBytes)
			totalUsage.Duration.Add(totalUsage.Duration, usage.Duration)

			if index < 5 {
				// Only retrieve used memory/storage from <= 5 latest pid
				totalUsage.UsedMemory.Add(totalUsage.UsedMemory, usage.UsedMemory)
				totalUsage.UsedStorage.Add(totalUsage.UsedStorage, usage.UsedStorage)

				totalStorageRecord.Add(totalStorageRecord, big.NewInt(1))
			}
		}

		// Get average used memory/storage from <= 5 latest pid
		totalUsage.UsedMemory.Div(totalUsage.UsedMemory, totalStorageRecord)
		totalUsage.UsedStorage.Div(totalUsage.UsedStorage, totalStorageRecord)

		return totalUsage, nil
	}

	// No record found from datastore
	// Get usage from smart contract
	usage, err := s.subnetAppRegistry.GetAppNode(nil, appId, &s.subnetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource usage from datastore & smart contract for appId %s:%s %v", appId.String(), s.subnetID.String(), err)
	}

	return convertToResourceUsage(usage, appId, &s.subnetID), nil
}

// Get resource usage created by a specific pid for appId from datastore
func (s *Service) getUsageFromStorage(ctx context.Context, appId *big.Int, pid uint32) (*ResourceUsage, error) {
	resourceUsageKey := getUsageKey(appId, pid)
	usageData, err := s.Datastore.Get(ctx, resourceUsageKey)

	if err == nil { // If the record exists in the datastore
		var usage pusage.ResourceUsage
		if err := proto.Unmarshal(usageData, &usage); err == nil {
			return convertUsageFromProto(usage), nil
		}
	}

	return nil, fmt.Errorf("failed to get resource usage from datastore for appId %s: %v", appId.String(), err)
}

// Update resource usage created by a specific pid for appId to datastore
func (s *Service) updateUsage(ctx context.Context, appId *big.Int, pid uint32, usage *ResourceUsage) error {
	resourceUsageData, err := proto.Marshal(convertUsageToProto(*usage))
	if err != nil {
		return fmt.Errorf("failed to marshal resource usage for appId %s: %v", appId.String(), err)
	}

	resourceUsageKey := getUsageKey(appId, pid)
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
		containerId := container.ID()
		appIdStr := strings.TrimPrefix(containerId, "subnet-app-")
		appId, ok := new(big.Int).SetString(appIdStr, 10)
		if !ok {
			return fmt.Errorf("invalid container ID: %s", containerId)
		}

		usage, pid, err := s.getUsage(ctx, appId)
		if err != nil {
			return fmt.Errorf("failed to get latest resource usage for appId %s: %v", appId.String(), err)
		}

		if pid == nil {
			return fmt.Errorf("failed to get pid of appId %s: %v", appId.String(), err)
		}

		// Update usage metadata if the pid is new for appId
		err = s.syncUsageMetadata(ctx, appId, *pid)
		if err != nil {
			return fmt.Errorf("failed to sync pid %d for appId %s: %v", *pid, appId.String(), err)
		}
		err = s.updateUsage(ctx, appId, *pid, usage)

		if err != nil {
			return fmt.Errorf("failed to update resource usage for container %s: %w", containerId, err)
		}
	}

	return nil
}

func (s *Service) getUsageMetadata(ctx context.Context, appId *big.Int) (*pusage.ResourceUsageMetadata, error) {
	usageMetadataKey := getUsageMetadataKey(appId)
	usageMetadataBytes, err := s.Datastore.Get(ctx, usageMetadataKey)
	if err != nil {
		return nil, err
	}

	var usageMetadata pusage.ResourceUsageMetadata
	err = proto.Unmarshal(usageMetadataBytes, &usageMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal usage metadata for appId %s: %w", appId.String(), err)
	}

	return &usageMetadata, nil
}

func (s *Service) updateUsageMetadata(ctx context.Context, appId *big.Int, usageMetadata *pusage.ResourceUsageMetadata) error {
	usageMetadataBytes, err := proto.Marshal(usageMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal usage metadata for appId %s: %v", appId.String(), err)
	}

	usageMetadataKey := getUsageMetadataKey(appId)
	if err := s.Datastore.Put(ctx, usageMetadataKey, usageMetadataBytes); err != nil {
		return fmt.Errorf("failed to update usage metadata for appId %s: %v", appId.String(), err)
	}

	return nil
}

// Check if an `appId` already has an usage metadata or not
func (s *Service) hasUsageMetadata(ctx context.Context, appId *big.Int) (bool, error) {
	usageMetadataKey := getUsageMetadataKey(appId)
	hasUsageMetadata, err := s.Datastore.Has(ctx, usageMetadataKey)

	if err != nil {
		return false, fmt.Errorf("failed to check usage metadata for appId %s: %w", appId.String(), err)
	}

	return hasUsageMetadata, nil
}

// Check if a pid creating resource usage of an `appId` has existed or not
func (s *Service) hasUsageKey(ctx context.Context, appId *big.Int, pid uint32) (bool, error) {
	resourceUsageKey := getUsageKey(appId, pid)
	hasUsage, err := s.Datastore.Has(ctx, resourceUsageKey)

	if err != nil {
		return false, fmt.Errorf("failed to check resource usage for appId %s: %w", appId.String(), err)
	}

	return hasUsage, nil
}

// Add new pid into usage metadata if it hadn't run appId before
func (s *Service) syncUsageMetadata(ctx context.Context, appId *big.Int, pid uint32) error {
	hasPid, err := s.hasUsageKey(ctx, appId, pid)
	if err != nil {
		return fmt.Errorf("failed to check pid for appId %s: %w", appId.String(), err)
	}

	if hasPid {
		return nil
	}

	hasUsageMetadata, err := s.hasUsageMetadata(ctx, appId)
	if err != nil {
		return fmt.Errorf("failed to check usage metadata for appId %s: %w", appId.String(), err)
	}

	usageMetadata := &pusage.ResourceUsageMetadata{
		Pids: *new([]uint32),
	}
	if hasUsageMetadata {
		usageMetadata, err = s.getUsageMetadata(ctx, appId)
		if err != nil {
			return fmt.Errorf("failed to get pidList for appId %s: %w", appId.String(), err)
		}
	}

	// Append new pid on the top of `Pids` list
	usageMetadata.Pids = append([]uint32{pid}, usageMetadata.Pids...)

	return s.updateUsageMetadata(ctx, appId, usageMetadata)
}

// Get key to retrieve resource usage from datastore created by a pid which had run appId
func getUsageKey(appId *big.Int, pid uint32) datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s-%s-%d", RESOURCE_USAGE_KEY, appId.String(), pid))
}

// Get key to retrieve resource usage metadata of an appId
func getUsageMetadataKey(appId *big.Int) datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s-%s", RESOURCE_USAGE_KEY, appId.String()))
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
	return s.subnetRegistry.PeerToSubnet(nil, s.peerId.String())
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
