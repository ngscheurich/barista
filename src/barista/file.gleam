import child_process
import child_process/stdio
import envoy
import gleam/list
import gleam/result
import gleam/string
import simplifile
import snag.{type Result}

/// Get the file path for a file belonging to a particular theme.
///
/// ## Example
/// ```gleam
/// theme_filepath("catppuccin-mocha", "neovim.lua.mustache")
/// ```
pub fn theme_filepath(theme_name: String, filename: String) -> Result(String) {
  use dir <- result.try(barista_config_dir())

  let filepath = build_path([dir, "flavors", theme_name, filename])

  case simplifile.is_file(filepath) {
    Ok(True) -> Ok(filepath)
    Ok(False) -> snag.error("Not a file (" <> filepath <> ")")
    Error(_) -> snag.error("Could not check file (" <> filepath <> ")")
  }
}

/// Read a file from disk.
pub fn read(filepath: String) -> Result(String) {
  simplifile.read(filepath)
  |> result.map_error(fn(_) { snag.new("Could not read file: " <> filepath) })
}

/// Write a file from disk.
pub fn write(filepath: String, contents: String) -> Result(Nil) {
  simplifile.write(filepath, contents)
  |> result.map_error(fn(_) { snag.new("Could not write file: " <> filepath) })
}

/// Build a file system path by joining a list of parts.
pub fn build_path(parts: List(String)) -> String {
  string.join(parts, "/")
}

/// Get the path to the user-specific data directory. Checks for
/// `$XDG_DATA_HOME` then falls back to `~/.local/share`.
pub fn user_data_dir() -> Result(String) {
  let xdg_path = get_env("XDG_DATA_HOME")
  let default_path = build_path([get_env("HOME"), ".local", "share"])

  [xdg_path, default_path]
  |> find_existing_dir
  |> snag.context("Could not get user data directory")
}

/// Get the path to the user-specific Barista data directory.
pub fn barista_data_dir() -> Result(String) {
  case user_data_dir() {
    Ok(dir) -> Ok(build_path([dir, "barista"]))
    Error(snag) ->
      snag.layer(snag, "Could not get Barista data directory") |> Error
  }
}

/// Get the path to the user-specific Barista config directory. Checks for
/// `$XDG_CONFIG_HOME` then falls back to `~/.config`.
pub fn user_config_dir() -> Result(String) {
  let xdg_path = get_env("XDG_CONFIG_HOME")
  let default_path = build_path([get_env("HOME"), ".config"])

  [xdg_path, default_path]
  |> find_existing_dir
  |> snag.context("Could not get user config directory")
}

/// Get the path to the user-specific Barista config directory.
pub fn barista_config_dir() -> Result(String) {
  case user_config_dir() {
    Ok(dir) -> Ok(build_path([dir, "barista"]))
    Error(snag) ->
      snag.layer(snag, "Could not get Barista config directory") |> Error
  }
}

fn get_env(var: String) -> String {
  result.unwrap(envoy.get(var), "")
}

fn find_existing_dir(paths: List(String)) -> Result(String) {
  list.filter(paths, fn(path) {
    result.unwrap(simplifile.is_directory(path), False)
  })
  |> list.first
  |> result.map_error(fn(_) { snag.new("Could not find an existing directory") })
}

/// Create the directory at `path`, including all intermediate directories.
pub fn make_dirpath(path: String) -> snag.Result(String) {
  let result =
    child_process.from_name("mkdir")
    |> child_process.args(["-p", path])
    |> child_process.run(stdio.capture(True))

  case result {
    Ok(_) -> Ok(path)
    Error(_) -> snag.error("Could not create directory: " <> path)
  }
}
