package fzf_test

import (
	"errors"
	"os"
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

// writeFlavorDir creates a flavors dir containing <dirname>/fzf.rc.mustache
// and returns the flavors dir path.
func writeFlavorDir(t *testing.T, dirname, tmpl string) string {
	t.Helper()
	dir := t.TempDir()
	flavorDir := filepath.Join(dir, dirname)
	require.NoError(t, os.MkdirAll(flavorDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(flavorDir, "fzf.rc.mustache"), []byte(tmpl), 0o644))
	return dir
}

// When the output file exists, Run appends the rendered template after a
// blank line, preserving the existing content.
func TestRunAppendsToExistingFile(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha",
		"# {{name}}\n--color=bg:{{palette.mantle}}\n")
	outFile := filepath.Join(t.TempDir(), "fzfrc")
	require.NoError(t, os.WriteFile(outFile,
		[]byte("# My fzf options\n--layout=reverse\n"), 0o644))

	r := fzf.New(flavorsDir, outFile)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
	got, err := os.ReadFile(outFile)
	require.NoError(t, err)
	want := "# My fzf options\n--layout=reverse\n\n" +
		"# Catppuccin Mocha\n--color=bg:#11111b\n"
	assert.Equal(t, want, string(got))
}

// When the output file does not exist, Run writes the rendered template
// as a new file.
func TestRunWritesNewFileWhenAbsent(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha", "# {{name}}\n")
	outFile := filepath.Join(t.TempDir(), "fzfrc")

	r := fzf.New(flavorsDir, outFile)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
	got, err := os.ReadFile(outFile)
	require.NoError(t, err)
	assert.Equal(t, "# Catppuccin Mocha\n", string(got))
}

// Run overwrites an existing file in place by appending, not by replacing.
func TestRunPreservesExistingContentOnAppend(t *testing.T) {
	flavorsDir := writeFlavorDir(t, "mocha", "--color=bg:{{palette.mantle}}\n")
	outFile := filepath.Join(t.TempDir(), "fzfrc")
	require.NoError(t, os.WriteFile(outFile, []byte("--foo=bar\n"), 0o644))

	r := fzf.New(flavorsDir, outFile)
	err := r.Run(sampleFlavor())

	require.NoError(t, err)
	got, err := os.ReadFile(outFile)
	require.NoError(t, err)
	assert.Contains(t, string(got), "--foo=bar", "existing content must survive the append")
	assert.Contains(t, string(got), "--color=bg:#11111b")
}

// A missing template is not-applicable rather than a failure: Run returns
// an error wrapping recipe.ErrNotApplicable, which the orchestrator treats
// as a skip.
func TestRunMissingTemplateIsNotApplicable(t *testing.T) {
	flavorsDir := t.TempDir() // no flavor dir written
	outFile := filepath.Join(t.TempDir(), "fzfrc")

	r := fzf.New(flavorsDir, outFile)
	err := r.Run(sampleFlavor())

	require.Error(t, err)
	assert.True(t, errors.Is(err, recipe.ErrNotApplicable),
		"missing template should wrap recipe.ErrNotApplicable; got %v", err)
}

// OutputFilePath returns $FZF_DEFAULT_OPTS_FILE when it is set.
func TestOutputFilePathHonorsEnvVar(t *testing.T) {
	t.Setenv("FZF_DEFAULT_OPTS_FILE", "/custom/path/fzfrc")
	t.Setenv("HOME", "/home/user")

	got, err := fzf.OutputFilePath()

	require.NoError(t, err)
	assert.Equal(t, "/custom/path/fzfrc", got)
}

// OutputFilePath falls back to $HOME/.fzfrc when FZF_DEFAULT_OPTS_FILE is
// unset.
func TestOutputFilePathFallsBackToHome(t *testing.T) {
	t.Setenv("FZF_DEFAULT_OPTS_FILE", "")
	t.Setenv("HOME", "/home/user")

	got, err := fzf.OutputFilePath()

	require.NoError(t, err)
	assert.Equal(t, "/home/user/.fzfrc", got)
}

// OutputFilePath errors when neither FZF_DEFAULT_OPTS_FILE nor HOME is set.
func TestOutputFilePathErrorsWhenNeitherSet(t *testing.T) {
	t.Setenv("FZF_DEFAULT_OPTS_FILE", "")
	t.Setenv("HOME", "")

	_, err := fzf.OutputFilePath()

	require.Error(t, err)
}
