package scenes

import (
	"testing"

	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

func TestByName(t *testing.T) {
	cfg := config.Default().Scene
	for _, name := range Names() {
		if s := ByName(name, cfg); s == nil {
			t.Errorf("ByName(%q) returned nil", name)
		}
	}
	if s := ByName("does-not-exist", cfg); s != nil {
		t.Errorf("ByName(unknown) should return nil, got %v", s)
	}
	if s := ByName("orrery", cfg); s == nil {
		t.Fatal("ByName(orrery) returned nil")
	}
}

func TestScenesInitDoNotPanic(t *testing.T) {
	theme := scene.Themes["cosmic"]
	for _, s := range All(config.Default().Scene) {
		t.Run(s.Name(), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("scene %q panicked on Init: %v", s.Name(), r)
				}
			}()
			s.Init(120, 40, theme)
			s.Update(0.033)
			s.Resize(80, 24)
		})
	}
}
