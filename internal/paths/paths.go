// Package paths resolves the filesystem locations Barista reads from and
// writes to.
//
// Three directories matter: the config directory (where themes and app
// templates live), the themes directory (a subdirectory of the config
// directory), and the data directory (where recipes write their output).
// Each is resolved from an XDG environment variable with a home-relative
// fallback, selecting the first candidate that exists and is a directory.
package paths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrNoDir is returned when none of the candidate directories for a path
// exist on disk.
var ErrNoDir = errors.New("no existing directory found")

// ConfigDir returns the barista config directory: $XDG_CONFIG_HOME/barista,
// falling back to ~/.config/barista. The XDG variable is checked for an
// existing directory before the fallback is tried; the returned barista
// subdir itself is not required to exist.
func ConfigDir() (string, error) {
	base, err := resolve(os.Getenv("XDG_CONFIG_HOME"), homeConfigDir())
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(base, "barista"), nil
}

// ThemesDir returns the themes directory, located under the config
// directory at <config dir>/themes. The directory is not required to exist.
func ThemesDir() (string, error) {
	cfg, err := ConfigDir()
	if err != nil {
		return "", fmt.Errorf("themes dir: %w", err)
	}
	return filepath.Join(cfg, "themes"), nil
}

// DataDir returns the barista data directory: $XDG_DATA_HOME/barista, falling
// back to ~/.local/share/barista. The directory is not required to exist;
// use EnsureDataDir to create it.
func DataDir() (string, error) {
	base, err := resolve(os.Getenv("XDG_DATA_HOME"), homeDataDir())
	if err != nil {
		return "", fmt.Errorf("data dir: %w", err)
	}
	return filepath.Join(base, "barista"), nil
}

// EnsureDataDir resolves the data directory and creates it (with mkdir -p
// semantics) if it does not yet exist, returning the resulting path. It is
// called before any recipe runs so recipe writes never fail on a missing
// directory.
func EnsureDataDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", fmt.Errorf("ensure data dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("ensure data dir %s: %w", dir, err)
	}
	return dir, nil
}

// Join composes a path from the given parts using filepath.Join, which
// handles platform separators and cleaning. Prefer this over string
// concatenation with "/".
func Join(parts ...string) string {
	return filepath.Join(parts...)
}

// resolve picks the first of primary and fallback that exists and is a
// directory. An unset (empty) or non-existent primary falls through to the
// fallback; if neither exists it returns a wrapped ErrNoDir.
func resolve(primary, fallback string) (string, error) {
	for _, candidate := range []string{primary, fallback} {
		if candidate == "" {
			continue
		}
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if info.IsDir() {
			return candidate, nil
		}
	}
	return "", ErrNoDir
}

// homeConfigDir returns the ~/.config directory, built from $HOME.
func homeConfigDir() string {
	return filepath.Join(os.Getenv("HOME"), ".config")
}

// homeDataDir returns the ~/.local/share directory, built from $HOME.
func homeDataDir() string {
	return filepath.Join(os.Getenv("HOME"), ".local", "share")
}
