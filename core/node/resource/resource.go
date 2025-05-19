package resource

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/showwin/speedtest-go/speedtest"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("node", "resource")

type ResourceInfo struct {
	Region     RegionInfo    `json:"region"`
	CPU        CpuInfo       `json:"cpu"`       // CPU cores
	GPU        GpuInfo       `json:"gpu"`       // GPU cores
	Memory     MemoryInfo    `json:"memory"`    // RAM (MB)
	Bandwidth  BandwidthInfo `json:"bandwidth"` // Bandwidth (Mbps)
	Storage    StorageInfo   `json:"storage"`   // Storage (MB)
	LastUpdate time.Time     `json:"last_update"`
}

type RegionInfo struct {
	Region  string `json:"region"`
	Country string `json:"country"`
}

type MemoryInfo struct {
	Total uint64 `json:"total"`
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

type StorageInfo struct {
	Total uint64 `json:"total"`
}

func (r *ResourceInfo) Topic() string {
	return fmt.Sprintf("topic/resources/%s/%d", r.Region.Country, r.CPU.Count)
}

func getRegion() (RegionInfo, error) {
	client := ipinfo.NewClient(nil, nil, "")
	info, err := client.GetIPInfo(nil)
	if err != nil {
		return RegionInfo{
			Region:  "unknown",
			Country: "unknown",
		}, err
	}
	return RegionInfo{
		Region:  info.Region,
		Country: info.Country,
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

func getStorage() (StorageInfo, error) {
	usage, err := disk.Usage("/")
	if err != nil {
		return StorageInfo{}, err
	}
	return StorageInfo{
		Total: usage.Total,
	}, nil
}

func GetResource() (*ResourceInfo, error) {
	var (
		wg       sync.WaitGroup
		errs     *multierror.Error
		resource ResourceInfo
		mutex    sync.Mutex
	)

	resource = ResourceInfo{
		Region: RegionInfo{
			Region:  "unknown",
			Country: "unknown",
		},
		CPU: CpuInfo{
			Count: 0,
			Name:  "unknown",
		},
		GPU: GpuInfo{
			Count: 0,
			Name:  "unknown",
		},
		Memory: MemoryInfo{
			Total: 0,
		},
		Bandwidth: BandwidthInfo{
			Latency:       0,
			UploadSpeed:   0,
			DownloadSpeed: 0,
		},
		Storage: StorageInfo{
			Total: 0,
		},
	}

	// Run resource fetchers in parallel
	wg.Add(5)

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

	go func() {
		defer wg.Done()
		storage, err := getStorage()
		mutex.Lock()
		defer mutex.Unlock()
		if err != nil {
			errs = multierror.Append(err)
		} else {
			resource.Storage = storage
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

	resource.LastUpdate = time.Now()

	// Return the result and aggregated errors
	return &resource, errs.ErrorOrNil()
}
