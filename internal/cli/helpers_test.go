package cli_test

import (
	"bytes"

	"github.com/spf13/cobra"

	"github.com/ngscheurich/barista/internal/cli"
)

func newRoot(args []string) (root *cobra.Command, out, errOut *bytes.Buffer) {
	out, errOut = &bytes.Buffer{}, &bytes.Buffer{}
	root = cli.NewRoot()
	root.SetOut(out)
	root.SetErr(errOut)
	root.SetArgs(args)
	return root, out, errOut
}
