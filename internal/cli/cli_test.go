package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// apply <theme> with a real theme and every template writes each
// recipe's artifact and prints the served-up block: a header with the
// theme's name followed by a ✓ row per app that applied.
func TestApplyServesUpThemeAndWritesAllArtifacts(t *testing.T) {
	configDir, dataDir := useTempDirs(t)
	writeTheme(t, configDir, "catppuccin-mocha")

	root, out, _ := newRoot([]string{"apply", "catppuccin-mocha"})

	require.NoError(t, root.Execute())

	want := "☕ Served up Catppuccin Mocha to:\n\n" +
		"  ✓ fzf\n" +
		"  ✓ Ghostty\n" +
		"  ✓ Neovim\n" +
		"  ✓ Zellij\n\n"
	assert.Equal(t, want, out.String())

	gotGhostty, err := os.ReadFile(filepath.Join(dataDir, "barista", "ghostty"))
	require.NoError(t, err)
	assert.Equal(t, "name = Catppuccin Mocha\nbase = #1e1e2e", string(gotGhostty))

	gotNvim, err := os.ReadFile(filepath.Join(dataDir, "barista", "nvim", "lua", "barista.lua"))
	require.NoError(t, err)
	assert.Equal(t, "local name = Catppuccin Mocha", string(gotNvim))

	gotZellij, err := os.ReadFile(filepath.Join(configDir, "zellij", "themes", "barista.kdl"))
	require.NoError(t, err)
	assert.Equal(t, "name = Catppuccin Mocha", string(gotZellij))
}

// apply prints the theme's name, not the dirname.
func TestApplyUsesNameNotDirname(t *testing.T) {
	configDir, _ := useTempDirs(t)
	writeTheme(t, configDir, "catppuccin-mocha")

	root, out, _ := newRoot([]string{"apply", "catppuccin-mocha"})

	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Catppuccin Mocha")
	assert.NotContains(t, out.String(), "catppuccin-mocha")
}

// apply respects a custom icon from config.toml.
func TestApplyCustomIcon(t *testing.T) {
	configDir, _ := useTempDirs(t)
	writeTheme(t, configDir, "catppuccin-mocha")
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "barista", "config.toml"),
		[]byte("icon = \"🍵\""),
		0o644,
	))

	root, out, _ := newRoot([]string{"apply", "catppuccin-mocha"})

	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "🍵 Served up")
	assert.NotContains(t, out.String(), "☕ Served up")
}

// A theme that carries only some templates: the apps with templates
// apply (✓); the apps without templates are skipped (•), not errors, and
// the run exits zero. The served-up list always names every configured app.
func TestApplyMissingTemplateSkipsNotErrors(t *testing.T) {
	configDir, dataDir := useTempDirs(t)
	// Write the theme with only the Ghostty template; fzf, Neovim, and
	// Zellij templates are absent, so those recipes are skipped (•).
	themeDir := filepath.Join(configDir, "barista", "themes", "catppuccin-mocha")
	require.NoError(t, os.MkdirAll(themeDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "flavor.toml"), []byte(fullFlavorTOML), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "ghostty.mustache"), []byte("name = {{name}}"), 0o644))

	root, out, _ := newRoot([]string{"apply", "catppuccin-mocha"})

	err := root.Execute()
	require.NoError(t, err)

	assert.Contains(t, out.String(), "  • fzf\n")
	assert.Contains(t, out.String(), "  ✓ Ghostty\n")
	assert.Contains(t, out.String(), "  • Neovim\n")
	assert.Contains(t, out.String(), "  • Zellij\n")

	// Ghostty still wrote its artifact.
	assert.FileExists(t, filepath.Join(dataDir, "barista", "ghostty"))
}

// A missing theme surfaces an error mentioning the dirname.
func TestApplyMissingThemeFails(t *testing.T) {
	useTempDirs(t)

	root, _, _ := newRoot([]string{"apply", "nope"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nope")
}

// apply rejects more than one positional arg.
func TestApplyRejectsExtraArgs(t *testing.T) {
	root, _, _ := newRoot([]string{"apply", "one", "two"})

	assert.Error(t, root.Execute())
}

func TestRootHelpHasShortDescription(t *testing.T) {
	root, out, _ := newRoot([]string{"--help"})

	assert.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Serves up a new theme for your terminal apps.")
}

func TestRootNoArgsPrintsHelp(t *testing.T) {
	root, out, _ := newRoot(nil)

	assert.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Usage:",
		"root with no args should print help with a Usage block; got:\n%s", out.String())
}

// apply -v logs each recipe step to stderr, including the input files
// read, artifacts written, and reload actions performed.
func TestApplyVerboseLogsRecipeSteps(t *testing.T) {
	configDir, dataDir := useTempDirs(t)
	writeTheme(t, configDir, "catppuccin-mocha")

	root, _, errOut := newRoot([]string{"apply", "-v", "catppuccin-mocha"})

	require.NoError(t, root.Execute())

	got := errOut.String()
	themesDir := filepath.Join(configDir, "barista", "themes", "catppuccin-mocha")

	// fzf: template read, render, env var read, merge, fish.
	assert.Contains(t, got, filepath.Join(themesDir, "fzf.mustache"))
	assert.Contains(t, got, "Read existing env var")
	assert.Contains(t, got, "Setting env var via fish")

	// Ghostty: template read, artifact write, reload (pgrep no-op in tests).
	assert.Contains(t, got, filepath.Join(themesDir, "ghostty.mustache"))
	assert.Contains(t, got, filepath.Join(dataDir, "barista", "ghostty"))
	assert.Contains(t, got, "Discovering ghostty pid")
	assert.Contains(t, got, "reload skipped")

	// Neovim: template read, dir create, artifact write, reload.
	assert.Contains(t, got, filepath.Join(themesDir, "neovim.lua.mustache"))
	assert.Contains(t, got, filepath.Join(dataDir, "barista", "nvim", "lua", "barista.lua"))
	assert.Contains(t, got, "Scanning neovim runtime directory")

	// Zellij: template read, dir create, artifact write, config touch.
	assert.Contains(t, got, filepath.Join(themesDir, "zellij.kdl.mustache"))
	assert.Contains(t, got, filepath.Join(configDir, "zellij", "themes", "barista.kdl"))
	assert.Contains(t, got, "Touching config file")
	assert.Contains(t, got, filepath.Join(configDir, "zellij", "config.kdl"))
}

// apply --verbose works as the long form of -v.
func TestApplyVerboseLongFlagWorks(t *testing.T) {
	configDir, _ := useTempDirs(t)
	writeTheme(t, configDir, "catppuccin-mocha")

	root, _, errOut := newRoot([]string{"apply", "--verbose", "catppuccin-mocha"})

	require.NoError(t, root.Execute())
	assert.Contains(t, errOut.String(), "Reading template")
	assert.Contains(t, errOut.String(), "Writing artifact")
}

// apply without -v emits no per-step logs to stderr.
func TestApplyNonVerboseOmitsLogs(t *testing.T) {
	configDir, _ := useTempDirs(t)
	writeTheme(t, configDir, "catppuccin-mocha")

	root, _, errOut := newRoot([]string{"apply", "catppuccin-mocha"})

	require.NoError(t, root.Execute())
	assert.NotContains(t, errOut.String(), "Reading template")
	assert.NotContains(t, errOut.String(), "Writing artifact")
}

// apply -v logs steps only for applied apps; skipped apps log the
// template search and skip, but no render/write/reload lines.
func TestApplyVerboseSkipsLogsForSkippedApps(t *testing.T) {
	configDir, _ := useTempDirs(t)
	// Only the Ghostty template is present; fzf, Neovim, and Zellij are
	// skipped and should log the locate + skip but no render/write lines.
	themeDir := filepath.Join(configDir, "barista", "themes", "catppuccin-mocha")
	require.NoError(t, os.MkdirAll(themeDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "flavor.toml"), []byte(fullFlavorTOML), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "ghostty.mustache"), []byte("name = {{name}}"), 0o644))

	root, _, errOut := newRoot([]string{"apply", "-v", "catppuccin-mocha"})

	require.NoError(t, root.Execute())
	got := errOut.String()

	// Ghostty applied: its render/write lines are present.
	assert.Contains(t, got, "Rendering template")
	assert.Contains(t, got, "Writing artifact")
	assert.Contains(t, got, "Template not found; skipping")
	assert.NotContains(t, got, "Setting env var via fish")
	// Neovim skipped: no render line.
	assert.NotContains(t, got, "Creating plugin directory")
	// Zellij skipped: no config touch line.
	assert.NotContains(t, got, "Touching config file")
}

// apply with no argument and a non-terminal stdin fails with a plain
// list of the available themes' dirnames — the scripted and
// screen-reader surface — instead of opening a picker it cannot run.
// The dirnames are what the user would type to retry.
func TestApplyNoArgNonTerminalListsThemes(t *testing.T) {
	configDir, _ := useTempDirs(t)
	writeTheme(t, configDir, "catppuccin-mocha")
	writeTheme(t, configDir, "catppuccin-latte")

	root, _, _ := newRoot([]string{"apply"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no theme given")
	assert.Contains(t, err.Error(), "catppuccin-mocha")
	assert.Contains(t, err.Error(), "catppuccin-latte")
}

// apply with no argument and BARISTA_ACCESSIBLE set runs the picker in
// accessible mode against stdin, so piping a choice applies that theme
// end-to-end through every recipe.
func TestApplyNoArgAccessibleAppliesChoice(t *testing.T) {
	configDir, dataDir := useTempDirs(t)
	writeTheme(t, configDir, "catppuccin-mocha")
	t.Setenv("BARISTA_ACCESSIBLE", "1")

	root, out, _ := newRoot([]string{"apply"})
	root.SetIn(strings.NewReader("1\n"))

	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Served up Catppuccin Mocha")

	gotGhostty, err := os.ReadFile(filepath.Join(dataDir, "barista", "ghostty"))
	require.NoError(t, err)
	assert.Equal(t, "name = Catppuccin Mocha\nbase = #1e1e2e", string(gotGhostty))
}

// apply with no argument and no themes installed fails before any
// picker opens, naming the directory where themes would live.
func TestApplyNoArgNoThemes(t *testing.T) {
	useTempDirs(t)

	root, _, _ := newRoot([]string{"apply"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no themes found")
}

// Two themes sharing a Flavor name (writeTheme's TOML has a fixed name)
// get their dirnames as picker labels, so the rows stay distinguishable;
// piping the row number still applies the right theme.
func TestApplyNoArgAccessibleDisambiguatesSharedNames(t *testing.T) {
	configDir, _ := useTempDirs(t)
	writeTheme(t, configDir, "catppuccin-mocha")
	writeTheme(t, configDir, "catppuccin-latte")
	t.Setenv("BARISTA_ACCESSIBLE", "1")

	root, out, _ := newRoot([]string{"apply"})
	root.SetIn(strings.NewReader("2\n"))

	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "1. catppuccin-latte")
	assert.Contains(t, out.String(), "2. catppuccin-mocha")

	gotZellij, err := os.ReadFile(filepath.Join(configDir, "zellij", "themes", "barista.kdl"))
	require.NoError(t, err)
	assert.Equal(t, "name = Catppuccin Mocha", string(gotZellij))
}
