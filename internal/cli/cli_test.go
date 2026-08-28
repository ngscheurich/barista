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
	// Write the flavor with only the Ghostty template; Fish, Neovim, and
	// Zellij templates are absent, so those recipes are skipped (☐).
	flavorDir := filepath.Join(configDir, "barista", "flavors", "catppuccin-mocha")
	require.NoError(t, os.MkdirAll(flavorDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "flavor.toml"), []byte(fullFlavorTOML), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "ghostty.mustache"), []byte("name = {{name}}"), 0o644))

	root, out, _ := newRoot([]string{"apply", "catppuccin-mocha"})

	err := root.Execute()
	require.NoError(t, err)

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
