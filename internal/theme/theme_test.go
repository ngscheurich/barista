package theme_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/theme"
)

// validFlavorTOML is a complete flavor.toml with all 26 colors.
const validFlavorTOML = `name = "Catppuccin Mocha"

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

func wantFlavor() flavor.Flavor {
	return flavor.Flavor{
		Name: "Catppuccin Mocha",
		Palette: flavor.Palette{
			Rosewater: "#f5e0dc",
			Flamingo:  "#f2cdcd",
			Pink:      "#f5c2e7",
			Mauve:     "#cba6f7",
			Red:       "#f38ba8",
			Maroon:    "#eba0ac",
			Peach:     "#fab387",
			Yellow:    "#f9e2af",
			Green:     "#a6e3a1",
			Teal:      "#94e2d5",
			Sky:       "#89dceb",
			Sapphire:  "#74c7ec",
			Blue:      "#89b4fa",
			Lavender:  "#b4befe",
			Text:      "#cdd6f4",
			Subtext1:  "#bac2de",
			Subtext0:  "#a6adc8",
			Overlay2:  "#9399b2",
			Overlay1:  "#7f849c",
			Overlay0:  "#6c7086",
			Surface2:  "#585b70",
			Surface1:  "#45475a",
			Surface0:  "#313244",
			Base:      "#1e1e2e",
			Mantle:    "#181825",
			Crust:     "#11111b",
		},
	}
}

// Load reads the theme's flavor.toml and returns a Theme pairing the
// decoded Flavor with its directory name.
func TestLoad(t *testing.T) {
	themesDir := writeTheme(t, "catppuccin-mocha", validFlavorTOML)

	got, err := theme.Load(themesDir, "catppuccin-mocha")

	require.NoError(t, err)
	assert.Equal(t, theme.Theme{
		Dirname: "catppuccin-mocha",
		Flavor:  wantFlavor(),
	}, got)
}

// A missing flavor.toml returns an error wrapping ErrNotFound, so callers
// can distinguish a missing theme from a parse failure via errors.Is.
func TestLoadMissingFileWrapsErrNotFound(t *testing.T) {
	themesDir := t.TempDir()

	_, err := theme.Load(themesDir, "nope")

	require.Error(t, err)
	assert.ErrorIs(t, err, theme.ErrNotFound)
	assert.NotErrorIs(t, err, os.ErrNotExist,
		"ErrNotFound is the domain sentinel, not the raw os error")
}

// A flavor.toml that fails to parse wraps the parse error rather than
// ErrNotFound.
func TestLoadParseErrorIsNotNotFound(t *testing.T) {
	themesDir := writeTheme(t, "broken", "name = \"Broken\"\n")

	_, err := theme.Load(themesDir, "broken")

	require.Error(t, err)
	assert.NotErrorIs(t, err, theme.ErrNotFound)
}

// A compile-time guard that the sentinel is an error value, not a type.
var _ error = theme.ErrNotFound

// List enumerates the themes directory: every subdirectory whose
// flavor.toml loads becomes a Theme, sorted by the Flavor's Name.
// Directories without a flavor.toml, entries that fail to load, and plain
// files are skipped; a missing themes directory is an error.
func TestList(t *testing.T) {
	dir := t.TempDir()
	writeThemeTOML(t, dir, "catppuccin-mocha", validFlavorTOML)
	latte := strings.Replace(validFlavorTOML, `"Catppuccin Mocha"`, `"Catppuccin Latte"`, 1)
	writeThemeTOML(t, dir, "catppuccin-latte", latte)

	// A directory with no flavor.toml is not a theme.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "not-a-theme"), 0o755))
	// A flavor.toml that fails to load is skipped, not fatal.
	writeThemeTOML(t, dir, "broken", "name = \"Broken\"\n")
	// A plain file among the directories is ignored.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("hi"), 0o644))

	themes, err := theme.List(dir)
	require.NoError(t, err)

	wantMocha := theme.Theme{Dirname: "catppuccin-mocha", Flavor: wantFlavor()}
	wantLatte := theme.Theme{Dirname: "catppuccin-latte", Flavor: wantFlavor()}
	wantLatte.Flavor.Name = "Catppuccin Latte"

	assert.Equal(t, []theme.Theme{wantLatte, wantMocha}, themes)
}

// An empty themes directory yields no themes and no error.
func TestListEmptyDir(t *testing.T) {
	themes, err := theme.List(t.TempDir())
	require.NoError(t, err)
	assert.Empty(t, themes)
}

// A themes directory that cannot be read is an error.
func TestListMissingDir(t *testing.T) {
	_, err := theme.List(filepath.Join(t.TempDir(), "nope"))
	assert.Error(t, err)
}

// writeTheme creates <themesDir>/<dirname>/flavor.toml with the given
// content and returns themesDir so Load can be pointed at it.
func writeTheme(t *testing.T, dirname, toml string) string {
	t.Helper()
	themesDir := t.TempDir()
	writeThemeTOML(t, themesDir, dirname, toml)
	return themesDir
}

// writeThemeTOML creates <dir>/<dirname>/flavor.toml with the given
// contents, making the parent directory first.
func writeThemeTOML(t *testing.T, dir, dirname, toml string) {
	t.Helper()
	d := filepath.Join(dir, dirname)
	require.NoError(t, os.MkdirAll(d, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(d, "flavor.toml"), []byte(toml), 0o644))
}
