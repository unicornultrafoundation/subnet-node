package commands

import (
	"github.com/urfave/cli/v2"
)

var RootCmd = &cli.Command{
	Subcommands: []*cli.Command{
		SwarmCmd,
	},
}
