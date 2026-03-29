package scene

import (
	"math"
	"math/rand"
)

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
