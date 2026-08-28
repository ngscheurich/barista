package cli_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyPrintsThemeAndExitsZero(t *testing.T) {
	root, out, _ := newRoot([]string{"apply", "catppuccin-mocha"})

	assert.NoError(t, root.Execute())
	assert.Equal(t, "catppuccin-mocha\n", out.String())
}

func TestApplyRejectsExtraArgs(t *testing.T) {
	root, _, _ := newRoot([]string{"apply", "one", "two"})

	assert.Error(t, root.Execute())
}

func TestApplyRequiresArg(t *testing.T) {
	root, _, _ := newRoot([]string{"apply"})

	assert.Error(t, root.Execute())
}

func TestRootHelpHasShortDescription(t *testing.T) {
	root, out, _ := newRoot([]string{"--help"})

	assert.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Serves up a new <theme> for your terminal apps.")
}

func TestRootNoArgsPrintsHelp(t *testing.T) {
	root, out, _ := newRoot(nil)

	assert.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Usage:",
		"root with no args should print help with a Usage block; got:\n%s", out.String())
}
