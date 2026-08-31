// Package ui holds Barista's interactive surfaces. Today that is the
// picker: a huh select presented when `barista apply` is invoked with no
// flavor argument.
package ui

import (
	"fmt"
	"io"
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// Option is one row of the picker: Label is what the user sees and
// chooses; Value is what Pick returns.
type Option struct {
	Label string
	Value string
}

// pickTitle is the picker's one line of framing text. It pairs with the
// numbered labels — a screen reader gets the word "flavor" in the title,
// not just a bare list.
const pickTitle = "Choose a flavor"

// pickerTheme is Base16 with the form chrome stripped: the bordered,
// padded box becomes plain inline text — title, rows, help footer —
// rendered where the cursor sits and erased on completion, never a
// screen takeover. The colors stay the named 16-color ANSI indices, so
// the picker wears the user's live terminal palette.
func pickerTheme(isDark bool) *huh.Styles {
	t := huh.ThemeBase16(isDark)
	t.Focused.Base = lipgloss.NewStyle()
	t.Blurred.Base = lipgloss.NewStyle()
	return t
}

// Pick prompts the user to choose one option, returning the chosen
// Value. opts must be non-empty; callers guard the empty case before
// presenting a picker. accessible runs the prompt in huh's accessible
// mode: plain numbered rows and a typed number instead of a TUI, which
// is what screen readers need — and what works over a pipe. When the
// user cancels the form, the returned error wraps huh.ErrUserAborted.
func Pick(out io.Writer, in io.Reader, opts []Option, accessible bool) (string, error) {
	choice := ""
	huhOptions := make([]huh.Option[string], len(opts))
	for i, o := range opts {
		huhOptions[i] = huh.NewOption(o.Label, o.Value)
	}

	// The accessible prompt is plain output, so gate it through the
	// same colorprofile seam the rest of the output uses: a piped
	// destination gets plain text, and NO_COLOR is honored. The TUI
	// path renders through bubbletea, which manages its own profile.
	if accessible {
		out = colorprofile.NewWriter(out, os.Environ())
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(pickTitle).
				Options(huhOptions...).
				Value(&choice),
		),
	).
		WithTheme(huh.ThemeFunc(pickerTheme)).
		WithAccessible(accessible).
		WithInput(in).
		WithOutput(out)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("pick flavor: %w", err)
	}
	return choice, nil
}
