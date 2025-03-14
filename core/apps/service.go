package apps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gogo/protobuf/proto"
	"github.com/patrickmn/go-cache"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"

	dockerCli "github.com/docker/docker/client"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipfs/go-datastore"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/apps/verifier"

	"github.com/moby/moby/errdefs"
	"github.com/unicornultrafoundation/subnet-node/core/apps/stats"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
	pvtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/app/verifier"
)

var log = logrus.WithField("service", "apps")

const RESOURCE_USAGE_KEY = "resource-usage-v2"

type Service struct {
	peerId                peer.ID
	IsProvider            bool
	IsVerifier            bool
	cfg                   *config.C
	ethClient             *ethclient.Client
	dockerClient          *dockerCli.Client
	P2P                   *p2p.P2P
	PeerHost              p2phost.Host  `optional:"true"` // the network host (server+client)
	stopChan              chan struct{} // Channel to stop background tasks
	accountService        *account.AccountService
	statService           *stats.Stats
	Datastore             datastore.Datastore // Datastore for storing resource usage
	verifier              *verifier.Verifier  // Verifier for resource usage
	pow                   *verifier.Pow       // Proof of Work for resource usage
	signatureResponseChan chan *pvtypes.SignatureResponse

	// Caching fields
	gitHubAppCache  *cache.Cache
	subnetAppCache  *cache.Cache
	gitHubAppsCache *cache.Cache
}

// Initializes the Service with Ethereum and docker clients.
func New(peerHost p2phost.Host, peerId peer.ID, cfg *config.C, P2P *p2p.P2P, ds datastore.Datastore, acc *account.AccountService) *Service {
	return &Service{
		peerId:                peerId,
		PeerHost:              peerHost,
		P2P:                   P2P,
		cfg:                   cfg,
		Datastore:             ds,
		IsProvider:            cfg.GetBool("provider.enable", false),
		IsVerifier:            cfg.GetBool("verifier.enable", false),
		stopChan:              make(chan struct{}),
		accountService:        acc,
		ethClient:             acc.GetClient(),
		signatureResponseChan: make(chan *pvtypes.SignatureResponse, 100),
		gitHubAppCache:        cache.New(1*time.Minute, 2*time.Minute),
		subnetAppCache:        cache.New(1*time.Minute, 2*time.Minute),
		gitHubAppsCache:       cache.New(1*time.Minute, 2*time.Minute),
	}
}

func (s *Service) Start(ctx context.Context) error {
	if s.IsVerifier {
		s.verifier = verifier.NewVerifier(s.Datastore, s.PeerHost, s.P2P, s.accountService)
		// Register the P2P protocol for signing
		if err := s.verifier.Register(); err != nil {
			return fmt.Errorf("failed to register signing protocol: %w", err)
		}
	}

	if s.IsProvider {
		s.pow = verifier.NewPow(verifier.NodeProvider, s.PeerHost, s.P2P)
		s.PeerHost.SetStreamHandler(atypes.ProtocollAppSignatureReceive, s.onSignatureReceive)

		// Connect to docker daemon
		var err error
		s.dockerClient, err = dockerCli.NewClientWithOpts(dockerCli.FromEnv, dockerCli.WithAPIVersionNegotiation())
		if err != nil {
			return fmt.Errorf("error connecting to docker: %v", err)
		}
		s.statService = stats.NewStats(s.dockerClient)

		s.RestartStoppedContainers(ctx)
		s.upgradeAppVersion(ctx)

		// Start app sub-services
		go s.startUpgradeAppVersion(ctx)
		go s.statService.Start()
		go s.startReportLoop(ctx)
		go s.handleSignatureResponses(ctx)
	}

	log.Info("Subnet Apps Service started successfully.")
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	log.Info("Stopping Subnet Apps Service...")

	// Close stopChan to stop all background tasks
	close(s.stopChan)

	// Close the docker client
	if s.dockerClient != nil {
		err := s.dockerClient.Close()
		if err != nil {
			return fmt.Errorf("failed to close docker client: %w", err)
		}
	}

	if s.statService != nil {
		// Close sub-services
		s.statService.Stop()
	}

	close(s.signatureResponseChan)

	log.Info("Subnet Apps Service stopped successfully.")
	return nil
}

func (s *Service) GetAppCount() (*big.Int, error) {
	return s.accountService.AppStore().AppCount(nil)
}

func (s *Service) GetContainerConfigProto(ctx context.Context, appId *big.Int) (*atypes.AppMetadata, error) {
	configKey := datastore.NewKey(fmt.Sprintf("container-config-%s", appId.String()))
	configData, err := s.Datastore.Get(ctx, configKey)
	if err != nil {
		return nil, err
	}

	var protoConfig pbapp.AppMetadata
	if err := proto.Unmarshal(configData, &protoConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal container config: %w", err)
	}

	return atypes.ProtoToAppMetadata(&protoConfig), nil
}

// Retrieves the status of a container associated with a specific app.
func (s *Service) GetContainerStatus(ctx context.Context, appId *big.Int) (atypes.ProcessStatus, error) {
	containerId := atypes.GetContainerIdFromAppId(appId)

	// Load the container for the app
	container, err := s.dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return atypes.NotFound, nil
		}
		return atypes.Unknown, err
	}

	// Retrieve the status of the container
	status := container.State.Status

	// TODO: Add other Docker statuses
	switch status {
	case atypes.DockerCreated:
		return atypes.Created, nil
	case atypes.DockerPaused:
		return atypes.Paused, nil
	case atypes.DockerRunning:
		return atypes.Running, nil
	case atypes.DockerExited:
		return atypes.Stopped, nil
	default:
		return atypes.Unknown, nil
	}
}

// Retrieves the IP address of a running container.
func (s *Service) GetContainerIP(ctx context.Context, appId *big.Int) (string, error) {
	container, err := s.ContainerInspect(ctx, appId)
	if err != nil {
		return "", err
	}
	return container.NetworkSettings.IPAddress, nil
}

func (s *Service) SaveContainerConfigProto(ctx context.Context, appId *big.Int, config atypes.ContainerConfig) error {
	protoConfig := atypes.AppMetadataToProto(&atypes.AppMetadata{ContainerConfig: config})
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

// Fetches the list of apps from GitHub with caching.
func (s *Service) getGitHubApps(url string) ([]GitHubApp, error) {
	if cachedApps, found := s.gitHubAppsCache.Get(url); found {
		return cachedApps.([]GitHubApp), nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch GitHub apps: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gitHubApps []GitHubApp
	if err := json.Unmarshal(body, &gitHubApps); err != nil {
		return nil, err
	}

	s.gitHubAppsCache.Set(url, gitHubApps, cache.DefaultExpiration)
	return gitHubApps, nil
}

// Fetches a GitHub app by its ID with caching.
func (s *Service) getGitHubAppByID(id int64) (*GitHubApp, error) {
	if cachedApp, found := s.gitHubAppCache.Get(fmt.Sprintf("%d", id)); found {
		return cachedApp.(*GitHubApp), nil
	}

	gitHubApps, err := s.getGitHubApps(gitHubAppsURL)
	if err != nil {
		return nil, err
	}

	for _, app := range gitHubApps {
		if app.ID == id {
			s.gitHubAppCache.Set(fmt.Sprintf("%d", id), &app, cache.DefaultExpiration)
			return &app, nil
		}
	}

	return nil, fmt.Errorf("GitHub app with ID %d not found", id)
}

// Fetches a subnet app by its ID with caching.
func (s *Service) getSubnetAppByID(_ context.Context, appId *big.Int) (*atypes.App, error) {
	if cachedApp, found := s.subnetAppCache.Get(appId.String()); found {
		return cachedApp.(*atypes.App), nil
	}

	subnetApp, err := s.accountService.AppStore().GetApp(nil, appId)
	if err != nil {
		return nil, err
	}

	app := atypes.ConvertToApp(subnetApp, appId, atypes.Unknown)
	s.subnetAppCache.Set(appId.String(), app, cache.DefaultExpiration)
	return app, nil
}

// getDeviceCapability retrieves the CPU, memory, and storage capabilities of the machine.
// It uses the gopsutil library to fetch detailed system information.
// If gopsutil fails, it falls back to using the runtime package for CPU count.

func (s *Service) getDeviceCapability() (atypes.DeviceCapability, error) {
	// Get CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		// Fallback to runtime.NumCPU() if gopsutil fails
		return atypes.DeviceCapability{
			AvailableCPU:     big.NewInt(int64(runtime.NumCPU())),
			AvailableMemory:  big.NewInt(0),
			AvailableStorage: big.NewInt(0),
		}, fmt.Errorf("failed to get CPU info: %w", err)
	}

	var totalCores int64
	for _, cpu := range cpuInfo {
		totalCores += int64(cpu.Cores)
	}
	if totalCores == 0 {
		totalCores = int64(runtime.NumCPU())
	}

	// Get memory information
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return atypes.DeviceCapability{
			AvailableCPU:     big.NewInt(totalCores),
			AvailableMemory:  big.NewInt(0),
			AvailableStorage: big.NewInt(0),
		}, fmt.Errorf("failed to get memory info: %w", err)
	}

	// Get storage information
	var availableStorage uint64
	partitions, err := disk.Partitions(false) // false means only physical partitions
	if err == nil {
		for _, partition := range partitions {
			usage, err := disk.Usage(partition.Mountpoint)
			if err == nil {
				availableStorage += usage.Free
			}
		}
	}

	return atypes.DeviceCapability{
		AvailableCPU:     big.NewInt(totalCores),
		AvailableMemory:  big.NewInt(int64(vmStat.Available)),
		AvailableStorage: big.NewInt(int64(availableStorage)),
	}, nil
}

// Parsing resource request from human reading to big.Int
func (s *Service) parseResourceRequest(appResourceRequest atypes.Requests) (*big.Int, *big.Int, *big.Int, error) {
	// request CPU
	cpuInt, err := strconv.ParseInt(appResourceRequest.CPU, 10, 64)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse CPU request: %w", err)
	}
	requestCPU := new(big.Int).SetInt64(cpuInt)

	// request memory
	memoryStr := appResourceRequest.Memory
	memoryValue, err := humanize.ParseBytes(memoryStr)

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse memory request: %w", err)
	}
	requestMemory := new(big.Int).SetInt64(int64(memoryValue))

	// request storage
	var storageValue uint64
	storageStr := appResourceRequest.Storage
	if storageStr != "" {
		storageValue, err = humanize.ParseBytes(storageStr)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse storage request: %w", err)
		}
	}
	requestStorage := new(big.Int).SetInt64(int64(storageValue))
	return requestCPU, requestMemory, requestStorage, nil
}

func (s *Service) validateNodeCompatibility(resourceUsage atypes.ResourceUsage, appResourceRequest atypes.Requests, deviceCapability atypes.DeviceCapability) (*big.Int, *big.Int, *big.Int, error) { // requestCPU, requestMemory, requestStorage
	// calculate remain CPU and memory
	remainCpu := new(big.Int).Sub(deviceCapability.AvailableCPU, resourceUsage.UsedCpu)
	remainMemory := new(big.Int).Sub(deviceCapability.AvailableMemory, resourceUsage.UsedMemory)
	remainStorage := new(big.Int).Sub(deviceCapability.AvailableStorage, resourceUsage.UsedStorage)

	requestCPU, requestMemory, requestStorage, err := s.parseResourceRequest(appResourceRequest)
	if err != nil {
		return nil, nil, nil, err
	}

	// compare remain CPU, memory, storage with request
	if remainCpu.Cmp(requestCPU) < 0 {
		return nil, nil, nil, errors.New("insufficient CPU")
	}

	if remainMemory.Cmp(requestMemory) < 0 {
		return nil, nil, nil, errors.New("insufficient memory")
	}

	if remainStorage.Cmp(requestStorage) < 0 {
		return nil, nil, nil, errors.New("insufficient storage")
	}

	return requestCPU, requestMemory, requestStorage, nil
}

func (s *Service) setNodeResourceUsage(resourceUsage atypes.ResourceUsage) {
	s.subnetAppCache.Set(RESOURCE_USAGE_KEY, resourceUsage, cache.DefaultExpiration)
}

func (s *Service) getNodeResourceUsage() (atypes.ResourceUsage, error) {
	if cachedUsage, found := s.subnetAppCache.Get(RESOURCE_USAGE_KEY); found {
		return cachedUsage.(atypes.ResourceUsage), nil
	}

	return atypes.ResourceUsage{
		UsedCpu:     big.NewInt(0),
		UsedMemory:  big.NewInt(0),
		UsedStorage: big.NewInt(0),
	}, nil
}
