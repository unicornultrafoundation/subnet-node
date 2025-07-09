package sdl

import "fmt"

type gpuInterface string

type v1GPUNvidia struct {
	Model     string          `yaml:"model"`
	RAM       *memoryQuantity `yaml:"ram,omitempty"`
	Interface *gpuInterface   `yaml:"interface,omitempty"`
}

func (sdl *v1GPUNvidia) String() string {
	key := sdl.Model
	if sdl.RAM != nil {
		key = fmt.Sprintf("%s/ram/%s", key, sdl.RAM.StringWithSuffix("Gi"))
	}

	if sdl.Interface != nil {
		key = fmt.Sprintf("%s/interface/%s", key, *sdl.Interface)
	}

	return key
}

type v1GPUsNvidia []v1GPUNvidia

type gpuVendor struct {
	Nvidia v1GPUsNvidia `yaml:"nvidia"`
}

type v1GPU struct {
	Units  gpuQuantity `yaml:"units"`
	Vendor gpuVendor   `yaml:"vendor"`
}
