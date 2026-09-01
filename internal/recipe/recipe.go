// Package recipe defines the Recipe interface each terminal application
// recipe implements.
//
// A Recipe is the procedure for one application: render its template from
// the Theme's Flavor, write the output to the right location, and reload
// the application. The single-method interface follows Go's small-interface
// idiom; the orchestration in internal/cli loops over recipes, collecting
// errors rather than stopping at the first.
package recipe

import (
	"errors"

	"github.com/ngscheurich/barista/internal/theme"
)

// ErrNotApplicable indicates a recipe's template file is absent from the
// theme directory, i.e. the theme does not carry an output for that
// application. Orchestrators treat it as a skip rather than a failure:
// the application is listed but not served, and the run does not abort.
var ErrNotApplicable = errors.New("recipe not applicable: theme has no template for this application")

// Recipe is the procedure for theming and reloading one terminal
// application. Run renders the recipe's template against t.Flavor, writes
// the output, and reloads the application; an error indicates any step
// failed. Orchestrators run recipes independently and aggregate errors.
type Recipe interface {
	Run(t theme.Theme) error
}
