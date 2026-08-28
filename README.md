# Barista

Barista takes a flavor (a named color palette in Catppuccin's format) and
serves it to terminal applications — rendering templates into theme files
and triggering a reload in each.

# Flavors

Flavors consist of a `flavor.toml` data file and one or more application
templates. By default, user-specified flavors are located at
`$XDG_CONFIG_HOME/barista/flavors`.

## Example

A flavor is a directory under `flavors/` containing a `flavor.toml` data
file and one Mustache template per application:

```
~/.config
└── barista
    └── flavors
        └── catppuccin-mocha
            ├── flavor.toml
            ├── ghostty.mustache
            ├── neovim.lua.mustache
            └── zellij.kdl.mustache
```

`flavor.toml` carries the flavor's name and a `[palette]` table of all 26
Catppuccin colors:

```toml
name = "Catppuccin Mocha"

[palette]
rosewater = "#f5e0dc"
flamingo  = "#f2cdcd"
pink      = "#f5c2e7"
mauve     = "#cba6f7"
red       = "#f38ba8"
maroon    = "#eba0ac"
peach     = "#fab387"
yellow    = "#f9e2af"
green     = "#a6e3a1"
teal      = "#94e2d5"
sky       = "#89dceb"
sapphire  = "#74c7ec"
blue      = "#89b4fa"
lavender  = "#b4befe"
text      = "#cdd6f4"
subtext_1 = "#bac2de"
subtext_0 = "#a6adc8"
overlay_2 = "#9399b2"
overlay_1 = "#7f849c"
overlay_0 = "#6c7086"
surface_2 = "#585b70"
surface_1 = "#45475a"
surface_0 = "#313244"
base      = "#1e1e2e"
mantle    = "#181825"
crust     = "#11111b"
```

Each template is a Mustache file that references `{{name}}` and
`{{palette.<color>}}`. For example, `ghostty.mustache`:

```
# {{name}}
background = {{palette.base}}
foreground = {{palette.text}}
cursor-color = {{palette.rosewater}}
```

Running `barista apply catppuccin-mocha` renders each template against the
flavor and writes the result to the application's theme location.

# Recipes

A recipe tells Barista how to:

1. Convert a flavor template into a theme for a particular application
2. Load the new theme in the application after conversion

Recipes output application themes to `$XDG_DATA_HOME/barista`, except
Zellij, which reads its themes from the user's config directory.

## Ghostty

The [Ghostty] recipe writes a config file called `ghostty`. Load this file
in your main Ghostty config.

```
~/.local/share/barista/ghostty
```

## Neovim

The [Neovim] recipe has two dependencies:

- [Barista for Neovim], to load the current flavor
- [Catppuccin for Neovim], the theming framework

The recipe outputs a flavor plugin to `~/.local/share/barista/nvim`. Load
the plugin in your Neovim config and run `setup`.

> [!IMPORTANT]
> Ensure the dependencies are available when you run `setup`.

```lua
-- Add the required plugins
vim.pack.add({
  "https://github.com/ngscheurich/barista.nvim",
  { src = "https://github.com/catppuccin/nvim", name = "catppuccin" },
})

-- Run Barista setup
require("barista").setup()

-- Or, with an optional callback function
require("barista").setup(require("my.statusline").setup)
```

## Zellij

The [Zellij] recipe writes a theme file called `barista.kdl` to
`~/.config/zellij/themes/`. Zellij picks up themes from this directory on
reload, which the recipe triggers by touching `config.kdl`.

```
~/.config/zellij/themes/barista.kdl
```

# Usage

```sh
barista apply <theme>
```

`<theme>` is the directory name of a flavor under the flavors directory.
On success Barista prints `☕︎ Served up <name>` (the flavor's name, not
the dirname) and exits zero; on any failure it prints the error(s) to
stderr and exits non-zero.

## Development

```sh
go run ./cmd/barista apply <theme>   # Run the CLI
go test ./...                         # Run the tests
go build -o bin/barista ./cmd/barista # Build the binary
```

## Tasks

- [x] Load palette from disk (~/.local/share/barista/<theme>/palette.toml)
- [x] Compile and write Ghostty theme (~/.local/share/barista/<theme>/ghossty.mustache)
- [x] Reload Ghostty them
- [x] Compile and write Neovim theme (~/.local/share/barista/<theme>/neovim.lua.mustache)
- [x] Reload Neovim theme
- [x] Compile and write Zellij theme (~/.local/share/barista/<theme>/zellij.kdl.mustache)
- [x] Reload Neovim theme
- [x] CLI wrapper
- [ ] Check for an application before applying theme
- [ ] Allow skipping applications via CLI flags
- [ ] Allow skipping applications via config file
- [ ] Only try applications with an available template?

[barista for neovim]: https://github.com/ngscheurich/barista.nvim
[catppuccin for neovim]: https://github.com/catppuccin/nvim
[ghostty]: https://ghostty.org/
[neovim]: https://neovim.io/
[zellij]: https://zellij.dev/
