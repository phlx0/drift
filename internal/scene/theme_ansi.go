package scene

import (
	"github.com/gdamore/tcell/v2"
)

// ansiRGB holds conventional values for the 16 ANSI colors. They're never
// rendered directly (the terminal substitutes its own palette for the index),
// so they only need to be representative reference points for snapping.
var ansiRGB = [16]RGBColor{
	{0, 0, 0},       // 0  black
	{128, 0, 0},     // 1  maroon
	{0, 128, 0},     // 2  green
	{128, 128, 0},   // 3  olive
	{0, 0, 128},     // 4  navy
	{128, 0, 128},   // 5  purple
	{0, 128, 128},   // 6  teal
	{192, 192, 192}, // 7  silver
	{128, 128, 128}, // 8  gray
	{255, 0, 0},     // 9  red
	{0, 255, 0},     // 10 lime
	{255, 255, 0},   // 11 yellow
	{0, 0, 255},     // 12 blue
	{255, 0, 255},   // 13 fuchsia
	{0, 255, 255},   // 14 aqua
	{255, 255, 255}, // 15 white
}

// The 2:4:3 RGB weighting is a cheap approximation of perceived difference;
// Lab would be more accurate but too slow to run per cell per frame.
func ANSIIndex(c RGBColor) int {
	best, bestDist := 0, -1
	for i, ref := range ansiRGB {
		dr := int(c.R) - int(ref.R)
		dg := int(c.G) - int(ref.G)
		db := int(c.B) - int(ref.B)
		d := 2*dr*dr + 4*dg*dg + 3*db*db
		if bestDist < 0 || d < bestDist {
			best, bestDist = i, d
		}
	}
	return best
}

func ANSIColor(c RGBColor) tcell.Color {
	return tcell.PaletteColor(ANSIIndex(c))
}

func ansiTheme() Theme {
	return Theme{
		Name: "ansi",
		ANSI: true,
		Palette: []RGBColor{
			ansiRGB[12],
			ansiRGB[13],
			ansiRGB[14],
			ansiRGB[10],
		},
		Dim: []RGBColor{
			ansiRGB[4],
			ansiRGB[5],
			ansiRGB[6],
			ansiRGB[2],
		},
		Bright: ansiRGB[15],
	}
}
