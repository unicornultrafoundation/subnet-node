package api

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

type appResult struct {
	ID                   *hexutil.Big         `json:"id,omitempty"`
	PeerIds              []string             `json:"peer_ids,omitempty"`
	Owner                common.Address       `json:"owner,omitempty"`
	Name                 string               `json:"name,omitempty"`
	Symbol               string               `json:"symbol,omitempty"`
	Description          string               `json:"description,omitempty"`
	Logo                 string               `json:"logo,omitempty"`
	BannerUrls           []string             `json:"banners_urls,omitempty"`
	DefaultBannerIndex   int64                `json:"default_banner_index,omitempty"`
	Budget               *hexutil.Big         `json:"budget,omitempty"`
	SpentBudget          *hexutil.Big         `json:"spent_budget,omitempty"`
	MaxNodes             *hexutil.Big         `json:"max_nodes,omitempty"`
	MinCpu               *hexutil.Big         `json:"min_cpu,omitempty"`
	MinGpu               *hexutil.Big         `json:"min_gpu,omitempty"`
	MinMemory            *hexutil.Big         `json:"min_memory,omitempty"`
	MinUploadBandwidth   *hexutil.Big         `json:"min_upload_bandwidth,omitempty"`
	MinDownloadBandwidth *hexutil.Big         `json:"min_download_bandwidth,omitempty"`
	NodeCount            *hexutil.Big         `json:"node_count,omitempty"`
	PricePerCpu          *hexutil.Big         `json:"price_per_cpu,omitempty"`
	PricePerGpu          *hexutil.Big         `json:"price_per_gpu,omitempty"`
	PricePerMemoryGB     *hexutil.Big         `json:"price_per_memory_gb,omitempty"`
	PricePerStorageGB    *hexutil.Big         `json:"price_per_storage_gb,omitempty"`
	PricePerBandwidthGB  *hexutil.Big         `json:"price_per_bandwidth_gb,omitempty"`
	Status               atypes.ProcessStatus `json:"status,omitempty"`
	Metadata             *atypes.AppMetadata  `json:"metadata,omitempty"`
	Usage                *resourceUsageResult `json:"usage,omitempty"`
	IP                   string               `json:"ip"`
}

type resourceUsageResult struct {
	UsedCpu           *hexutil.Big `json:"usedCpu"`
	UsedGpu           *hexutil.Big `json:"usedGpu"`
	UsedMemory        *hexutil.Big `json:"usedMemory"`
	UsedStorage       *hexutil.Big `json:"usedStorage"`
	UsedUploadBytes   *hexutil.Big `json:"usedUploadBytes"`
	UsedDownloadBytes *hexutil.Big `json:"usedDownloadBytes"`
	Duration          *hexutil.Big `json:"duration"`
}

func convertToUsageResult(usage *atypes.ResourceUsage) *resourceUsageResult {
	if usage == nil {
		return nil
	}
	return &resourceUsageResult{
		UsedCpu:           (*hexutil.Big)(usage.UsedCpu),
		UsedGpu:           (*hexutil.Big)(usage.UsedGpu),
		UsedMemory:        (*hexutil.Big)(usage.UsedMemory),
		UsedStorage:       (*hexutil.Big)(usage.UsedStorage),
		UsedUploadBytes:   (*hexutil.Big)(usage.UsedUploadBytes),
		UsedDownloadBytes: (*hexutil.Big)(usage.UsedDownloadBytes),
		Duration:          (*hexutil.Big)(usage.Duration),
	}
}

func convertToAppResult(app *atypes.App) *appResult {
	return &appResult{
		ID:                   (*hexutil.Big)(app.ID),
		Name:                 app.Name,
		PeerIds:              app.PeerIds,
		Owner:                app.Owner,
		Symbol:               app.Symbol,
		Description:          app.Description,
		Logo:                 app.Logo,
		BannerUrls:           app.BannerUrls,
		DefaultBannerIndex:   app.DefaultBannerIndex,
		Budget:               (*hexutil.Big)(app.Budget),
		SpentBudget:          (*hexutil.Big)(app.SpentBudget),
		MaxNodes:             (*hexutil.Big)(app.MaxNodes),
		MinCpu:               (*hexutil.Big)(app.MinCpu),
		MinGpu:               (*hexutil.Big)(app.MinGpu),
		MinMemory:            (*hexutil.Big)(app.MinMemory),
		MinUploadBandwidth:   (*hexutil.Big)(app.MinUploadBandwidth),
		MinDownloadBandwidth: (*hexutil.Big)(app.MinDownloadBandwidth),
		NodeCount:            (*hexutil.Big)(app.NodeCount),
		PricePerCpu:          (*hexutil.Big)(app.PricePerCpu),
		PricePerMemoryGB:     (*hexutil.Big)(app.PricePerMemoryGB),
		PricePerGpu:          (*hexutil.Big)(app.PricePerGpu),
		PricePerStorageGB:    (*hexutil.Big)(app.PricePerStorageGB),
		PricePerBandwidthGB:  (*hexutil.Big)(app.PricePerBandwidthGB),
		Status:               app.Status,
		Metadata:             app.Metadata,
		Usage:                convertToUsageResult(app.Usage),
		IP:                   app.IP,
	}
}

type AppAPI struct {
	appService *apps.Service
}

// NewAppAPI creates a new instance of AppAPI.
func NewAppAPI(appService *apps.Service) *AppAPI {
	return &AppAPI{appService: appService}
}

func (api *AppAPI) GetAppCount(ctx context.Context, appFilter *apps.AppFilter) (*hexutil.Big, error) {
	var appCount *big.Int
	if appFilter == nil {
		total, err := api.appService.GetAppCount()
		if err != nil {
			return nil, err
		}
		appCount = total
	} else {
		_, total, err := api.appService.GetApps(ctx, big.NewInt(0), big.NewInt(0), *appFilter)
		if err != nil {
			return nil, err
		}
		appCount = big.NewInt(int64(total))
	}

	return (*hexutil.Big)(appCount), nil
}

func (api *AppAPI) GetApps(ctx context.Context, start int64, end int64, appFilter *apps.AppFilter) ([]appResult, error) {
	if appFilter == nil {
		appFilter = &apps.AppFilter{
			Status: "",
			Query:  "",
		}
	}

	apps, _, err := api.appService.GetApps(ctx, big.NewInt(start), big.NewInt(end), *appFilter)
	if err != nil {
		return nil, err
	}

	result := make([]appResult, len(apps))

	for i, app := range apps {
		result[i] = *convertToAppResult(app)
	}
	return result, nil
}

func (api *AppAPI) GetApp(ctx context.Context, appId hexutil.Big) (*appResult, error) {
	subnetApp, err := api.appService.GetApp(ctx, appId.ToInt())
	if err != nil {
		return nil, err
	}
	return convertToAppResult(subnetApp), nil
}

func (api *AppAPI) RunApp(ctx context.Context, appId hexutil.Big) (*appResult, error) {
	subnetApp, err := api.appService.RunApp(ctx, appId.ToInt())
	if err != nil {
		return nil, err
	}

	return convertToAppResult(subnetApp), nil
}

func (api *AppAPI) RemoveApp(ctx context.Context, appId hexutil.Big) (*appResult, error) {
	subnetApp, err := api.appService.RemoveApp(ctx, appId.ToInt())
	if err != nil {
		return nil, err
	}
	return convertToAppResult(subnetApp), nil
}

func (api *AppAPI) GetUsage(ctx context.Context, appId hexutil.Big) (*resourceUsageResult, error) {
	usage, err := api.appService.GetUsage(ctx, appId.ToInt())
	if err != nil {
		return &resourceUsageResult{
			UsedCpu:           &hexutil.Big{},
			UsedGpu:           &hexutil.Big{},
			UsedMemory:        &hexutil.Big{},
			UsedStorage:       &hexutil.Big{},
			UsedUploadBytes:   &hexutil.Big{},
			UsedDownloadBytes: &hexutil.Big{},
			Duration:          &hexutil.Big{},
		}, nil
	}
	return convertToUsageResult(usage), nil
}

func (api *AppAPI) GetAllUsage(ctx context.Context) (*resourceUsageResult, error) {
	usage, err := api.appService.GetAllRunningContainersUsage(ctx)
	if err != nil {
		return nil, err
	}
	return convertToUsageResult(usage), nil
}

func (api *AppAPI) UpdateAppConfig(ctx context.Context, appId hexutil.Big, cfg atypes.ContainerConfig) error {
	return api.appService.UpdateAppConfig(ctx, appId.ToInt(), cfg)
}
