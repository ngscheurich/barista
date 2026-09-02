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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"charm.land/log/v2"
)

// DiscoverSockets discovers running Neovim server sockets, returning the
// paths of every entry matching nvim.*.0 one subdirectory level deep
// under the runtime directory (e.g. $XDG_RUNTIME_DIR/*/nvim.*.0). It
// tries $XDG_RUNTIME_DIR first, then falls back to $TMPDIR/nvim.<user>.
// A missing directory yields no sockets rather than an error, so a
// recipe can treat "Neovim is not running" as a no-op reload.
func DiscoverSockets() ([]string, error) {
	dir := resolveRuntimeDir()
	log.Info("Scanning neovim runtime directory", "dir", dir)

	pattern := filepath.Join(dir, "*", "nvim.*.0")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob %s: %w", pattern, err)
	}

	// Filter out any directory entries that matched the glob.
	var sockets []string
	for _, m := range matches {
		info, err := os.Lstat(m)
		if err != nil || info.IsDir() {
			continue
		}
		sockets = append(sockets, m)
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

// reloadKeys is the keystroke sequence sent to every Neovim instance.
// Lua's require caches modules, so discard the Barista module loaded at
// startup before requiring the newly written artifact and applying it.
const reloadKeys = "<Cmd>lua package.loaded['barista'] = nil; require('barista').setup()<CR>"

// processAlive returns true when pid is a running process.
// Uses syscall.Kill(pid, 0) which probes without delivering a signal.
func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

// Reload sends the reload keystroke to every discovered Neovim socket.
// Stale sockets (whose PID is no longer alive) are skipped rather than
// treated as failures; an empty socket list is a no-op.
func Reload() error {
	sockets, err := DiscoverSockets()
	if err != nil {
		return fmt.Errorf("discover sockets: %w", err)
	}

	for _, s := range sockets {
		log.Info("Sending reload keystroke to neovim socket", "socket", s)
		if err := RemoteSend(s, reloadKeys).Run(); err != nil {
			if pid := extractPID(s); pid > 0 && !processAlive(pid) {
				log.Info("Skipping stale neovim socket", "socket", s, "pid", pid)
				continue
			}
			return fmt.Errorf("reload neovim: nvim --server %s: %w", s, err)
		}
	}
	return nil
}

// extractPID parses the PID from a socket path like
// ".../nvim.25015.0", returning -1 when the filename doesn't match.
func extractPID(path string) int {
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "nvim.") || !strings.HasSuffix(base, ".0") {
		return -1
	}
	pidStr := strings.TrimSuffix(strings.TrimPrefix(base, "nvim."), ".0")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return -1
	}
	return pid
}

// resolveRuntimeDir picks the directory to scan for Neovim sockets:
// $XDG_RUNTIME_DIR when set, otherwise $TMPDIR/nvim.<user>.
func resolveRuntimeDir() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(os.TempDir(), "nvim."+os.Getenv("USER"))
}
