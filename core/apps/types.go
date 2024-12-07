package apps

import (
	"math/big"

	"github.com/containerd/containerd"
)

type App struct {
	ID                   big.Int  `json:"id"`
	PeerId               string   `json:"peerId"`
	Owner                string   `json:"owner"`
	Name                 string   `json:"name"`
	Symbol               string   `json:"symbol"`
	Budget               *big.Int `json:"budget"`
	SpentBudget          *big.Int `json:"spentBudget"`
	MaxNodes             uint64   `json:"maxNodes"`
	MinCpu               uint64   `json:"minCpu"`
	MinGpu               uint64   `json:"minGpu"`
	MinMemory            uint64   `json:"minMemory"`
	MinUploadBandwidth   uint64   `json:"minUploadBandwidth"`
	MinDownloadBandwidth uint64   `json:"minDownloadBandwidth"`
	NodeCount            uint64   `json:"nodeCount"`
	PricePerCpu          uint64   `json:"pricePerCpu"`
	PricePerGpu          uint64   `json:"pricePerGpu"`
	PricePerMemoryGB     uint64   `json:"pricePerMemoryGB"`
	PricePerStorageGB    uint64   `json:"pricePerStorageGB"`
	PricePerBandwidthGB  uint64   `json:"pricePerBandwidthGB"`
	PaymentMethod        uint8    `json:"paymentMethod"`
	Status               containerd.ProcessStatus
}

// ResourceUsage represents the resource usage data
type ResourceUsage struct {
	AppId             big.Int `json:"appId"`
	SubnetId          big.Int `json:"subnetId"`
	UsedCpu           uint64  `json:"usedCpu"`
	UsedGpu           uint64  `json:"usedGpu"`
	UsedMemory        uint64  `json:"usedMemory"`
	UsedStorage       uint64  `json:"usedStorage"`
	UsedUploadBytes   uint64  `json:"usedUploadBytes"`
	UsedDownloadBytes uint64  `json:"usedDownloadBytes"`
	Duration          uint64  `json:"duration"`
}

// SignatureResponse represents the response containing the signature
type SignatureResponse struct {
	Signature string `json:"signature"`
}
