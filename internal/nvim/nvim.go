// Package nvim discovers running Neovim server sockets and sends keys to
// them, replacing the Gleam version's priv/scripts/nvim_send.sh helper so
// the Go binary stays self-contained.
//
// The contract with Neovim's server discovery is two details that must be
// preserved exactly: sockets are named nvim.<pid>.0 (the nvim.*.0 glob),
// and they live under $XDG_RUNTIME_DIR with a fallback of
// $TMPDIR/nvim.<user> when that variable is unset.
package nvim

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

// DiscoverSockets walks the Neovim runtime directory for server sockets,
// returning the paths of every entry named nvim.*.0. It tries
// $XDG_RUNTIME_DIR first, then falls back to $TMPDIR/nvim.<user>. A
// missing directory yields no sockets rather than an error, so a recipe
// can treat "Neovim is not running" as a no-op reload.
func DiscoverSockets() ([]string, error) {
	dir := resolveRuntimeDir()
	log.Info("Scanning neovim runtime directory", "dir", dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read nvim runtime dir %s: %w", dir, err)
	}

	var sockets []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "nvim.") || !strings.HasSuffix(name, ".0") {
			continue
		}
		if e.IsDir() {
			continue
		}
		sockets = append(sockets, filepath.Join(dir, name))
	}
	log.Info("Discovered neovim sockets", "count", len(sockets))
	return sockets, nil
}

// RemoteSend constructs the nvim --server <socket> --remote-send <keys>
// command without running it, so tests can assert the args and the recipe
// can run it once per discovered socket.
func RemoteSend(socket, keys string) *exec.Cmd {
	return exec.Command("nvim", "--server", socket, "--remote-send", keys)
}

// reloadKeys is the keystroke sequence sent to every Neovim instance to
// reload the Barista flavor plugin, matching the Gleam version's
// nvim_send.sh argument.
const reloadKeys = "<Cmd>lua require('barista')<CR>"

// Reload sends the reload keystroke to every discovered Neovim socket.
// Errors from individual sockets are collected and joined so one dead
// socket does not hide a failure on another; an empty socket list is a
// no-op.
func Reload() error {
	sockets, err := DiscoverSockets()
	if err != nil {
		return fmt.Errorf("discover sockets: %w", err)
	}

	var errs []error
	for _, s := range sockets {
		log.Info("Sending reload keystroke to neovim socket", "socket", s)
		if err := RemoteSend(s, reloadKeys).Run(); err != nil {
			errs = append(errs, fmt.Errorf("nvim --server %s: %w", s, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("reload neovim: %w", errors.Join(errs...))
	}
	return nil
}

// resolveRuntimeDir picks the directory to scan for Neovim sockets:
// $XDG_RUNTIME_DIR when set, otherwise $TMPDIR/nvim.<user>.
func resolveRuntimeDir() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(os.TempDir(), "nvim."+os.Getenv("USER"))
}
