package fzf_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/recipe"
	"github.com/ngscheurich/barista/internal/recipe/fzf"
)

// sampleFlavor carries distinct palette values so the rendered output
// makes it obvious which field landed where.
func sampleFlavor() flavor.Flavor {
	return flavor.Flavor{
		Name:    "Catppuccin Mocha",
		Dirname: "mocha",
		Palette: flavor.Palette{
			Mantle: "#11111b",
		},
	}
}

// writeFlavorDir creates a flavors dir containing <dirname>/fzf.mustache
// and returns the flavors dir path.
func writeFlavorDir(t *testing.T, dirname, tmpl string) string {
	t.Helper()
	dir := t.TempDir()
	flavorDir := filepath.Join(dir, dirname)
	require.NoError(t, os.MkdirAll(flavorDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "fzf.mustache"), []byte(tmpl), 0o644))
	return dir
}

// fakeFish replaces fzf.Fish with a no-op that succeeds, so tests do not
// spawn a real fish process. It restores the original on cleanup.
func fakeFish(t *testing.T) {
	t.Helper()
	orig := fzf.Fish
	t.Cleanup(func() { fzf.Fish = orig })
	fzf.Fish = func(string) *exec.Cmd { return exec.Command("true") }
}

// fakeFishCapturing replaces fzf.Fish with a no-op that captures the
// merged string it was called with, so a test can assert on it.
func fakeFishCapturing(t *testing.T) *string {
	t.Helper()
	orig := fzf.Fish
	t.Cleanup(func() { fzf.Fish = orig })
	var captured string
	fzf.Fish = func(merged string) *exec.Cmd {
		captured = merged
		return exec.Command("true")
	}
	return &captured
}

// Run renders the template, replaces color flags in the existing env
// var, and runs fish to set the variable.
func TestRunSetsEnvVarViaFish(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha", "--color=bg:{{palette.mantle}}\n")
	t.Setenv("FZF_DEFAULT_OPTS", "--layout=inline --color=bg:#000000")
	fakeFish(t)

	r := fzf.New(flavorsDir)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
}

// Run passes the merged value (theme colors replacing existing colors)
// to fish.
func TestRunPassesMergedValueToFish(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha",
		"--color=bg:{{palette.mantle}}\n--color=fg:#fefefe\n")
	t.Setenv("FZF_DEFAULT_OPTS", "--layout=inline --color=bg:#000000 --info=true")
	captured := fakeFishCapturing(t)

	r := fzf.New(flavorsDir)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
	assert.Equal(t,
		"--color=bg:#11111b --color=fg:#fefefe --layout=inline --info=true",
		*captured,
	)
}

// A missing template is not-applicable rather than a failure.
func TestRunMissingTemplateIsNotApplicable(t *testing.T) {
	flavorsDir := t.TempDir()
	fakeFish(t)

	r := fzf.New(flavorsDir)
	err := r.Run(sampleFlavor())

	require.Error(t, err)
	assert.True(t, errors.Is(err, recipe.ErrNotApplicable),
		"missing template should wrap recipe.ErrNotApplicable; got %v", err)
}

// The fish seam: Fish builds fish -c "set -Ux FZF_DEFAULT_OPTS '<merged>'",
// wrapping the value in single quotes so inner double quotes and spaces
// survive fish's parser. Single quotes in the value are escaped fish-style.
func TestFishCommandShape(t *testing.T) {
	cmd := fzf.Fish(`--color=bg:#11111b --prompt=": " --layout=inline`)

	assert.Equal(t, "fish", filepath.Base(cmd.Path))
	assert.Equal(t,
		[]string{"-c", `set -Ux FZF_DEFAULT_OPTS '--color=bg:#11111b --prompt=": " --layout=inline'`},
		cmd.Args[1:])
}

// The fish seam escapes single quotes in the value fish-style.
func TestFishCommandShapeEscapesSingleQuotes(t *testing.T) {
	cmd := fzf.Fish(`--prompt=': ' --layout=reverse`)

	assert.Equal(t,
		[]string{"-c", `set -Ux FZF_DEFAULT_OPTS '--prompt='\'': '\'' --layout=reverse'`},
		cmd.Args[1:])
}

// MergeOpts replaces existing color flags with the template's colors and
// leaves non-color options alone.
func TestMergeOptsReplacesColors(t *testing.T) {
	got := fzf.MergeOpts(
		"--color=bg:1\n--color=fg:0\n",
		"--layout=inline --color=bg:#000000 --info=true",
	)
	assert.Equal(t,
		"--color=bg:1 --color=fg:0 --layout=inline --info=true",
		got)
}

// MergeOpts discards template lines that do not start with --color.
func TestMergeOptsDiscardsNonColorTemplateLines(t *testing.T) {
	got := fzf.MergeOpts(
		"# this is a comment\n--color=bg:1\n--layout=reverse\n",
		"--info=true",
	)
	assert.Equal(t, "--color=bg:1 --info=true", got)
}

// MergeOpts with no existing color flags just prepends the theme colors.
func TestMergeOptsNoExistingColors(t *testing.T) {
	got := fzf.MergeOpts("--color=bg:1 --color=fg:0", "--layout=inline")
	assert.Equal(t, "--color=bg:1 --color=fg:0 --layout=inline", got)
}

// MergeOpts with an empty existing value produces only the theme colors.
func TestMergeOptsEmptyExisting(t *testing.T) {
	got := fzf.MergeOpts("--color=bg:1 --color=fg:0", "")
	assert.Equal(t, "--color=bg:1 --color=fg:0", got)
}

// MergeOpts with no color flags in the template leaves the existing
// value unchanged.
func TestMergeOptsNoTemplateColors(t *testing.T) {
	got := fzf.MergeOpts("# comment\n--layout=reverse", "--info=true --color=bg:#000000")
	assert.Equal(t, "--info=true", got)
}

// MergeOpts is idempotent: applying it twice with the same template
// produces the same result.
func TestMergeOptsIdempotent(t *testing.T) {
	tmpl := "--color=bg:1\n--color=fg:0\n"
	existing := "--layout=inline --color=bg:#000000 --info=true"

	first := fzf.MergeOpts(tmpl, existing)
	second := fzf.MergeOpts(tmpl, first)

	assert.Equal(t, first, second)
}

// MergeOpts preserves double-quoted values containing spaces.
func TestMergeOptsPreservesDoubleQuotedSpaces(t *testing.T) {
	got := fzf.MergeOpts(
		"--color=bg:1",
		"--key=\"  value \" --layout=inline",
	)
	assert.Equal(t, "--color=bg:1 --key=\"  value \" --layout=inline", got)
}

// MergeOpts preserves single-quoted values containing spaces.
func TestMergeOptsPreservesSingleQuotedSpaces(t *testing.T) {
	got := fzf.MergeOpts(
		"--color=bg:1",
		"--key='va l u e' --layout=inline",
	)
	assert.Equal(t, "--color=bg:1 --key='va l u e' --layout=inline", got)
}

// MergeOpts preserves quoted color flags and strips them correctly even
// with inner spaces.
func TestMergeOptsStripsQuotedColorFlags(t *testing.T) {
	got := fzf.MergeOpts(
		"--color=bg:1",
		"--color=\"  value \" --layout=inline",
	)
	assert.Equal(t, "--color=bg:1 --layout=inline", got)
}
