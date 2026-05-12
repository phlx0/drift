package boids

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

func defaultBoids() *Boids {
	b := New(config.Default().Scene.Boids)
	b.Init(80, 24, scene.Themes["cosmic"])
	return b
}

func TestBoidsInitDoesNotPanic(t *testing.T) {
	defaultBoids()
}

func TestBoidsCountMatchesConfig(t *testing.T) {
	cfg := config.Default().Scene.Boids
	cfg.Count = 30
	b := New(cfg)
	b.Init(80, 24, scene.Themes["cosmic"])
	if len(b.boids) != 30 {
		t.Errorf("expected 30 boids, got %d", len(b.boids))
	}
}

func TestBoidsTrailMatchesDimensions(t *testing.T) {
	b := defaultBoids()
	if len(b.trail) != 80 {
		t.Errorf("trail width: got %d, want 80", len(b.trail))
	}
	for x, col := range b.trail {
		if len(col) != 24 {
			t.Errorf("trail[%d] height: got %d, want 24", x, len(col))
		}
	}
}

func TestBoidsResizeRebuildsTrail(t *testing.T) {
	b := defaultBoids()
	b.Resize(40, 12)
	if len(b.trail) != 40 {
		t.Errorf("trail width after Resize: got %d, want 40", len(b.trail))
	}
	for x, col := range b.trail {
		if len(col) != 12 {
			t.Errorf("trail[%d] height after Resize: got %d, want 12", x, len(col))
		}
	}
}

func TestBoidsUpdateDoesNotPanic(t *testing.T) {
	b := defaultBoids()
	for range 60 {
		b.Update(0.033)
	}
}

func TestBoidsPositionStaysInBounds(t *testing.T) {
	b := defaultBoids()
	for i := range 300 {
		b.Update(0.033)
		for _, bd := range b.boids {
			if bd.x < 0 || bd.x >= float64(b.w) || bd.y < 0 || bd.y >= float64(b.h) {
				t.Fatalf("boid out of bounds at update %d: x=%f y=%f", i, bd.x, bd.y)
			}
		}
	}
}

func TestBoidsDrawDoesNotPanic(t *testing.T) {
	b := defaultBoids()
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen init: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)
	b.Draw(screen)
}

func TestBoidsSmallTerminalDoesNotPanic(t *testing.T) {
	b := New(config.Default().Scene.Boids)
	b.Init(5, 3, scene.Themes["cosmic"])
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen init: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(5, 3)
	b.Update(0.033)
	b.Draw(screen)
}
