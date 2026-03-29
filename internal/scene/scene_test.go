package scene

import (
	"testing"
)

func TestThemesAllHavePalette(t *testing.T) {
	for name, theme := range Themes {
		if len(theme.Palette) == 0 {
			t.Errorf("theme %q has empty Palette", name)
		}
		if len(theme.Dim) == 0 {
			t.Errorf("theme %q has empty Dim", name)
		}
		if len(theme.Palette) != len(theme.Dim) {
			t.Errorf("theme %q: Palette len %d != Dim len %d", name, len(theme.Palette), len(theme.Dim))
		}
	}
}

func TestLerp(t *testing.T) {
	black := RGBColor{0, 0, 0}
	white := RGBColor{255, 255, 255}

	mid := Lerp(black, white, 0.5)
	if mid.R != 127 && mid.R != 128 {
		t.Errorf("Lerp(black, white, 0.5).R = %d, want ~127", mid.R)
	}

	at0 := Lerp(black, white, 0)
	if at0 != black {
		t.Errorf("Lerp at t=0 should equal a, got %v", at0)
	}

	at1 := Lerp(black, white, 1)
	if at1 != white {
		t.Errorf("Lerp at t=1 should equal b, got %v", at1)
	}

	// Clamp test: t > 1 should not exceed b
	over := Lerp(black, white, 2)
	if over != white {
		t.Errorf("Lerp with t>1 should clamp to b, got %v", over)
	}
}
