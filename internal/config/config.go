// Package config handles loading and merging drift configuration.
// Defaults are compiled in; a TOML file at XDG_CONFIG_HOME/drift/config.toml
// is merged on top when present.
package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/phlx0/drift/internal/scene"
)

type Config struct {
	Engine EngineConfig                  `toml:"engine"`
	Scene  SceneConfig                   `toml:"scene"`
	Theme  map[string]CustomThemeConfig  `toml:"theme"`
}

// CustomThemeConfig defines a user-supplied theme in config.toml.
// Colors are hex strings in #RRGGBB format.
//
//	[theme.mytheme]
//	bright  = "#e8e8e8"
//	palette = ["#ff6b6b", "#4ecdc4", "#45b7d1", "#96ceb4"]
//	dim     = ["#3d1010", "#0e2e2c", "#0e2530", "#1a2e20"]
type CustomThemeConfig struct {
	Bright  string   `toml:"bright"`
	Palette []string `toml:"palette"`
	Dim     []string `toml:"dim"`
}

type EngineConfig struct {
	FPS int `toml:"fps"`
	// CycleSeconds is how long to show each scene before cycling. 0 disables cycling.
	CycleSeconds float64 `toml:"cycle_seconds"`
	// FadeSeconds is the duration of the fade-to-black transition between scenes.
	// 0 disables the transition (instant cut).
	FadeSeconds float64 `toml:"fade_seconds"`
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
	DVD           DVDConfig           `toml:"dvd"`
}

type DVDConfig struct {
	Speed float64 `toml:"speed"` // movement speed multiplier
	Label string  `toml:"label"` // text displayed inside the bouncing logo
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
			FadeSeconds:  0.3,
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
			DVD: DVDConfig{
				Speed: 1.0,
				Label: "drift",
			},
		},
	}
}

// AllThemes returns the merged set of built-in and user-defined themes.
// Custom themes override built-ins with the same name.
func (c *Config) AllThemes() map[string]scene.Theme {
	all := make(map[string]scene.Theme, len(scene.Themes)+len(c.Theme))
	maps.Copy(all, scene.Themes)
	for name, ct := range c.Theme {
		palette := make([]scene.RGBColor, len(ct.Palette))
		dim := make([]scene.RGBColor, len(ct.Dim))
		for i, h := range ct.Palette {
			if col, err := parseHex(h); err == nil {
				palette[i] = col
			}
		}
		for i, h := range ct.Dim {
			if col, err := parseHex(h); err == nil {
				dim[i] = col
			}
		}
		bright, _ := parseHex(ct.Bright)
		all[name] = scene.Theme{Name: name, Palette: palette, Dim: dim, Bright: bright}
	}
	return all
}

// parseHex converts a #RRGGBB hex string to an RGBColor.
func parseHex(s string) (scene.RGBColor, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return scene.RGBColor{}, fmt.Errorf("invalid hex color %q: must be #RRGGBB", "#"+s)
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return scene.RGBColor{}, fmt.Errorf("invalid hex color #%s: %w", s, err)
	}
	return scene.RGBColor{R: uint8(n >> 16), G: uint8((n >> 8) & 0xFF), B: uint8(n & 0xFF)}, nil
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
	if c.Engine.FadeSeconds < 0 {
		errs = append(errs, fmt.Sprintf("engine.fade_seconds must be >= 0, got %.2f", c.Engine.FadeSeconds))
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
	if f := c.Scene.Particles.Friction; f < 0 || f > 1 {
		errs = append(errs, fmt.Sprintf("scene.particles.friction must be between 0.0 and 1.0, got %.2f", f))
	}

	// waveform
	if l := c.Scene.Waveform.Layers; l < 1 || l > 3 {
		errs = append(errs, fmt.Sprintf("scene.waveform.layers must be between 1 and 3, got %d", l))
	}
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

	// dvd
	if s := c.Scene.DVD.Speed; s <= 0 {
		errs = append(errs, fmt.Sprintf("scene.dvd.speed must be > 0, got %.2f", s))
	}

	// custom themes
	for name, ct := range c.Theme {
		if len(ct.Palette) == 0 {
			errs = append(errs, fmt.Sprintf("theme.%s.palette must have at least 1 color", name))
		}
		if len(ct.Dim) == 0 {
			errs = append(errs, fmt.Sprintf("theme.%s.dim must have at least 1 color", name))
		}
		if len(ct.Palette) > 0 && len(ct.Dim) > 0 && len(ct.Palette) != len(ct.Dim) {
			errs = append(errs, fmt.Sprintf("theme.%s: palette (%d colors) and dim (%d colors) must be the same length", name, len(ct.Palette), len(ct.Dim)))
		}
		if ct.Bright != "" {
			if _, err := parseHex(ct.Bright); err != nil {
				errs = append(errs, fmt.Sprintf("theme.%s.bright: %v", name, err))
			}
		}
		for i, h := range ct.Palette {
			if _, err := parseHex(h); err != nil {
				errs = append(errs, fmt.Sprintf("theme.%s.palette[%d]: %v", name, i, err))
			}
		}
		for i, h := range ct.Dim {
			if _, err := parseHex(h); err != nil {
				errs = append(errs, fmt.Sprintf("theme.%s.dim[%d]: %v", name, i, err))
			}
		}
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
fade_seconds  = 0.3   # fade-to-black duration between scenes, 0 = instant cut
scenes        = "all"  # comma-separated list or "all"
theme         = "cosmic" # cosmic | nord | dracula | catppuccin | gruvbox | forest | wildberries | mono | rosepine
shuffle       = true   # randomise scene order
hide_tmux_status = false  # tmux: hide status bar while displaying scene

[scene.constellation]
star_count      = 80
connect_radius  = 0.18  # connection threshold as a fraction of screen diagonal (0.0–1.0)
twinkle         = true  # animate star brightness; false = steady glow
max_connections = 4     # max lines drawn from each star

[scene.rain]
charset = "ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉ0123456789"
density = 0.4
speed   = 1.0

[scene.particles]
count    = 120
gravity  = 0.0   # downward acceleration in cells/s²; negative pulls upward
friction = 0.98  # velocity damping per frame (0.0 = instant stop, 1.0 = no damping)

[scene.waveform]
layers    = 3    # number of overlapping sine waves (1–3)
amplitude = 0.70
speed     = 1.0

[scene.orrery]
bodies      = 8    # number of planets (4–8)
trail_decay = 2.4  # seconds for orbit trails to fade; lower = shorter trails

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

[scene.dvd]
speed = 1.0     # movement speed multiplier
label = "drift" # text displayed inside the bouncing logo

# Custom themes — define your own palette with #RRGGBB hex colors.
# palette and dim must have the same number of entries.
#
# [theme.terminal]
# bright  = "#ffffff"
# palette = ["#ff5555", "#50fa7b", "#f1fa8c", "#bd93f9"]
# dim     = ["#3d0000", "#003d00", "#3d3d00", "#1e003d"]
`
