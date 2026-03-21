package scene

import (
	"math"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

// particleGlyphs is the set of characters used to represent particles.
var particleGlyphs = []rune{'◦', '·', '○', '•', '.', '°', '∘'}

// Particles simulates a field of gently drifting particles driven by a
// sinusoidal flow field.  Particles leave fading trails as they move.
type Particles struct {
	w, h      int
	theme     Theme
	particles []particle

	// trail[x][y] brightness [0, 1].
	trail [][]float64
	time  float64
	rng   *rand.Rand

	gravity  float64
	friction float64
}

type particle struct {
	x, y       float64
	vx, vy     float64
	glyph      rune
	paletteIdx int
	phase      float64 // for individual shimmer
}

// NewParticles returns a fresh Particles scene.
func NewParticles() *Particles { return &Particles{gravity: 0, friction: 0.98} }

func (p *Particles) Name() string { return "particles" }

func (p *Particles) Init(w, h int, t Theme) {
	p.w, p.h = w, h
	p.theme = t
	p.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	p.rebuildTrail()

	count := 120
	p.particles = make([]particle, count)
	for i := range p.particles {
		p.particles[i] = p.newParticle(true)
	}
}

func (p *Particles) Resize(w, h int) {
	p.w, p.h = w, h
	p.rebuildTrail()
	for i := range p.particles {
		if p.particles[i].x >= float64(w) || p.particles[i].y >= float64(h) {
			p.particles[i] = p.newParticle(true)
		}
	}
}

func (p *Particles) rebuildTrail() {
	p.trail = make([][]float64, p.w)
	for x := range p.trail {
		p.trail[x] = make([]float64, p.h)
	}
}

func (p *Particles) newParticle(scattered bool) particle {
	speed := 0.4 + p.rng.Float64()*2.2
	angle := p.rng.Float64() * 2 * math.Pi
	x, y := 0.0, 0.0
	if scattered {
		x = p.rng.Float64() * float64(p.w)
		y = p.rng.Float64() * float64(p.h)
	} else {
		// Spawn at a random screen edge so particles flow inward.
		switch p.rng.Intn(4) {
		case 0:
			x, y = p.rng.Float64()*float64(p.w), 0
		case 1:
			x, y = p.rng.Float64()*float64(p.w), float64(p.h-1)
		case 2:
			x, y = 0, p.rng.Float64()*float64(p.h)
		case 3:
			x, y = float64(p.w-1), p.rng.Float64()*float64(p.h)
		}
	}
	return particle{
		x:          x,
		y:          y,
		vx:         math.Cos(angle) * speed,
		vy:         math.Sin(angle) * speed,
		glyph:      particleGlyphs[p.rng.Intn(len(particleGlyphs))],
		paletteIdx: p.rng.Intn(len(p.theme.Palette)),
		phase:      p.rng.Float64() * 2 * math.Pi,
	}
}

// flowField returns a gentle velocity nudge at (x, y, t) using overlapping
// sinusoids — a cheap but convincing organic flow without external deps.
func flowField(x, y, t float64) (fx, fy float64) {
	fx = math.Sin(x*0.04+t*0.25) * math.Cos(y*0.06+t*0.18) * 0.6
	fy = math.Cos(x*0.06+t*0.20) * math.Sin(y*0.04+t*0.22) * 0.6
	return
}

func (p *Particles) Update(dt float64) {
	p.time += dt

	// Decay trail brightness.
	decay := dt * 2.8
	for x := range p.trail {
		for y := range p.trail[x] {
			if v := p.trail[x][y] - decay; v > 0 {
				p.trail[x][y] = v
			} else {
				p.trail[x][y] = 0
			}
		}
	}

	for i := range p.particles {
		pt := &p.particles[i]

		// Apply flow field nudge.
		fx, fy := flowField(pt.x, pt.y, p.time)
		pt.vx += fx * dt
		pt.vy += fy * dt

		// Apply gravity (default 0).
		pt.vy += p.gravity * dt

		// Dampen velocity.
		pt.vx *= math.Pow(p.friction, dt*60)
		pt.vy *= math.Pow(p.friction, dt*60)

		// Clamp speed.
		speed := math.Sqrt(pt.vx*pt.vx + pt.vy*pt.vy)
		maxSpeed := 3.0
		if speed > maxSpeed {
			pt.vx = pt.vx / speed * maxSpeed
			pt.vy = pt.vy / speed * maxSpeed
		}

		pt.x += pt.vx * dt
		pt.y += pt.vy * dt
		pt.phase += dt * 1.2

		// Stamp trail at current position.
		ix, iy := int(pt.x+0.5), int(pt.y+0.5)
		if ix >= 0 && ix < p.w && iy >= 0 && iy < p.h {
			if p.trail[ix][iy] < 0.9 {
				p.trail[ix][iy] = 0.9
			}
		}

		// Respawn particles that stray off-screen.
		if pt.x < -2 || pt.x > float64(p.w)+2 ||
			pt.y < -2 || pt.y > float64(p.h)+2 {
			p.particles[i] = p.newParticle(false)
		}
	}
}

func (p *Particles) Draw(screen tcell.Screen) {
	// Render trails first so particles draw on top.
	for x := 0; x < p.w; x++ {
		for y := 0; y < p.h; y++ {
			b := p.trail[x][y]
			if b < 0.08 {
				continue
			}
			pIdx := (x + y) % len(p.theme.Dim)
			color := Lerp(p.theme.Dim[pIdx], p.theme.Palette[pIdx], b*0.45)
			screen.SetContent(x, y, '·', nil, color.Style())
		}
	}

	// Render particles.
	for _, pt := range p.particles {
		x, y := int(pt.x+0.5), int(pt.y+0.5)
		if x < 0 || x >= p.w || y < 0 || y >= p.h {
			continue
		}
		shimmer := 0.65 + 0.35*math.Sin(pt.phase)
		pIdx := pt.paletteIdx
		color := Lerp(p.theme.Palette[pIdx], p.theme.Bright, shimmer*0.5)
		screen.SetContent(x, y, pt.glyph, nil, color.Style())
	}
}
