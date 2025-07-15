package v1

import (
	"math/big"

	"github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/utils"
)

type resourceLimits struct {
	cpu     utils.BigInt
	gpu     utils.BigInt
	memory  utils.BigInt
	storage []utils.BigInt
}

func newLimits() resourceLimits {
	return resourceLimits{
		cpu:    utils.NewInt(0),
		gpu:    utils.NewInt(0),
		memory: utils.NewInt(0),
	}
}

func (u *resourceLimits) add(rhs resourceLimits) {
	u.cpu = utils.NewInt(u.cpu.Int.Add(&u.cpu.Int, &rhs.cpu.Int).Int64())
	u.gpu = utils.NewInt(u.gpu.Int.Add(&u.gpu.Int, &rhs.gpu.Int).Int64())
	u.memory = utils.NewInt(u.memory.Int.Add(&u.memory.Int, &rhs.memory.Int).Int64())

	// u.storage = u.storage.Add(rhs.storage)
}

func (u *resourceLimits) mul(count uint32) {
	u.cpu = utils.NewInt(u.cpu.Int.Mul(&u.cpu.Int, big.NewInt(int64(count))).Int64())
	u.gpu = utils.NewInt(u.gpu.Int.Mul(&u.gpu.Int, big.NewInt(int64(count))).Int64())
	u.memory = utils.NewInt(u.memory.Int.Mul(&u.memory.Int, big.NewInt(int64(count))).Int64())

	for i := range u.storage {
		u.storage[i] = utils.NewInt(u.storage[i].Int.Mul(&u.storage[i].Int, big.NewInt(int64(count))).Int64())
	}
}
