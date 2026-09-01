package neovim_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/recipe/neovim"
	"github.com/ngscheurich/barista/internal/theme"
)

// sampleTheme carries a Flavor with distinct palette values so the
// rendered output makes it obvious which field landed where.
func sampleTheme() theme.Theme {
	return theme.Theme{
		Dirname: "mocha",
		Flavor: flavor.Flavor{
			Name: "Mocha",
			Palette: flavor.Palette{
				Base:  "#1e1e2e",
				Text:  "#cdd6f4",
				Crust: "#11111b",
			},
		},
	}
}

// writeThemeDir creates a themes dir containing <dirname>/neovim.lua.mustache
// and returns the themes dir path.
func writeThemeDir(t *testing.T, dirname, tmpl string) string {
	t.Helper()
	dir := t.TempDir()
	themeDir := filepath.Join(dir, dirname)
	require.NoError(t, os.MkdirAll(themeDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "neovim.lua.mustache"), []byte(tmpl), 0o644))
	return dir
}

// Run renders the template, creates <dataDir>/nvim/lua/, and writes
// barista.lua there. No Neovim is running in the test env, so reload is a
// no-op and Run should succeed.
func TestRunWritesBaristaLuaUnderNvimLua(t *testing.T) {
	themesDir := writeThemeDir(t, "mocha", "local name = {{name}}\nlocal base = {{palette.base}}")
	dataDir := t.TempDir()
	// Point XDG_RUNTIME_DIR at an empty temp dir so DiscoverSockets finds
	// nothing and reload is a no-op.
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	r := neovim.New(themesDir, dataDir)
	err := r.Run(sampleTheme())

	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(dataDir, "nvim", "lua", "barista.lua"))
	require.NoError(t, err)
	assert.Equal(t, "local name = Mocha\nlocal base = #1e1e2e", string(got))
}

// Run creates the nvim/lua/ directory tree when it does not yet exist.
func TestRunCreatesNvimLuaDir(t *testing.T) {
	themesDir := writeThemeDir(t, "mocha", "{{name}}")
	dataDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	r := neovim.New(themesDir, dataDir)
	err := r.Run(sampleTheme())

	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(dataDir, "nvim", "lua"))
}

// A missing template surfaces as a wrapped error naming the recipe.
func TestRunMissingTemplateFails(t *testing.T) {
	themesDir := t.TempDir() // no theme dir written
	dataDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	r := neovim.New(themesDir, dataDir)
	err := r.Run(sampleTheme())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "neovim")
}
