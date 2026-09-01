package zellij_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/recipe/zellij"
	"github.com/ngscheurich/barista/internal/theme"
)

// sampleTheme carries a Flavor with distinct palette values so the
// rendered template makes it obvious which field landed where.
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

// writeThemeDir creates a themes dir containing <dirname>/zellij.kdl.mustache
// and returns the themes dir path.
func writeThemeDir(t *testing.T, dirname, tmpl string) string {
	t.Helper()
	dir := t.TempDir()
	themeDir := filepath.Join(dir, dirname)
	require.NoError(t, os.MkdirAll(themeDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "zellij.kdl.mustache"), []byte(tmpl), 0o644))
	return dir
}

// Run renders the template, creates <configDir>/themes/, and writes
// barista.kdl there. Reload touches config.kdl; the test asserts the file
// is touched (created if missing) rather than spawning a real touch.
func TestRunWritesBaristaKdlUnderZellijThemes(t *testing.T) {
	themesDir := writeThemeDir(t, "mocha", "name = {{name}}\nbase = {{palette.base}}")
	configDir := t.TempDir()

	r := zellij.New(themesDir, configDir)
	err := r.Run(sampleTheme())

	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(configDir, "themes", "barista.kdl"))
	require.NoError(t, err)
	assert.Equal(t, "name = Mocha\nbase = #1e1e2e", string(got))
}

// Run creates the themes/ directory tree when it does not yet exist.
func TestRunCreatesZellijThemesDir(t *testing.T) {
	themesDir := writeThemeDir(t, "mocha", "{{name}}")
	configDir := t.TempDir()

	r := zellij.New(themesDir, configDir)
	err := r.Run(sampleTheme())

	require.NoError(t, err)
	assert.DirExists(t, filepath.Join(configDir, "themes"))
}

// Run touches <configDir>/config.kdl as its reload, creating it if
// it did not exist; the test asserts the file exists after Run.
func TestRunTouchesConfigKdl(t *testing.T) {
	themesDir := writeThemeDir(t, "mocha", "{{name}}")
	configDir := t.TempDir()

	r := zellij.New(themesDir, configDir)
	err := r.Run(sampleTheme())

	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(configDir, "config.kdl"))
}

// A missing template is not-applicable rather than a failure: Run returns
// an error wrapping recipe.ErrNotApplicable, which the orchestrator treats
// as a skip.
func TestRunMissingTemplateIsNotApplicable(t *testing.T) {
	themesDir := t.TempDir() // no theme dir written
	configDir := t.TempDir()

	r := zellij.New(themesDir, configDir)
	err := r.Run(sampleTheme())

	require.Error(t, err)
	assert.True(t, errors.Is(err, recipe.ErrNotApplicable),
		"missing template should wrap recipe.ErrNotApplicable; got %v", err)
}

// The reload seam: Touch builds touch <path>, calling the binary directly
// rather than shelling out.
func TestTouchCommandShape(t *testing.T) {
	cmd := zellij.Touch("/home/u/.config/zellij/config.kdl")

	assert.Equal(t, "touch", filepath.Base(cmd.Path))
	assert.Equal(t, []string{"/home/u/.config/zellij/config.kdl"}, cmd.Args[1:])
}
