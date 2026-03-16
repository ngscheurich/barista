# Barista

[![Package Version](https://img.shields.io/hexpm/v/barista)](https://hex.pm/packages/barista)
[![Hex Docs](https://img.shields.io/badge/hex-docs-ffaff3)](https://hexdocs.pm/barista/)

# Flavors

Flavors consist of a `flavor.toml` data file and one or more application templates. By default, user-specified flavors are located at `$XDG_CONFIG_HOME/barista/flavors`.

## Example

```
~/.config
└── barista
    └── flavors
        └── catppuccin-mocha
            ├── flavor.toml
            ├── ghostty.mustache
            └── neovim.lua.mustache
```

# Recipes

A recipe tells Barista how to:

1. Convert a flavor template into a theme for a particular application
2. Load the new theme in the application after conversion

Recipes output application themes to `$XDG_DATA_HOME/barista`.

## Ghostty

The [Ghostty] recipe writes a config file called `ghostty`. Load this file in your main Ghostty config.

```
?~/.local/share/barista/ghostty`
```

## Neovim

The [Neovim] recipe has two dependencies:

- [Barista for Neovim], to load the current flavor
- [Catppuccin for Neovim], the theming framework

The recipe outputs a flavor plugin to `~/.local/share/barista/nvim`. Load the plugin in your Neovim config and run `setup`.

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

# Usage

```sh
gleam add barista@1
```

```gleam
import barista

pub fn main() -> Nil {
  // TODO: An example of the project in use
}
```

Further documentation can be found at <https://hexdocs.pm/barista>.

## Development

```sh
gleam run   # Run the project
gleam test  # Run the tests
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

[barista for neovim]: https://github.com/ngscheurich/barista-nvim
[catppuccin for neovim]: https://github.com/catppuccin/nvim
[ghostty]: https://ghostty.org/
[neovim]: https://neovim.io/
