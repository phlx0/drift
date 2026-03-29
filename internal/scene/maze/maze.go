package maze

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

var mazeDirs = [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

type mazeState int

const (
	mazeBuilding mazeState = iota
	mazeComplete
	mazeFading
)

type mazePos struct{ x, y int }

type Maze struct {
	w, h   int
	mw, mh int
	theme  scene.Theme
	rng    *rand.Rand

	walls   [][]bool
	visited [][]bool

	stack      []mazePos
	state      mazeState
	buildTimer float64
	stateTimer float64

	wallList []mazePos
	fadeIdx  int

	cfgPauseSeconds float64
	cfgFadeSeconds  float64
	cfgSpeed        float64
}

func New(cfg config.MazeConfig) *Maze {
	return &Maze{
		cfgPauseSeconds: cfg.PauseSeconds,
		cfgFadeSeconds:  cfg.FadeSeconds,
		cfgSpeed:        cfg.Speed,
	}
}

func (m *Maze) Name() string { return "maze" }

func (m *Maze) Init(w, h int, t scene.Theme) {
	m.w, m.h = w, h
	m.theme = t
	m.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	m.reset()
}

func (m *Maze) Resize(w, h int) {
	m.w, m.h = w, h
	m.reset()
}

func (m *Maze) reset() {
	m.mw = (m.w - 1) / 2
	m.mh = (m.h - 1) / 2
	if m.mw < 1 {
		m.mw = 1
	}
	if m.mh < 1 {
		m.mh = 1
	}

	m.walls = make([][]bool, m.w)
	for x := range m.walls {
		m.walls[x] = make([]bool, m.h)
		for y := range m.walls[x] {
			m.walls[x][y] = true
		}
	}

	m.visited = make([][]bool, m.mw)
	for x := range m.visited {
		m.visited[x] = make([]bool, m.mh)
	}

	sx, sy := m.rng.Intn(m.mw), m.rng.Intn(m.mh)
	m.visited[sx][sy] = true
	m.carve(2*sx+1, 2*sy+1)
	m.stack = []mazePos{{sx, sy}}

	m.state = mazeBuilding
	m.buildTimer = 0
	m.stateTimer = 0
	m.wallList = nil
	m.fadeIdx = 0
}

func (m *Maze) carve(tx, ty int) {
	if tx >= 0 && tx < m.w && ty >= 0 && ty < m.h {
		m.walls[tx][ty] = false
	}
}

func (m *Maze) step() {
	if len(m.stack) == 0 {
		m.state = mazeComplete
		return
	}

	cur := m.stack[len(m.stack)-1]

	var neighbours [][2]int
	for _, d := range mazeDirs {
		nx, ny := cur.x+d[0], cur.y+d[1]
		if nx >= 0 && nx < m.mw && ny >= 0 && ny < m.mh && !m.visited[nx][ny] {
			neighbours = append(neighbours, [2]int{nx, ny})
		}
	}

	if len(neighbours) == 0 {
		m.stack = m.stack[:len(m.stack)-1]
		return
	}

	next := neighbours[m.rng.Intn(len(neighbours))]
	dx, dy := next[0]-cur.x, next[1]-cur.y

	m.carve(2*cur.x+1+dx, 2*cur.y+1+dy)
	m.carve(2*next[0]+1, 2*next[1]+1)

	m.visited[next[0]][next[1]] = true
	m.stack = append(m.stack, mazePos{next[0], next[1]})
}

func (m *Maze) Update(dt float64) {
	switch m.state {
	case mazeBuilding:
		m.buildTimer += dt
		stepsPerSec := float64(m.mw*m.mh) / 10.0 * m.cfgSpeed
		if stepsPerSec < 8 {
			stepsPerSec = 8
		}
		stepDur := 1.0 / stepsPerSec
		for m.buildTimer >= stepDur && m.state == mazeBuilding {
			m.buildTimer -= stepDur
			m.step()
		}

	case mazeComplete:
		m.stateTimer += dt
		if m.stateTimer >= m.cfgPauseSeconds {
			m.wallList = m.wallList[:0]
			for x := 0; x < m.w; x++ {
				for y := 0; y < m.h; y++ {
					if m.walls[x][y] {
						m.wallList = append(m.wallList, mazePos{x, y})
					}
				}
			}
			m.rng.Shuffle(len(m.wallList), func(i, j int) {
				m.wallList[i], m.wallList[j] = m.wallList[j], m.wallList[i]
			})
			m.fadeIdx = 0
			m.state = mazeFading
			m.stateTimer = 0
		}

	case mazeFading:
		m.stateTimer += dt
		target := int(m.stateTimer / m.cfgFadeSeconds * float64(len(m.wallList)))
		if target > len(m.wallList) {
			target = len(m.wallList)
		}
		for m.fadeIdx < target {
			p := m.wallList[m.fadeIdx]
			m.walls[p.x][p.y] = false
			m.fadeIdx++
		}
		if m.fadeIdx >= len(m.wallList) {
			m.reset()
		}
	}
}

// wallCharTable maps a 4-bit neighbour mask to the corresponding box-drawing
// rune. Bits: 0=left, 1=right, 2=up, 3=down.
var wallCharTable = [16]rune{
	0b0000: '·',
	0b0001: '╴',
	0b0010: '╶',
	0b0011: '─',
	0b0100: '╵',
	0b0101: '┘',
	0b0110: '└',
	0b0111: '┴',
	0b1000: '╷',
	0b1001: '┐',
	0b1010: '┌',
	0b1011: '┬',
	0b1100: '│',
	0b1101: '┤',
	0b1110: '├',
	0b1111: '┼',
}

func (m *Maze) wallChar(x, y int) rune {
	isWall := func(tx, ty int) bool {
		return tx >= 0 && tx < m.w && ty >= 0 && ty < m.h && m.walls[tx][ty]
	}
	var mask int
	if isWall(x-1, y) {
		mask |= 1
	}
	if isWall(x+1, y) {
		mask |= 2
	}
	if isWall(x, y-1) {
		mask |= 4
	}
	if isWall(x, y+1) {
		mask |= 8
	}
	return wallCharTable[mask]
}

func (m *Maze) Draw(screen tcell.Screen) {
	pal := m.theme.Palette

	for x := 0; x < m.w; x++ {
		for y := 0; y < m.h; y++ {
			if !m.walls[x][y] {
				continue
			}
			pIdx := (x + y) % len(pal)
			screen.SetContent(x, y, m.wallChar(x, y), nil, pal[pIdx].Style())
		}
	}

	if m.state == mazeBuilding && len(m.stack) > 0 {
		cur := m.stack[len(m.stack)-1]
		tx, ty := 2*cur.x+1, 2*cur.y+1
		if tx < m.w && ty < m.h {
			screen.SetContent(tx, ty, m.wallChar(tx, ty), nil, m.theme.Bright.Style())
		}
	}
}
