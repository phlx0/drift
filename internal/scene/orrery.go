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

	// Hash multipliers for the deterministic RNG seed.
	orreryRNGSeedMulW int64 = 91841
	orreryRNGSeedMulH int64 = 69457
	orreryRNGSeedBase int64 = 0x77a31
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
	o.rng = rand.New(rand.NewSource(int64(w)*orreryRNGSeedMulW ^ int64(h)*orreryRNGSeedMulH ^ orreryRNGSeedBase))
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
	o.rng = rand.New(rand.NewSource(int64(w)*orreryRNGSeedMulW ^ int64(h)*orreryRNGSeedMulH ^ orreryRNGSeedBase))
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

func (o *Orrery) buildStars() {
	if o.pw <= 0 || o.ph <= 0 {
		o.stars = nil
		return
	}

	seed := int64(o.w)*44549 ^ int64(o.h)*98317 ^ 0x45a11
	rng := rand.New(rand.NewSource(seed))
	count := max((o.w*o.h)/90, 36)
	stars := make([]orreryStar, 0, count)

	addStar := func(xMin, xMax float64, idx int) {
		if xMax <= xMin {
			return
		}
		for attempt := 0; attempt < 8; attempt++ {
			x := xMin + rng.Float64()*(xMax-xMin)
			y := rng.Float64() * float64(o.ph)
			if math.Abs(x-o.centerX) < 20 && math.Abs(y-o.centerY) < 14 {
				continue
			}
			stars = append(stars, orreryStar{
				x:          x,
				y:          y,
				paletteIdx: uint8(idx % len(o.theme.Palette)),
				phase:      rng.Float64() * 2 * math.Pi,
				static:     rng.Intn(3) == 0,
			})
			return
		}
	}

	leftCount := count / 2
	rightCount := count - leftCount
	for i := 0; i < leftCount; i++ {
		addStar(0, o.centerX-12, i)
	}
	for i := 0; i < rightCount; i++ {
		addStar(o.centerX+12, float64(o.pw), leftCount+i)
	}
	o.stars = stars
}

func (o *Orrery) buildBodies() {
	count := o.effectiveBodyCount()
	seed := int64(o.w)*73856093 ^ int64(o.h)*19349663 ^ int64(count)*83492791
	rng := rand.New(rand.NewSource(seed))

	maxRadius := math.Min(float64(o.pw)*0.26, float64(o.ph)*0.26)
	minRadius := 9.0
	if maxRadius < minRadius+6 {
		maxRadius = minRadius + 6
	}
	span := maxRadius - minRadius

	sizePattern := []float64{1.0, 1.3, 1.7, 1.5, 2.0, 2.1, 2.1, 1.8}
	o.bodies = make([]orreryBody, count)
	o.orbitRadii = make([]float64, count)
	for i := 0; i < count; i++ {
		radius := minRadius + span*float64(i+1)/float64(count+1)
		speed := 1.45 / math.Pow(radius+4, 0.72)

		o.bodies[i] = orreryBody{
			radius:     radius,
			angle:      rng.Float64() * 2 * math.Pi,
			speed:      speed,
			size:       sizePattern[min(i, len(sizePattern)-1)],
			paletteIdx: i % len(o.theme.Palette),
			hasRing:    i == min(5, count-1),
		}
		o.orbitRadii[i] = radius
	}
}

func (o *Orrery) resetAsteroid() {
	o.asteroid = orreryAsteroid{
		active:   false,
		cooldown: 4 + o.rng.Float64()*5,
		size:     1.0,
	}
}

func (o *Orrery) resetUFO() {
	o.ufo = orreryUFO{
		active:   false,
		cooldown: 9 + o.rng.Float64()*10,
	}
}

func (o *Orrery) spawnAsteroid() {
	margin := 18.0
	side := o.rng.Intn(4)
	maxFlybyRadius := 24.0
	if len(o.orbitRadii) > 0 {
		maxFlybyRadius = o.orbitRadii[len(o.orbitRadii)-1] + 14
	}

	var x, y float64
	switch side {
	case 0:
		x = -margin
		y = o.rng.Float64() * float64(o.ph)
	case 1:
		x = float64(o.pw) + margin
		y = o.rng.Float64() * float64(o.ph)
	case 2:
		x = o.rng.Float64() * float64(o.pw)
		y = -margin
	default:
		x = o.rng.Float64() * float64(o.pw)
		y = float64(o.ph) + margin
	}

	size := 0.7 + o.rng.Float64()*1.5
	safeRadius := orrerySunExclusionRadius + size + orreryAsteroidClearance
	minMissDistance := safeRadius + 0.8
	targetX, targetY := x, y
	foundAny := false
	sunDX := o.centerX - x
	sunDY := o.centerY - y
	sunDist := math.Hypot(sunDX, sunDY)
	if sunDist < 0.000001 {
		sunDX, sunDY, sunDist = 1, 0, 1
	}
	sunUX := sunDX / sunDist
	sunUY := sunDY / sunDist

	for attempt := 0; attempt < 24; attempt++ {
		flybyRadius := (safeRadius + 1.2) + o.rng.Float64()*(maxFlybyRadius-(safeRadius+1.2))
		flybyAngle := o.rng.Float64() * 2 * math.Pi
		flybyX := o.centerX + math.Cos(flybyAngle)*flybyRadius
		flybyY := o.centerY + math.Sin(flybyAngle)*flybyRadius
		tangentX := -math.Sin(flybyAngle)
		tangentY := math.Cos(flybyAngle)
		lead := 10.0 + flybyRadius*0.9

		bestScore := math.MaxFloat64
		found := false
		for _, sign := range []float64{-1, 1} {
			candidateX := flybyX + tangentX*lead*sign
			candidateY := flybyY + tangentY*lead*sign
			miss := orreryPointSegmentDistance(o.centerX, o.centerY, x, y, candidateX, candidateY)
			if miss < minMissDistance {
				continue
			}

			dirX := candidateX - x
			dirY := candidateY - y
			dirDist := math.Hypot(dirX, dirY)
			if dirDist < 0.000001 {
				continue
			}
			dirX /= dirDist
			dirY /= dirDist

			radialDot := dirX*sunUX + dirY*sunUY
			if radialDot > 0.75 {
				continue
			}

			score := radialDot - miss*0.01
			if !found || score < bestScore {
				bestScore = score
				targetX = candidateX
				targetY = candidateY
				found = true
				foundAny = true
			}
		}

		if found {
			break
		}
	}

	if !foundAny {
		o.asteroid = orreryAsteroid{
			active:   false,
			cooldown: 0.8 + o.rng.Float64()*1.2,
			size:     size,
		}
		return
	}

	dx := targetX - x
	dy := targetY - y
	dist := math.Sqrt(dx*dx + dy*dy)
	speed := 14 + o.rng.Float64()*7

	o.asteroid = orreryAsteroid{
		active: true,
		x:      x,
		y:      y,
		vx:     dx / math.Max(dist, 0.01) * speed,
		vy:     dy / math.Max(dist, 0.01) * speed,
		size:   size,
	}
}

func (o *Orrery) updateAsteroid(dt float64) {
	if !o.asteroid.active {
		o.asteroid.cooldown -= dt
		if o.asteroid.cooldown <= 0 {
			o.spawnAsteroid()
		}
		return
	}

	speed := math.Sqrt(o.asteroid.vx*o.asteroid.vx + o.asteroid.vy*o.asteroid.vy)
	substeps := max(1, min(6, int(math.Ceil(speed*dt/3.0))))
	stepDT := dt / float64(substeps)
	safeRadius := orrerySunExclusionRadius + o.asteroid.size + orreryAsteroidClearance

	for step := 0; step < substeps; step++ {
		currDX := o.asteroid.x - o.centerX
		currDY := o.asteroid.y - o.centerY
		currDist := math.Hypot(currDX, currDY)
		if currDist < safeRadius {
			if currDist < 0.000001 {
				currDX, currDY, currDist = 1, 0, 1
			}

			ux := currDX / currDist
			uy := currDY / currDist
			o.asteroid.x = o.centerX + ux*(safeRadius+0.2)
			o.asteroid.y = o.centerY + uy*(safeRadius+0.2)

			normalVel := o.asteroid.vx*ux + o.asteroid.vy*uy
			if normalVel < 0 {
				o.asteroid.vx -= normalVel * ux * 1.2
				o.asteroid.vy -= normalVel * uy * 1.2
			}
		}

		dx := o.centerX - o.asteroid.x
		dy := o.centerY - o.asteroid.y
		dist2 := dx*dx + dy*dy
		dist := math.Sqrt(math.Max(dist2, 1))
		accel := 2200 / math.Max(dist2, 160)

		o.asteroid.vx += dx / dist * accel * stepDT
		o.asteroid.vy += dy / dist * accel * stepDT

		maxSpeed := 24.0
		speed = math.Sqrt(o.asteroid.vx*o.asteroid.vx + o.asteroid.vy*o.asteroid.vy)
		if speed > maxSpeed {
			o.asteroid.vx = o.asteroid.vx / speed * maxSpeed
			o.asteroid.vy = o.asteroid.vy / speed * maxSpeed
			speed = maxSpeed
		}

		currX := o.asteroid.x
		currY := o.asteroid.y
		nextX := currX + o.asteroid.vx*stepDT
		nextY := currY + o.asteroid.vy*stepDT

		if hit, t := orrerySegmentCircleHit(currX, currY, nextX, nextY, o.centerX, o.centerY, safeRadius); hit {
			hitX := currX + (nextX-currX)*t
			hitY := currY + (nextY-currY)*t
			nx := hitX - o.centerX
			ny := hitY - o.centerY
			norm := math.Hypot(nx, ny)
			if norm < 0.000001 {
				nx, ny, norm = 1, 0, 1
			}

			ux := nx / norm
			uy := ny / norm
			tx := -uy
			ty := ux
			normalVel := o.asteroid.vx*ux + o.asteroid.vy*uy
			tangentialVel := o.asteroid.vx*tx + o.asteroid.vy*ty
			outwardSpeed := math.Max(10.0, math.Max(-normalVel*1.05, speed*0.45))

			o.asteroid.vx = ux*outwardSpeed + tx*tangentialVel*0.55
			o.asteroid.vy = uy*outwardSpeed + ty*tangentialVel*0.55

			o.asteroid.x = hitX + ux*0.25
			o.asteroid.y = hitY + uy*0.25
			continue
		}

		o.asteroid.x = nextX
		o.asteroid.y = nextY
	}

	distFromSun := math.Hypot(o.asteroid.x-o.centerX, o.asteroid.y-o.centerY)
	if distFromSun > safeRadius+0.8 {
		o.stampDisc(o.asteroid.x, o.asteroid.y, clamp64(o.asteroid.size*0.6, 0.5, 1.5), orreryAsteroidOwner, 0.22, true)
	}

	margin := 26.0
	if o.asteroid.x < -margin || o.asteroid.x > float64(o.pw)+margin || o.asteroid.y < -margin || o.asteroid.y > float64(o.ph)+margin {
		o.resetAsteroid()
	}
}

func (o *Orrery) spawnUFO() {
	margin := 24.0
	side := o.rng.Intn(2)
	radius := o.orbitRadii[max(len(o.orbitRadii)-2, 0)] + 8 + o.rng.Float64()*10
	angle := -math.Pi + o.rng.Float64()*2*math.Pi
	targetX, targetY := o.pointOnOrbit(radius, angle)

	x := -margin
	if side == 1 {
		x = float64(o.pw) + margin
	}
	y := targetY - 8 - o.rng.Float64()*10
	if y < 10 {
		y = 10
	}
	if y > float64(o.ph)-10 {
		y = float64(o.ph) - 10
	}

	o.ufo = orreryUFO{
		active:     true,
		x:          x,
		y:          y,
		targetX:    targetX,
		targetY:    targetY,
		hoverTime:  1.4 + o.rng.Float64()*1.4,
		wobbleSeed: o.rng.Float64() * 2 * math.Pi,
	}
}

func (o *Orrery) updateUFO(dt float64) {
	if !o.ufo.active {
		o.ufo.cooldown -= dt
		if o.ufo.cooldown <= 0 {
			o.spawnUFO()
		}
		return
	}

	if !o.ufo.departing {
		dx := o.ufo.targetX - o.ufo.x
		dy := o.ufo.targetY - o.ufo.y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > 2.5 {
			speed := 16.0
			o.ufo.vx = dx / math.Max(dist, 0.01) * speed
			o.ufo.vy = dy / math.Max(dist, 0.01) * speed
			o.ufo.x += o.ufo.vx * dt
			o.ufo.y += o.ufo.vy * dt
		} else {
			o.ufo.hoverTime -= dt
			o.ufo.x = o.ufo.targetX + math.Sin(o.time*2.6+o.ufo.wobbleSeed)*1.8
			o.ufo.y = o.ufo.targetY + math.Sin(o.time*4.1+o.ufo.wobbleSeed)*0.8
			if o.ufo.hoverTime <= 0 {
				o.ufo.departing = true
				o.ufo.vx = 44 + o.rng.Float64()*18
				if o.ufo.x > o.centerX {
					o.ufo.vx *= -1
				}
				o.ufo.vy = -6 + o.rng.Float64()*12
			}
		}
	} else {
		o.ufo.x += o.ufo.vx * dt
		o.ufo.y += o.ufo.vy * dt
		o.ufo.vx *= math.Pow(1.02, dt*60)
	}

	margin := 30.0
	if o.ufo.x < -margin || o.ufo.x > float64(o.pw)+margin || o.ufo.y < -margin || o.ufo.y > float64(o.ph)+margin {
		o.resetUFO()
	}
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

func (o *Orrery) drawStars() {
	for _, star := range o.stars {
		var brightness float64
		if star.static {
			brightness = 0.24
		} else {
			brightness = 0.18 + 0.08*math.Sin(o.time*0.45+star.phase)
		}
		o.stampDisc(star.x, star.y, 0.35, star.paletteIdx, brightness, false)
	}
}

func (o *Orrery) drawOrbits() {
	for _, radius := range o.orbitRadii {
		samples := max(80, int(math.Ceil(radius*4.2)))
		for step := 0; step < samples; step++ {
			angle := (float64(step) / float64(samples)) * 2 * math.Pi
			x, y := o.pointOnOrbit(radius, angle)
			o.stampPixel(int(x+0.5), int(y+0.5), orreryOrbitOwner, 0.16)
		}
	}
}

func (o *Orrery) drawSun() {
	o.stampDisc(o.centerX, o.centerY, 4.1, orrerySunOwner, 1.0, false)
	o.stampDisc(o.centerX, o.centerY, 6.4, orrerySunOwner, 0.30, false)
}

func (o *Orrery) drawBody(body orreryBody) {
	haloRadius := body.size + 0.9
	if body.hasRing {
		haloRadius = body.size + 2.8
	}
	o.clearPlanetHalo(body.x, body.y, haloRadius)
	o.stampPlanetDisc(body.x, body.y, body.size, uint8(body.paletteIdx), 0.95)
	if body.hasRing {
		o.stampRing(body.x, body.y, body.size+1.4, body.size+2.4, uint8(body.paletteIdx), 0.42)
	}
}

func (o *Orrery) clearPlanetHalo(cx, cy, radius float64) {
	minCellX := max(0, int(math.Floor((cx-radius-1)/2)))
	maxCellX := min(o.w-1, int(math.Floor((cx+radius+1)/2)))
	minCellY := max(0, int(math.Floor((cy-radius-1)/4)))
	maxCellY := min(o.h-1, int(math.Floor((cy+radius+1)/4)))

	for cellX := minCellX; cellX <= maxCellX; cellX++ {
		for cellY := minCellY; cellY <= maxCellY; cellY++ {
			rectMinX := float64(cellX * 2)
			rectMaxX := rectMinX + 1
			rectMinY := float64(cellY * 4)
			rectMaxY := rectMinY + 3

			nearestX := clamp64(cx, rectMinX, rectMaxX)
			nearestY := clamp64(cy, rectMinY, rectMaxY)
			dx := nearestX - cx
			dy := nearestY - cy
			if math.Sqrt(dx*dx+dy*dy) > radius {
				continue
			}

			for subCol := 0; subCol < 2; subCol++ {
				for subRow := 0; subRow < 4; subRow++ {
					px := cellX*2 + subCol
					py := cellY*4 + subRow
					if px < 0 || px >= o.pw || py < 0 || py >= o.ph {
						continue
					}
					o.pixels[px][py] = 0
					o.pixelOwner[px][py] = 0
				}
			}
		}
	}
}

func (o *Orrery) drawAsteroid() {
	if !o.asteroid.active {
		return
	}

	drawX := o.asteroid.x
	drawY := o.asteroid.y
	safeRadius := orrerySunExclusionRadius + o.asteroid.size + orreryAsteroidClearance
	renderSafeRadius := safeRadius + 1.0
	dx := drawX - o.centerX
	dy := drawY - o.centerY
	dist := math.Hypot(dx, dy)
	if dist < renderSafeRadius {
		if dist < 0.000001 {
			dx, dy, dist = 1, 0, 1
		}
		ux := dx / dist
		uy := dy / dist
		drawX = o.centerX + ux*renderSafeRadius
		drawY = o.centerY + uy*renderSafeRadius
	}

	o.stampDisc(drawX, drawY, o.asteroid.size, orreryAsteroidOwner, 0.88, false)
}

func (o *Orrery) drawUFO() {
	if !o.ufo.active {
		return
	}

	o.stampEllipse(o.ufo.x, o.ufo.y+0.2, 4.2, 1.4, orreryUFOOwner, 0.72, false)
	o.stampEllipseRing(o.ufo.x, o.ufo.y+0.15, 4.0, 1.2, 4.8, 1.7, orreryUFOOwner, 0.28)
	o.stampEllipse(o.ufo.x, o.ufo.y-1.3, 1.7, 0.9, orreryUFODomeOwner, 0.64, false)
	o.stampEllipse(o.ufo.x, o.ufo.y+0.9, 2.3, 0.45, orreryUFOOwner, 0.40, false)
}
