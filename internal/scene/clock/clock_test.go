package clock

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

func TestClockDigitBitmapsAreNonZero(t *testing.T) {
	for d, bitmap := range clockDigits {
		hasPixel := false
		for _, row := range bitmap {
			if row != 0 {
				hasPixel = true
				break
			}
		}
		if !hasPixel {
			t.Errorf("digit %d bitmap is all zeros", d)
		}
	}
}

func TestClockDigitCountIsTen(t *testing.T) {
	if len(clockDigits) != 10 {
		t.Errorf("expected 10 digit bitmaps (0-9), got %d", len(clockDigits))
	}
}

func TestClockDrawDoesNotPanic(t *testing.T) {
	c := New(config.Default().Scene.Clock)
	c.Init(80, 24, scene.Themes["cosmic"])

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)

	c.Draw(screen)
}

func TestClockSmallTerminalDoesNotPanic(t *testing.T) {
	c := New(config.Default().Scene.Clock)
	c.Init(10, 4, scene.Themes["cosmic"])

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(10, 4)

	c.Draw(screen)
}

func TestClockTinyTerminalDoesNotPanic(t *testing.T) {
	c := New(config.Default().Scene.Clock)
	c.Init(1, 1, scene.Themes["cosmic"])

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(1, 1)

	c.Draw(screen)
}

func TestClockResizeDoesNotPanic(t *testing.T) {
	c := New(config.Default().Scene.Clock)
	c.Init(80, 24, scene.Themes["cosmic"])
	c.Resize(40, 12)
	if c.w != 40 || c.h != 12 {
		t.Errorf("expected w=40 h=12 after Resize, got w=%d h=%d", c.w, c.h)
	}
}

func TestClockShowDateFalseDoesNotPanic(t *testing.T) {
	cfg := config.Default().Scene.Clock
	cfg.ShowDate = false
	c := New(cfg)
	c.Init(80, 24, scene.Themes["cosmic"])

	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)

	c.Draw(screen)
}
