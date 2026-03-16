import barista/file
import child_process
import child_process/stdio
import gleam/result
import gleam/string
import gleam/string_tree.{type StringTree}
import snag

/// Writes Ghostty config file to `$XDG_DATA_HOME/barista/ghostty`.
pub fn write_theme(contents: StringTree) -> snag.Result(Nil) {
  use data_dir <- result.try(file.barista_data_dir())

  let filepath = file.build_path([data_dir, "ghostty"])
  let contents = string_tree.to_string(contents)

  file.write(filepath, contents)
  |> result.map_error(fn(_) { snag.new("Could not write file: " <> filepath) })
}

// Sends `SIGUSR2` to the Ghostty process to trigger a config reload.
pub fn apply_theme() -> snag.Result(Nil) {
  use pid <- result.try(get_pid())

  case kill_pid(pid) {
    Ok(_) -> Ok(Nil)
    Error(error) -> Error(error)
  }
}

fn get_pid() -> snag.Result(String) {
  let cmd = "pgrep"

  let result =
    child_process.from_name("pgrep")
    |> child_process.args(["ghostty"])
    |> child_process.run(stdio.capture(True))

  case result {
    Ok(output) -> Ok(string.trim(output.output))
    Error(_) -> snag.error("Process exited with error (" <> cmd <> ")")
  }
}

fn kill_pid(pid: String) -> snag.Result(Nil) {
  let cmd = "kill"

  let result =
    child_process.from_name("kill")
    |> child_process.args(["-s", "USR2", pid])
    |> child_process.run(stdio.capture(True))

  case result {
    Ok(_) -> Ok(Nil)
    Error(_) -> snag.error("Process exited with error (" <> cmd <> ")")
  }
}
