package apps

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
	AppInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Logo        string `json:"logo"`
		Website     string `json:"website"`
	} `json:"appInfo"`
	ContainerConfig struct {
		Image     string            `json:"image"`
		Command   []string          `json:"command"`
		Env       map[string]string `json:"env"`
		Resources struct {
			CPU     string `json:"cpu"`
			Memory  string `json:"memory"`
			Storage string `json:"storage"`
		} `json:"resources"`
		Ports []struct {
			ContainerPort int    `json:"containerPort"`
			Protocol      string `json:"protocol"`
		} `json:"ports"`
		Volumes []struct {
			Name      string `json:"name"`
			MountPath string `json:"mountPath"`
		} `json:"volumes"`
	} `json:"containerConfig"`
	ContactInfo struct {
		Email  string `json:"email"`
		Github string `json:"github"`
	} `json:"contactInfo"`
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

func convertToApp(subnetApp SubnetAppRegistryApp, id *big.Int, status ProcessStatus) *App {
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
