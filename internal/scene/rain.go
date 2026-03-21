package scene

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
)

type Rain struct {
	w, h  int
	theme Theme

	// grid[x][y] holds cell brightness [0, 1].
	// 1.0 = freshly lit, 0.0 = dark.
	grid  [][]float64
	drops []rainDrop
	time  float64
	rng   *rand.Rand

	charset []rune
	speed   float64

	cfgCharset string
	cfgDensity float64
	cfgSpeed   float64
}

type rainDrop struct {
	col        int
	y          float64 // float for smooth sub-cell movement
	speed      float64 // cells per second
	headChar   rune
	frameAge   int // frames since last char change
	paletteIdx int
}

func NewRain(cfg config.RainConfig) *Rain {
	return &Rain{
		cfgCharset: cfg.Charset,
		cfgDensity: cfg.Density,
		cfgSpeed:   cfg.Speed,
	}
}

func (r *Rain) Name() string { return "rain" }

func (r *Rain) dropCount(w int) int {
	if r.cfgDensity <= 0 {
		return 5
	}
	// At default density 0.4: w/3 + 5, scales linearly with density.
	return int(float64(w)*r.cfgDensity/1.2) + 5
}

func (r *Rain) Init(w, h int, t Theme) {
	r.w, r.h = w, h
	r.theme = t
	r.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.cfgCharset != "" {
		r.charset = []rune(r.cfgCharset)
	} else {
		r.charset = []rune("ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉ0123456789")
	}
	r.speed = r.cfgSpeed

	r.rebuildGrid()

	count := r.dropCount(w)
	r.drops = make([]rainDrop, count)
	for i := range r.drops {
		r.drops[i] = r.newDrop(true)
	}
}

func (r *Rain) Resize(w, h int) {
	r.w, r.h = w, h
	r.rebuildGrid()

	count := r.dropCount(w)
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

		d.frameAge++
		if d.frameAge >= 3+r.rng.Intn(4) {
			d.headChar = r.randomChar()
			d.frameAge = 0
		}

		if d.y > float64(r.h+trailLen) {
			r.drops[i] = r.newDrop(false)
		}
	}
}

func (r *Rain) Draw(screen tcell.Screen) {
	pal := r.theme.Palette
	dim := r.theme.Dim

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

	// Draw heads on top of the trail grid.
	for _, d := range r.drops {
		if d.col < 0 || d.col >= r.w {
			continue
		}
		hy := int(d.y)
		if hy >= 0 && hy < r.h {
			screen.SetContent(d.col, hy, d.headChar, nil, r.theme.Bright.Style())
		}
	}
}
