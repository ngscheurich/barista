// Package theme defines the Theme domain type and its loading from disk.
//
// A Theme is what barista serves: a Flavor (a named Catppuccin palette
// variant) plus the per-application templates that render that Flavor
// into each application's effect. On disk a Theme is a directory under
// the themes directory containing a flavor.toml (the Flavor) and one
// template per application.
package theme

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ngscheurich/barista/internal/flavor"
)

// ErrNotFound is returned when a theme's flavor.toml is missing, i.e. the
// directory does not name a Theme. Callers distinguish missing-theme from
// parse failures via errors.Is.
var ErrNotFound = errors.New("theme not found")

// Theme pairs a Flavor with the directory it was loaded from. Dirname is
// the theme's directory name (the argument to `barista apply`); Flavor is
// the named palette variant the templates render.
type Theme struct {
	Dirname string
	Flavor  flavor.Flavor
}

// List enumerates the themes directory, returning every Theme that loads,
// sorted by the Flavor's Name. Entries that are not themes (plain files,
// directories without a flavor.toml) and themes whose flavor.toml fails
// to load are skipped: they are not available, and `apply <dirname>`
// reports their errors when invoked directly. A themes directory that
// cannot be read is an error.
func List(themesDir string) ([]Theme, error) {
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return nil, fmt.Errorf("list themes %s: %w", themesDir, err)
	}
	var themes []Theme
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		t, err := Load(themesDir, e.Name())
		if err != nil {
			continue
		}
		themes = append(themes, t)
	}
	slices.SortFunc(themes, func(a, b Theme) int {
		return strings.Compare(a.Flavor.Name, b.Flavor.Name)
	})
	return themes, nil
}

// Load reads the theme named dirname from the themes directory: it reads
// <themesDir>/<dirname>/flavor.toml, decodes the Flavor, and returns the
// Theme pairing that Flavor with its directory.
func Load(themesDir, dirname string) (Theme, error) {
	p := filepath.Join(themesDir, dirname, "flavor.toml")
	raw, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Theme{}, fmt.Errorf("load theme %s: %w", dirname, ErrNotFound)
		}
		return Theme{}, fmt.Errorf("load theme %s: %w", dirname, err)
	}

	f, err := flavor.Parse(raw)
	if err != nil {
		return Theme{}, fmt.Errorf("load theme %s: %w", dirname, err)
	}
	return Theme{Dirname: dirname, Flavor: f}, nil
}
