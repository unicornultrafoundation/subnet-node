package apps

import (
	"context"
	"math/big"
	"strings"

	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

const DefaultGithubAppsURL = "https://raw.githubusercontent.com/unicornultrafoundation/subnet-apps/refs/heads/main/index.json"

type AppFilter struct {
	Status atypes.ProcessStatus `json:"status"`
	Query  string               `json:"query"`
}

type GitHubApp struct {
	ID                 int64                  `json:"id"`
	Symbol             string                 `json:"symbol"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Logo               string                 `json:"logo"`
	BannerUrls         []string               `json:"banners_urls"`
	DefaultBannerIndex int64                  `json:"default_banner_index"`
	Website            string                 `json:"website"`
	Container          atypes.ContainerConfig `json:"container"`
}

// Checks if the app matches the filter criteria.
func appMatchesFilter(app GitHubApp, status atypes.ProcessStatus, filter AppFilter) bool {
	return (filter.Status == "" || status == filter.Status) &&
		(filter.Query == "" || strings.Contains(strings.ToLower(app.Name), strings.ToLower(filter.Query)) ||
			strings.Contains(strings.ToLower(app.Symbol), strings.ToLower(filter.Query)))
}

// Retrieves a list of apps from the Ethereum contract and checks their container status.
func (s *Service) GetApps(ctx context.Context, start *big.Int, end *big.Int, filter AppFilter) ([]*atypes.App, int, error) {
	// Fetch JSON file from GitHub
	githubAppsURL := s.cfg.GetString("apps.github_apps", DefaultGithubAppsURL)
	gitHubApps, err := s.getGitHubApps(githubAppsURL)
	if err != nil {
		return nil, 0, err
	}
	apps := make([]*atypes.App, 0)
	for _, gitHubApp := range gitHubApps {
		appId := big.NewInt(gitHubApp.ID)
		appStatus, err := s.GetContainerStatus(ctx, appId)
		if err != nil {
			return nil, 0, err
		}

		if appMatchesFilter(gitHubApp, appStatus, filter) {
			app, err := s.GetApp(ctx, appId)
			if err != nil {
				log.Errorf("Failed to get app %s: %v", appId, err)
				continue
			}

			apps = append(apps, app)
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
