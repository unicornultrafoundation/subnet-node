package init

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/spf13/cobra"
	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
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

	// Generate a new private key and keystore
	ksDir := filepath.Join(expPath, "keystore")
	if err := os.MkdirAll(ksDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create keystore dir: %w", err)
	}
	ks := keystore.NewKeyStore(ksDir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, _, _ := CreateKeystoreAccount(ks, "", "Enter keystore password: ")
	fmt.Fprintf(out, "Keystore created at %s\n", account.URL.Path)

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

// PromptPassword prompts the user for a password with the given message.
func PromptPassword(message string) (string, error) {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	pw, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return strings.TrimSpace(pw), nil
}

// CreateKeystoreAccount prompts for password if needed and creates a new keystore account
func CreateKeystoreAccount(ks *keystore.KeyStore, password string, promptMsg string) (accounts.Account, string, error) {
	var err error
	if password == "" {
		password, err = PromptPassword(promptMsg)
		if err != nil {
			return accounts.Account{}, "", err
		}
	}
	acc, err := ks.NewAccount(password)
	if err != nil {
		return accounts.Account{}, "", fmt.Errorf("failed to create new keystore account: %w", err)
	}
	return acc, password, nil
}

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the subnet node data directory",
		Run: func(cmd *cobra.Command, args []string) {
			dataPath, _ := cmd.Flags().GetString("datadir")
			out := cmd.OutOrStdout()
			if _, err := Init(dataPath, out); err != nil {
				fmt.Fprintf(out, "Error initializing: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringP("datadir", "d", "~/.subnet-node", "Path to data directory")
	return cmd
}
