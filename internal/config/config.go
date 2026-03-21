// Package config handles loading and merging drift configuration.
// Defaults are compiled in; a TOML file at XDG_CONFIG_HOME/drift/config.toml
// is merged on top when present.
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Idle   IdleConfig   `toml:"idle"`
	Engine EngineConfig `toml:"engine"`
	Scene  SceneConfig  `toml:"scene"`
}

// IdleConfig controls the built-in idle timer.
// When using shell integration the shell handles idle detection;
// these values serve as a reference for what to set in TMOUT / DRIFT_TIMEOUT.
type IdleConfig struct {
	Timeout int `toml:"timeout"`
}

type EngineConfig struct {
	FPS int `toml:"fps"`
	// CycleSeconds is how long to show each scene before cycling. 0 disables cycling.
	CycleSeconds float64 `toml:"cycle_seconds"`
	// Scenes is a comma-separated list of scene names, or "all".
	Scenes  string `toml:"scenes"`
	Theme   string `toml:"theme"`
	Shuffle bool   `toml:"shuffle"`
}

type SceneConfig struct {
	Constellation ConstellationConfig `toml:"constellation"`
	Rain          RainConfig          `toml:"rain"`
	Particles     ParticlesConfig     `toml:"particles"`
	Waveform      WaveformConfig      `toml:"waveform"`
	Pipes         PipesConfig         `toml:"pipes"`
	Maze          MazeConfig          `toml:"maze"`
	Life          LifeConfig          `toml:"life"`
}

type ConstellationConfig struct {
	StarCount      int     `toml:"star_count"`
	ConnectRadius  float64 `toml:"connect_radius"`  // fraction of screen diagonal
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
		Idle: IdleConfig{
			Timeout: 120,
		},
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

	return cfg, nil
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

[idle]
# Informational only — changing this value has no effect.
# The shell integration controls idle activation via DRIFT_TIMEOUT (bash/fish)
# or TMOUT (zsh). Set that in your shell config to change the timeout.
timeout = 120

[engine]
fps           = 30     # target frames per second
cycle_seconds = 60     # seconds per scene, 0 = stay on one scene
scenes        = "all"  # comma-separated list or "all"
theme         = "cosmic" # cosmic | nord | dracula | catppuccin | gruvbox | forest | wildberries | mono | rosepine
shuffle       = true   # randomise scene order

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
`
