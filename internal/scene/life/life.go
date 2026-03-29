package life

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

// Life implements Conway's Game of Life as a drift scene.
// Cells are colored by age: newborns flash bright, mature cells use the
// palette, and long-lived cells shift to dim variants for visual depth.
// The grid wraps at both axes. A forced reset fires after reset_seconds,
// and a stagnation guard resets early if the grid freezes for 3 seconds.
type Life struct {
	w, h  int
	theme scene.Theme
	rng   *rand.Rand

	cells [][]bool
	next  [][]bool
	age   [][]int // consecutive generations this cell has been alive

	stepTimer   float64
	staticTimer float64
	resetTimer  float64
	changed     bool

	cfgDensity      float64
	cfgSpeed        float64
	cfgResetSeconds float64
}

func New(cfg config.LifeConfig) *Life {
	return &Life{
		cfgDensity:      cfg.Density,
		cfgSpeed:        cfg.Speed,
		cfgResetSeconds: cfg.ResetSeconds,
	}
}

func (l *Life) Name() string { return "life" }

func (l *Life) Init(w, h int, t scene.Theme) {
	l.w, l.h = w, h
	l.theme = t
	l.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	l.reset()
}

func (l *Life) Resize(w, h int) {
	l.w, l.h = w, h
	l.reset()
}

func (l *Life) reset() {
	l.cells = make([][]bool, l.w)
	l.next = make([][]bool, l.w)
	l.age = make([][]int, l.w)
	for x := range l.cells {
		l.cells[x] = make([]bool, l.h)
		l.next[x] = make([]bool, l.h)
		l.age[x] = make([]int, l.h)
		for y := range l.cells[x] {
			l.cells[x][y] = l.rng.Float64() < l.cfgDensity
		}
	}
	l.stepTimer = 0
	l.staticTimer = 0
	l.resetTimer = 0
	l.changed = true
}

func (l *Life) neighbours(x, y int) int {
	n := 0
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx := (x + dx + l.w) % l.w
			ny := (y + dy + l.h) % l.h
			if l.cells[nx][ny] {
				n++
			}
		}
	}
	return n
}

func (l *Life) step() {
	l.changed = false
	for x := 0; x < l.w; x++ {
		for y := 0; y < l.h; y++ {
			n := l.neighbours(x, y)
			alive := l.cells[x][y]
			var next bool
			if alive {
				next = n == 2 || n == 3
			} else {
				next = n == 3
			}
			l.next[x][y] = next
			if next != alive {
				l.changed = true
			}
			if next {
				l.age[x][y]++
			} else {
				l.age[x][y] = 0
			}
		}
	}
	l.cells, l.next = l.next, l.cells
}

func (l *Life) Update(dt float64) {
	l.resetTimer += dt
	if l.resetTimer >= l.cfgResetSeconds {
		l.reset()
		return
	}

	stepsPerSec := 8.0 * l.cfgSpeed
	if stepsPerSec < 1 {
		stepsPerSec = 1
	}
	l.stepTimer += dt
	stepDur := 1.0 / stepsPerSec
	for l.stepTimer >= stepDur {
		l.stepTimer -= stepDur
		l.step()
	}

	if !l.changed {
		l.staticTimer += dt
		if l.staticTimer >= 3.0 {
			l.reset()
		}
	} else {
		l.staticTimer = 0
	}
}

func (l *Life) Draw(screen tcell.Screen) {
	pal := l.theme.Palette
	dim := l.theme.Dim

	for x := 0; x < l.w; x++ {
		for y := 0; y < l.h; y++ {
			if !l.cells[x][y] {
				continue
			}
			age := l.age[x][y]
			var style tcell.Style
			switch {
			case age <= 1:
				// newborn — flash bright
				style = l.theme.Bright.Style()
			case age <= 6:
				// young — full palette color
				style = pal[(x+y)%len(pal)].Style()
			default:
				// old — dim variant for depth
				style = dim[(x+y)%len(dim)].Style()
			}
			screen.SetContent(x, y, '█', nil, style)
		}
	}
}
