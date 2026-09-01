// Package zellij implements the recipe for theming and reloading Zellij:
// locate zellij.kdl.mustache under the theme's directory, render it
// against the theme's Flavor, create <config dir>/zellij/themes/ and write
// barista.kdl there, then touch <config dir>/zellij/config.kdl to trigger
// a reload.
//
// Unlike the Ghostty and Neovim recipes, which write under the data
// directory, Zellij reads its themes from the user's config directory, so
// this recipe carries the config dir rather than the data dir.
package zellij

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"charm.land/log/v2"

	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/template"
	"github.com/ngscheurich/barista/internal/theme"
)

// templateName is the Mustache file a theme directory carries for
// Zellij; the recipe looks for it under <themesDir>/<theme.Dirname>.
const templateName = "zellij.kdl.mustache"

// outputDir is the directory the rendered barista.kdl is written to,
// under the user's config directory; it is created if missing. Zellij
// reads theme files from this directory on reload.
const outputDir = "zellij/themes"

// outputName is the file the rendered output is written to inside outputDir.
const outputName = "barista.kdl"

// configFile is the Zellij main config file the recipe touches to trigger
// a reload, under the user's config directory.
const configFile = "zellij/config.kdl"

// Touch constructs the touch command used to reload Zellij. It returns
// the command without running it, so tests can assert the args without
// spawning a real touch.
var Touch = func(path string) *exec.Cmd {
	return exec.Command("touch", path)
}

// Recipe is the Zellij recipe: it carries the directories it resolves
// templates and output from and runs the locate-render-write-reload
// procedure against a Theme.
type Recipe struct {
	themesDir string
	configDir string
}

// New builds a Zellij recipe that reads templates from themesDir and
// writes output under configDir (Zellij looks for theme files under the
// user config dir, not the data dir).
func New(themesDir, configDir string) *Recipe {
	return &Recipe{themesDir: themesDir, configDir: configDir}
}

// Run renders the Zellij template against t.Flavor, writes barista.kdl
// under <configDir>/zellij/themes/, and touches <configDir>/zellij/config.kdl
// to reload Zellij. A failure at any step is wrapped with this layer's
// role prefix; the orchestrator aggregates errors across recipes.
func (r *Recipe) Run(t theme.Theme) error {
	tmplPath := filepath.Join(r.themesDir, t.Dirname, templateName)
	log.Info("Locating template", "app", "zellij", "path", tmplPath)
	raw, err := os.ReadFile(tmplPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Info("Template not found; skipping", "app", "zellij")
			return fmt.Errorf("zellij: %w", recipe.ErrNotApplicable)
		}
		return fmt.Errorf("zellij: read template %s: %w", tmplPath, err)
	}
	log.Info("Reading template", "app", "zellij", "path", tmplPath)

	log.Info("Rendering template", "app", "zellij")
	rendered, err := template.Render(string(raw), t.Flavor)
	if err != nil {
		return fmt.Errorf("zellij: %w", err)
	}

	outDir := filepath.Join(r.configDir, outputDir)
	log.Info("Creating themes directory", "app", "zellij", "path", outDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("zellij: create themes dir %s: %w", outDir, err)
	}

	outPath := filepath.Join(outDir, outputName)
	log.Info("Writing output", "app", "zellij", "path", outPath)
	if err := os.WriteFile(outPath, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("zellij: write output %s: %w", outPath, err)
	}

	cfgPath := filepath.Join(r.configDir, configFile)
	log.Info("Touching config file for reload", "app", "zellij", "path", cfgPath)
	if err := Touch(cfgPath).Run(); err != nil {
		return fmt.Errorf("zellij: touch %s: %w", cfgPath, err)
	}
	return nil
}
