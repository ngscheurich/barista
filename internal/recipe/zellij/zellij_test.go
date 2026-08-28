package zellij_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/recipe/zellij"
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

// writeFlavorDir creates a flavors dir containing <dirname>/zellij.kdl.mustache
// and returns the flavors dir path.
func writeFlavorDir(t *testing.T, dirname, tmpl string) string {
	t.Helper()
	dir := t.TempDir()
	flavorDir := filepath.Join(dir, dirname)
	require.NoError(t, os.MkdirAll(flavorDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "zellij.kdl.mustache"), []byte(tmpl), 0o644))
	return dir
}

// Run renders the template, creates <configDir>/zellij/themes/, and writes
// barista.kdl there. Reload touches config.kdl; the test asserts the file
// is touched (created if missing) rather than spawning a real touch.
func TestRunWritesBaristaKdlUnderZellijThemes(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha", "name = {{name}}\nbase = {{palette.base}}")
	configDir := t.TempDir()

	r := zellij.New(flavorsDir, configDir)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(configDir, "zellij", "themes", "barista.kdl"))
	require.NoError(t, err)
	assert.Equal(t, "name = Mocha\nbase = #1e1e2e", string(got))
}

// Run creates the zellij/themes/ directory tree when it does not yet exist.
func TestRunCreatesZellijThemesDir(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha", "{{name}}")
	configDir := t.TempDir()

	r := zellij.New(flavorsDir, configDir)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(configDir, "zellij", "themes"))
}

// Run touches <configDir>/zellij/config.kdl as its reload, creating it if
// it did not exist; the test asserts the file exists after Run.
func TestRunTouchesConfigKdl(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha", "{{name}}")
	configDir := t.TempDir()

	r := zellij.New(flavorsDir, configDir)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(configDir, "zellij", "config.kdl"))
}

// A missing template surfaces as a wrapped error naming the recipe.
func TestRunMissingTemplateFails(t *testing.T) {
	flavorsDir := t.TempDir() // no flavor dir written
	configDir := t.TempDir()

	r := zellij.New(flavorsDir, configDir)
	err := r.Run(sampleFlavor())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "zellij")
}

// The reload seam: Touch builds touch <path>, calling the binary directly
// rather than shelling out.
func TestTouchCommandShape(t *testing.T) {
	cmd := zellij.Touch("/home/u/.config/zellij/config.kdl")

	assert.Equal(t, "touch", filepath.Base(cmd.Path))
	assert.Equal(t, []string{"/home/u/.config/zellij/config.kdl"}, cmd.Args[1:])
}
