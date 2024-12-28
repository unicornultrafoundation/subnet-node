package apps

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/core/contracts"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
	pusage "github.com/unicornultrafoundation/subnet-node/proto/subnet/usage"
)

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
	PeerId               string
	Owner                common.Address
	Name                 string
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
	SubnetId          *big.Int `json:"subnetId"`
	UsedCpu           *big.Int `json:"usedCpu"`
	UsedGpu           *big.Int `json:"usedGpu"`
	UsedMemory        *big.Int `json:"usedMemory"`
	UsedStorage       *big.Int `json:"usedStorage"`
	UsedUploadBytes   *big.Int `json:"usedUploadBytes"`
	UsedDownloadBytes *big.Int `json:"usedDownloadBytes"`
	Duration          *big.Int `json:"duration"`
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
	Name        string `json:"name"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Banner      string `json:"banner"`
	Website     string `json:"website"`
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

func decodeAndParseMetadata(encodedMetadata string) (*AppMetadata, error) {
	// Giải mã Base64
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

func convertToApp(subnetApp contracts.SubnetAppRegistryApp, id *big.Int, status ProcessStatus) *App {
	metadata, err := decodeAndParseMetadata(subnetApp.Metadata)
	if err != nil {
		fmt.Printf("Warning: Failed to parse metadata for app %s: %v\n", subnetApp.Name, err)
		metadata = nil
	}

	return &App{
		ID:                   id,
		PeerId:               subnetApp.PeerId,
		Owner:                subnetApp.Owner,
		Name:                 subnetApp.Name,
		Symbol:               subnetApp.Symbol,
		Budget:               subnetApp.Budget,
		SpentBudget:          subnetApp.SpentBudget,
		MaxNodes:             subnetApp.MaxNodes,
		MinCpu:               subnetApp.MinCpu,
		MinGpu:               subnetApp.MinGpu,
		MinMemory:            subnetApp.MinMemory,
		MinUploadBandwidth:   subnetApp.MinUploadBandwidth,
		MinDownloadBandwidth: subnetApp.MinDownloadBandwidth,
		NodeCount:            subnetApp.NodeCount,
		PricePerCpu:          subnetApp.PricePerCpu,
		PricePerGpu:          subnetApp.PricePerGpu,
		PricePerMemoryGB:     subnetApp.PricePerMemoryGB,
		PricePerStorageGB:    subnetApp.PricePerStorageGB,
		PricePerBandwidthGB:  subnetApp.PricePerBandwidthGB,
		Status:               status,
		Metadata:             metadata,
	}
}

func convertToResourceUsage(usage contracts.SubnetAppRegistryAppNode, appId *big.Int, subnetId *big.Int) *ResourceUsage {
	return &ResourceUsage{
		AppId:             appId,
		SubnetId:          subnetId,
		UsedCpu:           usage.UsedCpu,
		UsedGpu:           usage.UsedGpu,
		UsedMemory:        usage.UsedMemory,
		UsedStorage:       usage.UsedStorage,
		UsedUploadBytes:   usage.UsedUploadBytes,
		UsedDownloadBytes: usage.UsedDownloadBytes,
		Duration:          usage.Duration,
	}
}

func fillDefaultResourceUsage(usage *ResourceUsage) *ResourceUsage {
	if usage.AppId == nil {
		usage.AppId = big.NewInt(0)
	}
	if usage.SubnetId == nil {
		usage.SubnetId = big.NewInt(0)
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
	return usage
}

// Helper function to convert []byte to *big.Int
func bytesToBigInt(data []byte) *big.Int {
	if data == nil {
		return nil // Handle nil bytes
	}
	return new(big.Int).SetBytes(data)
}

func convertUsageFromProto(usage pusage.ResourceUsage) *ResourceUsage {
	return &ResourceUsage{
		AppId:             bytesToBigInt(usage.AppId),
		SubnetId:          bytesToBigInt(usage.SubnetId),
		UsedCpu:           bytesToBigInt(usage.UsedCpu),
		UsedGpu:           bytesToBigInt(usage.UsedGpu),
		UsedMemory:        bytesToBigInt(usage.UsedMemory),
		UsedStorage:       bytesToBigInt(usage.UsedStorage),
		UsedUploadBytes:   bytesToBigInt(usage.UsedUploadBytes),
		UsedDownloadBytes: bytesToBigInt(usage.UsedDownloadBytes),
		Duration:          bytesToBigInt(usage.Duration),
	}
}

// Helper function to convert *big.Int to []byte
func bigIntToBytes(value *big.Int) []byte {
	if value == nil {
		return nil // Handle nil big.Int
	}
	return value.Bytes()
}

func convertUsageToProto(usage ResourceUsage) *pusage.ResourceUsage {
	return &pusage.ResourceUsage{
		AppId:             bigIntToBytes(usage.AppId),
		SubnetId:          bigIntToBytes(usage.SubnetId),
		UsedCpu:           bigIntToBytes(usage.UsedCpu),
		UsedGpu:           bigIntToBytes(usage.UsedGpu),
		UsedMemory:        bigIntToBytes(usage.UsedMemory),
		UsedStorage:       bigIntToBytes(usage.UsedStorage),
		UsedUploadBytes:   bigIntToBytes(usage.UsedUploadBytes),
		UsedDownloadBytes: bigIntToBytes(usage.UsedDownloadBytes),
		Duration:          bigIntToBytes(usage.Duration),
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

	usage, err := ProtoToResourceUsage(protoApp.Usage)
	if err != nil {
		return nil, err
	}

	return &App{
		ID:                   id,
		PeerId:               protoApp.PeerId,
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
		Usage:                usage,
		Metadata:             ProtoToAppMetadata(protoApp.Metadata),
		IP:                   protoApp.Ip,
	}, nil
}

func AppToProto(app *App) *pbapp.App {
	return &pbapp.App{
		Id:                   app.ID.String(),
		PeerId:               app.PeerId,
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

func ProtoToResourceUsage(protoUsage *pbapp.ResourceUsage) (*ResourceUsage, error) {
	appId, ok := new(big.Int).SetString(protoUsage.AppId, 10)
	if !ok {
		return nil, fmt.Errorf("invalid app id: %s", protoUsage.AppId)
	}
	subnetId, ok := new(big.Int).SetString(protoUsage.SubnetId, 10)
	if !ok {
		return nil, fmt.Errorf("invalid subnet id: %s", protoUsage.SubnetId)
	}
	usedCpu, ok := new(big.Int).SetString(protoUsage.UsedCpu, 10)
	if !ok {
		return nil, fmt.Errorf("invalid used cpu: %s", protoUsage.UsedCpu)
	}
	usedGpu, ok := new(big.Int).SetString(protoUsage.UsedGpu, 10)
	if !ok {
		return nil, fmt.Errorf("invalid used gpu: %s", protoUsage.UsedGpu)
	}
	usedMemory, ok := new(big.Int).SetString(protoUsage.UsedMemory, 10)
	if !ok {
		return nil, fmt.Errorf("invalid used memory: %s", protoUsage.UsedMemory)
	}
	usedStorage, ok := new(big.Int).SetString(protoUsage.UsedStorage, 10)
	if !ok {
		return nil, fmt.Errorf("invalid used storage: %s", protoUsage.UsedStorage)
	}
	usedUploadBytes, ok := new(big.Int).SetString(protoUsage.UsedUploadBytes, 10)
	if !ok {
		return nil, fmt.Errorf("invalid used upload bytes: %s", protoUsage.UsedUploadBytes)
	}
	usedDownloadBytes, ok := new(big.Int).SetString(protoUsage.UsedDownloadBytes, 10)
	if !ok {
		return nil, fmt.Errorf("invalid used download bytes: %s", protoUsage.UsedDownloadBytes)
	}
	duration, ok := new(big.Int).SetString(protoUsage.Duration, 10)
	if !ok {
		return nil, fmt.Errorf("invalid duration: %s", protoUsage.Duration)
	}

	return &ResourceUsage{
		AppId:             appId,
		SubnetId:          subnetId,
		UsedCpu:           usedCpu,
		UsedGpu:           usedGpu,
		UsedMemory:        usedMemory,
		UsedStorage:       usedStorage,
		UsedUploadBytes:   usedUploadBytes,
		UsedDownloadBytes: usedDownloadBytes,
		Duration:          duration,
	}, nil
}

func ResourceUsageToProto(usage *ResourceUsage) *pbapp.ResourceUsage {
	return &pbapp.ResourceUsage{
		AppId:             usage.AppId.String(),
		SubnetId:          usage.SubnetId.String(),
		UsedCpu:           usage.UsedCpu.String(),
		UsedGpu:           usage.UsedGpu.String(),
		UsedMemory:        usage.UsedMemory.String(),
		UsedStorage:       usage.UsedStorage.String(),
		UsedUploadBytes:   usage.UsedUploadBytes.String(),
		UsedDownloadBytes: usage.UsedDownloadBytes.String(),
		Duration:          usage.Duration.String(),
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
				CPU:     protoMetadata.ContainerConfig.Resources.Cpu,
				Memory:  protoMetadata.ContainerConfig.Resources.Memory,
				Storage: protoMetadata.ContainerConfig.Resources.Storage,
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
		},
		ContainerConfig: &pbapp.ContainerConfig{
			Image:   metadata.ContainerConfig.Image,
			Command: metadata.ContainerConfig.Command,
			Env:     metadata.ContainerConfig.Env,
			Resources: &pbapp.Resources{
				Cpu:     metadata.ContainerConfig.Resources.CPU,
				Memory:  metadata.ContainerConfig.Resources.Memory,
				Storage: metadata.ContainerConfig.Resources.Storage,
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
