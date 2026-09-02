// Package ghostty implements the recipe for theming and reloading the
// Ghostty terminal: locate ghostty.mustache under the theme's directory,
// render it against the theme's Flavor, write the artifact to
// <data dir>/ghostty, and send SIGUSR2 to the running Ghostty process to
// trigger a reload.
package ghostty

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"charm.land/log/v2"

	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/template"
	"github.com/ngscheurich/barista/internal/theme"
)

// templateName is the Mustache file a theme directory carries for
// Ghostty; the recipe looks for it under <themesDir>/<theme.Dirname>.
const templateName = "ghostty.mustache"

// artifactName is the file the rendered artifact is written to, under
// the barista data directory.
const artifactName = "ghostty"

// Pgrep constructs the pgrep command used to discover the Ghostty pid.
// It returns the command without running it, so tests can assert the
// args and integration code can swap in a fake that returns a fixed pid.
var Pgrep = func() *exec.Cmd {
	return exec.Command("pgrep", "ghostty")
}

// Kill constructs the kill command used to signal a reload. It returns
// the command without running it, so tests can assert the args without
// sending a real signal.
var Kill = func(pid string) *exec.Cmd {
	return exec.Command("kill", "-s", "USR2", pid)
}

// Recipe is the Ghostty recipe: it carries the directories it resolves
// templates and artifacts from and runs the locate-render-write-reload
// procedure against a Theme.
type Recipe struct {
	themesDir string
	dataDir   string
}

// New builds a Ghostty recipe that reads templates from themesDir and
// writes artifacts under dataDir.
func New(themesDir, dataDir string) *Recipe {
	return &Recipe{themesDir: themesDir, dataDir: dataDir}
}

// Run renders the Ghostty template against t.Flavor, writes the artifact,
// and reloads Ghostty. A failure at any step is wrapped with this
// layer's role prefix; the orchestrator aggregates errors across recipes.
func (r *Recipe) Run(t theme.Theme) error {
	tmplPath := filepath.Join(r.themesDir, t.Dirname, templateName)
	log.Info("Locating template", "app", "ghostty", "path", tmplPath)
	raw, err := os.ReadFile(tmplPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Info("Template not found; skipping", "app", "ghostty")
			return fmt.Errorf("ghostty: %w", recipe.ErrNotApplicable)
		}
		return fmt.Errorf("ghostty: read template %s: %w", tmplPath, err)
	}
	log.Info("Reading template", "app", "ghostty", "path", tmplPath)

	log.Info("Rendering template", "app", "ghostty")
	rendered, err := template.Render(string(raw), t.Flavor)
	if err != nil {
		return fmt.Errorf("ghostty: %w", err)
	}

	outPath := filepath.Join(r.dataDir, artifactName)
	log.Info("Writing artifact", "app", "ghostty", "path", outPath)
	if err := os.WriteFile(outPath, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("ghostty: write artifact %s: %w", outPath, err)
	}

	if err := reload(); err != nil {
		return fmt.Errorf("ghostty: reload: %w", err)
	}
	return nil
}

// reload discovers the Ghostty pid via pgrep and sends it SIGUSR2. A
// missing pgrep result (Ghostty not running) is treated as a no-op so
// applying a theme without the app open does not fail the run.
func reload() error {
	log.Info("Discovering ghostty pid via pgrep", "app", "ghostty")
	out, err := Pgrep().Output()
	if err != nil {
		// pgrep exits non-zero when no process matches; treat as not
		// running rather than aborting the recipe.
		log.Info("No ghostty process running; reload skipped", "app", "ghostty")
		return nil
	}
	pid := firstNonEmptyLine(string(out))
	if pid == "" {
		log.Info("No ghostty process running; reload skipped", "app", "ghostty")
		return nil
	}
	log.Info("Sending SIGUSR2 to ghostty", "app", "ghostty", "pid", pid)
	if err := Kill(pid).Run(); err != nil {
		return fmt.Errorf("kill -s USR2 %s: %w", pid, err)
	}
	return nil
}

// firstNonEmptyLine returns the trimmed first non-empty line of s, the
// line pgrep writes for the oldest matching process.
func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			return t
		}
	}
	return ""
}
