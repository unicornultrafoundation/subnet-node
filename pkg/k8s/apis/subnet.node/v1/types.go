package v1

import (
	"fmt"
	"math"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	types "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/base/v1"
	mani "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
	mtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"
)

var (
	ErrInvalidArgs = fmt.Errorf("crd/%s: invalid args", crdVersion)
)

type Status struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}

type deployment struct {
	lid             mtypes.LeaseID
	group           mani.Group
	cparams         interface{}
	resourceVersion string
}

func (d deployment) LeaseID() mtypes.LeaseID {
	return d.lid
}

func (d deployment) ManifestGroup() *mani.Group {
	return &d.group
}

func (d deployment) ClusterParams() interface{} {
	return d.cparams
}

func (d deployment) ResourceVersion() string {
	return d.resourceVersion
}

// LeaseID stores deployment, group sequence, order, provider and metadata
type LeaseID struct {
	Owner    string `json:"owner"`
	DSeq     string `json:"dseq"`
	GSeq     uint32 `json:"gseq"`
	OSeq     uint32 `json:"oseq"`
	Provider string `json:"provider"`
}

// FromCRD returns LeaseID from LeaseID details
func (id LeaseID) FromCRD() (mtypes.LeaseID, error) {
	if !common.IsHexAddress(id.Owner) {
		return mtypes.LeaseID{}, errors.New("invalid owner address")
	}

	owner := common.HexToAddress(id.Owner)

	provider := common.HexToAddress(id.Provider)

	dseq, err := strconv.ParseUint(id.DSeq, 10, 64)
	if err != nil {
		return mtypes.LeaseID{}, err
	}

	return mtypes.LeaseID{
		Owner:    owner.String(),
		DSeq:     dseq,
		GSeq:     id.GSeq,
		OSeq:     id.OSeq,
		Provider: provider.String(),
	}, nil
}

// LeaseIDFromSubnet returns LeaseID instance from subnet
func LeaseIDFromSubnet(id mtypes.LeaseID) LeaseID {
	return LeaseID{
		Owner:    id.Owner,
		DSeq:     strconv.FormatUint(id.DSeq, 10),
		GSeq:     id.GSeq,
		OSeq:     id.OSeq,
		Provider: id.Provider,
	}
}

type ResourceCPU struct {
	Units      uint32           `json:"units"`
	Attributes types.Attributes `json:"attributes,omitempty"`
}

type ResourceGPU struct {
	Units      uint32           `json:"units"`
	Attributes types.Attributes `json:"attributes,omitempty"`
}

type ResourceMemory struct {
	Size       string           `json:"size"`
	Attributes types.Attributes `json:"attributes,omitempty"`
}

type ResourceVolume struct {
	Name       string           `json:"name"`
	Size       string           `json:"size"`
	Attributes types.Attributes `json:"attributes,omitempty"`
}

type ResourceStorage []ResourceVolume

func (ru ResourceCPU) ToSubnet() *types.CPU {
	return &types.CPU{
		Units:      types.NewResourceValue(uint64(ru.Units)),
		Attributes: ru.Attributes.Dup(),
	}
}

func (ru ResourceGPU) ToSubnet() *types.GPU {
	return &types.GPU{
		Units:      types.NewResourceValue(uint64(ru.Units)),
		Attributes: ru.Attributes.Dup(),
	}
}

func (ru ResourceMemory) ToSubnet() *types.Memory {
	size, _ := strconv.ParseUint(ru.Size, 10, 64)

	return &types.Memory{
		Quantity:   types.NewResourceValue(size),
		Attributes: ru.Attributes.Dup(),
	}
}

func (ru ResourceVolume) ToSubnet() types.Storage {
	size, _ := strconv.ParseUint(ru.Size, 10, 64)

	return types.Storage{
		Name:       ru.Name,
		Quantity:   types.NewResourceValue(size),
		Attributes: ru.Attributes.Dup(),
	}
}

func (ru ResourceStorage) ToSubnet() types.Volumes {
	res := make(types.Volumes, 0, len(ru))

	for _, vol := range ru {
		res = append(res, vol.ToSubnet())
	}

	return res
}

// Resources stores cpu, memory and storage details
type Resources struct {
	ID      uint32          `json:"id"`
	CPU     ResourceCPU     `json:"cpu"`
	GPU     ResourceGPU     `json:"gpu"`
	Memory  ResourceMemory  `json:"memory"`
	Storage ResourceStorage `json:"storage,omitempty"`
}

func (ru *Resources) ToSubnet() (types.Resources, error) {
	return types.Resources{
		ID:      ru.ID,
		CPU:     ru.CPU.ToSubnet(),
		GPU:     ru.GPU.ToSubnet(),
		Memory:  ru.Memory.ToSubnet(),
		Storage: ru.Storage.ToSubnet(),
	}, nil
}

func resourcesFromSubnet(aru types.Resources) (Resources, error) {
	res := Resources{
		ID: aru.ID,
	}

	if aru.CPU != nil {
		if aru.CPU.Units.Value() > math.MaxUint32 {
			return Resources{}, errors.New("k8s api: cpu units value overflows uint32")
		}
		res.CPU.Units = uint32(aru.CPU.Units.Value()) // nolint: gosec
		res.CPU.Attributes = aru.CPU.Attributes.Dup()
	}

	if aru.Memory != nil {
		res.Memory.Size = strconv.FormatUint(aru.Memory.Quantity.Value(), 10)
		res.Memory.Attributes = aru.Memory.Attributes.Dup()
	}

	if aru.GPU != nil {
		if aru.GPU.Units.Value() > math.MaxUint32 {
			return Resources{}, errors.New("k8s api: gpu units value overflows uint32")
		}
		res.GPU.Units = uint32(aru.GPU.Units.Value()) // nolint: gosec
		res.GPU.Attributes = aru.GPU.Attributes
	}

	res.Storage = make(ResourceStorage, 0, len(aru.Storage))
	for _, storage := range aru.Storage {
		res.Storage = append(res.Storage, ResourceVolume{
			Name:       storage.Name,
			Size:       strconv.FormatUint(storage.Quantity.Value(), 10),
			Attributes: storage.Attributes.Dup(),
		})
	}

	return res, nil
}
