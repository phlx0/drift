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

func TestValidateAcceptsDefaults(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() should accept defaults, got: %v", err)
	}
}

func TestValidateRejectsOutOfRangeValues(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Config)
	}{
		// engine
		{"negative fps", func(c *Config) { c.Engine.FPS = -1 }},
		{"excessive fps", func(c *Config) { c.Engine.FPS = 200 }},
		{"negative cycle_seconds", func(c *Config) { c.Engine.CycleSeconds = -5 }},

		// constellation
		{"star_count < 1", func(c *Config) { c.Scene.Constellation.StarCount = 0 }},
		{"connect_radius > 1", func(c *Config) { c.Scene.Constellation.ConnectRadius = 1.5 }},
		{"max_connections < 1", func(c *Config) { c.Scene.Constellation.MaxConnections = 0 }},

		// rain
		{"rain density > 1", func(c *Config) { c.Scene.Rain.Density = 1.5 }},
		{"rain speed <= 0", func(c *Config) { c.Scene.Rain.Speed = 0 }},

		// particles
		{"particles count < 1", func(c *Config) { c.Scene.Particles.Count = 0 }},

		// waveform
		{"waveform amplitude > 1", func(c *Config) { c.Scene.Waveform.Amplitude = 2.0 }},
		{"waveform speed <= 0", func(c *Config) { c.Scene.Waveform.Speed = -1 }},

		// orrery
		{"orrery bodies < 4", func(c *Config) { c.Scene.Orrery.Bodies = 2 }},
		{"orrery bodies > 8", func(c *Config) { c.Scene.Orrery.Bodies = 12 }},
		{"orrery trail_decay <= 0", func(c *Config) { c.Scene.Orrery.TrailDecay = 0 }},

		// pipes
		{"pipes heads < 1", func(c *Config) { c.Scene.Pipes.Heads = 0 }},
		{"pipes turn_chance > 1", func(c *Config) { c.Scene.Pipes.TurnChance = 1.5 }},
		{"pipes speed <= 0", func(c *Config) { c.Scene.Pipes.Speed = 0 }},
		{"pipes reset_seconds <= 0", func(c *Config) { c.Scene.Pipes.ResetSeconds = -1 }},

		// maze
		{"maze pause_seconds < 0", func(c *Config) { c.Scene.Maze.PauseSeconds = -1 }},
		{"maze fade_seconds < 0", func(c *Config) { c.Scene.Maze.FadeSeconds = -1 }},
		{"maze speed <= 0", func(c *Config) { c.Scene.Maze.Speed = 0 }},

		// life
		{"life density < 0", func(c *Config) { c.Scene.Life.Density = -0.1 }},
		{"life speed <= 0", func(c *Config) { c.Scene.Life.Speed = 0 }},
		{"life reset_seconds <= 0", func(c *Config) { c.Scene.Life.ResetSeconds = -1 }},

		// starfield
		{"starfield count < 1", func(c *Config) { c.Scene.Starfield.Count = 0 }},
		{"starfield speed <= 0", func(c *Config) { c.Scene.Starfield.Speed = 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.modify(cfg)
			if err := cfg.Validate(); err == nil {
				t.Errorf("Validate() should return error for %s", tt.name)
			}
		})
	}
}
