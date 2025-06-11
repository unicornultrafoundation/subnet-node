package init

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"gopkg.in/yaml.v2"
)

// ContainerACLConfig represents the structure of the container allow IPs configuration
type ContainerACLConfig struct {
	Global     GlobalACLConfig                    `json:"global"`
	Containers map[string]ContainerSpecificConfig `json:"containers"`
}

// GlobalACLConfig represents global allow IP settings
type GlobalACLConfig struct {
	AllowIPs []string `json:"allow_ips"`
	Enabled  bool     `json:"enabled"`
}

// ContainerSpecificConfig represents container-specific allow IP settings
type ContainerSpecificConfig struct {
	AllowIPs []string `json:"allow_ips"`
	Enabled  bool     `json:"enabled"`
}

func Init(repoPath string, out io.Writer) (map[interface{}]interface{}, error) {
	return InitWithOptions(repoPath, out, InitOptions{})
}

// InitOptions provides options for initialization
type InitOptions struct {
	CustomAllowIPs []string // Custom allow IPs to use instead of defaults
	DisableACL     bool     // Disable ACL checking by default
}

func InitWithOptions(repoPath string, out io.Writer, options InitOptions) (map[interface{}]interface{}, error) {
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

	accountPrivateKey, err := createPrivateKey()
	if err != nil {
		return nil, err
	}

	config["account"] = map[interface{}]interface{}{
		"private_key": accountPrivateKey,
	}

	// Create container ACL configuration
	containerACLPath := filepath.Join(expPath, "container_allow_ips.json")
	if err := createContainerACLConfigWithOptions(containerACLPath, out, options); err != nil {
		return nil, fmt.Errorf("failed to create container ACL config: %w", err)
	}

	// Save config to file
	configPath := filepath.Join(expPath, "config.yaml")
	if err := saveConfigToFile(config, configPath); err != nil {
		return nil, err
	}

	fmt.Fprintf(out, "Config saved to %s\n", configPath)
	return config, nil
}

// InitWithCustomAllowIPs initializes with custom allow IPs
func InitWithCustomAllowIPs(repoPath string, out io.Writer, allowIPs []string) (map[interface{}]interface{}, error) {
	return InitWithOptions(repoPath, out, InitOptions{
		CustomAllowIPs: allowIPs,
		DisableACL:     false,
	})
}

func InitWithIdentity(identity Identity) (map[interface{}]interface{}, error) {
	return map[interface{}]interface{}{
		"identity": map[interface{}]interface{}{
			"peer_id": identity.PeerID,
			"privkey": identity.PrivKey,
		},
		"addresses": map[interface{}]interface{}{
			"swarm": []interface{}{
				"/ip4/0.0.0.0/tcp/4001",
				"/ip6/::/tcp/4001",
				"/ip4/0.0.0.0/udp/4001/webrtc-direct",
				"/ip4/0.0.0.0/udp/4001/quic-v1",
				"/ip4/0.0.0.0/udp/4001/quic-v1/webtransport",
				"/ip6/::/udp/4001/webrtc-direct",
				"/ip6/::/udp/4001/quic-v1",
				"/ip6/::/udp/4001/quic-v1/webtransport",
			},
			"api": []interface{}{
				"/ip4/127.0.0.1/tcp/8080",
			},
		},
		"bootstrap": []interface{}{
			"/ip4/47.129.250.9/tcp/4001/p2p/12D3KooWDK63y6sxFi3dNqrS8yRetgbB81Tzszvs2yLoEtWtPCDa",
			"/ip4/54.255.247.224/tcp/4001/p2p/12D3KooWQT4FPJiYSSRNk8qaWhHVSxAFmKzuCwDAJskZpxUnhcd2",
		},
	}, nil
}

// createContainerACLConfigWithOptions creates the container ACL configuration file with custom options
func createContainerACLConfigWithOptions(configPath string, out io.Writer, options InitOptions) error {
	// Set default allow IPs if none provided
	allowIPs := options.CustomAllowIPs
	if len(allowIPs) == 0 {
		allowIPs = []string{
			"192.168.0.0/16", // Private networks
			"10.0.0.0/8",     // Private networks
			"172.16.0.0/12",  // Private networks
			"127.0.0.1",      // Localhost
		}
	}

	// Create simple container ACL configuration
	config := ContainerACLConfig{
		Global: GlobalACLConfig{
			AllowIPs: allowIPs,
			Enabled:  !options.DisableACL,
		},
		Containers: map[string]ContainerSpecificConfig{},
	}

	// Create the file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create container ACL config file: %w", err)
	}
	defer file.Close()

	// Write JSON with pretty formatting
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write container ACL config: %w", err)
	}

	fmt.Fprintf(out, "Container ACL configuration:\n")
	fmt.Fprintf(out, "  - Global allow IPs: %v\n", config.Global.AllowIPs)
	fmt.Fprintf(out, "  - ACL checking: %s\n", map[bool]string{true: "enabled", false: "disabled"}[config.Global.Enabled])

	if options.DisableACL {
		fmt.Fprintf(out, "  - WARNING: ACL checking is DISABLED - all connections will go through cluster membership check\n")
	}

	if len(options.CustomAllowIPs) > 0 {
		fmt.Fprintf(out, "  - Using custom allow IPs: %v\n", options.CustomAllowIPs)
	}

	fmt.Fprintf(out, "  - Edit %s to customize container access rules\n", configPath)

	return nil
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

// createPrivateKey generates a new private key and returns its hex representation
func createPrivateKey() (string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", err
	}

	// Convert the private key to bytes
	privateKeyBytes := crypto.FromECDSA(privateKey)
	// Encode the bytes to a hexadecimal string
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	return privateKeyHex, nil
}
