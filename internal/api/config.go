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

func (api *ConfigAPI) Update(ctx context.Context, newConfig map[string]interface{}) error {
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

	return yaml.NewEncoder(file).Encode(cfg.Settings)
}

// Get retrieves the current configuration.
func (api *ConfigAPI) Get(ctx context.Context) (map[string]interface{}, error) {
	cfg := api.repo.Config()
	settings := convertToMapStringInterface(cfg.Settings)
	hideSensitiveKeys(settings)
	return settings, nil
}

func updateNestedKey(settings map[interface{}]interface{}, key string, value interface{}) {
	parts := strings.Split(key, ".")
	last := len(parts) - 1
	for i, part := range parts {
		if i == last {
			settings[part] = value
			return
		}
		if _, ok := settings[part].(map[interface{}]interface{}); !ok {
			settings[part] = make(map[interface{}]interface{})
		}
		settings = settings[part].(map[interface{}]interface{})
	}
}

func convertToMapStringInterface(input map[interface{}]interface{}) map[string]interface{} {
	output := make(map[string]interface{})
	for key, value := range input {
		strKey := key.(string)
		switch v := value.(type) {
		case map[interface{}]interface{}:
			output[strKey] = convertToMapStringInterface(v)
		default:
			output[strKey] = v
		}
	}
	return output
}

func hideSensitiveKeys(settings map[string]interface{}) {
	for key, value := range settings {
		if strings.Contains(strings.ToLower(key), "privkey") || strings.Contains(strings.ToLower(key), "private_key") {
			settings[key] = "HIDDEN"
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			hideSensitiveKeys(nestedMap)
		}
	}
}
