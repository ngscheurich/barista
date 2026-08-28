# Barista Go — Specification

A faithful port of Barista from Gleam/BEAM to Go, preserving observable behavior. See ADR-0001 (Go rewrite) and ADR-0002 (TOML + Mustache).

## Scope

**In scope:** a Go binary that reproduces today's `barista <theme>` behavior — load a flavor, render and write themes for Ghostty, Neovim, and Zellij, reload each app.

**Out of scope (follow-up tickets, not folded in):** checking for an application before applying, skipping applications via CLI flags or config file, only trying applications with an available template, any TUI. These remain on the roadmap after the port lands.

## Command surface

```
barista apply <theme>
```

`<theme>` is the directory name of a flavor under the flavors directory. Exactly one positional argument, zero unnamed args. Help text: `Serves up a new <theme> for your terminal apps.` Cobra provides `--help` and the standard flag handling.

(The bare `barista <theme>` form from the Gleam version is **not** preserved; the Go version uses the `apply` subcommand to leave room for future subcommands. This is the one intentional behavior change.)

## Behavior contract

`barista apply <theme>` runs these steps in order:

1. **Ensure the data directory exists.** Create `$XDG_DATA_HOME/barista` (fallback `~/.local/share/barista`) with `mkdir -p` semantics. This happens before the flavor is loaded. A failure here aborts before any recipe runs.
2. **Load the flavor.** Read `<flavors dir>/<theme>/flavor.toml`, parse it as TOML, and build a `Flavor` (a `name` string plus a `Palette` of the 26 Catppuccin colors). The `dirname` is `<theme>`. A failure here aborts before any recipe runs (there is no flavor to render).
3. **Run all three recipes** (Ghostty, Neovim, Zellij), collecting errors rather than stopping at the first. The recipes are independent (each writes its own file and reloads its own app), so a Ghostty failure does not block Neovim or Zellij.
4. **Report.** On success (no errors), print `☕︎ Served up <name>` (the flavor's `name` field, not the dirname). If any recipe errored, print all collected errors to stderr and exit non-zero.

This continue-on-error behavior is a deliberate change from the Gleam version's short-circuit.

### Directory resolution

| Directory | Primary | Fallback | Created by Barista? |
| --- | --- | --- | --- |
| Config dir | `$XDG_CONFIG_HOME/barista` | `~/.config/barista` | No (assumed to exist) |
| Flavors dir | `<config dir>/flavors` | — | No (assumed to exist) |
| Data dir | `$XDG_DATA_HOME/barista` | `~/.local/share/barista` | Yes, on every run |

Resolution picks the first candidate in `[primary, fallback]` that exists and is a directory. If `XDG_*` is set but points at a non-existent path, the fallback is used. If neither exists, the relevant step errors. Env vars that are unset read as the empty string (replicated from the Gleam version).

### Recipes

Each recipe: locate the template, read it, render it against the flavor, write the output, then reload the app.

| App | Template file | Output file | Reload mechanism |
| --- | --- | --- | --- |
| Ghostty | `<flavors dir>/<theme>/ghostty.mustache` | `<data dir>/ghostty` | `pgrep ghostty` → first pid → `kill -s USR2 <pid>` |
| Neovim | `<flavors dir>/<theme>/neovim.lua.mustache` | `<data dir>/nvim/lua/flavor.lua` (creates `nvim/lua/`) | Send `<Cmd>lua require('barista')<CR>` to every `nvim.*.0` socket under `$XDG_RUNTIME_DIR` (fallback `$TMPDIR/nvim.<user>`) via `nvim --server <socket> --remote-send` |
| Zellij | `<flavors dir>/<theme>/zellij.kdl.mustache` | `<config dir>/zellij/themes/barista.kdl` (creates `zellij/themes/`) | `touch <config dir>/zellij/config.kdl` |

The Neovim reload is reimplemented in Go (discovering sockets and invoking `nvim --server --remote-send`), replacing the `priv/scripts/nvim_send.sh` shell helper, so the binary stays self-contained.

### Template context

A Mustache context with two top-level keys:

- `name` — the flavor's `name` string.
- `palette` — a map of the 26 Catppuccin color names to their hex string values: `rosewater`, `flamingo`, `pink`, `mauve`, `red`, `maroon`, `peach`, `yellow`, `green`, `teal`, `sky`, `sapphire`, `blue`, `lavender`, `text`, `subtext_1`, `subtext_0`, `overlay_2`, `overlay_1`, `overlay_0`, `surface_2`, `surface_1`, `surface_0`, `base`, `mantle`, `crust`.

Templates access these as `{{name}}` and `{{palette.rosewater}}` etc., as in the Gleam version.

### `flavor.toml` shape

```toml
name = "Catppuccin Mocha"

[palette]
rosewater = "#f5e0dc"
# ... all 26 colors
crust = "#11111b"
```

## Error strategy

Go idioms, not Gleam `Result`-chains. Each function that can fail returns `error`; callers use `if err != nil { return fmt.Errorf("...: %w", err) }` to add context. The top-level `apply` runs all three recipes and collects their errors into a slice; if the slice is non-empty, it prints every error to stderr and exits non-zero. Per-recipe failures do not abort the run.

## Proposed module layout (for reaction)

```
cmd/barista/main.go        # entry point; wires Cobra root + apply command
internal/flavor/flavor.go  # Flavor, Palette types; Load(dirname) reads+parses TOML
internal/paths/paths.go    # XDG dir resolution, flavors/data/config paths, filepath join
internal/template/template.go  # Mustache render wrapper
internal/recipe/recipe.go  # Recipe interface
internal/recipe/ghostty/ghostty.go
internal/recipe/neovim/neovim.go
internal/recipe/zellij/zellij.go
internal/nvim/nvim.go      # socket discovery + nvim --server --remote-send (replaces nvim_send.sh)
```

The `Recipe` interface (single-method, per Go's small-interface idiom):

```go
type Recipe interface {
    Run(f flavor.Flavor) error  // render template, write theme file, reload app
}
```

`apply` loops over the three recipes in order, calling `Run` on each and collecting errors into a slice. The render-and-write step inside each recipe stays unit-testable through package-level helper functions, not through the interface.

## Resolved decisions

- **Module layout** — `cmd/` + `internal/` split; the standard idiomatic Go CLI layout.
- **Recipe interface** — single method `Run(f Flavor) error`, per Go's small-interface idiom.
- **Error handling** — continue-on-error across recipes, aggregate, report all at the end. Deliberate change from the Gleam version's short-circuit.
- **Command form** — `barista apply <theme>` (subcommand form, for headroom).

## Testing

Stdlib `testing` + `testify/assert`. The Gleam test suite is a single placeholder (`hello_world_test`), so there is no existing behavior coverage to port — we write fresh tests for the Go implementation. Priority targets: flavor loading/parsing, path resolution, template rendering, and the recipe render step (the reload steps shell out to external processes and are harder to unit-test; an integration test that exercises the full `apply` against a temp flavors dir is the likely seam).
