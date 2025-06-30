package manifest

import (
	"fmt"
	"strconv"
	"strings"
)

// parseRAMString parses a RAM string like "8192MB" or "8GB" and returns the value in MB.
func parseRAMString(ram string) (int64, error) {
	ram = strings.TrimSpace(ram)
	if strings.HasSuffix(ram, "MB") {
		val := strings.TrimSuffix(ram, "MB")
		mb, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
		if err != nil {
			return 0, err
		}
		return mb, nil
	} else if strings.HasSuffix(ram, "GB") {
		val := strings.TrimSuffix(ram, "GB")
		gb, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
		if err != nil {
			return 0, err
		}
		return gb * 1024, nil
	}
	return 0, fmt.Errorf("unsupported RAM unit: %s", ram)
}

// CalculateGlobalResourceCap calculates the total resources for all services/replicas
func (m *Manifest) CalculateGlobalResourceCap() (cpuCores float64, gpuCores int64, gpuMemoryMB int64, memoryMB int64, diskGB int64, err error) {
	cpuCores = 0
	gpuCores = 0
	gpuMemoryMB = 0
	memoryMB = 0
	diskGB = 0

	for _, group := range m.Groups {
		for _, service := range group.Services {
			count := service.Count
			if count <= 0 {
				count = 1
			}
			if service.Resources != nil {
				// CPU
				if service.Resources.CPU != nil {
					cpuValue := service.Resources.CPU.Units.Value
					cpuUnit := service.Resources.CPU.Units.Unit
					switch cpuUnit {
					case "", "m":
						cpuCores += float64(cpuValue) * float64(count) / 1000.0
					case "core", "cores":
						cpuCores += float64(cpuValue) * float64(count)
					default:
						err = fmt.Errorf("unsupported CPU unit: %s", cpuUnit)
						return
					}
				}
				// Memory
				if service.Resources.Memory != nil {
					memValue := service.Resources.Memory.Size.Value
					memUnit := service.Resources.Memory.Size.Unit
					switch memUnit {
					case "", "Mi", "MiB":
						memoryMB += memValue * int64(count)
					case "Gi", "GiB", "GB":
						memoryMB += memValue * int64(count) * 1024
					case "Ki", "KiB", "kB":
						memoryMB += (memValue * int64(count)) / 1024
					case "MB":
						memoryMB += memValue * int64(count)
					case "B", "bytes":
						memoryMB += (memValue * int64(count)) / (1024 * 1024)
					default:
						err = fmt.Errorf("unsupported memory unit: %s", memUnit)
						return
					}
				}
				// Disk (Storage)
				if service.Resources.Storage != nil {
					storageValue := service.Resources.Storage.Size.Value
					unit := service.Resources.Storage.Size.Unit
					switch unit {
					case "Gi", "GiB", "GB":
						diskGB += storageValue * int64(count)
					case "Mi", "MiB", "MB":
						diskGB += (storageValue * int64(count)) / 1024
					case "Ki", "KiB", "kB":
						diskGB += (storageValue * int64(count)) / (1024 * 1024)
					case "B", "bytes":
						diskGB += (storageValue * int64(count)) / (1024 * 1024 * 1024)
					case "":
						// Assume bytes if unit is empty
						diskGB += (storageValue * int64(count)) / (1024 * 1024 * 1024)
					default:
						err = fmt.Errorf("unsupported storage unit: %s", unit)
						return
					}
				}
				// GPU
				if service.Resources.GPU != nil {
					gpuCores += int64(service.Resources.GPU.Units) * int64(count)
					// GPU Memory: sum for each model in Attributes.Vendor
					for _, models := range service.Resources.GPU.Attributes.Vendor {
						for _, model := range models {
							ramMB, parseErr := parseRAMString(model.RAM)
							if parseErr != nil {
								err = parseErr
								return
							}
							gpuMemoryMB += ramMB * int64(service.Resources.GPU.Units) * int64(count)
						}
					}
				}
			}
		}
	}

	return cpuCores, gpuCores, gpuMemoryMB, memoryMB, diskGB, nil
}
