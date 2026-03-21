package scene

import (
	"math"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
)

type decoration struct {
	x, y int
	char rune
	col  tcell.Style
}

type OrbitalCherries struct {
	w, h        int
	theme       Theme
	t           float64
	cfg         config.CherriesConfig
	decorations []decoration
}

func NewCherries(cfg config.CherriesConfig) Scene {
	return &OrbitalCherries{
		cfg: cfg,
	}
}

func (c *OrbitalCherries) Name() string {
	return "cherries"
}

func (c *OrbitalCherries) Init(w, h int, t Theme) {
	c.w = w
	c.h = h
	c.theme = t
	c.t = float64(time.Now().UnixNano()) / 1e9

	rand.Seed(time.Now().UnixNano())

	// generate decorations once
	c.decorations = []decoration{}
	cx := c.w / 2
	cy := c.h / 2
	radius := float64(min(c.w, c.h) / 4)
	margin := int(radius) + 3

	for i := 0; i < c.cfg.Decorations; i++ {
		var sx, sy int
		for {
			sx = rand.Intn(c.w)
			sy = rand.Intn(c.h)
			dx := sx - cx
			dy := sy - cy
			dist := math.Hypot(float64(dx), float64(dy))
			if int(dist) > margin {
				break
			}
		}
		r := []rune{'*', '+', '.', '♥', '•'}[rand.Intn(5)]
		color := tcell.StyleDefault.Foreground(tcell.NewRGBColor(
			int32(rand.Intn(256)), int32(rand.Intn(256)), int32(rand.Intn(256))))
		c.decorations = append(c.decorations, decoration{x: sx, y: sy, char: r, col: color})
	}
}

func (c *OrbitalCherries) Update(dt float64) {
	c.t += dt * c.cfg.Speed
}

func (c *OrbitalCherries) Draw(screen tcell.Screen) {
	cx := c.w / 2
	cy := c.h / 2
	radius := float64(min(c.w, c.h) / 4)

	red := tcell.NewRGBColor(220, 40, 80)
	green := tcell.NewRGBColor(80, 200, 120)
	fruit := tcell.StyleDefault.Foreground(red)
	stem := tcell.StyleDefault.Foreground(green)

	numCherries := c.cfg.NumCherries
	for i := 0; i < numCherries; i++ {
		angle := c.t + float64(i)*2*math.Pi/float64(numCherries)
		x := cx + int(radius*math.Cos(angle))
		y := cy + int(radius*math.Sin(angle))

		screen.SetContent(x, y-1, '|', nil, stem)
		screen.SetContent(x+1, y-1, '/', nil, stem)
		screen.SetContent(x, y, 'o', nil, fruit)
	}

	// draw decorations (static)
	for _, d := range c.decorations {
		screen.SetContent(d.x, d.y, d.char, nil, d.col)
	}
}

func (c *OrbitalCherries) Resize(w, h int) {
	c.Init(w, h, c.theme)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
