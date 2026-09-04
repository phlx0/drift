package scene

import (
	"github.com/gdamore/tcell/v2"
)

// ansiRGB holds the conventional true-color value of each of the 16 ANSI
// colors, in palette order (black, maroon, green, olive, navy, purple, teal,
// silver, then their bright counterparts).
//
// drift uses them for two things: as the nominal colors of the "ansi" theme,
// and as reference points when snapping an arbitrary color to the nearest
// palette index. Nothing renders these values directly — the engine emits the
// index, and the terminal substitutes whatever the user configured for it —
// so they only have to be representative, not exact.
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

// ANSIIndex returns the index (0–15) of the ANSI palette color closest to c.
//
// Distance is squared RGB weighted 2:4:3, the usual cheap stand-in for
// perceived difference — green dominates brightness, blue contributes least.
// A proper Lab distance is more accurate but far too slow to run on every
// cell of every frame.
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

// ANSIColor snaps c to the nearest ANSI palette entry. The returned color
// renders with the terminal's own palette rather than a fixed RGB value.
func ANSIColor(c RGBColor) tcell.Color {
	return tcell.PaletteColor(ANSIIndex(c))
}

// ansiTheme builds the "ansi" theme out of the standard palette: the bright
// colors as accents, their normal counterparts as the dim variants, and white
// as the highlight. Because the theme sets ANSI, the engine snaps every color
// the scenes produce — interpolated trails included — back to a palette index,
// so what a user actually sees is their own terminal colorscheme.
func ansiTheme() Theme {
	return Theme{
		Name: "ansi",
		ANSI: true,
		Palette: []RGBColor{
			ansiRGB[12], // bright blue
			ansiRGB[13], // bright magenta
			ansiRGB[14], // bright cyan
			ansiRGB[10], // bright green
		},
		Dim: []RGBColor{
			ansiRGB[4], // blue
			ansiRGB[5], // magenta
			ansiRGB[6], // cyan
			ansiRGB[2], // green
		},
		Bright: ansiRGB[15], // white
	}
}
