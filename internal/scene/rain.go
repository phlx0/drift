package scene

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Rain renders columns of falling characters with bright heads and fading
// trails, inspired by the classic matrix aesthetic.
type Rain struct {
	w, h  int
	theme Theme

	// grid[x][y] holds the brightness of the cell [0, 1].
	// 1.0 = freshly lit (raindrop just passed), 0.0 = dark.
	grid  [][]float64
	drops []rainDrop
	time  float64
	rng   *rand.Rand

	charset []rune
	speed   float64
}

type rainDrop struct {
	col       int
	y         float64  // current head position (float for smooth movement)
	speed     float64  // cells per second
	headChar  rune
	frameAge  int // frames since last char change
	paletteIdx int
}

// NewRain returns a fresh Rain scene.
func NewRain() *Rain { return &Rain{} }

func (r *Rain) Name() string { return "rain" }

func (r *Rain) Init(w, h int, t Theme) {
	r.w, r.h = w, h
	r.theme = t
	r.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	r.charset = []rune("ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉ0123456789")
	r.speed = 1.0

	r.rebuildGrid()

	// Number of concurrent drops scales with terminal width.
	count := w/3 + 5
	r.drops = make([]rainDrop, count)
	for i := range r.drops {
		r.drops[i] = r.newDrop(true)
	}
}

func (r *Rain) Resize(w, h int) {
	r.w, r.h = w, h
	r.rebuildGrid()

	count := w/3 + 5
	if len(r.drops) < count {
		extra := make([]rainDrop, count-len(r.drops))
		for i := range extra {
			extra[i] = r.newDrop(true)
		}
		r.drops = append(r.drops, extra...)
	} else {
		r.drops = r.drops[:count]
	}
}

func (r *Rain) rebuildGrid() {
	r.grid = make([][]float64, r.w)
	for x := range r.grid {
		r.grid[x] = make([]float64, r.h)
	}
}

func (r *Rain) newDrop(scattered bool) rainDrop {
	y := 0.0
	if scattered {
		y = r.rng.Float64() * float64(r.h)
	}
	speed := (5.0 + r.rng.Float64()*12.0) * r.speed
	return rainDrop{
		col:        r.rng.Intn(r.w),
		y:          y,
		speed:      speed,
		headChar:   r.randomChar(),
		paletteIdx: r.rng.Intn(len(r.theme.Palette)),
	}
}

func (r *Rain) randomChar() rune {
	return r.charset[r.rng.Intn(len(r.charset))]
}

func (r *Rain) Update(dt float64) {
	r.time += dt

	// Decay all cells toward darkness.
	decay := dt * 3.5
	for x := range r.grid {
		for y := range r.grid[x] {
			if v := r.grid[x][y] - decay; v > 0 {
				r.grid[x][y] = v
			} else {
				r.grid[x][y] = 0
			}
		}
	}

	trailLen := 14

	for i := range r.drops {
		d := &r.drops[i]

		// Guard against stale column index after resize.
		if d.col < 0 || d.col >= r.w {
			r.drops[i] = r.newDrop(false)
			d = &r.drops[i]
		}

		d.y += d.speed * dt

		// Illuminate trail behind the head.
		headY := int(d.y)
		for t := 0; t < trailLen; t++ {
			cy := headY - t
			if cy >= 0 && cy < r.h {
				brightness := 1.0 - float64(t)/float64(trailLen)
				if brightness > r.grid[d.col][cy] {
					r.grid[d.col][cy] = brightness
				}
			}
		}

		// Randomise head character occasionally for flicker.
		d.frameAge++
		if d.frameAge >= 3+r.rng.Intn(4) {
			d.headChar = r.randomChar()
			d.frameAge = 0
		}

		// Reset drop when it fully exits the screen.
		if d.y > float64(r.h+trailLen) {
			r.drops[i] = r.newDrop(false)
		}
	}
}

func (r *Rain) Draw(screen tcell.Screen) {
	pal := r.theme.Palette
	dim := r.theme.Dim

	// Render the brightness grid.
	for x := 0; x < r.w; x++ {
		for y := 0; y < r.h; y++ {
			b := r.grid[x][y]
			if b < 0.04 {
				continue
			}

			var ch rune
			var color RGBColor
			pIdx := x % len(pal)

			switch {
			case b > 0.85:
				ch = '│'
				color = Lerp(pal[pIdx], r.theme.Bright, (b-0.85)/0.15)
			case b > 0.55:
				ch = '╎'
				color = pal[pIdx]
			case b > 0.30:
				ch = '╌'
				color = Lerp(dim[pIdx%len(dim)], pal[pIdx], (b-0.30)/0.25)
			default:
				ch = '·'
				color = Lerp(dim[pIdx%len(dim)], pal[pIdx], b/0.30)
			}

			screen.SetContent(x, y, ch, nil, color.Style())
		}
	}

	// Draw bright head characters on top.
	for _, d := range r.drops {
		if d.col < 0 || d.col >= r.w {
			continue
		}
		hy := int(d.y)
		if hy >= 0 && hy < r.h {
			color := r.theme.Bright
			screen.SetContent(d.col, hy, d.headChar, nil, color.Style())
		}
	}
}
