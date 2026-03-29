package scene

import (
	"math"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
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

	theme Theme
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

func NewOrrery(cfg config.OrreryConfig) *Orrery {
	return &Orrery{
		cfgBodies:     cfg.Bodies,
		cfgTrailDecay: cfg.TrailDecay,
	}
}

func (o *Orrery) Name() string { return "orrery" }

func (o *Orrery) Init(w, h int, t Theme) {
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
	o.rng = rand.New(rand.NewSource(int64(w)*91841 ^ int64(h)*69457 ^ 0x77a31))
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
	return clamp64(o.cfgTrailDecay, 1.8, 8.0)
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

func orreryPointSegmentDistance(px, py, x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	den := dx*dx + dy*dy
	if den <= 0.000001 {
		return math.Hypot(px-x1, py-y1)
	}

	t := ((px-x1)*dx + (py-y1)*dy) / den
	t = clamp64(t, 0, 1)
	sx := x1 + t*dx
	sy := y1 + t*dy
	return math.Hypot(px-sx, py-sy)
}

func orrerySegmentCircleHit(x1, y1, x2, y2, cx, cy, r float64) (bool, float64) {
	dx := x2 - x1
	dy := y2 - y1
	fx := x1 - cx
	fy := y1 - cy

	a := dx*dx + dy*dy
	if a <= 0.000001 {
		return false, 0
	}

	b := 2 * (fx*dx + fy*dy)
	c := fx*fx + fy*fy - r*r
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return false, 0
	}

	sqrtDisc := math.Sqrt(discriminant)
	t1 := (-b - sqrtDisc) / (2 * a)
	t2 := (-b + sqrtDisc) / (2 * a)

	switch {
	case t1 >= 0 && t1 <= 1:
		return true, t1
	case t2 >= 0 && t2 <= 1:
		return true, t2
	default:
		return false, 0
	}
}

func (o *Orrery) buildBrailleCell(cx, cy int) (rune, tcell.Style, bool) {
	var mask uint8
	var paletteEnergy [8]float64
	orbitEnergy := 0.0
	sunEnergy := 0.0
	starEnergy := 0.0
	asteroidEnergy := 0.0
	ufoEnergy := 0.0
	ufoDomeEnergy := 0.0
	total := 0.0

	for subRow := 0; subRow < 4; subRow++ {
		for subCol := 0; subCol < 2; subCol++ {
			px := cx*2 + subCol
			py := cy*4 + subRow
			if px >= o.pw || py >= o.ph {
				continue
			}

			brightness := o.pixels[px][py]
			if brightness < 0.10 {
				continue
			}

			mask |= uint8(1) << brailleOffsets[subRow][subCol]
			total += brightness

			switch owner := o.pixelOwner[px][py]; {
			case owner == orrerySunOwner:
				sunEnergy += brightness
			case owner == orreryOrbitOwner:
				orbitEnergy += brightness
			case owner == orreryStarOwner:
				starEnergy += brightness
			case owner == orreryAsteroidOwner:
				asteroidEnergy += brightness
			case owner == orreryUFODomeOwner:
				ufoDomeEnergy += brightness
			case owner == orreryUFOOwner:
				ufoEnergy += brightness
			default:
				paletteEnergy[int(owner)%len(paletteEnergy)] += brightness
			}
		}
	}

	if mask == 0 {
		return 0, tcell.StyleDefault, false
	}

	paletteMax := maxPaletteEnergy(paletteEnergy)

	var color RGBColor
	switch {
	case sunEnergy >= orbitEnergy && sunEnergy >= starEnergy && sunEnergy >= paletteMax:
		color = Lerp(o.theme.Palette[0], o.theme.Bright, clamp64(0.72+sunEnergy*0.18, 0, 1))
	case ufoDomeEnergy >= ufoEnergy && ufoDomeEnergy >= asteroidEnergy && ufoDomeEnergy >= orbitEnergy && ufoDomeEnergy >= starEnergy && ufoDomeEnergy >= paletteMax:
		color = Lerp(o.theme.Palette[1%len(o.theme.Palette)], o.theme.Bright, clamp64(0.52+ufoDomeEnergy*0.24, 0, 1))
	case ufoEnergy >= asteroidEnergy && ufoEnergy >= orbitEnergy && ufoEnergy >= starEnergy && ufoEnergy >= paletteMax:
		color = Lerp(o.theme.Palette[3%len(o.theme.Palette)], o.theme.Bright, clamp64(0.56+ufoEnergy*0.22, 0, 1))
	case asteroidEnergy >= orbitEnergy && asteroidEnergy >= starEnergy && asteroidEnergy >= paletteMax:
		color = Lerp(o.theme.Palette[2%len(o.theme.Palette)], o.theme.Bright, clamp64(0.42+asteroidEnergy*0.24, 0, 1))
	case orbitEnergy >= starEnergy && orbitEnergy >= paletteMax:
		color = Lerp(o.theme.Dim[0], o.theme.Bright, clamp64(orbitEnergy*0.35, 0.18, 0.4))
	case starEnergy >= paletteMax:
		color = Lerp(o.theme.Dim[1%len(o.theme.Dim)], o.theme.Palette[1%len(o.theme.Palette)], clamp64(starEnergy*0.45, 0.12, 0.28))
	default:
		best := 0
		for i := 1; i < len(paletteEnergy); i++ {
			if paletteEnergy[i] > paletteEnergy[best] {
				best = i
			}
		}
		color = Lerp(o.theme.Palette[best], o.theme.Bright, clamp64(total*0.16, 0, 0.32))
	}

	return '\u2800' | rune(mask), color.Style(), true
}

func maxPaletteEnergy(vals [8]float64) float64 {
	best := 0.0
	for _, v := range vals {
		if v > best {
			best = v
		}
	}
	return best
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
