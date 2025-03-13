package apps

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/sirupsen/logrus"
	pbstream "github.com/unicornultrafoundation/subnet-node/common/io"
	signer "github.com/unicornultrafoundation/subnet-node/common/signer"
	"github.com/unicornultrafoundation/subnet-node/core/apps/stats"
	"github.com/unicornultrafoundation/subnet-node/core/contracts"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
	pvtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/app/verifier"
)

var log = logrus.WithField("service", "apps")

const ProtocolAppVerifierUsageReport = protocol.ID("/x/app/verifier/usagereport/0.0.1")
const ProtocolAppSignatureRequest = protocol.ID("/x/app/verifier/signreq/0.0.1")
const ProtocollAppSignatureReceive = protocol.ID("/x/app/verifier/signrev/0.0.1")
const ProtocolAppPoWChallenge = protocol.ID("/x/app/verifier/powchallenge/0.0.1")
const ProtocolAppPoWRequest protocol.ID = "/app/pow/request"
const ProtocolAppPoWResponse protocol.ID = "/app/pow/response"

// ProcessStatus returns a human readable status for the Process representing its current status
type ProcessStatus string

const (
	// Running indicates the process is currently executing
	Running ProcessStatus = "running"
	// Created indicates the process has been created within containerd but the
	// user's defined process has not started
	Created ProcessStatus = "created"
	// Stopped indicates that the process has ran and exited
	Stopped ProcessStatus = "stopped"
	// Paused indicates that the process is currently paused
	Paused ProcessStatus = "paused"
	// Pausing indicates that the process is currently switching from a
	// running state into a paused state
	Pausing ProcessStatus = "pausing"
	// Unknown indicates that we could not determine the status from the runtime
	Unknown  ProcessStatus = "unknown"
	NotFound ProcessStatus = "notfound"
)

type App struct {
	ID                   *big.Int
	PeerIds              []string
	Owner                common.Address
	Name                 string
	Description          string
	Logo                 string
	BannerUrls           []string
	DefaultBannerIndex   int64
	Website              string
	Symbol               string
	Budget               *big.Int
	SpentBudget          *big.Int
	MaxNodes             *big.Int
	MinCpu               *big.Int
	MinGpu               *big.Int
	MinMemory            *big.Int
	MinUploadBandwidth   *big.Int
	MinDownloadBandwidth *big.Int
	NodeCount            *big.Int
	PricePerCpu          *big.Int
	PricePerGpu          *big.Int
	PricePerMemoryGB     *big.Int
	PricePerStorageGB    *big.Int
	PricePerBandwidthGB  *big.Int
	Status               ProcessStatus
	Usage                *ResourceUsage
	Metadata             *AppMetadata
	IP                   string
}

func (app *App) ContainerId() string {
	return fmt.Sprintf("subnet-app-%s", app.ID.String())
}

// ResourceUsage represents the resource usage data
type ResourceUsage struct {
	AppId             *big.Int `json:"appId"`
	ProviderId        *big.Int `json:"providerId"`
	PeerId            string   `json:"peerId"`
	UsedCpu           *big.Int `json:"usedCpu"`
	UsedGpu           *big.Int `json:"usedGpu"`
	UsedMemory        *big.Int `json:"usedMemory"`
	UsedStorage       *big.Int `json:"usedStorage"`
	UsedUploadBytes   *big.Int `json:"usedUploadBytes"`
	UsedDownloadBytes *big.Int `json:"usedDownloadBytes"`
	Duration          *big.Int `json:"duration"`
	Timestamp         *big.Int `json:"timestamp"`
}

// SignatureResponse represents the response containing the signature
type SignatureResponse struct {
	Signature string `json:"signature"`
}

// AppMetadata structure
type AppMetadata struct {
	AppInfo         AppInfo         `json:"appInfo"`
	ContainerConfig ContainerConfig `json:"containerConfig"`
	ContactInfo     ContactInfo     `json:"contactInfo"`
}

type ContactInfo struct {
	Email  string `json:"email"`
	Github string `json:"github"`
}

type AppInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Logo        string   `json:"logo"`
	Banner      string   `json:"banner"`
	Website     string   `json:"website"`
	Preview     []string `json:"preview"`
}

type ContainerConfig struct {
	Image     string            `json:"image"`
	Command   []string          `json:"command"`
	Env       map[string]string `json:"env"`
	Resources Resources         `json:"resources"`
	Ports     []Port            `json:"ports"`
	Volumes   []Volume          `json:"volumes"`
}

type Resources struct {
	Requests Requests `json:"requests"`
}

type Requests struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
}

type Port struct {
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

type Volume struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
}

// DeviceCapability represents the hardware capabilities of the device
type DeviceCapability struct {
	AvailableCPU    *big.Int `json:"availableCpu"`    // Available CPU cores
	AvailableMemory *big.Int `json:"availableMemory"` // Available memory in bytes
}

func decodeAndParseMetadata(encodedMetadata string) (*AppMetadata, error) {

	// Decode Base64
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedMetadata)

	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 metadata: %w", err)
	}

	// Parse JSON
	var metadata AppMetadata
	err = json.Unmarshal(decodedBytes, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON metadata: %w", err)
	}

	return &metadata, nil
}

func ConvertToApp(subnetApp contracts.SubnetAppStoreApp, id *big.Int, status ProcessStatus) *App {
	metadata, err := decodeAndParseMetadata(subnetApp.Metadata)
	if err != nil {
		metadata = nil
	}

	return &App{
		ID:                  id,
		PeerIds:             subnetApp.PeerIds,
		Owner:               subnetApp.Owner,
		Name:                subnetApp.Name,
		Symbol:              subnetApp.Symbol,
		Budget:              subnetApp.Budget,
		SpentBudget:         subnetApp.SpentBudget,
		PricePerCpu:         subnetApp.PricePerCpu,
		PricePerGpu:         subnetApp.PricePerGpu,
		PricePerMemoryGB:    subnetApp.PricePerMemoryGB,
		PricePerStorageGB:   subnetApp.PricePerStorageGB,
		PricePerBandwidthGB: subnetApp.PricePerBandwidthGB,
		Status:              status,
		Metadata:            metadata,
	}
}

func FillDefaultResourceUsage(usage *ResourceUsage) *ResourceUsage {
	if usage.AppId == nil {
		usage.AppId = big.NewInt(0)
	}
	if usage.ProviderId == nil {
		usage.ProviderId = big.NewInt(0)
	}
	if usage.UsedCpu == nil {
		usage.UsedCpu = big.NewInt(0)
	}
	if usage.UsedGpu == nil {
		usage.UsedGpu = big.NewInt(0)
	}
	if usage.UsedMemory == nil {
		usage.UsedMemory = big.NewInt(0)
	}
	if usage.UsedStorage == nil {
		usage.UsedStorage = big.NewInt(0)
	}
	if usage.UsedUploadBytes == nil {
		usage.UsedUploadBytes = big.NewInt(0)
	}
	if usage.UsedDownloadBytes == nil {
		usage.UsedDownloadBytes = big.NewInt(0)
	}
	if usage.Duration == nil {
		usage.Duration = big.NewInt(0)
	}

	if usage.Timestamp == nil {
		usage.Timestamp = big.NewInt(0)
	}
	return usage
}

// Helper function to convert []byte to *big.Int
func bytesToBigInt(data []byte) *big.Int {
	if data == nil {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(data)
}

// Helper function to convert *big.Int to []byte
func bigIntToBytes(value *big.Int) []byte {
	if value == nil {
		return []byte{}
	}
	return value.Bytes()
}

func ConvertUsageToProto(usage ResourceUsage) *pbapp.ResourceUsage {
	return &pbapp.ResourceUsage{
		AppId:             bigIntToBytes(usage.AppId),
		ProviderId:        bigIntToBytes(usage.ProviderId),
		PeerId:            usage.PeerId,
		UsedCpu:           bigIntToBytes(usage.UsedCpu),
		UsedGpu:           bigIntToBytes(usage.UsedGpu),
		UsedMemory:        bigIntToBytes(usage.UsedMemory),
		UsedStorage:       bigIntToBytes(usage.UsedStorage),
		UsedUploadBytes:   bigIntToBytes(usage.UsedUploadBytes),
		UsedDownloadBytes: bigIntToBytes(usage.UsedDownloadBytes),
		Duration:          bigIntToBytes(usage.Duration),
		Timestamp:         bigIntToBytes(usage.Timestamp),
	}
}

func ProtoToApp(protoApp *pbapp.App) (*App, error) {
	id, ok := new(big.Int).SetString(protoApp.Id, 10)
	if !ok {
		return nil, fmt.Errorf("invalid id: %s", protoApp.Id)
	}
	budget, ok := new(big.Int).SetString(protoApp.Budget, 10)
	if !ok {
		return nil, fmt.Errorf("invalid budget: %s", protoApp.Budget)
	}
	spentBudget, ok := new(big.Int).SetString(protoApp.SpentBudget, 10)
	if !ok {
		return nil, fmt.Errorf("invalid spent budget: %s", protoApp.SpentBudget)
	}
	maxNodes, ok := new(big.Int).SetString(protoApp.MaxNodes, 10)
	if !ok {
		return nil, fmt.Errorf("invalid max nodes: %s", protoApp.MaxNodes)
	}
	minCpu, ok := new(big.Int).SetString(protoApp.MinCpu, 10)
	if !ok {
		return nil, fmt.Errorf("invalid min cpu: %s", protoApp.MinCpu)
	}
	minGpu, ok := new(big.Int).SetString(protoApp.MinGpu, 10)
	if !ok {
		return nil, fmt.Errorf("invalid min gpu: %s", protoApp.MinGpu)
	}
	minMemory, ok := new(big.Int).SetString(protoApp.MinMemory, 10)
	if !ok {
		return nil, fmt.Errorf("invalid min memory: %s", protoApp.MinMemory)
	}
	minUploadBandwidth, ok := new(big.Int).SetString(protoApp.MinUploadBandwidth, 10)
	if !ok {
		return nil, fmt.Errorf("invalid min upload bandwidth: %s", protoApp.MinUploadBandwidth)
	}
	minDownloadBandwidth, ok := new(big.Int).SetString(protoApp.MinDownloadBandwidth, 10)
	if !ok {
		return nil, fmt.Errorf("invalid min download bandwidth: %s", protoApp.MinDownloadBandwidth)
	}
	nodeCount, ok := new(big.Int).SetString(protoApp.NodeCount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid node count: %s", protoApp.NodeCount)
	}
	pricePerCpu, ok := new(big.Int).SetString(protoApp.PricePerCpu, 10)
	if !ok {
		return nil, fmt.Errorf("invalid price per cpu: %s", protoApp.PricePerCpu)
	}
	pricePerGpu, ok := new(big.Int).SetString(protoApp.PricePerGpu, 10)
	if !ok {
		return nil, fmt.Errorf("invalid price per gpu: %s", protoApp.PricePerGpu)
	}
	pricePerMemoryGB, ok := new(big.Int).SetString(protoApp.PricePerMemoryGb, 10)
	if !ok {
		return nil, fmt.Errorf("invalid price per memory gb: %s", protoApp.PricePerMemoryGb)
	}
	pricePerStorageGB, ok := new(big.Int).SetString(protoApp.PricePerStorageGb, 10)
	if !ok {
		return nil, fmt.Errorf("invalid price per storage gb: %s", protoApp.PricePerStorageGb)
	}
	pricePerBandwidthGB, ok := new(big.Int).SetString(protoApp.PricePerBandwidthGb, 10)
	if !ok {
		return nil, fmt.Errorf("invalid price per bandwidth gb: %s", protoApp.PricePerBandwidthGb)
	}

	return &App{
		ID:                   id,
		PeerIds:              protoApp.PeerIds,
		Owner:                common.HexToAddress(protoApp.Owner),
		Name:                 protoApp.Name,
		Symbol:               protoApp.Symbol,
		Budget:               budget,
		SpentBudget:          spentBudget,
		MaxNodes:             maxNodes,
		MinCpu:               minCpu,
		MinGpu:               minGpu,
		MinMemory:            minMemory,
		MinUploadBandwidth:   minUploadBandwidth,
		MinDownloadBandwidth: minDownloadBandwidth,
		NodeCount:            nodeCount,
		PricePerCpu:          pricePerCpu,
		PricePerGpu:          pricePerGpu,
		PricePerMemoryGB:     pricePerMemoryGB,
		PricePerStorageGB:    pricePerStorageGB,
		PricePerBandwidthGB:  pricePerBandwidthGB,
		Status:               ProcessStatus(protoApp.Status),
		Usage:                ProtoToResourceUsage(protoApp.Usage),
		Metadata:             ProtoToAppMetadata(protoApp.Metadata),
		IP:                   protoApp.Ip,
	}, nil
}

func AppToProto(app *App) *pbapp.App {
	return &pbapp.App{
		Id:                   app.ID.String(),
		PeerIds:              app.PeerIds,
		Owner:                app.Owner.Hex(),
		Name:                 app.Name,
		Symbol:               app.Symbol,
		Budget:               app.Budget.String(),
		SpentBudget:          app.SpentBudget.String(),
		MaxNodes:             app.MaxNodes.String(),
		MinCpu:               app.MinCpu.String(),
		MinGpu:               app.MinGpu.String(),
		MinMemory:            app.MinMemory.String(),
		MinUploadBandwidth:   app.MinUploadBandwidth.String(),
		MinDownloadBandwidth: app.MinDownloadBandwidth.String(),
		NodeCount:            app.NodeCount.String(),
		PricePerCpu:          app.PricePerCpu.String(),
		PricePerGpu:          app.PricePerGpu.String(),
		PricePerMemoryGb:     app.PricePerMemoryGB.String(),
		PricePerStorageGb:    app.PricePerStorageGB.String(),
		PricePerBandwidthGb:  app.PricePerBandwidthGB.String(),
		Status:               string(app.Status),
		Usage:                ResourceUsageToProto(app.Usage),
		Metadata:             AppMetadataToProto(app.Metadata),
		Ip:                   app.IP,
	}
}

func ResourceUsageToProto(usage *ResourceUsage) *pbapp.ResourceUsage {
	return &pbapp.ResourceUsage{
		AppId:             usage.AppId.Bytes(),
		ProviderId:        usage.ProviderId.Bytes(),
		PeerId:            usage.PeerId,
		UsedCpu:           usage.UsedCpu.Bytes(),
		UsedGpu:           usage.UsedGpu.Bytes(),
		UsedMemory:        usage.UsedMemory.Bytes(),
		UsedStorage:       usage.UsedStorage.Bytes(),
		UsedUploadBytes:   usage.UsedUploadBytes.Bytes(),
		UsedDownloadBytes: usage.UsedDownloadBytes.Bytes(),
		Duration:          usage.Duration.Bytes(),
		Timestamp:         usage.Timestamp.Bytes(),
	}
}

func ProtoToResourceUsage(usage *pbapp.ResourceUsage) *ResourceUsage {
	return &ResourceUsage{
		AppId:             bytesToBigInt(usage.AppId),
		ProviderId:        bytesToBigInt(usage.ProviderId),
		PeerId:            usage.PeerId,
		UsedCpu:           bytesToBigInt(usage.UsedCpu),
		UsedGpu:           bytesToBigInt(usage.UsedGpu),
		UsedMemory:        bytesToBigInt(usage.UsedMemory),
		UsedStorage:       bytesToBigInt(usage.UsedStorage),
		UsedUploadBytes:   bytesToBigInt(usage.UsedUploadBytes),
		UsedDownloadBytes: bytesToBigInt(usage.UsedDownloadBytes),
		Duration:          bytesToBigInt(usage.Duration),
		Timestamp:         bytesToBigInt(usage.Timestamp),
	}
}

func ProtoToAppMetadata(protoMetadata *pbapp.AppMetadata) *AppMetadata {
	return &AppMetadata{
		AppInfo: AppInfo{
			Name:        protoMetadata.AppInfo.Name,
			Description: protoMetadata.AppInfo.Description,
			Logo:        protoMetadata.AppInfo.Logo,
			Banner:      protoMetadata.AppInfo.Banner,
			Website:     protoMetadata.AppInfo.Website,
		},
		ContainerConfig: ContainerConfig{
			Image:   protoMetadata.ContainerConfig.Image,
			Command: protoMetadata.ContainerConfig.Command,
			Env:     protoMetadata.ContainerConfig.Env,
			Resources: Resources{
				Requests: Requests{
					CPU:     protoMetadata.ContainerConfig.Resources.Requests.Cpu,
					Memory:  protoMetadata.ContainerConfig.Resources.Requests.Memory,
					Storage: protoMetadata.ContainerConfig.Resources.Requests.Storage,
				},
			},
			Ports:   ProtoToPorts(protoMetadata.ContainerConfig.Ports),
			Volumes: ProtoToVolumes(protoMetadata.ContainerConfig.Volumes),
		},
		ContactInfo: ContactInfo{
			Email:  protoMetadata.ContactInfo.Email,
			Github: protoMetadata.ContactInfo.Github,
		},
	}
}

func AppMetadataToProto(metadata *AppMetadata) *pbapp.AppMetadata {
	return &pbapp.AppMetadata{
		AppInfo: &pbapp.AppInfo{
			Name:        metadata.AppInfo.Name,
			Description: metadata.AppInfo.Description,
			Logo:        metadata.AppInfo.Logo,
			Banner:      metadata.AppInfo.Banner,
			Website:     metadata.AppInfo.Website,
			Preview:     metadata.AppInfo.Preview,
		},
		ContainerConfig: &pbapp.ContainerConfig{
			Image:   metadata.ContainerConfig.Image,
			Command: metadata.ContainerConfig.Command,
			Env:     metadata.ContainerConfig.Env,
			Resources: &pbapp.Resources{
				Requests: &pbapp.Requests{
					Cpu:     metadata.ContainerConfig.Resources.Requests.CPU,
					Memory:  metadata.ContainerConfig.Resources.Requests.Memory,
					Storage: metadata.ContainerConfig.Resources.Requests.Storage,
				},
			},
			Ports:   PortsToProto(metadata.ContainerConfig.Ports),
			Volumes: VolumesToProto(metadata.ContainerConfig.Volumes),
		},
		ContactInfo: &pbapp.ContactInfo{
			Email:  metadata.ContactInfo.Email,
			Github: metadata.ContactInfo.Github,
		},
	}
}

func ProtoToPorts(protoPorts []*pbapp.Port) []Port {
	ports := make([]Port, len(protoPorts))
	for i, protoPort := range protoPorts {
		ports[i] = Port{
			ContainerPort: protoPort.ContainerPort,
			Protocol:      protoPort.Protocol,
		}
	}
	return ports
}

func PortsToProto(ports []Port) []*pbapp.Port {
	protoPorts := make([]*pbapp.Port, len(ports))
	for i, port := range ports {
		protoPorts[i] = &pbapp.Port{
			ContainerPort: port.ContainerPort,
			Protocol:      port.Protocol,
		}
	}
	return protoPorts
}

func ProtoToVolumes(protoVolumes []*pbapp.Volume) []Volume {
	volumes := make([]Volume, len(protoVolumes))
	for i, protoVolume := range protoVolumes {
		volumes[i] = Volume{
			Name:      protoVolume.Name,
			MountPath: protoVolume.MountPath,
		}
	}
	return volumes
}

func VolumesToProto(volumes []Volume) []*pbapp.Volume {
	protoVolumes := make([]*pbapp.Volume, len(volumes))
	for i, volume := range volumes {
		protoVolumes[i] = &pbapp.Volume{
			Name:      volume.Name,
			MountPath: volume.MountPath,
		}
	}
	return protoVolumes
}

func ConvertStatEntryToResourceUsage(entry *stats.StatEntry, appId, providerId *big.Int) *ResourceUsage {
	usage := ResourceUsage{
		AppId:             appId,
		ProviderId:        providerId,
		UsedCpu:           big.NewInt(int64(entry.UsedCpu)),
		UsedGpu:           big.NewInt(int64(entry.UsedGpu)),
		UsedMemory:        big.NewInt(int64(entry.UsedMemory)),
		UsedStorage:       big.NewInt(int64(entry.UsedStorage)),
		UsedUploadBytes:   big.NewInt(int64(entry.UsedDownloadBytes)),
		UsedDownloadBytes: big.NewInt(int64(entry.UsedDownloadBytes)),
		Duration:          big.NewInt(entry.Duration),
	}

	return &usage
}

// Extract appId from container ID
func GetAppIdFromContainerId(containerId string) (*big.Int, error) {
	appIDStr := strings.TrimPrefix(containerId, "subnet-app-")
	appID, ok := new(big.Int).SetString(appIDStr, 10)

	if !ok {
		return nil, fmt.Errorf("invalid container ID: %s", containerId)
	}

	return appID, nil
}

// Extract containerId from appId
func GetContainerIdFromAppId(appId *big.Int) string {
	app := App{ID: appId}

	return app.ContainerId()
}

func ReceiveSignRequest(s network.Stream) (*pbapp.SignatureRequest, error) {
	response := &pbapp.SignatureRequest{}

	err := pbstream.ReadProtoBuffered(s, response)
	if err != nil {
		s.Reset()
		return nil, fmt.Errorf("failed to receive signature request: %v", err)
	}

	return response, nil
}

func ConvertUsageToTypedData(usage *pvtypes.SignedUsage, chainid *big.Int, cAddr string) (*signer.TypedData, error) {
	var domainType = []signer.Type{
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	}

	chainID := math.HexOrDecimal256(*chainid)

	usageTypedData := signer.TypedData{
		Types: signer.Types{
			"EIP712Domain": domainType,
			"Usage": []signer.Type{
				{Name: "appId", Type: "uint256"},
				{Name: "providerId", Type: "uint256"},
				{Name: "peerId", Type: "string"},
				{Name: "usedCpu", Type: "uint256"},
				{Name: "usedGpu", Type: "uint256"},
				{Name: "usedMemory", Type: "uint256"},
				{Name: "usedStorage", Type: "uint256"},
				{Name: "usedUploadBytes", Type: "uint256"},
				{Name: "usedDownloadBytes", Type: "uint256"},
				{Name: "duration", Type: "uint256"},
				{Name: "timestamp", Type: "uint256"},
			},
		},
		Domain: signer.TypedDataDomain{
			Name:              "SubnetAppStore",
			Version:           "1",
			ChainId:           &chainID,
			VerifyingContract: cAddr,
		},
		PrimaryType: "Usage",
		Message: signer.TypedDataMessage{
			"appId":             big.NewInt(usage.AppId),
			"providerId":        big.NewInt(usage.ProviderId),
			"peerId":            usage.PeerId,
			"usedCpu":           big.NewInt(usage.Cpu),
			"usedGpu":           big.NewInt(usage.Gpu),
			"usedMemory":        big.NewInt(usage.Memory),
			"usedStorage":       big.NewInt(usage.Storage),
			"usedUploadBytes":   big.NewInt(usage.UploadBytes),
			"usedDownloadBytes": big.NewInt(usage.DownloadBytes),
			"duration":          big.NewInt(usage.Duration),
			"timestamp":         big.NewInt(usage.Timestamp),
		},
	}

	return &usageTypedData, nil
}
