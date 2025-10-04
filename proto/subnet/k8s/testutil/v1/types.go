package testutil

import (
	"testing"

	types "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/base/v1"
)

func Resources(_ testing.TB) types.Resources {
	return types.Resources{
		ID: 1,
		CPU: &types.CPU{
			Units: types.NewResourceValue(uint64(RandCPUUnits())),
		},
		Memory: &types.Memory{
			Quantity: types.NewResourceValue(RandMemoryQuantity()),
		},
		GPU: &types.GPU{
			Units: types.NewResourceValue(uint64(RandGPUUnits())),
		},
		Storage: types.Volumes{
			types.Storage{
				Quantity: types.NewResourceValue(RandStorageQuantity()),
			},
		},
	}
}
