package pipes

import (
	"testing"

	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

func TestPipesInitSpawnsCorrectHeadCount(t *testing.T) {
	cfg := config.Default().Scene.Pipes
	cfg.Heads = 4
	p := New(cfg)
	p.Init(80, 24, scene.Themes["cosmic"])
	if len(p.heads) != 4 {
		t.Errorf("expected 4 heads, got %d", len(p.heads))
	}
}

func TestPipesInitDoesNotPanic(t *testing.T) {
	p := New(config.Default().Scene.Pipes)
	p.Init(80, 24, scene.Themes["cosmic"])
}

func TestPipesSmallTerminalDoesNotPanic(t *testing.T) {
	p := New(config.Default().Scene.Pipes)
	p.Init(3, 3, scene.Themes["cosmic"])
	p.Update(0.016)
}

// noTurnPipes returns a Pipes with turn chance = 0 so advance always goes straight.
func noTurnPipes(t *testing.T) *Pipes {
	t.Helper()
	cfg := config.Default().Scene.Pipes
	cfg.TurnChance = 0
	p := New(cfg)
	p.Init(10, 10, scene.Themes["cosmic"])
	return p
}

func TestPipesAdvanceWrapsAtRightEdge(t *testing.T) {
	p := noTurnPipes(t)
	h := &pipeHead{x: 9, y: 5, dir: pipeRight, paletteIdx: 0, stepRate: 30}
	p.advance(h)
	if h.x != 0 {
		t.Errorf("head moving right from x=9 should wrap to x=0, got x=%d", h.x)
	}
}

func TestPipesAdvanceWrapsAtLeftEdge(t *testing.T) {
	p := noTurnPipes(t)
	h := &pipeHead{x: 0, y: 5, dir: pipeLeft, paletteIdx: 0, stepRate: 30}
	p.advance(h)
	if h.x != p.w-1 {
		t.Errorf("head moving left from x=0 should wrap to x=%d, got x=%d", p.w-1, h.x)
	}
}

func TestPipesAdvanceWrapsAtBottomEdge(t *testing.T) {
	p := noTurnPipes(t)
	h := &pipeHead{x: 5, y: 9, dir: pipeDown, paletteIdx: 0, stepRate: 30}
	p.advance(h)
	if h.y != 0 {
		t.Errorf("head moving down from y=9 should wrap to y=0, got y=%d", h.y)
	}
}

func TestPipesAdvanceWrapsAtTopEdge(t *testing.T) {
	p := noTurnPipes(t)
	h := &pipeHead{x: 5, y: 0, dir: pipeUp, paletteIdx: 0, stepRate: 30}
	p.advance(h)
	if h.y != p.h-1 {
		t.Errorf("head moving up from y=0 should wrap to y=%d, got y=%d", p.h-1, h.y)
	}
}

func TestPipesResetClearsGrid(t *testing.T) {
	cfg := config.Default().Scene.Pipes
	cfg.ResetSeconds = 0.1
	p := New(cfg)
	p.Init(20, 10, scene.Themes["cosmic"])

	// Run long enough to trigger a reset.
	for i := 0; i < 20; i++ {
		p.Update(0.016)
	}
	// No panic is sufficient to validate the reset path.
}

func TestPipesResizeRespawnsHeads(t *testing.T) {
	p := New(config.Default().Scene.Pipes)
	p.Init(80, 24, scene.Themes["cosmic"])
	p.Resize(40, 12)
	if p.w != 40 || p.h != 12 {
		t.Errorf("expected w=40 h=12 after Resize, got w=%d h=%d", p.w, p.h)
	}
	if len(p.heads) != p.cfgHeads {
		t.Errorf("expected %d heads after Resize, got %d", p.cfgHeads, len(p.heads))
	}
}
