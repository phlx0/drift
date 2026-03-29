package scene

import "math"

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
