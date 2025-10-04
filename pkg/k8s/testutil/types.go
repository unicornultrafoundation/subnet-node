package testutil

import (
	"math/rand"
	"testing"

	types "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/base/v1"
)

func RandRangeInt(min, max int) int {
	return rand.Intn(max-min) + min // nolint: gosec
}

func RandRangeUint(min, max uint) uint {
	val := rand.Uint64() // nolint: gosec
	val %= uint64(max - min)
	val += uint64(min)
	return uint(val)
}

func RandRangeUint64(min, max uint64) uint64 {
	val := rand.Uint64() // nolint: gosec
	val %= max - min
	val += min
	return val
}

func ResourceUnits(_ testing.TB) types.Resources {
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
