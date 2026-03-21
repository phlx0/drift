package scene

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
)

// Cherries scene
type Cherries struct {
	cfg       config.CherriesConfig
	width     int
	height    int
	positions []Position
	r         *rand.Rand
}

type Position struct {
	X, Y int
}

// NewCherries creates a new Cherries scene.
func NewCherries(cfg config.CherriesConfig) *Cherries {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	positions := make([]Position, cfg.NumCherries)
	for i := range positions {
		positions[i] = Position{
			X: r.Intn(80), // default, will resize in Init
			Y: r.Intn(24),
		}
	}
	return &Cherries{
		cfg:       cfg,
		positions: positions,
		r:         r,
	}
}

func (c *Cherries) Name() string {
	return "cherries"
}

// Init sets up the scene dimensions
func (c *Cherries) Init(w, h int, t Theme) {
	c.width = w
	c.height = h
	for i := range c.positions {
		c.positions[i] = Position{
			X: c.r.Intn(w),
			Y: c.r.Intn(h),
		}
	}
}

// Update moves cherries around (simple example)
func (c *Cherries) Update(dt float64) {
	for i := range c.positions {
		// Move randomly
		c.positions[i].X += c.r.Intn(3) - 1
		c.positions[i].Y += c.r.Intn(3) - 1

		// Wrap around
		if c.positions[i].X < 0 {
			c.positions[i].X = c.width - 1
		} else if c.positions[i].X >= c.width {
			c.positions[i].X = 0
		}
		if c.positions[i].Y < 0 {
			c.positions[i].Y = c.height - 1
		} else if c.positions[i].Y >= c.height {
			c.positions[i].Y = 0
		}
	}
}

// Draw renders the cherries
func (c *Cherries) Draw(screen tcell.Screen) {
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)
	for _, pos := range c.positions {
		screen.SetContent(pos.X, pos.Y, '🍒', nil, style)
	}
}

// Resize updates scene dimensions
func (c *Cherries) Resize(w, h int) {
	c.width = w
	c.height = h
}
