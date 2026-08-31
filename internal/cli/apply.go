package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/ngscheurich/barista/internal/config"
	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/paths"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/recipe/fzf"
	"github.com/ngscheurich/barista/internal/recipe/ghostty"
	"github.com/ngscheurich/barista/internal/recipe/neovim"
	"github.com/ngscheurich/barista/internal/recipe/zellij"
	"github.com/ngscheurich/barista/internal/ui"
)

// ErrAborted reports that the user cancelled the picker: nothing was
// applied. The caller exits non-zero without printing — the user changed
// their mind, and a silent exit is gum's own contract for a cancelled
// choose.
var ErrAborted = huh.ErrUserAborted

func newApplyCmd() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "apply [flavor]",
		Short: "Apply a flavor to all configured applications",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dirname := ""
			if len(args) > 0 {
				dirname = args[0]
			} else {
				var err error
				dirname, err = resolveFlavor(cmd)
				if err != nil {
					return err
				}
			}
			return runApply(cmd, dirname, verbose)
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"log each step the recipes perform, including files read and written")
	return cmd
}

// resolveFlavor decides which flavor apply serves when no argument was
// given: the picker's choice. The picker opens only when stdin is a
// terminal or accessible mode is requested via BARISTA_ACCESSIBLE;
// otherwise the error lists the available flavors' dirnames in plain
// text — the surface scripts and screen readers consume. An empty
// flavors directory never reaches the picker.
func resolveFlavor(cmd *cobra.Command) (string, error) {
	flavorsDir, err := paths.FlavorsDir()
	if err != nil {
		return "", fmt.Errorf("apply: %w", err)
	}
	flavors, err := flavor.List(flavorsDir)
	if err != nil {
		// A flavors directory that does not exist yet is the same
		// situation as an empty one: nothing to pick from.
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("no flavors found in %s", flavorsDir)
		}
		return "", fmt.Errorf("apply: %w", err)
	}
	if len(flavors) == 0 {
		return "", fmt.Errorf("no flavors found in %s", flavorsDir)
	}

	accessible := os.Getenv("BARISTA_ACCESSIBLE") != ""
	in := cmd.InOrStdin()
	if !accessible && !isTerminal(in) {
		names := make([]string, len(flavors))
		for i, f := range flavors {
			names[i] = f.Dirname
		}
		return "", fmt.Errorf("no flavor given; available: %s", strings.Join(names, ", "))
	}

	choice, err := ui.Pick(cmd.OutOrStdout(), in, pickerOptions(flavors), accessible)
	if err != nil {
		return "", fmt.Errorf("apply: %w", err)
	}
	return choice, nil
}

// pickerOptions maps flavors to picker rows: each flavor's Name, with
// its Dirname as the label when two flavors share a Name, so no two
// rows are indistinguishable. The chosen row's Value is the Dirname
// apply takes as its argument.
func pickerOptions(flavors []flavor.Flavor) []ui.Option {
	counts := make(map[string]int)
	for _, f := range flavors {
		counts[f.Name]++
	}
	opts := make([]ui.Option, len(flavors))
	for i, f := range flavors {
		label := f.Name
		if counts[f.Name] > 1 {
			label = f.Dirname
		}
		opts[i] = ui.Option{Label: label, Value: f.Dirname}
	}
	return opts
}

// isTerminal reports whether r is an interactive terminal. Non-file
// readers — test buffers, pipes — are never terminals.
func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	return ok && term.IsTerminal(f.Fd())
}

// runApply is the apply command's body, split out so a test can call it
// with a controlled command (and captured stdout/stderr) without going
// through the cobra tree. The returned error is surfaced to the caller
// (cobra, then main), which is responsible for printing it to stderr;
// the root command sets SilenceErrors so cobra does not print it itself.
func runApply(cmd *cobra.Command, dirname string, verbose bool) error {
	level := log.WarnLevel
	if verbose {
		level = log.InfoLevel
	}
	log.SetOutput(cmd.ErrOrStderr())
	log.SetLevel(level)
	return apply(cmd.OutOrStdout(), dirname)
}

// apply runs the full theming pipeline for one flavor: ensure the data
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
func apply(out io.Writer, dirname string) error {
	dataDir, err := paths.EnsureDataDir()
	if err != nil {
		return fmt.Errorf("apply %s: %w", dirname, err)
	}

	configDir, err := paths.ConfigDir()
	if err != nil {
		return fmt.Errorf("apply %s: %w", dirname, err)
	}

	flavorsDir, err := paths.FlavorsDir()
	if err != nil {
		return fmt.Errorf("apply %s: %w", dirname, err)
	}

	f, err := flavor.Load(flavorsDir, dirname)
	if err != nil {
		return fmt.Errorf("apply %s: %w", dirname, err)
	}

	cfg, err := config.Load(configDir)
	if err != nil {
		return fmt.Errorf("apply %s: %w", dirname, err)
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

	// Style rendering emits full ANSI; the colorprofile writer is the seam
	// that downsamples or strips it per destination, honoring NO_COLOR,
	// CLICOLOR_FORCE, and TERM=dumb, so styling degrades to plain text when
	// out is not a color-capable terminal (a test buffer, a pipe, ...).
	styled := colorprofile.NewWriter(out, os.Environ())

	// lipgloss.Color takes an ANSI index ("1"–"255") or hex ("#rrggbb"),
	// not a color name. The status symbol carries the state's color and
	// the app name stays plain, mirroring how gum and gh-dash render
	// state glyphs.
	appliedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	skippedStyle := lipgloss.NewStyle().Faint(true)
	boldStyle := lipgloss.NewStyle().Bold(true)

	fmt.Fprintf(styled, "%s\n\n", boldStyle.Render(fmt.Sprintf("%s Served up %s to:", cfg.Icon, f.Name)))
	var errs []error
	for _, a := range apps {
		err := a.r.Run(f)
		if err == nil {
			fmt.Fprintf(styled, "  %s %s\n", appliedStyle.Render("✓"), a.name)
			continue
		}
		fmt.Fprintf(styled, "  %s %s\n", skippedStyle.Render("•"), skippedStyle.Render(a.name))
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
