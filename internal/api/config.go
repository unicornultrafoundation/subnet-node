package api

import (
	"context"
	"os"
	"path/filepath"
	"strings"

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

func (api *ConfigAPI) Update(ctx context.Context, newConfig map[string]any) error {
	cfg := api.repo.Config()

	// Update only the fields that have changed, including nested keys
	for key, value := range newConfig {
		updateNestedKey(cfg.Settings, key, value)
	}

	configPath := filepath.Join(api.repo.Path(), "config.yaml")
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = yaml.NewEncoder(file).Encode(cfg.Settings)

	if err != nil {
		return err
	}

	cfg.ReloadConfig()
	return nil
}

// Get retrieves the current configuration.
func (api *ConfigAPI) Get(ctx context.Context) (map[string]any, error) {
	cfg := api.repo.Config()
	settings := convertToMapStringInterface(cfg.Settings)
	hideSensitiveKeys(settings)
	return settings, nil
}

func updateNestedKey(settings map[string]any, key string, value any) {
	parts := strings.Split(key, ".")
	last := len(parts) - 1
	for i, part := range parts {
		if i == last {
			settings[part] = value
			return
		}
		if _, ok := settings[part].(map[string]any); !ok {
			settings[part] = make(map[string]any)
		}
		settings = settings[part].(map[string]any)
	}
}

func convertToMapStringInterface(input map[string]any) map[string]any {
	output := make(map[string]any)
	for key, value := range input {
		switch v := value.(type) {
		case map[string]any:
			output[key] = convertToMapStringInterface(v)
		default:
			output[key] = v
		}
	}
	return output
}

func hideSensitiveKeys(settings map[string]any) {
	for key, value := range settings {
		if strings.Contains(strings.ToLower(key), "privkey") ||
			strings.Contains(strings.ToLower(key), "private_key") ||
			strings.Contains(strings.ToLower(key), "auth_secret") {
			settings[key] = "HIDDEN"
		} else if nestedMap, ok := value.(map[string]any); ok {
			hideSensitiveKeys(nestedMap)
		}
	}
}
