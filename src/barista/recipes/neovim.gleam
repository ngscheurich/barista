import barista/file
import child_process
import child_process/stdio
import gleam/erlang/application
import gleam/result
import gleam/string_tree.{type StringTree}
import snag

/// Writes flavor plugin to `$XDG_DATA_HOME/barista/nvim`, which is loaded by
/// [Barista for Neovim](https://github.com/ngscheurich/barista-nvim).
pub fn write_theme(contents: StringTree) -> snag.Result(Nil) {
  use data_dir <- result.try(file.barista_data_dir())

  use plugin_dirpath <- result.try(
    file.build_path([data_dir, "nvim", "lua"])
    |> file.make_dirpath,
  )

  let filepath = file.build_path([plugin_dirpath, "flavor.lua"])
  let contents = string_tree.to_string(contents)

  file.write(filepath, contents)
  |> result.map_error(fn(_) { snag.new("Could not write file: " <> filepath) })
}

/// Sends a command to all running Neovim instances to reload Barista.
pub fn apply_theme() -> snag.Result(Nil) {
  let assert Ok(priv_dir) = application.priv_directory("barista")
  let cmd = priv_dir <> "/scripts/nvim_send.sh"

  let result =
    child_process.from_name(cmd)
    |> child_process.args(["<Cmd>lua require('barista')<CR>"])
    |> child_process.run(stdio.capture(True))

  case result {
    Ok(_) -> Ok(Nil)
    Error(_) -> snag.error("Process exited with error (" <> cmd <> ")")
  }
}
