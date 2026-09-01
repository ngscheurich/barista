// Package flavor defines the Flavor domain type and its decoding from
// flavor.toml content.
//
// A Flavor is a named Catppuccin palette variant: a Name plus a Palette of
// the 26 colors. flavor.toml is the on-disk serialization of a Flavor; the
// Theme is the container that pairs a Flavor with the per-application
// templates that render it. The on-disk TOML shape is decoded into an
// unexported mirror struct so adding a TOML field later does not force
// the exported domain type to change shape in lockstep.
package flavor

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Palette is the 26 Catppuccin color values inside a Flavor.
type Palette struct {
	Rosewater string
	Flamingo  string
	Pink      string
	Mauve     string
	Red       string
	Maroon    string
	Peach     string
	Yellow    string
	Green     string
	Teal      string
	Sky       string
	Sapphire  string
	Blue      string
	Lavender  string
	Text      string
	Subtext1  string
	Subtext0  string
	Overlay2  string
	Overlay1  string
	Overlay0  string
	Surface2  string
	Surface1  string
	Surface0  string
	Base      string
	Mantle    string
	Crust     string
}

// Flavor is a named Catppuccin palette variant: a Name plus a Palette of
// the 26 colors. This is the same unit Catppuccin calls a flavor; Barista
// renders it through per-application templates grouped into a Theme.
type Flavor struct {
	Name    string
	Palette Palette
}

// AsMap exposes the 26 Catppuccin colors keyed by the snake_case names that
// templates reference, so a Mustache context can offer {{palette.rosewater}}
// and the rest. This is the single enumeration of the color list for
// rendering; the only other site is the paletteFile decode tags in Parse.
func (p Palette) AsMap() map[string]string {
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

// Parse decodes flavor.toml content into a Flavor, enforcing that the
// name and all 26 colors are present. BurntSushi/toml leaves missing keys
// as zero values rather than erroring, so a flavor.toml missing a color
// would otherwise decode to an empty string and render silently; the spec
// requires a missing color to fail the load, so required-ness is enforced
// here.
func Parse(raw []byte) (Flavor, error) {
	var f flavorFile
	if err := toml.Unmarshal(raw, &f); err != nil {
		return Flavor{}, fmt.Errorf("parse flavor: %w", err)
	}
	if missing := missingFields(f); missing != "" {
		return Flavor{}, fmt.Errorf("parse flavor: missing required field %s", missing)
	}
	return Flavor{
		Name: f.Name,
		// Tags are ignored for struct conversion, so the untagged Palette
		// is built from the tagged paletteFile in one move.
		Palette: Palette(f.Palette),
	}, nil
}

// flavorFile mirrors the on-disk flavor.toml shape. TOML keys are
// snake_case; the tags bridge them to MixedCaps Go identifiers.
type flavorFile struct {
	Name    string      `toml:"name"`
	Palette paletteFile `toml:"palette"`
}

type paletteFile struct {
	Rosewater string `toml:"rosewater"`
	Flamingo  string `toml:"flamingo"`
	Pink      string `toml:"pink"`
	Mauve     string `toml:"mauve"`
	Red       string `toml:"red"`
	Maroon    string `toml:"maroon"`
	Peach     string `toml:"peach"`
	Yellow    string `toml:"yellow"`
	Green     string `toml:"green"`
	Teal      string `toml:"teal"`
	Sky       string `toml:"sky"`
	Sapphire  string `toml:"sapphire"`
	Blue      string `toml:"blue"`
	Lavender  string `toml:"lavender"`
	Text      string `toml:"text"`
	Subtext1  string `toml:"subtext_1"`
	Subtext0  string `toml:"subtext_0"`
	Overlay2  string `toml:"overlay_2"`
	Overlay1  string `toml:"overlay_1"`
	Overlay0  string `toml:"overlay_0"`
	Surface2  string `toml:"surface_2"`
	Surface1  string `toml:"surface_1"`
	Surface0  string `toml:"surface_0"`
	Base      string `toml:"base"`
	Mantle    string `toml:"mantle"`
	Crust     string `toml:"crust"`
}

// missingFields returns the name of the first required field that is
// absent from the decoded file, or the empty string if all are present.
func missingFields(f flavorFile) string {
	if f.Name == "" {
		return "name"
	}
	type kv struct{ key, val string }
	for _, c := range []kv{
		{"rosewater", f.Palette.Rosewater},
		{"flamingo", f.Palette.Flamingo},
		{"pink", f.Palette.Pink},
		{"mauve", f.Palette.Mauve},
		{"red", f.Palette.Red},
		{"maroon", f.Palette.Maroon},
		{"peach", f.Palette.Peach},
		{"yellow", f.Palette.Yellow},
		{"green", f.Palette.Green},
		{"teal", f.Palette.Teal},
		{"sky", f.Palette.Sky},
		{"sapphire", f.Palette.Sapphire},
		{"blue", f.Palette.Blue},
		{"lavender", f.Palette.Lavender},
		{"text", f.Palette.Text},
		{"subtext_1", f.Palette.Subtext1},
		{"subtext_0", f.Palette.Subtext0},
		{"overlay_2", f.Palette.Overlay2},
		{"overlay_1", f.Palette.Overlay1},
		{"overlay_0", f.Palette.Overlay0},
		{"surface_2", f.Palette.Surface2},
		{"surface_1", f.Palette.Surface1},
		{"surface_0", f.Palette.Surface0},
		{"base", f.Palette.Base},
		{"mantle", f.Palette.Mantle},
		{"crust", f.Palette.Crust},
	} {
		if c.val == "" {
			return "palette." + c.key
		}
	}
	return ""
}
