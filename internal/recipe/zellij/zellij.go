// Package zellij implements the recipe for theming and reloading Zellij:
// locate zellij.kdl.mustache under the flavor's directory, render it
// against the flavor, create <config dir>/zellij/themes/ and write
// barista.kdl there, then touch <config dir>/zellij/config.kdl to trigger
// a reload.
//
// Unlike the Ghostty and Neovim recipes, which write under the data
// directory, Zellij reads its themes from the user's config directory, so
// this recipe carries the config dir rather than the data dir.
package zellij

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/template"
)

// templateName is the Mustache file a flavor directory carries for
// Zellij; the recipe looks for it under <flavorsDir>/<flavor.Dirname>.
const templateName = "zellij.kdl.mustache"

// themesDir is the directory the rendered barista.kdl is written to,
// under the user's config directory; it is created if missing.
const themesDir = "zellij/themes"

// outputName is the file the rendered theme is written to inside themesDir.
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
// procedure against a Flavor.
type Recipe struct {
	flavorsDir string
	configDir  string
}

// New builds a Zellij recipe that reads templates from flavorsDir and
// writes themes under configDir (Zellij looks for themes under the user
// config dir, not the data dir).
func New(flavorsDir, configDir string) *Recipe {
	return &Recipe{flavorsDir: flavorsDir, configDir: configDir}
}

// Run renders the Zellij template against f, writes barista.kdl under
// <configDir>/zellij/themes/, and touches <configDir>/zellij/config.kdl
// to reload Zellij. A failure at any step is wrapped with this layer's
// role prefix; the orchestrator aggregates errors across recipes.
func (r *Recipe) Run(f flavor.Flavor) error {
	tmplPath := filepath.Join(r.flavorsDir, f.Dirname, templateName)
	raw, err := os.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("zellij: read template %s: %w", tmplPath, err)
	}

	rendered, err := template.Render(string(raw), f)
	if err != nil {
		return fmt.Errorf("zellij: %w", err)
	}

	outDir := filepath.Join(r.configDir, themesDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("zellij: create themes dir %s: %w", outDir, err)
	}

	outPath := filepath.Join(outDir, outputName)
	if err := os.WriteFile(outPath, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("zellij: write theme %s: %w", outPath, err)
	}

	cfgPath := filepath.Join(r.configDir, configFile)
	if err := Touch(cfgPath).Run(); err != nil {
		return fmt.Errorf("zellij: touch %s: %w", cfgPath, err)
	}
	return nil
}
