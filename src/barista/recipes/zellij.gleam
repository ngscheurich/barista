import barista/file
import child_process
import child_process/stdio
import gleam/result
import gleam/string_tree.{type StringTree}
import snag

// Writes Zellij config file to `$XDG_CONFIG_HOME/zellij/themes/barista.kdl`.
pub fn write_theme(contents: StringTree) -> snag.Result(Nil) {
  use config_dir <- result.try(file.user_config_dir())

  use theme_dirpath <- result.try(
    file.build_path([config_dir, "zellij", "themes"])
    |> file.make_dirpath,
  )

  let filepath = file.build_path([theme_dirpath, "barista.kdl"])
  let contents = string_tree.to_string(contents)

  file.write(filepath, contents)
  |> result.map_error(fn(_) { snag.new("Could not write file: " <> filepath) })
}

/// Runs `touch` on Zellij main config file to trigger a reload.
pub fn apply_theme() -> snag.Result(Nil) {
  use config_dir <- result.try(file.user_config_dir())
  let filepath = file.build_path([config_dir, "zellij", "config.kdl"])

  let cmd = "touch"

  let result =
    child_process.from_name(cmd)
    |> child_process.args([filepath])
    |> child_process.run(stdio.capture(True))

  case result {
    Ok(_) -> Ok(Nil)
    Error(_) -> snag.error("Process exited with error (" <> cmd <> ")")
  }
}
