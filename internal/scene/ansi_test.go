package scene

import "testing"

func TestANSIIndexExactMatches(t *testing.T) {
	for want, c := range ansiRGB {
		if got := ANSIIndex(c); got != want {
			t.Errorf("ANSIIndex(%v) = %d, want %d", c, got, want)
		}
	}
}

func TestANSIIndexNearest(t *testing.T) {
	cases := []struct {
		name string
		in   RGBColor
		want int
	}{
		{"near black", RGBColor{10, 10, 10}, 0},
		{"dark red", RGBColor{90, 0, 0}, 1},
		{"saturated red", RGBColor{255, 10, 10}, 9},
		{"light grey", RGBColor{200, 200, 200}, 7},
		{"mid grey", RGBColor{120, 130, 125}, 8},
		{"strong blue", RGBColor{0, 0, 200}, 12},
	}
	for _, tc := range cases {
		if got := ANSIIndex(tc.in); got != tc.want {
			t.Errorf("%s: ANSIIndex(%v) = %d, want %d", tc.name, tc.in, got, tc.want)
		}
	}
}

// The ansi theme's own colors must survive snapping untouched, otherwise a
// scene drawing straight from the palette would land on a different terminal
// color than the one the theme names.
func TestANSIThemeColorsRoundTrip(t *testing.T) {
	th, ok := Themes["ansi"]
	if !ok {
		t.Fatal("theme \"ansi\" is not registered")
	}
	if !th.ANSI {
		t.Error("theme \"ansi\" must set ANSI so the engine snaps its output")
	}

	all := append(append([]RGBColor{th.Bright}, th.Palette...), th.Dim...)
	for _, c := range all {
		if got := ansiRGB[ANSIIndex(c)]; got != c {
			t.Errorf("theme color %v snaps to %v, want itself", c, got)
		}
	}
}
