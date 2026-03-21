package scene

import (
	"math"

	"github.com/gdamore/tcell/v2"
)

// brailleOffsets[row][col] gives the Unicode bit offset for the dot at
// (col, row) inside a braille cell (col ∈ {0,1}, row ∈ {0,1,2,3}).
//
// Unicode braille bit layout (U+2800 base):
//
//	col 0   col 1
//	bit 0   bit 3   ← row 0 (top)
//	bit 1   bit 4   ← row 1
//	bit 2   bit 5   ← row 2
//	bit 6   bit 7   ← row 3 (bottom)
var brailleOffsets = [4][2]uint{
	{0, 3},
	{1, 4},
	{2, 5},
	{6, 7},
}

// waveLayer is one sine wave in the stack.
type waveLayer struct {
	freq  float64  // spatial cycles across the full width
	amp   float64  // base amplitude in pixels (not characters)
	speed float64  // radians per second (positive = rightward)
	phase float64  // initial phase offset (radians)
	color RGBColor
}

// Waveform renders multiple breathing sine waves using braille Unicode
// characters for sub-character precision, giving smooth, animated curves.
type Waveform struct {
	w, h int
	theme Theme

	// Pixel resolution: pw = w*2, ph = h*4
	pw, ph int

	// pixels[px][py] = 1-based wave index that owns the pixel (0 = empty).
	pixels [][]uint8
	layers []waveLayer

	time float64
}

// NewWaveform returns a fresh Waveform scene.
func NewWaveform() *Waveform { return &Waveform{} }

func (wf *Waveform) Name() string { return "waveform" }

func (wf *Waveform) Init(w, h int, t Theme) {
	wf.w, wf.h = w, h
	wf.theme = t
	wf.pw, wf.ph = w*2, h*4
	wf.allocPixels()

	baseAmp := float64(wf.ph) * 0.22 // ~22% of screen height in pixels
	wf.layers = []waveLayer{
		{
			freq:  2.0,
			amp:   baseAmp,
			speed: 0.45,
			phase: 0,
			color: t.Palette[0%len(t.Palette)],
		},
		{
			freq:  3.7,
			amp:   baseAmp * 0.65,
			speed: -0.70,
			phase: math.Pi / 3,
			color: t.Palette[1%len(t.Palette)],
		},
		{
			freq:  1.2,
			amp:   baseAmp * 0.45,
			speed: 0.28,
			phase: math.Pi,
			color: t.Palette[2%len(t.Palette)],
		},
	}
}

func (wf *Waveform) Resize(w, h int) {
	wf.w, wf.h = w, h
	wf.pw, wf.ph = w*2, h*4
	wf.allocPixels()

	baseAmp := float64(wf.ph) * 0.22
	for i := range wf.layers {
		switch i {
		case 0:
			wf.layers[i].amp = baseAmp
		case 1:
			wf.layers[i].amp = baseAmp * 0.65
		case 2:
			wf.layers[i].amp = baseAmp * 0.45
		}
	}
}

func (wf *Waveform) allocPixels() {
	wf.pixels = make([][]uint8, wf.pw)
	for i := range wf.pixels {
		wf.pixels[i] = make([]uint8, wf.ph)
	}
}

func (wf *Waveform) Update(dt float64) {
	wf.time += dt
}

func (wf *Waveform) Draw(screen tcell.Screen) {
	// Clear pixel buffer.
	for px := range wf.pixels {
		for py := range wf.pixels[px] {
			wf.pixels[px][py] = 0
		}
	}

	centerPY := wf.ph / 2

	// Paint each wave layer into the pixel buffer.
	for li, layer := range wf.layers {
		// Breathing: amplitude slowly oscillates.
		breathe := 0.72 + 0.28*math.Sin(wf.time*0.38+float64(li)*1.1)
		amp := layer.amp * breathe

		for px := 0; px < wf.pw; px++ {
			// Map pixel column to [0, 1] and evaluate sine.
			fx := float64(px) / float64(wf.pw)
			py := centerPY + int(amp*math.Sin(fx*layer.freq*2*math.Pi+layer.phase+wf.time*layer.speed))

			// Draw 2-pixel thick line for visibility.
			for offset := 0; offset <= 1; offset++ {
				yy := py + offset
				if yy >= 0 && yy < wf.ph {
					wf.pixels[px][yy] = uint8(li + 1) // 1-based
				}
			}
		}
	}

	// Convert pixel buffer → braille cells → screen.
	for cx := 0; cx < wf.w; cx++ {
		for cy := 0; cy < wf.h; cy++ {
			var mask uint8
			// waveCounts[li] = number of bits set by layer li in this cell.
			var waveCounts [3]int

			for subRow := 0; subRow < 4; subRow++ {
				for subCol := 0; subCol < 2; subCol++ {
					px := cx*2 + subCol
					py := cy*4 + subRow
					if px >= wf.pw || py >= wf.ph {
						continue
					}
					waveIdx := wf.pixels[px][py]
					if waveIdx == 0 {
						continue
					}
					bitOffset := brailleOffsets[subRow][subCol]
					mask |= uint8(1) << bitOffset
					if int(waveIdx-1) < len(waveCounts) {
						waveCounts[waveIdx-1]++
					}
				}
			}

			if mask == 0 {
				continue
			}

			// Choose color from the wave that contributed the most bits.
			bestWave := 0
			for li := 1; li < len(wf.layers); li++ {
				if waveCounts[li] > waveCounts[bestWave] {
					bestWave = li
				}
			}

			color := wf.layers[bestWave].color
			// Add a slight brightness boost toward Bright based on density.
			totalBits := waveCounts[0] + waveCounts[1] + waveCounts[2]
			if totalBits > 0 {
				boost := clamp64(float64(totalBits)/8.0, 0, 1) * 0.35
				color = Lerp(color, wf.theme.Bright, boost)
			}

			ch := '\u2800' | rune(mask)
			screen.SetContent(cx, cy, ch, nil, color.Style())
		}
	}
}
