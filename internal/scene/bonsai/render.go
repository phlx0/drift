package bonsai

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/scene"
)

var (
	leafChars  = [2]rune{'*', '&'}
	petalChars = [2]rune{'·', '*'}
)

func (b *Bonsai) buildPot() int {
	pw := b.w / 3
	if pw > 23 {
		pw = 23
	}
	if pw < 11 {
		pw = 11
	}
	col := scene.Lerp(b.theme.Dim[0], b.theme.Bright, 0.38)
	left := (b.w - pw) / 2

	feet := b.h >= 10
	rim := b.h - 2
	if feet {
		rim = b.h - 3
	}

	for i := 0; i < pw; i++ {
		ch := '─'
		switch i {
		case 0:
			ch = '╭'
		case pw - 1:
			ch = '╮'
		}
		b.pot = append(b.pot, segment{x: left + i, y: rim, ch: ch, color: col})
	}
	for i := 1; i < pw-1; i++ {
		ch := '─'
		switch i {
		case 1:
			ch = '╰'
		case pw - 2:
			ch = '╯'
		}
		b.pot = append(b.pot, segment{x: left + i, y: rim + 1, ch: ch, color: col})
	}
	if feet {
		footCol := scene.Lerp(col, b.theme.Dim[0], 0.4)
		for _, x := range []int{left + 3, left + pw - 4} {
			b.pot = append(b.pot, segment{x: x, y: rim + 2, ch: '╵', color: footCol})
		}
	}

	return rim - 1
}

func (b *Bonsai) stampTrunkBase(x, y int) {
	col := b.woodColor(0)
	b.put(x-1, y, '/', col)
	b.put(x, y, '│', col)
	b.put(x+1, y, '\\', col)
}

func (b *Bonsai) stamp(px, py, cx, cy int, t *tip) {
	dx, dy := cx-px, cy-py
	steps := scene.AbsInt(dx)
	if scene.AbsInt(dy) > steps {
		steps = scene.AbsInt(dy)
	}
	if steps == 0 {
		return
	}

	col := b.woodColor(t.depth)
	width := 1
	if t.depth == 0 {
		width = trunkWidth(1 - float64(t.life)/float64(t.baseLife))
	}

	for i := 1; i <= steps; i++ {
		x := px + dx*i/steps
		y := py + dy*i/steps
		ch := branchChar(x-(px+dx*(i-1)/steps), y-(py+dy*(i-1)/steps))
		b.put(x, y, ch, col)

		if width >= 2 {
			b.put(x+1, y, ch, col)
		}
		if width >= 3 {
			b.put(x-1, y, ch, col)
		}
	}

	if t.depth >= 2 && b.rng.Float64() < 0.25 {
		b.putLeaf(cx, cy)
	}
}

func trunkWidth(prog float64) int {
	switch {
	case prog < 0.30:
		return 3
	case prog < 0.62:
		return 2
	default:
		return 1
	}
}

func branchChar(dx, dy int) rune {
	switch {
	case dx == 0:
		return '│'
	case dy == 0:
		return '─'
	case (dx > 0) == (dy < 0):
		return '/'
	default:
		return '\\'
	}
}

func (b *Bonsai) growPad(cx, cy, depth int) {
	rx := 3.0 + b.rng.Float64()*2.5
	ry := 1.0
	if depth <= 1 || b.rng.Intn(3) == 0 {
		ry = 2.0
	}

	for y := -int(ry); y <= int(ry); y++ {
		for x := -int(rx); x <= int(rx); x++ {
			fx, fy := float64(x)/rx, float64(y)/ry
			d := fx*fx + fy*fy
			if d > 1 {
				continue
			}
			if b.rng.Float64() < 0.15+0.55*d {
				continue
			}
			b.putLeaf(cx+x, cy+y)
		}
	}
}

func (b *Bonsai) putLeaf(x, y int) {
	col := b.theme.Palette[b.rng.Intn(len(b.theme.Palette))]
	if b.rng.Float64() < 0.10 {
		col = b.theme.Bright
	}
	ch := leafChars[b.rng.Intn(len(leafChars))]
	if b.put(x, y, ch, col) {
		b.leaves = append(b.leaves, segment{x: x, y: y, ch: ch, color: col})
	}
}

func (b *Bonsai) put(x, y int, ch rune, col scene.RGBColor) bool {
	if x < 0 || x >= b.w || y < 0 || y > b.groundY {
		return false
	}
	if len(b.segs) >= b.maxSegs {
		return false
	}
	b.segs = append(b.segs, segment{x: x, y: y, ch: ch, color: col})
	return true
}

func (b *Bonsai) woodColor(depth int) scene.RGBColor {
	i := depth % len(b.theme.Dim)
	t := 0.45 - float64(depth)*0.07
	return scene.Lerp(b.theme.Dim[i], b.theme.Palette[i%len(b.theme.Palette)], scene.Clamp64(t, 0, 1))
}

func (b *Bonsai) updatePetals(dt float64) {
	b.petalTimer += dt
	if b.petalTimer >= 0.4 && len(b.petals) < maxPetals && len(b.leaves) > 0 {
		b.petalTimer = 0
		b.spawnPetal()
	}

	alive := b.petals[:0]
	for _, p := range b.petals {
		p.phase += dt * 2.4
		p.y += p.fall * dt
		p.x += math.Sin(p.phase) * p.sway * dt
		if p.y > float64(b.h-1) || p.x < 0 || p.x >= float64(b.w) {
			continue
		}
		alive = append(alive, p)
	}
	b.petals = alive
}

func (b *Bonsai) spawnPetal() {
	src := b.leaves[b.rng.Intn(len(b.leaves))]
	b.petals = append(b.petals, petal{
		x:     float64(src.x),
		y:     float64(src.y),
		fall:  1.4 + b.rng.Float64()*1.8,
		sway:  1.5 + b.rng.Float64()*2.0,
		phase: b.rng.Float64() * math.Pi * 2,
		ch:    petalChars[b.rng.Intn(len(petalChars))],
		color: scene.Lerp(src.color, b.theme.Dim[0], 0.35),
	})
}

func (b *Bonsai) Draw(screen tcell.Screen) {
	if b.state == bonsaiFading {
		for _, s := range b.fadeList[:b.visible] {
			screen.SetContent(s.x, s.y, s.ch, nil, s.color.Style())
		}
	} else {
		for _, s := range b.pot {
			screen.SetContent(s.x, s.y, s.ch, nil, s.color.Style())
		}
		for _, s := range b.segs {
			screen.SetContent(s.x, s.y, s.ch, nil, s.color.Style())
		}
	}

	for _, p := range b.petals {
		screen.SetContent(int(p.x), int(p.y), p.ch, nil, p.color.Style())
	}
}
