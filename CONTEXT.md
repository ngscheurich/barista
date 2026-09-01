# CONTEXT.md — Barista

## What Barista is

Barista takes a **theme** (a Catppuccin flavor plus per-application templates) and **serves** it to terminal applications — rendering the templates into each application's output and triggering a reload in each.

## Glossary

- **Theme** — A named bundle that pairs a Flavor with the templates that render it. Lives on disk as a directory under a themes directory, containing a `flavor.toml` and one or more application templates.
- **Flavor** — A named Catppuccin palette variant: a name plus the 26 color values. Serialized as `flavor.toml`.
- **Palette** — The 26 Catppuccin color values (rosewater … crust) inside a Flavor.
- **Application** — A terminal program Barista can theme. Today: fzf, Ghostty, Neovim, Zellij.
- **Recipe** — The procedure for one Application: render its template from the Flavor, write the output to the Application's location, and reload the Application.
- **Template** — A Mustache file (e.g. `ghostty.mustache`) that, given a Flavor as context, produces an Application's output.
- **Picker** — The interactive chooser `barista apply` opens when invoked with no theme argument. Lists available Themes by name; serves the one the user chooses. Not available when input is not a terminal, in which case `apply` fails and lists the Themes instead.
- **config directory** — The root of Barista's user configuration. `$XDG_CONFIG_HOME/barista` (fallback `~/.config/barista`). Not created by Barista; assumed to exist.
- **themes directory** — Where user-specified themes live. A subset of the config directory: `<config directory>/themes`.
- **data directory** — Where recipes write their output. `$XDG_DATA_HOME/barista` (fallback `~/.local/share/barista`). Created by Barista on every run.

## Implementation

- **Language:** Go (single static binary, no runtime dependency).
- **CLI:** Cobra (`spf13/cobra`); command shape `barista apply [theme]` — one theme argument, or none to open the Picker.
- **Interactive prompts:** huh (`charm.land/huh/v2`), chosen over Bubble Tea directly because of its accessible mode (ADR-0003). Accessible mode is opt-in via the `BARISTA_ACCESSIBLE` environment variable.
- **Styling:** lipgloss and log from the `charm.land` v2 line (see ADR-0003).
- **Flavor format:** TOML via `BurntSushi/toml`. The `internal/flavor` package holds `Flavor`/`Palette` and decodes `flavor.toml` content; the `internal/theme` package holds `Theme` and loads it from the themes directory.
- **Templates:** Mustache via `cbroglie/mustache`.
- **Tests:** stdlib `testing` + `testify/assert`.
- **Module path:** `github.com/ngscheurich/barista`.
- The Neovim reload helper (`priv/scripts/nvim_send.sh` in the Gleam version) is reimplemented in Go so the binary stays self-contained.
- Gleam-idiomatic patterns (`Result`-chains, `use`-callbacks, `snag` layering) are **not** carried over; Go code uses `if err != nil` and `fmt.Errorf("...: %w", err)`.

## Decision log

- **ADR-0001** — Rewrite from Gleam to Go. See `docs/adrs/0001-rewrite-in-go.md`.
- **ADR-0002** — Flavor format is TOML; templates are Mustache. See `docs/adrs/0002-formats-toml-and-mustache.md`.
- **ADR-0003** — The picker is a huh form; Charm stack on the v2 line. See `docs/adrs/0003-huh-for-the-picker.md`.
- **ADR-0004** — "Theme" is the bundle; "flavor" is the Catppuccin palette variant. See `docs/adrs/0004-theme-is-the-bundle-flavor-is-the-palette.md`.

## Style guides

- `docs/style/go.md` — idiomatic Go for this codebase.
- `docs/style/accessibility.md` — accessibility as a core principle for CLI/TUI output.

## Repo conventions

- **ADRs live in `docs/adrs/`** (not `docs/adr/`), numbered `NNNN-kebab-title.md`.
- **Tickets live in `.tracker/<feature-slug>/issues/`**, one file per ticket, numbered from `01` in dependency order.
