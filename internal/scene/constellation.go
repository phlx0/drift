package scene

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
)

// starGlyphs is an ordered rune slice from faintest to brightest.
// Each star is assigned one based on its base brightness.
var starGlyphs = []rune{'·', '∘', '○', '◦', '*', '✦'}

// Constellation animates a field of slowly drifting stars that connect to
// nearby neighbours with thin dotted lines.
type Constellation struct {
	w, h  int
	theme Theme
	stars []star
	time  float64
	rng   *rand.Rand

	connectDist    float64 // pixel distance threshold for connections
	maxConnections int
}

type star struct {
	x, y         float64
	vx, vy       float64
	twinkle      float64 // phase, radians
	twinkleFreq  float64 // radians per second
	glyphIdx     int
	paletteIdx   int
}

// NewConstellation returns a fresh Constellation scene.
func NewConstellation() *Constellation { return &Constellation{} }

func (c *Constellation) Name() string { return "constellation" }

func (c *Constellation) Init(w, h int, t Theme) {
	c.w, c.h = w, h
	c.theme = t
	c.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	c.connectDist = math.Sqrt(float64(w*w+h*h)) * 0.18
	c.maxConnections = 4

	count := 80
	c.stars = make([]star, count)
	for i := range c.stars {
		c.stars[i] = c.randomStar(true)
	}
}

func (c *Constellation) Resize(w, h int) {
	c.w, c.h = w, h
	c.connectDist = math.Sqrt(float64(w*w+h*h)) * 0.18
	// Re-scatter stars that are now outside the new bounds.
	for i := range c.stars {
		s := &c.stars[i]
		if s.x >= float64(w) || s.y >= float64(h) {
			*s = c.randomStar(true)
		}
	}
}

func (c *Constellation) randomStar(scattered bool) star {
	speed := 0.3 + c.rng.Float64()*1.2
	angle := c.rng.Float64() * 2 * math.Pi
	x, y := 0.0, 0.0
	if scattered {
		x = c.rng.Float64() * float64(c.w)
		y = c.rng.Float64() * float64(c.h)
	} else {
		// Spawn at a random edge.
		switch c.rng.Intn(4) {
		case 0:
			x, y = c.rng.Float64()*float64(c.w), 0
		case 1:
			x, y = c.rng.Float64()*float64(c.w), float64(c.h-1)
		case 2:
			x, y = 0, c.rng.Float64()*float64(c.h)
		case 3:
			x, y = float64(c.w-1), c.rng.Float64()*float64(c.h)
		}
	}
	return star{
		x:           x,
		y:           y,
		vx:          math.Cos(angle) * speed,
		vy:          math.Sin(angle) * speed,
		twinkle:     c.rng.Float64() * 2 * math.Pi,
		twinkleFreq: 0.4 + c.rng.Float64()*0.8,
		glyphIdx:    c.rng.Intn(len(starGlyphs)),
		paletteIdx:  c.rng.Intn(len(c.theme.Palette)),
	}
}

func (c *Constellation) Update(dt float64) {
	c.time += dt
	fw, fh := float64(c.w), float64(c.h)
	for i := range c.stars {
		s := &c.stars[i]
		s.x += s.vx * dt
		s.y += s.vy * dt
		s.twinkle += s.twinkleFreq * dt

		// Elastic bounce at edges.
		if s.x < 0 {
			s.x = -s.x
			s.vx = -s.vx
		} else if s.x >= fw {
			s.x = 2*fw - s.x - 0.01
			s.vx = -s.vx
		}
		if s.y < 0 {
			s.y = -s.y
			s.vy = -s.vy
		} else if s.y >= fh {
			s.y = 2*fh - s.y - 0.01
			s.vy = -s.vy
		}
	}
}

func (c *Constellation) Draw(screen tcell.Screen) {
	// Pre-sort connection candidates per star (limit to maxConnections closest
	// neighbours) so dense screens don't become a wall of lines.
	type edge struct{ i, j int; dist float64 }
	edges := make([]edge, 0, len(c.stars)*c.maxConnections)
	connCount := make([]int, len(c.stars))

	for i := 0; i < len(c.stars); i++ {
		type nb struct {
			j    int
			dist float64
		}
		nbs := make([]nb, 0, 8)
		for j := i + 1; j < len(c.stars); j++ {
			dx := c.stars[j].x - c.stars[i].x
			dy := c.stars[j].y - c.stars[i].y
			d := math.Sqrt(dx*dx + dy*dy)
			if d <= c.connectDist {
				nbs = append(nbs, nb{j, d})
			}
		}
		sort.Slice(nbs, func(a, b int) bool { return nbs[a].dist < nbs[b].dist })
		for _, nb := range nbs {
			if connCount[i] >= c.maxConnections || connCount[nb.j] >= c.maxConnections {
				continue
			}
			connCount[i]++
			connCount[nb.j]++
			edges = append(edges, edge{i, nb.j, nb.dist})
		}
	}

	// Draw connection lines (dotted, fading by distance).
	for _, e := range edges {
		a, b := &c.stars[e.i], &c.stars[e.j]
		alpha := (1.0 - e.dist/c.connectDist) * 0.55
		color := Lerp(c.theme.Dim[a.paletteIdx%len(c.theme.Dim)], c.theme.Palette[a.paletteIdx], alpha)
		c.drawDots(screen, int(a.x+0.5), int(a.y+0.5), int(b.x+0.5), int(b.y+0.5), color)
	}

	// Draw stars on top of the lines.
	for i := range c.stars {
		s := &c.stars[i]
		brightness := 0.55 + 0.45*math.Sin(s.twinkle)
		dim := c.theme.Dim[s.paletteIdx%len(c.theme.Dim)]
		pal := c.theme.Palette[s.paletteIdx]
		color := Lerp(dim, pal, brightness)
		// Bright center flash at peak twinkle.
		if brightness > 0.9 {
			color = Lerp(pal, c.theme.Bright, (brightness-0.9)*10)
		}
		x, y := int(s.x+0.5), int(s.y+0.5)
		if x >= 0 && x < c.w && y >= 0 && y < c.h {
			screen.SetContent(x, y, starGlyphs[s.glyphIdx], nil, color.Style())
		}
	}
}

// drawDots interpolates from (x0,y0) to (x1,y1) and places '·' at each step.
// Endpoints are skipped so stars always render on top.
func (c *Constellation) drawDots(screen tcell.Screen, x0, y0, x1, y1 int, color RGBColor) {
	dx := x1 - x0
	dy := y1 - y0
	steps := absInt(dx)
	if d := absInt(dy); d > steps {
		steps = d
	}
	if steps < 2 {
		return
	}
	style := color.Style()
	for i := 1; i < steps; i++ {
		t := float64(i) / float64(steps)
		x := x0 + int(float64(dx)*t+0.5)
		y := y0 + int(float64(dy)*t+0.5)
		if x >= 0 && x < c.w && y >= 0 && y < c.h {
			screen.SetContent(x, y, '·', nil, style)
		}
	}
}
