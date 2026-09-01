// Package neovim implements the recipe for theming and reloading Neovim:
// locate neovim.lua.mustache under the theme's directory, render it
// against the theme's Flavor, create <data dir>/nvim/lua/ and write
// barista.lua there, then send a reload keystroke to every discovered
// Neovim server socket via internal/nvim.
package neovim

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/log/v2"

	"github.com/ngscheurich/barista/internal/nvim"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/template"
	"github.com/ngscheurich/barista/internal/theme"
)

// templateName is the Mustache file a theme directory carries for
// Neovim; the recipe looks for it under <themesDir>/<theme.Dirname>.
const templateName = "neovim.lua.mustache"

// pluginDir is the directory the rendered output is written to, under the
// barista data directory; it is created if missing.
const pluginDir = "nvim/lua"

// outputName is the file the rendered output is written to inside
// pluginDir. barista.nvim reads this file; renaming it is a coordinated
// change with that repo.
const outputName = "barista.lua"

// Recipe is the Neovim recipe: it carries the directories it resolves
// templates and output from and runs the locate-render-write-reload
// procedure against a Theme.
type Recipe struct {
	themesDir string
	dataDir   string
}

// New builds a Neovim recipe that reads templates from themesDir and
// writes output under dataDir.
func New(themesDir, dataDir string) *Recipe {
	return &Recipe{themesDir: themesDir, dataDir: dataDir}
}

// Run renders the Neovim template against t.Flavor, writes barista.lua
// under <dataDir>/nvim/lua/, and reloads every running Neovim instance. A
// failure at any step is wrapped with this layer's role prefix; the
// orchestrator aggregates errors across recipes.
func (r *Recipe) Run(t theme.Theme) error {
	tmplPath := filepath.Join(r.themesDir, t.Dirname, templateName)
	log.Info("Locating template", "app", "neovim", "path", tmplPath)
	raw, err := os.ReadFile(tmplPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Info("Template not found; skipping", "app", "neovim")
			return fmt.Errorf("neovim: %w", recipe.ErrNotApplicable)
		}
		return fmt.Errorf("neovim: read template %s: %w", tmplPath, err)
	}
	log.Info("Reading template", "app", "neovim", "path", tmplPath)

	log.Info("Rendering template", "app", "neovim")
	rendered, err := template.Render(string(raw), t.Flavor)
	if err != nil {
		return fmt.Errorf("neovim: %w", err)
	}

	outDir := filepath.Join(r.dataDir, pluginDir)
	log.Info("Creating plugin directory", "app", "neovim", "path", outDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("neovim: create plugin dir %s: %w", outDir, err)
	}

	outPath := filepath.Join(outDir, outputName)
	log.Info("Writing output", "app", "neovim", "path", outPath)
	if err := os.WriteFile(outPath, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("neovim: write output %s: %w", outPath, err)
	}

	log.Info("Reloading neovim instances", "app", "neovim")
	if err := nvim.Reload(); err != nil {
		return fmt.Errorf("neovim: reload: %w", err)
	}
	return nil
}
