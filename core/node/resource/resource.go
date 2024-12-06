package resource

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/hashicorp/go-multierror"
	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/showwin/speedtest-go/speedtest"
	"github.com/sirupsen/logrus"
)

var log = logrus.New().WithField("node", "resource")

type ResourceInfo struct {
	Region    RegionInfo    `json:"region"`
	PeerID    string        `json:"peer_id"`
	CPU       CpuInfo       `json:"cpu"`       // CPU cores
	GPU       GpuInfo       `json:"gpu"`       // GPU cores
	Memory    MemoryInfo    `json:"memory"`    // RAM (MB)
	Bandwidth BandwidthInfo `json:"bandwidth"` // Bandwidth (Mbps)
}

type RegionInfo struct {
	Region  string `json:"region"`
	Country string `json:"country"`
}

type MemoryInfo struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

type CpuInfo struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
}

type GpuInfo struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
}

type BandwidthInfo struct {
	Latency       time.Duration `json:"latency"`
	UploadSpeed   uint64        `json:"upload"`
	DownloadSpeed uint64        `json:"download"`
}

func (r *ResourceInfo) Topic() string {
	return fmt.Sprintf("topic/resources/%s/%d", r.Region.Country, r.CPU.Count)
}

func getRegion() (RegionInfo, error) {
	client := ipinfo.NewClient(nil, nil, "")
	info, err := client.GetIPInfo(nil)
	if err != nil {
		return RegionInfo{}, err
	}
	return RegionInfo{
		Region:  info.Region,
		Country: info.Country,
	}, nil
}

func getAppleSiliconGPU() (GpuInfo, error) {
	// Get chip info
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	output, err := cmd.Output()
	if err != nil {
		return GpuInfo{}, fmt.Errorf("failed to get CPU info: %v", err)
	}
	chipInfo := strings.TrimSpace(string(output))

	// Get GPU cores count using ioreg
	cmd = exec.Command("sh", "-c", "ioreg -l | grep gpu-core-count")
	output, err = cmd.Output()
	if err != nil {
		return GpuInfo{}, fmt.Errorf("failed to get GPU core count: %v", err)
	}

	// Extract the number from output like: "gpu-core-count" = 10
	parts := strings.Split(string(output), "=")
	if len(parts) != 2 {
		return GpuInfo{}, fmt.Errorf("unexpected ioreg output format")
	}

	// Clean and parse the number
	coreCount := strings.TrimSpace(parts[1])
	gpuCores, err := strconv.Atoi(coreCount)
	if err != nil {
		return GpuInfo{}, fmt.Errorf("failed to parse GPU core count: %v", err)
	}

	return GpuInfo{
		Count: gpuCores,
		Name:  fmt.Sprintf("Apple Silicon GPU (%s)", chipInfo),
	}, nil
}

func getGpu() (GpuInfo, error) {
	// Check if running on macOS
	if runtime.GOOS == "darwin" {
		// Check for arm64 architecture (Apple Silicon)
		if runtime.GOARCH == "arm64" {
			return getAppleSiliconGPU()
		}
		// For Intel Macs, return integrated graphics
		return GpuInfo{
			Count: 1,
			Name:  "Apple Integrated Graphics",
		}, nil
	}

	// For non-macOS systems, try NVIDIA GPU detection
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return GpuInfo{}, fmt.Errorf("unable to initialize NVML: %v", nvml.ErrorString(ret))
	}
	defer func() {
		if ret := nvml.Shutdown(); ret != nvml.SUCCESS {
			log.Errorf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
		}
	}()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return GpuInfo{}, fmt.Errorf("unable to get device count: %v", nvml.ErrorString(ret))
	}

	gpuNames := make([]string, count)
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			return GpuInfo{}, fmt.Errorf("unable to get device at index %d: %v", i, nvml.ErrorString(ret))
		}
		name, ret := device.GetName()
		if ret != nvml.SUCCESS {
			return GpuInfo{}, fmt.Errorf("unable to get name of device at index %d: %v", i, nvml.ErrorString(ret))
		}
		gpuNames[i] = name
	}

	return GpuInfo{
		Count: count,
		Name:  gpuNames[0], // Simplify to the first GPU for now
	}, nil
}

func getBandwidth() (BandwidthInfo, error) {
	speedtestClient := speedtest.New()
	servers, err := speedtestClient.FetchServers()
	if err != nil {
		return BandwidthInfo{}, err
	}

	targets, err := servers.FindServer([]int{})
	if err != nil {
		return BandwidthInfo{}, err
	}

	// Select the fastest server
	if len(targets) == 0 {
		return BandwidthInfo{}, errors.New("no speedtest servers found")
	}
	server := targets[0]

	err = server.PingTest(nil)
	if err != nil {
		log.Errorf("Ping test failed: %v", err)
		return BandwidthInfo{}, err
	}
	err = server.DownloadTest()
	if err != nil {
		log.Errorf("Download test failed: %v", err)
		return BandwidthInfo{}, err
	}
	err = server.UploadTest()
	if err != nil {
		log.Errorf("Upload test failed: %v", err)
		return BandwidthInfo{}, err
	}

	return BandwidthInfo{
		Latency:       server.Latency,
		UploadSpeed:   uint64(server.ULSpeed),
		DownloadSpeed: uint64(server.DLSpeed),
	}, nil
}

func GetResource() (*ResourceInfo, error) {
	var (
		wg       sync.WaitGroup
		errs     *multierror.Error
		resource ResourceInfo
		mutex    sync.Mutex
	)

	// Run resource fetchers in parallel
	wg.Add(4)

	go func() {
		defer wg.Done()
		region, err := getRegion()
		mutex.Lock()
		defer mutex.Unlock()
		if err != nil {
			errs = multierror.Append(errs, err)
		} else {
			resource.Region = region
		}
	}()

	go func() {
		defer wg.Done()
		gpu, err := getGpu()
		mutex.Lock()
		defer mutex.Unlock()
		if err != nil {
			errs = multierror.Append(errs, err)
		} else {
			resource.GPU = gpu
		}
	}()

	go func() {
		defer wg.Done()
		v, err := mem.VirtualMemory()
		mutex.Lock()
		defer mutex.Unlock()
		if err != nil {
			errs = multierror.Append(err)
		} else {
			resource.Memory = MemoryInfo{
				Total: v.Total,
				Used:  v.Used,
			}
		}
	}()

	go func() {
		defer wg.Done()
		bandwidth, err := getBandwidth()
		mutex.Lock()
		defer mutex.Unlock()
		if err != nil {
			errs = multierror.Append(err)
		} else {
			resource.Bandwidth = bandwidth
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	// Get CPU information
	cpuCounts, err := cpu.Counts(false)
	if err != nil {
		errs = multierror.Append(err)
	} else {
		cpuInfo, err := cpu.Info()
		if err != nil {
			errs = multierror.Append(err)
		} else {
			resource.CPU = CpuInfo{
				Count: cpuCounts,
				Name:  cpuInfo[0].ModelName,
			}
		}
	}

	// Return the result and aggregated errors
	return &resource, errs.ErrorOrNil()
}
