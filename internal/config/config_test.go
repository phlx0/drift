package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultsAreValid(t *testing.T) {
	cfg := Default()

	if cfg.Engine.FPS <= 0 {
		t.Errorf("default FPS should be > 0, got %d", cfg.Engine.FPS)
	}
	if cfg.Engine.Theme == "" {
		t.Error("default Theme should not be empty")
	}
	if cfg.Idle.Timeout <= 0 {
		t.Errorf("default Timeout should be > 0, got %d", cfg.Idle.Timeout)
	}
}

func TestLoadReturnsDefaultsWhenNoFile(t *testing.T) {
	// Set XDG_CONFIG_HOME to a temp dir that has no config file.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	def := Default()
	if cfg.Engine.FPS != def.Engine.FPS {
		t.Errorf("FPS: got %d, want %d", cfg.Engine.FPS, def.Engine.FPS)
	}
	if cfg.Engine.Theme != def.Engine.Theme {
		t.Errorf("Theme: got %q, want %q", cfg.Engine.Theme, def.Engine.Theme)
	}
}

func TestLoadReadsConfigWhenXDGConfigHomeUsesTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "~/.config")

	cfgPath := filepath.Join(home, ".config", "drift", "config.toml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(cfgPath, []byte(`[engine]
fps = 60

[scene.waveform]
layers = 1
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Engine.FPS != 60 {
		t.Fatalf("expected FPS=60, got %d", cfg.Engine.FPS)
	}
	if cfg.Scene.Waveform.Layers != 1 {
		t.Fatalf("expected waveform.layers=1, got %d", cfg.Scene.Waveform.Layers)
	}
}

func TestPathUsesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/testxdg")

	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if p != "/tmp/testxdg/drift/config.toml" {
		t.Errorf("unexpected path: %s", p)
	}
}

func TestPathExpandsTildeInXDGConfigHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "~/.config")

	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	want := home + "/.config/drift/config.toml"
	if p != want {
		t.Errorf("unexpected path: %s (want %s)", p, want)
	}
}
