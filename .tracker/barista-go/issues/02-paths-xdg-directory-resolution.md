# 02: Paths — XDG directory resolution

**What to build:** The `internal/paths` package, the foundation every recipe and the flavor loader depend on. Resolve three directories per the spec's directory table: the config dir (`$XDG_CONFIG_HOME/barista`, fallback `~/.config/barista`), the flavors dir (`<config dir>/flavors`), and the data dir (`$XDG_DATA_HOME/barista`, fallback `~/.local/share/barista`). Resolution picks the first candidate in `[primary, fallback]` that exists and is a directory; if an `XDG_*` var is set but points at a non-existent path, the fallback is used; if neither exists, the relevant function returns a wrapped error. Env vars that are unset read as the empty string (replicated from the Gleam version). Provide a filepath-join helper built on `filepath.Join` (not string concatenation). Provide an "ensure the data dir exists" helper that creates it with `os.MkdirAll` (the `mkdir -p` semantics the Gleam version shelled out for). Unit-test against `testing.T.TempDir()` with explicit env var manipulation (`t.Setenv`): set `XDG_CONFIG_HOME`/`XDG_DATA_HOME` to a temp path, assert the right directory is returned; unset them and assert the fallback; point them at a non-existent path and assert the fallback is used; assert the data-dir helper creates a nested path that didn't exist. Follow `docs/style/go.md` — package comment, MixedCaps, `%w` wrapping, no `docs/` references in comments.

**Blocked by:** 01 (scaffold Go module and CLI command)

**Status:** ready-for-agent

- [x] `internal/paths` resolves config, flavors, and data dirs per the spec's directory table
- [x] First-existing-directory selection; non-existent XDG path falls back; unset env reads as empty string
- [x] `filepath.Join`-based path building (no string concatenation with `"/"`)
- [x] Data-dir "ensure exists" helper uses `os.MkdirAll`
- [x] Unit tests with `t.TempDir()` + `t.Setenv` cover primary, fallback, non-existent, and creation cases
- [x] `go build ./...`, `go test ./...`, `go vet ./...` pass
