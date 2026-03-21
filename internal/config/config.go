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

type IdleConfig struct {
	Timeout int `toml:"timeout"`
}

type EngineConfig struct {
	FPS          int     `toml:"fps"`
	CycleSeconds float64 `toml:"cycle_seconds"`
	Scenes       string  `toml:"scenes"`
	Theme        string  `toml:"theme"`
	Shuffle      bool    `toml:"shuffle"`
}

// --- Scene Configs ---
type SceneConfig struct {
	Constellation ConstellationConfig `toml:"constellation"`
	Rain          RainConfig          `toml:"rain"`
	Particles     ParticlesConfig     `toml:"particles"`
	Waveform      WaveformConfig      `toml:"waveform"`
	Pipes         PipesConfig         `toml:"pipes"`
	Maze          MazeConfig          `toml:"maze"`
	Cherries      CherriesConfig      `toml:"cherries"`
}

type ConstellationConfig struct {
	StarCount      int     `toml:"star_count"`
	ConnectRadius  float64 `toml:"connect_radius"`
	Twinkle        bool    `toml:"twinkle"`
	MaxConnections int     `toml:"max_connections"`
}

type RainConfig struct {
	Charset string  `toml:"charset"`
	Density float64 `toml:"density"`
	Speed   float64 `toml:"speed"`
}

type ParticlesConfig struct {
	Count    int     `toml:"count"`
	Gravity  float64 `toml:"gravity"`
	Friction float64 `toml:"friction"`
}

type WaveformConfig struct {
	Layers    int     `toml:"layers"`
	Amplitude float64 `toml:"amplitude"`
	Speed     float64 `toml:"speed"`
}

type PipesConfig struct {
	Heads        int     `toml:"heads"`
	TurnChance   float64 `toml:"turn_chance"`
	Speed        float64 `toml:"speed"`
	ResetSeconds float64 `toml:"reset_seconds"`
}

type MazeConfig struct {
	PauseSeconds float64 `toml:"pause_seconds"`
	FadeSeconds  float64 `toml:"fade_seconds"`
	Speed        float64 `toml:"speed"`
}

type CherriesConfig struct {
	NumCherries int     `toml:"num_cherries"`
	Decorations int     `toml:"decorations"`
	Speed       float64 `toml:"speed"`
}

// --- Default config ---
func Default() *Config {
	return &Config{
		Idle: IdleConfig{Timeout: 120},
		Engine: EngineConfig{
			FPS:          30,
			CycleSeconds: 60,
			Scenes:       "all",
			Theme:        "cosmic",
			Shuffle:      true,
		},
		Scene: SceneConfig{
			Constellation: ConstellationConfig{StarCount: 80, ConnectRadius: 0.18, Twinkle: true, MaxConnections: 4},
			Rain:          RainConfig{Charset: "ｱｲｳｴｵｶｷｸｹｺ0123456789", Density: 0.4, Speed: 1.0},
			Particles:     ParticlesConfig{Count: 120, Gravity: 0.0, Friction: 0.98},
			Waveform:      WaveformConfig{Layers: 3, Amplitude: 0.7, Speed: 1.0},
			Pipes:         PipesConfig{Heads: 6, TurnChance: 0.15, Speed: 1.0, ResetSeconds: 45.0},
			Maze:          MazeConfig{PauseSeconds: 3.0, FadeSeconds: 2.0, Speed: 1.0},
			Cherries:      CherriesConfig{NumCherries: 28, Decorations: 60, Speed: 0.5},
		},
	}
}

// --- Load / Path ---
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

// --- WriteDefault ---
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

const defaultTOML = `[idle]
timeout = 120

[engine]
fps = 30
cycle_seconds = 60
scenes = "all"
theme = "cosmic"
shuffle = true

[scene.constellation]
star_count = 80
connect_radius = 0.18
twinkle = true
max_connections = 4

[scene.rain]
charset = "ｱｲｳｴｵｶｷｸｹｺ0123456789"
density = 0.4
speed = 1.0

[scene.particles]
count = 120
gravity = 0.0
friction = 0.98

[scene.waveform]
layers = 3
amplitude = 0.7
speed = 1.0

[scene.pipes]
heads = 6
turn_chance = 0.15
speed = 1.0
reset_seconds = 45.0

[scene.maze]
pause_seconds = 3.0
fade_seconds = 2.0
speed = 1.0

[scene.cherries]
num_cherries = 28
decorations = 60
speed = 0.5
`
