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
// dir, load the flavor, run every recipe, and print the served-up list
// marking each app ☑ (applied) or ☐ (skipped, because the flavor carries
// no template for it). Pre-recipe failures (missing data dir, missing
// flavor) short-circuit and return. A recipe that returns
// recipe.ErrNotApplicable is a skip, not a failure; every other recipe
// error is aggregated via errors.Join so one app's failure does not hide
// another's. The served-up list always names every configured app; the
// caller is responsible for printing the returned error to stderr.
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

	// apps pairs each recipe with the display name printed in the
	// served-up list. Order is alphabetical by name, which governs the
	// row order in the output.
	apps := []struct {
		name string
		r    recipe.Recipe
	}{
		{name: "Ghostty", r: ghostty.New(flavorsDir, dataDir)},
		{name: "Neovim", r: neovim.New(flavorsDir, dataDir)},
		{name: "Zellij", r: zellij.New(flavorsDir, configDir)},
	}

	fmt.Fprintf(out, "☕ Served up %s to:\n\n", f.Name)
	var errs []error
	for _, a := range apps {
		err := a.r.Run(f)
		switch {
		case err == nil:
			fmt.Fprintf(out, "  ☑ %s\n", a.name)
		case errors.Is(err, recipe.ErrNotApplicable):
			fmt.Fprintf(out, "  ☐ %s\n", a.name)
		default:
			fmt.Fprintf(out, "  ☐ %s\n", a.name)
			errs = append(errs, err)
		}
	}
	fmt.Fprintln(out)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
