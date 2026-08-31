# ADR-0003: The picker is a huh form

Date: 2026-08-31
Status: Accepted

## Context

`barista apply` invoked with no flavor argument will open an interactive picker of available flavors. "Stay in the Charm ecosystem" admits three idiomatic candidates:

| Candidate | External dependency | Accessibility story |
| --- | --- | --- |
| Shell out to `gum choose` | Yes — the gum binary at runtime; contradicts the single-static-binary property | gum has none of its own |
| Bubble Tea + `bubbles/list` directly | No | None: a full-screen TUI is opaque to a screen reader |
| huh | No | **First-class accessible mode** |

Accessibility is a core principle in this repo ([`docs/style/accessibility.md`](../style/accessibility.md)): no piece of meaning may ride on a single perceptual channel, and a full-screen Bubble Tea application is explicitly called out as unreadable by screen readers. Verified against primary sources (2026-08-31): huh's accessible mode (`form.WithAccessible(true)`) drops the TUI in favor of plain numbered prompts, read linearly, with no in-place redraws; the bubbletea and lipgloss READMEs contain no accessibility support at all. The picker is the repo's first interactive surface, so its screen-reader story is decided here, not retrofitted.

Three related decisions settled in the same conversation:

1. **Accessible mode stays opt-in.** In huh's source, an accessible prompt that hits EOF on stdin silently returns the default option — so barista never enables it implicitly. It is wired to the `BARISTA_ACCESSIBLE` environment variable (following huh's own recommendation), and when `apply` has no argument and stdin is not a terminal, it fails with a plain list of available flavors instead of prompting.
2. **Theme: Base16.** Of huh's predefined themes, Base16 is the only one styled solely with named 16-color ANSI indices — which is exactly what the accessibility guide prescribes — rather than hardcoded hexes. The picker therefore renders in the user's live terminal palette (their currently applied flavor).
3. **Version: huh v2 (`charm.land/huh/v2`).** The v1 stack is frozen by release cadence (lipgloss v1 last moved Mar 2025, bubbletea v1 Sep 2025; the `charm.land` v2 line is actively released), so huh v1 would land new code on a dead end. huh v2 pulls the entire Charm v2 line, so the existing lipgloss and log dependencies migrate to `charm.land` v2 paths first, keeping one coherent stack with no dual majors.

## Decision

- Build the picker on **huh** (`charm.land/huh/v2`) — a single `Select` field, not a raw Bubble Tea program and not an external gum invocation.
- Accessible mode is enabled only by the `BARISTA_ACCESSIBLE` environment variable, never implicitly.
- The picker uses huh's **Base16** theme.
- The Charm stack migrates to the `charm.land` v2 line **before** the picker lands (migration ticket blocks picker ticket).

## Consequences

- Barista keeps its single-static-binary property while gaining its first interactive surface.
- Screen-reader users get a linear, numbered prompt; scripts and non-terminal stdin get a complete, plain-text list of flavors in the error message.
- The migration is a separate, blocking ticket: import-path swaps for lipgloss and log, verified against the v2 API deltas, with no functional change.
- One lipgloss major version in the tree — no v1/v2 coexistence debt.
- If a future `barista tui` command arrives (per the accessibility guide), its prompts start from this same huh foundation.
