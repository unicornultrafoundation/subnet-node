package main

import (
	"fmt"
	"os"

	logging "github.com/ipfs/go-log/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ninit "github.com/unicornultrafoundation/subnet-node/cmd/init"
	accountcmd "github.com/unicornultrafoundation/subnet-node/cmd/subnet/account"
	"github.com/unicornultrafoundation/subnet-node/cmd/subnet/config"
	"github.com/unicornultrafoundation/subnet-node/subnet"
)

// A version string that can be set with
//
//	-ldflags "-X main.Build=SOMEVERSION"
//
// at compile-time.
var Build string

func main() {
	cmd := rootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
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
	cmd.AddCommand(config.ConfigCmd())
	cmd.AddCommand(accountcmd.AccountCmd())

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of the subnet node",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Subnet Node Version: %s\n", Build)
		},
	}
}
