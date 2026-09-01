// Package template renders Mustache templates against a Flavor.
//
// Render is the single entry point: it builds a context exposing the
// flavor's name and its 26-color palette (keyed by the Catppuccin color
// names, so templates use {{name}} and {{palette.rosewater}}) and renders
// via cbroglie/mustache's error-returning Render.
package template

import (
	"fmt"

	"github.com/cbroglie/mustache"

	"github.com/ngscheurich/barista/internal/flavor"
)

// Render renders tmpl as a Mustache template against f, returning the
// rendered string. Template parse or runtime errors are wrapped with a
// render template: prefix so callers can identify the failing layer.
// The context is a map keyed by lowercase name and palette so template
// tags ({{name}}, {{palette.rosewater}}) resolve case-sensitively;
// cbroglie/mustache looks up struct fields via reflect.FieldByName with
// no tag support, so a struct cannot expose the lowercase keys.
//
// An unknown template variable is a hard error, not a silent empty
// string: cbroglie/mustache defaults to the Mustache-spec behaviour of
// emitting nothing for a miss, which would let a typo in a recipe
// template ({{palette.raosewater}}) write broken output with no
// signal. AllowMissingVariables is a package-level switch, so Render
// flips it off for the whole process on first use. Barista is the binary's
// only mustache consumer and recipes render sequentially, so the global
// is safe to set here.
func Render(tmpl string, f flavor.Flavor) (string, error) {
	mustache.AllowMissingVariables = false

	rendered, err := mustache.Render(tmpl, map[string]interface{}{
		"name":    f.Name,
		"palette": f.Palette.AsMap(),
	})
	if err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}
	return rendered, nil
}
