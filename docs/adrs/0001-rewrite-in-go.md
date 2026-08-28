# ADR-0001: Rewrite Barista in Go

Date: 2026-08-28
Status: Accepted

## Context

Barista is currently a Gleam program targeting the Erlang/BEAM runtime. The motivation to rewrite is twofold:

1. **Drop the BEAM runtime dependency.** Distributing Barista today means shipping (or assuming) an Erlang runtime. We want a single self-contained binary with no runtime prerequisite.
2. **Reach for a richer CLI/TUI ecosystem.** The current CLI is built on Gleam's `glint`. We want headroom for interactive prompts now and a full-screen TUI later.

## Considered options

### TypeScript (Bun-compiled single binary)

- Dependency surface maps almost 1:1 onto the Node ecosystem (TOML, Mustache, env, fs, child_process all have mature libs).
- Clack gives interactive prompts; Bun can compile a standalone binary.
- Single-binary story is real but is an *extra build step* (Bun compile) rather than native, and would need verification for a CLI that shells out to `pgrep`/`kill`/`nvim`.
- TUI ceiling is lower: Clack is prompt-oriented; a full reactive TUI would mean adopting a less-proven library.

### Go (Cobra + Bubble Tea)

- Single static binary natively, zero runtime dependency. Strongest possible answer to motivation #1.
- Cobra is idiomatic for the `barista apply <theme>` command shape and the subcommands the roadmap implies (skip flags, config-file skips, `list`). Bubble Tea is the strongest TUI story on either side — directly serves motivation #2's "full TUI later" possibility.
- TOML (`BurntSushi/toml`) and Mustache (`cbroglie/mustache`) are workable in Go, though less central than the JS equivalents. This is the main cost.
- A Gleam→Go rewrite is a genuine rethink of structure (packages, error handling as `if err != nil`) rather than a mechanical port. Accepted: the code should fit the language naturally, not carry Gleam idioms over.

## Decision

Rewrite Barista in **Go**.

- CLI framework: **Cobra** (`spf13/cobra`), command shape `barista apply <theme>`.
- TUI (when reached for): **Bubble Tea** (`charmbracelet/bubbletea`).
- We do **not** carry over Gleam-idiomatic patterns (the `Result`-chain, `use`-callbacks, `snag` error layering). Go code uses `if err != nil` and `fmt.Errorf("...: %w", err)` for error context.
- We **do** preserve Barista's observable behavior in the port (faithful port first; roadmap items are follow-up tickets, not folded into the rewrite).

## Consequences

- The Gleam source, `gleam.toml`, `manifest.toml`, and `build/` tree are removed and replaced with a Go module (`go.mod`, `cmd/`, `internal/`).
- The `priv/scripts/nvim_send.sh` helper is **reimplemented in Go** (walk `$XDG_RUNTIME_DIR`/`$TMPDIR` for `nvim.*.0` sockets, invoke `nvim --server --remote-send`), so the binary stays self-contained.
- Distribution becomes `go build` → single binary; a future `goreleaser` config can produce cross-platform release binaries.
- Testing: stdlib `testing` + `testify/assert`.
