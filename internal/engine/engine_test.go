package engine

import (
	"math/rand"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func simScreen(t *testing.T, w, h int) tcell.SimulationScreen {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(w, h)
	t.Cleanup(s.Fini)
	return s
}

// countScreen counts the cells passed to tcell.
type countScreen struct {
	tcell.Screen
	n int
}

func (c *countScreen) SetContent(x, y int, main rune, comb []rune, style tcell.Style) {
	c.n++
	c.Screen.SetContent(x, y, main, comb, style)
}

type put struct {
	x, y  int
	r     rune
	style tcell.Style
}

// randomFrame returns a frame that keeps most of the previous one and moves
// or recolours a few cells, the way a scene changes from frame to frame.
func randomFrame(rng *rand.Rand, prev []put, w, h int) []put {
	runes := []rune{'*', '.', '│', '─', '░', '█'}
	colors := []tcell.Color{tcell.ColorRed, tcell.ColorGreen, tcell.NewRGBColor(10, 20, 30)}
	var f []put
	for _, p := range prev {
		if rng.Intn(10) > 0 {
			f = append(f, p)
		}
	}
	for range rng.Intn(20) {
		f = append(f, put{
			x: rng.Intn(w), y: rng.Intn(h),
			r:     runes[rng.Intn(len(runes))],
			style: tcell.StyleDefault.Foreground(colors[rng.Intn(len(colors))]),
		})
	}
	return f
}

// TestFrameScreenShowsWhatFillShows draws the same frames the old way (Fill,
// then draw) and through frameScreen, and compares every cell.
func TestFrameScreenShowsWhatFillShows(t *testing.T) {
	const w, h = 40, 12
	old, cur := simScreen(t, w, h), simScreen(t, w, h)
	var fs frameScreen

	rng := rand.New(rand.NewSource(1))
	var frame []put
	for i := range 200 {
		frame = randomFrame(rng, frame, w, h)

		old.Fill(' ', tcell.StyleDefault)
		for _, p := range frame {
			old.SetContent(p.x, p.y, p.r, nil, p.style)
		}
		old.Show()

		fs.Screen = cur
		fs.begin()
		for _, p := range frame {
			fs.SetContent(p.x, p.y, p.r, nil, p.style)
		}
		fs.flush()
		cur.Show()

		oc, _, _ := old.GetContents()
		cc, _, _ := cur.GetContents()
		for j := range oc {
			if string(oc[j].Runes) != string(cc[j].Runes) || oc[j].Style != cc[j].Style {
				t.Fatalf("frame %d, cell (%d,%d): Fill drew %q %v, frameScreen %q %v",
					i, j%w, j/w, oc[j].Runes, oc[j].Style, cc[j].Runes, cc[j].Style)
			}
		}
	}
}

func TestFrameScreenSkipsUnchangedCells(t *testing.T) {
	cs := &countScreen{Screen: simScreen(t, 20, 5)}
	fs := frameScreen{Screen: cs}

	draw := func() {
		fs.begin()
		fs.SetContent(3, 2, '*', nil, tcell.StyleDefault.Foreground(tcell.ColorRed))
		fs.flush()
	}

	draw()
	if cs.n != 20*5 {
		t.Errorf("first frame passed %d cells to tcell, want all %d", cs.n, 20*5)
	}
	cs.n = 0
	draw()
	if cs.n != 0 {
		t.Errorf("an unchanged frame passed %d cells to tcell, want 0", cs.n)
	}

	cs.n = 0
	fs.begin()
	fs.SetContent(4, 2, '*', nil, tcell.StyleDefault.Foreground(tcell.ColorRed))
	fs.flush()
	if cs.n != 2 {
		t.Errorf("moving one cell passed %d cells to tcell, want 2 (old and new place)", cs.n)
	}
}

func TestFrameScreenRedrawsAfterResize(t *testing.T) {
	sim := simScreen(t, 10, 4)
	cs := &countScreen{Screen: sim}
	fs := frameScreen{Screen: cs}

	fs.begin()
	fs.flush()
	sim.SetSize(12, 5)
	cs.n = 0
	fs.begin()
	fs.flush()
	if cs.n != 12*5 {
		t.Errorf("after a resize %d cells were passed to tcell, want all %d", cs.n, 12*5)
	}
}
