# 07: Zellij recipe and remove Gleam source

**What to build:** The Zellij recipe, completing the three-recipe roster, followed by the removal of the Gleam source now that the Go port is the implementation. `internal/recipe/zellij` locates `<flavors dir>/<theme>/zellij.kdl.mustache`, reads it, renders via `internal.template.Render`, creates `<config dir>/zellij/themes/` with `os.MkdirAll`, writes `barista.kdl` there, then reloads by running `touch <config dir>/zellij/config.kdl` via `os/exec` (call the `touch` binary directly). Reuse the command-construction seam from ticket 05. Wire the Zellij recipe into the `apply` command's recipe loop (now all three recipes in the aggregated-error slice). Then remove the Gleam implementation: `src/`, `gleam.toml`, `manifest.toml`, `build/`, `priv/` (including `priv/scripts/nvim_send.sh`, now superseded by `internal/nvim`), and any Erlang-specific entries from `.gitignore` that ticket 01 didn't already drop. The port is the implementation. Final integration test: one `barista apply <theme>` invocation against a `t.TempDir()` flavors dir holding all three templates runs all three recipes and writes all three output files; a recipe whose template is missing errors but does not block the others (continue-on-error, per the spec). Update `README.md` if its development commands (`gleam run` / `gleam test`) no longer apply — replace with `go run ./cmd/barista apply <theme>` and `go test ./...`.

**Blocked by:** 06 (Neovim recipe with Go socket discovery)

**Status:** ready-for-agent

- [ ] `internal/recipe/zellij` locates, reads, renders, and writes `barista.kdl` to `<config dir>/zellij/themes/` (creating the dir)
- [ ] Zellij reload runs `touch <config dir>/zellij/config.kdl` via `os/exec` (no `sh -c`)
- [ ] Zellij recipe wired into the `apply` loop (all three recipes in the aggregated-error slice)
- [ ] Gleam source removed (`src/`, `gleam.toml`, `manifest.toml`, `build/`, `priv/`)
- [ ] Any remaining Erlang-specific `.gitignore` entries removed
- [ ] `README.md` development commands updated to `go run` / `go test`
- [ ] Integration test: all three recipes run from one `apply` invocation; a missing-template recipe errors without blocking the others
- [ ] `go build ./...`, `go test ./...`, `go vet ./...` pass
