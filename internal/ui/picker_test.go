package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/ui"
)

func pickerOptions() []ui.Option {
	return []ui.Option{
		{Label: "Catppuccin Latte", Value: "catppuccin-latte"},
		{Label: "Catppuccin Mocha", Value: "catppuccin-mocha"},
	}
}

// In accessible mode the picker prints the title, a numbered list of
// labels, and a plain prompt, then returns the Value of the chosen row.
// The numbered rows are what a screen reader reads, one line each.
func TestPickAccessible(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("1\n")

	got, err := ui.Pick(&out, in, pickerOptions(), true)

	require.NoError(t, err)
	assert.Equal(t, "catppuccin-latte", got)
	assert.Contains(t, out.String(), "Choose a theme")
	assert.Contains(t, out.String(), "1. Catppuccin Latte")
	assert.Contains(t, out.String(), "2. Catppuccin Mocha")
	assert.Contains(t, out.String(), "Enter a number between 1 and 2")
	// The prompt degrades to plain text for a non-terminal writer: no
	// escape sequences reach a piped destination.
	assert.NotContains(t, out.String(), "\x1b[")
}

// Invalid input re-prompts until a valid number arrives; nothing is
// applied from the bad attempts.
func TestPickAccessibleReprompts(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("nope\n9\n2\n")

	got, err := ui.Pick(&out, in, pickerOptions(), true)

	require.NoError(t, err)
	assert.Equal(t, "catppuccin-mocha", got)
}

// EOF on stdin selects the first option: huh's accessible prompt
// treats end-of-input as accepting the default. Pinned here because
// it is why accessible mode is opt-in and never enabled implicitly —
// an implicit run over a dead pipe would apply a theme unasked.
func TestPickAccessibleEOFSelectsDefault(t *testing.T) {
	var out bytes.Buffer

	got, err := ui.Pick(&out, strings.NewReader(""), pickerOptions(), true)

	require.NoError(t, err)
	assert.Equal(t, "catppuccin-latte", got)
}
