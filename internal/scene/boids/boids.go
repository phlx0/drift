package boids

import (
	"math"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

const (
	sepRadius     = 3.5
	alignRadius   = 8.0
	cohRadius     = 14.0
	sepWeight     = 3.0
	alignStrength = 2.0
	cohStrength   = 0.15
	boundMargin   = 2.5
	boundForce    = 12.0
	maxBoidSpeed  = 8.0
	minBoidSpeed  = 2.5
)

type Boids struct {
	w, h  int
	theme scene.Theme
	boids []boid
	trail [][]float64
	time  float64
	rng   *rand.Rand

	cfgCount int
	cfgSpeed float64
}

type boid struct {
	x, y       float64
	vx, vy     float64
	paletteIdx int
}

func New(cfg config.BoidsConfig) *Boids {
	return &Boids{
		cfgCount: cfg.Count,
		cfgSpeed: cfg.Speed,
	}
}

func (b *Boids) Name() string { return "boids" }

func (b *Boids) Init(w, h int, t scene.Theme) {
	b.w, b.h = w, h
	b.theme = t
	b.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	b.rebuildTrail()
	b.boids = make([]boid, b.cfgCount)
	for i := range b.boids {
		b.boids[i] = b.newBoid()
	}
}

func (b *Boids) Resize(w, h int) {
	b.w, b.h = w, h
	b.rebuildTrail()
	fw, fh := float64(w), float64(h)
	for i := range b.boids {
		bd := &b.boids[i]
		if bd.x >= fw || bd.y >= fh || bd.x < 0 || bd.y < 0 {
			b.boids[i] = b.newBoid()
		}
	}
}

func (b *Boids) rebuildTrail() {
	b.trail = make([][]float64, b.w)
	for x := range b.trail {
		b.trail[x] = make([]float64, b.h)
	}
}

func (b *Boids) newBoid() boid {
	angle := b.rng.Float64() * 2 * math.Pi
	speed := (minBoidSpeed + b.rng.Float64()*(maxBoidSpeed-minBoidSpeed)) * b.cfgSpeed
	return boid{
		x:          1 + b.rng.Float64()*float64(b.w-2),
		y:          1 + b.rng.Float64()*float64(b.h-2),
		vx:         math.Cos(angle) * speed,
		vy:         math.Sin(angle) * speed,
		paletteIdx: b.rng.Intn(len(b.theme.Palette)),
	}
}

func (b *Boids) Update(dt float64) {
	b.time += dt
	if b.w < 5 || b.h < 3 {
		return
	}

	decay := dt * 3.0
	for x := range b.trail {
		for y := range b.trail[x] {
			if v := b.trail[x][y] - decay; v > 0 {
				b.trail[x][y] = v
			} else {
				b.trail[x][y] = 0
			}
		}
	}

	maxSpd := maxBoidSpeed * b.cfgSpeed
	minSpd := minBoidSpeed * b.cfgSpeed
	fw, fh := float64(b.w), float64(b.h)

	for i := range b.boids {
		bd := &b.boids[i]

		var sepX, sepY float64
		var alignVX, alignVY float64
		var cohX, cohY float64
		var sepN, alignN, cohN int

		for j := range b.boids {
			if i == j {
				continue
			}
			o := &b.boids[j]
			dx := o.x - bd.x
			dy := o.y - bd.y
			d := math.Sqrt(dx*dx + dy*dy)

			if d < sepRadius && d > 0.01 {
				sepX -= dx / d
				sepY -= dy / d
				sepN++
			}
			if d < alignRadius {
				alignVX += o.vx
				alignVY += o.vy
				alignN++
			}
			if d < cohRadius {
				cohX += o.x
				cohY += o.y
				cohN++
			}
		}

		var ax, ay float64

		if sepN > 0 {
			ax += (sepX / float64(sepN)) * sepWeight
			ay += (sepY / float64(sepN)) * sepWeight
		}
		if alignN > 0 {
			ax += (alignVX/float64(alignN) - bd.vx) * alignStrength
			ay += (alignVY/float64(alignN) - bd.vy) * alignStrength
		}
		if cohN > 0 {
			ax += (cohX/float64(cohN) - bd.x) * cohStrength
			ay += (cohY/float64(cohN) - bd.y) * cohStrength
		}

		if bd.x < boundMargin {
			ax += boundForce * (1.0 - bd.x/boundMargin)
		} else if bd.x > fw-boundMargin {
			ax -= boundForce * (1.0 - (fw-bd.x)/boundMargin)
		}
		if bd.y < boundMargin {
			ay += boundForce * (1.0 - bd.y/boundMargin)
		} else if bd.y > fh-boundMargin {
			ay -= boundForce * (1.0 - (fh-bd.y)/boundMargin)
		}

		bd.vx += ax * dt
		bd.vy += ay * dt

		spd := math.Sqrt(bd.vx*bd.vx + bd.vy*bd.vy)
		if spd > maxSpd {
			bd.vx = bd.vx / spd * maxSpd
			bd.vy = bd.vy / spd * maxSpd
			spd = maxSpd
		}
		if spd < minSpd {
			if spd > 1e-9 {
				bd.vx = bd.vx / spd * minSpd
				bd.vy = bd.vy / spd * minSpd
			} else {
				angle := b.rng.Float64() * 2 * math.Pi
				bd.vx = math.Cos(angle) * minSpd
				bd.vy = math.Sin(angle) * minSpd
			}
		}

		bd.x += bd.vx * dt
		bd.y += bd.vy * dt

		bd.x = scene.Clamp64(bd.x, 0, fw-0.01)
		bd.y = scene.Clamp64(bd.y, 0, fh-0.01)

		ix, iy := int(bd.x+0.5), int(bd.y+0.5)
		if ix >= 0 && ix < b.w && iy >= 0 && iy < b.h {
			if b.trail[ix][iy] < 0.85 {
				b.trail[ix][iy] = 0.85
			}
		}
	}
}

func dirGlyph(vx, vy float64) rune {
	angle := math.Atan2(vy, vx)
	switch {
	case angle >= -math.Pi/4 && angle < math.Pi/4:
		return '▸'
	case angle >= math.Pi/4 && angle < 3*math.Pi/4:
		return '▾'
	case angle >= -3*math.Pi/4 && angle < -math.Pi/4:
		return '▴'
	default:
		return '◂'
	}
}

func (b *Boids) Draw(screen tcell.Screen) {
	if b.w < 10 || b.h < 3 {
		return
	}
	for x := 0; x < b.w; x++ {
		for y := 0; y < b.h; y++ {
			br := b.trail[x][y]
			if br < 0.05 {
				continue
			}
			pIdx := (x + y) % len(b.theme.Dim)
			color := scene.Lerp(b.theme.Dim[pIdx], b.theme.Palette[pIdx], br*0.45)
			screen.SetContent(x, y, '·', nil, color.Style())
		}
	}
	for _, bd := range b.boids {
		x, y := int(bd.x+0.5), int(bd.y+0.5)
		if x < 0 || x >= b.w || y < 0 || y >= b.h {
			continue
		}
		pIdx := bd.paletteIdx % len(b.theme.Palette)
		color := scene.Lerp(b.theme.Palette[pIdx], b.theme.Bright, 0.4)
		screen.SetContent(x, y, dirGlyph(bd.vx, bd.vy), nil, color.Style())
	}
}
