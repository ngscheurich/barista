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
// tags ({{name}}, {{palette.rosewater}}) resolve case-sensitively.
func Render(tmpl string, f flavor.Flavor) (string, error) {
	rendered, err := mustache.Render(tmpl, map[string]interface{}{
		"name":    f.Name,
		"palette": paletteAsMap(f.Palette),
	})
	if err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}
	return rendered, nil
}

// paletteAsMap exposes the 26 Catppuccin colors keyed by the snake_case
// names templates reference. This is the one place the color list is
// enumerated for rendering; flavor.Load is the only other site, where the
// palette is decoded from TOML.
func paletteAsMap(p flavor.Palette) map[string]string {
	return map[string]string{
		"rosewater": p.Rosewater,
		"flamingo":  p.Flamingo,
		"pink":      p.Pink,
		"mauve":     p.Mauve,
		"red":       p.Red,
		"maroon":    p.Maroon,
		"peach":     p.Peach,
		"yellow":    p.Yellow,
		"green":     p.Green,
		"teal":      p.Teal,
		"sky":       p.Sky,
		"sapphire":  p.Sapphire,
		"blue":      p.Blue,
		"lavender":  p.Lavender,
		"text":      p.Text,
		"subtext_1": p.Subtext1,
		"subtext_0": p.Subtext0,
		"overlay_2": p.Overlay2,
		"overlay_1": p.Overlay1,
		"overlay_0": p.Overlay0,
		"surface_2": p.Surface2,
		"surface_1": p.Surface1,
		"surface_0": p.Surface0,
		"base":      p.Base,
		"mantle":    p.Mantle,
		"crust":     p.Crust,
	}
}
