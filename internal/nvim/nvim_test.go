package nvim_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/nvim"
)

// DiscoverSockets returns only the entries whose names match the
// nvim.*.0 contract one subdirectory level deep, ignoring non-matching
// files and directories.
func TestDiscoverSocketsReturnsOnlyMatchingEntries(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", dir)

	subdir := filepath.Join(dir, "session")
	require.NoError(t, os.MkdirAll(subdir, 0o755))

	// Matching sockets.
	mustWrite(t, filepath.Join(subdir, "nvim.123.0"))
	mustWrite(t, filepath.Join(subdir, "nvim.456.0"))
	// Non-matching: wrong prefix, wrong suffix, a directory, a log file.
	mustWrite(t, filepath.Join(subdir, "other.0"))
	mustWrite(t, filepath.Join(subdir, "nvim.789.log"))
	require.NoError(t, os.Mkdir(filepath.Join(subdir, "nvim.dir.0"), 0o755))
	mustWrite(t, filepath.Join(subdir, "README"))

	got, err := nvim.DiscoverSockets()

	require.NoError(t, err)
	want := []string{
		filepath.Join(subdir, "nvim.123.0"),
		filepath.Join(subdir, "nvim.456.0"),
	}
	assert.ElementsMatch(t, want, got)
}

// When XDG_RUNTIME_DIR is set but points at a non-existent path, no
// sockets are found and no error is returned -- the recipe treats this as
// "Neovim is not running" rather than aborting.
func TestDiscoverSocketsMissingDirReturnsEmpty(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", filepath.Join(t.TempDir(), "does-not-exist"))

	got, err := nvim.DiscoverSockets()

	require.NoError(t, err)
	assert.Empty(t, got)
}

// When XDG_RUNTIME_DIR is unset, the fallback path $TMPDIR/nvim.<user>
// is walked, one subdirectory level deep.
func TestDiscoverSocketsFallsBackWhenXDGUnset(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("TMPDIR", tmp)
	t.Setenv("USER", "tester")

	fallback := filepath.Join(tmp, "nvim.tester")
	session := filepath.Join(fallback, "session")
	require.NoError(t, os.MkdirAll(session, 0o755))
	mustWrite(t, filepath.Join(session, "nvim.999.0"))

	got, err := nvim.DiscoverSockets()

	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(session, "nvim.999.0")}, got)
}

// RemoteSend builds nvim --server <socket> --remote-send <keys>, calling
// the nvim binary directly rather than going through sh -c.
func TestRemoteSendCommandShape(t *testing.T) {
	cmd := nvim.RemoteSend("/tmp/nvim.1.0", "<Cmd>lua require('barista')<CR>")

	assert.Equal(t, "nvim", filepath.Base(cmd.Path))
	assert.Equal(t,
		[]string{"--server", "/tmp/nvim.1.0", "--remote-send", "<Cmd>lua require('barista')<CR>"},
		cmd.Args[1:])
}

// Reload discovers Neovim's per-instance socket beneath the fallback
// runtime directory and asks the instance to discard the cached Barista
// module, load the newly written artifact, and apply it. A bare require is
// insufficient because Lua returns the module cached by the initial setup.
func TestReloadLoadsAndAppliesNewArtifact(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("TMPDIR", tmp)
	t.Setenv("USER", "tester")

	socket := filepath.Join(tmp, "nvim.tester", "9Hrtkz", "nvim.41727.0")
	require.NoError(t, os.MkdirAll(filepath.Dir(socket), 0o755))
	mustWrite(t, socket)

	binDir := filepath.Join(tmp, "bin")
	require.NoError(t, os.Mkdir(binDir, 0o755))
	argsFile := filepath.Join(tmp, "nvim-args")
	fakeNvim := filepath.Join(binDir, "nvim")
	require.NoError(t, os.WriteFile(fakeNvim, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$NVIM_ARGS_FILE\"\n"), 0o755))
	t.Setenv("NVIM_ARGS_FILE", argsFile)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	require.NoError(t, nvim.Reload())

	rawArgs, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	got := strings.Split(strings.TrimSpace(string(rawArgs)), "\n")
	assert.Equal(t, []string{
		"--server",
		socket,
		"--remote-send",
		"<Cmd>lua package.loaded['barista'] = nil; require('barista').setup()<CR>",
	}, got)
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte{}, 0o644))
}
