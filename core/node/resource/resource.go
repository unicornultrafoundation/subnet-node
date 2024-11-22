package resource

import (
	"fmt"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
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
	GpuInfo   string
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
	// Tạo client
	client := ipinfo.NewClient(nil, nil, "")

	// Lấy thông tin IP
	info, err := client.GetIPInfo(nil)
	if err != nil {
		return "Unknown", err
	}

	return info.Region + "-" + info.Country, nil
}

func getGpu() (GpuInfo, error) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return GpuInfo{}, fmt.Errorf("Unable to initialize NVML: %v", nvml.ErrorString(ret))
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
		}
	}()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return GpuInfo{}, fmt.Errorf("Unable to get device count: %v", nvml.ErrorString(ret))
	}
	name := ""
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			return GpuInfo{}, fmt.Errorf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
		}

		deviceName, ret := device.GetName()
		if ret != nvml.SUCCESS {
			return GpuInfo{}, fmt.Errorf("Unable to get uuid of device at index %d: %v", i, nvml.ErrorString(ret))
		}

		name = deviceName
	}

	return GpuInfo{
		Count: count,
		Name:  name,
	}, nil
}

func getBandwidth() (BandwidthInfo, error) {
	var speedtestClient = speedtest.New()
	serverList, err := speedtestClient.FetchServers()

	if err != nil {
		return BandwidthInfo{}, err
	}

	targets, err := serverList.FindServer([]int{})

	if err != nil {
		return BandwidthInfo{}, err
	}

	if targets.Len() == 0 {
		return BandwidthInfo{}, err
	}

	s := targets[0]

	err = s.PingTest(nil)
	if err != nil {
		log.Error(err)
		return BandwidthInfo{}, err
	}
	err = s.DownloadTest()
	if err != nil {
		log.Error(err)
		return BandwidthInfo{}, err
	}
	err = s.UploadTest()
	if err != nil {
		log.Error(err)
		return BandwidthInfo{}, err
	}
	return BandwidthInfo{
		Latency:       s.Latency,
		UploadSpeed:   uint64(s.ULSpeed),
		DownloadSpeed: uint64(s.DLSpeed),
	}, nil
}

func GetResource() (*ResourceInfo, error) {
	res := &ResourceInfo{}
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	cpuCounts, err := cpu.Counts(false)
	if err != nil {
		return nil, err
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	bw, err := getBandwidth()

	if err != nil {
		return nil, err
	}

	rg, err := getRegion()
	if err != nil {
		return nil, err
	}

	g, err := getGpu()
	if err != nil {
		return nil, err
	}
	res.Memory = MemoryInfo{
		Total: v.Total,
		Used:  v.Used,
	}
	res.CPU = CpuInfo{
		Count: cpuCounts,
		Name:  cpuInfo[0].ModelName,
	}
	res.Bandwidth = bw
	res.Region = rg
	res.GPU = g

	return res, nil
}
