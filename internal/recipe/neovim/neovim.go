// Package neovim implements the recipe for theming and reloading Neovim:
// locate neovim.lua.mustache under the flavor's directory, render it
// against the flavor, create <data dir>/nvim/lua/ and write flavor.lua
// there, then send a reload keystroke to every discovered Neovim server
// socket via internal/nvim.
package neovim

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/nvim"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/template"
)

// templateName is the Mustache file a flavor directory carries for
// Neovim; the recipe looks for it under <flavorsDir>/<flavor.Dirname>.
const templateName = "neovim.lua.mustache"

// pluginDir is the directory the rendered flavor.lua is written to,
// under the barista data directory; it is created if missing.
const pluginDir = "nvim/lua"

// outputName is the file the rendered theme is written to inside pluginDir.
const outputName = "flavor.lua"

// Recipe is the Neovim recipe: it carries the directories it resolves
// templates and output from and runs the locate-render-write-reload
// procedure against a Flavor.
type Recipe struct {
	flavorsDir string
	dataDir    string
}

// New builds a Neovim recipe that reads templates from flavorsDir and
// writes themes under dataDir.
func New(flavorsDir, dataDir string) *Recipe {
	return &Recipe{flavorsDir: flavorsDir, dataDir: dataDir}
}

// Run renders the Neovim template against f, writes flavor.lua under
// <dataDir>/nvim/lua/, and reloads every running Neovim instance. A
// failure at any step is wrapped with this layer's role prefix; the
// orchestrator aggregates errors across recipes.
func (r *Recipe) Run(f flavor.Flavor) error {
	tmplPath := filepath.Join(r.flavorsDir, f.Dirname, templateName)
	raw, err := os.ReadFile(tmplPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("neovim: %w", recipe.ErrNotApplicable)
		}
		return fmt.Errorf("neovim: read template %s: %w", tmplPath, err)
	}

	rendered, err := template.Render(string(raw), f)
	if err != nil {
		return fmt.Errorf("neovim: %w", err)
	}

	outDir := filepath.Join(r.dataDir, pluginDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("neovim: create plugin dir %s: %w", outDir, err)
	}

	outPath := filepath.Join(outDir, outputName)
	if err := os.WriteFile(outPath, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("neovim: write theme %s: %w", outPath, err)
	}

	if err := nvim.Reload(); err != nil {
		return fmt.Errorf("neovim: reload: %w", err)
	}
	return nil
}
