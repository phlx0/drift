package plasma

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

func defaultPlasma() *Plasma {
	p := New(config.Default().Scene.Plasma)
	p.Init(80, 24, scene.Themes["cosmic"])
	return p
}

func TestPlasmaInitDoesNotPanic(t *testing.T) {
	defaultPlasma()
}

func TestPlasmaUpdateDoesNotPanic(t *testing.T) {
	p := defaultPlasma()
	for range 60 {
		p.Update(0.033)
	}
}

func TestPlasmaDrawDoesNotPanic(t *testing.T) {
	p := defaultPlasma()
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen init: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)
	p.Update(0.033)
	p.Draw(screen)
}

func TestPlasmaSmallTerminalDoesNotPanic(t *testing.T) {
	p := New(config.Default().Scene.Plasma)
	p.Init(5, 3, scene.Themes["cosmic"])
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen init: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(5, 3)
	p.Update(0.033)
	p.Draw(screen)
}

func TestPlasmaValueInRange(t *testing.T) {
	p := defaultPlasma()
	p.Update(1.0)
	for cy := 0; cy < p.h; cy++ {
		for cx := 0; cx < p.w; cx++ {
			v := p.value(cx, cy)
			if v < 0 || v > 1 {
				t.Fatalf("value(%d,%d) = %f, want [0,1]", cx, cy, v)
			}
		}
	}
}

func TestPlasmaResizeDoesNotPanic(t *testing.T) {
	p := defaultPlasma()
	p.Resize(40, 12)
	p.Update(0.033)
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen init: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(40, 12)
	p.Draw(screen)
}
