package ghostty_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/recipe/ghostty"
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
				Rosewater: "#f5e0dc",
				Base:      "#1e1e2e",
				Text:      "#cdd6f4",
				Crust:     "#11111b",
			},
		},
	}
}

// writeThemeDir creates a themes dir containing <dirname>/ghostty.mustache
// with the given template content and returns the themes dir path.
func writeThemeDir(t *testing.T, dirname, tmpl string) string {
	t.Helper()
	dir := t.TempDir()
	themeDir := filepath.Join(dir, dirname)
	require.NoError(t, os.MkdirAll(themeDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(themeDir, "ghostty.mustache"), []byte(tmpl), 0o644))
	return dir
}

// Run writes the rendered template to <dataDir>/ghostty and reloads.
func TestRunWritesRenderedTheme(t *testing.T) {
	themesDir := writeThemeDir(t, "mocha", "name = {{name}}\nbase = {{palette.base}}\ntext = {{palette.text}}")
	dataDir := t.TempDir()
	// No Ghostty is running in the test env; pgrep will exit non-zero and
	// reload is a no-op, so Run should still succeed and write the file.

	r := ghostty.New(themesDir, dataDir)
	err := r.Run(sampleTheme())

	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(dataDir, "ghostty"))
	require.NoError(t, err)
	assert.Equal(t, "name = Mocha\nbase = #1e1e2e\ntext = #cdd6f4", string(got))
}

// A missing template is not-applicable rather than a failure: Run returns
// an error wrapping recipe.ErrNotApplicable, which the orchestrator treats
// as a skip.
func TestRunMissingTemplateIsNotApplicable(t *testing.T) {
	themesDir := t.TempDir() // no theme dir written
	dataDir := t.TempDir()

	r := ghostty.New(themesDir, dataDir)
	err := r.Run(sampleTheme())

	require.Error(t, err)
	assert.True(t, errors.Is(err, recipe.ErrNotApplicable),
		"missing template should wrap recipe.ErrNotApplicable; got %v", err)
}

// The reload seam: pgrep is built with the binary name and a single arg
// "ghostty", without going through sh -c.
func TestPgrepCommandShape(t *testing.T) {
	cmd := ghostty.Pgrep()

	assert.Equal(t, "pgrep", filepath.Base(cmd.Path))
	assert.Equal(t, []string{"ghostty"}, cmd.Args[1:])
}

// The reload seam: kill is built as kill -s USR2 <pid>, calling the binary
// directly rather than shelling out.
func TestKillCommandShape(t *testing.T) {
	cmd := ghostty.Kill("4242")

	assert.Equal(t, "kill", filepath.Base(cmd.Path))
	assert.Equal(t, []string{"-s", "USR2", "4242"}, cmd.Args[1:])
}
