package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// apply <theme> with a real flavor and every template writes each
// recipe's output file and prints the served-up block: a header with the
// flavor's Name followed by a ☑ row per app that applied.
func TestApplyServesUpFlavorAndWritesAllThemes(t *testing.T) {
	configDir, dataDir := useTempDirs(t)
	writeFlavor(t, configDir, "catppuccin-mocha")

	root, out, _ := newRoot([]string{"apply", "catppuccin-mocha"})

	require.NoError(t, root.Execute())

	want := "☕ Served up Catppuccin Mocha to:\n\n" +
		"  ☑ fzf\n" +
		"  ☑ Ghostty\n" +
		"  ☑ Neovim\n" +
		"  ☑ Zellij\n\n"
	assert.Equal(t, want, out.String())

	gotGhostty, err := os.ReadFile(filepath.Join(dataDir, "barista", "ghostty"))
	require.NoError(t, err)
	assert.Equal(t, "name = Catppuccin Mocha\nbase = #1e1e2e", string(gotGhostty))

	gotNvim, err := os.ReadFile(filepath.Join(dataDir, "barista", "nvim", "lua", "flavor.lua"))
	require.NoError(t, err)
	assert.Equal(t, "local name = Catppuccin Mocha", string(gotNvim))

	gotZellij, err := os.ReadFile(filepath.Join(configDir, "barista", "zellij", "themes", "barista.kdl"))
	require.NoError(t, err)
	assert.Equal(t, "name = Catppuccin Mocha", string(gotZellij))

	gotFzf, err := os.ReadFile(filepath.Join(configDir, "barista", "fzfrc"))
	require.NoError(t, err)
	assert.Equal(t, "# Catppuccin Mocha\n", string(gotFzf))
}

// apply prints the flavor's Name, not the dirname.
func TestApplyUsesNameNotDirname(t *testing.T) {
	configDir, _ := useTempDirs(t)
	writeFlavor(t, configDir, "catppuccin-mocha")

	root, out, _ := newRoot([]string{"apply", "catppuccin-mocha"})

	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Catppuccin Mocha")
	assert.NotContains(t, out.String(), "catppuccin-mocha")
}

// A flavor that carries only some templates: the apps with templates
// apply (☑); the apps without templates are skipped (☐), not errors, and
// the run exits zero. The served-up list always names every configured app.
func TestApplyMissingTemplateSkipsNotErrors(t *testing.T) {
	configDir, dataDir := useTempDirs(t)
	// Write the flavor with only the Ghostty template; fzf, Neovim, and
	// Zellij templates are absent, so those recipes are skipped (☐).
	flavorDir := filepath.Join(configDir, "barista", "flavors", "catppuccin-mocha")
	require.NoError(t, os.MkdirAll(flavorDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "flavor.toml"), []byte(fullFlavorTOML), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "ghostty.mustache"), []byte("name = {{name}}"), 0o644))

	root, out, _ := newRoot([]string{"apply", "catppuccin-mocha"})

	err := root.Execute()
	require.NoError(t, err)

	assert.Contains(t, out.String(), "  ☐ fzf\n")
	assert.Contains(t, out.String(), "  ☑ Ghostty\n")
	assert.Contains(t, out.String(), "  ☐ Neovim\n")
	assert.Contains(t, out.String(), "  ☐ Zellij\n")

	// Ghostty still wrote its theme.
	assert.FileExists(t, filepath.Join(dataDir, "barista", "ghostty"))
}

// A missing flavor surfaces an error mentioning the dirname.
func TestApplyMissingFlavorFails(t *testing.T) {
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

// apply requires exactly one positional arg.
func TestApplyRequiresArg(t *testing.T) {
	root, _, _ := newRoot([]string{"apply"})

	assert.Error(t, root.Execute())
}

func TestRootHelpHasShortDescription(t *testing.T) {
	root, out, _ := newRoot([]string{"--help"})

	assert.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Serves up a new <theme> for your terminal apps.")
}

func TestRootNoArgsPrintsHelp(t *testing.T) {
	root, out, _ := newRoot(nil)

	assert.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Usage:",
		"root with no args should print help with a Usage block; got:\n%s", out.String())
}

// apply -v logs each recipe step to stderr, including the input files
// read, output files written, and reload actions performed.
func TestApplyVerboseLogsRecipeSteps(t *testing.T) {
	configDir, dataDir := useTempDirs(t)
	writeFlavor(t, configDir, "catppuccin-mocha")

	root, _, errOut := newRoot([]string{"apply", "-v", "catppuccin-mocha"})

	require.NoError(t, root.Execute())

	got := errOut.String()
	flavorsDir := filepath.Join(configDir, "barista", "flavors", "catppuccin-mocha")

	// fzf: template locate/read, render, file write.
	assert.Contains(t, got, filepath.Join(flavorsDir, "fzf.rc.mustache"))
	assert.Contains(t, got, "FZF_DEFAULT_OPTS_FILE is set")
	assert.Contains(t, got, "writing new file")

	// Ghostty: template read, theme write, reload (pgrep no-op in tests).
	assert.Contains(t, got, filepath.Join(flavorsDir, "ghostty.mustache"))
	assert.Contains(t, got, filepath.Join(dataDir, "barista", "ghostty"))
	assert.Contains(t, got, "Discovering ghostty pid")
	assert.Contains(t, got, "reload skipped")

	// Neovim: template read, dir create, theme write, reload.
	assert.Contains(t, got, filepath.Join(flavorsDir, "neovim.lua.mustache"))
	assert.Contains(t, got, filepath.Join(dataDir, "barista", "nvim", "lua", "flavor.lua"))
	assert.Contains(t, got, "Scanning neovim runtime directory")

	// Zellij: template read, dir create, theme write, config touch.
	assert.Contains(t, got, filepath.Join(flavorsDir, "zellij.kdl.mustache"))
	assert.Contains(t, got, filepath.Join(configDir, "barista", "zellij", "themes", "barista.kdl"))
	assert.Contains(t, got, "Touching config file")
	assert.Contains(t, got, filepath.Join(configDir, "barista", "zellij", "config.kdl"))
}

// apply --verbose works as the long form of -v.
func TestApplyVerboseLongFlagWorks(t *testing.T) {
	configDir, _ := useTempDirs(t)
	writeFlavor(t, configDir, "catppuccin-mocha")

	root, _, errOut := newRoot([]string{"apply", "--verbose", "catppuccin-mocha"})

	require.NoError(t, root.Execute())
	assert.Contains(t, errOut.String(), "Reading template")
	assert.Contains(t, errOut.String(), "Writing theme")
}

// apply without -v emits no per-step logs to stderr.
func TestApplyNonVerboseOmitsLogs(t *testing.T) {
	configDir, _ := useTempDirs(t)
	writeFlavor(t, configDir, "catppuccin-mocha")

	root, _, errOut := newRoot([]string{"apply", "catppuccin-mocha"})

	require.NoError(t, root.Execute())
	assert.NotContains(t, errOut.String(), "Reading template")
	assert.NotContains(t, errOut.String(), "Writing theme")
}

// apply -v logs steps only for applied apps; skipped apps log the
// template search and skip, but no render/write/reload lines.
func TestApplyVerboseSkipsLogsForSkippedApps(t *testing.T) {
	configDir, _ := useTempDirs(t)
	// Only the Ghostty template is present; fzf, Neovim, and Zellij are
	// skipped and should log the locate + skip but no render/write lines.
	flavorDir := filepath.Join(configDir, "barista", "flavors", "catppuccin-mocha")
	require.NoError(t, os.MkdirAll(flavorDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "flavor.toml"), []byte(fullFlavorTOML), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "ghostty.mustache"), []byte("name = {{name}}"), 0o644))

	root, _, errOut := newRoot([]string{"apply", "-v", "catppuccin-mocha"})

	require.NoError(t, root.Execute())
	got := errOut.String()

	// Ghostty applied: its render/write lines are present.
	assert.Contains(t, got, "Rendering template")
	assert.Contains(t, got, "Writing theme")
	assert.Contains(t, got, "Template not found; skipping")
	assert.NotContains(t, got, "writing new file")
	// Neovim skipped: no render line.
	assert.NotContains(t, got, "Creating plugin directory")
	// Zellij skipped: no config touch line.
	assert.NotContains(t, got, "Touching config file")
}
