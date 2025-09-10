package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// A version string that can be set with
//
//	-ldflags "-X main.Build=SOMEVERSION"
//
// at compile-time.
var Build string

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of the subnet node",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Subnet Node Version: %s\n", Build)
		},
	}
}
