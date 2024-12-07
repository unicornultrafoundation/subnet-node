package apps

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/p2p"
)

var log = logrus.New().WithField("service", "core")

type Service struct {
	privKey                   *ecdsa.PrivateKey
	ethClient                 *ethclient.Client
	subnetAppRegistryContract common.Address
	containerdClient          *containerd.Client
	P2P                       *p2p.P2P
	PeerHost                  p2phost.Host `optional:"true"` // the network host (server+client)
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

func (s *Service) RegisterSignProtocol(protoID protocol.ID) error {
	// Define the signing logic
	signHandler := func(data network.Stream) (string, error) {
		// Replace with your actual signing logic
		return "dummy-signature", nil
	}

	// Create the listener for the signing protocol
	listener := p2p.NewSignProtocolListener(protoID, signHandler)

	// Register the listener in the ListenersP2P
	err := s.P2P.ListenersP2P.Register(listener)
	if err != nil {
		return err
	}

	log.Printf("Registered signing protocol: %s", protoID)
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
