package nvim_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/nvim"
)

// DiscoverSockets returns only the entries whose names match the
// nvim.*.0 contract, ignoring non-matching files and directories.
func TestDiscoverSocketsReturnsOnlyMatchingEntries(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", dir)

	// Matching sockets.
	mustWrite(t, filepath.Join(dir, "nvim.123.0"))
	mustWrite(t, filepath.Join(dir, "nvim.456.0"))
	// Non-matching: wrong prefix, wrong suffix, a directory, a log file.
	mustWrite(t, filepath.Join(dir, "other.0"))
	mustWrite(t, filepath.Join(dir, "nvim.789.log"))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "nvim.dir.0"), 0o755))
	mustWrite(t, filepath.Join(dir, "README"))

	got, err := nvim.DiscoverSockets()

	require.NoError(t, err)
	want := []string{
		filepath.Join(dir, "nvim.123.0"),
		filepath.Join(dir, "nvim.456.0"),
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
// is walked.
func TestDiscoverSocketsFallsBackWhenXDGUnset(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("TMPDIR", tmp)
	t.Setenv("USER", "tester")

	fallback := filepath.Join(tmp, "nvim.tester")
	require.NoError(t, os.MkdirAll(fallback, 0o755))
	mustWrite(t, filepath.Join(fallback, "nvim.999.0"))

	got, err := nvim.DiscoverSockets()

	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(fallback, "nvim.999.0")}, got)
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

func mustWrite(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte{}, 0o644))
}
