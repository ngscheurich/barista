// Package cli wires the cobra command tree for the barista binary.
//
// The root command owns the program's short description and silence settings;
// each subcommand lives in its own file and is attached here.
package cli

import (
	"github.com/spf13/cobra"
)

// NewRoot builds the root command. RunE errors are surfaced by the caller
// in main, so usage and error printing are silenced here.
func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "barista",
		Short:         "Serves up a new flavor for your terminal apps.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newApplyCmd())
	return cmd
}
