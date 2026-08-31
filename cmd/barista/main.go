// Command barista serves up a new flavor for your terminal apps.
//
// main is intentionally tiny: it builds the cobra command tree from internal/cli
// and executes it against the program context. All command logic lives in
// internal/cli so it stays testable without going through the binary.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ngscheurich/barista/internal/cli"
)

func main() {
	if err := cli.NewRoot().ExecuteContext(context.Background()); err != nil {
		if errors.Is(err, cli.ErrAborted) {
			// The user cancelled the picker: nothing was applied,
			// and nothing needs saying.
			os.Exit(130)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
