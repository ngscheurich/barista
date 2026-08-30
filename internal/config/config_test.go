package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/config"
)

// Load returns the default config when no config.toml exists.
func TestLoadReturnsDefaultsWhenFileMissing(t *testing.T) {
	dir := t.TempDir()

	cfg, err := config.Load(dir)

	require.NoError(t, err)
	assert.Equal(t, "☕", cfg.Icon)
}

// Load reads the icon from a present config.toml.
func TestLoadReadsIcon(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte("icon = \"🍵\""), 0o644))

	cfg, err := config.Load(dir)

	require.NoError(t, err)
	assert.Equal(t, "🍵", cfg.Icon)
}

// Load applies the default when the file exists but icon is absent.
func TestLoadDefaultsWhenIconAbsent(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(""), 0o644))

	cfg, err := config.Load(dir)

	require.NoError(t, err)
	assert.Equal(t, "☕", cfg.Icon)
}

// Load returns a parse error when the file is malformed.
func TestLoadMalformedTOML(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte("icon = [bad"), 0o644))

	_, err := config.Load(dir)

	require.Error(t, err)
}
