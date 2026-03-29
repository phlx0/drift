package scene

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
)

// waveformAmplitudeScale converts configured amplitude to pixel height.
// At the default amplitude of 0.70 this yields ~22% of pixel height.
const waveformAmplitudeScale = 0.31

// brailleOffsets maps (row, col) inside a braille cell to its Unicode bit offset.
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
	freq  float64 // spatial cycles across the full width
	amp   float64 // amplitude in pixels (not characters)
	speed float64 // radians per second (positive = rightward)
	phase float64 // initial phase offset (radians)
	color RGBColor
}

// Waveform renders multiple breathing sine waves using braille Unicode
// characters for sub-character precision (2×4 dots per terminal cell).
type Waveform struct {
	w, h  int
	theme Theme

	// Pixel resolution: pw = w*2, ph = h*4
	pw, ph int

	// pixels[px][py] = 1-based wave index that owns the pixel (0 = empty).
	pixels [][]uint8
	layers []waveLayer

	time float64

	cfgLayers    int
	cfgAmplitude float64
	cfgSpeed     float64
}

func NewWaveform(cfg config.WaveformConfig) *Waveform {
	return &Waveform{
		cfgLayers:    cfg.Layers,
		cfgAmplitude: cfg.Amplitude,
		cfgSpeed:     cfg.Speed,
	}
}

func (wf *Waveform) Name() string { return "waveform" }

func (wf *Waveform) Init(w, h int, t Theme) {
	wf.w, wf.h = w, h
	wf.theme = t
	wf.pw, wf.ph = w*2, h*4
	wf.allocPixels()
	wf.buildLayers()
}

// buildLayers (re)builds the wave layer slice from stored config and theme.
// Called from Init and Resize so both paths stay in sync.
func (wf *Waveform) buildLayers() {
	// amplitude * 0.31 gives ~22% of pixel height at the default value of 0.70.
	baseAmp := float64(wf.ph) * wf.cfgAmplitude * waveformAmplitudeScale

	numLayers := wf.cfgLayers
	if numLayers < 1 {
		numLayers = 1
	}
	if numLayers > 3 {
		numLayers = 3
	}

	pal := wf.theme.Palette
	all := []waveLayer{
		{freq: 2.0, amp: baseAmp, speed: 0.45 * wf.cfgSpeed, phase: 0, color: pal[0%len(pal)]},
		{freq: 3.7, amp: baseAmp * 0.65, speed: -0.70 * wf.cfgSpeed, phase: math.Pi / 3, color: pal[1%len(pal)]},
		{freq: 1.2, amp: baseAmp * 0.45, speed: 0.28 * wf.cfgSpeed, phase: math.Pi, color: pal[2%len(pal)]},
	}
	wf.layers = all[:numLayers]
}

func (wf *Waveform) Resize(w, h int) {
	wf.w, wf.h = w, h
	wf.pw, wf.ph = w*2, h*4
	wf.allocPixels()
	wf.buildLayers()
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
	for px := range wf.pixels {
		for py := range wf.pixels[px] {
			wf.pixels[px][py] = 0
		}
	}

	centerPY := wf.ph / 2

	for li, layer := range wf.layers {
		// Amplitude breathes slowly over time.
		breathe := 0.72 + 0.28*math.Sin(wf.time*0.38+float64(li)*1.1)
		amp := layer.amp * breathe

		for px := 0; px < wf.pw; px++ {
			fx := float64(px) / float64(wf.pw)
			py := centerPY + int(amp*math.Sin(fx*layer.freq*2*math.Pi+layer.phase+wf.time*layer.speed))

			// 2-pixel thick line for visibility at small amplitudes.
			for offset := 0; offset <= 1; offset++ {
				yy := py + offset
				if yy >= 0 && yy < wf.ph {
					wf.pixels[px][yy] = uint8(li + 1) // 1-based
				}
			}
		}
	}

	// Convert pixel buffer → braille characters → screen.
	for cx := 0; cx < wf.w; cx++ {
		for cy := 0; cy < wf.h; cy++ {
			ch, style, ok := wf.buildBrailleCell(cx, cy)
			if ok {
				screen.SetContent(cx, cy, ch, nil, style)
			}
		}
	}
}

// buildBrailleCell samples the pixel buffer for the terminal cell at (cx, cy),
// returns the braille rune, its style, and whether the cell has any dots.
func (wf *Waveform) buildBrailleCell(cx, cy int) (rune, tcell.Style, bool) {
	var mask uint8
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
		return 0, tcell.StyleDefault, false
	}

	// Color from the wave that contributed the most dots.
	bestWave := 0
	for li := 1; li < len(wf.layers); li++ {
		if waveCounts[li] > waveCounts[bestWave] {
			bestWave = li
		}
	}

	color := wf.layers[bestWave].color
	totalBits := waveCounts[0] + waveCounts[1] + waveCounts[2]
	if totalBits > 0 {
		boost := clamp64(float64(totalBits)/8.0, 0, 1) * 0.35
		color = Lerp(color, wf.theme.Bright, boost)
	}

	return '\u2800' | rune(mask), color.Style(), true
}
