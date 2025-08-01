package cmd

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ninit "github.com/unicornultrafoundation/subnet-node/cmd/init"
	"github.com/unicornultrafoundation/subnet-node/subnet"
)

func RootCmd() *cobra.Command {
	var (
		configPath string
		dataPath   string
		debug      bool
	)

	cmd := &cobra.Command{
		Use:   "subnet-node",
		Short: "Subnet Node",
		Long:  "Subnet Node - U2U Subnet Node Service",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
				logging.SetAllLoggers(logging.LevelDebug)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			subnet.Main(dataPath, &configPath)
		},
	}

	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to config file or directory")
	cmd.PersistentFlags().StringVarP(&dataPath, "datadir", "d", "~/.subnet-node", "Path to data directory")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Enable debug logging")

	cmd.AddCommand(versionCmd())
	cmd.AddCommand(ninit.InitCmd())
	cmd.AddCommand(configCmd())
	cmd.AddCommand(accountCmd())

	return cmd
}
