package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/cli"
	"github.com/ngscheurich/barista/internal/recipe/fzf"
)

// newRoot builds the cobra root command with its stdout and stderr
// captured into buffers and its args set, so a test can run a command
// and assert on its output.
func newRoot(args []string) (root *cobra.Command, out, errOut *bytes.Buffer) {
	out, errOut = &bytes.Buffer{}, &bytes.Buffer{}
	root = cli.NewRoot()
	root.SetOut(out)
	root.SetErr(errOut)
	root.SetArgs(args)
	return root, out, errOut
}

// useTempDirs points XDG_CONFIG_HOME, XDG_DATA_HOME, and XDG_RUNTIME_DIR
// at freshly-made temp directories and returns the config and data dirs,
// so the apply command resolves paths inside the test sandbox rather than
// the user's real config. XDG_RUNTIME_DIR is sandboxed so Neovim socket
// discovery does not touch the real runtime dir.
func useTempDirs(t *testing.T) (configDir, dataDir string) {
	t.Helper()
	configDir = t.TempDir()
	dataDir = t.TempDir()
	runtimeDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "barista"), 0o755))
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", dataDir)
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	origFish := fzf.Fish
	t.Cleanup(func() { fzf.Fish = origFish })
	fzf.Fish = func(string) *exec.Cmd { return exec.Command("true") }
	return configDir, dataDir
}

// writeTheme creates a themes/<dirname> directory under configDir with a
// complete flavor.toml and a mustache template per recipe, so apply can
// load and render the theme end-to-end across every recipe.
func writeTheme(t *testing.T, configDir, dirname string) {
	t.Helper()
	themeDir := filepath.Join(configDir, "barista", "themes", dirname)
	require.NoError(t, os.MkdirAll(themeDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "flavor.toml"), []byte(fullFlavorTOML), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "fzf.mustache"), []byte("# {{name}}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "ghostty.mustache"), []byte("name = {{name}}\nbase = {{palette.base}}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "neovim.lua.mustache"), []byte("local name = {{name}}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "zellij.kdl.mustache"), []byte("name = {{name}}"), 0o644))
}

// fullFlavorTOML is a complete flavor.toml with all 26 colors so theme.Load
// succeeds; only a few colors are referenced by the test template.
const fullFlavorTOML = `name = "Catppuccin Mocha"

[palette]
rosewater = "#f5e0dc"
flamingo = "#f2cdcd"
pink = "#f5c2e7"
mauve = "#cba6f7"
red = "#f38ba8"
maroon = "#eba0ac"
peach = "#fab387"
yellow = "#f9e2af"
green = "#a6e3a1"
teal = "#94e2d5"
sky = "#89dceb"
sapphire = "#74c7ec"
blue = "#89b4fa"
lavender = "#b4befe"
text = "#cdd6f4"
subtext_1 = "#bac2de"
subtext_0 = "#a6adc8"
overlay_2 = "#9399b2"
overlay_1 = "#7f849c"
overlay_0 = "#6c7086"
surface_2 = "#585b70"
surface_1 = "#45475a"
surface_0 = "#313244"
base = "#1e1e2e"
mantle = "#181825"
crust = "#11111b"
`
