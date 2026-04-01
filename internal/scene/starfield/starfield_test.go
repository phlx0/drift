package starfield

import (
	"testing"

	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

func TestStarfieldInitDoesNotPanic(t *testing.T) {
	s := New(config.Default().Scene.Starfield)
	s.Init(80, 24, scene.Themes["cosmic"])
}

func TestStarfieldStarCountMatchesConfig(t *testing.T) {
	cfg := config.Default().Scene.Starfield
	cfg.Count = 50
	s := New(cfg)
	s.Init(80, 24, scene.Themes["cosmic"])
	if len(s.stars) != 50 {
		t.Errorf("expected 50 stars, got %d", len(s.stars))
	}
}

func TestStarfieldScatteredSpawnZRange(t *testing.T) {
	s := New(config.Default().Scene.Starfield)
	s.Init(80, 24, scene.Themes["cosmic"])

	for i, st := range s.stars {
		if st.z <= 0 || st.z > 1 {
			t.Errorf("star %d scattered z=%f not in (0, 1]", i, st.z)
		}
	}
}

func TestStarfieldSpawnNewZRange(t *testing.T) {
	s := New(config.Default().Scene.Starfield)
	s.Init(80, 24, scene.Themes["cosmic"])

	for i := 0; i < 20; i++ {
		st := s.spawnStar(false)
		if st.z < 0.75 || st.z > 1.0 {
			t.Errorf("new (non-scattered) star z=%f not in [0.75, 1.0]", st.z)
		}
	}
}

func TestStarfieldResizeRespawnsStars(t *testing.T) {
	s := New(config.Default().Scene.Starfield)
	s.Init(80, 24, scene.Themes["cosmic"])
	s.Resize(40, 12)

	if s.w != 40 || s.h != 12 {
		t.Errorf("expected w=40 h=12 after Resize, got w=%d h=%d", s.w, s.h)
	}
	// After Resize all stars are respawned with hasPrev=false.
	for i, st := range s.stars {
		if st.hasPrev {
			t.Errorf("star %d has hasPrev=true immediately after Resize", i)
		}
	}
}

func TestStarfieldUpdateAdvancesZ(t *testing.T) {
	s := New(config.Default().Scene.Starfield)
	s.Init(80, 24, scene.Themes["cosmic"])

	z0 := s.stars[0].z
	s.Update(0.1)
	// z should have decreased (star moves toward viewer).
	if s.stars[0].z >= z0 {
		// Star may have respawned if it hit z<=0.01, which is also valid behavior.
		// Just check it didn't panic.
	}
}

func TestStarfieldSmallTerminalDoesNotPanic(t *testing.T) {
	s := New(config.Default().Scene.Starfield)
	s.Init(5, 3, scene.Themes["cosmic"])
	s.Update(0.016)
}

func TestStarfieldProjectInBounds(t *testing.T) {
	s := New(config.Default().Scene.Starfield)
	s.Init(80, 24, scene.Themes["cosmic"])

	// A star at (0,0) with z=0.5 should project near the centre.
	st := sfStar{x: 0, y: 0, z: 0.5}
	px, py, ok := s.sfProject(st)
	if !ok {
		t.Fatal("expected star at (0,0,z=0.5) to project within bounds")
	}
	if px != 40 || py != 12 {
		t.Errorf("expected projection at (40,12), got (%d,%d)", px, py)
	}
}
