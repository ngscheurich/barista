# 06: Neovim recipe with Go socket discovery

**What to build:** The `internal/nvim` package and the `internal/recipe/neovim` implementation, completing the second recipe and dropping the `priv/scripts/nvim_send.sh` shell dependency. `internal/nvim.DiscoverSockets()` walks `$XDG_RUNTIME_DIR` (fallback `$TMPDIR/nvim.<user>`) for entries whose names start with `nvim.` and end with `.0`, returning their paths — the `nvim.*.0` glob and the fallback path are the contract with Neovim's server discovery and must be preserved exactly. The recipe locates `<flavors dir>/<theme>/neovim.lua.mustache`, reads it, renders via `internal.template.Render`, creates `<data dir>/nvim/lua/` with `os.MkdirAll`, writes `flavor.lua` there, then reloads by calling `nvim --server <socket> --remote-send "<Cmd>lua require('barista')<CR>"` on every discovered socket via `os/exec` (call the `nvim` binary directly, never `sh -c`). Reuse the command-construction seam from ticket 05 for testability. Wire the Neovim recipe into the `apply` command's recipe loop (now two recipes in the aggregated-error slice). Unit-test `DiscoverSockets` against a `t.TempDir()` with fake `nvim.*.0` files and non-matching entries, using `t.Setenv` for `XDG_RUNTIME_DIR` and `TMPDIR`; assert only matching sockets are returned and the fallback path is used when `XDG_RUNTIME_DIR` is unset. Test the reload step via the command-construction seam. With this ticket, `priv/scripts/nvim_send.sh` is no longer referenced by the Go code (the file itself is removed in ticket 07 with the rest of the Gleam source).

**Blocked by:** 05 (Ghostty recipe — tracer bullet through the CLI)

**Status:** ready-for-agent

- [ ] `internal/nvim.DiscoverSockets()` walks `$XDG_RUNTIME_DIR` (fallback `$TMPDIR/nvim.<user>`) for `nvim.*.0` entries
- [ ] The `nvim.*.0` glob and fallback path preserved exactly (Neovim server-discovery contract)
- [ ] `internal/recipe/neovim` locates, reads, renders, and writes `flavor.lua` to `<data dir>/nvim/lua/` (creating the dir)
- [ ] Reload calls `nvim --server <socket> --remote-send` on every discovered socket via `os/exec` (no `sh -c`)
- [ ] Neovim recipe wired into the `apply` loop (two recipes in the aggregated-error slice)
- [ ] `DiscoverSockets` unit-tested with a temp dir, fake socket files, `t.Setenv`, and the fallback case
- [ ] Reload tested via the command-construction seam
- [ ] `go build ./...`, `go test ./...`, `go vet ./...` pass
