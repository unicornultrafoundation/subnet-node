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
	var (
		configPath string
		dataPath   string
		debug      bool
	)

	rootCmd := &cobra.Command{
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

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to config file or directory")
	rootCmd.PersistentFlags().StringVarP(&dataPath, "datadir", "d", "~/.subnet-node", "Path to data directory")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Enable debug logging")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", Build)
		},
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize node data directory",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := ninit.Init(dataPath, os.Stdout)
			if err != nil {
				fmt.Printf("init err :%v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(config.ConfigCmd())
	rootCmd.AddCommand(accountcmd.AccountCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
