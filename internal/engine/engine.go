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

// Engine manages scene lifecycle and the main render loop.
type Engine struct {
	cfg config.Config

	screen tcell.Screen
	scenes []scene.Scene
	cur    int   // index into scenes
	theme  scene.Theme

	sceneAge float64 // seconds the current scene has been displayed
}

// New creates an Engine from the given configuration.
func New(cfg config.Config) *Engine {
	return &Engine{cfg: cfg}
}

// Run initialises the terminal and blocks until a keypress or interrupt.
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

	// Resolve theme.
	t, ok := scene.Themes[e.cfg.Engine.Theme]
	if !ok {
		t = scene.Themes["cosmic"]
	}
	e.theme = t

	// Build scene list.
	e.scenes = e.buildScenes()
	if len(e.scenes) == 0 {
		return fmt.Errorf("no scenes available")
	}

	// Optionally shuffle.
	if e.cfg.Engine.Shuffle {
		rand.Shuffle(len(e.scenes), func(i, j int) {
			e.scenes[i], e.scenes[j] = e.scenes[j], e.scenes[i]
		})
	}

	// Initialise first scene.
	w, h := screen.Size()
	e.scenes[e.cur].Init(w, h, t)

	// Event pump goroutine.
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
				return nil // any key exits
			case *tcell.EventMouse:
				return nil // any click exits
			case *tcell.EventResize:
				w, h = screen.Size()
				e.scenes[e.cur].Resize(w, h)
				screen.Sync()
			}

		case now := <-ticker.C:
			dt := now.Sub(lastTick).Seconds()
			lastTick = now
			// Cap dt to prevent large jumps after e.g. sleep/wake.
			if dt > 0.1 {
				dt = 0.1
			}

			cur := e.scenes[e.cur]
			cur.Update(dt)

			screen.Fill(' ', tcell.StyleDefault)
			cur.Draw(screen)
			screen.Show()

			// Scene cycling.
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

// buildScenes returns the ordered list of scenes to cycle through.
func (e *Engine) buildScenes() []scene.Scene {
	spec := strings.TrimSpace(e.cfg.Engine.Scenes)
	if spec == "" || spec == "all" {
		return scene.All()
	}

	var result []scene.Scene
	for _, name := range strings.Split(spec, ",") {
		name = strings.TrimSpace(name)
		if s := scene.ByName(name); s != nil {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return scene.All()
	}
	return result
}
