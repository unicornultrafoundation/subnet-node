package cmd

import (
	"github.com/spf13/cobra"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/operator"
)

func k8sCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "k8s-services",
		Short:        "K8s services commands",
		Long:         "Commands to run k8s sub-services",
		SilenceUsage: true,
	}

	cmd.AddCommand(operator.OperatorsCmd())
	cmd.AddCommand(operator.ToolsCmd())

	return cmd
}
