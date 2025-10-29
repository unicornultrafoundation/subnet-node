package inventory

// GPUDevice represents a single GPU device
type GPUDevice struct {
	Name       string `json:"name"`
	Interface  string `json:"interface"`
	MemorySize string `json:"memory_size"`
}

// GPUVendor represents a GPU vendor with their devices
type GPUVendor struct {
	Name    string                `json:"name"`
	Devices map[string]*GPUDevice `json:"devices"`
}

// GPURegistry contains all GPU vendor and device information
type GPURegistry map[string]*GPUVendor

// GetGPURegistry returns the complete GPU registry with vendor and device information
func GetGPURegistry() GPURegistry {
	return GPURegistry{
		"1002": {
			Name: "amd",
			Devices: map[string]*GPUDevice{
				"66a1": {
					Name:       "mi60",
					Interface:  "PCIe",
					MemorySize: "32Gi",
				},
				"738c": {
					Name:       "mi100",
					Interface:  "PCIe",
					MemorySize: "32Gi",
				},
			},
		},
		"10de": {
			Name: "nvidia",
			Devices: map[string]*GPUDevice{
				"1401": {
					Name:       "gtx960",
					Interface:  "PCIe",
					MemorySize: "2Gi",
				},
				"1406": {
					Name:       "gtx960",
					Interface:  "PCIe",
					MemorySize: "4Gi",
				},
				"2182": {
					Name:       "gtx1660ti",
					Interface:  "PCIe",
					MemorySize: "6Gi",
				},
				"2184": {
					Name:       "gtx1660",
					Interface:  "PCIe",
					MemorySize: "6Gi",
				},
				"2187": {
					Name:       "gtx1650super",
					Interface:  "PCIe",
					MemorySize: "4Gi",
				},
				"2203": {
					Name:       "rtx3090ti",
					Interface:  "PCIe",
					MemorySize: "24Gi",
				},
				"2204": {
					Name:       "rtx3090",
					Interface:  "PCIe",
					MemorySize: "24Gi",
				},
				"2216": {
					Name:       "rtx3080",
					Interface:  "PCIe",
					MemorySize: "10Gi",
				},
				"2230": {
					Name:       "rtxa6000",
					Interface:  "PCIe",
					MemorySize: "48Gi",
				},
				"2235": {
					Name:       "a40",
					Interface:  "PCIe",
					MemorySize: "48Gi",
				},
				"2330": {
					Name:       "h100",
					Interface:  "SXM5",
					MemorySize: "80Gi",
				},
				"2331": {
					Name:       "h100",
					Interface:  "PCIe",
					MemorySize: "80Gi",
				},
				"2335": {
					Name:       "h200",
					Interface:  "SXM5",
					MemorySize: "141Gi",
				},
				"2484": {
					Name:       "rtx3070",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"2486": {
					Name:       "rtx3060ti",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"2487": {
					Name:       "rtx3060",
					Interface:  "PCIe",
					MemorySize: "12Gi",
				},
				"2488": {
					Name:       "rtx3070",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"2489": {
					Name:       "rtx3060ti",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"2504": {
					Name:       "rtx3060",
					Interface:  "PCIe",
					MemorySize: "12Gi",
				},
				"2531": {
					Name:       "rtxa2000",
					Interface:  "PCIe",
					MemorySize: "6Gi",
				},
				"2684": {
					Name:       "rtx4090",
					Interface:  "PCIe",
					MemorySize: "24Gi",
				},
				"2702": {
					Name:       "rtx4080super",
					Interface:  "PCIe",
					MemorySize: "16Gi",
				},
				"2704": {
					Name:       "rtx4080",
					Interface:  "PCIe",
					MemorySize: "16Gi",
				},
				"2782": {
					Name:       "rtx4070ti",
					Interface:  "PCIe",
					MemorySize: "12Gi",
				},
				"2786": {
					Name:       "rtx4070",
					Interface:  "PCIe",
					MemorySize: "12Gi",
				},
				"2803": {
					Name:       "rtx4060ti",
					Interface:  "PCIe",
					MemorySize: "16Gi",
				},
				"2805": {
					Name:       "rtx4060ti",
					Interface:  "PCIe",
					MemorySize: "16Gi",
				},
				"2882": {
					Name:       "rtx4060",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"13c0": {
					Name:       "gtx980",
					Interface:  "PCIe",
					MemorySize: "4Gi",
				},
				"13c2": {
					Name:       "gtx970",
					Interface:  "PCIe",
					MemorySize: "4Gi",
				},
				"13f1": {
					Name:       "m4000",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"147f": {
					Name:       "a100",
					Interface:  "SXM4",
					MemorySize: "80Gi",
				},
				"15f8": {
					Name:       "p100",
					Interface:  "PCIe",
					MemorySize: "16Gi",
				},
				"17c8": {
					Name:       "gtx980ti",
					Interface:  "PCIe",
					MemorySize: "6Gi",
				},
				"1b06": {
					Name:       "gtx1080ti",
					Interface:  "PCIe",
					MemorySize: "11Gi",
				},
				"1b30": {
					Name:       "p6000",
					Interface:  "PCIe",
					MemorySize: "24Gi",
				},
				"1b38": {
					Name:       "p40",
					Interface:  "PCIe",
					MemorySize: "24Gi",
				},
				"1b80": {
					Name:       "gtx1080",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1b81": {
					Name:       "gtx1070",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1b82": {
					Name:       "gtx1070ti",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1bb0": {
					Name:       "p5000",
					Interface:  "PCIe",
					MemorySize: "16Gi",
				},
				"1bb1": {
					Name:       "p4000",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1bb3": {
					Name:       "p4",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1c02": {
					Name:       "gtx1060",
					Interface:  "PCIe",
					MemorySize: "3Gi",
				},
				"1c03": {
					Name:       "gtx1060",
					Interface:  "PCIe",
					MemorySize: "6Gi",
				},
				"1c04": {
					Name:       "gtx1060",
					Interface:  "PCIe",
					MemorySize: "5Gi",
				},
				"1c06": {
					Name:       "gtx1060",
					Interface:  "PCIe",
					MemorySize: "6Gi",
				},
				"1c30": {
					Name:       "p2000",
					Interface:  "PCIe",
					MemorySize: "5Gi",
				},
				"1c81": {
					Name:       "gtx1050",
					Interface:  "PCIe",
					MemorySize: "2Gi",
				},
				"1c82": {
					Name:       "gtx1050ti",
					Interface:  "PCIe",
					MemorySize: "4Gi",
				},
				"1c83": {
					Name:       "gtx1050",
					Interface:  "PCIe",
					MemorySize: "3Gi",
				},
				"1cba": {
					Name:       "p2000",
					Interface:  "PCIe",
					MemorySize: "4Gi",
				},
				"1d01": {
					Name:       "gt1030",
					Interface:  "PCIe",
					MemorySize: "2Gi",
				},
				"1d10": {
					Name:       "mx150",
					Interface:  "PCIe",
					MemorySize: "2Gi",
				},
				"1db1": {
					Name:       "v100",
					Interface:  "SXM2",
					MemorySize: "16Gi",
				},
				"1db5": {
					Name:       "v100",
					Interface:  "SXM2",
					MemorySize: "32Gi",
				},
				"1db8": {
					Name:       "v100",
					Interface:  "SXM3",
					MemorySize: "32Gi",
				},
				"1e04": {
					Name:       "rtx2080ti",
					Interface:  "PCIe",
					MemorySize: "11Gi",
				},
				"1e07": {
					Name:       "rtx2080ti",
					Interface:  "PCIe",
					MemorySize: "11Gi",
				},
				"1e30": {
					Name:       "rtx8000",
					Interface:  "PCIe",
					MemorySize: "48Gi",
				},
				"1e78": {
					Name:       "rtx8000",
					Interface:  "PCIe",
					MemorySize: "48Gi",
				},
				"1e81": {
					Name:       "rtx2080super",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1e84": {
					Name:       "rtx2070super",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1e87": {
					Name:       "rtx2080",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1e89": {
					Name:       "rtx2060",
					Interface:  "PCIe",
					MemorySize: "6Gi",
				},
				"1eb1": {
					Name:       "rtx4000",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1eb5": {
					Name:       "rtx5000",
					Interface:  "PCIe",
					MemorySize: "16Gi",
				},
				"1eb8": {
					Name:       "t4",
					Interface:  "PCIe",
					MemorySize: "16Gi",
				},
				"1f02": {
					Name:       "rtx2070",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1f06": {
					Name:       "rtx2060super",
					Interface:  "PCIe",
					MemorySize: "8Gi",
				},
				"1f95": {
					Name:       "gtx1650ti",
					Interface:  "PCIe",
					MemorySize: "4Gi",
				},
				"20b0": {
					Name:       "a100",
					Interface:  "SXM4",
					MemorySize: "40Gi",
				},
				"20b1": {
					Name:       "a100",
					Interface:  "PCIe",
					MemorySize: "40Gi",
				},
				"20b2": {
					Name:       "a100",
					Interface:  "SXM4",
					MemorySize: "80Gi",
				},
				"20b3": {
					Name:       "a100",
					Interface:  "SXM",
					MemorySize: "64Gi",
				},
				"20b5": {
					Name:       "a100",
					Interface:  "PCIe",
					MemorySize: "80Gi",
				},
				"20f1": {
					Name:       "a100",
					Interface:  "PCIe",
					MemorySize: "40Gi",
				},
				"20f3": {
					Name:       "a800",
					Interface:  "SXM4",
					MemorySize: "80Gi",
				},
				"27b0": {
					Name:       "rtx4000sffada",
					Interface:  "PCIe",
					MemorySize: "20Gi",
				},
				"27b2": {
					Name:       "rtx4000ada",
					Interface:  "PCIe",
					MemorySize: "20Gi",
				},
				"28b0": {
					Name:       "rtx2000eada",
					Interface:  "PCIe",
					MemorySize: "16Gi",
				},
				"2b85": {
					Name:       "rtx5090",
					Interface:  "PCIe",
					MemorySize: "32Gi",
				},
				"f297": {
					Name:       "rtx4090",
					Interface:  "PCIe",
					MemorySize: "24Gi",
				},
			},
		},
	}
}

// GetGPUInfo returns GPU device information by vendor ID and device ID
func GetGPUInfo(vendorID, deviceID string) (*GPUDevice, bool) {
	registry := GetGPURegistry()
	vendor, exists := registry[vendorID]
	if !exists {
		return nil, false
	}

	device, exists := vendor.Devices[deviceID]
	return device, exists
}

// GetVendorName returns the vendor name by vendor ID
func GetVendorName(vendorID string) (string, bool) {
	registry := GetGPURegistry()
	vendor, exists := registry[vendorID]
	if !exists {
		return "", false
	}
	return vendor.Name, true
}

// GetAllVendors returns all vendor IDs and names
func GetAllVendors() map[string]string {
	registry := GetGPURegistry()
	vendors := make(map[string]string)
	for vendorID, vendor := range registry {
		vendors[vendorID] = vendor.Name
	}
	return vendors
}

// GetDevicesByVendor returns all devices for a specific vendor
func GetDevicesByVendor(vendorID string) (map[string]*GPUDevice, bool) {
	registry := GetGPURegistry()
	vendor, exists := registry[vendorID]
	if !exists {
		return nil, false
	}
	return vendor.Devices, true
}
