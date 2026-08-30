// Package fzf implements the recipe for theming fzf: locate fzf.rc.mustache
// under the flavor's directory, render it against the flavor, and merge the
// result into the user's fzf opts file.
//
// fzf reads its colors from FZF_DEFAULT_OPTS (or a file the user sources),
// so there is no process to signal after the write; the new theme takes
// effect on the next sourcing. The output file is $FZF_DEFAULT_OPTS_FILE,
// falling back to $HOME/.fzfrc. When the file already exists the rendered
// template is appended to it (separated by a blank line); when it does
// not, the rendered template is written as a new file.
package fzf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/template"
)

// templateName is the Mustache file a flavor directory carries for fzf;
// the recipe looks for it under <flavorsDir>/<flavor.Dirname>.
const templateName = "fzf.rc.mustache"

// OutputFilePath resolves the file the rendered template is written to:
// $FZF_DEFAULT_OPTS_FILE when set, otherwise $HOME/.fzfrc. An empty
// FZF_DEFAULT_OPTS_FILE is treated as unset and falls through to the
// home fallback.
func OutputFilePath() (string, error) {
	if p := os.Getenv("FZF_DEFAULT_OPTS_FILE"); p != "" {
		log.Info("FZF_DEFAULT_OPTS_FILE is set", "path", p)
		return p, nil
	}
	log.Info("FZF_DEFAULT_OPTS_FILE is unset; falling back to $HOME/.fzfrc")
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("fzf: resolve output file: %w", errNoOutputPath)
	}
	path := filepath.Join(home, ".fzfrc")
	log.Info("Resolved output file", "path", path)
	return path, nil
}

// errNoOutputPath is returned when neither FZF_DEFAULT_OPTS_FILE nor HOME
// is set, so there is nowhere to write the rendered template.
var errNoOutputPath = errors.New("neither FZF_DEFAULT_OPTS_FILE nor HOME is set")

// Recipe is the fzf recipe: it carries the flavors directory it reads
// templates from and the file path it writes the rendered result to.
type Recipe struct {
	flavorsDir  string
	outputFile  string
}

// New builds an fzf recipe that reads templates from flavorsDir and writes
// the rendered theme to outputFile. Use OutputFilePath to resolve
// outputFile from the environment.
func New(flavorsDir, outputFile string) *Recipe {
	return &Recipe{flavorsDir: flavorsDir, outputFile: outputFile}
}

// Run renders the fzf template against f and merges it into the output
// file: when the file exists the rendered text is appended after a blank
// line, preserving the existing content; when it does not the rendered
// text is written as a new file. A missing template reports
// recipe.ErrNotApplicable so the orchestrator marks fzf skipped. A
// failure at any step is wrapped with this layer's role prefix; the
// orchestrator aggregates errors across recipes.
func (r *Recipe) Run(f flavor.Flavor) error {
	tmplPath := filepath.Join(r.flavorsDir, f.Dirname, templateName)
	log.Info("Locating template", "app", "fzf", "path", tmplPath)
	raw, err := os.ReadFile(tmplPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Info("Template not found; skipping", "app", "fzf")
			return fmt.Errorf("fzf: %w", recipe.ErrNotApplicable)
		}
		return fmt.Errorf("fzf: read template %s: %w", tmplPath, err)
	}
	log.Info("Reading template", "app", "fzf", "path", tmplPath)

	log.Info("Rendering template", "app", "fzf")
	rendered, err := template.Render(string(raw), f)
	if err != nil {
		return fmt.Errorf("fzf: %w", err)
	}
	rendered = ensureTrailingNewline(rendered)

	existing, err := os.ReadFile(r.outputFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Info("Output file does not exist; writing new file", "app", "fzf", "path", r.outputFile)
			if err := os.WriteFile(r.outputFile, []byte(rendered), 0o644); err != nil {
				return fmt.Errorf("fzf: write %s: %w", r.outputFile, err)
			}
			return nil
		}
		return fmt.Errorf("fzf: read %s: %w", r.outputFile, err)
	}

	log.Info("Appending to existing file", "app", "fzf", "path", r.outputFile)
	content := ensureTrailingNewline(string(existing)) + "\n" + rendered
	if err := os.WriteFile(r.outputFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("fzf: write %s: %w", r.outputFile, err)
	}
	return nil
}

// ensureTrailingNewline returns s with exactly one trailing newline,
// stripping any trailing whitespace first so the separator blank line is
// the only blank line between the existing and appended content.
func ensureTrailingNewline(s string) string {
	return strings.TrimRight(s, "\n") + "\n"
}
