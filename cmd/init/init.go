package init

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/unicornultrafoundation/subnet-node/misc/fsutil"
	"gopkg.in/yaml.v2"
)

func Init(repoPath string, out io.Writer) (map[interface{}]interface{}, error) {
	expPath, err := fsutil.ExpandHome(filepath.Clean(repoPath))
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(expPath, 0755); err != nil {
		return nil, err
	}

	identity, err := CreateIdentity(out)
	if err != nil {
		return nil, err
	}

	// Initialize config with identity
	config, err := InitWithIdentity(identity)
	if err != nil {
		return nil, err
	}

	// Save config to file
	configPath := filepath.Join(expPath, "config.yaml")
	if err := saveConfigToFile(config, configPath); err != nil {
		return nil, err
	}

	fmt.Fprintf(out, "Config saved to %s\n", configPath)
	return config, nil
}

func InitWithIdentity(identity Identity) (map[interface{}]interface{}, error) {
	return map[interface{}]interface{}{
		"identity": map[interface{}]interface{}{
			"peer_id": identity.PeerID,
			"privkey": identity.PrivKey,
		},
	}, nil
}

// SaveConfigToFile saves the configuration map to a YAML file
func saveConfigToFile(config map[interface{}]interface{}, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Encode the map to YAML and write to the file
	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config to file: %w", err)
	}

	return nil
}
