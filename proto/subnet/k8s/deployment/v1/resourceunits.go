package v1

import (
	"fmt"
	"sort"

	"github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/utils"
)

type ResourceUnits []ResourceUnit

var _ sort.Interface = (*ResourceUnits)(nil)

func (s ResourceUnits) Len() int {
	return len(s)
}

func (s ResourceUnits) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ResourceUnits) Less(i, j int) bool {
	return s[i].ID < s[j].ID
}

func (s ResourceUnits) Dup() ResourceUnits {
	res := make(ResourceUnits, 0, len(s))

	for _, ru := range s {
		res = append(res, ru.Dup())
	}

	return res
}

func (s ResourceUnits) Validate() error {
	// if count := len(s); count > validationConfig.MaxGroupUnits {
	// 	return fmt.Errorf("too many units (%v > %v)", count, validationConfig.MaxGroupUnits)
	// }

	ids := make(map[uint32]bool)
	for _, res := range s {
		if err := res.validate(); err != nil {
			return err
		}
		if res.ID == 0 {
			return fmt.Errorf("resources ID must be > 0")
		}

		if _, exists := ids[res.ID]; exists {
			return fmt.Errorf("duplicate resources ID (%d) within group", res.ID)
		}

		ids[res.ID] = true
	}

	limits := newLimits()

	for idx := range s {
		limits.add(s[idx].totalResources())
	}

	cpuLimit := utils.NewInt(int64(validationConfig.Group.Max.CPU))
	gpuLimit := utils.NewInt(int64(validationConfig.Group.Max.GPU))
	memoryLimit := utils.NewInt(int64(validationConfig.Group.Max.Memory))
	storageLimit := utils.NewInt(int64(validationConfig.Group.Max.Storage))
	zero := utils.NewInt(0)

	if limits.cpu.GT(&cpuLimit) || limits.cpu.LTE(&zero) {
		return fmt.Errorf("invalid total CPU (%v > %v > %v fails)", validationConfig.Group.Max.CPU, limits.cpu, 0)
	}

	if !limits.gpu.IsZero() && (limits.gpu.GT(&gpuLimit) || limits.gpu.LTE(&zero)) {
		return fmt.Errorf("invalid total GPU (%v > %v > %v fails)", validationConfig.Group.Max.GPU, limits.gpu, 0)
	}

	if limits.memory.GT(&memoryLimit) || limits.memory.LTE(&zero) {
		return fmt.Errorf("invalid total memory (%v > %v > %v fails)", validationConfig.Group.Max.Memory, limits.memory, 0)
	}

	for i := range limits.storage {
		if limits.storage[i].GT(&storageLimit) || limits.storage[i].LTE(&zero) {
			return fmt.Errorf("invalid total storage (%v > %v > %v fails)", validationConfig.Group.Max.Storage, limits.storage, 0)
		}
	}

	return nil
}
