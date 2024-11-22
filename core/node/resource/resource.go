package resource

import (
	"errors"
	"fmt"
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
	Region    string
	PeerID    string        `json:"peer_id"`
	CPU       CpuInfo       `json:"cpu"`       // CPU cores
	GPU       GpuInfo       `json:"gpu"`       // GPU cores
	Memory    MemoryInfo    `json:"memory"`    // RAM (MB)
	Bandwidth BandwidthInfo `json:"bandwidth"` // Bandwidth (Mbps)
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

func getRegion() (string, error) {
	client := ipinfo.NewClient(nil, nil, "")
	info, err := client.GetIPInfo(nil)
	if err != nil {
		return "Unknown", err
	}
	return info.Region + "-" + info.Country, nil
}

func getGpu() (GpuInfo, error) {
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
