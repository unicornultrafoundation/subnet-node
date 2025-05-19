package main

import (
	"flag"
	"fmt"
	"os"

	logging "github.com/ipfs/go-log/v2"

	"github.com/sirupsen/logrus"
	ninit "github.com/unicornultrafoundation/subnet-node/cmd/init"
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
	configPath := flag.String("config", "", "Path to either a file or directory to load configuration from")
	dataPath := flag.String("datadir", "~/.subnet-node", "Path to either a file or directory to load configuration from")
	initFlag := flag.Bool("init", false, "Init")
	debugFlag := flag.Bool("debug", false, "Debug")

	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")

	flag.Parse()

	if *debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
		logging.SetAllLoggers(logging.LevelDebug)
	}

	if *printVersion {
		fmt.Printf("Version: %s\n", Build)
		os.Exit(0)
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	if *dataPath == "" {
		fmt.Println("-datadir flag must be set")
		flag.Usage()
		os.Exit(1)
	}

	// Check for subcommand
	if len(flag.Args()) > 0 {
		subcommand := flag.Args()[0]

		if subcommand == "edit-config" {
			if err := config.EditConfig(*dataPath, flag.Args()[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	if *initFlag {
		_, err := ninit.Init(*dataPath, os.Stdout)
		if err != nil {
			fmt.Printf("init err :%v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	subnet.Main(*dataPath, configPath)
}
