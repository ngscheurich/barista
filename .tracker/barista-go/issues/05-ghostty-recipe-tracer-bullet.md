# 05: Ghostty recipe — tracer bullet through the CLI

**What to build:** The first end-to-end vertical slice, proving the architecture by cutting through every layer. Define the `internal/recipe` package with the `Recipe` interface (`Run(f flavor.Flavor) error` — single method, per Go's small-interface idiom). Implement `internal/recipe/ghostty`: locate `<flavors dir>/<theme>/ghostty.mustache` (via `internal.paths`), read it, render via `internal.template.Render`, write the output to `<data dir>/ghostty`, then reload by running `pgrep ghostty`, taking the first pid from its output, and `kill -s USR2 <pid>`. Construct child processes with `os/exec` (call binaries directly, never `sh -c`); extract command construction into a package-level helper that returns `*exec.Cmd` without running it, so the reload step is testable without spawning real `pgrep`/`kill`. Wire the recipe into the `apply` command: `barista apply <theme>` ensures the data dir exists, loads the flavor, runs the Ghostty recipe, and on success prints `☕︎ Served up <name>` (the flavor's `Name`, not the dirname) to stdout; on failure prints the error to stderr and exits non-zero. Introduce the aggregated-error collection loop (a `[]error` slice, joined with `errors.Join` at the end) — currently holding one recipe, ready for 06 and 07 to add two more without changing the orchestration. Integration-test against a `t.TempDir()` flavors dir with a real `ghostty.mustache` template: assert the output file is written with the rendered content. Test the reload step via the command-construction seam: assert the `pgrep` and `kill` args are correct without spawning processes. This is the tracer bullet — it establishes the recipe interface, the exec seam, the error-aggregation pattern, and the CLI-to-recipe wiring that 06 and 07 reuse.

**Blocked by:** 04 (template rendering)

**Status:** ready-for-agent

- [x] `internal/recipe` defines the `Recipe` interface (`Run(f flavor.Flavor) error`)
- [x] `internal/recipe/ghostty` locates, reads, renders, and writes the Ghostty theme to `<data dir>/ghostty`
- [x] Ghostty reload runs `pgrep ghostty` → first pid → `kill -s USR2 <pid>` via `os/exec` (no `sh -c`)
- [x] Command construction extracted into a testable seam (returns `*exec.Cmd` without running)
- [x] `barista apply <theme>` ensures data dir, loads flavor, runs Ghostty, prints `☕︎ Served up <name>` on success
- [x] Aggregated-error collection loop (`[]error` + `errors.Join`) introduced, holding one recipe
- [x] Integration test writes a real template into a temp flavors dir and asserts the rendered output file
- [x] Reload tested via the command-construction seam (assert args, no process spawn)
- [x] `go build ./...`, `go test ./...`, `go vet ./...` pass
