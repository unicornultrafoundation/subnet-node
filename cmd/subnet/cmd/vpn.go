package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	ninit "github.com/unicornultrafoundation/subnet-node/cmd/init"
	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"github.com/unicornultrafoundation/subnet-node/internal/api"
	"github.com/unicornultrafoundation/subnet-node/repo/snrepo"
	"github.com/unicornultrafoundation/subnet-node/subnet"
)

func vpnCmd() *cobra.Command {
	var (
		ipFlag   string
		dataPath string
	)

	cmd := &cobra.Command{
		Use:          "vpn",
		Short:        "Enable VPN and optionally set virtual IP",
		Long:         "Enable the VPN service in the config and optionally set vpn.virtual_ip via --ip without requiring it.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !snrepo.IsInitialized(dataPath) {
				fmt.Printf("Initializing Subnet Node at %s...\n", dataPath)
				if _, err := ninit.Init(dataPath, os.Stdout); err != nil {
					return fmt.Errorf("failed to initialize repo: %v", err)
				}
			}

			expPath, err := fsutil.ExpandHome(filepath.Clean(dataPath))
			if err != nil {
				return fmt.Errorf("failed to expand data path: %v", err)
			}

			configFile := filepath.Join(expPath, "config.yaml")
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				return fmt.Errorf("configuration file does not exist at %s", configFile)
			}

			r, err := snrepo.Open(dataPath, &configFile)
			if err != nil {
				return fmt.Errorf("failed to open repo: %v", err)
			}
			defer r.Close()

			cfgAPI := api.NewConfigAPI(r)
			updates := map[string]interface{}{
				"vpn.enable":                         true,
				"internal.libp2p_force_reachability": "private",
			}
			if ipFlag != "" {
				updates["vpn.virtual_ip"] = ipFlag
			}

			if err := cfgAPI.Update(context.Background(), updates); err != nil {
				return fmt.Errorf("failed to update config: %v", err)
			}

			fmt.Println("VPN enabled in configuration")
			if ipFlag != "" {
				fmt.Printf("Set vpn.virtual_ip to %s\n", ipFlag)
			}

			// Start node using the explicit config file we just updated
			configPath := configFile
			subnet.Main(dataPath, &configPath, "")
			return nil
		},
	}

	cmd.Flags().StringVar(&dataPath, "datadir", "~/.subnet-node", "Path to data directory")
	cmd.Flags().StringVar(&ipFlag, "ip", "", "Optional virtual IP to set for VPN (e.g., 100.68.1.10)")
	return cmd
}
