# 03: Flavor loading and parsing

**What to build:** The `internal/flavor` package — the `Flavor` and `Palette` domain types and the `Load` function that reads a flavor from disk. `Flavor` carries a `Name` string, a `Dirname` string, and a `Palette` of the 26 Catppuccin colors (rosewater through crust). `Load(flavorsDir, dirname)` reads `<flavorsDir>/<dirname>/flavor.toml`, decodes it via `BurntSushi/toml` into an unexported mirror struct with `toml:"..."` tags (TOML keys are `snake_case` like `subtext_1`; Go identifiers are MixedCaps like `Subtext1`), and builds the exported `Flavor` (with `Dirname` set to `dirname`). The `[palette]` table maps to a nested mirror struct. Define a sentinel `ErrNotFound` and return it wrapped with `%w` when the file is missing (use `errors.Is(err, os.ErrNotExist)`); wrap parse errors with a `parse flavor <dirname>:` prefix. Unit-test with table-driven cases: a valid full 26-color palette decodes to the expected `Flavor`; a missing color fails; a wrong-typed value fails; a missing file returns a wrappable `ErrNotFound`. Tests write `flavor.toml` into a `t.TempDir()` and point `Load` at it. Follow `docs/style/go.md` — unexported mirror struct decoupling the on-disk shape from the domain type, `%w` wrapping, sentinel + typed errors per the guide.

**Blocked by:** 02 (paths — XDG directory resolution)

**Status:** ready-for-agent

- [x] `internal/flavor` exports `Flavor` (Name, Dirname, Palette) and `Palette` (all 26 Catppuccin colors)
- [x] `Load(flavorsDir, dirname)` reads and parses `flavor.toml` via `BurntSushi/toml`
- [x] Unexported mirror struct with `toml:"..."` tags decouples on-disk shape from domain type
- [x] `snake_case` TOML keys map to MixedCaps Go identifiers (`subtext_1` → `Subtext1`)
- [x] Missing file returns a wrappable `ErrNotFound`; parse errors wrapped with `%w`
- [x] Table-driven tests cover valid palette, missing color, wrong type, missing file
- [x] `go build ./...`, `go test ./...`, `go vet ./...` pass
