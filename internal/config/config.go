// Package config handles loading and merging drift configuration.
// Defaults are compiled in; a TOML file at XDG_CONFIG_HOME/drift/config.toml
// is merged on top when present.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Engine EngineConfig `toml:"engine"`
	Scene  SceneConfig  `toml:"scene"`
}

type EngineConfig struct {
	FPS int `toml:"fps"`
	// CycleSeconds is how long to show each scene before cycling. 0 disables cycling.
	CycleSeconds float64 `toml:"cycle_seconds"`
	// Scenes is a comma-separated list of scene names, or "all".
	Scenes  string `toml:"scenes"`
	Theme   string `toml:"theme"`
	Shuffle bool   `toml:"shuffle"`
	// runs `tmux set status off` while drift runs when inside tmux.
	HideTmuxStatus bool `toml:"hide_tmux_status"`
	// Showcase keeps running continuously; arrow/wasd keys cycle scenes and
	// themes, Escape exits.
	Showcase bool `toml:"showcase"`
}

type SceneConfig struct {
	Constellation ConstellationConfig `toml:"constellation"`
	Rain          RainConfig          `toml:"rain"`
	Particles     ParticlesConfig     `toml:"particles"`
	Waveform      WaveformConfig      `toml:"waveform"`
	Orrery        OrreryConfig        `toml:"orrery"`
	Pipes         PipesConfig         `toml:"pipes"`
	Maze          MazeConfig          `toml:"maze"`
	Life          LifeConfig          `toml:"life"`
	Clock         ClockConfig         `toml:"clock"`
	Starfield     StarfieldConfig     `toml:"starfield"`
}

type StarfieldConfig struct {
	Count int     `toml:"count"` // number of stars
	Speed float64 `toml:"speed"` // warp speed multiplier
}

type ClockConfig struct {
	ShowDate bool `toml:"show_date"`
}

type ConstellationConfig struct {
	StarCount      int     `toml:"star_count"`
	ConnectRadius  float64 `toml:"connect_radius"` // fraction of screen diagonal
	Twinkle        bool    `toml:"twinkle"`
	MaxConnections int     `toml:"max_connections"` // per star
}

type RainConfig struct {
	Charset string  `toml:"charset"`
	Density float64 `toml:"density"` // 0.0–1.0
	Speed   float64 `toml:"speed"`   // multiplier
}

type ParticlesConfig struct {
	Count    int     `toml:"count"`
	Gravity  float64 `toml:"gravity"`
	Friction float64 `toml:"friction"`
}

type WaveformConfig struct {
	Layers    int     `toml:"layers"`
	Amplitude float64 `toml:"amplitude"` // 0.0–1.0
	Speed     float64 `toml:"speed"`     // multiplier
}

type OrreryConfig struct {
	Bodies     int     `toml:"bodies"`
	TrailDecay float64 `toml:"trail_decay"`
}

type PipesConfig struct {
	Heads        int     `toml:"heads"`         // number of pipe heads
	TurnChance   float64 `toml:"turn_chance"`   // probability of turning per step (0.0–1.0)
	Speed        float64 `toml:"speed"`         // step rate multiplier
	ResetSeconds float64 `toml:"reset_seconds"` // seconds before screen clears and restarts
}

type MazeConfig struct {
	PauseSeconds float64 `toml:"pause_seconds"` // seconds to display completed maze before fading
	FadeSeconds  float64 `toml:"fade_seconds"`  // seconds the fade-out takes
	Speed        float64 `toml:"speed"`         // build speed multiplier
}

type LifeConfig struct {
	Density      float64 `toml:"density"`       // initial fill probability (0.0–1.0)
	Speed        float64 `toml:"speed"`         // step rate multiplier
	ResetSeconds float64 `toml:"reset_seconds"` // seconds before forced reset
}

func Default() *Config {
	return &Config{
		Engine: EngineConfig{
			FPS:          30,
			CycleSeconds: 60,
			Scenes:       "all",
			Theme:        "cosmic",
			Shuffle:      true,
		},
		Scene: SceneConfig{
			Constellation: ConstellationConfig{
				StarCount:      80,
				ConnectRadius:  0.18,
				Twinkle:        true,
				MaxConnections: 4,
			},
			Rain: RainConfig{
				Charset: "ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉ0123456789",
				Density: 0.4,
				Speed:   1.0,
			},
			Particles: ParticlesConfig{
				Count:    120,
				Gravity:  0.0,
				Friction: 0.98,
			},
			Waveform: WaveformConfig{
				Layers:    3,
				Amplitude: 0.70,
				Speed:     1.0,
			},
			Orrery: OrreryConfig{
				Bodies:     8,
				TrailDecay: 2.4,
			},
			Pipes: PipesConfig{
				Heads:        6,
				TurnChance:   0.15,
				Speed:        1.0,
				ResetSeconds: 45.0,
			},
			Maze: MazeConfig{
				PauseSeconds: 3.0,
				FadeSeconds:  2.0,
				Speed:        1.0,
			},
			Life: LifeConfig{
				Density:      0.35,
				Speed:        1.0,
				ResetSeconds: 30.0,
			},
			Clock: ClockConfig{
				ShowDate: true,
			},
			Starfield: StarfieldConfig{
				Count: 200,
				Speed: 1.0,
			},
		},
	}
}

// Load reads the config file and merges it with compiled-in defaults.
// Missing keys retain their default values. Returns defaults if no file exists.
func Load() (*Config, error) {
	cfg := Default()

	path, err := Path()
	if err != nil {
		return cfg, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks config values for out-of-range settings and returns
// a human-readable error if any are found. All numeric fields with
// meaningful bounds are covered; boolean and free-form string fields
// are intentionally excluded.
func (c *Config) Validate() error {
	var errs []string

	// engine
	if c.Engine.FPS < 1 || c.Engine.FPS > 120 {
		errs = append(errs, fmt.Sprintf("engine.fps must be between 1 and 120, got %d", c.Engine.FPS))
	}
	if c.Engine.CycleSeconds < 0 {
		errs = append(errs, fmt.Sprintf("engine.cycle_seconds must be >= 0, got %.1f", c.Engine.CycleSeconds))
	}

	// constellation
	if n := c.Scene.Constellation.StarCount; n < 1 {
		errs = append(errs, fmt.Sprintf("scene.constellation.star_count must be >= 1, got %d", n))
	}
	if r := c.Scene.Constellation.ConnectRadius; r < 0 || r > 1 {
		errs = append(errs, fmt.Sprintf("scene.constellation.connect_radius must be between 0.0 and 1.0, got %.2f", r))
	}
	if n := c.Scene.Constellation.MaxConnections; n < 1 {
		errs = append(errs, fmt.Sprintf("scene.constellation.max_connections must be >= 1, got %d", n))
	}

	// rain
	if d := c.Scene.Rain.Density; d < 0 || d > 1 {
		errs = append(errs, fmt.Sprintf("scene.rain.density must be between 0.0 and 1.0, got %.2f", d))
	}
	if s := c.Scene.Rain.Speed; s <= 0 {
		errs = append(errs, fmt.Sprintf("scene.rain.speed must be > 0, got %.2f", s))
	}

	// particles
	if n := c.Scene.Particles.Count; n < 1 {
		errs = append(errs, fmt.Sprintf("scene.particles.count must be >= 1, got %d", n))
	}

	// waveform
	if a := c.Scene.Waveform.Amplitude; a < 0 || a > 1 {
		errs = append(errs, fmt.Sprintf("scene.waveform.amplitude must be between 0.0 and 1.0, got %.2f", a))
	}
	if s := c.Scene.Waveform.Speed; s <= 0 {
		errs = append(errs, fmt.Sprintf("scene.waveform.speed must be > 0, got %.2f", s))
	}

	// orrery
	if n := c.Scene.Orrery.Bodies; n < 4 || n > 8 {
		errs = append(errs, fmt.Sprintf("scene.orrery.bodies must be between 4 and 8, got %d", n))
	}
	if d := c.Scene.Orrery.TrailDecay; d <= 0 {
		errs = append(errs, fmt.Sprintf("scene.orrery.trail_decay must be > 0, got %.2f", d))
	}

	// pipes
	if n := c.Scene.Pipes.Heads; n < 1 {
		errs = append(errs, fmt.Sprintf("scene.pipes.heads must be >= 1, got %d", n))
	}
	if tc := c.Scene.Pipes.TurnChance; tc < 0 || tc > 1 {
		errs = append(errs, fmt.Sprintf("scene.pipes.turn_chance must be between 0.0 and 1.0, got %.2f", tc))
	}
	if s := c.Scene.Pipes.Speed; s <= 0 {
		errs = append(errs, fmt.Sprintf("scene.pipes.speed must be > 0, got %.2f", s))
	}
	if rs := c.Scene.Pipes.ResetSeconds; rs <= 0 {
		errs = append(errs, fmt.Sprintf("scene.pipes.reset_seconds must be > 0, got %.2f", rs))
	}

	// maze
	if ps := c.Scene.Maze.PauseSeconds; ps < 0 {
		errs = append(errs, fmt.Sprintf("scene.maze.pause_seconds must be >= 0, got %.2f", ps))
	}
	if fs := c.Scene.Maze.FadeSeconds; fs < 0 {
		errs = append(errs, fmt.Sprintf("scene.maze.fade_seconds must be >= 0, got %.2f", fs))
	}
	if s := c.Scene.Maze.Speed; s <= 0 {
		errs = append(errs, fmt.Sprintf("scene.maze.speed must be > 0, got %.2f", s))
	}

	// life
	if d := c.Scene.Life.Density; d < 0 || d > 1 {
		errs = append(errs, fmt.Sprintf("scene.life.density must be between 0.0 and 1.0, got %.2f", d))
	}
	if s := c.Scene.Life.Speed; s <= 0 {
		errs = append(errs, fmt.Sprintf("scene.life.speed must be > 0, got %.2f", s))
	}
	if rs := c.Scene.Life.ResetSeconds; rs <= 0 {
		errs = append(errs, fmt.Sprintf("scene.life.reset_seconds must be > 0, got %.2f", rs))
	}

	// starfield
	if n := c.Scene.Starfield.Count; n < 1 {
		errs = append(errs, fmt.Sprintf("scene.starfield.count must be >= 1, got %d", n))
	}
	if s := c.Scene.Starfield.Speed; s <= 0 {
		errs = append(errs, fmt.Sprintf("scene.starfield.speed must be > 0, got %.2f", s))
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// Path returns the config file path, respecting XDG_CONFIG_HOME.
func Path() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	} else if strings.HasPrefix(base, "~/") || base == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if base == "~" {
			base = home
		} else {
			base = filepath.Join(home, strings.TrimPrefix(base, "~/"))
		}
	}

	return filepath.Join(base, "drift", "config.toml"), nil
}

// WriteDefault writes the default config, creating directories as needed.
// Uses O_EXCL so it never overwrites an existing file.
func WriteDefault() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	_, err = f.WriteString(defaultTOML)
	return err
}

const defaultTOML = `# drift configuration
# Generated by: drift config --init
# Full documentation: https://github.com/phlx0/drift

# Idle timeout is controlled by your shell, not drift.
# Set DRIFT_TIMEOUT (bash/fish) or TMOUT (zsh) in your shell config.
# Example: export DRIFT_TIMEOUT=120

[engine]
fps           = 30     # target frames per second
cycle_seconds = 60     # seconds per scene, 0 = stay on one scene
scenes        = "all"  # comma-separated list or "all"
theme         = "cosmic" # cosmic | nord | dracula | catppuccin | gruvbox | forest | wildberries | mono | rosepine
shuffle       = true   # randomise scene order
hide_tmux_status = false  # tmux: hide status bar while displaying scene

[scene.constellation]
star_count      = 80
connect_radius  = 0.18  # fraction of screen diagonal
twinkle         = true
max_connections = 4     # max connections per star

[scene.rain]
charset = "ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉ0123456789"
density = 0.4
speed   = 1.0

[scene.particles]
count    = 120
gravity  = 0.0
friction = 0.98

[scene.waveform]
layers    = 3
amplitude = 0.70
speed     = 1.0

[scene.orrery]
bodies      = 8    # clamped to at least 4 for readability
trail_decay = 2.4

[scene.pipes]
heads         = 6
turn_chance   = 0.15  # probability of turning per step (0.0–1.0)
speed         = 1.0
reset_seconds = 45.0  # seconds before screen clears and restarts

[scene.maze]
pause_seconds = 3.0   # seconds to display completed maze before fading
fade_seconds  = 2.0   # seconds the fade-out takes
speed         = 1.0   # build speed multiplier

[scene.life]
density       = 0.35  # initial fill probability (0.0–1.0)
speed         = 1.0   # step rate multiplier
reset_seconds = 30.0  # seconds before forced reset

[scene.clock]
show_date = true  # show date below the time

[scene.starfield]
count = 200   # number of stars
speed = 1.0   # warp speed multiplier
`
