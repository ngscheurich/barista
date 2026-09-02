package template_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/flavor"
	"github.com/ngscheurich/barista/internal/template"
)

// sampleFlavor is a Flavor whose palette values are the bare color names,
// so a rendered template makes it obvious which field landed where.
func sampleFlavor() flavor.Flavor {
	return flavor.Flavor{
		Name: "Mocha",
		Palette: flavor.Palette{
			Rosewater: "rosewater-val",
			Flamingo:  "flamingo-val",
			Pink:      "pink-val",
			Mauve:     "mauve-val",
			Red:       "red-val",
			Maroon:    "maroon-val",
			Peach:     "peach-val",
			Yellow:    "yellow-val",
			Green:     "green-val",
			Teal:      "teal-val",
			Sky:       "sky-val",
			Sapphire:  "sapphire-val",
			Blue:      "blue-val",
			Lavender:  "lavender-val",
			Text:      "text-val",
			Subtext1:  "subtext_1-val",
			Subtext0:  "subtext_0-val",
			Overlay2:  "overlay_2-val",
			Overlay1:  "overlay_1-val",
			Overlay0:  "overlay_0-val",
			Surface2:  "surface_2-val",
			Surface1:  "surface_1-val",
			Surface0:  "surface_0-val",
			Base:      "base-val",
			Mantle:    "mantle-val",
			Crust:     "crust-val",
		},
	}
}

func TestRender(t *testing.T) {
	cases := []struct {
		name string
		tmpl string
		want string
	}{
		{
			name: "name only",
			tmpl: "name = {{name}}",
			want: "name = Mocha",
		},
		{
			name: "nested palette colors use snake_case keys",
			tmpl: "{{palette.rosewater}} {{palette.subtext_1}} {{palette.surface_0}} {{palette.crust}}",
			want: "rosewater-val subtext_1-val surface_0-val crust-val",
		},
		{
			name: "name and palette together",
			tmpl: "# {{name}}\nbackground = {{palette.base}}\ntext = {{palette.text}}",
			want: "# Mocha\nbackground = base-val\ntext = text-val",
		},
		{
			name: "whitespace-padded tags",
			tmpl: "{{  name  }}",
			want: "Mocha",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := template.Render(tc.tmpl, sampleFlavor())

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// Every one of the 26 Catppuccin color names resolves through the palette
// map, so a recipe template never silently renders an empty value for a
// color it references.
func TestRenderExposesAll26Colors(t *testing.T) {
	colors := []string{
		"rosewater", "flamingo", "pink", "mauve", "red", "maroon", "peach",
		"yellow", "green", "teal", "sky", "sapphire", "blue", "lavender",
		"text", "subtext_1", "subtext_0", "overlay_2", "overlay_1",
		"overlay_0", "surface_2", "surface_1", "surface_0", "base",
		"mantle", "crust",
	}
	f := sampleFlavor()

	for _, c := range colors {
		t.Run(c, func(t *testing.T) {
			got, err := template.Render("{{palette."+c+"}}", f)

			require.NoError(t, err)
			assert.Equal(t, c+"-val", got)
		})
	}
}

// An unknown template variable is a hard error rather than a silent empty
// string, so a typo in a recipe template (e.g. {{palette.raosewater}})
// surfaces instead of producing broken content. cbroglie/mustache names the
// offending variable in the error.
func TestRenderUnknownVariableFails(t *testing.T) {
	_, err := template.Render("[{{palette.nonexistent}}]", sampleFlavor())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "render template")
	assert.Contains(t, err.Error(), "nonexistent")
}

// An unknown top-level variable fails the same way a missing palette key
// does; the strictness covers the whole context, not just the palette map.
func TestRenderUnknownTopLevelVariableFails(t *testing.T) {
	_, err := template.Render("{{bogus}}", sampleFlavor())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus")
}

// An unbalanced tag is a parse error from mustache, wrapped with %w so
// callers can inspect the underlying error if they need to.
func TestRenderInvalidTemplateFails(t *testing.T) {
	_, err := template.Render("{{name", sampleFlavor())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "render template")
}
