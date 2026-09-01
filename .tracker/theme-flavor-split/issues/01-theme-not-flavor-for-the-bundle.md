# 01: "Theme" is the bundle; "flavor" is the palette variant

**What to build:** The domain-model change from ADR-0004. Split the overloaded "flavor" concept: `Flavor` keeps its Catppuccin meaning (a named 26-color palette variant) and a new `Theme` names the bundle (a Flavor plus its per-application templates). The user-facing `flavors/` directory renames to `themes/` (hard cut, no migration — ADR-0002 already declined to preserve pre-1.0 flavors, and `flavor.toml` keeps its name). The CLI argument becomes `apply [theme]`, and the rendered per-application artifact is called "output" rather than "theme file". The Neovim recipe's `flavor.lua` renames to `barista.lua`, which is a coordinated breaking change for the external `barista.nvim` repo.

- [x] `internal/flavor` owns `Flavor{Name, Palette}` and `Parse` (decodes `flavor.toml` content); drops `Dirname`
- [x] `internal/theme` owns `Theme{Dirname, Flavor}` with `Load`/`List`
- [x] `Recipe.Run(t theme.Theme)`; every recipe's `flavorsDir` field/param becomes `themesDir`
- [x] `paths.FlavorsDir` becomes `paths.ThemesDir` (`<config dir>/themes`)
- [x] CLI: `apply [theme]`, "Serves up a new theme", "Choose a theme", "no themes found"
- [x] Neovim output `flavor.lua` → `barista.lua`
- [x] Recipe write logs and comments say "Writing output", not "Writing theme"
- [x] CONTEXT.md, README, spec, `docs/style/go.md`, `docs/style/accessibility.md` swept to the new vocabulary
- [x] ADR-0004 recorded in `docs/adrs/`
- [ ] `barista.nvim` updated to read `barista.lua` instead of `flavor.lua` (external repo, coordinated change)
