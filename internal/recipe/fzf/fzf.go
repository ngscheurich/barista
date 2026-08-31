// Package fzf implements the recipe for theming fzf: locate fzf.mustache
// under the flavor's directory, render it against the flavor, and set
// FZF_DEFAULT_OPTS via fish with the theme's --color flags replacing any
// existing ones.
//
// fzf reads its colors from the FZF_DEFAULT_OPTS environment variable.
// The recipe reads the current value, strips every --color flag from it,
// extracts the --color flags from the rendered template (discarding any
// non-color lines), and prepends the theme colors to the remaining user
// options. The result is set as a fish universal export so it persists
// across sessions.
package fzf

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"charm.land/log/v2"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/template"
)

// templateName is the Mustache file a flavor directory carries for fzf;
// the recipe looks for it under <flavorsDir>/<flavor.Dirname>.
const templateName = "fzf.mustache"

// envVar is the environment variable fzf reads its options from.
const envVar = "FZF_DEFAULT_OPTS"

// colorFlag is the prefix every fzf color option shares.
const colorFlag = "--color"

// Fish constructs the fish command used to set FZF_DEFAULT_OPTS as a
// universal export. The merged value is wrapped in single quotes so fish
// treats it as a literal string, preserving any double quotes and inner
// spaces in option values. Single quotes inside the value are escaped
// fish-style, each becoming a backslash-escaped quote. It returns the
// command without running it, so tests can assert the args and
// integration code can swap in a fake.
var Fish = func(merged string) *exec.Cmd {
	escaped := strings.ReplaceAll(merged, "'", "'\\''")
	return exec.Command("fish", "-c", "set -Ux "+envVar+" '"+escaped+"'")
}

// Recipe is the fzf recipe: it carries the flavors directory it reads
// templates from.
type Recipe struct {
	flavorsDir string
}

// New builds an fzf recipe that reads templates from flavorsDir.
func New(flavorsDir string) *Recipe {
	return &Recipe{flavorsDir: flavorsDir}
}

// Run renders the fzf template against f, replaces the --color flags in
// the existing $FZF_DEFAULT_OPTS with the template's color flags, and
// runs fish to set the variable as a universal export. A missing
// template reports recipe.ErrNotApplicable so the orchestrator marks fzf
// skipped. A failure at any step is wrapped with this layer's role
// prefix; the orchestrator aggregates errors across recipes.
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

	existing := os.Getenv(envVar)
	log.Info("Read existing env var", "app", "fzf", "var", envVar, "value", existing)

	merged := MergeOpts(rendered, existing)
	log.Info("Merged options", "app", "fzf", "value", merged)

	log.Info("Setting env var via fish", "app", "fzf")
	if err := Fish(merged).Run(); err != nil {
		return fmt.Errorf("fzf: set %s via fish: %w", envVar, err)
	}
	return nil
}

// MergeOpts builds the new FZF_DEFAULT_OPTS value from the rendered
// template and the existing env var value. It extracts the --color
// segments from the rendered template (discarding any segment that does
// not start with --color), strips every --color segment from the
// existing value, and prepends the theme colors to the remaining user
// options. Quoted values (single or double) are preserved as a single
// token with their quotes intact.
func MergeOpts(rendered, existing string) string {
	colors := filterColors(rendered)
	rest := stripColors(existing)
	merged := strings.Join(colors, " ")
	if len(rest) > 0 {
		if merged != "" {
			merged += " "
		}
		merged += strings.Join(rest, " ")
	}
	return merged
}

// filterColors returns the tokens of s that start with the --color flag,
// discarding everything else. Quoted values are kept as a single token
// with their quotes intact.
func filterColors(s string) []string {
	var out []string
	for _, field := range tokenize(s) {
		if strings.HasPrefix(field, colorFlag) {
			out = append(out, field)
		}
	}
	return out
}

// stripColors returns the tokens of s that do not start with the --color
// flag, removing every color option. Quoted values are kept as a single
// token with their quotes intact.
func stripColors(s string) []string {
	var out []string
	for _, field := range tokenize(s) {
		if !strings.HasPrefix(field, colorFlag) {
			out = append(out, field)
		}
	}
	return out
}

// tokenize splits s into tokens on whitespace outside single or double
// quotes. Quoted spans (including the quote characters) are kept as part
// of the containing token, so a value like --key="  value " survives as
// a single token with its quotes and inner spaces intact. An unclosed
// quote runs to the end of the string.
func tokenize(s string) []string {
	var tokens []string
	var cur strings.Builder
	inSingle, inDouble, inToken := false, false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inSingle:
			cur.WriteByte(c)
			if c == '\'' {
				inSingle = false
			}
		case inDouble:
			cur.WriteByte(c)
			if c == '"' {
				inDouble = false
			}
		case c == '\'':
			inToken = true
			inSingle = true
			cur.WriteByte(c)
		case c == '"':
			inToken = true
			inDouble = true
			cur.WriteByte(c)
		case c == ' ' || c == '\t' || c == '\n':
			if inToken {
				tokens = append(tokens, cur.String())
				cur.Reset()
				inToken = false
			}
		default:
			inToken = true
			cur.WriteByte(c)
		}
	}
	if inToken {
		tokens = append(tokens, cur.String())
	}
	return tokens
}
