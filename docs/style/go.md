# Writing Go in Barista

A guide to writing idiomatic Go for this codebase, distilled from [Effective Go](https://go.dev/doc/effective_go), the Go team's [Code Review Comments](https://go.dev/wiki/CodeReviewComments), the standard library, and the patterns established by the Bubble Tea / Charm ecosystem (`bubbletea`, `bubbles`, `lipgloss`, `huh`, `glamour`) and the cobra / Kubernetes lineage that shaped most modern CLI tooling.

Barista is a CLI written in Go that reads TOML flavor palettes, renders Mustache templates, writes theme files for terminal applications (Ghostty, Neovim, Zellij), and triggers a reload in each. There is no server, no RPC, no network boundary — the "boundaries" that shape Barista's Go design are the filesystem (reading flavors, writing themes) and the child-process calls that reload each application (`pgrep`/`kill`, `nvim --server`, `touch`). A Bubble Tea TUI is a future possibility, not current scope; the TUI section below is forward-looking so the conventions are settled before any TUI code lands.

Read §§1–4 before writing your first package. Jump to whichever layer your task touches, then come back to §16 for our local conventions and §17 for the open questions where the ecosystem hasn't converged. When in doubt, prefer stdlib idioms over inventing new ones.

---

## 1. Project layout

A Go module has a conventional shape. The Go team's [`golang-standards/project-layout`](https://github.com/golang-standards/project-layout) is *not* an official standard, but most popular CLIs converge on a subset of it. Barista uses this:

```
go.mod                              # module manifest
go.sum                              # locked deps (committed)
cmd/
  barista/
    main.go                         # CLI binary entry point — tiny, just wires deps and calls into internal/
internal/                           # private code; nothing outside this module can import these packages
  cli/                              # cobra command tree
    root.go
    apply.go
  flavor/                           # Flavor, Palette types; Load reads + parses flavor.toml
  paths/                            # XDG dir resolution, flavors/data/config paths, filepath join
  template/                         # Mustache render wrapper
  recipe/                           # Recipe interface
    ghostty/
    neovim/
    zellij/
  nvim/                             # Neovim socket discovery + nvim --server --remote-send (replaces priv/scripts/nvim_send.sh)
testdata/                           # golden files, fixtures (Go reserves this directory name)
```

`go.mod`:

```
module github.com/ngscheurich/barista

go 1.27

require (
    github.com/spf13/cobra v1.8.0
    github.com/BurntSushi/toml v3.x
    github.com/cbroglie/mustache v1.x
)
```

The toolchain is pinned in `mise.toml` at the repo root (TBD — confirm whether Barista uses mise). Use `mise install` to match.

Common commands:

| What | Command |
| --- | --- |
| Build | `go build ./...` |
| Build the CLI binary | `go build -o bin/barista ./cmd/barista` |
| Run | `go run ./cmd/barista apply <theme>` |
| Test | `go test ./...` |
| Test with race detector | `go test -race ./...` |
| Format | `gofmt -w .` or `goimports -w .` |
| Vet | `go vet ./...` |
| Lint | `golangci-lint run` |
| Update deps | `go get -u ./... && go mod tidy` |

`gofmt` (and its strict superset `goimports`, which also fixes import grouping) is non-negotiable, has no options, and is enforced by CI. Like Gleam's formatter, this is on purpose: it eliminates bikeshedding, and every Go project looks the same to a new reader.

### Why `internal/`?

A directory named `internal` is special to the `go` tool: only packages rooted at the parent of `internal/` may import it. For a single-module project that means *only this repo's code* can import `internal/...`. This is the cheapest possible way to keep our domain types from leaking into anyone else's dependency tree and to give ourselves freedom to break internal APIs.

The rule of thumb: **default everything to `internal/`**. Promote a package to `pkg/` only when you have a concrete external consumer and have decided to make stability commitments to them. Barista has none today.

## 2. Packages and imports

A Go file declares its package in the first non-comment line. Every file in the same directory must share that package name. There is no other naming knob — the directory *is* the package.

```go
// Package flavor defines the Flavor domain type and its loading from disk.
//
// A Flavor is a named collection of 26 Catppuccin colors (a Palette) read
// from a flavor.toml file under the flavors directory.
package flavor

import (
    "fmt"
    "os"

    "github.com/BurntSushi/toml"

    "github.com/ngscheurich/barista/internal/paths"
)
```

Conventions, in order of preference:

1. **`goimports` groups imports into three blocks** separated by blank lines: standard library, third-party, this module. Don't fight this — set up your editor to run `goimports` on save.
2. **No dot imports** (`import . "fmt"`). They obscure where identifiers come from.
3. **No bare aliases for stylistic reasons.** Alias only on genuine name collision or when the import path's last segment doesn't match the package name (rare; usually a sign of a poorly-named third-party module).
4. **Side-effect imports** (`import _ "embed"`) get their own block at the bottom with the `_` blank identifier and a brief comment explaining why.

### Package naming

- **Lowercase, single word, no underscores or mixedCaps.** `flavor`, not `Flavor` or `flavor_loader`.
- **Singular.** `recipe`, not `recipes`. The package is one concept; the slice type is plural.
- **No stutter with the type it owns.** A `flavor` package exports `Flavor`, not `FlavorFlavor`. It exports `Load`, not `LoadFlavor` (called as `flavor.Load(...)`).
- **Avoid grab-bag names.** `util`, `common`, `helpers`, `misc` are smells — they grow without bound and tell the reader nothing. If you need three string helpers, put them in the package that already uses them, or inline.

## 3. Naming

Hard rules from the compiler:

- **Exported identifiers start with an uppercase letter**: `Flavor`, `Load`, `Palette`. Anything that begins lowercase is package-private.
- **No other visibility modifiers.** No `public` / `private` keywords; just case.

Strong conventions (Go team and stdlib):

- **MixedCaps** for multi-word identifiers, not `snake_case` or `kebab-case`. `Subtext1`, not `subtext_1` or `Subtext_1`. (TOML keys stay `snake_case` — see §9 — but Go identifiers that wrap them are MixedCaps.)
- **Initialisms keep their case.** `ID`, `URL`, `JSON`, `HTTP`, `UDS`. So `FlavorID`, `parseTOML`. *Mixed case versions like `Id` or `Url` are wrong* — `go vet` and most linters will flag this.
- **Receivers are short**: `f *Flavor`, `r *Recipe`. One or two letters, the same letter every time the same type is the receiver. Never `self` or `this`.
- **Getters drop the `Get` prefix.** A getter for the `name` field is `Name()`, not `GetName()`. Setters keep `Set`: `SetName(s string)`.
- **Errors**: variables of type `error` start with `Err` (`ErrFlavorNotFound`); error types end in `Error` (`type ParseError struct{...}`).
- **Interfaces named after the method**, with `-er` suffix when there's one method: `Reader`, `Writer`, `Stringer`. Multi-method interfaces describe the role: `Recipe`, `FlavorSource`.
- **Boolean-returning functions are predicates** starting with `Is`, `Has`, `Can`, or read as one: `IsEmpty()`, `HasTemplate()`, `CanApply()`. A method named `Applied()` returning `bool` is fine; `GetApplied()` is not.
- **Constructors are `New` or `NewT` where `T` disambiguates.** A `recipe` package that builds different recipe kinds exports `recipe.NewGhostty(...)`, `recipe.NewNeovim(...)`.

Avoid stuttering across package boundaries: `flavor.Flavor` reads as `flavor.Flavor` from outside the package, which is fine but mildly stuttery. Where a type is the package's central concept a little stutter is accepted (the stdlib has `time.Time`, `context.Context`); resist it for secondary types.

## 4. Functions, methods, and receivers

```go
// Load reads and parses the flavor named dirname from the flavors directory,
// returning a Flavor ready to be rendered against a template.
func Load(dirname string) (Flavor, error) {
    // ...
}
```

Conventions:

- **`context.Context` is the first argument**, conventionally named `ctx`. Functions that do I/O, that may block, or that should be cancellable take a context. Pure functions (string parsing, palette construction from known values) don't. Most of Barista's I/O (file reads, child processes) is short-lived and synchronous; reach for context when a caller might want to cancel or bound it.
- **`error` is the last return value.** Always. `(T, error)`, never `(error, T)`.
- **Multiple return values are normal**, but if you find yourself returning four or five, define a struct.
- **No named return values** except as documentation for `(int, int, error)`-style returns where the meaning isn't obvious from types. Naked `return` (relying on named returns) is a smell outside of very short functions; it makes the flow harder to follow.

### Receivers: value vs pointer

The choice is mostly mechanical:

- **Use a pointer receiver if the method mutates the receiver**, or if the type contains a `sync.Mutex` / other non-copyable field, or if the type is large.
- **Use a value receiver only for small, immutable, value-like types** — a `Palette` is 26 strings and immutable after construction, so value semantics may fit; a `Flavor` that owns a `Palette` plus a name and dirname is borderline and §17.7 decides the default.
- **Be consistent within a type**: if any method needs a pointer receiver, *all* methods on that type take a pointer receiver. Mixing the two is a documented Go gotcha — value receivers on a pointer-receiver type silently copy.

```go
// Good — all methods on *Flavor take a pointer.
func (f *Flavor) Reload() error
func (f *Flavor) Name() string   // even read-only methods stay on the pointer
```

### Functional options

For constructors with more than two or three configurable knobs, use **functional options** rather than a config struct or a long parameter list:

```go
type Recipe struct { /* ... */ }

type Option func(*Recipe)

func WithTemplatePath(p string) Option { return func(r *Recipe) { r.templatePath = p } }

func NewGhostty(opts ...Option) *Recipe {
    r := &Recipe{app: "ghostty"}
    for _, opt := range opts {
        opt(r)
    }
    return r
}
```

This pattern gives callers a future-proof API: adding a new option is non-breaking, defaults are explicit, and required arguments stay positional. *Don't* reach for it when you only have one optional knob — a second function or a small `Config` struct is fine.

## 5. Structs and embedding

Structs are the workhorse. Model domain entities as structs with named, typed fields:

```go
type Flavor struct {
    Name    string
    Dirname string
    Palette Palette
}

type Palette struct {
    Rosewater string
    Flamingo  string
    Pink      string
    Mauve     string
    // ... all 26 Catppuccin colors
    Crust string
}
```

- **Zero values should be useful.** A `Flavor{}` should not panic on common operations; if it must be constructed through `Load`, document that and consider making the fields unexported where practical.
- **Composition over inheritance**: Go has no inheritance. Embedding (a field whose type is itself a struct or interface, declared without a field name) gives you method promotion. Use it sparingly — it works well for *capability composition* (mixing in a `sync.Mutex`, a logger) but poorly for *domain modelling*. When you find yourself reaching for it to express "a Ghostty recipe is a kind of Recipe," use an explicit field or an interface instead.

### Struct literals

Always use **keyed struct literals**: `Flavor{Name: n, Dirname: d}`, not `Flavor{n, d}`. Positional literals break silently when a field is added or reordered. `go vet` enforces this for stdlib types; we enforce it everywhere via `golangci-lint`.

### Field tags and TOML

`BurntSushi/toml` uses `toml:"..."` tags to map struct fields to TOML keys:

```go
type flavorFile struct {
    Name    string         `toml:"name"`
    Palette tomlPalette    `toml:"palette"`
}

type tomlPalette struct {
    Rosewater string `toml:"rosewater"`
    Flamingo  string `toml:"flamingo"`
    // ...
}
```

Use an unexported `flavorFile` struct that mirrors the TOML shape, then build the exported `Flavor`/`Palette` from it. This keeps the on-disk format decoupled from the in-memory domain type — if `flavor.toml` gains a field later, the domain type doesn't have to change shape in lockstep. See §9.

## 6. Interfaces (consumer-side)

Go interfaces are structural — any type with the right methods satisfies the interface, no `implements` declaration needed. The standard library makes heavy use of this with tiny, single-method interfaces (`io.Reader`, `io.Writer`, `fmt.Stringer`).

**Define interfaces where they are consumed, not where they are implemented.** The `cli` package that needs to run recipes defines its own minimal interface:

```go
// package cli
type recipeRunner interface {
    Run(f flavor.Flavor) error
}
```

The concrete implementation lives in `internal/recipe/ghostty/ghostty.go` and exports `*Recipe` — no interface declaration on the *producer* side. This is the most-cited difference from Java/C# practice and it is genuinely important:

- It keeps interfaces small (only the methods the caller actually needs).
- It avoids speculative abstractions ("I might want to swap this out one day").
- It makes mocking trivial — the test in `internal/cli` defines its own fake.

Concrete corollary: **a Go function should usually accept interfaces and return concrete types.** Accepting an interface lets callers pass anything; returning `*Recipe` gives callers the full API. Returning an interface is appropriate when the concrete type is genuinely an implementation detail.

The bar for declaring an interface is: *there are, today, two or more concrete implementations, or there will be in the next change.* Barista has three recipes (Ghostty, Neovim, Zellij), so a `Recipe` interface is justified at the `cli`/orchestration seam. Otherwise just use the concrete type.

## 7. Error handling

Go has no exceptions. Errors are values, returned alongside results, checked explicitly.

```go
f, err := flavor.Load(dirname)
if err != nil {
    return fmt.Errorf("apply %s: %w", dirname, err)
}
```

### 7.1 The `if err != nil` block

The single most common piece of Go code. Read it as a return-on-error short-circuit, the rough equivalent of Gleam's `use _ <- result.try(...)`. Don't try to make it less verbose with helpers; the language has rejected several proposals to do so, and the explicitness is a feature.

### 7.2 Wrapping with `%w` (Go 1.13+)

`fmt.Errorf("context: %w", err)` produces a new error that *wraps* the original. Wrap when you cross a layer boundary and want to add context but preserve the underlying error for `errors.Is` / `errors.As`:

```go
func loadFlavor(dirname string) (flavor.Flavor, error) {
    raw, err := os.ReadFile(filepath.Join(flavorsDir, dirname, "flavor.toml"))
    if err != nil {
        return flavor.Flavor{}, fmt.Errorf("read flavor %s: %w", dirname, err)
    }
    // ...
}
```

The verb is `%w`, not `%s` or `%v`. `%w` preserves the chain; `%s` flattens it to a string and loses `errors.Is` support.

### 7.3 `errors.Is` and `errors.As`

```go
// Sentinel comparison
if errors.Is(err, ErrFlavorNotFound) { return nil }

// Typed extraction
var pe *toml.ParseError
if errors.As(err, &pe) {
    return fmt.Errorf("malformed flavor.toml at line %d", pe.Line)
}
```

Use these for *behaviour* (not-found, missing-template, reload-failed). Don't write `if err.Error() == "foo"` — string-comparing errors is brittle, and `errors.Is` is the supported idiom.

### 7.4 Sentinel errors vs typed errors

Two patterns, both common, both fine in their place:

**Sentinel** — a package-level error value, compared with `errors.Is`:

```go
var ErrFlavorNotFound = errors.New("flavor not found")

if errors.Is(err, flavor.ErrFlavorNotFound) { ... }
```

Use for *categorical* failures where the caller only needs to know "this happened" — `io.EOF`, `os.ErrNotExist`, `context.Canceled`.

**Typed** — a struct that satisfies `error`, with fields the caller might want:

```go
type ReloadError struct {
    App  string
    PID  string
    Cmd  string
}

func (e *ReloadError) Error() string {
    return fmt.Sprintf("reload %s: %s failed", e.App, e.Cmd)
}
```

Use when the caller wants *details* — which app, which pid, which command. Extract with `errors.As`.

In practice a single subsystem often exports both: sentinels for the cheap cases, typed errors for the rich cases. Don't agonise — start with sentinels, promote to typed when a caller actually needs structured data.

### 7.5 Aggregated errors across recipes

Barista runs all three recipes and reports every error at the end (a deliberate change from the Gleam version's short-circuit). The orchestration layer collects errors into a slice and, if non-empty, returns them wrapped so the CLI can print each. Go's stdlib has `errors.Join` (Go 1.20+) for joining multiple errors into one that `errors.Is` traverses; use it, or build a small typed `MultiError` if individual errors need to be iterated for per-recipe reporting:

```go
var errs []error
for _, r := range recipes {
    if err := r.Run(f); err != nil {
        errs = append(errs, fmt.Errorf("%s: %w", r.App(), err))
    }
}
if len(errs) > 0 {
    return errors.Join(errs...)
}
```

### 7.6 `panic` and `recover`

`panic` is for *programmer errors* — invariants the code itself violates, like indexing past a slice or dereferencing a nil pointer. Never use it for control flow, and never use it to signal "expected" failures (missing flavor files, TOML parse errors, reload failures).

**Reserve `panic` for `main` and `init`-time impossibilities**, paired with a comment that makes the invariant explicit. `recover` is even rarer — the only legitimate use is at goroutine boundaries to keep one bad worker from killing the program.

### 7.7 Don't drop errors

```go
// Wrong
result, _ := flavor.Load(dirname)

// Right
result, err := flavor.Load(dirname)
if err != nil { /* handle or propagate */ }
```

The blank identifier silences the compiler but lies to the reader. The only legitimate uses of `_` for an error are:

- Closing a `*os.File` you only opened to read (when the close error genuinely doesn't matter — and even then, log it).
- Type assertions you've already validated.

If an error is truly safe to ignore, write a comment saying why.

## 8. Concurrency: goroutines, channels, context

Barista is mostly serial: load a flavor, then run three recipes in sequence. There is no long-running server and no event loop today. So most code should *not* spawn goroutines directly — but when concurrency is introduced (running the three recipes in parallel is a plausible future change), follow these rules.

### 8.1 Goroutines have owners

Every goroutine you launch must have a clear answer to two questions: *how does it stop*, and *who is waiting for it?*

- **How it stops**: a `context.Context` it watches, a channel close, or a finite amount of work it will complete on its own.
- **Who waits**: a `sync.WaitGroup`, an `errgroup.Group`, or a result channel the caller reads.

A goroutine without both is a leak. `go func() { for { ... } }()` with no exit condition is almost always wrong.

### 8.2 `context.Context` for cancellation

`context.Context` propagates cancellation signals through call chains. Pass it explicitly; never store it in a struct except for the rare top-of-program long-lived context.

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

return reloadGhostty(ctx)
```

- **Always call `cancel`** — `defer cancel()` is the right pattern even when the context will expire by timeout, because cancelling early releases resources.
- **Don't pass `nil`** as a context; use `context.TODO()` to mark "I don't know yet" or `context.Background()` for genuine top-of-program.
- **Don't put values in context** that aren't request-scoped. `context.Value` is not a parameter-passing mechanism.

### 8.3 Parallel recipes (future)

If the three recipes ever run concurrently, `golang.org/x/sync/errgroup` is the clean shape — it gives cancellation + first-error collection in one type, and pairs naturally with the aggregated-error strategy in §7.5:

```go
g, ctx := errgroup.WithContext(ctx)
for _, r := range recipes {
    r := r
    g.Go(func() error { return r.Run(ctx, f) })
}
return g.Wait()
```

Until then, the serial loop is correct and simpler.

### 8.4 `sync` primitives

For shared mutable state, `sync.Mutex` (or `sync.RWMutex`) is the standard. Keep critical sections short. `sync.Once` for one-shot initialisation. `sync.WaitGroup` for fan-in.

### 8.5 Race detector

`go test -race ./...` catches data races at test time. Run it in CI on every change. The flag costs ~5–10× CPU and memory; that's fine for tests, never for production builds.

## 9. TOML and Mustache

### 9.1 TOML

`BurntSushi/toml` decodes into a struct with `toml:"..."` tags. Decode into an unexported mirror struct, then build the domain type:

```go
type flavorFile struct {
    Name    string      `toml:"name"`
    Palette paletteFile `toml:"palette"`
}

type paletteFile struct {
    Rosewater string `toml:"rosewater"`
    Flamingo  string `toml:"flamingo"`
    Pink      string `toml:"pink"`
    Mauve     string `toml:"mauve"`
    Red       string `toml:"red"`
    Maroon    string `toml:"maroon"`
    Peach     string `toml:"peach"`
    Yellow    string `toml:"yellow"`
    Green     string `toml:"green"`
    Teal      string `toml:"teal"`
    Sky       string `toml:"sky"`
    Sapphire  string `toml:"sapphire"`
    Blue      string `toml:"blue"`
    Lavender  string `toml:"lavender"`
    Text      string `toml:"text"`
    Subtext1  string `toml:"subtext_1"`
    Subtext0  string `toml:"subtext_0"`
    Overlay2  string `toml:"overlay_2"`
    Overlay1  string `toml:"overlay_1"`
    Overlay0  string `toml:"overlay_0"`
    Surface2  string `toml:"surface_2"`
    Surface1  string `toml:"surface_1"`
    Surface0  string `toml:"surface_0"`
    Base      string `toml:"base"`
    Mantle    string `toml:"mantle"`
    Crust     string `toml:"crust"`
}
```

Note the mapping: TOML keys are `snake_case` (`subtext_1`), Go identifiers are MixedCaps (`Subtext1`). The tag bridges them.

### 9.2 Mustache

`cbroglie/mustache` renders a template string against a data value. The data value can be a struct, a map, or a combination — Mustache resolves fields by name. For Barista, build the context as a struct or map matching the template's `{{name}}` and `{{palette.rosewater}}` references:

```go
type templateContext struct {
    Name    string
    Palette map[string]string
}

func renderTemplate(tmpl string, f flavor.Flavor) (string, error) {
    ctx := templateContext{
        Name:    f.Name,
        Palette: f.Palette.AsMap(),
    }
    rendered, err := mustache.RenderString(tmpl, ctx)
    if err != nil {
        return "", fmt.Errorf("render template: %w", err)
    }
    return rendered, nil
}
```

`cbroglie/mustache` uses an idiomatic error-returning API (unlike its parent fork), so handle errors at every call.

## 10. External processes and file I/O

This is Barista's actual boundary — the filesystem and the child processes that reload each app. There is no network, no RPC, no socket protocol (except the Neovim UDS discovery, which is filesystem path enumeration plus `nvim --server` invocations, not a protocol we speak).

### 10.1 File I/O

Prefer `os.ReadFile` / `os.WriteFile` for whole-file operations and `os.MkdirAll` for directory creation (the Go equivalent of `mkdir -p`, replacing the Gleam version's `child_process` call to `mkdir`). Check existence with `os.Stat` and `errors.Is(err, os.ErrNotExist)` rather than a separate `is_file` call:

```go
func themeFilepath(flavorsDir, dirname, filename string) (string, error) {
    p := filepath.Join(flavorsDir, dirname, filename)
    info, err := os.Stat(p)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return "", fmt.Errorf("%s: not a file", p)
        }
        return "", fmt.Errorf("stat %s: %w", p, err)
    }
    if info.IsDir() {
        return "", fmt.Errorf("%s: not a file", p)
    }
    return p, nil
}
```

Use `filepath.Join`, not string concatenation with `"/"` — it handles platform separators and cleaning, where the Gleam version's `build_path` did not.

### 10.2 Child processes

`os/exec` is the stdlib way to run external commands. Always set up the command, then run it; check the error, distinguishing `*exec.ExitError` for non-zero exit codes:

```go
func reloadGhostty() error {
    out, err := exec.Command("pgrep", "ghostty").Output()
    if err != nil {
        return fmt.Errorf("pgrep ghostty: %w", err)
    }
    pid := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
    if err := exec.Command("kill", "-s", "USR2", pid).Run(); err != nil {
        return fmt.Errorf("kill -USR2 %s: %w", pid, err)
    }
    return nil
}
```

Conventions:

- **`exec.Command(name, args...)`** builds the command; `.Run()` waits for completion and errors on non-zero exit, `.Output()` captures stdout and is the right call when you need the output (like `pgrep`).
- **Don't shell out via `sh -c`** when you can call the binary directly. `exec.Command("kill", "-s", "USR2", pid)` is safer than `exec.Command("sh", "-c", "kill -s USR2 "+pid)` — no shell injection, no quoting bugs.
- **Capture or discard stderr deliberately.** `cmd.Stderr = &buf` when you want it for an error message; otherwise it goes nowhere.
- **Set up `cmd.Env` / `cmd.Dir` explicitly** if the command cares — child processes inherit the parent's environment by default, which is what Barista wants for `pgrep`/`kill`/`nvim`/`touch`.

### 10.3 Neovim socket discovery

The Gleam version's `priv/scripts/nvim_send.sh` walks `$XDG_RUNTIME_DIR` (fallback `$TMPDIR/nvim.<user>`) for `nvim.*.0` sockets and calls `nvim --server <socket> --remote-send` on each. Reimplement this in Go under `internal/nvim/`:

```go
func DiscoverSockets() ([]string, error) {
    dir := os.Getenv("XDG_RUNTIME_DIR")
    if dir == "" {
        dir = filepath.Join(os.TempDir(), "nvim."+os.Getenv("USER"))
    }
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, fmt.Errorf("read nvim runtime dir %s: %w", dir, err)
    }
    var sockets []string
    for _, e := range entries {
        if strings.HasPrefix(e.Name(), "nvim.") && strings.HasSuffix(e.Name(), ".0") {
            sockets = append(sockets, filepath.Join(dir, e.Name()))
        }
    }
    return sockets, nil
}
```

The `nvim.*.0` glob and the fallback path are the two details to preserve exactly — they are the contract with Neovim's server discovery, and getting either wrong silently makes reloads no-ops.

## 11. CLI patterns: cobra

Barista uses [cobra](https://github.comgithub.com/spf13/cobra), which underpins kubectl, helm, gh, and most other modern Go CLIs. Cobra's model is a tree of `*cobra.Command`s, each with its own flags, args, and `RunE`.

```go
// internal/cli/root.go

func NewRoot() *cobra.Command {
    cmd := &cobra.Command{
        Use:           "barista",
        Short:         "Serves up a new <theme> for your terminal apps.",
        SilenceUsage:  true,  // don't print usage on RunE error
        SilenceErrors: true,  // we print errors ourselves
    }
    cmd.AddCommand(newApplyCmd())
    return cmd
}
```

The `apply` command:

```go
// internal/cli/apply.go

func newApplyCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "apply <theme>",
        Short: "Apply a flavor to all configured applications",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            return apply.Run(ctx, args[0])
        },
    }
}
```

Conventions:

- **One file per top-level command tree** (`root.go`, `apply.go`).
- **`RunE`, not `Run`.** `RunE` returns an error; cobra prints it and exits non-zero. `Run` panics if you try to fail.
- **`cmd.Context()` is your context**, threaded from `cmd.ExecuteContext(ctx)` in `main`.
- **`cmd.OutOrStdout()` and `cmd.OutOrStderr()`** — never reach for `os.Stdout` / `os.Stderr` directly from a `RunE`. Going through the cobra accessor makes the command testable (you can pass a `*bytes.Buffer`).
- **Flags bound to local vars in a closure**; avoid package-global flag variables.
- **Required flags**: `cmd.MarkFlagRequired(name)`. Don't validate "is this empty?" in `RunE` if the framework can.
- **`SilenceUsage: true`** on commands that do real work — printing a giant usage block on a runtime error is noise.

### Output

For human-facing output (the `☕︎ Served up <name>` line, error messages), write to `cmd.OutOrStdout()` / `cmd.OutOrStderr()`. The CLI's exit code matters: `0` success, non-zero on any error, with cobra setting that automatically when `RunE` returns non-nil. The aggregated-error strategy (§7.5) means `RunE` returns one joined error that the top-level prints; or the `apply` package prints each error itself and returns a sentinel to force a non-zero exit — pick one and keep it consistent.

## 12. TUI patterns: Bubble Tea (forward-looking)

A `barista tui` command is a future possibility, not current scope. This section is written now so the conventions are settled before any TUI code lands.

Bubble Tea is the Elm Architecture (Model–Update–View) for terminals. Every component is a value satisfying `tea.Model`:

```go
type Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (Model, tea.Cmd)
    View() string
}
```

- **`Model`** is your state — a plain Go struct, by value. Updates *return a new Model* rather than mutating; this is enforced by the value-typed signature.
- **`Update`** is a giant `switch` on the `tea.Msg` type.
- **`View`** is a pure function from model to string. No side effects, no I/O. Render with [lipgloss](https://github.com/charmbracelet/lipgloss).
- **`tea.Cmd`** is `func() tea.Msg` — a unit of asynchronous work the runtime will execute and feed back as a message.

Rules of thumb when the TUI arrives:

- **All I/O lives inside a `tea.Cmd`.** Never call file reads or child processes directly from `Update` or `View` — they run on the main loop and will block the UI.
- **Messages are values.** Define a named type per kind of message; don't reuse anonymous structs.
- **One `Model` per screen**, composed by a top-level `app` model that delegates to the focused child. The `bubbles` package provides reusable building blocks (list, table, textinput, viewport, spinner).
- **Forms use [huh](https://github.com/charmbracelet/huh)**, which composes with Bubble Tea models naturally.
- **Styling stays in `View`**, in per-component `lipgloss.Style` package-level vars. Don't compute styles inside the render loop; build them once and reuse.
- **`Update` and `View` are pure**, so they are trivial to unit-test: call `Update` with a message, assert on the returned model; call `View`, assert on the string.

## 13. Testing

The standard library's `testing` package is sufficient. Conventions:

- **`*_test.go` files live alongside the code they test**, in the same package (or in `package foo_test` for black-box testing).
- **Test functions are `func TestXxx(t *testing.T)`**.
- **`t.Run("name", func(t *testing.T) { ... })`** for subtests — gives per-case failures and selective re-running with `go test -run TestSomething/case_one`.
- **`t.Parallel()`** at the top of any test that doesn't share state with siblings.
- **`t.Helper()`** in any helper function that calls `t.Fatal` / `t.Error`, so failure messages point at the caller.
- **`t.Cleanup(fn)`** instead of `defer` for teardown; runs even when a parent test fails.
- **`testing.T.TempDir()`** and **`testing.T.Context()`** (Go 1.24+) — let the framework manage scratch directories and cancellation. Barista's tests should lean on `TempDir` heavily: build a fake flavors directory, write a `flavor.toml` and templates into it, point `paths` at it, and run `apply` end-to-end without touching the user's real config.

### Table-driven tests

The dominant style. One test function, a slice of cases, a `t.Run` per case:

```go
func TestPaletteFromTOML(t *testing.T) {
    cases := []struct {
        name    string
        input   string
        want    flavor.Palette
        wantErr bool
    }{
        {"full palette", validPaletteTOML, expectedPalette, false},
        {"missing color", missingColorTOML, flavor.Palette{}, true},
        {"wrong type", wrongTypeTOML, flavor.Palette{}, true},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            // ...
        })
    }
}
```

### Assertion style — open question

ADR-0001 records "stdlib `testing` + `testify/assert`." The Go team explicitly recommends *against* `testify` (over-stuffed, encourages bad patterns like `suite`), and `golangci-lint` defaults flag it. The lean in this guide is stdlib `if got != want { t.Errorf(...) }` plus `github.com/google/go-cmp/cmp` for struct diffs:

```go
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("Flavor mismatch (-want +got):\n%s", diff)
}
```

**This conflicts with ADR-0001 and must be reconciled** — see §17.2. Until it's decided, follow the ADR (testify) if you want to match the recorded decision, or raise it and flip the ADR.

### Golden files

For tests of template output or generated theme files, write the expected output to `testdata/<name>.golden` and compare. Add a `-update` flag to regenerate on changes:

```go
var update = flag.Bool("update", false, "update golden files")

if *update {
    os.WriteFile(goldenPath, got, 0644)
}
want, _ := os.ReadFile(goldenPath)
if !bytes.Equal(got, want) { /* fail with diff */ }
```

### Testing the reload steps

The reload steps shell out to external processes (`pgrep`, `kill`, `nvim`, `touch`) and are harder to unit-test. Two seams:

- **Extract the command construction** into a function that returns `*exec.Cmd` without running it, so a test can assert the command and args without spawning a process.
- **Integration test the full `apply`** against a `TempDir` flavors directory, mocking or skipping the reload step via a flag or a seam (`var reloadFunc = reloadGhostty` that the test swaps).

## 14. Logging

Use the standard library's `log/slog` (Go 1.21+). Structured, levelled, contextual. Barista is a short-lived CLI, so logging is light — most output is the user-facing `☕︎ Served up <name>` line and errors, not a log stream. Reach for `slog` for diagnostic output that isn't user-facing (e.g. behind a `--debug` flag).

```go
import "log/slog"

logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

logger.Debug("resolved flavors dir", "path", flavorsDir)
```

Conventions:

- **Text handler for a CLI** (human-readable); JSON handler only if some tool consumes the output.
- **Levels**: `Debug` (verbose dev), `Info` (normal events), `Warn` (degraded but functional), `Error` (operation failed). No `Fatal` — errors should propagate to a caller that can decide.
- **Key conventions**: `snake_case`, stable across the codebase.
- **No secrets in logs.**

## 15. Documentation comments

Godoc renders any comment *immediately preceding* a top-level declaration as that declaration's documentation. Conventions:

- **Package comment** on the file that declares the package, starts with `// Package <name> `:

  ```go
  // Package flavor defines the Flavor domain type and its loading from disk.
  //
  // A Flavor is a named collection of 26 Catppuccin colors (a Palette) read
  // from a flavor.toml file under the flavors directory.
  package flavor
  ```

- **Exported declarations** start their doc comment with the identifier name:

  ```go
  // Load reads and parses the flavor named dirname from the flavors
  // directory, returning a Flavor ready to be rendered against a template.
  func Load(dirname string) (Flavor, error) { ... }
  ```

  Starting with the identifier matters because tools (`go doc`, IDE hovers) display the first sentence as a one-liner.

- **Doc links** use bracket notation: `[Flavor.Palette]`, `[paths.FlavorsDir]`. Renders as a link on pkg.go.dev and in modern IDEs.
- **Examples** (`func ExampleLoad()` etc.) are surfaced inline in godoc and double as tests with an `// Output:` comment.
- **Don't document obvious things.** `// Flavor represents a flavor.` is worse than no comment; delete it. Document *why* and *invariants*, not *what the name already says*.
- **If a comment explains the name, the name is wrong.** Rename the declaration rather than annotate it.
- **Ordering-dependency comments state the dependency and stop.** When a comment survives because one statement must run before another, write it in that shape — `// Ensure the data dir before loading the flavor, so the recipe writes don't fail on a missing dir` — and nothing more.
- **No references to documents outside the code.** A comment never cites `docs/adrs/…`, `docs/specs/…`, or a bare `ADR-0001`. Compress any non-obvious *why* into the comment itself and drop the pointer — the reasoning is found from `docs/adrs/`, not a footnote in the source. A `[paths.FlavorsDir]` doc link to another Go symbol is fine; a link out to prose under `docs/` is not. Succinct by default, not by hard cap: keep the lines a real invariant needs, cut everything else. Unexported functions rarely earn a comment at all — name them well and leave them bare.
- **Wrap comment prose at 80 columns.** `gofmt` reflows code but leaves comment text as written, so wrap by hand. (See [`prose.md`](prose.md).)

## 16. Conventions to adopt for Barista

These are the locally-decided defaults. Override only with a comment justifying the divergence.

1. **`internal/` for everything by default.** Promote to `pkg/` only when a concrete external consumer exists.
2. **`goimports` (not just `gofmt`) on save.** Three import blocks, alphabetised within each.
3. **Package names are singular, lowercase, no underscores.** No `util`, `common`, `helpers`.
4. **MixedCaps everywhere, initialisms preserved.** `FlavorID`, `parseTOML`, never `FlavorId` or `Parse_toml`. TOML keys stay `snake_case`; the Go identifiers that wrap them are MixedCaps.
5. **`cobra` for the CLI** (`barista apply <theme>`).
6. **Functional options** for constructors with more than two optional arguments; small `Config` struct otherwise.
7. **Errors wrapped with `%w`** at every layer boundary. Wrap with a one-phrase prefix identifying *this* layer's role (`fmt.Errorf("read flavor %s: %w", dirname, err)`).
8. **Sentinel errors for categorical failures, typed errors for structured detail.** Start with sentinels; promote to typed only when a caller needs fields.
9. **Pointer receivers throughout a type if any method needs one.** Don't mix.
10. **No package-level mutable state** beyond logger and configuration.
11. **`stdlib testing` + `cmp.Diff` for assertions** (provisional — see §17.2; conflicts with ADR-0001).
12. **`log/slog` (stdlib), text handler for a CLI.**
13. **Doc comments on every exported identifier.** Start with the identifier name. **No references to `docs/`** — compress the *why* inline, drop the citation.
14. **No `init()`** outside `cmd/barista/main.go`. Hidden init is hidden control flow.
15. **`go vet ./...` and `golangci-lint run` are CI gates.** Locally: pre-commit hooks run both.
16. **Command nouns are singular; collections stay plural.** `barista apply <theme>` acts on one flavor; a directory listing many flavors is the plural `flavors`. A command invokes a verb on one noun; a collection holds many.

## 17. Open questions

Style points where the ecosystem is split, or where this repo hasn't decided. Decide and delete.

1. **`testify` vs stdlib `testing` + `cmp.Diff`.** ADR-0001 records `testify/assert`; this guide leans stdlib + `cmp.Diff` per the Go team's guidance. **These conflict and must be reconciled.** `testify` is more concise; stdlib is what the Go team and `golangci-lint` defaults recommend. ADR-worthy — flip the ADR one way or the other before the first test lands.

2. **`filepath.Join` vs preserving the Gleam `build_path` semantics.** The Gleam version joined with `"/"` and read unset env vars as `""`, producing malformed default paths. Go's `filepath.Join` is platform-correct and cleaner. The spec says replicate the empty-string-on-unset behavior for fidelity, but `filepath.Join` + `os.UserHomeDir()` / `os.UserConfigDir()` / `os.UserCacheDir()` is the idiomatic Go way to resolve XDG dirs. Decide whether "faithful port" means reproducing the v1 edge cases or adopting Go's proper XDG resolution — the latter is almost certainly right and may warrant a spec amendment.

3. **Mocking strategy.** Three viable patterns: (a) hand-written fakes in `_test.go` files; (b) generated mocks (`go.uber.org/mock`, `mockery`); (c) interface-free testing with seams (`var nowFunc = time.Now`). Small codebase suggests (a) for now.

4. **Logger placement.** Options: (i) one `*slog.Logger` passed through every constructor, (ii) `slog.Default()` as an ambient global with `slog.SetDefault` once in `main`, (iii) context-carried logger. The Go team's guidance is (i); (ii) is widespread. For a small CLI with light logging, (ii) may be pragmatic; pick.

5. **`context.Context` placement on cobra commands.** Cobra commands carry a `Context()` accessor, seeded via `cmd.ExecuteContext(ctx)` in `main`. Some projects stash request-scoped values (config, paths) in the context; others wire those explicitly. Wiring explicitly is cleaner; context-stashing is more cobra-idiomatic. Pick.

6. **Pointer vs value for domain types.** `*Flavor` vs `Flavor` in returned values, function parameters, and slice element types. `Flavor` owns a 26-field `Palette`, so it's not tiny — pointers avoid copying it on every pass. Lean toward value types for small immutable domain values (an `ID` wrapper) and pointers for `Flavor`/`Palette`. §4's receiver rule then follows: `Flavor` methods take pointer receivers.

7. **Build tags.** If integration tests that spawn real `pgrep`/`nvim`/`touch` processes land, they want build tags (`//go:build integration`) so `go test ./...` doesn't run them by default. Decide on a small set of canonical tags and document them.

8. **Generics.** Go 1.18+ supports type parameters; the stdlib uses them sparingly (`slices`, `maps`, `cmp`). Probably reach for the stdlib `slices`/`maps` packages and inline the rest rather than writing generic helpers. The temptation to write `MapErr` will be strong; resist unless a concrete duplication justifies it.

9. **Linter configuration.** `golangci-lint` ships with a dozen analysers; the defaults are conservative. Pick a profile — enable `errcheck`, `gocritic`, `gosec`, `revive` at least — and check it in as `.golangci.yml`. Pin a version; the tool changes defaults regularly.

10. **Cross-compilation.** Barista targets macOS and Linux. `GOOS`/`GOARCH` cross-compilation works out of the box for pure-Go code, but anything that links C (CGO) will fight us. Keep the binary CGO-free; if a dependency requires CGO, build per-platform in CI. A `goreleaser` config is the natural distribution story for release binaries.

11. **Module path.** `github.com/ngscheurich/barista` (per ADR-0001). If this moves to an org, every import rewrites — worth confirming before there are many imports to rewrite.

12. **`mise.toml` / toolchain pinning.** The Spiritualism guide pins the Go toolchain via `mise.toml`. Confirm whether Barista uses mise (or another version manager) and pin the toolchain before the first contributor hits a "works on my machine" mismatch.

---

## Appendix A: Quick reference

```go
// Package flavor defines the Flavor domain type and its loading from disk.
package flavor

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "github.com/BurntSushi/toml"
    "github.com/google/go-cmp/cmp"

    "github.com/ngscheurich/barista/internal/paths"
)

// Palette is the 26 Catppuccin color values inside a Flavor.
type Palette struct {
    Rosewater string
    Flamingo  string
    Pink      string
    Mauve     string
    // ... all 26 colors
    Crust string
}

// Flavor is a named collection of colors that maps to the Catppuccin palette
// spec.
type Flavor struct {
    Name    string
    Dirname string
    Palette Palette
}

// ErrNotFound is returned when a flavor directory or file is missing.
var ErrNotFound = errors.New("flavor not found")

// Load reads and parses the flavor named dirname from the flavors directory.
func Load(flavorsDir, dirname string) (Flavor, error) {
    p := filepath.Join(flavorsDir, dirname, "flavor.toml")
    raw, err := os.ReadFile(p)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return Flavor{}, fmt.Errorf("load flavor %s: %w", dirname, ErrNotFound)
        }
        return Flavor{}, fmt.Errorf("load flavor %s: %w", dirname, err)
    }
    var f flavorFile
    if err := toml.Unmarshal(raw, &f); err != nil {
        return Flavor{}, fmt.Errorf("parse flavor %s: %w", dirname, err)
    }
    return fromFile(dirname, f), nil
}
```

## Appendix B: Reading list

- [Effective Go](https://go.dev/doc/effective_go) — the closest thing to an official style document.
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) — terse, opinionated, the source of half the conventions in this guide.
- [Practical Go](https://dave.cheney.net/practical-go) by Dave Cheney — long-form opinions on package design, error handling, naming.
- [`cobra` user guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md) — command tree, flags, completion.
- [`BurntSushi/toml` docs](https://pkg.go.dev/github.com/BurntSushi/toml) — struct-tag decoding, `Marshaler`/`Unmarshaler`.
- [`cbroglie/mustache` docs](https://pkg.go.dev/github.com/cbroglie/mustache) — `RenderString`, error-returning API.
- [`pkg.go.dev/os/exec`](https://pkg.go.dev/os/exec) — the canonical child-process API.
- [`pkg.go.dev/log/slog`](https://pkg.go.dev/log/slog) — structured logging in the stdlib.
- [Bubble Tea tutorial](https://github.com/charmbracelet/bubbletea/tree/main/tutorials) — start here when the TUI lands.
- [Lipgloss README](https://github.com/charmbracelet/lipgloss) — the styling layer.
