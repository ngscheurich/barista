# CONTEXT.md — Barista

## What Barista is

Barista takes a **flavor** (a named color palette in Catppuccin's format) and **serves** it to terminal applications — rendering templates into theme files and triggering a reload in each app.

## Glossary

- **Flavor** — A named collection of colors. Lives on disk as `flavor.toml` under a flavors directory, alongside one or more application templates.
- **Palette** — The 26 Catppuccin color values (rosewater … crust) inside a Flavor.
- **Application** — A terminal program Barista can theme. Today: Ghostty, Neovim, Zellij.
- **Recipe** — The procedure for one Application: render its template from the Flavor, write the theme file to the right location, and reload the Application.
- **Template** — A Mustache file (e.g. `ghostty.mustache`) that, given a Flavor as context, produces an application theme.
- **Picker** — The interactive chooser `barista apply` opens when invoked with no Flavor argument. Lists available Flavors by Name; serves the one the user chooses. Not available when input is not a terminal, in which case `apply` fails and lists the Flavors instead.
- **config directory** — The root of Barista's user configuration. `$XDG_CONFIG_HOME/barista` (fallback `~/.config/barista`). Not created by Barista; assumed to exist.
- **flavors directory** — Where user-specified flavors live. A subset of the config directory: `<config directory>/flavors`.
- **data directory** — Where Barista writes generated themes. `$XDG_DATA_HOME/barista` (fallback `~/.local/share/barista`). Created by Barista on every run.

## Implementation

- **Language:** Go (single static binary, no runtime dependency).
- **CLI:** Cobra (`spf13/cobra`); command shape `barista apply [flavor]` — one Flavor argument, or none to open the Picker.
- **Interactive prompts:** huh (`charm.land/huh/v2`), chosen over Bubble Tea directly because of its accessible mode (ADR-0003). Accessible mode is opt-in via the `BARISTA_ACCESSIBLE` environment variable.
- **Styling:** lipgloss and log from the `charm.land` v2 line (see ADR-0003).
- **Flavor format:** TOML via `BurntSushi/toml`.
- **Templates:** Mustache via `cbroglie/mustache`.
- **Tests:** stdlib `testing` + `testify/assert`.
- **Module path:** `github.com/ngscheurich/barista`.
- The Neovim reload helper (`priv/scripts/nvim_send.sh` in the Gleam version) is reimplemented in Go so the binary stays self-contained.
- Gleam-idiomatic patterns (`Result`-chains, `use`-callbacks, `snag` layering) are **not** carried over; Go code uses `if err != nil` and `fmt.Errorf("...: %w", err)`.

## Decision log

- **ADR-0001** — Rewrite from Gleam to Go. See `docs/adrs/0001-rewrite-in-go.md`.
- **ADR-0002** — Flavor format is TOML; templates are Mustache. See `docs/adrs/0002-formats-toml-and-mustache.md`.
- **ADR-0003** — The picker is a huh form; Charm stack on the v2 line. See `docs/adrs/0003-huh-for-the-picker.md`.

## Style guides

- `docs/style/go.md` — idiomatic Go for this codebase.
- `docs/style/accessibility.md` — accessibility as a core principle for CLI/TUI output.

## Repo conventions

- **ADRs live in `docs/adrs/`** (not `docs/adr/`), numbered `NNNN-kebab-title.md`.
- **Tickets live in `.tracker/<feature-slug>/issues/`**, one file per ticket, numbered from `01` in dependency order.
