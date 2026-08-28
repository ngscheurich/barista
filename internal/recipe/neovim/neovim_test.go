package neovim_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/recipe/neovim"
)

// sampleFlavor carries distinct palette values so the rendered output
// makes it obvious which field landed where.
func sampleFlavor() flavor.Flavor {
	return flavor.Flavor{
		Name:    "Mocha",
		Dirname: "mocha",
		Palette: flavor.Palette{
			Base:  "#1e1e2e",
			Text:  "#cdd6f4",
			Crust: "#11111b",
		},
	}
}

// writeFlavorDir creates a flavors dir containing <dirname>/neovim.lua.mustache
// and returns the flavors dir path.
func writeFlavorDir(t *testing.T, dirname, tmpl string) string {
	t.Helper()
	dir := t.TempDir()
	flavorDir := filepath.Join(dir, dirname)
	require.NoError(t, os.MkdirAll(flavorDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "neovim.lua.mustache"), []byte(tmpl), 0o644))
	return dir
}

// Run renders the template, creates <dataDir>/nvim/lua/, and writes
// flavor.lua there. No Neovim is running in the test env, so reload is a
// no-op and Run should succeed.
func TestRunWritesFlavorLuaUnderNvimLua(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha", "local name = {{name}}\nlocal base = {{palette.base}}")
	dataDir := t.TempDir()
	// Point XDG_RUNTIME_DIR at an empty temp dir so DiscoverSockets finds
	// nothing and reload is a no-op.
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	r := neovim.New(flavorsDir, dataDir)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(dataDir, "nvim", "lua", "flavor.lua"))
	require.NoError(t, err)
	assert.Equal(t, "local name = Mocha\nlocal base = #1e1e2e", string(got))
}

// Run creates the nvim/lua/ directory tree when it does not yet exist.
func TestRunCreatesNvimLuaDir(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha", "{{name}}")
	dataDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	r := neovim.New(flavorsDir, dataDir)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(dataDir, "nvim", "lua"))
}

// A missing template surfaces as a wrapped error naming the recipe.
func TestRunMissingTemplateFails(t *testing.T) {
	flavorsDir := t.TempDir() // no flavor dir written
	dataDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	r := neovim.New(flavorsDir, dataDir)
	err := r.Run(sampleFlavor())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "neovim")
}
