package orrery

import (
	"math"
	"math/rand"

	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

const (
	orreryUFODomeOwner  uint8 = 249
	orreryOrbitOwner    uint8 = 250
	orrerySunOwner      uint8 = 251
	orreryStarOwner     uint8 = 252
	orreryAsteroidOwner uint8 = 253
	orreryUFOOwner      uint8 = 254

	orrerySunExclusionRadius = 8.4
	orreryAsteroidClearance  = 0.15
)

type Orrery struct {
	w, h   int
	pw, ph int

	theme scene.Theme
	time  float64
	rng   *rand.Rand

	centerX, centerY float64

	orbitRadii []float64
	bodies     []orreryBody
	stars      []orreryStar
	asteroid   orreryAsteroid
	ufo        orreryUFO

	trail      [][]float64
	trailOwner [][]uint8
	pixels     [][]float64
	pixelOwner [][]uint8

	cfgBodies     int
	cfgTrailDecay float64
}

type orreryBody struct {
	radius     float64
	angle      float64
	speed      float64
	size       float64
	paletteIdx int
	hasRing    bool

	x, y float64
}

type orreryStar struct {
	x, y       float64
	paletteIdx uint8
	phase      float64
	static     bool
}

type orreryAsteroid struct {
	active   bool
	x, y     float64
	vx, vy   float64
	size     float64
	cooldown float64
}

type orreryUFO struct {
	active     bool
	x, y       float64
	vx, vy     float64
	targetX    float64
	targetY    float64
	hoverTime  float64
	cooldown   float64
	departing  bool
	wobbleSeed float64
}

func New(cfg config.OrreryConfig) *Orrery {
	return &Orrery{
		cfgBodies:     cfg.Bodies,
		cfgTrailDecay: cfg.TrailDecay,
	}
}

func (o *Orrery) Name() string { return "orrery" }

func (o *Orrery) Init(w, h int, t scene.Theme) {
	o.w, o.h = w, h
	o.pw, o.ph = w*2, h*4
	o.theme = t
	o.time = 0
	o.rng = rand.New(rand.NewSource(int64(w)*91841 ^ int64(h)*69457 ^ 0x77a31))
	o.centerX = float64(max(o.pw-1, 0)) / 2
	o.centerY = float64(max(o.ph-1, 0)) / 2
	o.allocBuffers()
	o.buildStars()
	o.buildBodies()
	o.resetAsteroid()
	o.resetUFO()
	o.updatePositions()
}

func (o *Orrery) Resize(w, h int) {
	oldAsteroid := o.asteroid
	oldUFO := o.ufo

	o.w, o.h = w, h
	o.pw, o.ph = w*2, h*4
	o.rng = rand.New(rand.NewSource(int64(w)*91841 ^ int64(h)*69457 ^ 0x77a31 ^ int64(o.time*1000)))
	o.centerX = float64(max(o.pw-1, 0)) / 2
	o.centerY = float64(max(o.ph-1, 0)) / 2
	o.allocBuffers()
	o.buildStars()
	o.buildBodies()
	if oldAsteroid.active {
		o.asteroid = oldAsteroid
	} else {
		o.resetAsteroid()
	}
	if oldUFO.active {
		o.ufo = oldUFO
	} else {
		o.resetUFO()
	}
	o.updatePositions()
}

func (o *Orrery) Update(dt float64) {
	o.time += dt

	decay := dt * o.trailDecay()
	for x := range o.trail {
		for y := range o.trail[x] {
			if v := o.trail[x][y] - decay; v > 0 {
				o.trail[x][y] = v
			} else {
				o.trail[x][y] = 0
			}
		}
	}

	for i := range o.bodies {
		body := &o.bodies[i]
		body.angle += body.speed * dt
		body.x, body.y = o.pointOnOrbit(body.radius, body.angle)
	}

	o.updateAsteroid(dt)
	o.updateUFO(dt)
}

func (o *Orrery) effectiveBodyCount() int {
	n := o.cfgBodies
	if n < 4 {
		n = 4
	}
	if n > 8 {
		n = 8
	}

	switch {
	case o.w < 28 || o.h < 10:
		return 4
	case o.w < 48 || o.h < 14:
		return min(n, 5)
	case o.w < 72 || o.h < 20:
		return min(n, 6)
	default:
		return n
	}
}

func (o *Orrery) trailDecay() float64 {
	return scene.Clamp64(o.cfgTrailDecay, 1.8, 8.0)
}

func (o *Orrery) updatePositions() {
	for i := range o.bodies {
		body := &o.bodies[i]
		body.x, body.y = o.pointOnOrbit(body.radius, body.angle)
	}
}

func (o *Orrery) pointOnOrbit(radius, angle float64) (float64, float64) {
	return o.centerX + radius*math.Cos(angle), o.centerY + radius*math.Sin(angle)
}

func (o *Orrery) allocBuffers() {
	o.trail = make([][]float64, o.pw)
	o.pixels = make([][]float64, o.pw)
	o.trailOwner = make([][]uint8, o.pw)
	o.pixelOwner = make([][]uint8, o.pw)
	for x := 0; x < o.pw; x++ {
		o.trail[x] = make([]float64, o.ph)
		o.pixels[x] = make([]float64, o.ph)
		o.trailOwner[x] = make([]uint8, o.ph)
		o.pixelOwner[x] = make([]uint8, o.ph)
	}
}

func (o *Orrery) clearScratch() {
	for x := range o.pixels {
		for y := range o.pixels[x] {
			o.pixels[x][y] = 0
			o.pixelOwner[x][y] = 0
		}
	}
}
