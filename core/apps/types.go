package apps

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/unicornultrafoundation/subnet-node/core/apps/stats"
	"github.com/unicornultrafoundation/subnet-node/core/contracts"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
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
	ProviderId        *big.Int `json:"providerId"`
	PeerId            string   `json:"peerId"`
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

func convertToApp(subnetApp contracts.SubnetAppStoreApp, id *big.Int, status ProcessStatus) *App {
	metadata, err := decodeAndParseMetadata(subnetApp.Metadata)
	if err != nil {
		log.Warnf("Warning: Failed to parse metadata for app %s: %v\n", subnetApp.Name, err)
		metadata = nil
	}

	return &App{
		ID:                  id,
		PeerId:              subnetApp.PeerId,
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

func fillDefaultResourceUsage(usage *ResourceUsage) *ResourceUsage {
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

func convertUsageToProto(usage ResourceUsage) *pbapp.ResourceUsage {
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
		Usage:                ProtoToResourceUsage(protoApp.Usage),
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
			Preview:     metadata.AppInfo.Preview,
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

var typedDataReferenceTypeRegexp = regexp.MustCompile(`^[A-Za-z](\w*)(\[\d*\])*$`)

// TypedData is a type to encapsulate EIP-712 typed messages
type TypedData struct {
	Types       Types            `json:"types"`
	PrimaryType string           `json:"primaryType"`
	Domain      TypedDataDomain  `json:"domain"`
	Message     TypedDataMessage `json:"message"`
}

// Type is the inner type of an EIP-712 message
type Type struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// isArray returns true if the type is a fixed or variable sized array.
// This method may return false positives, in case the Type is not a valid
// expression, e.g. "fooo[[[[".
func (t *Type) isArray() bool {
	return strings.IndexByte(t.Type, '[') > 0
}

// typeName returns the canonical name of the type. If the type is 'Person[]' or 'Person[2]', then
// this method returns 'Person'
func (t *Type) typeName() string {
	return strings.Split(t.Type, "[")[0]
}

type Types map[string][]Type

type TypePriority struct {
	Type  string
	Value uint
}

type TypedDataMessage = map[string]interface{}

// TypedDataDomain represents the domain part of an EIP-712 message.
type TypedDataDomain struct {
	Name              string                `json:"name"`
	Version           string                `json:"version"`
	ChainId           *math.HexOrDecimal256 `json:"chainId"`
	VerifyingContract string                `json:"verifyingContract"`
	Salt              string                `json:"salt"`
}

// TypedDataAndHash is a helper function that calculates a hash for typed data conforming to EIP-712.
// This hash can then be safely used to calculate a signature.
//
// See https://eips.ethereum.org/EIPS/eip-712 for the full specification.
//
// This gives context to the signed typed data and prevents signing of transactions.
func TypedDataAndHash(typedData TypedData) ([]byte, string, error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, "", err
	}
	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, "", err
	}
	rawData := fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash))
	return crypto.Keccak256([]byte(rawData)), rawData, nil
}

// HashStruct generates a keccak256 hash of the encoding of the provided data
func (typedData *TypedData) HashStruct(primaryType string, data TypedDataMessage) (hexutil.Bytes, error) {
	encodedData, err := typedData.EncodeData(primaryType, data, 1)
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(encodedData), nil
}

// Dependencies returns an array of custom types ordered by their hierarchical reference tree
func (typedData *TypedData) Dependencies(primaryType string, found []string) []string {
	primaryType = strings.Split(primaryType, "[")[0]

	if slices.Contains(found, primaryType) {
		return found
	}
	if typedData.Types[primaryType] == nil {
		return found
	}
	found = append(found, primaryType)
	for _, field := range typedData.Types[primaryType] {
		for _, dep := range typedData.Dependencies(field.Type, found) {
			if !slices.Contains(found, dep) {
				found = append(found, dep)
			}
		}
	}
	return found
}

// EncodeType generates the following encoding:
// `name ‖ "(" ‖ member₁ ‖ "," ‖ member₂ ‖ "," ‖ … ‖ memberₙ ")"`
//
// each member is written as `type ‖ " " ‖ name` encodings cascade down and are sorted by name
func (typedData *TypedData) EncodeType(primaryType string) hexutil.Bytes {
	// Get dependencies primary first, then alphabetical
	deps := typedData.Dependencies(primaryType, []string{})
	if len(deps) > 0 {
		slicedDeps := deps[1:]
		sort.Strings(slicedDeps)
		deps = append([]string{primaryType}, slicedDeps...)
	}

	// Format as a string with fields
	var buffer bytes.Buffer
	for _, dep := range deps {
		buffer.WriteString(dep)
		buffer.WriteString("(")
		for _, obj := range typedData.Types[dep] {
			buffer.WriteString(obj.Type)
			buffer.WriteString(" ")
			buffer.WriteString(obj.Name)
			buffer.WriteString(",")
		}
		buffer.Truncate(buffer.Len() - 1)
		buffer.WriteString(")")
	}
	return buffer.Bytes()
}

// TypeHash creates the keccak256 hash  of the data
func (typedData *TypedData) TypeHash(primaryType string) hexutil.Bytes {
	return crypto.Keccak256(typedData.EncodeType(primaryType))
}

// EncodeData generates the following encoding:
// `enc(value₁) ‖ enc(value₂) ‖ … ‖ enc(valueₙ)`
//
// each encoded member is 32-byte long
func (typedData *TypedData) EncodeData(primaryType string, data map[string]interface{}, depth int) (hexutil.Bytes, error) {
	if err := typedData.validate(); err != nil {
		return nil, err
	}

	buffer := bytes.Buffer{}

	// Verify extra data
	if exp, got := len(typedData.Types[primaryType]), len(data); exp < got {
		return nil, fmt.Errorf("there is extra data provided in the message (%d < %d)", exp, got)
	}

	// Add typehash
	buffer.Write(typedData.TypeHash(primaryType))

	// Add field contents. Structs and arrays have special handlers.
	for _, field := range typedData.Types[primaryType] {
		encType := field.Type
		encValue := data[field.Name]
		if encType[len(encType)-1:] == "]" {
			encodedData, err := typedData.encodeArrayValue(encValue, encType, depth)
			if err != nil {
				return nil, err
			}
			buffer.Write(encodedData)
		} else if typedData.Types[field.Type] != nil {
			mapValue, ok := encValue.(map[string]interface{})
			if !ok {
				return nil, dataMismatchError(encType, encValue)
			}
			encodedData, err := typedData.EncodeData(field.Type, mapValue, depth+1)
			if err != nil {
				return nil, err
			}
			buffer.Write(crypto.Keccak256(encodedData))
		} else {
			byteValue, err := typedData.EncodePrimitiveValue(encType, encValue, depth)
			if err != nil {
				return nil, err
			}
			buffer.Write(byteValue)
		}
	}
	return buffer.Bytes(), nil
}

func (typedData *TypedData) encodeArrayValue(encValue interface{}, encType string, depth int) (hexutil.Bytes, error) {
	arrayValue, err := convertDataToSlice(encValue)
	if err != nil {
		return nil, dataMismatchError(encType, encValue)
	}

	arrayBuffer := new(bytes.Buffer)
	parsedType := strings.Split(encType, "[")[0]
	for _, item := range arrayValue {
		if reflect.TypeOf(item).Kind() == reflect.Slice ||
			reflect.TypeOf(item).Kind() == reflect.Array {
			encodedData, err := typedData.encodeArrayValue(item, parsedType, depth+1)
			if err != nil {
				return nil, err
			}
			arrayBuffer.Write(encodedData)
		} else {
			if typedData.Types[parsedType] != nil {
				mapValue, ok := item.(map[string]interface{})
				if !ok {
					return nil, dataMismatchError(parsedType, item)
				}
				encodedData, err := typedData.EncodeData(parsedType, mapValue, depth+1)
				if err != nil {
					return nil, err
				}
				digest := crypto.Keccak256(encodedData)
				arrayBuffer.Write(digest)
			} else {
				bytesValue, err := typedData.EncodePrimitiveValue(parsedType, item, depth)
				if err != nil {
					return nil, err
				}
				arrayBuffer.Write(bytesValue)
			}
		}
	}
	return crypto.Keccak256(arrayBuffer.Bytes()), nil
}

// Attempt to parse bytes in different formats: byte array, hex string, hexutil.Bytes.
func parseBytes(encType interface{}) ([]byte, bool) {
	// Handle array types.
	val := reflect.ValueOf(encType)
	if val.Kind() == reflect.Array && val.Type().Elem().Kind() == reflect.Uint8 {
		v := reflect.MakeSlice(reflect.TypeOf([]byte{}), val.Len(), val.Len())
		reflect.Copy(v, val)
		return v.Bytes(), true
	}

	switch v := encType.(type) {
	case []byte:
		return v, true
	case hexutil.Bytes:
		return v, true
	case string:
		bytes, err := hexutil.Decode(v)
		if err != nil {
			return nil, false
		}
		return bytes, true
	default:
		return nil, false
	}
}

func parseInteger(encType string, encValue interface{}) (*big.Int, error) {
	var (
		length int
		signed = strings.HasPrefix(encType, "int")
		b      *big.Int
	)
	if encType == "int" || encType == "uint" {
		length = 256
	} else {
		lengthStr := ""
		if strings.HasPrefix(encType, "uint") {
			lengthStr = strings.TrimPrefix(encType, "uint")
		} else {
			lengthStr = strings.TrimPrefix(encType, "int")
		}
		atoiSize, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid size on integer: %v", lengthStr)
		}
		length = atoiSize
	}
	switch v := encValue.(type) {
	case *math.HexOrDecimal256:
		b = (*big.Int)(v)
	case *big.Int:
		b = v
	case string:
		var hexIntValue math.HexOrDecimal256
		if err := hexIntValue.UnmarshalText([]byte(v)); err != nil {
			return nil, err
		}
		b = (*big.Int)(&hexIntValue)
	case float64:
		// JSON parses non-strings as float64. Fail if we cannot
		// convert it losslessly
		if float64(int64(v)) == v {
			b = big.NewInt(int64(v))
		} else {
			return nil, fmt.Errorf("invalid float value %v for type %v", v, encType)
		}
	}
	if b == nil {
		return nil, fmt.Errorf("invalid integer value %v/%v for type %v", encValue, reflect.TypeOf(encValue), encType)
	}
	if b.BitLen() > length {
		return nil, fmt.Errorf("integer larger than '%v'", encType)
	}
	if !signed && b.Sign() == -1 {
		return nil, fmt.Errorf("invalid negative value for unsigned type %v", encType)
	}
	return b, nil
}

// EncodePrimitiveValue deals with the primitive values found
// while searching through the typed data
func (typedData *TypedData) EncodePrimitiveValue(encType string, encValue interface{}, depth int) ([]byte, error) {
	switch encType {
	case "address":
		retval := make([]byte, 32)
		switch val := encValue.(type) {
		case string:
			if common.IsHexAddress(val) {
				copy(retval[12:], common.HexToAddress(val).Bytes())
				return retval, nil
			}
		case []byte:
			if len(val) == 20 {
				copy(retval[12:], val)
				return retval, nil
			}
		case [20]byte:
			copy(retval[12:], val[:])
			return retval, nil
		}
		return nil, dataMismatchError(encType, encValue)
	case "bool":
		boolValue, ok := encValue.(bool)
		if !ok {
			return nil, dataMismatchError(encType, encValue)
		}
		if boolValue {
			return math.PaddedBigBytes(common.Big1, 32), nil
		}
		return math.PaddedBigBytes(common.Big0, 32), nil
	case "string":
		strVal, ok := encValue.(string)
		if !ok {
			return nil, dataMismatchError(encType, encValue)
		}
		return crypto.Keccak256([]byte(strVal)), nil
	case "bytes":
		bytesValue, ok := parseBytes(encValue)
		if !ok {
			return nil, dataMismatchError(encType, encValue)
		}
		return crypto.Keccak256(bytesValue), nil
	}
	if strings.HasPrefix(encType, "bytes") {
		lengthStr := strings.TrimPrefix(encType, "bytes")
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid size on bytes: %v", lengthStr)
		}
		if length < 0 || length > 32 {
			return nil, fmt.Errorf("invalid size on bytes: %d", length)
		}
		if byteValue, ok := parseBytes(encValue); !ok || len(byteValue) != length {
			return nil, dataMismatchError(encType, encValue)
		} else {
			// Right-pad the bits
			dst := make([]byte, 32)
			copy(dst, byteValue)
			return dst, nil
		}
	}
	if strings.HasPrefix(encType, "int") || strings.HasPrefix(encType, "uint") {
		b, err := parseInteger(encType, encValue)
		if err != nil {
			return nil, err
		}
		return math.U256Bytes(new(big.Int).Set(b)), nil
	}
	return nil, fmt.Errorf("unrecognized type '%s'", encType)
}

// dataMismatchError generates an error for a mismatch between
// the provided type and data
func dataMismatchError(encType string, encValue interface{}) error {
	return fmt.Errorf("provided data '%v' doesn't match type '%s'", encValue, encType)
}

func convertDataToSlice(encValue interface{}) ([]interface{}, error) {
	var outEncValue []interface{}
	rv := reflect.ValueOf(encValue)
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			outEncValue = append(outEncValue, rv.Index(i).Interface())
		}
	} else {
		return outEncValue, fmt.Errorf("provided data '%v' is not slice", encValue)
	}
	return outEncValue, nil
}

// validate makes sure the types are sound
func (typedData *TypedData) validate() error {
	if err := typedData.Types.validate(); err != nil {
		return err
	}
	if err := typedData.Domain.validate(); err != nil {
		return err
	}
	return nil
}

// Map generates a map version of the typed data
func (typedData *TypedData) Map() map[string]interface{} {
	dataMap := map[string]interface{}{
		"types":       typedData.Types,
		"domain":      typedData.Domain.Map(),
		"primaryType": typedData.PrimaryType,
		"message":     typedData.Message,
	}
	return dataMap
}

// Format returns a representation of typedData, which can be easily displayed by a user-interface
// without in-depth knowledge about 712 rules
func (typedData *TypedData) Format() ([]*NameValueType, error) {
	domain, err := typedData.formatData("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, err
	}
	ptype, err := typedData.formatData(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}
	var nvts []*NameValueType
	nvts = append(nvts, &NameValueType{
		Name:  "EIP712Domain",
		Value: domain,
		Typ:   "domain",
	})
	nvts = append(nvts, &NameValueType{
		Name:  typedData.PrimaryType,
		Value: ptype,
		Typ:   "primary type",
	})
	return nvts, nil
}

func (typedData *TypedData) formatData(primaryType string, data map[string]interface{}) ([]*NameValueType, error) {
	var output []*NameValueType

	// Add field contents. Structs and arrays have special handlers.
	for _, field := range typedData.Types[primaryType] {
		encName := field.Name
		encValue := data[encName]
		item := &NameValueType{
			Name: encName,
			Typ:  field.Type,
		}
		if field.isArray() {
			arrayValue, _ := convertDataToSlice(encValue)
			parsedType := field.typeName()
			for _, v := range arrayValue {
				if typedData.Types[parsedType] != nil {
					mapValue, _ := v.(map[string]interface{})
					mapOutput, err := typedData.formatData(parsedType, mapValue)
					if err != nil {
						return nil, err
					}
					item.Value = mapOutput
				} else {
					primitiveOutput, err := formatPrimitiveValue(field.Type, encValue)
					if err != nil {
						return nil, err
					}
					item.Value = primitiveOutput
				}
			}
		} else if typedData.Types[field.Type] != nil {
			if mapValue, ok := encValue.(map[string]interface{}); ok {
				mapOutput, err := typedData.formatData(field.Type, mapValue)
				if err != nil {
					return nil, err
				}
				item.Value = mapOutput
			} else {
				item.Value = "<nil>"
			}
		} else {
			primitiveOutput, err := formatPrimitiveValue(field.Type, encValue)
			if err != nil {
				return nil, err
			}
			item.Value = primitiveOutput
		}
		output = append(output, item)
	}
	return output, nil
}

func formatPrimitiveValue(encType string, encValue interface{}) (string, error) {
	switch encType {
	case "address":
		if stringValue, ok := encValue.(string); !ok {
			return "", fmt.Errorf("could not format value %v as address", encValue)
		} else {
			return common.HexToAddress(stringValue).String(), nil
		}
	case "bool":
		if boolValue, ok := encValue.(bool); !ok {
			return "", fmt.Errorf("could not format value %v as bool", encValue)
		} else {
			return fmt.Sprintf("%t", boolValue), nil
		}
	case "bytes", "string":
		return fmt.Sprintf("%s", encValue), nil
	}
	if strings.HasPrefix(encType, "bytes") {
		return fmt.Sprintf("%s", encValue), nil
	}
	if strings.HasPrefix(encType, "uint") || strings.HasPrefix(encType, "int") {
		if b, err := parseInteger(encType, encValue); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%d (%#x)", b, b), nil
		}
	}
	return "", fmt.Errorf("unhandled type %v", encType)
}

// validate checks if the types object is conformant to the specs
func (t Types) validate() error {
	for typeKey, typeArr := range t {
		if len(typeKey) == 0 {
			return errors.New("empty type key")
		}
		for i, typeObj := range typeArr {
			if len(typeObj.Type) == 0 {
				return fmt.Errorf("type %q:%d: empty Type", typeKey, i)
			}
			if len(typeObj.Name) == 0 {
				return fmt.Errorf("type %q:%d: empty Name", typeKey, i)
			}
			if typeKey == typeObj.Type {
				return fmt.Errorf("type %q cannot reference itself", typeObj.Type)
			}
			if isPrimitiveTypeValid(typeObj.Type) {
				continue
			}
			// Must be reference type
			if _, exist := t[typeObj.typeName()]; !exist {
				return fmt.Errorf("reference type %q is undefined", typeObj.Type)
			}
			if !typedDataReferenceTypeRegexp.MatchString(typeObj.Type) {
				return fmt.Errorf("unknown reference type %q", typeObj.Type)
			}
		}
	}
	return nil
}

var validPrimitiveTypes = map[string]struct{}{}

// build the set of valid primitive types
func init() {
	// Types those are trivially valid
	for _, t := range []string{
		"address", "address[]", "bool", "bool[]", "string", "string[]",
		"bytes", "bytes[]", "int", "int[]", "uint", "uint[]",
	} {
		validPrimitiveTypes[t] = struct{}{}
	}
	// For 'bytesN', 'bytesN[]', we allow N from 1 to 32
	for n := 1; n <= 32; n++ {
		validPrimitiveTypes[fmt.Sprintf("bytes%d", n)] = struct{}{}
		validPrimitiveTypes[fmt.Sprintf("bytes%d[]", n)] = struct{}{}
	}
	// For 'intN','intN[]' and 'uintN','uintN[]' we allow N in increments of 8, from 8 up to 256
	for n := 8; n <= 256; n += 8 {
		validPrimitiveTypes[fmt.Sprintf("int%d", n)] = struct{}{}
		validPrimitiveTypes[fmt.Sprintf("int%d[]", n)] = struct{}{}
		validPrimitiveTypes[fmt.Sprintf("uint%d", n)] = struct{}{}
		validPrimitiveTypes[fmt.Sprintf("uint%d[]", n)] = struct{}{}
	}
}

// Checks if the primitive value is valid
func isPrimitiveTypeValid(primitiveType string) bool {
	input := strings.Split(primitiveType, "[")[0]
	_, ok := validPrimitiveTypes[input]
	return ok
}

// validate checks if the given domain is valid, i.e. contains at least
// the minimum viable keys and values
func (domain *TypedDataDomain) validate() error {
	if domain.ChainId == nil && len(domain.Name) == 0 && len(domain.Version) == 0 && len(domain.VerifyingContract) == 0 && len(domain.Salt) == 0 {
		return errors.New("domain is undefined")
	}

	return nil
}

// Map is a helper function to generate a map version of the domain
func (domain *TypedDataDomain) Map() map[string]interface{} {
	dataMap := map[string]interface{}{}

	if domain.ChainId != nil {
		dataMap["chainId"] = domain.ChainId
	}

	if len(domain.Name) > 0 {
		dataMap["name"] = domain.Name
	}

	if len(domain.Version) > 0 {
		dataMap["version"] = domain.Version
	}

	if len(domain.VerifyingContract) > 0 {
		dataMap["verifyingContract"] = domain.VerifyingContract
	}

	if len(domain.Salt) > 0 {
		dataMap["salt"] = domain.Salt
	}
	return dataMap
}

// NameValueType is a very simple struct with Name, Value and Type. It's meant for simple
// json structures used to communicate signing-info about typed data with the UI
type NameValueType struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Typ   string      `json:"type"`
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
