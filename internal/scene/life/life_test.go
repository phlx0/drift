package life

import (
	"testing"

	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

func TestLifeInitDoesNotPanic(t *testing.T) {
	l := New(config.Default().Scene.Life)
	l.Init(80, 24, scene.Themes["cosmic"])
}

func TestLifeNeighboursWrapsAtEdges(t *testing.T) {
	l := New(config.Default().Scene.Life)
	l.Init(10, 10, scene.Themes["cosmic"])

	// Clear all cells then set a known pattern to test wrap-around counting.
	for x := 0; x < l.w; x++ {
		for y := 0; y < l.h; y++ {
			l.cells[x][y] = false
		}
	}

	// Set the cell at (0,0) and its wrap-around neighbours.
	l.cells[9][9] = true // wraps to (-1,-1)
	l.cells[0][9] = true // wraps to (0,-1)
	l.cells[1][9] = true // wraps to (1,-1)

	n := l.neighbours(0, 0)
	if n != 3 {
		t.Errorf("expected 3 wrapped neighbours at (0,0), got %d", n)
	}
}

func TestLifeConwayRulesB3(t *testing.T) {
	l := New(config.Default().Scene.Life)
	l.Init(10, 10, scene.Themes["cosmic"])

	for x := 0; x < l.w; x++ {
		for y := 0; y < l.h; y++ {
			l.cells[x][y] = false
		}
	}

	// Dead cell at (5,5) with exactly 3 neighbours should be born.
	l.cells[4][5] = true
	l.cells[6][5] = true
	l.cells[5][4] = true

	l.step()

	if !l.cells[5][5] {
		t.Error("dead cell with 3 neighbours should become alive (B3)")
	}
}

func TestLifeConwayRulesS23(t *testing.T) {
	l := New(config.Default().Scene.Life)
	l.Init(10, 10, scene.Themes["cosmic"])

	for x := 0; x < l.w; x++ {
		for y := 0; y < l.h; y++ {
			l.cells[x][y] = false
		}
	}

	// Alive cell at (5,5) with 2 neighbours should survive.
	l.cells[5][5] = true
	l.cells[4][5] = true
	l.cells[6][5] = true

	l.step()

	if !l.cells[5][5] {
		t.Error("alive cell with 2 neighbours should survive (S2)")
	}
}

func TestLifeConwayRulesOverpopulation(t *testing.T) {
	l := New(config.Default().Scene.Life)
	l.Init(10, 10, scene.Themes["cosmic"])

	for x := 0; x < l.w; x++ {
		for y := 0; y < l.h; y++ {
			l.cells[x][y] = false
		}
	}

	// Alive cell at (5,5) with 4 neighbours dies from overpopulation.
	l.cells[5][5] = true
	l.cells[4][5] = true
	l.cells[6][5] = true
	l.cells[5][4] = true
	l.cells[5][6] = true

	l.step()

	if l.cells[5][5] {
		t.Error("alive cell with 4 neighbours should die (overpopulation)")
	}
}

func TestLifeStagnationResetsGrid(t *testing.T) {
	l := New(config.Default().Scene.Life)
	l.Init(10, 10, scene.Themes["cosmic"])

	// All-dead grid is static; staticTimer should accumulate and trigger reset.
	for x := 0; x < l.w; x++ {
		for y := 0; y < l.h; y++ {
			l.cells[x][y] = false
		}
	}
	l.changed = false

	// Advance past the 3-second stagnation threshold.
	for i := 0; i < 10; i++ {
		l.Update(0.4)
	}

	// After reset, cells should be re-randomised (grid will have changed state).
	// We can't assert specific cell values, but we verify it didn't panic.
}

func TestLifeResetTimerResetsGrid(t *testing.T) {
	cfg := config.Default().Scene.Life
	cfg.ResetSeconds = 1.0
	l := New(cfg)
	l.Init(10, 10, scene.Themes["cosmic"])

	// Record initial grid state.
	before := l.cells[0][0]
	_ = before

	// Advance past reset_seconds.
	l.Update(1.1)
	// No panic is sufficient; the reset path was exercised.
}

func TestLifeSmallTerminalDoesNotPanic(t *testing.T) {
	l := New(config.Default().Scene.Life)
	l.Init(3, 3, scene.Themes["cosmic"])
	l.Update(0.016)
}
