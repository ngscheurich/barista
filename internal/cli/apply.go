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
	"github.com/ngscheurich/barista/internal/paths"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/recipe/fzf"
	"github.com/ngscheurich/barista/internal/recipe/ghostty"
	"github.com/ngscheurich/barista/internal/recipe/neovim"
	"github.com/ngscheurich/barista/internal/recipe/zellij"
	"github.com/ngscheurich/barista/internal/theme"
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
		Use:   "apply [theme]",
		Short: "Apply a theme to all configured applications",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dirname := ""
			if len(args) > 0 {
				dirname = args[0]
			} else {
				var err error
				dirname, err = resolveTheme(cmd)
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

// resolveTheme decides which theme apply serves when no argument was
// given: the picker's choice. The picker opens only when stdin is a
// terminal or accessible mode is requested via BARISTA_ACCESSIBLE;
// otherwise the error lists the available themes' dirnames in plain
// text — the surface scripts and screen readers consume. An empty
// themes directory never reaches the picker.
func resolveTheme(cmd *cobra.Command) (string, error) {
	themesDir, err := paths.ThemesDir()
	if err != nil {
		return "", fmt.Errorf("apply: %w", err)
	}
	themes, err := theme.List(themesDir)
	if err != nil {
		// A themes directory that does not exist yet is the same
		// situation as an empty one: nothing to pick from.
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("no themes found in %s", themesDir)
		}
		return "", fmt.Errorf("apply: %w", err)
	}
	if len(themes) == 0 {
		return "", fmt.Errorf("no themes found in %s", themesDir)
	}

	accessible := os.Getenv("BARISTA_ACCESSIBLE") != ""
	in := cmd.InOrStdin()
	if !accessible && !isTerminal(in) {
		names := make([]string, len(themes))
		for i, t := range themes {
			names[i] = t.Dirname
		}
		return "", fmt.Errorf("no theme given; available: %s", strings.Join(names, ", "))
	}

	choice, err := ui.Pick(cmd.OutOrStdout(), in, pickerOptions(themes), accessible)
	if err != nil {
		return "", fmt.Errorf("apply: %w", err)
	}
	return choice, nil
}

// pickerOptions maps themes to picker rows: each theme's Flavor name, with
// its Dirname as the label when two themes share a Flavor name, so no two
// rows are indistinguishable. The chosen row's Value is the Dirname
// apply takes as its argument.
func pickerOptions(themes []theme.Theme) []ui.Option {
	counts := make(map[string]int)
	for _, t := range themes {
		counts[t.Flavor.Name]++
	}
	opts := make([]ui.Option, len(themes))
	for i, t := range themes {
		label := t.Flavor.Name
		if counts[t.Flavor.Name] > 1 {
			label = t.Dirname
		}
		opts[i] = ui.Option{Label: label, Value: t.Dirname}
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

// apply runs the full theming pipeline for one theme: ensure the data
// dir, load the theme, run every recipe, and print the served-up list
// marking each app ✓ (applied) or • (skipped, because the theme carries
// no template for it). Pre-recipe failures (missing data dir, missing
// theme) short-circuit and return. A recipe that returns
// recipe.ErrNotApplicable is a skip, not a failure; every other recipe
// error is aggregated via errors.Join so one app's failure does not hide
// another's. The served-up list always names every configured app; the
// caller is responsible for printing the returned error to stderr.
//
// Per-step logs (template reads, artifact writes, reload actions) are
// emitted by each recipe via the package-level charmbracelet/log default
// logger, whose level and output are configured by runApply from the
// --verbose flag.
func apply(out io.Writer, dirname string) error {
	dataDir, err := paths.EnsureDataDir()
	if err != nil {
		return fmt.Errorf("apply %s: %w", dirname, err)
	}

	configDir, err := paths.ConfigDir()
	if err != nil {
		return fmt.Errorf("apply %s: %w", dirname, err)
	}

	zellijConfigDir, err := paths.ZellijConfigDir()
	if err != nil {
		return fmt.Errorf("apply %s: %w", dirname, err)
	}

	themesDir, err := paths.ThemesDir()
	if err != nil {
		return fmt.Errorf("apply %s: %w", dirname, err)
	}

	t, err := theme.Load(themesDir, dirname)
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
		{name: "fzf", r: fzf.New(themesDir)},
		{name: "Ghostty", r: ghostty.New(themesDir, dataDir)},
		{name: "Neovim", r: neovim.New(themesDir, dataDir)},
		{name: "Zellij", r: zellij.New(themesDir, zellijConfigDir)},
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

	fmt.Fprintf(styled, "%s\n\n", boldStyle.Render(fmt.Sprintf("%s Served up %s to:", cfg.Icon, t.Flavor.Name)))
	var errs []error
	for _, a := range apps {
		err := a.r.Run(t)
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
