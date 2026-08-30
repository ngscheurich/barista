// Package config reads the user config file from the barista config
// directory, returning a Config with defaults for any absent field.
//
// The config lives at <config dir>/config.toml. When the file is missing
// or empty, a zero-value Config is returned with all defaults applied;
// only malformed TOML returns an error.
package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const defaultIcon = "☕"

// Config holds user-facing settings read from config.toml.
type Config struct {
	// Icon is the glyph printed before the "Served up <name>" message.
	// Defaults to ☕ when absent.
	Icon string
}

// Load reads the barista config file from the given config directory.
// A missing file yields a default Config (no error). A present but
// malformed file returns a parse error.
func Load(configDir string) (Config, error) {
	cfg := Config{Icon: defaultIcon}

	path := filepath.Join(configDir, "config.toml")
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	var file configFile
	if err := toml.Unmarshal(raw, &file); err != nil {
		return cfg, err
	}
	if file.Icon != "" {
		cfg.Icon = file.Icon
	}
	return cfg, nil
}

// configFile mirrors the on-disk config.toml shape.
type configFile struct {
	Icon string `toml:"icon"`
}
