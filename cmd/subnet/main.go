package main

import (
	"flag"
	"fmt"
	"os"

	ninit "github.com/unicornultrafoundation/subnet-node/cmd/init"
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
	repoPath := flag.String("repo", "~/.subnet-node", "Path to either a file or directory to load configuration from")
	initFlag := flag.Bool("init", false, "Init")

	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")

	flag.Parse()

	if *printVersion {
		fmt.Printf("Version: %s\n", Build)
		os.Exit(0)
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	if *repoPath == "" {
		fmt.Println("-repo flag must be set")
		flag.Usage()
		os.Exit(1)
	}

	if *initFlag {
		_, err := ninit.Init(*repoPath, os.Stdout)
		if err != nil {
			fmt.Printf("init err :%v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	subnet.Main(*repoPath, configPath)
}
