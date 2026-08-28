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
)

// sampleFlavor is a Flavor with distinct palette values so the rendered
// output makes it obvious which field landed where.
func sampleFlavor() flavor.Flavor {
	return flavor.Flavor{
		Name:    "Mocha",
		Dirname: "mocha",
		Palette: flavor.Palette{
			Rosewater: "#f5e0dc",
			Base:      "#1e1e2e",
			Text:      "#cdd6f4",
			Crust:     "#11111b",
		},
	}
}

// writeFlavorDir creates a flavors dir containing <dirname>/ghostty.mustache
// with the given template content and returns the flavors dir path.
func writeFlavorDir(t *testing.T, dirname, tmpl string) string {
	t.Helper()
	dir := t.TempDir()
	flavorDir := filepath.Join(dir, dirname)
	require.NoError(t, os.MkdirAll(flavorDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "ghostty.mustache"), []byte(tmpl), 0o644))
	return dir
}

// Run writes the rendered template to <dataDir>/ghostty and reloads.
func TestRunWritesRenderedTheme(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha", "name = {{name}}\nbase = {{palette.base}}\ntext = {{palette.text}}")
	dataDir := t.TempDir()
	// No Ghostty is running in the test env; pgrep will exit non-zero and
	// reload is a no-op, so Run should still succeed and write the file.

	r := ghostty.New(flavorsDir, dataDir)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
	got, err := os.ReadFile(filepath.Join(dataDir, "ghostty"))
	require.NoError(t, err)
	assert.Equal(t, "name = Mocha\nbase = #1e1e2e\ntext = #cdd6f4", string(got))
}

// A missing template is not-applicable rather than a failure: Run returns
// an error wrapping recipe.ErrNotApplicable, which the orchestrator treats
// as a skip.
func TestRunMissingTemplateIsNotApplicable(t *testing.T) {
	flavorsDir := t.TempDir() // no flavor dir written
	dataDir := t.TempDir()

	r := ghostty.New(flavorsDir, dataDir)
	err := r.Run(sampleFlavor())

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
