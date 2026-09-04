package bonsai

import (
	"testing"

	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

func newTestBonsai(w, h int) *Bonsai {
	b := New(config.Default().Scene.Bonsai)
	b.Init(w, h, scene.Themes["cosmic"])
	return b
}

func TestBonsaiInitStartsGrowing(t *testing.T) {
	b := newTestBonsai(80, 24)
	if b.state != bonsaiGrowing {
		t.Errorf("expected bonsaiGrowing after Init, got %d", b.state)
	}
	if len(b.tips) != 1 {
		t.Errorf("expected exactly one trunk tip after Init, got %d", len(b.tips))
	}
	if len(b.pot) == 0 {
		t.Error("expected a pot to be built at 80x24")
	}
	if len(b.segs) == 0 {
		t.Error("expected the root flare to be stamped at Init")
	}
}

func TestBonsaiGrowsSegments(t *testing.T) {
	b := newTestBonsai(80, 24)
	for i := 0; i < 60; i++ {
		b.Update(0.05)
	}
	if len(b.segs) == 0 {
		t.Error("expected segments after growing for 3s")
	}
}

func TestBonsaiCompletesAndFades(t *testing.T) {
	b := newTestBonsai(80, 24)

	for i := 0; i < 600 && b.state == bonsaiGrowing; i++ {
		b.Update(0.05)
	}
	if b.state == bonsaiGrowing {
		t.Fatal("tree did not finish growing within time budget")
	}

	for i := 0; i < 400 && b.state == bonsaiHolding; i++ {
		b.Update(0.05)
	}
	if b.state != bonsaiFading {
		t.Fatalf("expected bonsaiFading after the hold, got %d", b.state)
	}
	if len(b.fadeList) < len(b.pot) {
		t.Error("fade list should include the pot")
	}
}

func TestBonsaiFadeReachesReset(t *testing.T) {
	b := newTestBonsai(80, 24)
	for i := 0; i < 2000; i++ {
		b.Update(0.05)
		if b.state == bonsaiGrowing && len(b.leaves) == 0 && len(b.tips) == 1 {
			return // reset happened
		}
	}
	t.Error("scene never cycled back to a fresh tree")
}

func TestBonsaiStaysInsideBounds(t *testing.T) {
	const w, h = 40, 14
	b := newTestBonsai(w, h)
	for i := 0; i < 600 && b.state == bonsaiGrowing; i++ {
		b.Update(0.05)
	}
	for _, s := range b.segs {
		if s.x < 0 || s.x >= w || s.y < 0 || s.y > b.groundY {
			t.Fatalf("segment out of bounds: (%d,%d) ground=%d", s.x, s.y, b.groundY)
		}
	}
}

func TestBonsaiSegmentCountCapped(t *testing.T) {
	b := newTestBonsai(80, 24)
	for i := 0; i < 2000; i++ {
		b.Update(0.05)
		if len(b.segs) > b.maxSegs {
			t.Fatalf("segment count %d exceeded cap %d", len(b.segs), b.maxSegs)
		}
	}
}

func TestBonsaiSmallTerminalDoesNotPanic(t *testing.T) {
	b := newTestBonsai(8, 4)
	for i := 0; i < 200; i++ {
		b.Update(0.05)
	}
	if len(b.pot) != 0 {
		t.Error("no pot should be drawn on a tiny terminal")
	}
}

func TestBonsaiTinyTerminalDoesNotPanic(t *testing.T) {
	b := newTestBonsai(1, 1)
	for i := 0; i < 200; i++ {
		b.Update(0.05)
	}
}

func TestBonsaiResizeReinits(t *testing.T) {
	b := newTestBonsai(80, 24)
	for i := 0; i < 40; i++ {
		b.Update(0.05)
	}
	b.Resize(40, 12)

	if b.w != 40 || b.h != 12 {
		t.Errorf("expected w=40 h=12 after Resize, got w=%d h=%d", b.w, b.h)
	}
	if b.state != bonsaiGrowing {
		t.Errorf("expected bonsaiGrowing after Resize, got %d", b.state)
	}
	if len(b.leaves) != 0 || len(b.tips) != 1 {
		t.Errorf("expected a fresh tree after Resize, got %d leaves and %d tips", len(b.leaves), len(b.tips))
	}
}

func TestBonsaiPetalsDriftWhileHeld(t *testing.T) {
	b := newTestBonsai(80, 24)
	for i := 0; i < 600 && b.state == bonsaiGrowing; i++ {
		b.Update(0.05)
	}
	if b.state != bonsaiHolding {
		t.Fatalf("expected bonsaiHolding after growth, got %d", b.state)
	}

	for i := 0; i < 40; i++ {
		b.Update(0.05)
	}
	if len(b.petals) == 0 {
		t.Fatal("expected petals to spawn while the tree is held")
	}
	if len(b.petals) > maxPetals {
		t.Errorf("petal count %d exceeded cap %d", len(b.petals), maxPetals)
	}

	before := b.petals[0].y
	for i := 0; i < 20; i++ {
		b.Update(0.05)
	}
	if len(b.petals) > 0 && b.petals[0].y <= before {
		t.Error("expected petals to fall")
	}
}

func TestBonsaiBranchBudgetRespected(t *testing.T) {
	b := newTestBonsai(80, 24)
	budget := b.branchesLeft
	for i := 0; i < 600 && b.state == bonsaiGrowing; i++ {
		b.Update(0.05)
		if b.branchesLeft < 0 {
			t.Fatal("branch budget went negative")
		}
		if len(b.tips) > maxTips {
			t.Fatalf("concurrent tips %d exceeded cap %d", len(b.tips), maxTips)
		}
	}
	if b.branchesLeft == budget {
		t.Error("expected the tree to spend some of its branch budget")
	}
}

func TestBonsaiPotHasFeetWhenTallEnough(t *testing.T) {
	tall := newTestBonsai(80, 24)
	var feet int
	for _, s := range tall.pot {
		if s.ch == '╵' {
			feet++
		}
	}
	if feet != 2 {
		t.Errorf("expected 2 pot feet on a tall terminal, got %d", feet)
	}

	short := newTestBonsai(40, 9)
	for _, s := range short.pot {
		if s.ch == '╵' {
			t.Error("short terminal pot should have no feet")
		}
	}
}

func TestTrunkWidthTapers(t *testing.T) {
	if w := trunkWidth(0.0); w != 3 {
		t.Errorf("trunk base width = %d, want 3", w)
	}
	if w := trunkWidth(0.45); w != 2 {
		t.Errorf("trunk mid width = %d, want 2", w)
	}
	if w := trunkWidth(0.9); w != 1 {
		t.Errorf("trunk apex width = %d, want 1", w)
	}
}

func TestBranchChar(t *testing.T) {
	cases := []struct {
		dx, dy int
		want   rune
	}{
		{0, -1, '│'},
		{0, 1, '│'},
		{1, 0, '─'},
		{-1, 0, '─'},
		{1, -1, '/'},
		{-1, 1, '/'},
		{-1, -1, '\\'},
		{1, 1, '\\'},
	}
	for _, c := range cases {
		if got := branchChar(c.dx, c.dy); got != c.want {
			t.Errorf("branchChar(%d,%d) = %q, want %q", c.dx, c.dy, got, c.want)
		}
	}
}

func TestClampAimStaysUpward(t *testing.T) {
	for _, a := range []float64{-10, -3.2, -1.5, -0.1, 2.0} {
		got := clampAim(a)
		if got < angleUp-aimSpread || got > angleUp+aimSpread {
			t.Errorf("clampAim(%.2f) = %.2f, outside the upward cone", a, got)
		}
	}
}
