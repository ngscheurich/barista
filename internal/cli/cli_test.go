package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// apply <theme> with a real flavor and template writes the rendered theme
// for every recipe to the data dir and prints the success line with the
// flavor's Name.
func TestApplyServesUpFlavorAndWritesTheme(t *testing.T) {
	configDir, dataDir := useTempDirs(t)
	writeFlavor(t, configDir, "catppuccin-mocha")

	root, out, _ := newRoot([]string{"apply", "catppuccin-mocha"})

	require.NoError(t, root.Execute())

	assert.Contains(t, out.String(), "☕︎ Served up Catppuccin Mocha")

	gotGhostty, err := os.ReadFile(filepath.Join(dataDir, "barista", "ghostty"))
	require.NoError(t, err)
	assert.Equal(t, "name = Catppuccin Mocha\nbase = #1e1e2e", string(gotGhostty))

	gotNvim, err := os.ReadFile(filepath.Join(dataDir, "barista", "nvim", "lua", "flavor.lua"))
	require.NoError(t, err)
	assert.Equal(t, "local name = Catppuccin Mocha", string(gotNvim))
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

// A missing flavor surfaces an error mentioning the dirname.
func TestApplyMissingFlavorFails(t *testing.T) {
	useTempDirs(t)

	root, _, errOut := newRoot([]string{"apply", "nope"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, errOut.String(), "nope")
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
