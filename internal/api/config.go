package api

import (
	"context"
	"os"
	"path/filepath"

	"github.com/unicornultrafoundation/subnet-node/repo"
	"gopkg.in/yaml.v2"
)

type ConfigAPI struct {
	repo repo.Repo
}

// NewConfigAPI creates a new instance of ConfigAPI.
func NewConfigAPI(repo repo.Repo) *ConfigAPI {
	return &ConfigAPI{repo: repo}
}

func (api *ConfigAPI) Update(ctx context.Context, newConfig map[string]interface{}) error {
	cfg := api.repo.Config()

	// Update only the fields that have changed
	for key, value := range newConfig {
		cfg.Settings[key] = value
	}
	configPath := filepath.Join(api.repo.Path(), "config.yaml")
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return yaml.NewEncoder(file).Encode(cfg.Settings)
}
