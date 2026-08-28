package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/paths"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/recipe/ghostty"
	"github.com/ngscheurich/barista/internal/recipe/neovim"
	"github.com/ngscheurich/barista/internal/recipe/zellij"
)

func newApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply <theme>",
		Short: "Apply a flavor to all configured applications",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(cmd, args[0])
		},
	}
}

// runApply is the apply command's body, split out so a test can call it
// with a controlled command (and captured stdout/stderr) without going
// through the cobra tree. The returned error is surfaced to the caller
// (cobra, then main), which is responsible for printing it to stderr;
// the root command sets SilenceErrors so cobra does not print it itself.
func runApply(cmd *cobra.Command, theme string) error {
	return apply(cmd.OutOrStdout(), theme)
}

// apply runs the full theming pipeline for one theme: ensure the data
// dir, load the flavor, run every recipe collecting errors, and on
// success print the served-up line with the flavor's Name. Pre-recipe
// failures (missing data dir, missing flavor) short-circuit and return;
// per-recipe failures are aggregated via errors.Join so one app's
// failure does not hide another's. The caller is responsible for
// printing the returned error to stderr.
func apply(out io.Writer, theme string) error {
	dataDir, err := paths.EnsureDataDir()
	if err != nil {
		return fmt.Errorf("apply %s: %w", theme, err)
	}

	configDir, err := paths.ConfigDir()
	if err != nil {
		return fmt.Errorf("apply %s: %w", theme, err)
	}

	flavorsDir, err := paths.FlavorsDir()
	if err != nil {
		return fmt.Errorf("apply %s: %w", theme, err)
	}

	f, err := flavor.Load(flavorsDir, theme)
	if err != nil {
		return fmt.Errorf("apply %s: %w", theme, err)
	}

	recipes := []recipe.Recipe{
		ghostty.New(flavorsDir, dataDir),
		neovim.New(flavorsDir, dataDir),
		zellij.New(flavorsDir, configDir),
	}

	var errs []error
	for _, r := range recipes {
		if err := r.Run(f); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	fmt.Fprintf(out, "☕︎ Served up %s\n", f.Name)
	return nil
}
