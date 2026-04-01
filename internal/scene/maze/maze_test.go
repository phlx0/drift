package maze

import (
	"testing"

	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

func TestMazeInitStartsInBuildingState(t *testing.T) {
	m := New(config.Default().Scene.Maze)
	m.Init(80, 24, scene.Themes["cosmic"])
	if m.state != mazeBuilding {
		t.Errorf("expected mazeBuilding state after Init, got %d", m.state)
	}
}

func TestMazeInitStackNonEmpty(t *testing.T) {
	m := New(config.Default().Scene.Maze)
	m.Init(80, 24, scene.Themes["cosmic"])
	if len(m.stack) == 0 {
		t.Error("expected non-empty stack after Init")
	}
}

func TestMazeSmallTerminalDoesNotPanic(t *testing.T) {
	m := New(config.Default().Scene.Maze)
	m.Init(3, 3, scene.Themes["cosmic"])
	m.Update(0.016)
}

func TestMazeTinyTerminalDoesNotPanic(t *testing.T) {
	m := New(config.Default().Scene.Maze)
	m.Init(1, 1, scene.Themes["cosmic"])
	m.Update(0.016)
}

func TestMazeWallCharIsolatedWall(t *testing.T) {
	m := New(config.Default().Scene.Maze)
	m.Init(9, 9, scene.Themes["cosmic"])

	// Clear all walls then set exactly one isolated wall.
	for x := 0; x < m.w; x++ {
		for y := 0; y < m.h; y++ {
			m.walls[x][y] = false
		}
	}
	m.walls[4][4] = true

	ch := m.wallChar(4, 4)
	if ch != '·' {
		t.Errorf("isolated wall should render as '·', got %q", ch)
	}
}

func TestMazeWallCharHorizontalSegment(t *testing.T) {
	m := New(config.Default().Scene.Maze)
	m.Init(9, 9, scene.Themes["cosmic"])

	for x := 0; x < m.w; x++ {
		for y := 0; y < m.h; y++ {
			m.walls[x][y] = false
		}
	}
	// Horizontal run: left and right neighbours are walls.
	m.walls[3][4] = true
	m.walls[4][4] = true
	m.walls[5][4] = true

	ch := m.wallChar(4, 4)
	if ch != '─' {
		t.Errorf("wall with left+right neighbours should render as '─', got %q", ch)
	}
}

func TestMazeWallCharVerticalSegment(t *testing.T) {
	m := New(config.Default().Scene.Maze)
	m.Init(9, 9, scene.Themes["cosmic"])

	for x := 0; x < m.w; x++ {
		for y := 0; y < m.h; y++ {
			m.walls[x][y] = false
		}
	}
	// Vertical run: up and down neighbours are walls.
	m.walls[4][3] = true
	m.walls[4][4] = true
	m.walls[4][5] = true

	ch := m.wallChar(4, 4)
	if ch != '│' {
		t.Errorf("wall with up+down neighbours should render as '│', got %q", ch)
	}
}

func TestMazeBuildCompletesWithinBudget(t *testing.T) {
	m := New(config.Default().Scene.Maze)
	m.Init(40, 20, scene.Themes["cosmic"])

	// Advance enough time for the maze to fully build (generous budget).
	for i := 0; i < 500 && m.state == mazeBuilding; i++ {
		m.Update(0.1)
	}

	if m.state == mazeBuilding {
		t.Error("maze did not complete building within time budget")
	}
}

func TestMazeResizeReinits(t *testing.T) {
	m := New(config.Default().Scene.Maze)
	m.Init(80, 24, scene.Themes["cosmic"])
	m.Resize(40, 12)

	if m.w != 40 || m.h != 12 {
		t.Errorf("expected w=40 h=12 after Resize, got w=%d h=%d", m.w, m.h)
	}
	if m.state != mazeBuilding {
		t.Errorf("expected mazeBuilding after Resize, got %d", m.state)
	}
}
