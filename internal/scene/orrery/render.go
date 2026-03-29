package orrery

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/scene"
)

func (o *Orrery) Draw(screen tcell.Screen) {
	if o.w <= 0 || o.h <= 0 {
		return
	}

	o.clearScratch()
	o.drawStars()
	o.drawOrbits()

	for x := 0; x < o.pw; x++ {
		for y := 0; y < o.ph; y++ {
			if b := o.trail[x][y]; b > 0.05 {
				o.stampPixel(x, y, o.trailOwner[x][y], b)
			}
		}
	}

	o.drawSun()
	for _, body := range o.bodies {
		o.drawBody(body)
	}
	o.drawAsteroid()
	o.drawUFO()

	for cx := 0; cx < o.w; cx++ {
		for cy := 0; cy < o.h; cy++ {
			ch, style, ok := o.buildBrailleCell(cx, cy)
			if ok {
				screen.SetContent(cx, cy, ch, nil, style)
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

func (o *Orrery) stampDisc(cx, cy, radius float64, owner uint8, brightness float64, trail bool) {
	minX := int(cx - radius - 1)
	maxX := int(cx + radius + 1)
	minY := int(cy - radius - 1)
	maxY := int(cy + radius + 1)

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}
			falloff := 1 - dist/math.Max(radius, 0.01)
			value := brightness * (0.45 + falloff*0.55)
			if trail {
				o.stampTrailPixel(x, y, owner, value)
			} else {
				o.stampPixel(x, y, owner, value)
			}
		}
	}
}

func (o *Orrery) stampPlanetDisc(cx, cy, radius float64, owner uint8, brightness float64) {
	minX := int(cx - radius - 1)
	maxX := int(cx + radius + 1)
	minY := int(cy - radius - 1)
	maxY := int(cy + radius + 1)

	softEdge := 0.28
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				continue
			}

			value := brightness
			if dist > radius-softEdge {
				edgeT := (radius - dist) / math.Max(softEdge, 0.01)
				value = brightness * (0.55 + scene.Clamp64(edgeT, 0, 1)*0.45)
			}

			o.stampPixel(x, y, owner, value)
		}
	}
}

func (o *Orrery) stampRing(cx, cy, inner, outer float64, owner uint8, brightness float64) {
	minX := int(cx - outer - 1)
	maxX := int(cx + outer + 1)
	minY := int(cy - outer - 1)
	maxY := int(cy + outer + 1)

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < inner || dist > outer {
				continue
			}
			falloff := 1 - math.Abs(dist-(inner+outer)/2)/math.Max((outer-inner)/2, 0.01)
			o.stampPixel(x, y, owner, brightness*(0.5+falloff*0.5))
		}
	}
}

func (o *Orrery) stampEllipse(cx, cy, rx, ry float64, owner uint8, brightness float64, trail bool) {
	minX := int(cx - rx - 1)
	maxX := int(cx + rx + 1)
	minY := int(cy - ry - 1)
	maxY := int(cy + ry + 1)

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			dx := (float64(x) - cx) / math.Max(rx, 0.01)
			dy := (float64(y) - cy) / math.Max(ry, 0.01)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 1 {
				continue
			}
			falloff := 1 - dist
			value := brightness * (0.45 + falloff*0.55)
			if trail {
				o.stampTrailPixel(x, y, owner, value)
			} else {
				o.stampPixel(x, y, owner, value)
			}
		}
	}
}

func (o *Orrery) stampEllipseRing(cx, cy, innerRx, innerRy, outerRx, outerRy float64, owner uint8, brightness float64) {
	minX := int(cx - outerRx - 1)
	maxX := int(cx + outerRx + 1)
	minY := int(cy - outerRy - 1)
	maxY := int(cy + outerRy + 1)

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			outerDX := (float64(x) - cx) / math.Max(outerRx, 0.01)
			outerDY := (float64(y) - cy) / math.Max(outerRy, 0.01)
			outerDist := math.Sqrt(outerDX*outerDX + outerDY*outerDY)
			if outerDist > 1 {
				continue
			}

			innerDX := (float64(x) - cx) / math.Max(innerRx, 0.01)
			innerDY := (float64(y) - cy) / math.Max(innerRy, 0.01)
			innerDist := math.Sqrt(innerDX*innerDX + innerDY*innerDY)
			if innerDist < 1 {
				continue
			}

			falloff := 1 - outerDist
			o.stampPixel(x, y, owner, brightness*(0.5+falloff*0.5))
		}
	}
}

func (o *Orrery) stampTrailPixel(px, py int, owner uint8, brightness float64) {
	if px < 0 || px >= o.pw || py < 0 || py >= o.ph {
		return
	}
	if brightness > o.trail[px][py] {
		o.trail[px][py] = brightness
		o.trailOwner[px][py] = owner
	}
}

func (o *Orrery) stampPixel(px, py int, owner uint8, brightness float64) {
	if px < 0 || px >= o.pw || py < 0 || py >= o.ph {
		return
	}
	if brightness > o.pixels[px][py] {
		o.pixels[px][py] = brightness
		o.pixelOwner[px][py] = owner
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

			mask |= uint8(1) << scene.BrailleOffsets[subRow][subCol]
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

	var color scene.RGBColor
	switch {
	case sunEnergy >= orbitEnergy && sunEnergy >= starEnergy && sunEnergy >= paletteMax:
		color = scene.Lerp(o.theme.Palette[0], o.theme.Bright, scene.Clamp64(0.72+sunEnergy*0.18, 0, 1))
	case ufoDomeEnergy >= ufoEnergy && ufoDomeEnergy >= asteroidEnergy && ufoDomeEnergy >= orbitEnergy && ufoDomeEnergy >= starEnergy && ufoDomeEnergy >= paletteMax:
		color = scene.Lerp(o.theme.Palette[1%len(o.theme.Palette)], o.theme.Bright, scene.Clamp64(0.52+ufoDomeEnergy*0.24, 0, 1))
	case ufoEnergy >= asteroidEnergy && ufoEnergy >= orbitEnergy && ufoEnergy >= starEnergy && ufoEnergy >= paletteMax:
		color = scene.Lerp(o.theme.Palette[3%len(o.theme.Palette)], o.theme.Bright, scene.Clamp64(0.56+ufoEnergy*0.22, 0, 1))
	case asteroidEnergy >= orbitEnergy && asteroidEnergy >= starEnergy && asteroidEnergy >= paletteMax:
		color = scene.Lerp(o.theme.Palette[2%len(o.theme.Palette)], o.theme.Bright, scene.Clamp64(0.42+asteroidEnergy*0.24, 0, 1))
	case orbitEnergy >= starEnergy && orbitEnergy >= paletteMax:
		color = scene.Lerp(o.theme.Dim[0], o.theme.Bright, scene.Clamp64(orbitEnergy*0.35, 0.18, 0.4))
	case starEnergy >= paletteMax:
		color = scene.Lerp(o.theme.Dim[1%len(o.theme.Dim)], o.theme.Palette[1%len(o.theme.Palette)], scene.Clamp64(starEnergy*0.45, 0.12, 0.28))
	default:
		best := 0
		for i := 1; i < len(paletteEnergy); i++ {
			if paletteEnergy[i] > paletteEnergy[best] {
				best = i
			}
		}
		color = scene.Lerp(o.theme.Palette[best], o.theme.Bright, scene.Clamp64(total*0.16, 0, 0.32))
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

func orreryPointSegmentDistance(px, py, x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	den := dx*dx + dy*dy
	if den <= 0.000001 {
		return math.Hypot(px-x1, py-y1)
	}

	t := ((px-x1)*dx + (py-y1)*dy) / den
	t = scene.Clamp64(t, 0, 1)
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
