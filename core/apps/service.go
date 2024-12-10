package apps

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	wstats "github.com/Microsoft/hcsshim/cmd/containerd-shim-runhcs-v1/stats"
	v1 "github.com/containerd/cgroups/v3/cgroup1/stats"
	v2 "github.com/containerd/cgroups/v3/cgroup2/stats"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/typeurl/v2"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/p2p"
)

var log = logrus.New().WithField("service", "core")

const NAMESPACE = "subnet-apps"
const PROTOCOL_ID = protocol.ID("subnet-apps")

type Service struct {
	IsProvider                bool
	cfg                       *config.C
	privKey                   *ecdsa.PrivateKey
	ethClient                 *ethclient.Client
	subnetAppRegistryContract common.Address
	containerdClient          *containerd.Client
	P2P                       *p2p.P2P
	PeerHost                  p2phost.Host `optional:"true"` // the network host (server+client)
}

// Initializes the Service with Ethereum and containerd clients.
func New(cfg *config.C, P2P *p2p.P2P) *Service {
	return &Service{
		P2P:                       P2P,
		cfg:                       cfg,
		IsProvider:                cfg.GetBool("provider.enable", false),
		subnetAppRegistryContract: common.HexToAddress(cfg.GetString("apps.subnet_app_registry_contract", "")),
	}
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

	if s.IsProvider {
		// Connect to containerd daemon
		s.containerdClient, err = containerd.New("/run/containerd/containerd.sock")
		if err != nil {
			return fmt.Errorf("error connecting to containerd: %v", err)
		}
	}

	log.Info("Subnet Apps Service started successfully.")
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	log.Info("Stopping Subnet Apps Service...")

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
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

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
		containerd.WithNewSnapshot(app.ID.String()+"-snapshot", image),
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
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

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
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)

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

func (s *Service) RequestSignature(ctx context.Context, peerID peer.ID, protoID protocol.ID, usage *ResourceUsage) (string, error) {
	// Open a stream to the remote peer
	stream, err := s.PeerHost.NewStream(ctx, peerID, protoID)
	if err != nil {
		return "", fmt.Errorf("failed to open stream to peer %s: %w", peerID, err)
	}
	defer stream.Close()

	// Send the resource usage data
	if err := json.NewEncoder(stream).Encode(usage); err != nil {
		return "", fmt.Errorf("failed to send resource usage data: %w", err)
	}

	// Receive the signature response
	var response SignatureResponse
	if err := json.NewDecoder(stream).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode signature response: %w", err)
	}

	return response.Signature, nil
}

func (s *Service) ClaimReward(ctx context.Context, subnetID, appID uint64, usage *ResourceUsage, signature string) (common.Hash, error) {
	// Pack the claimReward call
	data, err := subnetABI.Pack("claimReward",
		big.NewInt(int64(subnetID)),
		big.NewInt(int64(appID)),
		big.NewInt(int64(usage.UsedCpu)),
		big.NewInt(int64(usage.UsedGpu)),
		big.NewInt(int64(usage.UsedMemory)),
		big.NewInt(int64(usage.UsedStorage)),
		big.NewInt(int64(usage.UsedUploadBytes)),
		big.NewInt(int64(usage.UsedDownloadBytes)),
		big.NewInt(int64(usage.Duration)),
		[]byte(signature), // Signature from the app owner's peer
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack transaction data: %w", err)
	}

	// Get the nonce for the account
	fromAddress := crypto.PubkeyToAddress(s.privKey.PublicKey)
	nonce, err := s.ethClient.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get account nonce: %w", err)
	}

	// Get the suggested gas price
	gasPrice, err := s.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get suggested gas price: %w", err)
	}

	// Estimate the gas limit
	msg := ethereum.CallMsg{
		From: fromAddress,
		To:   &s.subnetAppRegistryContract,
		Data: data,
	}
	gasLimit, err := s.ethClient.EstimateGas(ctx, msg)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to estimate gas: %w", err)
	}

	toAddress := &s.subnetAppRegistryContract // Get the pointer to the contract address

	// Send the transaction
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       toAddress,
		Value:    big.NewInt(0),
		Data:     data,
	})

	// Sign the transaction
	chainID, err := s.ethClient.NetworkID(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get chain ID: %w", err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = s.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Infof("Reward claimed successfully. Transaction hash: %s", signedTx.Hash())
	return signedTx.Hash(), nil
}

// Calculate container duration
func (s *Service) CalculateContainerDuration(ctx context.Context, appId big.Int) (int64, error) {
	// Load the container
	container, err := s.containerdClient.LoadContainer(ctx, appId.String())
	if err != nil {
		return 0, err // Return 0 if container not found
	}

	// Get the task associated with the container
	info, err := container.Info(ctx, nil)
	if err != nil {
		log.Errorf("Failed to get container task: %v", err)
		return 0, err // Return 0 if task not found
	}

	// Calculate duration (current time - start time)
	startTime := info.CreatedAt
	if startTime.IsZero() {
		return 0, nil // If no start time, duration is 0
	}
	duration := time.Since(startTime).Seconds()
	return int64(duration), nil // Return duration in seconds
}

func (s *Service) GetAppResourceUsage(ctx context.Context, appId big.Int) (ResourceUsage, error) {
	// Load the container
	ctx = namespaces.WithNamespace(ctx, NAMESPACE)
	container, err := s.containerdClient.LoadContainer(ctx, appId.String())
	if err != nil {
		return ResourceUsage{}, fmt.Errorf("failed to load container: %w", err)
	}

	// Get the task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return ResourceUsage{}, fmt.Errorf("failed to get task for container: %w", err)
	}

	// Get metrics
	metric, err := task.Metrics(ctx)
	if err != nil {
		return ResourceUsage{}, fmt.Errorf("failed to get metrics: %w", err)
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
		return ResourceUsage{}, errors.New("cannot convert metric data to cgroups.Metrics or windows.Statistics")
	}
	if err := typeurl.UnmarshalTo(metric.Data, data); err != nil {
		return ResourceUsage{}, err
	}

	// Lấy thông tin Network I/O theo PID
	proc, err := process.NewProcess(int32(task.Pid()))
	if err != nil {
		return ResourceUsage{}, err
	}

	var totalRxBytes uint64 = 0
	var totalTxBytes uint64 = 0
	netIO, err := proc.NetIOCounters(false)

	if err != nil {
		return ResourceUsage{}, err
	}

	for _, stat := range netIO {
		totalRxBytes += stat.BytesRecv
		totalTxBytes += stat.BytesSent
	}

	createTime, err := proc.CreateTime()
	if err != nil {
		return ResourceUsage{}, fmt.Errorf("failed to get process create time: %w", err)
	}

	now := time.Now().UnixMilli()

	duration := now - createTime
	durationSeconds := duration / 1000

	// Extract resource usage details from the metrics
	var usage ResourceUsage
	switch metrics := data.(type) {
	case *v1.Metrics:
		// Parse cgroup v1 metrics
		usage = ResourceUsage{
			AppId:             appId,
			UsedCpu:           metrics.CPU.Usage.Total,
			UsedMemory:        metrics.Memory.Usage.Usage,
			UsedStorage:       0,                       // Storage usage might require custom logic
			UsedUploadBytes:   totalRxBytes,            // Set if applicable
			UsedDownloadBytes: totalTxBytes,            // Set if applicable
			Duration:          uint64(durationSeconds), // Calculate or set duration if available
		}
	case *v2.Metrics:
		// Parse cgroup v2 metrics
		usage = ResourceUsage{
			AppId:             appId,
			UsedCpu:           metrics.CPU.UsageUsec, // Convert microseconds to nanoseconds
			UsedMemory:        metrics.Memory.Usage,
			UsedStorage:       0,                       // Storage usage might require custom logic
			UsedUploadBytes:   totalRxBytes,            // Set if applicable
			UsedDownloadBytes: totalTxBytes,            // Set if applicable
			Duration:          uint64(durationSeconds), // Calculate or set duration if available
		}
	case *wstats.Statistics:
		// Parse Windows statistics
		usage = ResourceUsage{
			AppId:             appId,
			UsedCpu:           metrics.GetWindows().Processor.TotalRuntimeNS, // Convert 100-nanosecond intervals to nanoseconds
			UsedMemory:        metrics.GetWindows().Memory.MemoryUsageCommitBytes,
			UsedStorage:       0,                       // Storage usage might require custom logic
			UsedUploadBytes:   totalRxBytes,            // Set if applicable
			UsedDownloadBytes: totalTxBytes,            // Set if applicable
			Duration:          uint64(durationSeconds), // Calculate or set duration if available
		}
	default:
		return ResourceUsage{}, errors.New("unsupported metrics type")
	}

	// Optionally, calculate or set the duration if metrics provide a timestamp
	// For example:
	// usage.Duration = calculateDurationBasedOnMetrics(metrics)

	return usage, nil

}

// GetAppNode retrieves the details of an AppNode by its ID or similar identifier.
func (s *Service) GetAppNode(ctx context.Context, appId big.Int, subnetId big.Int) (ResourceUsage, error) {
	// Prepare input for the contract call to retrieve AppNode details
	input, err := subnetABI.Pack("getAppNode", appId, subnetId)
	if err != nil {
		log.Errorf("Failed to pack input for getAppNode: %v", err)
		return ResourceUsage{}, err
	}

	// Create the contract call message
	msg := ethereum.CallMsg{
		To:   &s.subnetAppRegistryContract,
		Data: input,
	}

	// Execute the contract call
	result, err := s.ethClient.CallContract(ctx, msg, nil)
	if err != nil {
		log.Errorf("Failed to call contract for AppNode details: %v", err)
		return ResourceUsage{}, err
	}

	// Unpack the result into the AppNode structure
	var usage ResourceUsage
	err = subnetABI.UnpackIntoInterface(&usage, "getAppNode", result)
	if err != nil {
		log.Errorf("Failed to unpack AppNode details: %v", err)
		return ResourceUsage{}, err
	}

	// Return the retrieved AppNode details
	return usage, nil
}
