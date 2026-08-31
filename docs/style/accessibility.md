# Accessibility in Barista

How Barista's output stays usable for everyone who reads a terminal — independent of any one language. For Go code style see [`go.md`](go.md); for the prose style see [`prose.md`](prose.md).

Accessibility is a core principle, not a finishing pass. The one rule the rest of this guide follows: **no piece of meaning rides on a single perceptual channel** — color, dimness, position, or motion — that some reader can't perceive. Two audiences make that concrete, and they fail differently:

- **Low vision or color vision deficiency** — sighted, but can't rely on color (red-green deficiency alone is roughly 8% of men of Northern European descent) or on fine contrast.
- **Screen reader and braille users** — read the terminal's rendered text linearly and perceive neither color nor layout.

## Color and de-emphasis are redundant, never the message

Barista's user-facing output is minimal today — the `☕︎ Served up <name>` line on success and error messages on failure — but the principle holds from the first line, and matters more once a TUI arrives.

- Pair every styled state with a word or glyph that carries the meaning on its own; color is decoration on top. If success is ever colored green and errors red, the words `served` and `error:` tell them apart on their own; color only speeds up the terminal case.
- For a screen reader, color isn't merely hard to tell apart — it's **absent**: the terminal consumes the escape codes before the reader reaches the text. Color-only encoding fails the color-vision audience and the screen reader audience at once.
- Don't let red-versus-green be the only distinction; differ by word, symbol, or position too.
- When styling arrives (Lip Gloss, in the future TUI), stick to the named 16-color ANSI palette (`lipgloss.Color("1")`–`"15"`). The user's theme defines what those look like, so they adapt to light and dark backgrounds for free. Reach for `lipgloss.AdaptiveColor` only when you hardcode specific shades, and never assume the background is light or dark.
- `Muted`/faint **reduces** contrast against a baseline you don't control, and terminals honor it inconsistently — some render it dimmer, some ignore it outright. Use it only on text that is already secondary; never to signal that something is secondary.

## Degrade by destination, and keep an explicit plain lever

Output should degrade by where it's going: styled at a terminal, plain when piped, and — if a machine-readable form is ever added — a `--json` flag for tools. Route all human-facing output through one place (a future `internal/ui` package) so that gate stays in one seam, rather than scattering `fmt.Println` and Lip Gloss calls across the codebase.

- Color detection is **inherited** from `termenv` beneath Lip Gloss: `NO_COLOR`, `CLICOLOR`/`CLICOLOR_FORCE`, and `TERM=dumb` are honored. Don't reimplement them, and don't honor `FORCE_COLOR` — that's a Node convention `termenv` doesn't read; `CLICOLOR_FORCE` is the knob. (Carried over from the Spiritualism guide's established finding; not independently re-verified for this doc.)
- That detection governs **color only**. Border glyphs are content, gated separately by whether stdout is a terminal, so `NO_COLOR` at an interactive terminal still draws borders. A screen reader user there gets the structural noise regardless. Since no portable signal says "a screen reader is attached," an explicit `--plain`/`--no-color` flag is the only lever that drops borders and in-place redraws at a TTY. Treat it as an accessibility lever, not just a piping convenience.
- The aggregated-error report (the spec's continue-on-error behavior) must read as complete and ordered in plain text: `error: ghostty: pgrep failed`, `error: neovim: no sockets found`, one per line, color optional. The severity word (`error:`) and the app name carry the meaning; color is decoration.

## Screen reader and braille friendliness

- The interactive picker (`barista apply` with no flavor argument) is a huh form, and huh's **accessible mode** is the screen-reader surface for it: the TUI is replaced by plain numbered prompts, read linearly, no in-place redraws. It is enabled with the `BARISTA_ACCESSIBLE` environment variable — set once in the user's shell, not a per-invocation flag — following huh's own recommendation to wire accessibility to an env var.
- Accessible mode happens to work over pipes and captured stdio, but with no input on stdin it silently falls back to a default selection. So it stays strictly opt-in: barista never enables it implicitly. When `apply` has no flavor argument and stdin is not a terminal, it fails with a plain, complete list of available flavors — that error message is the scripted and screen-reader surface, no interaction required.
- Output is read top to bottom, left to right. Borders, box-drawing, ASCII art, and space-aligned columns are announced as noise or carry no structure. Keep the meaning in words.
- The plain and any future `--json` paths are primary surfaces for assistive technology, not just for scripts — keep them complete. If a `--json` form is added, it must carry the same information as the human-readable form, including every error.
- Don't repaint in place outside a terminal, or when plain output is requested. Spinners that use `\r` and cursor moves flood a screen reader with repeated lines; emit discrete state-change lines instead. (Barista is a one-shot CLI today, so this is mostly a future-TUI concern — but the `☕︎` success glyph is the kind of in-place flourish to watch: keep it on a single line, not a redraw.)

## Glyphs, motion, and timing

- The `☕︎` success glyph pairs with the word `Served up` — the meaning survives without the glyph. Keep that pairing for any future glyph: pair it with text, provide an ASCII fallback, and don't assume the font covers it or that a screen reader pronounces it sensibly.
- Make animation possible to turn off (it folds into `--plain`): motion is distraction and vestibular load, and redraws flood screen readers. Outright strobing is unlikely in a text UI, but it's cheap to never emit.
- Don't assume fast sighted reading. Avoid timeouts on interactive prompts; if one is unavoidable, make it configurable. (Relevant when the TUI lands; the current one-shot CLI has no prompts.)

## The future TUI

A `barista tui` command is a future possibility, not current scope. When it arrives:

- Everything is reachable by keyboard — no mouse-only path. Keybindings are discoverable (a help overlay) and consistent.
- The focus indicator is more than color; the focus order is logical; quit and escape are reliable; on resize the layout reflows rather than truncating essential information.
- A full-screen Bubble Tea application is effectively opaque to a screen reader, so the CLI is the accessible surface. Every TUI operation must have a CLI equivalent and a `--json` form. Keep that parity from day one of the TUI, not as a retrofit.

## Where this is enforced

The contract belongs at the output seam. Today that seam is small — the `apply` command's success and error printing in `internal/cli`. As styled output grows, consolidate it behind a single `internal/ui` package so the rule holds by construction: semantic styles add color to text that already carries the meaning, and a test asserts the severity word survives a non-terminal writer. New styled output goes through that package, not raw Lip Gloss, so color-on-color-only encoding can't slip in.
