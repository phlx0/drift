// Package scene defines the Scene interface and shared types used by all
// drift animations.
package scene

import (
	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
)

// Scene is the interface every drift animation must implement.
// The engine calls methods in order: Init → (Update → Draw)* → Resize → …
type Scene interface {
	Name() string

	// Init initialises the scene for the given terminal dimensions and theme.
	// Called once when the scene becomes active and again after every Resize.
	Init(w, h int, t Theme)

	// Update advances the simulation by dt seconds.
	Update(dt float64)

	// Draw renders the current state onto the screen.
	// Must NOT call screen.Show() — the engine owns flushing.
	Draw(screen tcell.Screen)

	Resize(w, h int)
}

// All returns fresh instances of every scene, configured with the optional
// SceneConfig. If no config is provided, compiled-in defaults are used.
func All(cfgs ...config.SceneConfig) []Scene {
	cfg := config.Default().Scene
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	return []Scene{
		NewConstellation(cfg.Constellation),
		NewRain(cfg.Rain),
		NewParticles(cfg.Particles),
		NewWaveform(cfg.Waveform),
		NewPipes(cfg.Pipes),
		NewMaze(cfg.Maze),
		NewLife(cfg.Life),
	}
}

func ByName(name string, cfgs ...config.SceneConfig) Scene {
	for _, s := range All(cfgs...) {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

func Names() []string {
	all := All()
	names := make([]string, len(all))
	for i, s := range all {
		names[i] = s.Name()
	}
	return names
}

// RGBColor is a 24-bit true-color value.
type RGBColor struct{ R, G, B uint8 }

// Tcell converts to tcell.Color, degrading automatically on terminals that
// don't support true color.
func (c RGBColor) Tcell() tcell.Color {
	return tcell.NewRGBColor(int32(c.R), int32(c.G), int32(c.B))
}

// Style returns a tcell.Style with this color as foreground and the
// terminal's default background (never hardcodes a background color).
func (c RGBColor) Style() tcell.Style {
	return tcell.StyleDefault.Foreground(c.Tcell())
}

// Lerp linearly interpolates between two RGBColors, clamping t to [0, 1].
func Lerp(a, b RGBColor, t float64) RGBColor {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return RGBColor{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
	}
}

// Theme holds the color palette for a scene.
type Theme struct {
	Name string
	// Palette is the main accent colors. Scenes index with Palette[i % len(Palette)].
	Palette []RGBColor
	// Dim mirrors Palette with darker variants for trails and depth.
	// Always the same length as Palette.
	Dim []RGBColor
	// Bright is a near-white highlight for peaks and heads.
	Bright RGBColor
}

var Themes = map[string]Theme{
	"cosmic": {
		Name: "cosmic",
		Palette: []RGBColor{
			{100, 140, 230},
			{160, 100, 220},
			{80, 200, 220},
			{180, 140, 255},
		},
		Dim: []RGBColor{
			{25, 35, 70},
			{40, 22, 60},
			{18, 50, 60},
			{45, 30, 70},
		},
		Bright: RGBColor{230, 235, 255},
	},
	"nord": {
		Name: "nord",
		Palette: []RGBColor{
			{136, 192, 208},
			{129, 161, 193},
			{143, 188, 187},
			{163, 190, 140},
		},
		Dim: []RGBColor{
			{46, 52, 64},
			{59, 66, 82},
			{67, 76, 94},
			{76, 86, 106},
		},
		Bright: RGBColor{236, 239, 244},
	},
	"dracula": {
		Name: "dracula",
		Palette: []RGBColor{
			{189, 147, 249},
			{255, 121, 198},
			{139, 233, 253},
			{80, 250, 123},
		},
		Dim: []RGBColor{
			{48, 34, 78},
			{68, 28, 52},
			{32, 58, 68},
			{18, 62, 32},
		},
		Bright: RGBColor{248, 248, 242},
	},
	"catppuccin": {
		Name: "catppuccin",
		Palette: []RGBColor{
			{203, 166, 247},
			{137, 180, 250},
			{116, 199, 236},
			{166, 227, 161},
		},
		Dim: []RGBColor{
			{49, 35, 62},
			{33, 46, 70},
			{28, 51, 60},
			{40, 58, 40},
		},
		Bright: RGBColor{205, 214, 244},
	},
	"gruvbox": {
		Name: "gruvbox",
		Palette: []RGBColor{
			{251, 189, 35},
			{184, 187, 38},
			{214, 93, 14},
			{104, 157, 106},
		},
		Dim: []RGBColor{
			{58, 44, 8},
			{44, 46, 10},
			{50, 22, 4},
			{26, 38, 24},
		},
		Bright: RGBColor{235, 219, 178},
	},
	"forest": {
		Name: "forest",
		Palette: []RGBColor{
			{80, 200, 90},
			{60, 160, 100},
			{160, 220, 80},
			{40, 180, 140},
		},
		Dim: []RGBColor{
			{14, 38, 16},
			{12, 30, 20},
			{33, 48, 14},
			{10, 38, 28},
		},
		Bright: RGBColor{200, 240, 180},
	},
	"wildberries": {
		Name: "wildberries",
		Palette: []RGBColor{
			{255,13,130},
			{144,124,255},
			{72,200,160},
			{255,160,120},
		},
		Dim: []RGBColor{
			{90,5,45},
			{60,50,120},
			{30,100,90},
			{120,70,50},
		},
		Bright: RGBColor{0,253,181},
     },
	"mono": {
		Name: "mono",
		Palette: []RGBColor{
			{0, 200, 80},
			{0, 170, 65},
			{0, 230, 95},
			{0, 150, 55},
		},
		Dim: []RGBColor{
			{0, 38, 14},
			{0, 28, 11},
			{0, 46, 16},
			{0, 24, 9},
		},
		Bright: RGBColor{180, 255, 200},
	},
	"rosepine": {
		Name: "rosepine",
		Palette: []RGBColor{
			{235, 111, 146},
			{196, 167, 231},
			{156, 207, 216},
			{49, 116, 143},
		},
		Dim: []RGBColor{
			{25, 23, 36},
			{31, 29, 46},
			{38, 35, 58},
			{28, 26, 42},
		},
		Bright: RGBColor{224, 222, 244},
	},
}

func ThemeNames() []string {
	names := make([]string, 0, len(Themes))
	for k := range Themes {
		names = append(names, k)
	}
	return names
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func clamp64(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
