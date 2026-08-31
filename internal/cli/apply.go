package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/ngscheurich/barista/internal/config"
	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/paths"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/recipe/fzf"
	"github.com/ngscheurich/barista/internal/recipe/ghostty"
	"github.com/ngscheurich/barista/internal/recipe/neovim"
	"github.com/ngscheurich/barista/internal/recipe/zellij"
)

func newApplyCmd() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "apply <theme>",
		Short: "Apply a flavor to all configured applications",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(cmd, args[0], verbose)
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"log each step the recipes perform, including files read and written")
	return cmd
}

// runApply is the apply command's body, split out so a test can call it
// with a controlled command (and captured stdout/stderr) without going
// through the cobra tree. The returned error is surfaced to the caller
// (cobra, then main), which is responsible for printing it to stderr;
// the root command sets SilenceErrors so cobra does not print it itself.
func runApply(cmd *cobra.Command, theme string, verbose bool) error {
	level := log.WarnLevel
	if verbose {
		level = log.InfoLevel
	}
	log.SetOutput(cmd.ErrOrStderr())
	log.SetLevel(level)
	return apply(cmd.OutOrStdout(), theme)
}

// apply runs the full theming pipeline for one theme: ensure the data
// dir, load the flavor, run every recipe, and print the served-up list
// marking each app ✓ (applied) or • (skipped, because the flavor carries
// no template for it). Pre-recipe failures (missing data dir, missing
// flavor) short-circuit and return. A recipe that returns
// recipe.ErrNotApplicable is a skip, not a failure; every other recipe
// error is aggregated via errors.Join so one app's failure does not hide
// another's. The served-up list always names every configured app; the
// caller is responsible for printing the returned error to stderr.
//
// Per-step logs (template reads, theme writes, reload actions) are emitted
// by each recipe via the package-level charmbracelet/log default logger,
// whose level and output are configured by runApply from the --verbose
// flag.
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

	cfg, err := config.Load(configDir)
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
		{name: "fzf", r: fzf.New(flavorsDir)},
		{name: "Ghostty", r: ghostty.New(flavorsDir, dataDir)},
		{name: "Neovim", r: neovim.New(flavorsDir, dataDir)},
		{name: "Zellij", r: zellij.New(flavorsDir, configDir)},
	}

	// Bind a renderer to out; lipgloss detects the color profile from
	// the writer, so styling degrades to plain text when out is not a
	// color-capable terminal (a test buffer, a pipe, NO_COLOR, ...).
	renderer := lipgloss.NewRenderer(out)

	// lipgloss.Color takes an ANSI index ("1"–"255") or hex ("#rrggbb"),
	// not a color name. The status symbol carries the state's color and
	// the app name stays plain, mirroring how gum and gh-dash render
	// state glyphs.
	appliedStyle := renderer.NewStyle().Foreground(lipgloss.Color("2"))
	skippedStyle := renderer.NewStyle().Faint(true)
	boldStyle := renderer.NewStyle().Bold(true)

	fmt.Fprintf(out, "%s\n\n", boldStyle.Render(fmt.Sprintf("%s Served up %s to:", cfg.Icon, f.Name)))
	var errs []error
	for _, a := range apps {
		err := a.r.Run(f)
		if err == nil {
			fmt.Fprintf(out, "  %s %s\n", appliedStyle.Render("✓"), a.name)
			continue
		}
		fmt.Fprintf(out, "  %s %s\n", skippedStyle.Render("•"), skippedStyle.Render(a.name))
		if !errors.Is(err, recipe.ErrNotApplicable) {
			errs = append(errs, err)
		}
	}
	fmt.Fprintln(out)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
