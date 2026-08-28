# 04: Template rendering

**What to build:** The `internal/template` package — a `Render(tmpl string, f flavor.Flavor) (string, error)` function that renders a Mustache template against a flavor. Build the context as a struct or map with two top-level keys: `name` (the flavor's `Name` string) and `palette` (a `map[string]string` keyed by the 26 Catppuccin color names — `rosewater`, `flamingo`, …, `crust` — so templates can use `{{palette.rosewater}}` and the rest). Render via `cbroglie/mustache`'s `RenderString`, wrapping its error with a `render template:` prefix using `%w`. Unit-test against golden output: a template using `{{name}}` and several `{{palette.<color>}}` references renders to the expected string; an invalid template fails with a wrapped error. Consider a `testdata/` golden file for a representative template if the output is non-trivial. Follow `docs/style/go.md` — the `cbroglie/mustache` API returns errors (unlike its parent fork), so handle them at every call; do not drop errors.

**Blocked by:** 03 (flavor loading and parsing)

**Status:** ready-for-agent

- [ ] `internal/template.Render(tmpl, f)` renders a Mustache template against a flavor
- [ ] Context exposes `{{name}}` and `{{palette.<color>}}` for all 26 Catppuccin colors
- [ ] Uses `cbroglie/mustache`'s error-returning API; errors wrapped with `%w`
- [ ] Tests cover a valid template with `{{name}}` and nested `{{palette.<color>}}` access, and an invalid template
- [ ] `go build ./...`, `go test ./...`, `go vet ./...` pass
