// Package engine drives the drift render loop.
package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
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

	screen      tcell.Screen
	scenes      []scene.Scene
	cur         int // index into scenes
	theme       scene.Theme
	sceneAge    float64 // seconds the current scene has been displayed
	shiftTimer  float64 // seconds since last pixel shift
	shiftOX     int     // current x offset for OLED pixel shift
	shiftOY     int     // current y offset for OLED pixel shift
}

func New(cfg config.Config) *Engine {
	return &Engine{cfg: cfg}
}

// Run initialises the terminal and blocks until a keypress or click.
// The terminal is fully restored on return regardless of how Run exits.
func (e *Engine) Run() error {
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

	e.scenes = e.buildScenes()
	if len(e.scenes) == 0 {
		return fmt.Errorf("no scenes available")
	}

	if e.cfg.Engine.Shuffle {
		rand.Shuffle(len(e.scenes), func(i, j int) {
			e.scenes[i], e.scenes[j] = e.scenes[j], e.scenes[i]
		})
	}

	w, h := screen.Size()
	e.scenes[e.cur].Init(w, h, t)

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
			switch ev.(type) {
			case *tcell.EventKey:
				return nil
			case *tcell.EventMouse:
				return nil
			case *tcell.EventResize:
				w, h = screen.Size()
				e.scenes[e.cur].Resize(w, h)
				screen.Sync()
			}

		case now := <-ticker.C:
			dt := now.Sub(lastTick).Seconds()
			lastTick = now
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
			screen.Show()

			if e.cfg.Engine.CycleSeconds > 0 && len(e.scenes) > 1 {
				e.sceneAge += dt
				if e.sceneAge >= e.cfg.Engine.CycleSeconds {
					e.sceneAge = 0
					e.cur = (e.cur + 1) % len(e.scenes)
					w, h = screen.Size()
					e.scenes[e.cur].Init(w, h, t)
				}
			}
		}
	}
}

func (e *Engine) buildScenes() []scene.Scene {
	spec := strings.TrimSpace(e.cfg.Engine.Scenes)
	if spec == "" || spec == "all" {
		return scene.All(e.cfg.Scene)
	}

	var result []scene.Scene
	for _, name := range strings.Split(spec, ",") {
		name = strings.TrimSpace(name)
		if s := scene.ByName(name, e.cfg.Scene); s != nil {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return scene.All(e.cfg.Scene)
	}
	return result
}
