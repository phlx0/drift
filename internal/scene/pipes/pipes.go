package pipes

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

const (
	pipeRight = iota
	pipeDown
	pipeLeft
	pipeUp
)

var (
	pipeDx = [4]int{1, 0, -1, 0}
	pipeDy = [4]int{0, 1, 0, -1}

	pipeStraight = [4]rune{
		pipeRight: '─',
		pipeDown:  '│',
		pipeLeft:  '─',
		pipeUp:    '│',
	}

	pipeCorner = [4][4]rune{
		pipeRight: {0, '┐', 0, '┘'},
		pipeDown:  {'└', 0, '┘', 0},
		pipeLeft:  {0, '┌', 0, '└'},
		pipeUp:    {'┌', 0, '┐', 0},
	}
)

type pipeCell struct {
	ch    rune
	color scene.RGBColor
}

type pipeHead struct {
	x, y       int
	dir        int
	paletteIdx int
	stepTimer  float64
	stepRate   float64
}

type Pipes struct {
	w, h       int
	theme      scene.Theme
	rng        *rand.Rand
	heads      []pipeHead
	grid       [][]pipeCell
	resetTimer float64

	cfgHeads        int
	cfgTurnChance   float64
	cfgSpeed        float64
	cfgResetSeconds float64
}

func New(cfg config.PipesConfig) *Pipes {
	return &Pipes{
		cfgHeads:        cfg.Heads,
		cfgTurnChance:   cfg.TurnChance,
		cfgSpeed:        cfg.Speed,
		cfgResetSeconds: cfg.ResetSeconds,
	}
}

func (p *Pipes) Name() string { return "pipes" }

func (p *Pipes) Init(w, h int, t scene.Theme) {
	p.w, p.h = w, h
	p.theme = t
	p.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	p.rebuildGrid()
	p.spawnHeads()
	p.resetTimer = 0
}

func (p *Pipes) Resize(w, h int) {
	p.w, p.h = w, h
	p.rebuildGrid()
	p.spawnHeads()
}

func (p *Pipes) rebuildGrid() {
	p.grid = make([][]pipeCell, p.w)
	for x := range p.grid {
		p.grid[x] = make([]pipeCell, p.h)
	}
}

func (p *Pipes) spawnHeads() {
	p.heads = make([]pipeHead, p.cfgHeads)
	for i := range p.heads {
		p.heads[i] = p.newHead()
	}
}

func (p *Pipes) newHead() pipeHead {
	return pipeHead{
		x:          p.rng.Intn(p.w),
		y:          p.rng.Intn(p.h),
		dir:        p.rng.Intn(4),
		paletteIdx: p.rng.Intn(len(p.theme.Palette)),
		stepTimer:  p.rng.Float64() * 0.1,
		stepRate:   (10.0 + p.rng.Float64()*10.0) * p.cfgSpeed,
	}
}

func (p *Pipes) Update(dt float64) {
	p.resetTimer += dt
	if p.resetTimer >= p.cfgResetSeconds {
		p.rebuildGrid()
		p.spawnHeads()
		p.resetTimer = 0
		return
	}

	for i := range p.heads {
		h := &p.heads[i]
		h.stepTimer += dt
		stepDur := 1.0 / h.stepRate
		for h.stepTimer >= stepDur {
			h.stepTimer -= stepDur
			p.advance(h)
		}
	}
}

func (p *Pipes) advance(h *pipeHead) {
	newDir := h.dir
	if p.rng.Float64() < p.cfgTurnChance {
		perp := [2]int{(h.dir + 1) % 4, (h.dir + 3) % 4}
		newDir = perp[p.rng.Intn(2)]
	}

	if h.x >= 0 && h.x < p.w && h.y >= 0 && h.y < p.h {
		var ch rune
		if newDir != h.dir {
			ch = pipeCorner[h.dir][newDir]
		} else {
			ch = pipeStraight[h.dir]
		}
		p.grid[h.x][h.y] = pipeCell{ch: ch, color: p.theme.Palette[h.paletteIdx]}
	}

	h.dir = newDir
	h.x += pipeDx[h.dir]
	h.y += pipeDy[h.dir]

	if h.x < 0 {
		h.x = p.w - 1
	} else if h.x >= p.w {
		h.x = 0
	}
	if h.y < 0 {
		h.y = p.h - 1
	} else if h.y >= p.h {
		h.y = 0
	}
}

func (p *Pipes) Draw(screen tcell.Screen) {
	for x := 0; x < p.w; x++ {
		for y := 0; y < p.h; y++ {
			c := p.grid[x][y]
			if c.ch == 0 {
				continue
			}
			screen.SetContent(x, y, c.ch, nil, c.color.Style())
		}
	}

	for _, h := range p.heads {
		if h.x >= 0 && h.x < p.w && h.y >= 0 && h.y < p.h {
			screen.SetContent(h.x, h.y, pipeStraight[h.dir], nil, p.theme.Bright.Style())
		}
	}
}
