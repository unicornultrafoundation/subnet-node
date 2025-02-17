package apps

import (
	"context"
	"math/big"
	"strings"

	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

type AppFilter struct {
	Status atypes.ProcessStatus `json:"status"`
	Query  string               `json:"query"`
}

// Retrieves a list of apps from the Ethereum contract and checks their container status.
func (s *Service) GetApps(ctx context.Context, start *big.Int, end *big.Int, filter AppFilter) ([]*atypes.App, int, error) {
	appCount, err := s.GetAppCount()
	if err != nil {
		return nil, 0, err
	}
	subnetApps, err := s.accountService.AppStore().ListApps(nil, big.NewInt(1), appCount)
	if err != nil {
		return nil, 0, err
	}

	apps := make([]*atypes.App, 0)
	for i, subnetApp := range subnetApps {
		appId := big.NewInt(1)
		appId.Add(appId, big.NewInt(int64(i)))
		appStatus, err := s.GetContainerStatus(ctx, appId)
		if err != nil {
			return nil, 0, err
		}

		if (filter.Status == "" || appStatus == filter.Status) &&
			(filter.Query == "" || strings.Contains(strings.ToLower(subnetApp.Name), strings.ToLower(filter.Query)) ||
				strings.Contains(strings.ToLower(subnetApp.Symbol), strings.ToLower(filter.Query))) {
			apps = append(apps, atypes.ConvertToApp(subnetApp, appId, appStatus))
		}
	}

	total := len(apps)

	startIndex := int(start.Int64())
	endIndex := int(end.Int64())
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > total {
		endIndex = total
	}
	if startIndex > endIndex {
		return nil, total, nil
	}

	return apps[startIndex:endIndex], total, nil
}
