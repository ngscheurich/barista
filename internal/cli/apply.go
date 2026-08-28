package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply <theme>",
		Short: "Apply a flavor to all configured applications",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", theme)
			return nil
		},
	}
}
