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
