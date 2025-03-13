package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	logging "github.com/ipfs/go-log/v2"

	"github.com/sirupsen/logrus"
	ninit "github.com/unicornultrafoundation/subnet-node/cmd/init"
	"github.com/unicornultrafoundation/subnet-node/connect"
	"github.com/unicornultrafoundation/subnet-node/subnet"
)

// A version string that can be set with
//
//	-ldflags "-X main.Build=SOMEVERSION"
//
// at compile-time.
var Build string

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "[" + strings.Join(*i, ", ") + "]"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	connectCmd := flag.NewFlagSet("connect", flag.ExitOnError)
	connectPeer := connectCmd.String("peer", "", "The peer address connect to")
	connectApp := connectCmd.String("appId", "", "The ID of application want to connect to")
	var ports arrayFlags
	connectCmd.Var(&ports, "port", "Proxy mapping ports with format: {local port}:{app port}")

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "connect":
			connectCmd.Parse(os.Args[2:])
			connect.Main(*connectPeer, *connectApp, ports)
			os.Exit(0)
		}
	}

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
