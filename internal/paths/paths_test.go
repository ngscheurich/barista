package paths_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/paths"
)

// Resolution honours the XDG env var when it points at an existing directory,
// returning the barista subdir under it (the subdir itself need not exist).
func TestConfigDirUsesXDGConfigHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	got, err := paths.ConfigDir()

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, "barista"), got)
}

// An unset XDG var reads as the empty string and falls back to ~/.config.
func TestConfigDirFallsBackWhenXDGUnset(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".config"), 0o755))
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	got, err := paths.ConfigDir()

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "barista"), got)
}

// A set-but-non-existent XDG path falls back to the home candidate.
func TestConfigDirFallsBackWhenXDGNonexistent(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".config"), 0o755))
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "does-not-exist"))

	got, err := paths.ConfigDir()

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "barista"), got)
}

// When neither candidate exists, resolution returns a wrapped error.
func TestConfigDirErrorsWhenNeitherExists(t *testing.T) {
	home := t.TempDir() // no .config created
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	_, err := paths.ConfigDir()

	assert.Error(t, err)
}

// ThemesDir is the themes/ subdirectory of the resolved config dir.
func TestThemesDirIsUnderConfigDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	got, err := paths.ThemesDir()

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, "barista", "themes"), got)
}

func TestDataDirUsesXDGDataHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	got, err := paths.DataDir()

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, "barista"), got)
}

func TestDataDirFallsBackWhenXDGUnset(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".local", "share"), 0o755))
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	got, err := paths.DataDir()

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".local", "share", "barista"), got)
}

func TestDataDirFallsBackWhenXDGNonexistent(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".local", "share"), 0o755))
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "does-not-exist"))

	got, err := paths.DataDir()

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".local", "share", "barista"), got)
}

// EnsureDataDir creates the barista data directory when it does not yet
// exist, with mkdir -p semantics, and returns the resulting path.
func TestEnsureDataDirCreatesNestedPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	target := filepath.Join(tmp, "barista")
	assert.NoDirExists(t, target)

	got, err := paths.EnsureDataDir()

	assert.NoError(t, err)
	assert.Equal(t, target, got)
	assert.DirExists(t, target)
}

// EnsureDataDir is idempotent: calling it again on an existing dir succeeds.
func TestEnsureDataDirIsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "barista"), 0o755))

	got, err := paths.EnsureDataDir()

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, "barista"), got)
}

// Join composes paths via filepath.Join, cleaning separators.
func TestJoin(t *testing.T) {
	assert.Equal(t, filepath.Join("a", "b", "c"), paths.Join("a", "b", "c"))
	assert.Equal(t, "a", paths.Join("a"))
}
