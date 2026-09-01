package flavor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngscheurich/barista/internal/flavor"
)

// validPaletteTOML is a complete flavor.toml with all 26 colors. The hex
// values are arbitrary but distinct so a transposition in decoding surfaces.
const validPaletteTOML = `name = "Catppuccin Mocha"

[palette]
rosewater = "#f5e0dc"
flamingo = "#f2cdcd"
pink = "#f5c2e7"
mauve = "#cba6f7"
red = "#f38ba8"
maroon = "#eba0ac"
peach = "#fab387"
yellow = "#f9e2af"
green = "#a6e3a1"
teal = "#94e2d5"
sky = "#89dceb"
sapphire = "#74c7ec"
blue = "#89b4fa"
lavender = "#b4befe"
text = "#cdd6f4"
subtext_1 = "#bac2de"
subtext_0 = "#a6adc8"
overlay_2 = "#9399b2"
overlay_1 = "#7f849c"
overlay_0 = "#6c7086"
surface_2 = "#585b70"
surface_1 = "#45475a"
surface_0 = "#313244"
base = "#1e1e2e"
mantle = "#181825"
crust = "#11111b"
`

func wantPalette() flavor.Palette {
	return flavor.Palette{
		Rosewater: "#f5e0dc",
		Flamingo:  "#f2cdcd",
		Pink:      "#f5c2e7",
		Mauve:     "#cba6f7",
		Red:       "#f38ba8",
		Maroon:    "#eba0ac",
		Peach:     "#fab387",
		Yellow:    "#f9e2af",
		Green:     "#a6e3a1",
		Teal:      "#94e2d5",
		Sky:       "#89dceb",
		Sapphire:  "#74c7ec",
		Blue:      "#89b4fa",
		Lavender:  "#b4befe",
		Text:      "#cdd6f4",
		Subtext1:  "#bac2de",
		Subtext0:  "#a6adc8",
		Overlay2:  "#9399b2",
		Overlay1:  "#7f849c",
		Overlay0:  "#6c7086",
		Surface2:  "#585b70",
		Surface1:  "#45475a",
		Surface0:  "#313244",
		Base:      "#1e1e2e",
		Mantle:    "#181825",
		Crust:     "#11111b",
	}
}

// Parse decodes flavor.toml content into a Flavor with Name and Palette.
func TestParse(t *testing.T) {
	cases := []struct {
		name    string
		toml    string
		want    flavor.Flavor
		wantErr bool
	}{
		{
			name: "full 26-color palette",
			toml: validPaletteTOML,
			want: flavor.Flavor{
				Name:    "Catppuccin Mocha",
				Palette: wantPalette(),
			},
		},
		{
			name:    "missing color fails",
			toml:    validPaletteTOML[:len(validPaletteTOML)-len("crust = \"#11111b\"\n")],
			wantErr: true,
		},
		{
			name: "wrong-typed value fails",
			toml: `name = "Bad"
[palette]
rosewater = 123
`,
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := flavor.Parse([]byte(tc.toml))

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// A missing top-level name fails parsing with an error naming the field.
func TestParseMissingNameFails(t *testing.T) {
	toml := `[palette]
rosewater = "#f5e0dc"
flamingo = "#f2cdcd"
pink = "#f5c2e7"
mauve = "#cba6f7"
red = "#f38ba8"
maroon = "#eba0ac"
peach = "#fab387"
yellow = "#f9e2af"
green = "#a6e3a1"
teal = "#94e2d5"
sky = "#89dceb"
sapphire = "#74c7ec"
blue = "#89b4fa"
lavender = "#b4befe"
text = "#cdd6f4"
subtext_1 = "#bac2de"
subtext_0 = "#a6adc8"
overlay_2 = "#9399b2"
overlay_1 = "#7f849c"
overlay_0 = "#6c7086"
surface_2 = "#585b70"
surface_1 = "#45475a"
surface_0 = "#313244"
base = "#1e1e2e"
mantle = "#181825"
crust = "#11111b"
`

	_, err := flavor.Parse([]byte(toml))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field name")
}
