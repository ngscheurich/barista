# ADR-0004: "Theme" is the bundle; "flavor" is the Catppuccin palette variant

Date: 2026-08-31
Status: Accepted

Barista reads a named 26-color palette plus per-application templates and renders them for each app. Before this change, "flavor" named both that whole bundle and — since the palettes are in Catppuccin's format — Catppuccin's own "flavor" (a palette variant like Mocha). The README's first example nested one meaning inside the other (`catppuccin-mocha` barista-"flavor" whose palette *is* the Catppuccin Mocha flavor), which is an unacceptable collision given the tie-in.

## Decision

Split the two concepts and give each its own word:

- **Flavor** — a named Catppuccin palette variant (a `name` plus the 26 colors), serialized as `flavor.toml`. This is Catppuccin's own meaning.
- **Theme** — a Flavor plus the templates that render it, living as a directory under the **themes directory** (`<config dir>/themes`).
- The per-application artifact a recipe produces is called "output" and no longer "theme file"/"generated themes" — it is application code/config, not the palette.

The on-disk `flavors/` directory renames to `themes/` as a hard cut (no migration — ADR-0002 already declined to preserve existing flavors pre-1.0), while `flavor.toml` keeps its name. The CLI argument becomes `barista apply [theme]`. The Neovim recipe's rendered file renames `flavor.lua` → `barista.lua` (matching Zellij's `barista.kdl` output), which is a coordinated breaking change for the external `barista.nvim` repo.

## Considered options

- **Rename "flavor" → "theme" wholesale.** Rejected: it discards Catppuccin's word for the palette variant, which is the one concept threads through `flavor.toml`, the `[palette]` table, and the docs.
- **Keep "flavor" for both meanings and document the disambiguation.** Rejected: the collision is in the most user-visible example, not a footnote.
- **Name the output "flavor file".** Rejected: a rendered artifact may use some, none, or a re-expression of the palette colors, so it is not a flavor.
- **Name the output "application file" / "theme target".** Rejected: fzf's recipe writes no file (it sets an env var), and "target" already means the Application here.

## Consequences

- `internal/flavor` owns `Flavor`/`Palette` and decoding `flavor.toml` content; a new `internal/theme` owns `Theme{Dirname, Flavor}` and directory loading/listing.
- ADR-0001, ADR-0002, and ADR-0003 use "flavor" to mean the bundle; this ADR supersedes that usage without rewriting those records. (ADR-0002's "flavor format is TOML" remains literally true — `flavor.toml` still is the flavor's format.)
- `barista.nvim` must be updated in step to read `barista.lua` instead of `flavor.lua`.
