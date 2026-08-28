// Package recipe defines the Recipe interface each terminal application
// recipe implements.
//
// A Recipe is the procedure for one application: render its template from
// the Flavor, write the theme file to the right location, and reload the
// application. The single-method interface follows Go's small-interface
// idiom; the orchestration in internal/cli loops over recipes, collecting
// errors rather than stopping at the first.
package recipe

import "github.com/ngscheurich/barista/internal/flavor"

// Recipe is the procedure for theming and reloading one terminal
// application. Run renders the recipe's template against f, writes the
// theme file, and reloads the application; an error indicates any step
// failed. Orchestrators run recipes independently and aggregate errors.
type Recipe interface {
	Run(f flavor.Flavor) error
}
