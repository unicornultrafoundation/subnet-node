package cmd

import (
	"github.com/spf13/cobra"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "k8s-services",
		Short:        "K8s services commands",
		SilenceUsage: true,
	}

	cmd.AddCommand(operator.OperatorsCmd())
	cmd.AddCommand(operator.ToolsCmd())

	return cmd
}
