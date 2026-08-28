# ADR-0002: Flavor format is TOML; templates are Mustache

Date: 2026-08-28
Status: Accepted

## Context

Barista reads a **flavor** data file (`flavor.toml` today) and renders **Mustache** templates against it. The rewrite to Go (ADR-0001) reopened both choices: we are not bound to preserve existing flavor files, and we want formats that are idiomatic and well-supported in the Go community.

Two independent decisions:

1. **Flavor data format** — TOML vs YAML vs JSON.
2. **Template engine** — Mustache vs Go `text/template` vs Handlebars.

## Considered options

### Flavor data format

Verified against the Go ecosystem (2026-08-28):

| Format | Library | Stars | Status |
| --- | --- | --- | --- |
| TOML | `BurntSushi/toml` | ~5,000 | Active |
| YAML | `go-yaml/yaml` v3 | ~7,000 | **Archived** (canonical but unmaintained) |
| YAML | `goccy/go-yaml` | ~2,200 | Active |
| JSON | stdlib | — | Zero-dependency; noisy to hand-author |

For a hand-authored 26-color palette with a nested `palette` table, the candidates were TOML and JSON. TOML reads better for this shape of data (no quotes, no commas, comments allowed) and `BurntSushi/toml` is the most popular *dedicated* config-format library in Go, actively maintained. JSON's zero-dependency property is real but does not outweigh hand-editability here. YAML was ruled out because its canonical Go library is archived.

### Template engine

Mustache is logic-less and language-agnostic — the right property for theme templates, which should stay portable across applications and not pull template authors into Go's `text/template` syntax. Verified Mustache libraries in Go:

| Library | Stars | Status |
| --- | --- | --- |
| `hoisie/mustache` | ~1,100 | Last push Apr 2024; appears unmaintained |
| `cbroglie/mustache` | ~500 | Active (the maintained fork) |

`cbroglie/mustache` is the living fork with an idiomatic error-returning API and spec compliance. Go `text/template` was rejected because it would impose Go syntax on template authors and reduce portability.

## Decision

- **Flavor data format: TOML**, parsed with `BurntSushi/toml`.
- **Template engine: Mustache**, via `cbroglie/mustache`.

## Consequences

- `flavor.toml` keeps its current shape (a `name` field plus a `[palette]` table with the 26 Catppuccin colors). No migration needed since preserving existing flavors is not a concern.
- Templates (`.mustache` files) keep working as-is.
- Two non-stdlib dependencies for format handling, both actively maintained.
