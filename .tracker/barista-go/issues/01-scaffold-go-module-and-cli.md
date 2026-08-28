# 01: Scaffold Go module and CLI command

**What to build:** A buildable, testable Go project at the repo root that provides the `barista` CLI with an `apply <theme>` subcommand. `go.mod` declares the module `github.com/ngscheurich/barista` with Go 1.26 and the cobra dependency. `cmd/barista/main.go` wires a Cobra root command (name `barista`, short help `Serves up a new <theme> for your terminal apps.`, `SilenceUsage` and `SilenceErrors` on) and an `apply` subcommand taking exactly one positional arg (`cobra.ExactArgs(1)`) that, for now, prints the theme name to stdout and exits zero. The repo's `.gitignore` is rewritten to reflect a Go project (drop `*.beam`, `*.ez`, `/build`, `erl_crash.dump`; add `/bin/`, Go build/output conventions) — the Gleam entries are removed here because the Gleam source itself is removed in ticket 07, but a Go-appropriate ignore file belongs from the moment the Go module lands. `go build ./...`, `go test ./...`, and `go vet ./...` all pass. This is the skeleton every later ticket hangs off: it establishes the module, the `cmd/` + `internal/` layout, the Cobra command tree, the CI-green baseline, and the ignore-file shape.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] `go.mod` present with module `github.com/ngscheurich/barista`, `go 1.26`, cobra required
- [ ] `cmd/barista/main.go` wires a Cobra root + `apply <theme>` subcommand (one positional arg)
- [ ] `barista apply <theme>` prints the theme name and exits zero; `barista` with no args prints help
- [ ] `.gitignore` rewritten for Go (Gleam/Erlang entries dropped, `/bin/` and Go build conventions added)
- [ ] `go build ./...`, `go test ./...`, and `go vet ./...` pass
