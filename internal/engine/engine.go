// Package engine drives the drift render loop.
package engine

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
	"github.com/phlx0/drift/internal/scenes"
)

// shiftScreen wraps tcell.Screen and translates every SetContent call by
// (ox, oy), discarding writes that land outside the physical screen bounds.
// All other Screen methods delegate to the underlying screen unchanged.
type shiftScreen struct {
	tcell.Screen
	ox, oy int
}

func (s *shiftScreen) SetContent(x, y int, main rune, comb []rune, style tcell.Style) {
	w, h := s.Screen.Size()
	nx, ny := x+s.ox, y+s.oy
	if nx >= 0 && nx < w && ny >= 0 && ny < h {
		s.Screen.SetContent(nx, ny, main, comb, style)
	}
}

type Engine struct {
	cfg config.Config

	screen     tcell.Screen
	scenes     []scene.Scene
	cur        int // index into scenes
	theme      scene.Theme
	sceneAge   float64 // seconds the current scene has been displayed
	shiftTimer float64 // seconds since last pixel shift
	shiftOX    int     // current x offset for OLED pixel shift
	shiftOY    int     // current y offset for OLED pixel shift

	// showcase mode
	themeNames []string // sorted theme names
	themeIdx   int      // index into themeNames
	hudTimer   float64  // seconds remaining to show the HUD overlay
}

func New(cfg config.Config) *Engine {
	return &Engine{cfg: cfg}
}

// Run initialises the terminal and blocks until exit.
// In normal mode any keypress or click exits.
// In showcase mode navigation keys cycle scenes/themes and Escape exits.
// The terminal is fully restored on return regardless of how Run exits.
func (e *Engine) Run() error {
	if e.cfg.Engine.HideTmuxStatus && os.Getenv("TMUX") != "" {
		if err := exec.Command("tmux", "set", "status", "off").Run(); err == nil {
			defer func() { _ = exec.Command("tmux", "set", "status", "on").Run() }()
		}
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("create screen: %w", err)
	}
	if err := screen.Init(); err != nil {
		return fmt.Errorf("init screen: %w", err)
	}
	defer screen.Fini()

	screen.SetStyle(tcell.StyleDefault)
	screen.Clear()
	screen.HideCursor()
	screen.EnableMouse(tcell.MouseButtonEvents)

	e.screen = screen

	t, ok := scene.Themes[e.cfg.Engine.Theme]
	if !ok {
		t = scene.Themes["cosmic"]
	}
	e.theme = t

	if e.cfg.Engine.Showcase {
		e.scenes = scenes.All(e.cfg.Scene)
		names := scene.ThemeNames()
		sort.Strings(names)
		e.themeNames = names
		themeName := e.cfg.Engine.Theme
		if themeName == "" {
			themeName = "cosmic"
		}
		for i, n := range e.themeNames {
			if n == themeName {
				e.themeIdx = i
				break
			}
		}
		e.hudTimer = 3.0 // show controls on start
	} else {
		e.scenes = e.buildScenes()
		if e.cfg.Engine.Shuffle {
			rand.Shuffle(len(e.scenes), func(i, j int) {
				e.scenes[i], e.scenes[j] = e.scenes[j], e.scenes[i]
			})
		}
	}

	if len(e.scenes) == 0 {
		return fmt.Errorf("no scenes available")
	}

	w, h := screen.Size()
	e.scenes[e.cur].Init(w, h, e.theme)

	// PollEvent blocks, so it runs in its own goroutine and sends into a channel.
	events := make(chan tcell.Event, 16)
	stopPump := make(chan struct{})
	go func() {
		for {
			ev := screen.PollEvent()
			if ev == nil {
				return
			}
			select {
			case events <- ev:
			case <-stopPump:
				return
			}
		}
	}()
	defer close(stopPump)

	fps := e.cfg.Engine.FPS
	if fps <= 0 {
		fps = 30
	}
	ticker := time.NewTicker(time.Duration(int64(time.Second) / int64(fps)))
	defer ticker.Stop()

	lastTick := time.Now()

	for {
		select {
		case ev := <-events:
			if done := e.handleEvent(ev, screen, &w, &h); done {
				return nil
			}
		case now := <-ticker.C:
			dt := now.Sub(lastTick).Seconds()
			lastTick = now
			e.handleTick(dt, screen, &w, &h)
		}
	}
}

// handleEvent processes a single terminal event. Returns true if drift should exit.
func (e *Engine) handleEvent(ev tcell.Event, screen tcell.Screen, w, h *int) bool {
	if e.cfg.Engine.Showcase {
		return e.handleShowcaseEvent(ev, screen, w, h)
	}
	switch ev.(type) {
	case *tcell.EventKey, *tcell.EventMouse:
		return true
	case *tcell.EventResize:
		*w, *h = screen.Size()
		e.scenes[e.cur].Resize(*w, *h)
		screen.Sync()
	}
	return false
}

// handleShowcaseEvent handles input in showcase mode.
// Navigation keys cycle scenes/themes; Escape exits.
func (e *Engine) handleShowcaseEvent(ev tcell.Event, screen tcell.Screen, w, h *int) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			return true
		case tcell.KeyUp:
			e.prevScene(*w, *h)
			e.hudTimer = 3.0
		case tcell.KeyDown:
			e.nextScene(*w, *h)
			e.hudTimer = 3.0
		case tcell.KeyLeft:
			e.prevTheme(*w, *h)
			e.hudTimer = 3.0
		case tcell.KeyRight:
			e.nextTheme(*w, *h)
			e.hudTimer = 3.0
		default:
			switch ev.Rune() {
			case 'q', 'Q':
				return true
			case 'w', 'W':
				e.prevScene(*w, *h)
				e.hudTimer = 3.0
			case 's', 'S':
				e.nextScene(*w, *h)
				e.hudTimer = 3.0
			case 'a', 'A':
				e.prevTheme(*w, *h)
				e.hudTimer = 3.0
			case 'd', 'D':
				e.nextTheme(*w, *h)
				e.hudTimer = 3.0
			}
		}
	case *tcell.EventResize:
		*w, *h = screen.Size()
		e.scenes[e.cur].Resize(*w, *h)
		screen.Sync()
	}
	return false
}

func (e *Engine) nextScene(w, h int) {
	e.cur = (e.cur + 1) % len(e.scenes)
	e.sceneAge = 0
	e.scenes[e.cur].Init(w, h, e.theme)
}

func (e *Engine) prevScene(w, h int) {
	e.cur = (e.cur - 1 + len(e.scenes)) % len(e.scenes)
	e.sceneAge = 0
	e.scenes[e.cur].Init(w, h, e.theme)
}

func (e *Engine) nextTheme(w, h int) {
	e.themeIdx = (e.themeIdx + 1) % len(e.themeNames)
	e.theme = scene.Themes[e.themeNames[e.themeIdx]]
	e.scenes[e.cur].Init(w, h, e.theme)
}

func (e *Engine) prevTheme(w, h int) {
	e.themeIdx = (e.themeIdx - 1 + len(e.themeNames)) % len(e.themeNames)
	e.theme = scene.Themes[e.themeNames[e.themeIdx]]
	e.scenes[e.cur].Init(w, h, e.theme)
}

// drawHUD renders the showcase overlay at the bottom two rows of the screen.
// Drawn after the scene, so it appears on top. Visible while hudTimer > 0.
func (e *Engine) drawHUD(screen tcell.Screen, w, h int) {
	if h < 2 {
		return
	}

	hint := "  \u2191\u2193/ws scene   \u2190\u2192/ad theme   esc/q quit  "
	status := fmt.Sprintf("  scene: %-14s \u00b7  theme: %s",
		e.scenes[e.cur].Name(), e.themeNames[e.themeIdx])

	bgDark := tcell.NewRGBColor(15, 15, 20)
	statusStyle := tcell.StyleDefault.
		Foreground(e.theme.Bright.Tcell()).
		Background(bgDark)
	hintStyle := tcell.StyleDefault.
		Foreground(tcell.NewRGBColor(100, 100, 120)).
		Background(bgDark)

	// fill rows with background color first
	for x := range w {
		screen.SetContent(x, h-2, ' ', nil, hintStyle)
		screen.SetContent(x, h-1, ' ', nil, statusStyle)
	}
	for i, ch := range hint {
		if i >= w {
			break
		}
		screen.SetContent(i, h-2, ch, nil, hintStyle)
	}
	for i, ch := range status {
		if i >= w {
			break
		}
		screen.SetContent(i, h-1, ch, nil, statusStyle)
	}
}

// handleTick advances the simulation by dt seconds and redraws the screen.
func (e *Engine) handleTick(dt float64, screen tcell.Screen, w, h *int) {
	// Cap dt to prevent large jumps after sleep/wake.
	if dt > 0.1 {
		dt = 0.1
	}

	// Advance OLED pixel shift: nudge by 1 cell every 10 seconds,
	// cycling through a 3×3 grid so every position resets to (0,0)
	// after 90 seconds.
	e.shiftTimer += dt
	if e.shiftTimer >= 10.0 {
		e.shiftTimer -= 10.0
		e.shiftOX = (e.shiftOX + 1) % 3
		if e.shiftOX == 0 {
			e.shiftOY = (e.shiftOY + 1) % 3
		}
	}

	cur := e.scenes[e.cur]
	cur.Update(dt)

	shifted := &shiftScreen{Screen: screen, ox: e.shiftOX, oy: e.shiftOY}
	screen.Fill(' ', tcell.StyleDefault)
	cur.Draw(shifted)

	if e.cfg.Engine.Showcase {
		if e.hudTimer > 0 {
			e.hudTimer -= dt
			e.drawHUD(screen, *w, *h)
		}
	}

	screen.Show()

	if !e.cfg.Engine.Showcase && e.cfg.Engine.CycleSeconds > 0 && len(e.scenes) > 1 {
		e.sceneAge += dt
		if e.sceneAge >= e.cfg.Engine.CycleSeconds {
			e.sceneAge = 0
			e.cur = (e.cur + 1) % len(e.scenes)
			*w, *h = screen.Size()
			e.scenes[e.cur].Init(*w, *h, e.theme)
		}
	}
}

func (e *Engine) buildScenes() []scene.Scene {
	spec := strings.TrimSpace(e.cfg.Engine.Scenes)
	if spec == "" || spec == "all" {
		return scenes.All(e.cfg.Scene)
	}

	var result []scene.Scene
	for _, name := range strings.Split(spec, ",") {
		name = strings.TrimSpace(name)
		if s := scenes.ByName(name, e.cfg.Scene); s != nil {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return scenes.All(e.cfg.Scene)
	}
	return result
}
