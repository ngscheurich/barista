import argv
import barista/file
import barista/recipes/ghostty
import barista/recipes/neovim
import barista/recipes/zellij
import gleam/dict.{type Dict}
import gleam/io
import gleam/result
import gleam/string
import gleam/string_tree.{type StringTree}
import glint.{type Command}
import handles.{type Template}
import handles/ctx
import handles/format as handles_format
import snag
import tom.{type GetError, type Toml}

// TYPES -------------------------------------------------------------------------

/// A Config is a set of user-specified options that changes Barista's behavior.
type Config

/// An Application is a terminal program that Barista can theme.
type Application {
  Ghostty
  Neovim
  Zellij
}

/// A Flavor represents a named collection of colors that map to those outlined
/// in the [Catppuccin palette specification](https://catppuccin.com/palette/).
type Flavor {
  Flavor(name: String, dirname: String, palette: Palette)
}

/// A Palette is the color definitions of a `Flavor`.
type Palette {
  Palette(
    rosewater: String,
    flamingo: String,
    pink: String,
    mauve: String,
    red: String,
    maroon: String,
    peach: String,
    yellow: String,
    green: String,
    teal: String,
    sky: String,
    sapphire: String,
    blue: String,
    lavender: String,
    text: String,
    subtext_1: String,
    subtext_0: String,
    overlay_2: String,
    overlay_1: String,
    overlay_0: String,
    surface_2: String,
    surface_1: String,
    surface_0: String,
    base: String,
    mantle: String,
    crust: String,
  )
}

// FLAVORS -----------------------------------------------------------------------

fn load_flavor(dirname: String) -> snag.Result(Flavor) {
  use data_filepath <- result.try(file.theme_filepath(dirname, "flavor.toml"))

  use toml_input <- result.try(
    file.read(data_filepath)
    |> snag.context("Could not read data file"),
  )

  use toml <- result.try(parse_toml(toml_input))

  build_flavor(dirname, toml)
}

fn parse_toml(input: String) -> snag.Result(Dict(String, Toml)) {
  tom.parse(input)
  |> result.map_error(fn(_) { snag.new("Could not parse TOML") })
}

fn build_flavor(
  dirname: String,
  toml: Dict(String, Toml),
) -> snag.Result(Flavor) {
  from_toml(dirname, toml)
  |> result.map_error(fn(e) {
    case e {
      tom.NotFound(key:) -> snag.new(string.join(key, " not found"))
      tom.WrongType(key: _, expected:, got:) ->
        snag.new("Expected " <> expected <> ", got " <> got)
    }
  })
}

fn from_toml(
  dirname: String,
  toml: Dict(String, Toml),
) -> Result(Flavor, GetError) {
  use name <- result.try(tom.get_string(toml, ["name"]))
  use rosewater <- result.try(get_color(toml, "rosewater"))
  use flamingo <- result.try(get_color(toml, "flamingo"))
  use pink <- result.try(get_color(toml, "pink"))
  use mauve <- result.try(get_color(toml, "mauve"))
  use red <- result.try(get_color(toml, "red"))
  use maroon <- result.try(get_color(toml, "maroon"))
  use peach <- result.try(get_color(toml, "peach"))
  use yellow <- result.try(get_color(toml, "yellow"))
  use green <- result.try(get_color(toml, "green"))
  use teal <- result.try(get_color(toml, "teal"))
  use sky <- result.try(get_color(toml, "sky"))
  use sapphire <- result.try(get_color(toml, "sapphire"))
  use blue <- result.try(get_color(toml, "blue"))
  use lavender <- result.try(get_color(toml, "lavender"))
  use text <- result.try(get_color(toml, "text"))
  use subtext_1 <- result.try(get_color(toml, "subtext_1"))
  use subtext_0 <- result.try(get_color(toml, "subtext_0"))
  use overlay_2 <- result.try(get_color(toml, "overlay_2"))
  use overlay_1 <- result.try(get_color(toml, "overlay_1"))
  use overlay_0 <- result.try(get_color(toml, "overlay_0"))
  use surface_2 <- result.try(get_color(toml, "surface_2"))
  use surface_1 <- result.try(get_color(toml, "surface_1"))
  use surface_0 <- result.try(get_color(toml, "surface_0"))
  use base <- result.try(get_color(toml, "base"))
  use mantle <- result.try(get_color(toml, "mantle"))
  use crust <- result.map(get_color(toml, "crust"))

  Flavor(
    name:,
    dirname:,
    palette: Palette(
      rosewater:,
      flamingo:,
      pink:,
      mauve:,
      red:,
      maroon:,
      peach:,
      yellow:,
      green:,
      teal:,
      sky:,
      sapphire:,
      blue:,
      lavender:,
      text:,
      subtext_1:,
      subtext_0:,
      overlay_2:,
      overlay_1:,
      overlay_0:,
      surface_2:,
      surface_1:,
      surface_0:,
      base:,
      mantle:,
      crust:,
    ),
  )
}

fn get_color(
  toml: Dict(String, Toml),
  color: String,
) -> Result(String, GetError) {
  tom.get_string(toml, ["palette", color])
}

fn build_template_context(flavor: Flavor) -> ctx.Value {
  let palette = flavor.palette

  ctx.Dict([
    str_prop("name", flavor.name),
    ctx.Prop(
      "palette",
      ctx.Dict([
        str_prop("rosewater", palette.rosewater),
        str_prop("flamingo", palette.flamingo),
        str_prop("pink", palette.pink),
        str_prop("mauve", palette.mauve),
        str_prop("red", palette.red),
        str_prop("maroon", palette.maroon),
        str_prop("peach", palette.peach),
        str_prop("yellow", palette.yellow),
        str_prop("green", palette.green),
        str_prop("teal", palette.teal),
        str_prop("sky", palette.sky),
        str_prop("sapphire", palette.sapphire),
        str_prop("blue", palette.blue),
        str_prop("lavender", palette.lavender),
        str_prop("text", palette.text),
        str_prop("subtext_1", palette.subtext_1),
        str_prop("subtext_0", palette.subtext_0),
        str_prop("overlay_2", palette.overlay_2),
        str_prop("overlay_1", palette.overlay_1),
        str_prop("overlay_0", palette.overlay_0),
        str_prop("surface_2", palette.surface_2),
        str_prop("surface_1", palette.surface_1),
        str_prop("surface_0", palette.surface_0),
        str_prop("base", palette.base),
        str_prop("mantle", palette.mantle),
        str_prop("crust", palette.crust),
      ]),
    ),
  ])
}

fn str_prop(key: String, value: String) -> ctx.Prop {
  ctx.Prop(key, ctx.Str(value))
}

// TEMPLATING --------------------------------------------------------------------

// TODO: Improve errors around invalid templates
fn process_template(template: String, flavor: Flavor) -> snag.Result(StringTree) {
  use prepared <- result.try(prepare_template(template))
  run_template(prepared, flavor, template)
}

fn prepare_template(template: String) -> snag.Result(Template) {
  handles.prepare(template)
  |> result.map_error(fn(_) { snag.new("Could not tokenize template") })
}

fn run_template(
  template: Template,
  flavor: Flavor,
  template_str: String,
) -> snag.Result(StringTree) {
  let context = build_template_context(flavor)

  case handles.run(template, context, []) {
    Ok(template) -> Ok(template)
    Error(error) ->
      handles_format.format_runtime_error(error, template_str)
      |> result.unwrap("Template error")
      |> snag.error
  }
}

// MAIN --------------------------------------------------------------------------

fn barista() -> Command(Nil) {
  use <- glint.command_help("Serves up a new <theme> for your terminal apps.")
  use theme <- glint.named_arg("theme")
  use <- glint.unnamed_args(glint.EqArgs(0))

  use named, _, _ <- glint.command()

  theme(named)
  |> apply_theme
}

fn ensure_barista_data_dir() -> snag.Result(String) {
  use dirpath <- result.try(file.barista_data_dir())
  file.make_dirpath(dirpath)
}

fn apply_theme(dirname: String) -> Nil {
  let result = {
    use _ <- result.try(ensure_barista_data_dir())
    use flavor <- result.try(load_flavor(dirname))
    use _ <- result.try(apply_app_theme(Ghostty, flavor))
    use _ <- result.try(apply_app_theme(Neovim, flavor))
    use _ <- result.try(apply_app_theme(Zellij, flavor))

    Ok(flavor)
  }

  case result {
    Ok(flavor) -> io.println("☕︎ Served up " <> flavor.name)
    Error(snag) -> io.println(snag.pretty_print(snag))
  }
}

fn apply_app_theme(app: Application, flavor: Flavor) -> snag.Result(Nil) {
  case app {
    Ghostty -> apply_ghostty_theme(flavor)
    Neovim -> apply_neovim_theme(flavor)
    Zellij -> apply_zellij_theme(flavor)
  }
}

fn apply_ghostty_theme(flavor: Flavor) -> snag.Result(Nil) {
  use filepath <- result.try(file.theme_filepath(
    flavor.dirname,
    "ghostty.mustache",
  ))
  use _ <- result.try(write_theme(filepath, flavor, ghostty.write_theme))
  ghostty.apply_theme()
}

fn apply_neovim_theme(flavor: Flavor) -> snag.Result(Nil) {
  use filepath <- result.try(file.theme_filepath(
    flavor.dirname,
    "neovim.lua.mustache",
  ))
  use _ <- result.try(write_theme(filepath, flavor, neovim.write_theme))
  neovim.apply_theme()
}

fn apply_zellij_theme(flavor: Flavor) -> snag.Result(Nil) {
  use filepath <- result.try(file.theme_filepath(
    flavor.dirname,
    "zellij.kdl.mustache",
  ))
  use _ <- result.try(write_theme(filepath, flavor, zellij.write_theme))
  zellij.apply_theme()
}

fn write_theme(
  filepath: String,
  flavor: Flavor,
  writer: fn(StringTree) -> snag.Result(Nil),
) -> snag.Result(Nil) {
  use template <- result.try(file.read(filepath))
  use theme <- result.try(process_template(template, flavor))
  writer(theme)
}

pub fn main() -> Nil {
  glint.new()
  |> glint.with_name("barista")
  |> glint.pretty_help(glint.default_pretty_help())
  |> glint.add(at: [], do: barista())
  |> glint.run(argv.load().arguments)
}
